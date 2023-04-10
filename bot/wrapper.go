package bot

import (
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

type IBot interface {
	// Using from base package:
	Start()
	Handle(endpoint interface{}, h tb.HandlerFunc, m ...tb.MiddlewareFunc)
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
	Respond(c *tb.Callback, resp ...*tb.CallbackResponse) error
	// custom by me:
	Me() *tb.User
	SendSilent(logFn func(level helpers.Level, m *tb.Message, format string, v ...interface{}), to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
}

type Bot struct {
	bot *tb.Bot
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) Handle(endpoint interface{}, handler tb.HandlerFunc, m ...tb.MiddlewareFunc) {
	b.bot.Handle(endpoint, handler)
}

func (b *Bot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	return b.bot.Send(to, what, options...)
}

func (b *Bot) Respond(c *tb.Callback, resp ...*tb.CallbackResponse) error {
	return b.bot.Respond(c, resp...)
}

func (b *Bot) Me() *tb.User {
	return b.bot.Me
}

func (b *Bot) SendSilent(logFn func(level helpers.Level, m *tb.Message, format string, v ...interface{}), to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	m, err := b.Send(to, what, options...)
	if err != nil {
		logFn(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
	return m, err
}
