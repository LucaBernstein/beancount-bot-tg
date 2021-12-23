package web

import (
	"log"
	"net/http"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/web/health"
)

func StartWebServer(bc *bot.BotController) {
	http.HandleFunc("/health", health.MonitoringEndpoint(bc))
	log.Fatal(http.ListenAndServe(":8081", nil))
}
