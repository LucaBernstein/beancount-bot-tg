package bot_test

import (
	"strings"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
)

func TestReplyKeyboardMaximumOptions(t *testing.T) {
	options := []string{"some", "few", "options"}
	reply := bot.ReplyKeyboard(options)
	helpers.TestExpect(t, len(reply.ReplyKeyboard), 3, "reply keyboard options count")

	options = strings.Split(strings.Repeat("more_options ", 50), " ")
	reply = bot.ReplyKeyboard(options)
	helpers.TestExpect(t, len(reply.ReplyKeyboard), helpers.MAX_REPLY_KEYBOARD_ENTRIES, "reply keyboard options count")
}
