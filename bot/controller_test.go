package bot

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	tb "gopkg.in/tucnak/telebot.v2"
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

type MockDBFillCache struct {
}

func (db *MockDBFillCache) Ping() error  { return nil }
func (db *MockDBFillCache) Close() error { return nil }
func (db *MockDBFillCache) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}
func (db *MockDBFillCache) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// fill cache
	// return error as it does not matter, whether this function works for this test
	return nil, fmt.Errorf("Testerror")
}

// GitHub-Issue #16: Panic if plain message without state arrives
func TestTextHandlingWithoutPriorState(t *testing.T) {
	// create test dependencies
	db := &MockDBFillCache{}
	bc := NewBotController(db)
	bot := &MockBot{}
	bc.ConfigureAndAttachBot(bot)

	// Create simple tx and fill it completely
	bc.commandCreateSimpleTx(&tb.Message{Chat: &tb.Chat{ID: 12345}})
	tx := bc.State.states[12345]
	tx.Input(&tb.Message{Text: "17.34"})                                      // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                              // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                         // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"})         // description
	bc.handleTextState(&tb.Message{Chat: &tb.Chat{ID: 12345}, Text: "today"}) // date (via handleTextState)

	// After the first tx is done, send some command
	m := &tb.Message{Chat: &tb.Chat{ID: 12345}}
	bc.handleTextState(m)

	// should catch and send help instead of fail
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "you might need to start a transaction first") {
		t.Errorf("String did not contain substring as expected (was: '%s')", bot.LastSentWhat)
	}
}

// GitHub-Issue #16: Panic if plain message without state arrives
func TestTransactionDeletion(t *testing.T) {
	// create test dependencies
	db := &MockDBFillCache{}
	bc := NewBotController(db)
	bot := &MockBot{}
	bc.ConfigureAndAttachBot(bot)

	bc.commandDeleteTransactions(&tb.Message{Chat: &tb.Chat{ID: 12345}, Text: "/deleteAll"})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "to confirm the deletion of your transactions") {
		t.Errorf("Deletion should require 'yes' confirmation. Got: %s", bot.LastSentWhat)
	}

	bc.commandDeleteTransactions(&tb.Message{Chat: &tb.Chat{ID: 12345}, Text: "/deleteAll YeS"})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Permanently deleted all your transactions") {
		t.Errorf("Deletion should work with confirmation. Got: %s", bot.LastSentWhat)
	}
}
