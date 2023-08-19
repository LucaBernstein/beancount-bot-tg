package api

import (
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/v2/api/admin"
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/config"
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/health"
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/suggestions"
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/token"
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/transactions"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/gin-gonic/gin"
	"github.com/mandrigin/gin-spa/spa"
)

func StartWebServer(bc *bot.BotController) {
	r := gin.Default()
	r.Use(gin.Recovery())
	configureCors(r)

	r.GET("/health", gin.BasicAuth(gin.Accounts{
		helpers.EnvOrFb("MONITORING_USER", "beancount-bot-tg-health"): helpers.EnvOrFb("MONITORING_PASS", "this_service_should_be_healthy"),
	}), health.MonitoringEndpoint(bc))

	apiGroup := r.Group("/api")

	tokenGroup := apiGroup.Group("/token")
	token.NewRouter(bc).Hook(tokenGroup)

	transactionGroup := apiGroup.Group("/transactions")
	transactions.NewRouter(bc).Hook(transactionGroup)

	suggestionsGroup := apiGroup.Group("/suggestions")
	suggestions.NewRouter(bc).Hook(suggestionsGroup)

	configGroup := apiGroup.Group("/config")
	config.NewRouter(bc).Hook(configGroup)

	adminGroup := apiGroup.Group("/admin")
	admin.NewRouter(bc).Hook(adminGroup)

	r.Use(spa.Middleware("/", "./api/ui/build/web")) // Needs to come last

	port := ":8080"
	log.Printf("Web server started on %s", port)
	r.Run(port)
}
