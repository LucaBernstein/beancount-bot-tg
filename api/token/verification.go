package token

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/gin-gonic/gin"
)

func (r *Router) Verification(c *gin.Context) {
	pUserId := c.Param("userId")
	userId, err := strconv.ParseInt(pUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	nonce, err := r.bc.Repo.CreateApiVerification(userId)
	if err != nil {
		var httpStatus int
		switch err {
		case helpers.ErrApiDisabled:
			httpStatus = http.StatusExpectationFailed
		case helpers.ErrApiTokenChallengeInProgress:
			httpStatus = http.StatusConflict
		case helpers.ErrApiInvalidTokenVerification:
			httpStatus = http.StatusBadRequest
		default:
			httpStatus = http.StatusInternalServerError
		}
		c.JSON(httpStatus, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Send nonce to bot user chat
	r.bc.Bot.SendSilent(r.bc.Logf, bot.ReceiverImpl{ChatId: pUserId}, fmt.Sprintf(
		"To verify your API token creation attempt, please use this number: %s", nonce))

	c.String(http.StatusCreated, "Successfully created token. Please check for the verification message by the Telegram bot.")
}
