package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type IBot interface {
	// Using from base package:
	Start()
	Handle(endpoint interface{}, handler interface{})
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
	Respond(c *tb.Callback, resp ...*tb.CallbackResponse) error
	// custom by me:
	Me() *tb.User
}

type Bot struct {
	bot *tb.Bot
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) Handle(endpoint interface{}, handler interface{}) {
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
