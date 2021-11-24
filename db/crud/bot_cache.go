package crud

import (
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

var CACHE_LOCAL = make(map[int64]map[string][]string)

type GeneralCacheEntry struct {
	accounts []string
	date     string
	desc     string
	amount   string
}

func (r *Repo) PutCacheHints(m *tb.Message, values map[string]string) error {
	err := r.FillCache(m)
	if err != nil {
		return err
	}

	for key, value := range values {
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

func (r *Repo) GetCacheHints(m *tb.Message, key string) ([]string, error) {
	if _, exists := CACHE_LOCAL[m.Chat.ID]; !exists {
		log.Printf("No cached data found for current chat %d. Will fill cache first.", m.Chat.ID)
		err := r.FillCache(m)
		if err != nil {
			return nil, err
		}
	}
	cacheData := CACHE_LOCAL[m.Chat.ID][key]
	log.Printf("Got cached data for chat '%d', key '%s': %v", m.Chat.ID, key, cacheData)
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
	log.Printf("Filled cache for chat %d: %v", m.Chat.ID, cache)
	return nil
}

// TODO: Prune cache: Free from old entries after time span (async)
func (r *Repo) DeleteCache(m *tb.Message) {
	delete(CACHE_LOCAL, m.Chat.ID)
}

func (r *Repo) DeleteCacheEntries(m *tb.Message, t string, value string) error {
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
	_, err := r.db.Exec(q, params...)
	if err != nil {
		return err
	}
	return r.FillCache(m)
}
