package api

import (
	"log"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func configureCors(router *gin.Engine) {
	if !strings.HasPrefix(helpers.Env("BASE_URI"), "https") {
		log.Print("Configuring CORS to allow all origins, as BASE_URI env var does not have prefix 'https'.")
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOriginFunc = func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost")
		}
		corsConfig.AllowHeaders = []string{"authorization", "content-type"}
		corsConfig.AllowCredentials = true
		router.Use(cors.New(corsConfig))
	}
}
