package helpers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/gin-gonic/gin"
)

const K_CHAT_ID = "tgChatId"

func AttachChatId(bc *bot.BotController) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Authentication header: %v", c.GetHeader("Authorization"))
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
