package bot_test

import (
	"strings"
	"testing"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func TestHandleFloat(t *testing.T) {
	_, err := bot.HandleFloat(&tb.Message{Text: "Hello World!"})
	if err == nil {
		t.Errorf("Expected float parsing to return error.")
	}

	handledFloat, err := bot.HandleFloat(&tb.Message{Text: "27.5"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27.5")
	helpers.TestExpect(t, handledFloat, "27.5", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "27,8"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,8")
	helpers.TestExpect(t, handledFloat, "27.8", "Should come out as clean float")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "  27,12  "})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,12")
	helpers.TestExpect(t, handledFloat, "27.12", "Should come out as clean float (2)")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1.23456"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 1.23456")
	helpers.TestExpect(t, handledFloat, "1.23456", "Should work for precise floats")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "4.44 USD_CUSTOM"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 4.44 USD_CUSTOM")
	helpers.TestExpect(t, handledFloat, "4.44 USD_CUSTOM", "Should include custom currency")
}

func TestTransactionBuilding(t *testing.T) {
	tx, err := bot.CreateSimpleTx(&tb.Message{Text: "/simple"}, "")
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17"})                                 // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("USD", "")
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	today := time.Now().Format(helpers.BEANCOUNT_DATE_FORMAT)
	helpers.TestExpect(t, templated, today+` * "Buy something in the grocery store"
  Assets:Wallet                               -17.00 USD
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestTransactionBuildingCustomCurrencyInAmount(t *testing.T) {
	tx, err := bot.CreateSimpleTx(&tb.Message{Text: "/simple"}, "")
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})                   // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("EUR", "")
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	today := time.Now().Format(helpers.BEANCOUNT_DATE_FORMAT)
	helpers.TestExpect(t, templated, today+` * "Buy something in the grocery store"
  Assets:Wallet                               -17.3456 USD_TEST
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestTransactionBuildingWithDate(t *testing.T) {
	tx, err := bot.CreateSimpleTx(&tb.Message{Text: "/simple 2021-01-24"}, "")
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})                   // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("EUR", "")
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	helpers.TestExpect(t, templated, `2021-01-24 * "Buy something in the grocery store"
  Assets:Wallet                               -17.3456 USD_TEST
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestCountLeadingDigits(t *testing.T) {
	helpers.TestExpect(t, bot.CountLeadingDigits(12.34), 2, "")
	helpers.TestExpect(t, bot.CountLeadingDigits(0.34), 1, "")
	helpers.TestExpect(t, bot.CountLeadingDigits(1244.0), 4, "")
}

func TestTaggedTransaction(t *testing.T) {
	tx, _ := bot.CreateSimpleTx(&tb.Message{Text: "/simple 2021-01-24"}, "")
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})   // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"}) // to
	tx.Input(&tb.Message{Text: "Buy something"})      // description
	template, err := tx.FillTemplate("EUR", "someTag")
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if !strings.Contains(template, `2021-01-24 * "Buy something" #someTag`) {
		t.Errorf("Tx did not contain tag: %s", template)
	}
}

func TestParseAmount(t *testing.T) {
	helpers.TestExpect(t, bot.ParseAmount(-1), "-1.00", "At least two decimal places should be present")
	helpers.TestExpect(t, bot.ParseAmount(0), "0.00", "At least two decimal places should be present")
	helpers.TestExpect(t, bot.ParseAmount(17), "17.00", "At least two decimal places should be present")
	helpers.TestExpect(t, bot.ParseAmount(16.8), "16.80", "At least two decimal places should be present")
	helpers.TestExpect(t, bot.ParseAmount(9.8), "9.80", "At least two decimal places should be present")

	helpers.TestExpect(t, bot.ParseAmount(9.801), "9.801", "If higher precision is given, that should be applied")
	helpers.TestExpect(t, bot.ParseAmount(17.3456), "17.3456", "If higher precision is given, that should be applied")
}
