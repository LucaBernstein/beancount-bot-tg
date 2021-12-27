package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestRecordGetTransaction(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := crud.NewRepo(db)

	mock.ExpectExec(`INSERT INTO "bot::transaction"`).WithArgs(1122, "txContent").WillReturnResult(sqlmock.NewResult(1, 1))
	err = r.RecordTransaction(1122, "txContent")
	if err != nil {
		t.Errorf("No error should have been returned")
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::transaction"`).WithArgs(1122, true).
		WillReturnRows(
			sqlmock.NewRows([]string{"value"}).
				AddRow("tx1").
				AddRow("tx2"),
		)
	txs, err := r.GetTransactions(&tb.Message{Chat: &tb.Chat{ID: 1122}}, true)
	if err != nil {
		t.Errorf("No error should have been returned: %s", err.Error())
	}
	if !helpers.ArraysEqual(txs, []string{"tx1", "tx2"}) {
		t.Errorf("Resulting transactions list should contain expected values: %v", txs)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestArchiveDeleteTransactions(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := crud.NewRepo(db)

	mock.ExpectExec(`
		UPDATE "bot::transaction"
		SET "archived" = TRUE
		WHERE "tgChatId" = ?
	`).WithArgs(1122).
		WillReturnResult(sqlmock.NewResult(1, 1))
	r.ArchiveTransactions(&tb.Message{Chat: &tb.Chat{ID: 1122}})

	mock.ExpectExec(`
		DELETE FROM "bot::transaction"
		WHERE "tgChatId" = ?
	`).WithArgs(1122).
		WillReturnResult(sqlmock.NewResult(1, 1))
	r.DeleteTransactions(&tb.Message{Chat: &tb.Chat{ID: 1122}})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
