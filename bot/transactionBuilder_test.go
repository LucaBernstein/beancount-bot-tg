package bot_test

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

func TestTypesTemplateHintsOnlyContainsAllowed(t *testing.T) {
	var definedBot []string
	for key := range bot.TEMPLATE_TYPE_HINTS {
		definedBot = append(definedBot, string(key))
	}
	if len(helpers.AllowedSuggestionTypes()) > len(definedBot) {
		t.Errorf("Defined types mismatch (len)")
	}
	for _, key := range helpers.AllowedSuggestionTypes() {
		if !helpers.ArrayContains(definedBot, key) {
			t.Errorf("Defined types mismatch (undefined '%s')", key)
		}
	}
}

func TestHandleFloat(t *testing.T) {
	_, err := bot.HandleFloat(&tb.Message{Text: "Hello World!"})
	if err == nil {
		t.Errorf("Expected float parsing to return error.")
	}

	handledFloat, err := bot.HandleFloat(&tb.Message{Text: "27.5"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27.5")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"27.50", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "27,8"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,8")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"27.80", "Should come out as clean float")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "  27,12  "})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 27,12")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"27.12", "Should come out as clean float (2)")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1.23456"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 1.23456")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"1.23456", "Should work for precise floats")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "4.44 USD_CUSTOM"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 4.44 USD_CUSTOM")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"4.44 USD_CUSTOM", "Should include custom currency")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "-5.678"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for -5.678")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"5.678", "Should use absolute value")
}

func TestHandleFloatSimpleCalculations(t *testing.T) {
	// Additions should work
	handledFloat, err := bot.HandleFloat(&tb.Message{Text: "10+3"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 10+3")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"13.00", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "11.45+3,12345"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 11.45+3,12345")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"14.57345", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "006+9.999"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 006+9.999")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"15.999", "")

	// Multiplications should work
	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "10*3"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 10*3")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"30.00", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "10*3,12345"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 10*3,12345")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"31.2345", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "001.1*3.5"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 001.1*3.5")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"3.85", "")

	// Simple calculations also work with currencies
	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "11*3 TEST_CUR"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 11*3 TEST_CUR")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"33.00 TEST_CUR", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "14.5+16+1+1+3 ANOTHER_CURRENCY"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 14.5+16+1+1+3 ANOTHER_CURRENCY")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"35.50 ANOTHER_CURRENCY", "")

	// Check some error behaviors
	// Mixed calculation operators
	_, err = bot.HandleFloat(&tb.Message{Text: "1+1*2"})
	if err == nil || !strings.Contains(err.Error(), "failed at value '1*2'") {
		t.Errorf("Error message should state that mixing operators is not allowed: %e", err)
	}
	// Too many spaces in input
	_, err = bot.HandleFloat(&tb.Message{Text: "some many spaces"})
	if err == nil || !strings.Contains(err.Error(), "contained too many spaces") {
		t.Errorf("Error message should state that amount contained too many spaces")
	}
	// tx left open / spaced
	_, err = bot.HandleFloat(&tb.Message{Text: "1+ 1"})
	if err == nil || !strings.Contains(err.Error(), "additionally specified currency is allowed") {
		t.Errorf("Error message should state that no additionally specified currency is allowed for trailing + tx (left open)")
	}
	// Multiplications only work with exactly two multiplicators
	_, err = bot.HandleFloat(&tb.Message{Text: "1*1*1"})
	if err == nil || !strings.Contains(err.Error(), "exactly two multiplicators") {
		t.Errorf("Error message should state that parser expected exactly two multiplicators")
	}
	// some hiccup value in multiplication
	_, err = bot.HandleFloat(&tb.Message{Text: "1*EUR"})
	if err == nil || !strings.Contains(err.Error(), "failed at value 'EUR'") {
		t.Errorf("Error message should state that it could not interpret 'EUR' as a number: %s", err.Error())
	}
}

func TestHandleFloatThousandsSeparator(t *testing.T) {
	handledFloat, err := bot.HandleFloat(&tb.Message{Text: "100,000,000.00"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 100 million with comma thousands separator")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"100000000.00", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "100.000.000,00"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 100 million with dot thousands separator")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"100000000.00", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "24,123.7"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 24,123.7")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"24123.70", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1.000.000"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 1.000.000")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"1000000.00", "")

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1,000,000"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 1,000,000")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"1000000.00", "")

	_, err = bot.HandleFloat(&tb.Message{Text: "24,24.7"})
	if err == nil || !strings.Contains(err.Error(), "invalid separators in value '24,24.7'") {
		t.Errorf("Error message should state that separators used were invalid: %e", err)
	}

	_, err = bot.HandleFloat(&tb.Message{Text: "1.,23"})
	if err == nil || !strings.Contains(err.Error(), "invalid separators in value '1.,23'") {
		t.Errorf("Error message should state that separators used were invalid: %e", err)
	}

	_, err = bot.HandleFloat(&tb.Message{Text: "100,000.000,000"})
	if err == nil || !strings.Contains(err.Error(), "invalid separators in value '100,000.000,000'") {
		t.Errorf("Error message should state that separators used were invalid: %e", err)
	}

	_, err = bot.HandleFloat(&tb.Message{Text: "100,000.000,000.000"})
	if err == nil || !strings.Contains(err.Error(), "invalid separators in value '100,000.000,000.000'") {
		t.Errorf("Error message should state that separators used were invalid: %e", err)
	}

	handledFloat, err = bot.HandleFloat(&tb.Message{Text: "1,000.00+24"})
	helpers.TestExpect(t, err, nil, "Should not throw an error for 1,000.00+24")
	helpers.TestExpect(t, handledFloat, bot.FORMATTER_PLACEHOLDER+"1024.00", "")
}

func TestTransactionBuilding(t *testing.T) {
	tx, err := bot.CreateSimpleTx("", bot.TEMPLATE_SIMPLE_DEFAULT)
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17"})                                 // amount
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("USD", "", 0)
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
	tx, err := bot.CreateSimpleTx("", bot.TEMPLATE_SIMPLE_DEFAULT)
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})                   // amount
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("EUR", "", 0)
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
	tx, err := bot.CreateSimpleTx("", bot.TEMPLATE_SIMPLE_DEFAULT)
	tx.SetDate("2021-01-24")
	if err != nil {
		t.Errorf("Error creating simple tx: %s", err.Error())
	}
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})                   // amount
	tx.Input(&tb.Message{Text: "Buy something in the grocery store"}) // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})                      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"})                 // to

	if !tx.IsDone() {
		t.Errorf("With given input transaction data should be complete for SimpleTx")
	}

	templated, err := tx.FillTemplate("EUR", "", 0)
	if err != nil {
		t.Errorf("There should be no error raised during templating: %s", err.Error())
	}
	helpers.TestExpect(t, templated, `2021-01-24 * "Buy something in the grocery store"
  Assets:Wallet                               -17.3456 USD_TEST
  Expenses:Groceries
`, "Templated string should be filled with variables as expected.")
}

func TestTaggedTransaction(t *testing.T) {
	tx, _ := bot.CreateSimpleTx("", bot.TEMPLATE_SIMPLE_DEFAULT)
	tx.SetDate("2021-01-24")
	log.Print(tx.Debug())
	tx.Input(&tb.Message{Text: "17.3456 USD_TEST"})   // amount
	tx.Input(&tb.Message{Text: "Buy something"})      // description
	tx.Input(&tb.Message{Text: "Assets:Wallet"})      // from
	tx.Input(&tb.Message{Text: "Expenses:Groceries"}) // to
	template, err := tx.FillTemplate("EUR", "someTag", 0)
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

func TestParseTemplateFields(t *testing.T) {
	fields := bot.ParseTemplateFields(`this is a ${description} field, and ${-amount/3}`, "")

	helpers.TestExpect(t, fields[0].Raw, "-amount/3", "amount field raw")
	helpers.TestExpect(t, fields[0].FieldName, "amount", "amount field name")
	helpers.TestExpect(t, fields[0].IsNegative, true, "amount to be negative")
	helpers.TestExpect(t, fields[0].Fraction, 3, "amount fraction")

	helpers.TestExpect(t, fields[1].FieldName, "description", "description field name")
	helpers.TestExpect(t, fields[1].IsNegative, false, "description not be negative")
	helpers.TestExpect(t, fields[1].Fraction, 1, "description fraction default = 1")
}

func dateCase(t *testing.T, given, expected string) {
	handledDate, err := bot.ParseDate(given)
	helpers.TestExpect(t, err, nil, fmt.Sprintf("Should not throw an error for %s", given))
	helpers.TestExpect(t, handledDate, expected, "")
}

func TestEnhancedDateParsing(t *testing.T) {
	// helpers.BEANCOUNT_DATE_FORMAT
	today := time.Now().UTC()
	dateCase(t, "1999-04-14", "1999-04-14")
	dateCase(t, "19990414", "1999-04-14")
	dateCase(t, "03-31", fmt.Sprintf("%s-03-31", today.Format("2006")))
	dateCase(t, "0331", fmt.Sprintf("%s-03-31", today.Format("2006")))
	dateCase(t, "16", fmt.Sprintf("%s-16", today.Format("2006-01")))
	dateCase(t, "16", fmt.Sprintf("%s-16", today.Format("2006-01")))

	if _, err := bot.ParseDate("04-31"); err == nil {
		t.Errorf("Expected error for 04-31")
	}

	if _, err := bot.ParseDate("32"); err == nil {
		t.Errorf("Expected error for 32")
	}

	if _, err := bot.ParseDate("-01"); err == nil {
		t.Errorf("Expected error for 32")
	}
}
