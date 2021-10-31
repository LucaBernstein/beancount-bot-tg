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
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,8")
	helpers.TestExpect(t, handledFloat, "27.12", "Should come out as clean float (2)")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1.23456"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,8")
	helpers.TestExpect(t, handledFloat, "1.23456", "Should work for precise floats")
}

func TestTransactionBuilding(t *testing.T) {
	tx := bot.CreateSimpleTx()
	tx.Input(&tb.Message{Text: "17.34"})                              // amount
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "2021-01-24"})                         // date

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate()
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	expected := "; Created by beancount-bot-tg on " + time.Now().Format("2006-01-02")
	helpers.TestExpect(t, templated, expected+`
2021-01-24 * "Buy something in the grocery store"
  Assets:Wallet                               -17.34 EUR
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestCountLeadingDigits(t *testing.T) {
	helpers.TestExpect(t, bot.CountLeadingDigits(12.34), 2, "")
	helpers.TestExpect(t, bot.CountLeadingDigits(0.34), 1, "")
	helpers.TestExpect(t, bot.CountLeadingDigits(1244.0), 4, "")
}

func TestDateSpecialNow(t *testing.T) {
	h, err := bot.HandleDate(&tb.Message{Text: "Now"})
	if err != nil {
		t.Errorf("There should be no error handling date 'now': %s", err.Error())
	}
	helpers.TestExpect(t, h, time.Now().Format("2006-01-02"), "")

	h, err = bot.HandleDate(&tb.Message{Text: " ToDaY "})
	if err != nil {
		t.Errorf("There should be no error handling date 'now': %s", err.Error())
	}
	helpers.TestExpect(t, h, time.Now().Format("2006-01-02"), "")
}
