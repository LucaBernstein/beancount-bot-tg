package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestCacheOnlySuggestible(t *testing.T) {
	// create test dependencies
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
	// Should only insert description suggestion into db cache
	mock.
		ExpectExec(`INSERT INTO "bot::cache"`).
		WithArgs(chat.ID, helpers.STX_DESC, "description_value").
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
	err = bc.PutCacheHints(message, map[string]string{helpers.STX_DATE: "2021-01-01", helpers.STX_AMTF: "1234", helpers.STX_DESC: "description_value"})
	if err != nil {
		t.Errorf("PutCacheHints unexpectedly threw an error: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
