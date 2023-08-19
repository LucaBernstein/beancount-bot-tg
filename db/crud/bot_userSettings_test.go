package crud_test

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

func TestTzOffsetSettingAndGetting(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := crud.NewRepo(db)

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(1122, helpers.USERSET_TZOFF).WillReturnRows(mock.NewRows([]string{"value"}))
	tzOffset := r.UserGetTzOffset(&tb.Message{Chat: &tb.Chat{ID: 1122}})
	if tzOffset != 0 {
		t.Errorf("tzOffset should default to 0")
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(1122, helpers.USERSET_TZOFF).WillReturnRows(mock.NewRows([]string{"value"}).AddRow("-12"))
	tzOffset = r.UserGetTzOffset(&tb.Message{Chat: &tb.Chat{ID: 1122}})
	if tzOffset != -12 {
		t.Errorf("tzOffset should be parsed to -12")
	}

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(1122, helpers.USERSET_TZOFF).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::userSetting"`).WithArgs(1122, helpers.USERSET_TZOFF, "16").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	r.UserSetTzOffset(&tb.Message{Chat: &tb.Chat{ID: 1122}}, 16)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
