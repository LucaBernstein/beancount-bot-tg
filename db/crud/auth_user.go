package crud

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (r *Repo) EnrichUserData(m *tb.Message) error {
	if m == nil {
		return fmt.Errorf("provided message was nil")
	}
	tgChatId := m.Chat.ID
	tgUserId := m.Sender.ID
	tgUsername := m.Sender.Username

	userCachePrune()
	ce, err := r.getUser(m.Chat.ID)
	if err != nil {
		return err
	}
	if ce == nil {
		LogDbf(r, helpers.TRACE, m, "Creating user for the first time in the 'auth::user' db table")
		_, err := r.db.Exec(`INSERT INTO "auth::user" ("tgChatId", "tgUserId", "tgUsername")
			VALUES ($1, $2, $3);`, tgChatId, tgUserId, tgUsername)
		return err
	}
	// Check whether some changeable attributes differ
	if ce.TgUsername != m.Sender.Username {
		LogDbf(r, helpers.TRACE, m, "Updating attributes of user in table 'auth::user' (%s, %s)", ce.TgUsername, m.Sender.Username)
		_, err := r.db.Exec(`UPDATE "auth::user" SET "tgUserId" = $2, "tgUsername" = $3 WHERE "tgChatId" = $1`, tgChatId, tgUserId, tgUsername)
		return err
	}
	return nil
}

// User cache

type User struct {
	TgChatId   int64
	TgUserId   int
	TgUsername string
}

type UserCacheEntry struct {
	Expiry time.Time
	Value  *User
}

const CACHE_VALIDITY = 15 * time.Minute

var USER_CACHE = make(map[int64]*UserCacheEntry)

func userCachePrune() {
	for i, ce := range USER_CACHE {
		if ce.Expiry.Before(time.Now()) {
			delete(USER_CACHE, i)
		}
	}
}

func (r *Repo) getUser(id int64) (*User, error) {
	value, ok := USER_CACHE[id]
	if ok {
		return value.Value, nil
	}
	rows, err := r.db.Query(`
		SELECT "tgUserId", "tgUsername"
		FROM "auth::user"
		WHERE "tgChatId" = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tgUserId int
	var tgUsername sql.NullString
	if rows.Next() {
		err = rows.Scan(&tgUserId, &tgUsername)
		if err != nil {
			return nil, err
		}
		if !tgUsername.Valid {
			tgUsername.String = ""
		}
		user := &User{TgUserId: tgUserId, TgChatId: id, TgUsername: tgUsername.String}
		USER_CACHE[id] = &UserCacheEntry{Value: user, Expiry: time.Now().Add(CACHE_VALIDITY)}
		return user, nil
	}
	return nil, nil
}

const DEFAULT_CURRENCY = "EUR"

func (r *Repo) UserGetCurrency(m *tb.Message) string {
	rows, err := r.db.Query(`
		SELECT "currency"
		FROM "auth::user"
		WHERE "tgChatId" = $1
	`, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user currency: %s", err.Error())
	}
	defer rows.Close()

	var currency sql.NullString
	if rows.Next() {
		err = rows.Scan(&currency)
		if err != nil {
			LogDbf(r, helpers.ERROR, m, "Encountered error while scanning user currency into var: %s", err.Error())
		}
		if currency.Valid && currency.String != "" {
			return currency.String
		}
	}
	return DEFAULT_CURRENCY
}

func (r *Repo) UserIsAdmin(m *tb.Message) bool {
	rows, err := r.db.Query(`
		SELECT "isAdmin"
		FROM "auth::user"
		WHERE "tgChatId" = $1 AND "tgUserId" = "tgChatId" -- is a private chat
	`, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user currency: %s", err.Error())
	}
	defer rows.Close()

	isAdmin := false
	if rows.Next() {
		err = rows.Scan(&isAdmin)
		if err != nil {
			LogDbf(r, helpers.ERROR, m, "Encountered error while scanning user isAdmin into var: %s", err.Error())
			return false
		}
	}
	return isAdmin
}

func (r *Repo) IndividualsWithNotifications(chatId string) (recipients []string) {
	query := `
		SELECT "tgChatId"
		FROM "auth::user"
		WHERE "tgUserId" = "tgChatId" -- is a private chat
	`
	params := []interface{}{}

	if chatId != "" {
		i, err := strconv.ParseInt(chatId, 10, 64)
		if err != nil {
			LogDbf(r, helpers.ERROR, nil, "Error while parsing chatId to int64: %s", err.Error())
		}
		query += `AND "tgChatId" = $1`
		params = append(params, i)
	}
	rows, err := r.db.Query(query, params...)
	if err != nil {
		LogDbf(r, helpers.ERROR, nil, "Encountered error while getting user currency: %s", err.Error())
	}
	defer rows.Close()

	var rec string
	for rows.Next() {
		err = rows.Scan(&rec)
		if err != nil {
			LogDbf(r, helpers.ERROR, nil, "Encountered error while scanning into var: %s", err.Error())
			return []string{}
		}
		recipients = append(recipients, rec)
	}
	return
}

func (r *Repo) UserSetCurrency(m *tb.Message, currency string) error {
	_, err := r.db.Exec(`
		UPDATE "auth::user"
		SET "currency" = $2
		WHERE "tgChatId" = $1
	`, m.Chat.ID, currency)
	return err
}

func (r *Repo) UserGetTag(m *tb.Message) string {
	rows, err := r.db.Query(`
		SELECT "tag"
		FROM "auth::user"
		WHERE "tgChatId" = $1
	`, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user tag: %s", err.Error())
	}
	defer rows.Close()

	var tag sql.NullString
	if rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			LogDbf(r, helpers.ERROR, m, "Encountered error while scanning user tag into var: %s", err.Error())
		}
		if tag.Valid && tag.String != "" {
			return tag.String
		}
	}
	return ""
}

func (r *Repo) UserSetTag(m *tb.Message, tag string) error {
	q := `UPDATE "auth::user"`
	params := []interface{}{m.Chat.ID}
	if tag == "" {
		q += ` SET "tag" = NULL`
	} else {
		q += ` SET "tag" = $2`
		params = append(params, tag)
	}
	q += ` WHERE "tgChatId" = $1`
	_, err := r.db.Exec(q, params...)
	if err != nil {
		return fmt.Errorf("error while setting user default tag: %s", err.Error())
	}
	return nil
}

func (r *Repo) UserGetNotificationSetting(m *tb.Message) (daysDelay, hour int, err error) {
	rows, err := r.db.Query(`
		SELECT "delayHours", "notificationHour"
		FROM "bot::notificationSchedule"
		WHERE "tgChatId" = $1
	`, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user notification setting: %s", err.Error())
	}
	defer rows.Close()

	var delayHours int
	if rows.Next() {
		err = rows.Scan(&delayHours, &hour)
		if err != nil {
			LogDbf(r, helpers.ERROR, m, "Encountered error while scanning user notification setting into var: %s", err.Error())
		}
		return delayHours / 24, hour, nil
	}
	return -1, -1, nil
}

/**
UserSetNotificationSetting sets user's notification settings.
If daysDelay is < 0, schedule will be disabled.
*/
func (r *Repo) UserSetNotificationSetting(m *tb.Message, daysDelay, hour int) error {
	_, err := r.db.Exec(`DELETE FROM "bot::notificationSchedule" WHERE "tgChatId" = $1;`, m.Chat.ID)
	if daysDelay >= 0 && err == nil { // Condition to enable schedule
		_, err = r.db.Exec(`INSERT INTO "bot::notificationSchedule" ("tgChatId", "delayHours", "notificationHour")
			VALUES ($1, $2, $3);`, m.Chat.ID, daysDelay*24, hour)
	}
	if err != nil {
		return fmt.Errorf("error while setting user notifications schedule: %s", err.Error())
	}
	return nil
}

func (r *Repo) GetUsersToNotify() (*sql.Rows, error) {
	return r.db.Query(`
	SELECT DISTINCT u."tgChatId", COUNT(tx.id)
	FROM "auth::user" u, "bot::notificationSchedule" s, "bot::transaction" tx
	WHERE u."tgChatId" = s."tgChatId" AND s."tgChatId" = tx."tgChatId"
		AND tx.archived = FALSE
		AND s."notificationHour" = $1
		AND tx.created + INTERVAL '1 hour' * s."delayHours" <= NOW()
	GROUP BY u."tgChatId"
	`, time.Now().Hour())
}
