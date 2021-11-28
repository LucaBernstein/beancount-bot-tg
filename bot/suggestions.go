package bot

import (
	"fmt"
	"log"
	"strings"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) suggestionsHandler(m *tb.Message) {
	var (
		command    string
		subcommand string
		suggType   string
		value      string
	)
	command = m.Text
	if strings.Contains(m.Text, "\"") {
		// Assume quoted value, e.g. '/sugg rm type "some value with spaces"'
		splits := strings.Split(m.Text, "\"")

		command = splits[0]
		value = splits[1]
	}

	splits := strings.Split(strings.TrimSpace(command), " ")
	if len(splits) != 3 { // at least command + subcommand + type
		bc.suggestionsHelp(m, nil)
		return
	}

	subcommand = splits[1]
	suggType = splits[2]

	if len(splits) == 4 { // exactly 4 splits mean unspaced value. Remove quotes to make sure.
		value = strings.Trim(splits[3], "\"")
	}

	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), suggType, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
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
		bc.suggestionsHelp(m, nil)
		return
	}
}

func (bc *BotController) suggestionsHelp(m *tb.Message, err error) {
	suggestionTypes := strings.Join(h.AllowedSuggestionTypes(), ", ")
	errorMsg := ""
	if err != nil {
		errorMsg += fmt.Sprintf("Error executing your command: %s\n\n", err.Error())
	}
	bc.Bot.Send(m.Sender, errorMsg+fmt.Sprintf(`Usage help for /suggestions:
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
	log.Printf("(C%d): About to remove suggestion of type '%s' and value '%s'", m.Chat.ID, t, value)
	res, err := bc.Repo.DeleteCacheEntries(m, t, value)
	if err != nil {
		bc.Bot.Send(m.Sender, "Error encountered while removing suggestion: "+err.Error())
		return
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		bc.Bot.Send(m.Sender, "Error encountered while extracting affected entries: "+err.Error())
		return
	}
	if rowCount == 0 {
		bc.suggestionsHelp(m, fmt.Errorf("entry could not be found in the database. "+
			"If your value contains spaces, consider putting it in double quotes (\")"))
		return
	}
	bc.Bot.Send(m.Sender, "Successfully removed suggestion(s)")
}
