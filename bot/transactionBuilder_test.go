package bot_test

import (
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
	tx := bot.CreateSimpleTx()
	tx.Input(&tb.Message{Text: "17"})                                 // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "2021-01-24"})                         // date

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("USD")
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	helpers.TestExpect(t, templated, `2021-01-24 * "Buy something in the grocery store"
  Assets:Wallet                               -17.00 USD
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestTransactionBuildingCustomCurrencyInAmount(t *testing.T) {
	tx := bot.CreateSimpleTx()
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})                   // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "2021-01-24"})                         // date

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("EUR")
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

func TestDateSpecialNow(t *testing.T) {
	h, err := bot.HandleDate(&tb.Message{Text: " ToDaY "})
	if err != nil {
		t.Errorf("There should be no error handling date 'today': %s", err.Error())
	}
	helpers.TestExpect(t, h, time.Now().Format("2006-01-02"), "")

	// GitHub-Issue #14: Also accept parts of "today" string for quicker recording
	h, err = bot.HandleDate(&tb.Message{Text: " t"})
	if err != nil {
		t.Errorf("There should be no error handling date 't': %s", err.Error())
	}
	helpers.TestExpect(t, h, time.Now().Format("2006-01-02"), "")

	h, err = bot.HandleDate(&tb.Message{Text: " tod"})
	if err != nil {
		t.Errorf("There should be no error handling date 'tod': %s", err.Error())
	}
	helpers.TestExpect(t, h, time.Now().Format("2006-01-02"), "")

	_, err = bot.HandleDate(&tb.Message{Text: "tomorrow"})
	if err == nil {
		t.Errorf("There should be an error processing the date 'tomorrow': %s", err.Error())
	}
}
