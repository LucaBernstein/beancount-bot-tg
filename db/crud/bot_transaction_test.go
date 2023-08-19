package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	tb "gopkg.in/telebot.v3"
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

	mock.ExpectQuery(`SELECT "id", "value", "created" FROM "bot::transaction"`).WithArgs(1122, true).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "value", "created"}).
				AddRow(123, "tx1", "2022-03-30 14:24:50.390084").
				AddRow(124, "tx2", "2022-03-31 14:24:50.390084"),
		)
	txs, err := r.GetTransactions(&tb.Message{Chat: &tb.Chat{ID: 1122}}, true)
	if err != nil {
		t.Errorf("No error should have been returned: %s", err.Error())
	}
	if len(txs) != 2 {
		t.Errorf("Resulting transactions list should contain 2 values: %v", len(txs))
	}
	if txs[0].Tx != "tx1" || txs[1].Tx != "tx2" {
		t.Errorf("Resulting transactions list should contain expected Txs: %v", txs)
	}
	if txs[0].Date != "2022-03-30 14:24:50.390084" || txs[1].Date != "2022-03-31 14:24:50.390084" {
		t.Errorf("Resulting transactions list should contain expected Dates: %v", txs)
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
