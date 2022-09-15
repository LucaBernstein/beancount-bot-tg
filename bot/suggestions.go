package bot

import (
	"fmt"
	"strings"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func isAllowedSuggestionType(s string) bool {
	splits := strings.SplitN(s, ":", 2)
	_, exists := TEMPLATE_TYPE_HINTS[Type(splits[0])]
	return exists
}

func (bc *BotController) suggestionsHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/"+CMD_SUGGEST, true)
	sc.
		Add("list", bc.suggestionsHandleList).
		Add("add", bc.suggestionsHandleAdd).
		Add("rm", bc.suggestionsHandleRemove)
	_, err := sc.Handle(m)
	if err != nil {
		bc.suggestionsHelp(m, nil)
	}
}

func (bc *BotController) suggestionsHelp(m *tb.Message, err error) {
	suggestionTypes := []string{}
	for _, suggType := range h.AllowedSuggestionTypes() {
		if suggType == h.FIELD_ACCOUNT {
			suggType += ":[from,to,...]"
		}
		suggestionTypes = append(suggestionTypes, suggType)
	}
	errorMsg := ""
	if err != nil {
		errorMsg += fmt.Sprintf("Error executing your command: %s\n\n", err.Error())
	}

	_, err = bc.Bot.Send(Recipient(m), errorMsg+fmt.Sprintf(`Usage help for /suggestions:
/suggestions list <type>
/suggestions add <type> <value>
/suggestions rm <type> [value]

Parameter <type> is one from: [%s]`, strings.Join(suggestionTypes, ", ")))
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
	p.T = h.FqCacheKey(p.T)
	if !isAllowedSuggestionType(p.T) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	if p.Value != "" {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected value provided"))
		return
	}
	values, err := bc.Repo.GetCacheHints(m, p.T)
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), fmt.Sprintf("Error encountered while retrieving suggestions list for type '%s': %s", p.T, err.Error()))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if len(values) == 0 {
		_, err := bc.Bot.Send(Recipient(m), fmt.Sprintf("Your suggestions list for type '%s' is currently empty.", p.T))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("These suggestions are currently saved for type '%s':\n\n", p.T)+
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
	p.T = h.FqCacheKey(p.T)
	if !isAllowedSuggestionType(p.T) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	if p.Value == "" {
		bc.suggestionsHelp(m, fmt.Errorf("no value to add provided"))
		return
	}
	err = bc.Repo.PutCacheHints(m, map[string]string{p.T: p.Value})
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), fmt.Sprintf("Error encountered while adding suggestion (%s): %s", p.Value, err.Error()))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(Recipient(m), "Successfully added suggestion(s).")
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
	p.T = h.FqCacheKey(p.T)
	if !isAllowedSuggestionType(p.T) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	bc.Logf(TRACE, m, "About to remove suggestion of type '%s' and value '%s'", p.T, p.Value)
	res, err := bc.Repo.DeleteCacheEntries(m, p.T, p.Value)
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), "Error encountered while removing suggestion: "+err.Error())
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), "Error encountered while extracting affected entries: "+err.Error())
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
	_, err = bc.Bot.Send(Recipient(m), "Successfully removed suggestion(s)")
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}
