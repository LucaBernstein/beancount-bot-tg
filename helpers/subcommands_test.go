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
