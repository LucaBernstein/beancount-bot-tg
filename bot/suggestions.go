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

	bc.Bot.SendSilent(bc, Recipient(m), errorMsg+fmt.Sprintf(`Usage help for /suggestions:
/suggestions list <type>
/suggestions add <type> <value> [<value>...]
/suggestions rm <type> [value]

Parameter <type> is one of: [%s]

Adding multiple suggestions at once is supported either by space separation (with quotation marks) or using newlines.`, strings.Join(suggestionTypes, ", ")))
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
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Error encountered while retrieving suggestions list for type '%s': %s", p.T, err.Error()))
		return
	}
	if len(values) == 0 {
		bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Your suggestions list for type '%s' is currently empty.", p.T))
		return
	}
	bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("These suggestions are currently saved for type '%s':\n\n", p.T)+
		strings.Join(values, "\n"))
}

func (bc *BotController) suggestionsHandleAdd(m *tb.Message, params ...string) {
	if len(params) < 2 {
		bc.suggestionsHelp(m, fmt.Errorf("error encountered while retrieving suggestions list: Insufficient parameters count"))
	}
	suggestionTypeSplit := strings.SplitN(params[0], "\n", 2)
	suggestionType := suggestionTypeSplit[0]
	remainder := ""
	if len(suggestionTypeSplit) > 1 {
		remainder = suggestionTypeSplit[1]
	}
	// Undo splitting by spaces: concat and then split by newlines for bulk suggestions adding support
	restoredValue := remainder + " " + strings.Join(params[1:], " ")
	singleValues := strings.Split(strings.TrimSpace(restoredValue), "\n")

	suggestionType = h.FqCacheKey(suggestionType)
	if !isAllowedSuggestionType(suggestionType) {
		bc.suggestionsHelp(m, fmt.Errorf("unexpected subcommand"))
		return
	}
	if len(singleValues) == 0 || (len(singleValues) == 1 && singleValues[0] == "") {
		bc.suggestionsHelp(m, fmt.Errorf("no value to add provided"))
		return
	}
	for _, value := range singleValues {
		err := bc.Repo.PutCacheHints(m, map[string]string{suggestionType: value})
		if err != nil {
			bc.Bot.SendSilent(bc, Recipient(m), fmt.Sprintf("Error encountered while adding suggestion (%s): %s", value, err.Error()))
			return
		}
	}
	bc.Bot.SendSilent(bc, Recipient(m), "Successfully added suggestion(s).")
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
		bc.Bot.SendSilent(bc, Recipient(m), "Error encountered while removing suggestion: "+err.Error())
		return
	}
	rowCount, err := res.RowsAffected()
	if err != nil {
		bc.Bot.SendSilent(bc, Recipient(m), "Error encountered while extracting affected entries: "+err.Error())
		return
	}
	if rowCount == 0 {
		bc.suggestionsHelp(m, fmt.Errorf("entry could not be found in the database. "+
			"If your value contains spaces, consider putting it in double quotes (\")"))
		return
	}
	bc.Bot.SendSilent(bc, Recipient(m), "Successfully removed suggestion(s)")
}
