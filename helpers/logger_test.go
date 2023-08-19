package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

func TestLogLocal(t *testing.T) {
	representation := helpers.DEBUG.String()
	if representation != "DEBUG" {
		t.Errorf("DEBUG string representation was not 'DEBUG', but was %s", representation)
	}

	var a helpers.Level = 8
	if a.String() != "8" {
		t.Errorf("Level '8' string representation was not '8', but was %s", a.String())
	}

	helpers.LogLocalf(helpers.TRACE, &tb.Message{Chat: &tb.Chat{ID: 12345}, Sender: &tb.User{ID: 12345}}, "This is a test log")
	helpers.LogLocalf(helpers.TRACE, &tb.Message{Chat: &tb.Chat{ID: 12345}}, "This is a test log without sender (historical tests)")
	helpers.LogLocalf(helpers.TRACE, &tb.Message{}, "No chat")
	helpers.LogLocalf(helpers.TRACE, nil, "No message at all")
}
