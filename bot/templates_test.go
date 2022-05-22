package bot

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestTemplateHelpForNonexistingTemplate(t *testing.T) {
	// test dependencies
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
		ExpectQuery(regexp.QuoteMeta(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = $1 AND "name" LIKE $2`)).
		WithArgs(12345, "notexist%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t notexist"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /template", "send help for invalid command")

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t add nonesense"})
	bc.handleTextState(&tb.Message{Chat: chat, Text: "Blah"})

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "Usage help for /template", "send help for invalid command")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTemplateList(t *testing.T) {
	// test dependencies
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
		ExpectQuery(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = ?`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t list"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "not created any template yet", "no template yet message")

	mock.
		ExpectQuery(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = ?`).
		WithArgs(12345).
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}).AddRow("test1", "tpl1").AddRow("test2", "tpl2"))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t list"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), `test1:
tpl1

test2:
tpl2`, "templates")

	mock.
		ExpectQuery(regexp.QuoteMeta(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = $1 AND "name" LIKE $2`)).
		WithArgs(12345, "test3%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t list test3"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "No template name matched your query 'test3'", "single template response")

	mock.
		ExpectQuery(regexp.QuoteMeta(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = $1 AND "name" LIKE $2`)).
		WithArgs(12345, "test2%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}).AddRow("test2", "tpl2"))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t list test2"})
	helpers.TestExpect(t, strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "tpl1"), false, "should not contain tpl1")
	helpers.TestExpect(t, strings.Contains(fmt.Sprintf("%v", bot.LastSentWhat), "tpl2"), true, "but should contain tpl2")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTemplateAdd(t *testing.T) {
	// test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t add my template"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "parameter count mismatch", "parameter count mismatch response")

	// Step 1: Start template creation
	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t add myTemplate"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "Please provide a full transaction template", "template creation process response")
	helpers.TestExpect(t, bc.State.states[chatId(chat.ID)], ST_TPL, "state should show template process")
	helpers.TestExpect(t, bc.State.tplStates[chatId(chat.ID)], TemplateName("myTemplate"), "state should save template name")

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "bot::template" ("tgChatId", "name", "template") VALUES ($1, $2, $3)`)).
		WithArgs(12345, "myTemplate", "template data").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Step 2: Send template
	bc.handleTextState(&tb.Message{Chat: chat, Text: "template data"})

	helpers.TestExpect(t, bc.State.states[chatId(chat.ID)], ST_NONE, "state should be clean again")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTemplateRm(t *testing.T) {
	// test dependencies
	crud.TEST_MODE = true
	chat := &tb.Chat{ID: 12345}
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	bc := NewBotController(db)
	bot := &MockBot{}
	bc.AddBotAndStart(bot)

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "bot::template" WHERE "tgChatId" = $1 AND "name" = $2`)).
		WithArgs(12345, "myTemplate").
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/template rm myTemplate"}) // also test for /template instead of short alias /t
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "Successfully removed your template 'myTemplate'", "removed template")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestTemplateUse(t *testing.T) {
	// test dependencies
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
		ExpectQuery(regexp.QuoteMeta(`SELECT "name", "template" FROM "bot::template" WHERE "tgChatId" = $1 AND "name" LIKE $2`)).
		WithArgs(12345, "test%").
		WillReturnRows(sqlmock.NewRows([]string{"name", "template"}).AddRow("test", `${date} * "Test" "${description}"
  fromFix ${-amount}
  toFix1 ${amount/2}
  toFix2 ${amount/2}`))

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
		WillReturnRows(sqlmock.NewRows([]string{"value"}))
	mock.
		ExpectQuery(`SELECT "value" FROM "bot::userSetting"`).
		WithArgs(chat.ID, helpers.USERSET_TZOFF).
		WillReturnRows(sqlmock.NewRows([]string{"value"}))

	mock.
		ExpectExec(regexp.QuoteMeta(`INSERT INTO "bot::transaction" ("tgChatId", "value")
		VALUES ($1, $2);`)).
		WithArgs(chat.ID, `2022-04-11 * "Test" "Buy something"
  fromFix                                     -10.51 EUR_TEST
  toFix1                                        5.255 EUR_TEST
  toFix2                                        5.255 EUR_TEST
`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bc.commandTemplates(&tb.Message{Chat: chat, Text: "/t test 2022-04-11"})
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.AllLastSentWhat[len(bot.AllLastSentWhat)-2]), "Creating a new transaction from your template 'test'", "template tx starting msg")
	helpers.TestStringContains(t, fmt.Sprintf("%v", bot.LastSentWhat), "amount", "asking for amount")

	tx := bc.State.txStates[chatId(chat.ID)]
	tx.Input(&tb.Message{Text: "10.51 EUR_TEST"})                      // amount
	bc.handleTextState(&tb.Message{Chat: chat, Text: "Buy something"}) // description (via handleTextState)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
