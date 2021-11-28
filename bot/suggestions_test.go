package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	tb "gopkg.in/tucnak/telebot.v2"
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
		WithArgs(12345, "txDesc", "Some description with spaces").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.
		ExpectExec(`DELETE FROM "bot::cache"`).
		WithArgs(12345, "txDesc", "SomeDescriptionWithoutSpaces").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)

	bot := &MockBot{}
	bc.ConfigureAndAttachBot(bot)

	// missing subcommand
	bc.commandSuggestions(&tb.Message{Text: "/suggestions", Chat: chat})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("MissingType: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	// missing type
	bc.commandSuggestions(&tb.Message{Text: "/suggestions rm", Chat: chat})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("MissingType: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&tb.Message{Text: "/suggestions rm txDesc Too Many arguments with spaces", Chat: chat})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("TooManyArgs: Bot unexpectedly did not send usage help: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&tb.Message{Text: "/suggestions rm txDesc \"Some description with spaces\"", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("Spaced: Bot unexpectedly sent usage help instead of performing command: %s", bot.LastSentWhat)
	}

	bc.commandSuggestions(&tb.Message{Text: "/suggestions rm txDesc \"SomeDescriptionWithoutSpaces\"", Chat: chat})
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Usage help") {
		t.Errorf("NotSpaced: Bot unexpectedly sent usage help instead of performing command: %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
