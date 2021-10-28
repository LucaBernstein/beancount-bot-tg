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
	ce := getUser(m.Chat.ID)
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

func getUser(id int64) *User {
	return USER_CACHE[id].Value
}
