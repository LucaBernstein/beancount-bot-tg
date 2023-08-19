package suggestions

import (
	"github.com/LucaBernstein/beancount-bot-tg/v2/api/helpers"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
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

	g.GET("/list", r.List)
	g.DELETE("/list/:type/*name", r.ListDelete)
}
