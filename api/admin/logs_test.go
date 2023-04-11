package admin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/admin"
	"github.com/LucaBernstein/beancount-bot-tg/api/helpers/apiTest"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogs(t *testing.T) {
	token, mockBc, msg := apiTest.MockBcApiUser(t, 919)
	r := gin.Default()
	g := r.Group("")
	admin.NewRouter(mockBc).Hook(g)

	// Should be forbidden without admin priviledges
	err := apiTest.PromoteAdmin(msg.Chat.ID, false)
	botTest.HandleErr(t, err)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/logs", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 403, w.Result().StatusCode)

	err = apiTest.PromoteAdmin(msg.Chat.ID, true)
	botTest.HandleErr(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/logs", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Result().StatusCode)
}
