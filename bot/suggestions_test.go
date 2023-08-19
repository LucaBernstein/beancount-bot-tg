package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	tb "gopkg.in/telebot.v3"
)

func TestSuggestionsHandlingWithSpaces(t *testing.T) {
	// Test dependencies
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	mock.
		ExpectExec(`DELETE FROM "bot::cache"`).
		WithArgs(12345, "description:", "Some description with spaces").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectExec(`DELETE FROM "bot::cache"`).
		WithArgs(12345, "description:", "SomeDescriptionWithoutSpaces").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)

	bot := &botTest.MockBot{}
	bc.AddBotAndStart(bot)

	// missing subcommand
	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("MissingType: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	// missing type
	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions rm", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("MissingType: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions rm description Too Many arguments with spaces", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("TooManyArgs: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions rm description \"Some description with spaces\"", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("Spaced: Bot unexpectedly sent usage help instead of performing command: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions rm description \"SomeDescriptionWithoutSpaces\"", Chat: chat}})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("NotSpaced: Bot unexpectedly sent usage help instead of performing command: %s", bot.LastSentWhat)
	}

	// Add is missing required value
	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions add description ", Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("AddMissingValue: Bot did not send error: %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddMultipleSuggestionsAtOnce(t *testing.T) {
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

	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))
	mock.
		ExpectExec(`INSERT INTO "bot::cache"`).
		WithArgs(12345, "account:from", "First Suggestion").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))
	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))
	mock.
		ExpectExec(`INSERT INTO "bot::cache"`).
		WithArgs(12345, "account:from", "Second Suggestion").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))

	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions add account:from\nFirst Suggestion\nSecond Suggestion", Chat: chat}})

	// Add single suggestion
	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))
	mock.
		ExpectExec(`INSERT INTO "bot::cache"`).
		WithArgs(12345, "account:to", "One lonely suggestion").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectQuery(`SELECT "type", "value" FROM "bot::cache"`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"type", "value"}))

	bc.commandSuggestions(&botTest.MockContext{M: &tb.Message{Text: "/suggestions add account:to \"One lonely suggestion\"", Chat: chat}})

	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Successfully added") {
		t.Errorf("Unexpected bot response: %v", bot.AllLastSentWhat...)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
