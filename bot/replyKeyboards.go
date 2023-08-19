package bot

import (
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

func ReplyKeyboard(buttons []string) *tb.ReplyMarkup {
	if len(buttons) == 0 {
		return clearKeyboard()
	}
	kb := &tb.ReplyMarkup{ResizeKeyboard: true, OneTimeKeyboard: true}
	buttonsCreated := []tb.Row{}
	for _, label := range buttons {
		buttonsCreated = append(buttonsCreated, kb.Row(kb.Text(label)))
	}
	if len(buttonsCreated) > helpers.MAX_REPLY_KEYBOARD_ENTRIES {
		buttonsCreated = buttonsCreated[:helpers.MAX_REPLY_KEYBOARD_ENTRIES]
	}
	kb.Reply(buttonsCreated...)
	return kb
}
