package crud_test

import (
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
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

func TestSetCacheLimit(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bc := crud.NewRepo(db)
	message := &tb.Message{Chat: chat}

	mock.ExpectBegin()
	mock.
		ExpectExec(`DELETE FROM "bot::userSetting"`).
		WithArgs(chat.ID, "user.limitCache.txDesc").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectExec(`INSERT INTO "bot::userSetting"`).
		WithArgs(chat.ID, "user.limitCache.txDesc", "23").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = bc.CacheUserSettingSetLimit(message, "txDesc", 23)
	if err != nil {
		t.Errorf("CacheUserSettingSetLimit unexpectedly threw an error: %s", err.Error())
	}

	mock.ExpectBegin()
	mock.
		ExpectExec(`DELETE FROM "bot::userSetting"`).
		WithArgs(chat.ID, "user.limitCache.txDesc").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = bc.CacheUserSettingSetLimit(message, "txDesc", -1)
	if err != nil {
		t.Errorf("CacheUserSettingSetLimit (with delete only) unexpectedly threw an error: %s", err.Error())
	}

	err = bc.CacheUserSettingSetLimit(message, "thisCacheKeyDefinitelyIsInvalidAndShouldFail", 5)
	if err == nil {
		t.Errorf("CacheUserSettingSetLimit should fail for invalid cache key")
	}
	if !strings.Contains(err.Error(), "key you provided is invalid") {
		t.Errorf("CacheUserSettingSetLimit error message should provide further information for failure due to invalid key: %s", err.Error())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetCacheLimit(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bc := crud.NewRepo(db)
	message := &tb.Message{Chat: chat}

	mock.
		ExpectQuery(`SELECT "setting", "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, "user.limitCache.%").
		WillReturnRows(sqlmock.NewRows([]string{"setting", "value"}).AddRow("user.limitCache.txDesc", "79"))

	limits, err := bc.CacheUserSettingGetLimits(message)
	if err != nil {
		t.Errorf("TestSetCacheLimitGet unexpectedly threw an error: %s", err.Error())
	}
	if len(limits) != 3 {
		t.Errorf("TestSetCacheLimitGet unexpectedly threw an error")
	}
	if limits["txDesc"] != 79 {
		t.Errorf("TestSetCacheLimitGet should return values correctly: %d", limits["txDesc"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
