package crud

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

var CACHE_LOCAL = make(map[int64]map[string][]string)

func (r *Repo) PutCacheHints(m *tb.Message, values map[string]string) error {
	err := r.FillCache(m)
	if err != nil {
		return err
	}

	keyMappings := map[string]string{
		"description": helpers.STX_DESC,
		"from":        helpers.STX_ACCF,
		"to":          helpers.STX_ACCT,
	}
	for key, value := range values {
		if keyMappings[key] != "" {
			key = keyMappings[key]
		}
		if !helpers.ArrayContains(helpers.AllowedSuggestionTypes(), key) {
			// Don't cache non-suggestible data
			continue
		}
		if helpers.ArrayContains(CACHE_LOCAL[m.Chat.ID][key], value) {
			// TODO: Update all as single statement
			_, err = r.db.Exec(`
				UPDATE "bot::cache"
				SET "lastUsed" = NOW()
				WHERE "tgChatId" = $1 AND "type" = $2 AND "value" = $3`,
				m.Chat.ID, key, value)
			if err != nil {
				return err
			}
		} else {
			// TODO: Insert all as single statement
			_, err = r.db.Exec(`
				INSERT INTO "bot::cache" ("tgChatId", "type", "value")
				VALUES ($1, $2, $3)`,
				m.Chat.ID, key, value)
			if err != nil {
				return err
			}
		}
	}

	return r.FillCache(m)
}

func (r *Repo) PruneUserCachedSuggestions(m *tb.Message) error {
	limits, err := r.CacheUserSettingGetLimits(m)
	if err != nil {
		return err
	}
	for key, limit := range limits {
		if limit < 0 {
			continue
		}
		// Cleanup automatically added values above cache limit
		q := `DELETE FROM "bot::cache" WHERE "tgChatId" = $1 AND "type" = $2`
		params := []interface{}{m.Chat.ID, key}
		if limit != 0 {
			q += ` AND ID NOT IN (SELECT "id" FROM "bot::cache" WHERE "tgChatId" = $1 AND "type" = $2 ORDER BY "lastUsed" DESC LIMIT $3)`
			params = append(params, limit)
		}
		res, err := r.db.Exec(q, params...)
		if err != nil {
			return err
		}
		count, err := res.RowsAffected()
		if err != nil {
			LogDbf(r, helpers.WARN, m, "could not get affected rows count from pruning operation")
		}
		if count > 0 {
			LogDbf(r, helpers.INFO, m, "Pruned %d '%s' suggestions", count, key)
		}
	}
	err = r.FillCache(m)
	if err != nil {
		LogDbf(r, helpers.WARN, m, "could not get refill cache after pruning")
	}
	return nil
}

func (r *Repo) GetCacheHints(m *tb.Message, key string) ([]string, error) {
	if _, exists := CACHE_LOCAL[m.Chat.ID]; !exists {
		LogDbf(r, helpers.TRACE, m, "No cached data found for chat. Will fill cache first.")
		err := r.FillCache(m)
		if err != nil {
			return nil, err
		}
	}
	cacheData := CACHE_LOCAL[m.Chat.ID][key]
	LogDbf(r, helpers.TRACE, m, "Got cached data for chat, key '%s': %v", key, cacheData)
	return cacheData, nil
}

func (r *Repo) FillCache(m *tb.Message) error {
	r.DeleteCache(m)
	rows, err := r.db.Query(`
		SELECT "type", "value"
		FROM "bot::cache"
		WHERE "tgChatId" = $1
		ORDER BY "lastUsed" DESC`,
		m.Chat.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	cache := make(map[string][]string)

	var key string
	var value string
	for rows.Next() {
		err = rows.Scan(&key, &value)
		if err != nil {
			return err
		}
		cache[key] = append(cache[key], value)
	}
	CACHE_LOCAL[m.Chat.ID] = cache
	LogDbf(r, helpers.TRACE, m, "Filled cache for chat: %v", cache)
	return nil
}

// TODO: Prune cache: Free from old entries after time span (async)
func (r *Repo) DeleteCache(m *tb.Message) {
	delete(CACHE_LOCAL, m.Chat.ID)
}

func (r *Repo) DeleteCacheEntries(m *tb.Message, t string, value string) (sql.Result, error) {
	// if value is empty string, delete all entries for this user of this type
	q := `
		DELETE FROM "bot::cache"
		WHERE "tgChatId" = $1 AND "type" = $2
	`
	params := []interface{}{m.Chat.ID, t}
	if value != "" {
		q += ` AND "value" = $3`
		params = append(params, value)
	}
	res, err := r.db.Exec(q, params...)
	if err != nil {
		return nil, err
	}
	return res, r.FillCache(m)
}

func (r *Repo) CacheUserSettingGetLimits(m *tb.Message) (limits map[string]int, err error) {
	userSettingPrefix := helpers.USERSET_LIM_PREFIX

	rows, err := r.db.Query(`SELECT "setting", "value" FROM "bot::userSetting" WHERE "tgChatId" = $1 AND "setting" LIKE $2`,
		m.Chat.ID, userSettingPrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("error selecting userSettingCacheLimits: %s", err.Error())
	}

	allCacheLimits := map[string]int{}

	var setting string
	var value string
	for rows.Next() {
		err = rows.Scan(&setting, &value)
		if err != nil {
			return nil, err
		}
		valueI, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("could not convert stored limit number to integer: %s", err.Error())
		}
		allCacheLimits[strings.TrimPrefix(setting, userSettingPrefix)] = valueI
	}

	// Fill with possible but not modified cache limits
	for _, key := range helpers.AllowedSuggestionTypes() {
		_, exists := allCacheLimits[key]
		if !exists {
			allCacheLimits[key] = -1
		}
	}

	return allCacheLimits, nil
}

// CacheUserSettingSetLimit deletes probably existing and (re)creates caching limit in userSettings
func (r *Repo) CacheUserSettingSetLimit(m *tb.Message, key string, limit int) (err error) {
	if !helpers.ArrayContains(helpers.AllowedSuggestionTypes(), key) {
		return fmt.Errorf("the key you provided is invalid. Should be one of the following: %s", helpers.AllowedSuggestionTypes())
	}

	tx, err := r.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("caching userSettingLimit did not work as db tx could not be begun: %s", err.Error())
	}

	compositeKey := helpers.USERSET_LIM_PREFIX + key

	q := `DELETE FROM "bot::userSetting" WHERE "tgChatId" = $1 AND "setting" = $2;`
	params := []interface{}{m.Chat.ID, compositeKey}

	_, err = tx.Exec(q, params...)
	if err != nil {
		return fmt.Errorf("could not delete existing userSetting on set: %s", err.Error())
	}

	if limit >= 0 {
		q := `INSERT INTO "bot::userSetting" ("tgChatId", "setting", "value") VALUES ($1, $2, $3);`
		params := []interface{}{m.Chat.ID, compositeKey, strconv.Itoa(limit)}
		_, err = tx.Exec(q, params...)
		if err != nil {
			return fmt.Errorf("could not delete existing userSetting on set: %s", err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit db tx for setting userSetting: %s", err.Error())
	}

	return nil
}
