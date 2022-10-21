package bot

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	tb "gopkg.in/telebot.v3"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

// GitHub-Issue #16: Panic if plain message without state arrives
func TestTextHandlingWithoutPriorState(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CURRENCY"))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CURRENCY"))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TAG).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("vacation2021"))
	today := time.Now().Format(helpers.BEANCOUNT_DATE_FORMAT)
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	mock.
		ExpectExec(`INSERT INTO "bot::transaction"`).
		WithArgs(chat.ID, today+` * "Buy something in the grocery store" #vacation2021
  Assets:Wallet                               -17.34 TEST_CURRENCY
  Expenses:Groceries
`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	// Create simple tx and fill it completely
	bc.commandCreateSimpleTx(&MockContext{M: &tb.Message{Chat: chat}})
	tx := bc.State.txStates[12345]
	tx.Input(&tb.Message{Text: "17.34"})                                                     // amount
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"})                        // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                                             // from
	bc.handleTextState(&MockContext{M: &tb.Message{Chat: chat, Text: "Expenses:Groceries"}}) // to (via handleTextState)

	// After the first tx is done, send some command
	m := &MockContext{M: &tb.Message{Chat: chat, Sender: &tb.User{ID: chat.ID}}} // same ID: not group chat
	bc.handleTextState(m)

	// should catch and send help instead of fail
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "you might need to start a transaction first") {
		t.Errorf("String did not contain substring as expected (was: '%s')", bot.LastSentWhat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestStartTransactionWithPlainAmountThousandsSeparated(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CURRENCY"))

	bc.handleTextState(&MockContext{M: &tb.Message{Chat: chat, Text: "1,000,000"}})

	debugString := bc.State.txStates[12345].Debug()
	expected := "data=map[amount::${SPACE_FORMAT}1000000.00"
	helpers.TestStringContains(t, debugString, expected, "contain parsed amount")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

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
	bc.AddBotAndStart(bot)

	bc.commandDeleteTransactions(&MockContext{M: &tb.Message{Chat: chat, Text: "/deleteAll"}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "to confirm the deletion of your transactions") {
		t.Errorf("Deletion should require 'yes' confirmation. Got: %s", bot.LastSentWhat)
	}

	bc.commandDeleteTransactions(&MockContext{M: &tb.Message{Chat: chat, Text: "/deleteAll YeS"}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "Permanently deleted all your transactions") {
		t.Errorf("Deletion should work with confirmation. Got: %s", bot.LastSentWhat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTransactionListMaxLength(t *testing.T) {
	// create test dependencies
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	crud.TEST_MODE = true
	if err != nil {
		log.Fatal(err)
	}
	mock.
		ExpectQuery(`SELECT "id", "value", "created" FROM "bot::transaction"`).
		WithArgs(chat.ID, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "value", "created"}).AddRow(123, strings.Repeat("**********", 100), "").AddRow(124, strings.Repeat("**********", 100), "")) // 1000 + 1000
	mock.
		ExpectQuery(`SELECT "id", "value", "created" FROM "bot::transaction"`).
		WithArgs(chat.ID, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "value", "created"}).
			// 5 * 1000
			AddRow(123, strings.Repeat("**********", 100), "").
			AddRow(124, strings.Repeat("**********", 100), "").
			AddRow(125, strings.Repeat("**********", 100), "").
			AddRow(126, strings.Repeat("**********", 100), "").
			AddRow(127, strings.Repeat("**********", 100), ""),
		)

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	// < 4096 chars tx
	bc.commandList(&MockContext{M: &tb.Message{Chat: chat}})
	if len(bot.AllLastSentWhat) != 1 {
		t.Errorf("Expected exactly one message to be sent out: %v", bot.AllLastSentWhat)
	}

	bot.reset()

	// > 4096 chars tx
	bc.commandList(&MockContext{M: &tb.Message{Chat: chat}})
	if len(bot.AllLastSentWhat) != 2 {
		t.Errorf("Expected exactly two messages to be sent out: %v", strings.Join(stringArr(bot.AllLastSentWhat), ", "))
	}
	if bot.LastSentWhat != strings.Repeat("**********", 100) {
		t.Errorf("Expected last message to contain last transaction as it flowed over the first message: %v", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTransactionsListArchivedDated(t *testing.T) {
	// create test dependencies
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	crud.TEST_MODE = true
	if err != nil {
		log.Fatal(err)
	}
	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	// successful date enrichment
	mock.ExpectQuery(`SELECT "id", "value", "created" FROM "bot::transaction"`).WithArgs(12345, true).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "value", "created"}).
				AddRow(123, "tx1", "2022-03-30T14:24:50.390084Z").
				AddRow(124, "tx2", "2022-03-30T15:24:50.390084Z"),
		)
	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TZOFF).WillReturnRows(mock.NewRows([]string{"value"}))

	bc.commandList(&MockContext{M: &tb.Message{Chat: chat, Text: "/testListCommand(ignored) archived dated"}})

	if bot.LastSentWhat != "; recorded on 2022-03-30 14:24\ntx1\n; recorded on 2022-03-30 15:24\ntx2" {
		t.Errorf("Expected last message to contain transactions:\n%v", bot.LastSentWhat)
	}

	// fall back to undated if date parsing fails
	mock.ExpectQuery(`SELECT "id", "value", "created" FROM "bot::transaction"`).WithArgs(12345, true).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "value", "created"}).
				AddRow(123, "tx1", "123456789").
				AddRow(124, "tx2", "456789123"),
		)
	mock.ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).WithArgs(12345, helpers.USERSET_TZOFF).WillReturnRows(mock.NewRows([]string{"value"}))

	bc.commandList(&MockContext{M: &tb.Message{Chat: chat, Text: "/testListCommand(ignored) archived dated"}})

	if bot.LastSentWhat != "tx1\ntx2" {
		t.Errorf("Expected last message to contain transactions:\n%v", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestWritingComment(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	mock.
		ExpectExec(`INSERT INTO "bot::transaction"`).
		WithArgs(chat.ID, "; This is a comment"+"\n").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandAddComment(&MockContext{M: &tb.Message{Chat: chat, Text: "/comment \"; This is a comment\""}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "added the comment") {
		t.Errorf("Adding comment should have worked. Got message: %s", bot.LastSentWhat)
	}

	// Comment does not require quotes, as it only has a single parameter
	mock.
		ExpectExec(`INSERT INTO "bot::transaction"`).
		WithArgs(chat.ID, "This is another comment without \" (quotes)"+"\n").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc.commandAddComment(&MockContext{M: &tb.Message{Chat: chat, Text: "/c This is another comment without \\\" (quotes)"}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "added the comment") {
		t.Errorf("Adding comment should have worked. Got message: %s", bot.LastSentWhat)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func stringArr(i []interface{}) []string {
	arr := []string{}
	for _, e := range i {
		arr = append(arr, fmt.Sprintf("%v", e))
	}
	return arr
}

func TestCommandStartHelp(t *testing.T) {
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_ADM).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(false))
	bc.commandStart(&MockContext{M: &tb.Message{Chat: chat}})

	if !strings.Contains(fmt.Sprintf("%v", bot.AllLastSentWhat[0]), "Welcome") {
		t.Errorf("Bot should welcome user first")
	}
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "/help - List this command help") {
		t.Errorf("Bot should send help message as well")
	}
	if strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "admin_") {
		t.Errorf("Bot should not send admin commands in help message for default user")
	}

	// Admin check
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_ADM).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(true))
	bc.commandHelp(&MockContext{M: &tb.Message{Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "admin_") {
		t.Errorf("Bot should send admin commands in help message for admin user")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCommandCancel(t *testing.T) {
	chat := &tb.Chat{ID: 12345}
	bc := NewBotController(nil)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)
	bc.commandCancel(&MockContext{M: &tb.Message{Chat: chat}})
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "did not currently have any state or transaction open that could be cancelled") {
		t.Errorf("Unexpectedly there were open tx before")
	}
}

func TestTimezoneOffsetForAutomaticDate(t *testing.T) {
	// create test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CURRENCY"))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_CUR).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("TEST_CURRENCY"))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TAG).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("vacation2021"))
	yesterday_tzCorrection := time.Now().Add(-24 * time.Hour).Format(helpers.BEANCOUNT_DATE_FORMAT)
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow("-24"))
	mock.
		ExpectExec(`INSERT INTO "bot::transaction"`).
		WithArgs(chat.ID, yesterday_tzCorrection+` * "Buy something in the grocery store" #vacation2021
  Assets:Wallet                               -17.34 TEST_CURRENCY
  Expenses:Groceries
`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	// Create simple tx and fill it completely
	bc.commandCreateSimpleTx(&MockContext{M: &tb.Message{Chat: chat}})
	tx := bc.State.txStates[12345]
	tx.Input(&tb.Message{Text: "17.34"})                                                  // amount
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"})                     // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                                          // from
	bc.handleTextState(&MockContext{&tb.Message{Chat: chat, Text: "Expenses:Groceries"}}) // to (via handleTextState)

	// After the first tx is done, send some command
	m := &MockContext{M: &tb.Message{Chat: chat, Sender: &tb.User{ID: chat.ID}}}
	bc.handleTextState(m)

	// should catch and send help instead of fail
	if !strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "you might need to start a transaction first") {
		t.Errorf("String did not contain substring as expected (was: '%s')", bot.LastSentWhat)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
