package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestConfigCurrency(t *testing.T) {
	// Test dependencies
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
	bc.ConfigureAndAttachBot(bot)

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
