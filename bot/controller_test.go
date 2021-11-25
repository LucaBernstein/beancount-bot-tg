package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/DATA-DOG/go-sqlmock"
)

type MockBot struct {
	LastSentWhat interface{}
}

func (b *MockBot) Start()                                           {}
func (b *MockBot) Handle(endpoint interface{}, handler interface{}) {}
func (b *MockBot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	b.LastSentWhat = what
	return nil, nil
}
func (b *MockBot) Me() *tb.User {
	return &tb.User{Username: "Test bot"}
}

// GitHub-Issue #16: Panic if plain message without state arrives
func TestTextHandlingWithoutPriorState(t *testing.T) {
	// create test dependencies
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mock.
		ExpectQuery(`SELECT "currency" FROM "auth::user" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnRows(sqlmock.NewRows([]string{"TEST_CURRENCY"}))
	mock.
		ExpectExec(`INSERT INTO "bot::transaction"`).
		WithArgs(chat.ID, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.ConfigureAndAttachBot(bot)

	// Create simple tx and fill it completely
	bc.commandCreateSimpleTx(&tb.Message{Chat: chat})
	tx := bc.State.states[12345]
	tx.Input(&tb.Message{Text: "17.34"})                                                    // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                                            // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                                       // to
	bc.handleTextState(&tb.Message{Chat: chat, Text: "Buy something in the grocery store"}) // description (via handleTextState)

	// After the first tx is done, send some command
	m := &tb.Message{Chat: chat}
	bc.handleTextState(m)

	// should catch and send help instead of fail
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "you might need to start a transaction first") {
		t.Errorf("String did not contain substring as expected (was: '%s')", bot.LastSentWhat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// GitHub-Issue #16: Panic if plain message without state arrives
func TestTransactionDeletion(t *testing.T) {
	// create test dependencies
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	mock.
		ExpectExec(`DELETE FROM "bot::transaction" WHERE "tgChatId" = ?`).
		WithArgs(chat.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.ConfigureAndAttachBot(bot)

	bc.commandDeleteTransactions(&tb.Message{Chat: chat, Text: "/deleteAll"})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "to confirm the deletion of your transactions") {
		t.Errorf("Deletion should require 'yes' confirmation. Got: %s", bot.LastSentWhat)
	}

	bc.commandDeleteTransactions(&tb.Message{Chat: chat, Text: "/deleteAll YeS"})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Permanently deleted all your transactions") {
		t.Errorf("Deletion should work with confirmation. Got: %s", bot.LastSentWhat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
