package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

func ReplyKeyboard(buttons []string) *tb.ReplyMarkup {
	kb := &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	for _, label := range buttons {
		kb.Text(label)
	}
	return kb
}
