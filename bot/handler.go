package bot

import (
	"log"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func CreateBot(bc *BotController) *tb.Bot {
	botToken := helpers.Env(helpers.ENV_TG_BOT_API_KEY)
	if botToken == "" {
		log.Fatalf("Please provide Telegram bot API key as ENV var '%s'", helpers.ENV_TG_BOT_API_KEY)
	}

	poller := &tb.LongPoller{Timeout: 20 * time.Second}
	userGuardPoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		// TODO: Start goroutine to update data?
		bc.Repo.EnrichUserData(upd.Message)
		return true
	})

	b, err := tb.NewBot(tb.Settings{
		Token:  botToken,
		Poller: userGuardPoller,
	})
	if err != nil {
		log.Fatal(err)
	}

	return b
}
