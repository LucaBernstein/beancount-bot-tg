package crud_test

import (
	"log"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"gopkg.in/telebot.v3"
)

func TestNonceGeneration(t *testing.T) {
	nonce := crud.GenNonce(3)
	log.Printf("Created nonce: %s", nonce)
	if len(nonce) != 3 {
		t.Errorf("Expected nonce length to be 3. Got nonce: '%s'", nonce)
	}
}

func TestEnsureApiEnabled(t *testing.T) {
	repo := crud.NewRepo(db.Connection())
	var tgChatId int64 = -443322
	m := &telebot.Message{Chat: &telebot.Chat{ID: tgChatId}, Sender: &telebot.User{ID: tgChatId}}
	repo.EnrichUserData(m)
	err := repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "false", tgChatId)
	if err != nil {
		t.Errorf("Error while setting user setting: %e", err)
	}
	err = crud.EnsureApiEnabled(repo, tgChatId)
	if err != helpers.ErrApiDisabled {
		t.Errorf("Should yield error for disabled API: %e", err)
	}

	err = repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", tgChatId)
	if err != nil {
		t.Errorf("Error while setting user setting: %e", err)
	}
	err = crud.EnsureApiEnabled(repo, tgChatId)
	if err != nil {
		t.Errorf("Should not yield error for enabled API: %e", err)
	}
}

func TestApiTokenVerification(t *testing.T) {
	// Create user and enable API
	conn := db.Connection()
	repo := crud.NewRepo(conn)
	var tgChatId int64 = -443387
	m := &telebot.Message{Chat: &telebot.Chat{ID: tgChatId}, Sender: &telebot.User{ID: tgChatId}}
	repo.EnrichUserData(m)
	err := repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", tgChatId)
	if err != nil {
		t.Errorf("Error while setting user setting: %e", err)
	}

	// Cleanup existing tokens from previous runs
	_, err = conn.Exec(`DELETE FROM "app::apiToken" WHERE "tgChatId" = $1`, tgChatId)
	if err != nil {
		t.Errorf("Error while pruning token table: %e", err)
	}

	nonce, err := repo.CreateApiVerification(tgChatId)
	if err != nil {
		t.Errorf("Should not error for initial nonce creation: %e", err)
	}
	log.Printf("Created nonce: %s", nonce)

	// Creating a second nonce immediately should fail:
	_, err = repo.CreateApiVerification(tgChatId)
	if err != helpers.ErrApiTokenChallengeInProgress {
		t.Errorf("Should wait for timeout until recreation of nonce: %e", err)
	}

	// Verify nonce:
	_, err = repo.VerifyApiToken(tgChatId, nonce)
	if err != nil {
		t.Errorf("Should not error for nonce verification: %e", err)
	}

	// Verify nonce again:
	_, err = repo.VerifyApiToken(tgChatId, nonce)
	if err != helpers.ErrApiInvalidTokenVerification {
		t.Errorf("Should error for second nonce verification: %e", err)
	}
}
