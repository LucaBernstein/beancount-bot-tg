package api

import (
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/api/admin"
	"github.com/LucaBernstein/beancount-bot-tg/api/config"
	"github.com/LucaBernstein/beancount-bot-tg/api/token"
	"github.com/LucaBernstein/beancount-bot-tg/api/transactions"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/gin-gonic/gin"
	"github.com/mandrigin/gin-spa/spa"
)

func StartWebServer(bc *bot.BotController) {
	r := gin.Default()
	r.Use(gin.Recovery())
	configureCors(r)

	apiGroup := r.Group("/api")

	tokenGroup := apiGroup.Group("/token")
	token.NewRouter(bc).Hook(tokenGroup)

	transactionGroup := apiGroup.Group("/transactions")
	transactions.NewRouter(bc).Hook(transactionGroup)

	configGroup := apiGroup.Group("/config")
	config.NewRouter(bc).Hook(configGroup)

	adminGroup := apiGroup.Group("/admin")
	admin.NewRouter(bc).Hook(adminGroup)

	r.Use(spa.Middleware("/", "./api/ui/build/web")) // Needs to come last

	port := ":8080"
	log.Printf("Web server started on %s", port)
	r.Run(port)
}
