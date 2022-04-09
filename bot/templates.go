package bot

import (
	h "github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func (bc *BotController) templatesHandler(m *tb.Message) {
	sc := h.MakeSubcommandHandler("/"+CMD_TEMPLATE[0], true)
	sc.
		Add("list", bc.templatesHandleList).
		Add("add", bc.templatesHandleAdd).
		Add("rm", bc.templatesHandleRemove)
	parameters, err := sc.Handle(m)
	if err != nil {
		useErr := bc.templatesUse(m, parameters...)
		if useErr != nil {
			bc.Logf(ERROR, m, "could not handle templates command: %s - previous error for regular handle: %s", useErr.Error(), err.Error())
			bc.templatesHelp(m, nil)
		}
	}
}

func (bc *BotController) templatesHelp(m *tb.Message, err error) {

}

func (bc *BotController) templatesHandleList(m *tb.Message, params ...string) {

}

func (bc *BotController) templatesHandleAdd(m *tb.Message, params ...string) {

}

func (bc *BotController) templatesHandleRemove(m *tb.Message, params ...string) {

}

func (bc *BotController) processNewTemplateResponse(m *tb.Message, name TemplateName) (clearState bool) {
	return false
}

func (bc *BotController) templatesUse(m *tb.Message, params ...string) error {
	return nil
}
