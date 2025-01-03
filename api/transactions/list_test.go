package transactions_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/api/transactions"
	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/telebot.v3"
)

func handleErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}
}

func mockBcApiUser(t *testing.T) (r *gin.Engine, w *httptest.ResponseRecorder, token string, repo *crud.Repo, m *telebot.Message) {
	repo = crud.NewRepo(db.Connection())
	mockBc := &bot.BotController{
		Repo: repo,
	}
	r = gin.Default()
	g := r.Group("")
	transactions.NewRouter(mockBc).Hook(g)

	msg := &telebot.Message{Chat: &telebot.Chat{ID: 55}, Sender: &telebot.User{ID: 55}}
	err := mockBc.Repo.EnrichUserData(msg)
	handleErr(t, err)
	err = mockBc.Repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", msg.Chat.ID)
	handleErr(t, err)
	nonce, err := mockBc.Repo.CreateApiVerification(msg.Chat.ID)
	handleErr(t, err)
	token, err = mockBc.Repo.VerifyApiToken(msg.Chat.ID, nonce)
	handleErr(t, err)

	w = httptest.NewRecorder()
	return r, w, token, repo, msg
}

func TestList(t *testing.T) {
	r, w, token, repo, msg := mockBcApiUser(t)

	handleErr(t, repo.RecordTransaction(msg.Chat.ID, "my tx"))

	req, _ := http.NewRequest("GET", "/list", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"booking":"my tx"`)
}

func TestListTextFormat(t *testing.T) {
	r, w, token, _, _ := mockBcApiUser(t)


	req, _ := http.NewRequest("GET", "/list?format=text", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "my tx\n", w.Body.String())
}

func TestListDeleteSingle(t *testing.T) {
	r, w, token, repo, msg := mockBcApiUser(t)

	handleErr(t, repo.RecordTransaction(msg.Chat.ID, "my tx"))
	tx, err := repo.GetTransactions(msg, false)
	handleErr(t, err)
	id := tx[0].Id

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/list/%d", id), nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"affected":1`)
}

func TestListDeleteAll(t *testing.T) {
	r, w, token, repo, msg := mockBcApiUser(t)

	handleErr(t, repo.RecordTransaction(msg.Chat.ID, "my tx"))

	req, _ := http.NewRequest("DELETE", "/list", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"affected":`)
	assert.NotContains(t, w.Body.String(), `"affected":0`)

	tx, err := repo.GetTransactions(msg, false)
	handleErr(t, err)

	assert.Equal(t, 0, len(tx))
}
