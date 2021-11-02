package bot

import (
	"github.com/LucaBernstein/beancount-bot-tg/db/crud"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/tucnak/telebot.v2"
)

func ReplyKeyboard(buttons []string) *tb.ReplyMarkup {
	kb := &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	for _, label := range buttons {
		kb.Text(label)
	}
	return kb
}

func EnrichHint(r *crud.Repo, i Input) *Hint {
	if i.key == "description" {
		return HintDescription(r, i.hint)
	}
	if i.key == "date" {
		return HintDate(r, i.hint)
	}
	if helpers.ArrayContains([]string{"from", "to"}, i.key) {
		return HintAccount(r, i.hint)
	}
	return i.hint
}

func HintAccount(r *crud.Repo, h *Hint) *Hint {
	return h
}

func HintDescription(r *crud.Repo, h *Hint) *Hint {
	return h
}

func HintDate(r *crud.Repo, h *Hint) *Hint {
	return h
}
