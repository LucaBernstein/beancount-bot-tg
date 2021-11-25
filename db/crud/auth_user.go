package crud

import (
	"log"
	"strconv"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (r *Repo) EnrichUserData(m *tb.Message) error {
	tgChatId := m.Chat.ID
	tgUserId := m.Sender.ID
	tgUsername := m.Sender.Username

	userCachePrune()
	ce, err := r.getUser(m.Chat.ID)
	if err != nil {
		return err
	}
	if ce == nil {
		log.Printf("Creating user for the first time in the 'auth::user' db table: {chatId: %d, userId: %d, username: %s}", tgChatId, tgUserId, tgUsername)
		_, err := r.db.Exec(`INSERT INTO "auth::user" ("tgChatId", "tgUserId", "tgUsername")
			VALUES ($1, $2, $3);`, tgChatId, tgUserId, tgUsername)
		return err
	}
	// Check whether some changeable attributes differ
	if ce.TgUsername != m.Sender.Username {
		log.Printf("Updating attributes of user in table 'auth::user': {chatId: %d, userId: %d, username: %s}", tgChatId, tgUserId, tgUsername)
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
	var tgUsername string
	if rows.Next() {
		err = rows.Scan(&tgUserId, &tgUsername)
		if err != nil {
			return nil, err
		}
		user := &User{TgUserId: tgUserId, TgChatId: id, TgUsername: tgUsername}
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
		log.Printf("Encountered error while getting user currency (user: %d): %s", m.Chat.ID, err.Error())
	}
	defer rows.Close()

	var currency string
	if rows.Next() {
		err = rows.Scan(&currency)
		if err != nil {
			log.Printf("Encountered error while scanning user currency into var (user: %d): %s", m.Chat.ID, err.Error())
		}
		if currency != "" {
			return currency
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
		log.Printf("Encountered error while getting user currency (user: %d): %s", m.Chat.ID, err.Error())
	}
	defer rows.Close()

	isAdmin := false
	if rows.Next() {
		err = rows.Scan(&isAdmin)
		if err != nil {
			log.Printf("Encountered error while scanning user isAdmin into var (user: %d): %s", m.Chat.ID, err.Error())
			return false
		}
	}
	return isAdmin
}

func (r *Repo) IndividualsWithNotifications(myChatId int64, chatId string) (recipients []string) {
	query := `
		SELECT "tgChatId"
		FROM "auth::user"
		WHERE "tgUserId" = "tgChatId" -- is a private chat
			AND "tgChatId" != $1
	`
	params := []interface{}{myChatId}

	if chatId != "" {
		i, err := strconv.ParseInt(chatId, 10, 64)
		if err != nil {
			log.Printf("Error while parsing chatId to int64: %s", err.Error())
		}
		query += `AND "tgChatId" = $2`
		params = append(params, i)
	}
	rows, err := r.db.Query(query, params...)
	if err != nil {
		log.Printf("Encountered error while getting user currency: %s", err.Error())
	}
	defer rows.Close()

	var rec string
	if rows.Next() {
		err = rows.Scan(&rec)
		if err != nil {
			log.Printf("Encountered error while scanning into var: %s", err.Error())
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
