package apiTest

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
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
	botTest.HandleErr(t, err)
	err = mockBc.Repo.SetUserSetting(helpers.USERSET_ENABLEAPI, "true", msg.Chat.ID)
	botTest.HandleErr(t, err)
	nonce, err := mockBc.Repo.CreateApiVerification(msg.Chat.ID)
	botTest.HandleErr(t, err)
	token, err = mockBc.Repo.VerifyApiToken(msg.Chat.ID, nonce)
	botTest.HandleErr(t, err)

	return token, mockBc, msg
}

func PromoteAdmin(chatId int64, becomes bool) error {
	_, err := db.Connection().Exec(`DELETE FROM "bot::userSetting" WHERE "tgChatId" = $1 AND "setting" = $2`, chatId, helpers.USERSET_ADM)
	if err != nil {
		return err
	}
	if becomes {
		_, err = db.Connection().Exec(`INSERT INTO "bot::userSetting" ("tgChatId", "setting", "value") VALUES ($1, $2, $3)`, chatId, helpers.USERSET_ADM, "true")
	}
	return err
}
