package bot

import (
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

const (
	TRACE = helpers.TRACE
	DEBUG = helpers.DEBUG
	INFO  = helpers.INFO
	WARN  = helpers.WARN
	ERROR = helpers.ERROR
	FATAL = helpers.FATAL
)

func (bc *BotController) Logf(level helpers.Level, m *tb.Message, format string, v ...interface{}) {
	crud.LogDbf(bc.Repo, level, m, format, v...)
}
