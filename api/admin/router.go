package admin

import (
	"github.com/LucaBernstein/beancount-bot-tg/api/helpers"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/gin-gonic/gin"
)

type Router struct {
	bc *bot.BotController
}

func NewRouter(bc *bot.BotController) *Router {
	return &Router{
		bc: bc,
	}
}

func (r *Router) Hook(g *gin.RouterGroup) {
	g.Use(helpers.AttachChatId(r.bc))
	g.Use(helpers.EnsureAdmin(r.bc))

	g.GET("/logs", r.Logs)
}
