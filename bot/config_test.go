package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func TestConfigCurrency(t *testing.T) {
	// Test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("SOMEEUR"))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("SOMEEUR"))

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_CUR).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT`).WithArgs(12345, helpers.USERSET_CUR, "USD").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	bc := NewBotController(db)

	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config: %s", bot.LastSentWhat)
	}

	// Default currency
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config currency", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("/config currency: Unexpected usage help: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Your current currency is set to 'EUR'") {
		t.Errorf("/config currency default: Expected currency to be retrieved from db: %s", bot.LastSentWhat)
	}

	// Currency set in db
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config currency", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("/config currency: Unexpected usage help: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Your current currency is set to 'SOMEEUR") {
		t.Errorf("/config currency set (2): Expected currency to be retrieved from db: %s", bot.LastSentWhat)
	}

	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config currency USD", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("/config currency USD: Unexpected usage help: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "'SOMEEUR' to 'USD'") {
		t.Errorf("/config currency (2): Expected currency to be retrieved from db %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConfigTag(t *testing.T) {
	// Test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	bc := NewBotController(db)
	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config tag invalid amount of parameters", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag invalid amount of parameters: %s", bot.LastSentWhat)
	}

	// SET tag
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TAG).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::userSetting"`).
		WithArgs(12345, helpers.USERSET_TAG, "vacation2021").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config tag vacation2021", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag vacation2021: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "get the tag #vacation2021 added") {
		t.Errorf("/config tag vacation2021 response did not contain set tag: %s", bot.LastSentWhat)
	}

	// GET tag
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TAG).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("vacation2021"))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config tag", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "get the tag #vacation2021 added") {
		t.Errorf("/config tag vacation2021 response did not contain set tag: %s", bot.LastSentWhat)
	}

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TAG).
		WillReturnRows(sqlmock.NewRows([]string{"tag"}).AddRow(nil))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config tag", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "disabled") {
		t.Errorf("/config tag vacation2021 response did not contain set tag: %s", bot.LastSentWhat)
	}

	// DELETE tag
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TAG).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config tag off", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag off: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Disabled") {
		t.Errorf("/config tag off response did not contain 'Disabled': %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConfigHandleNotification(t *testing.T) {
	// Test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	bc := NewBotController(db)

	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	tz, _ := time.Now().Zone()

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("-7"))
	mock.ExpectQuery(`SELECT "delayHours", "notificationHour" FROM "bot::notificationSchedule"`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"delayHours", "notificationHour"}))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Notifications are disabled for open transactions") {
		t.Errorf("Notifications should be disabled: %s", bot.LastSentWhat)
	}

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("3"))
	mock.ExpectQuery(`SELECT "delayHours", "notificationHour" FROM "bot::notificationSchedule"`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"delayHours", "notificationHour"}).AddRow(24, 18))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat),
		"The bot will notify you daily at hour 18 ("+tz+"+3) if transactions are open for more than 1 day") {
		t.Errorf("Notifications should be disabled: %s", bot.LastSentWhat)
	}

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("14"))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify 17", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "invalid parameter") {
		t.Errorf("Single number as param should not be allowed: %s", bot.LastSentWhat)
	}

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("0"))
	mock.ExpectExec(`DELETE FROM "bot::notificationSchedule"`).WithArgs(chat.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify off", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Successfully disabled notifications") {
		t.Errorf("Single param should be allowed for 'off' to disable notifications: %s", bot.LastSentWhat)
	}

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("-2"))
	mock.ExpectExec(`DELETE FROM "bot::notificationSchedule"`).WithArgs(chat.ID).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::notificationSchedule"`).WithArgs(chat.ID, 4*24, 23).WillReturnResult(sqlmock.NewResult(1, 1))
	// Recursively called:
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("-2"))
	mock.ExpectQuery(`SELECT "delayHours", "notificationHour" FROM "bot::notificationSchedule"`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"delayHours", "notificationHour"}).AddRow(4*24, 23))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify 4 23", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat),
		"The bot will notify you daily at hour 23 ("+tz+"-2) if transactions are open for more than 4 days") {
		t.Errorf("Should successfully set notification: %s", bot.LastSentWhat)
	}

	// Invalid hour (0-23)
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("1"))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config notify 4 24", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat),
		"invalid hour (24 is out of valid range 1-23)") {
		t.Errorf("Out of bounds notification hour: %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConfigLeadingSlash(t *testing.T) {
	// Test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	bc := NewBotController(db)
	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	// no value set
	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_OMITCMDSLASH).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config omit_slash", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "currently turned off") {
		t.Errorf("message did not yield current disabled state: %s", bot.LastSentWhat)
	}
	// not true-ish value
	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_OMITCMDSLASH).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("fAlSe"))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config omit_slash", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "currently turned off") {
		t.Errorf("message did not yield current disabled state: %s", bot.LastSentWhat)
	}

	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_OMITCMDSLASH).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("tRuE"))
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config omit_slash", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "currently turned on") {
		t.Errorf("message did not yield current enabled state: %s", bot.LastSentWhat)
	}

	// SET setting
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_OMITCMDSLASH).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "bot::userSetting"`).
		WithArgs(12345, helpers.USERSET_OMITCMDSLASH, "true").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config omit_slash on", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "successfully") {
		t.Errorf("/config omit_slash on: %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestConfigAbout(t *testing.T) {
	// Test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	bc := NewBotController(nil)

	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandConfig(&botTest.MockContext{M: &tb.Message{Text: "/config about", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat),
		"LucaBernstein/beancount\\-bot\\-tg") {
		t.Errorf("Should contain repo link: %s", bot.LastSentWhat)
	}
}
