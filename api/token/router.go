package token

import (
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
	g.POST("/verification/:userId", r.Verification)
	g.POST("/grant/:userId/:nonce", r.Grant)
	g.POST("/revoke/:token", r.Revoke)
}
