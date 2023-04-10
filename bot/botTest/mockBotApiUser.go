package botTest

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"gopkg.in/telebot.v3"
)

func MockBcApiUser(t *testing.T, id int64) (token string, mockBc *bot.BotController, m *telebot.Message) {
	repo := crud.NewRepo(db.Connection())
	mockBc = &bot.BotController{
		Repo: repo,
	}

	msg := &telebot.Message{Chat: &telebot.Chat{ID: id}, Sender: &telebot.User{ID: id}}
	err := mockBc.Repo.EnrichUserData(msg)
	HandleErr(t, err)
	err = mockBc.Repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", msg.Chat.ID)
	HandleErr(t, err)
	nonce, err := mockBc.Repo.CreateApiVerification(msg.Chat.ID)
	HandleErr(t, err)
	token, err = mockBc.Repo.VerifyApiToken(msg.Chat.ID, nonce)
	HandleErr(t, err)

	return token, mockBc, msg
}
