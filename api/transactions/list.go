package transactions

import (
	"net/http"
	"strconv"

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
	queryArchived := c.Query("archived")
	if queryArchived == "" {
		queryArchived = "false"
	}
	isArchived, err := strconv.ParseBool(queryArchived)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	chatId := c.GetInt64(helpers.K_CHAT_ID)
	m := &telebot.Message{Chat: &telebot.Chat{ID: chatId}}
	tx, err := r.bc.Repo.GetTransactions(m, isArchived)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	transactions := []Transaction{}
	for _, t := range tx {
		transactions = append(transactions, Transaction{
			Id:         t.Id,
			CreatedAt:  t.Date,
			Booking:    t.Tx,
			IsArchived: isArchived,
		})
	}
	c.JSON(http.StatusOK, transactions)
}

func (r *Router) ListDeleteAll(c *gin.Context) {
	chatId := c.GetInt64(helpers.K_CHAT_ID)
	m := &telebot.Message{Chat: &telebot.Chat{ID: chatId}}
	count, err := r.bc.Repo.DeleteTransactions(m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"affected": count,
	})
}

func (r *Router) ListDeleteSingle(c *gin.Context) {
	queryArchived := c.Query("archived")
	if queryArchived == "" {
		queryArchived = "false"
	}
	isArchived, err := strconv.ParseBool(queryArchived)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	queryElementId := c.Param("id")
	if queryElementId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No element to delete provided",
		})
		return
	}
	elId, err := strconv.Atoi(queryElementId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No element to delete provided",
		})
		return
	}
	chatId := c.GetInt64(helpers.K_CHAT_ID)
	m := &telebot.Message{Chat: &telebot.Chat{ID: chatId}}
	count, err := r.bc.Repo.DeleteTransaction(m, isArchived, elId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"affected": count,
	})
}
