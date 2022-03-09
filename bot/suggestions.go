package bot

import (
	"fmt"
	"strings"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

var (
	suggestionsMenu    = &tb.ReplyMarkup{}
	btnSuggListAccFrom = suggestionsMenu.Data("/suggestions list accFrom", "btnSuggestionsListAccFrom")
	btnSuggListAccTo   = suggestionsMenu.Data("/suggestions list accTo", "btnSuggestionsListAccTo")
	btnSuggListTxDesc  = suggestionsMenu.Data("/suggestions list txDesc", "btnSuggestionsListTxDesc")
)

func (bc *BotController) suggestionsHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/"+CMD_SUGGEST, true)
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

	suggestionsMenu.Inline(
		suggestionsMenu.Row(btnSuggListAccFrom),
		suggestionsMenu.Row(btnSuggListAccTo),
		suggestionsMenu.Row(btnSuggListTxDesc),
	)

	_, err = bc.Bot.Send(m.Sender, errorMsg+fmt.Sprintf(`Usage help for /suggestions:
/suggestions list <type>
/suggestions add <type> <value>
/suggestions rm <type> [value]

Parameter <type> is one from: [%s]`, suggestionTypes), suggestionsMenu)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) suggestionsHandleList(m *tb.Message, params ...string) {
	p, err := h.ExtractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.T, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	if p.Value != "" {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected value provided"))
		return
	}
	values, err := bc.Repo.GetCacheHints(m, p.T)
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while retrieving suggestions list for type '%s': %s", p.T, err.Error()))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if len(values) == 0 {
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Your suggestions list for type '%s' is currently empty.", p.T))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, fmt.Sprintf("These suggestions are currently saved for type '%s':\n\n", p.T)+
		strings.Join(values, "\n"))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) suggestionsHandleAdd(m *tb.Message, params ...string) {
	p, err := h.ExtractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.T, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	if p.Value == "" {
		bc.suggestionsHelp(m, fmt.Errorf("no value to add provided"))
		return
	}
	err = bc.Repo.PutCacheHints(m, map[string]string{p.T: p.Value})
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, fmt.Sprintf("Error encountered while adding suggestion (%s): %s", p.Value, err.Error()))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(m.Sender, "Successfully added suggestion(s).")
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) suggestionsHandleRemove(m *tb.Message, params ...string) {
	p, err := h.ExtractTypeValue(params...)
	if err != nil {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: %s", err.Error()))
		return
	}
	if !h.ArrayContainsC(h.AllowedSuggestionTypes(), p.T, false) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	bc.Logf(TRACE, m, "About to remove suggestion of type '%s' and value '%s'", p.T, p.Value)
	res, err := bc.Repo.DeleteCacheEntries(m, p.T, p.Value)
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Error encountered while removing suggestion: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		_, err := bc.Bot.Send(m.Sender, "Error encountered while extracting affected entries: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if rowCount == 0 {
		bc.suggestionsHelp(m, fmt.Errorf("entry could not be found in the database. "+
			"If your value contains spaces, consider putting it in double quotes (\")"))
		return
	}
	_, err = bc.Bot.Send(m.Sender, "Successfully removed suggestion(s)")
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}
