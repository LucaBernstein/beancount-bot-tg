package bot

import (
	"fmt"
	"log"
	"strings"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) suggestionsHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/suggestions", true)
	sc.
		Add("list", bc.suggestionsHandleList).
		Add("add", bc.suggestionsHandleAdd).
		Add("rm", bc.suggestionsHandleRemove)
	err := sc.Handle(m)
	if err != nil {
		bc.suggestionsHelp(m, nil)
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

func (bc *BotController) suggestionsHandleList(m *tb.Message, params ...string) {
	p, err := extractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.t, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	values, err := bc.Repo.GetCacheHints(m, p.t)
	if err != nil {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while retrieving suggestions list for type '%s': %s", p.t, err.Error()))
		return
	}
	if len(values) == 0 {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Your suggestions list for type '%s' is currently empty.", p.t))
		return
	}
	bc.Bot.Send(m.Sender, fmt.Sprintf("These suggestions are currently saved for type '%s':\n\n", p.t)+
		strings.Join(values, "\n"))
}

func (bc *BotController) suggestionsHandleAdd(m *tb.Message, params ...string) {
	p, err := extractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.t, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	err = bc.Repo.PutCacheHints(m, map[string]string{p.t: p.value})
	if err != nil {
		bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while adding suggestion (%s): %s", p.value, err.Error()))
		return
	}
	bc.Bot.Send(m.Sender, "Successfully added suggestion(s).")
}

func (bc *BotController) suggestionsHandleRemove(m *tb.Message, params ...string) {
	p, err := extractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.t, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	log.Printf("(C%d): About to remove suggestion of type '%s' and value '%s'", m.Chat.ID, p.t, p.value)
	res, err := bc.Repo.DeleteCacheEntries(m, p.t, p.value)
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

type tv struct {
	t     string
	value string
}

func extractTypeValue(params ...string) (*tv, error) {
	e := &tv{}
	if len(params) < 1 || len(params) > 2 {
		return nil, fmt.Errorf("unexpected count of parameters")
	}
	if len(params) >= 1 {
		e.t = params[0]
	}
	if len(params) >= 2 {
		e.value = params[1]
	}
	return e, nil
}
