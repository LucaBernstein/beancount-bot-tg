package main

import (
	"github.com/LucaBernstein/beancount-bot-tg/api"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
)

func main() {
	db := db.Connection()
	defer db.Close()

	bc := bot.NewBotController(db)
	bc.ConfigureCronScheduler()

	go api.StartWebServer(bc)

	bot := bot.CreateBot(bc)
	bc.AddBotAndStart(bot)
}
