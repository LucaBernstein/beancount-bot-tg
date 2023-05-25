package suggestions

import (
	"net/http"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/api/helpers"
	"github.com/gin-gonic/gin"
	"gopkg.in/telebot.v3"
)

type Transaction struct {
	Id         int    `json:"id"`
	CreatedAt  string `json:"createdAt"`
	Booking    string `json:"booking"`
	IsArchived bool   `json:"isArchived"`
}

func (r *Router) List(c *gin.Context) {
	chatId := c.GetInt64(helpers.K_CHAT_ID)
	m := &telebot.Message{Chat: &telebot.Chat{ID: chatId}}
	suggestions, err := r.bc.Repo.GetAllSuggestions(m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, suggestions)
}

func (r *Router) ListDelete(c *gin.Context) {
	entryType := c.Param("type")
	entryName := c.Param("name")
	entryName = strings.TrimPrefix(entryName, "/")
	chatId := c.GetInt64(helpers.K_CHAT_ID)
	m := &telebot.Message{Chat: &telebot.Chat{ID: chatId}}
	count, err := r.bc.Repo.DeleteCacheEntries(m, entryType, entryName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	affected, err := count.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"affected": affected,
	})
}
