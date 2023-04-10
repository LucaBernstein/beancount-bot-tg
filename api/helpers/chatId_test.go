package helpers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/helpers"
	"github.com/LucaBernstein/beancount-bot-tg/api/helpers/apiTest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestChatIdEnrichment(t *testing.T) {
	token, mockBc, m := apiTest.MockBcApiUser(t, 442)
	handlerFn := helpers.AttachChatId(mockBc)

	r := gin.Default()
	r.Use(handlerFn)
	var tgChatId int64
	r.GET("/test", func(c *gin.Context) {
		tgChatId = c.GetInt64("tgChatId")
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, m.Chat.ID, tgChatId)
}
