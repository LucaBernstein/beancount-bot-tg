package token_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/token"
	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/telebot.v3"
)

func mockBcApiUser(t *testing.T) (r *gin.Engine, repo *crud.Repo, m *telebot.Message, mockBot *botTest.MockBot) {
	repo = crud.NewRepo(db.Connection())
	mockBot = &botTest.MockBot{}
	mockBc := &bot.BotController{
		Repo: repo,
		Bot:  mockBot,
	}
	r = gin.Default()
	g := r.Group("")
	token.NewRouter(mockBc).Hook(g)

	msg := &telebot.Message{Chat: &telebot.Chat{ID: 67}, Sender: &telebot.User{ID: 67}}
	err := mockBc.Repo.EnrichUserData(msg)
	botTest.HandleErr(t, err)
	err = mockBc.Repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", msg.Chat.ID)
	botTest.HandleErr(t, err)

	return r, repo, msg, mockBot
}

func TestGrant(t *testing.T) {
	r, _, msg, mockBot := mockBcApiUser(t)

	// Cleanup already running verifications
	_, err := db.Connection().Exec(`DELETE FROM "app::apiToken"`)
	botTest.HandleErr(t, err)

	// Verification
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/verification/%d", msg.Chat.ID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)
	assert.Contains(t, w.Body.String(), `Successfully created token`)

	nonce := strings.Split(fmt.Sprintf("%v", mockBot.LastSentWhat), "please use this number: ")[1]

	// Grant
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/grant/%d/%s", msg.Chat.ID, nonce), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"token":"`)

	token := strings.Split(w.Body.String(), `"`)[3]

	// Revoke
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", fmt.Sprintf("/revoke/%s", token), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"affected":1`)
}
