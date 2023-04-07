package bot_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"gopkg.in/telebot.v3"
)

func TestAccountDeletion(t *testing.T) {
	// Create a user, fill some content: comment, setting, cache, state
	repo := crud.NewRepo(db.Connection())
	var chatId int64 = -17
	var err error
	msg := &telebot.Message{Chat: &telebot.Chat{ID: chatId}, Sender: &telebot.User{ID: chatId}}
	err = repo.EnrichUserData(msg)
	if err != nil {
		t.Errorf("Error encountered while issuing command: %e", err)
	}
	err = repo.AddTemplate(chatId, "awesome_template", "some template values")
	if err != nil {
		t.Errorf("Error encountered while issuing command: %e", err)
	}
	err = repo.RecordTransaction(chatId, "my awesome transaction")
	if err != nil {
		t.Errorf("Error encountered while issuing command: %e", err)
	}
	err = repo.PutCacheHints(msg, map[string]string{"some_key:": "some_value"})
	if err != nil {
		t.Errorf("Error encountered while issuing command: %e", err)
	}

	// assert data has been created and persisted
	hints, err := repo.GetCacheHints(msg, "some_key:")
	if err != nil {
		t.Errorf("Getting cache hints should not error: %e", err)
	}
	if len(hints) == 0 || hints[len(hints)-1] != "some_value" {
		t.Errorf("unexpected hints content with len: %d", len(hints))
	}
	tx, err := repo.GetTransactions(msg, false)
	if err != nil {
		t.Errorf("Getting transactions list should not error: %e", err)
	}
	if len(tx) == 0 {
		t.Errorf("Unexpected tx count of 0. Expected at least 1.")
	}

	// Delete account
	err = repo.DeleteUser(msg)
	if err != nil {
		t.Errorf("User deletion should not error: %e", err)
	}

	// assert, content from former user is all gone
	hints, err = repo.GetCacheHints(msg, "some_key")
	if err != nil || len(hints) > 0 {
		t.Errorf("No more hints should be found for user. Got: %d. Err: %e", len(hints), err)
	}
	tx, err = repo.GetTransactions(msg, false)
	if err != nil || len(tx) > 0 {
		t.Errorf("No more transactions should be found for user. Got: %d. Err: %e", len(hints), err)
	}
	// Adding template for that user should fail now:
	err = repo.AddTemplate(chatId, "awesome_template", "some template values")
	if err == nil {
		t.Errorf("Adding a template should fail for a deleted user")
	}
}
