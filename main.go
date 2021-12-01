package main

import (
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
)

func main() {
	db := db.PostgresConnection()
	defer db.Close()

	bc := bot.NewBotController(db)
	bc.ConfigureCronScheduler()

	bot := bot.CreateBot(bc)
	bc.AddBotAndStart(bot)
}
