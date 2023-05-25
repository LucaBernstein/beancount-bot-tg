package helpers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/gin-gonic/gin"
	"gopkg.in/telebot.v3"
)

const K_CHAT_ID = "tgChatId"

func AttachChatId(bc *bot.BotController) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", ""))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No 'Authorization' header set in request",
			})
			c.Abort()
			return
		}
		chatId, err := bc.Repo.GetTokenChatId(authHeader)
		if err != nil {
			var httpStatus int
			switch err {
			case sql.ErrNoRows:
				httpStatus = http.StatusUnauthorized
			default:
				httpStatus = http.StatusInternalServerError
			}
			c.JSON(httpStatus, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Set(K_CHAT_ID, chatId)
		c.Next()
	}
}

func EnsureAdmin(bc *bot.BotController) gin.HandlerFunc {
	return func(c *gin.Context) {
		tgChatId := c.GetInt64("tgChatId")
		m := &telebot.Message{Chat: &telebot.Chat{ID: tgChatId}}
		isAdmin := bc.Repo.UserIsAdmin(m)
		if !isAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "missing priviledges",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
