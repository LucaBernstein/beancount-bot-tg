package bot

import (
	"fmt"
	"strings"

	. "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) suggestionsHandler(m *tb.Message) {
	// splits[0] is the command
	splits := strings.Split(m.Text, " ")
	var (
		subcommand string
		suggType   string
		value      string
	)
	if len(splits) <= 2 { // at least subcommand + type
		bc.suggestionsHelp(m)
		return
	}
	if len(splits) >= 3 {
		subcommand = splits[1]
		suggType = splits[2]
	}
	if len(splits) >= 4 {
		value = splits[3]
	}
	if len(splits) >= 5 {
		bc.suggestionsHelp(m)
		return
	}

	if !ArrayContainsC(AllowedSuggestionTypes(), suggType, false) {
		bc.suggestionsHelp(m)
		return
	}

	switch subcommand {
	case "list":
		bc.suggestionsHandleList(m, suggType)
	case "add":
		bc.suggestionsHandleAdd(m, suggType, value)
	case "rm":
		bc.suggestionsHandleRemove(m, suggType, value)
	default:
		bc.suggestionsHelp(m)
	}
}

func (bc *BotController) suggestionsHelp(m *tb.Message) {
	suggestionTypes := strings.Join(AllowedSuggestionTypes(), ", ")
	bc.Bot.Send(m.Sender, fmt.Sprintf(`Usage help for /suggestions:
/suggestions list <type>
/suggestions add <type> <value>
/suggestions rm <type> [value]

Parameter <type> is one from: [%s]`, suggestionTypes))
}

func (bc *BotController) suggestionsHandleList(m *tb.Message, t string) {
	values, err := bc.Repo.GetCacheHints(m, t)
	if err != nil {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while retrieving suggestions list for type '%s': %s", t, err.Error()))
		return
	}
	if len(values) == 0 {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Your suggestions list for type '%s' is currently empty.", t))
		return
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("These suggestions are currently saved for type '%s':\n\n", t)+
		strings.Join(values, "\n"))
}

func (bc *BotController) suggestionsHandleAdd(m *tb.Message, t string, value string) {
	if value == "" {
		// TODO: Not implemented yet. GH-Issue #17
		return
	}
	err := bc.Repo.PutCacheHints(m, map[string]string{t: value})
	if err != nil {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while adding suggestion (%s): %s", value, err.Error()))
		return
	}
	bc.Bot.Send(m.Sender, "Successfully added suggestion(s).")
}

func (bc *BotController) suggestionsHandleRemove(m *tb.Message, t string, value string) {
	err := bc.Repo.DeleteCacheEntries(m, t, value)
	if err != nil {
		bc.Bot.Send(m.Sender, "Error encountered while removing suggestion: "+err.Error())
		return
	}
	bc.Bot.Send(m.Sender, "Successfully removed suggestion(s)")
}
