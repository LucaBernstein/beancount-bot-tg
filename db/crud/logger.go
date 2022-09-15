package crud

import (
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

var TEST_MODE = false

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
	if !TEST_MODE {
		_, err := r.db.Exec(`INSERT INTO "app::log" ("level", "message", "chat") VALUES ($1, $2, $3)`, values...)
		if err != nil {
			helpers.LogLocalf(helpers.ERROR, nil, "Error inserting log statement into db: %s", err.Error())
			log.Fatal("Logging to database was not possible.")
		}
	} else {
		log.Printf("DB LOGGER IS IN TEST MODE")
	}
}
