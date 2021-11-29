package bot

import (
	"fmt"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) configHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/"+CMD_CONFIG, true)
	sc.
		Add("currency", bc.configHandleCurrency)
	err := sc.Handle(m)
	if err != nil {
		bc.configHelp(m)
	}
}

func (bc *BotController) configHelp(m *tb.Message) {
	bc.Bot.Send(m.Sender, fmt.Sprintf("Usage help for /%s:\n\n/%s currency <c> - Change default currency", CMD_CONFIG, CMD_CONFIG))
}

func (bc *BotController) configHandleCurrency(m *tb.Message, params ...string) {
	currency := bc.Repo.UserGetCurrency(m)
	if len(params) == 0 { // 0 params: GET currency
		// Return currently set currency
		bc.Bot.Send(m.Sender, fmt.Sprintf("Your current currency is set to '%s'. To change it add the new currency to use to the command like this: '/%s currency EUR'.", currency, CMD_CONFIG))
		return
	} else if len(params) > 1 { // 2 or more params: too many

	}
	// Set new currency
	newCurrency := params[0]
	err := bc.Repo.UserSetCurrency(m, newCurrency)
	if err != nil {
		bc.Bot.Send(m.Sender, "An error ocurred saving your currency preference: "+err.Error())
		return
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("Changed default currency for all future transactions from '%s' to '%s'.", currency, newCurrency))
}
