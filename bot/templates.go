package bot

import (
	"fmt"
	"strings"

	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func (bc *BotController) templatesHandler(m *tb.Message) {
	base := CMD_TEMPLATE[0]
	if !strings.HasPrefix(m.Text, "/"+base) {
		base = CMD_TEMPLATE[1]
	}
	sc := h.MakeSubcommandHandler("/"+base, true)
	sc.
		Add("list", bc.templatesHandleList).
		Add("add", bc.templatesHandleAdd).
		Add("rm", bc.templatesHandleRemove)
	parameters, err := sc.Handle(m)
	if err != nil {
		useErr := bc.templatesUse(m, parameters...)
		if useErr != nil {
			bc.Logf(ERROR, m, "could not handle templates command: %s - previous error for regular handle: %s", useErr.Error(), err.Error())
			bc.templatesHelp(m, useErr)
		}
	}
}

func (bc *BotController) templatesHelp(m *tb.Message, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg += fmt.Sprintf("Error executing your command: %s\n\n", err.Error())
	}
	_, err = bc.Bot.Send(Recipient(m), errorMsg+`Usage help for /template:
	/template list [name]
	/template add <name>
	/template rm <name>
	
	To use an existing template, type:
	/template <name> [date]
	or use the short form:
	/t <name> [date]
	
	If omitted, date defaults to today.`)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) templatesHandleList(m *tb.Message, params ...string) {
	searchTemplate := ""
	if len(params) == 1 {
		searchTemplate = params[0]
	}
	templates, err := bc.Repo.GetTemplates(m, searchTemplate)
	if err != nil {
		bc.Logf(ERROR, m, "Error loading templates: %s", err.Error())
		_, err = bc.Bot.Send(Recipient(m), "There has been an error loading your templates.")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
	}
	if len(templates) == 0 {
		if searchTemplate == "" {
			_, err = bc.Bot.Send(Recipient(m), "You have not created any template yet. Please see /template")
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		} else {
			_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("No template name matched your query '%s'", searchTemplate))
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
	} else {
		templateList := []string{"These templates are currently available to you:"}
		for _, t := range templates {
			templateList = append(templateList, fmt.Sprintf("%s:\n%s", t.Name, t.Template))
		}
		messageSplits := bc.MergeMessagesHonorSendLimit(templateList, "\n\n")
		for _, message := range messageSplits {
			_, err = bc.Bot.Send(Recipient(m), message, clearKeyboard())
			if err != nil {
				bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
			}
		}
	}
}

func (bc *BotController) templatesHandleAdd(m *tb.Message, params ...string) {
	state := bc.State.GetType(m)
	if state != ST_NONE {
		_, err := bc.Bot.Send(Recipient(m), "There is another operation currently running for you. Please complete it or /cancel it before proceeding.")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if len(params) != 1 {
		bc.templatesHelp(m, fmt.Errorf("parameter count mismatch"))
		return
	}
	name := params[0]
	if strings.TrimSpace(name) == "" {
		bc.templatesHelp(m, fmt.Errorf("please name your template"))
		return
	}
	bc.State.StartTpl(m, name)
	_, err := bc.Bot.Send(Recipient(m), `Please provide a full transaction template. Variables are to be inserted as '${<variable>}'. The following variables can be used:
- ${amount}, ${-amount}, ${amount/i} (e.g. ${amount/2})
- ${date}
- ${description}
- ${account:from}
- ${account:to}
- ${account:<yourName>:<yourHint>}

Example:

${date} * "Store" "${description}"
  CheckingAccount ${-amount}
  Destination1 ${amount/2}
  Destination2
	
On templating out the amount will be auto-formatted. The date will either be filled with a specified value or fallback to the then current date.
The amount will be inserted with the currency.`)
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) templatesHandleRemove(m *tb.Message, params ...string) {
	if len(params) != 1 {
		bc.templatesHelp(m, fmt.Errorf("parameter count mismatch"))
		return
	}
	name := params[0]
	wasRemoved, err := bc.Repo.RmTemplate(m.Chat.ID, string(name))
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), "Something went wrong while deleting your template.")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	if !wasRemoved {
		_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("There was no template called '%s' to remove. Please check '/t list'.", name))
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return
	}
	_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("Successfully removed your template '%s'.", name))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
}

func (bc *BotController) processNewTemplateResponse(m *tb.Message, name TemplateName) (clearState bool) {
	template := m.Text
	err := bc.Repo.AddTemplate(m.Chat.ID, string(name), template)
	if err != nil {
		_, err := bc.Bot.Send(Recipient(m), "Something went wrong while saving your template. Please check whether the name already exists.")
		if err != nil {
			bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
		}
		return false
	}
	_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("Successfully created your template. You can use it from now on by typing '/t %s' (/t is short for /template).", name))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
	return true
}

type TemplateTx struct {
}

func (bc *BotController) templatesUse(m *tb.Message, params ...string) error {
	if len(params) < 1 || len(params) > 2 {
		return fmt.Errorf("parameter count mismatch")
	}
	name := params[0]
	var date string
	if len(params) >= 2 {
		date = params[1]
	}
	if name == "" {
		bc.templatesHelp(m, nil)
		return nil
	}
	res, err := bc.Repo.GetTemplates(m, name)
	if err != nil {
		bc.Logf(ERROR, m, "Getting template to create tx failed: %s", err.Error())
		return fmt.Errorf("unable to get the template you specified from the database at the moment")
	}
	if len(res) != 1 {
		bc.Logf(ERROR, m, "Getting template to create tx failed: Got multiple results for name '%s'.", name)
		return fmt.Errorf("could not find the template you specified. Please create it first")
	}
	tpl := res[0]
	tx, err := bc.State.TemplateTx(m, tpl.Template, bc.Repo.UserGetCurrency(m), date)
	if err != nil {
		bc.Logf(ERROR, m, "Creating tx from template failed: %s", err.Error())
		return fmt.Errorf("something went wrong creating a transaction from your template: %s", err.Error())
	}
	_, err = bc.Bot.Send(Recipient(m), fmt.Sprintf("Creating a new transaction from your template '%s'.", tpl.Name))
	if err != nil {
		bc.Logf(ERROR, m, "Sending bot message failed: %s", err.Error())
	}
	if tx.IsDone() {
		bc.finishTransaction(m, tx)
		return nil
	}
	hint := tx.NextHint(bc.Repo, m)
	bc.sendNextTxHint(hint, m)
	return nil
}
