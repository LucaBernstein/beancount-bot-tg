package web

import (
	"log"
	"net/http"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/web/health"
)

func StartWebServer(bc *bot.BotController) {
	http.HandleFunc("/health", health.MonitoringEndpoint(bc))
	port := ":8081"
	log.Printf("Web server started on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
