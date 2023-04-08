package crud

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func IsGroupChat(m *tb.Message) bool {
	return m.Chat.ID != m.Sender.ID
}

func (r *Repo) EnrichUserData(m *tb.Message) error {
	if m == nil {
		return fmt.Errorf("provided message was nil")
	}
	tgChatId := m.Chat.ID
	tgUsername := m.Sender.Username

	if IsGroupChat(m) {
		tgUsername = m.Chat.Title
	}

	userCachePrune(0)
	ce, err := r.getUser(m.Chat.ID)
	if err != nil {
		return err
	}
	if ce == nil {
		LogDbf(r, helpers.TRACE, m, "Creating user for the first time in the 'auth::user' db table")
		_, err := r.db.Exec(`INSERT INTO "auth::user" ("tgChatId", "tgUsername")
			VALUES ($1, $2);`, tgChatId, tgUsername)
		return err
	}
	// Check whether some changeable attributes differ
	if ce.TgUsername != tgUsername {
		LogDbf(r, helpers.TRACE, m, "Updating attributes of user in table 'auth::user' (%s, %s)", ce.TgUsername, tgUsername)
		_, err := r.db.Exec(`UPDATE "auth::user" SET "tgUsername" = $2 WHERE "tgChatId" = $1`, tgChatId, tgUsername)
		return err
	}
	return nil
}

func (r *Repo) DeleteUser(m *tb.Message) error {
	if m == nil {
		return fmt.Errorf("provided message was nil")
	}
	tgChatId := m.Chat.ID

	userCachePrune(tgChatId)

	_, err := r.db.Exec(`DELETE FROM "auth::user" WHERE "tgChatId" = $1`, tgChatId)
	return err
}

// User cache

type User struct {
	TgChatId   int64
	TgUsername string
}

type UserCacheEntry struct {
	Expiry time.Time
	Value  *User
}

const CACHE_VALIDITY = 15 * time.Minute

var USER_CACHE = make(map[int64]*UserCacheEntry)

func userCachePrune(tgChatId int64) {
	for i, ce := range USER_CACHE {
		if ce.Expiry.Before(time.Now()) || (i == tgChatId && i != 0) {
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
		SELECT "tgUsername"
		FROM "auth::user"
		WHERE "tgChatId" = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tgUsername sql.NullString
	if rows.Next() {
		err = rows.Scan(&tgUsername)
		if err != nil {
			return nil, err
		}
		if !tgUsername.Valid {
			tgUsername.String = ""
		}
		user := &User{TgChatId: id, TgUsername: tgUsername.String}
		USER_CACHE[id] = &UserCacheEntry{Value: user, Expiry: time.Now().Add(CACHE_VALIDITY)}
		return user, nil
	}
	return nil, nil
}

func (r *Repo) IndividualsWithNotifications(chatId string) (recipients []string) {
	query := `
		SELECT "tgChatId"
		FROM "auth::user"
	`
	params := []interface{}{}

	if chatId != "" {
		i, err := strconv.ParseInt(chatId, 10, 64)
		if err != nil {
			LogDbf(r, helpers.ERROR, nil, "Error while parsing chatId to int64: %s", err.Error())
		}
		query += `WHERE "tgChatId" = $1`
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

/*
*
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
	var query string
	if strings.ToUpper(helpers.Env("DB_TYPE")) == "POSTGRES" {
		query = `
		SELECT
			overdue."tgChatId",
			overdue."count" overdue,
			COUNT(tx2.*) "allTx"
		FROM
			(
				SELECT DISTINCT
					u."tgChatId",
					COUNT(tx.id)
				FROM
					"auth::user" u
						LEFT OUTER JOIN "bot::userSetting" userset
							ON u."tgChatId" = userset."tgChatId"
								AND userset."setting" = 'user.tzOffset',
					"bot::notificationSchedule" s,
					"bot::transaction" tx
				WHERE
					u."tgChatId" = s."tgChatId" AND
					u."tgChatId" = tx."tgChatId" AND
					
					tx.archived = FALSE AND
					MOD(s."notificationHour" + 24 - CASE WHEN userset."value" IS NULL THEN 0 ELSE userset."value"::DECIMAL END, 24) = $1 AND
					tx.created + INTERVAL '1 hour' * s."delayHours" <= NOW()
				GROUP BY u."tgChatId"
			) AS overdue,
			"bot::transaction" tx2
		WHERE
			tx2."tgChatId" = overdue."tgChatId" AND
			tx2.archived = FALSE
		GROUP BY overdue."tgChatId", overdue."count"
		`
	} else {
		query = `
		WITH tx2 AS (SELECT "tgChatId", COUNT(*) AS "allTx" FROM "bot::transaction" GROUP BY "tgChatId")
		SELECT
			overdue."tgChatId",
			overdue."count" overdue,
			tx2."allTx"
		FROM (
			SELECT u."tgChatId", COUNT(*) AS "count"
			FROM "auth::user" u
				LEFT OUTER JOIN "bot::transaction" tx ON u."tgChatId" = tx."tgChatId" AND tx.archived = FALSE
				JOIN "bot::notificationSchedule" s ON u."tgChatId" = s."tgChatId"
				LEFT OUTER JOIN "bot::userSetting" userset ON u."tgChatId" = userset."tgChatId" AND userset."setting" = 'user.tzOffset'
			WHERE
				datetime(tx.created,'+1 hour', '+' || s."delayHours" || ' hour') <= datetime()
				AND (s."notificationHour" + 24 - CASE WHEN userset."value" IS NULL THEN 0 ELSE userset."value" END)%24 = $1
			GROUP BY u."tgChatId"
		) AS "overdue"
		JOIN "tx2" "tx2" ON overdue."tgChatId" = tx2."tgChatId"
		WHERE
			overdue."tgChatId" IS NOT NULL
		GROUP BY overdue."tgChatId"
		`
	}
	return r.db.Query(query, time.Now().UTC().Hour())
}
