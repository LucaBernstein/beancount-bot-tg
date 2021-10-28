package crud

import (
	"log"
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
