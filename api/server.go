package api

import (
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/api/token"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/gin-gonic/gin"
)

func StartWebServer(bc *bot.BotController) {
	r := gin.Default()
	r.Use(gin.Recovery())
	configureCors(r)

	apiGroup := r.Group("/api")

	tokenGroup := apiGroup.Group("/token")
	token.NewRouter(bc).Hook(tokenGroup)

	port := ":80"
	log.Printf("Web server started on %s", port)
	r.Run(port)
}
