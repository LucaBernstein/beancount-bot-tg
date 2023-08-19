package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	tb "gopkg.in/telebot.v3"
)

func TestCacheOnlySuggestible(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mock.
		ExpectQuery(`
			SELECT "type", "value"
			FROM "bot::cache"
			WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))
	mock.
		ExpectExec(`INSERT INTO "bot::cache"`).
		WithArgs(chat.ID, "description:", "description_value").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectQuery(`
			SELECT "type", "value"
			FROM "bot::cache"
			WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))

	bc := crud.NewRepo(db)
	message := &tb.Message{Chat: chat}
	tx, err := bot.CreateSimpleTx("", "${date} ${amount} ${description}")
	if err != nil {
		t.Errorf("PutCacheHints unexpectedly threw an error: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "12.34"})
	tx.Input(&tb.Message{Text: "description_value"})
	cacheData := tx.CacheData()
	err = bc.PutCacheHints(message, cacheData)
	if err != nil {
		t.Errorf("PutCacheHints unexpectedly threw an error: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
