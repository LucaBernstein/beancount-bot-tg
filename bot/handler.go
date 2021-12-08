package bot

import (
	"log"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func CreateBot(bc *BotController) IBot {
	const ENV_TG_BOT_API_KEY = "BOT_API_KEY"
	botToken := helpers.Env(ENV_TG_BOT_API_KEY)
	if botToken == "" {
		log.Fatalf("Please provide Telegram bot API key as ENV var '%s'", ENV_TG_BOT_API_KEY)
	}

	poller := &tb.LongPoller{Timeout: 20 * time.Second}
	userGuardPoller := tb.NewMiddlewarePoller(poller, func(upd *tb.Update) bool {
		// TODO: Start goroutine to update data?
		err := bc.Repo.EnrichUserData(upd.Message)
		if err != nil {
			bc.Logf(ERROR, nil, "Error encountered in middlewarePoller: %s", err.Error())
		}
		return true
	})

	b, err := tb.NewBot(tb.Settings{
		Token:    botToken,
		Poller:   userGuardPoller,
		Reporter: func(e error) { bc.Logf(WARN, nil, "%s", e.Error()) },
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Bot{bot: b}
}
