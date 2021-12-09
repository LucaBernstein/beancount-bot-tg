package crud

import (
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func LogDbf(r *Repo, level helpers.Level, m *tb.Message, format string, v ...interface{}) {
	prefix, message := helpers.LogLocalf(level, m, format, v...)
	go logToDb(r, prefix, level, message)
}

func logToDb(r *Repo, chat string, level helpers.Level, message string) {
	values := []interface{}{int(level), message}
	if chat != "" {
		values = append(values, chat)
	} else {
		values = append(values, nil)
	}
	_, err := r.db.Exec(`INSERT INTO "app::log" ("level", "message", "chat") VALUES ($1, $2, $3)`, values...)
	if err != nil {
		helpers.LogLocalf(helpers.ERROR, nil, "Error inserting log statement into db: %s", err.Error())
	}
}
