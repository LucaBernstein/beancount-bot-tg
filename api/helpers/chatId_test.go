package helpers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/helpers"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/telebot.v3"
)

func mockBcApiUser(t *testing.T) (r *gin.Engine, token string, mockBc *bot.BotController, m *telebot.Message) {
	repo := crud.NewRepo(db.Connection())
	mockBc = &bot.BotController{
		Repo: repo,
	}

	msg := &telebot.Message{Chat: &telebot.Chat{ID: 101}, Sender: &telebot.User{ID: 101}}
	err := mockBc.Repo.EnrichUserData(msg)
	botTest.HandleErr(t, err)
	err = mockBc.Repo.SetUserSetting(h.USERSET_ENABLEAPI, "true", msg.Chat.ID)
	botTest.HandleErr(t, err)
	nonce, err := mockBc.Repo.CreateApiVerification(msg.Chat.ID)
	botTest.HandleErr(t, err)
	token, err = mockBc.Repo.VerifyApiToken(msg.Chat.ID, nonce)
	botTest.HandleErr(t, err)

	return r, token, mockBc, msg
}

func TestChatIdEnrichment(t *testing.T) {
	_, token, mockBc, m := mockBcApiUser(t)
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
