package token

import (
	"net/http"
	"strconv"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-gonic/gin"
)

func (r *Router) Grant(c *gin.Context) {
	pUserId := c.Param("userId")
	nonce := c.Param("nonce")
	userId, err := strconv.ParseInt(pUserId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := r.bc.Repo.VerifyApiToken(userId, nonce)
	if err != nil {
		var httpStatus int
		switch err {
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

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
