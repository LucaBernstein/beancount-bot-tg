package token

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *Router) Revoke(c *gin.Context) {
	token := c.Param("token")
	count, err := r.bc.Repo.RevokeApiToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"affected": count,
	})
}
