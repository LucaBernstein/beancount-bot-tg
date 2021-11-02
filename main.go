package main

import (
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
)

func main() {
	db := db.PostgresConnection()
	defer db.Close()

	bc := bot.NewBotController(db)

	bot := bot.CreateBot(bc)
	bc.ConfigureAndAttachBot(bot)
}
