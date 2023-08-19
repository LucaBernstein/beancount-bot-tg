package bot

import (
	"log"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

func CreateBot(bc *BotController) IBot {
	const ENV_TG_BOT_API_KEY = "BOT_API_KEY"
	botToken := helpers.Env(ENV_TG_BOT_API_KEY)
	if botToken == "" {
		log.Fatalf("Please provide Telegram bot API key as ENV var '%s'", ENV_TG_BOT_API_KEY)
	}

	poller := &tb.LongPoller{Timeout: 20 * time.Second}
	userGuardPoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		message := upd.Message
		if message == nil && upd.Callback != nil {
			bc.Logf(TRACE, nil, "Message was nil. Seems to have been a callback. Proceeding.")
			return true
		}
		// TODO: Start goroutine to update data?
		err := bc.Repo.EnrichUserData(message)
		if err != nil {
			bc.Logf(ERROR, nil, "Error encountered in middlewarePoller: %s", err.Error())
		}
		return true
	})

	b, err := tb.NewBot(tb.Settings{
		Token:   botToken,
		Poller:  userGuardPoller,
		OnError: func(e error, context tb.Context) { bc.Logf(WARN, nil, "%s - context: %v", e.Error(), context) },
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Bot{bot: b}
}
