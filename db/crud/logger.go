package crud

import (
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func LogDbf(r *Repo, level helpers.Level, m *tb.Message, format string, v ...interface{}) {
	helpers.LogLocalf(level, m, format, v...)
	// TODO: Implement db logging
}
