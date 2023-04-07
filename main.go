package main

import (
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/web"
)

func main() {
	db := db.Connection()
	defer db.Close()

	bc := bot.NewBotController(db)
	bc.ConfigureCronScheduler()

	go web.StartWebServer(bc)

	bot := bot.CreateBot(bc)
	bc.AddBotAndStart(bot)
}
