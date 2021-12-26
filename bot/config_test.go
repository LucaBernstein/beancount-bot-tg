package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	tb "gopkg.in/tucnak/telebot.v2"
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
		ExpectQuery(`SELECT "currency" FROM "auth::user" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"currency"}))
	mock.
		ExpectQuery(`SELECT "currency" FROM "auth::user" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"currency"}).AddRow("SOMEEUR"))
	mock.
		ExpectQuery(`SELECT "currency" FROM "auth::user" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"currency"}).AddRow("SOMEEUR"))
	mock.
		ExpectExec(`UPDATE "auth::user"`).
		WithArgs(12345, "USD").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)

	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandConfig(&tb.Message{Text: "/config", Chat: chat})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config: %s", bot.LastSentWhat)
	}

	// Default currency
	bc.commandConfig(&tb.Message{Text: "/config currency", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("/config currency: Unexpected usage help: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Your current currency is set to 'EUR'") {
		t.Errorf("/config currency default: Expected currency to be retrieved from db: %s", bot.LastSentWhat)
	}

	// Currency set in db
	bc.commandConfig(&tb.Message{Text: "/config currency", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("/config currency: Unexpected usage help: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Your current currency is set to 'SOMEEUR") {
		t.Errorf("/config currency set (2): Expected currency to be retrieved from db: %s", bot.LastSentWhat)
	}

	bc.commandConfig(&tb.Message{Text: "/config currency USD", Chat: chat})
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
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	mock. // SET
		ExpectExec(`UPDATE "auth::user" SET "tag" = ?`).
		WithArgs(chat.ID, "vacation2021").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock. // GET
		ExpectQuery(`SELECT "tag" FROM "auth::user" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"tag"}).AddRow("vacation2021"))
	mock. // DELETE
		ExpectExec(`UPDATE "auth::user" SET "tag" = NULL WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandConfig(&tb.Message{Text: "/config tag invalid amount of parameters", Chat: chat})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag invalid amount of parameters: %s", bot.LastSentWhat)
	}

	// SET tag
	bc.commandConfig(&tb.Message{Text: "/config tag vacation2021", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag vacation2021: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "get the tag #vacation2021 added") {
		t.Errorf("/config tag vacation2021 response did not contain set tag: %s", bot.LastSentWhat)
	}

	// GET tag
	bc.commandConfig(&tb.Message{Text: "/config tag", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /config") {
		t.Errorf("/config tag: %s", bot.LastSentWhat)
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "get the tag #vacation2021 added") {
		t.Errorf("/config tag vacation2021 response did not contain set tag: %s", bot.LastSentWhat)
	}

	// DELETE tag
	bc.commandConfig(&tb.Message{Text: "/config tag off", Chat: chat})
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
