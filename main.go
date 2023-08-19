package main

import (
	"github.com/LucaBernstein/beancount-bot-tg/v2/api"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db"
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
