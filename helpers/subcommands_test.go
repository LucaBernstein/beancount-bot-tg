package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

type subcommandTestState struct {
	lastCalledParams []string
}

func (state *subcommandTestState) record(m *tb.Message, params ...string) {
	state.lastCalledParams = params
}

func (state *subcommandTestState) testCase(t *testing.T, sh *helpers.SubcommandHandler, msg string, shouldErr bool, expected []string) {
	err := sh.Handle(&tb.Message{Text: msg})
	should := "should have occurred"
	if !shouldErr {
		should = "should not occur"
	}
	// If err but no err should occur: fail, check also next
	// If !err but err should occur: fail and return
	if shouldErr {
		if err == nil {
			t.Errorf("%s: Error %s: %e", msg, should, err)
		}
		return
	}
	if !shouldErr && err != nil {
		t.Errorf("%s: Error %s: %e", msg, should, err)
	}
	if !helpers.ArraysEqual(state.lastCalledParams, expected) {
		t.Errorf("%s:\n expected: %v\n got: %v", msg, expected, state.lastCalledParams)
		return
	}
}

func TestSubcommandHelper(t *testing.T) {
	state := &subcommandTestState{}

	sh := helpers.MakeSubcommandHandler("base", true)
	sh.Add("subcommand", state.record)

	state.testCase(t, sh, "base subcommand", false, []string{})
	state.testCase(t, sh, "base subcommand singlearg", false, []string{"singlearg"})
	state.testCase(t, sh, "base subcommand multi arg", false, []string{"multi", "arg"})
	state.testCase(t, sh, "base subcommand \"single arg\" multi arg", false, []string{"single arg", "multi", "arg"})
	state.testCase(t, sh, "base notregistered subcommand", true, nil)
}

func TestSplitQuotedCommand(t *testing.T) {
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`something`), []string{"something"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command`), []string{"/command"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`  /command  `), []string{"/command"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`  /command  anotherone `), []string{"/command", "anotherone"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command "anotherone"`), []string{"/command", "anotherone"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command "is quoted" non-quoted and sep`), []string{"/command", "is quoted", "non-quoted", "and", "sep"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command hello"world"`), []string{"/command", "helloworld"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command hello "world"`), []string{"/command", "hello", "world"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command hello" world"`), []string{"/command", "hello world"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command "wrongly quoted`), []string{}, "quotes opened but not closed")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command \"correctly quoted`), []string{"/command", "\"correctly", "quoted"}, "quotes opened but not closed")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`/command hello\ world`), []string{"/command", "hello world"}, "")
	helpers.TestExpectArrEq(t, helpers.SplitQuotedCommand(`"onlyquoted"`), []string{"onlyquoted"}, "")
}

func TestSubcommandAddingWarnings(t *testing.T) {
	sh := helpers.MakeSubcommandHandler("base", true)
	result := 0
	sh.Add("subcommand", func(m *tb.Message, params ...string) { result++ })
	sh.Add("subcommand", func(m *tb.Message, params ...string) { result += 2 }) // again though already exists
	err := sh.Handle(&tb.Message{Text: "base subcommand"})
	if err != nil {
		t.Errorf("No error should have been returned: %s", err.Error())
	}
	if result != 2 {
		t.Errorf("Second mapping should take precedence. Expected %d to be 2.", result)
	}

	sh.Add("subcommand with spaces", func(m *tb.Message, params ...string) { result++ }) // for coverage

	err = sh.Handle(&tb.Message{Text: "base subcommand with \"invalid quoting"})
	if err == nil {
		t.Errorf("Should return error for handling with invalid quoting")
	}
}

func TestExtractTypeValue(t *testing.T) {
	tv, err := helpers.ExtractTypeValue("onlyType")
	if err != nil {
		t.Errorf("No error should have been returned")
	}
	if tv.T != "onlyType" || tv.Value != "" {
		t.Errorf("Only key should be set")
	}

	tv, err = helpers.ExtractTypeValue("aKey", "and_a_value")
	if err != nil {
		t.Errorf("No error should have been returned")
	}
	if tv.T != "aKey" || tv.Value != "and_a_value" {
		t.Errorf("Only key should be set")
	}

	_, err = helpers.ExtractTypeValue()
	if err == nil {
		t.Errorf("Should return error for no params")
	}

	// Error case: Too many params
	_, err = helpers.ExtractTypeValue("", "", "")
	if err == nil {
		t.Errorf("Should return error for too many params")
	}
}
