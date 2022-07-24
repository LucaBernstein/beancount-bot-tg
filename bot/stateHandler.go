package bot

import (
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"
)

type chatId int64
type StateType string
type TemplateName string

const (
	ST_NONE StateType = ""
	ST_TX   StateType = "tx"
	ST_TPL  StateType = "tpl"
)

type StateHandler struct {
	states    map[chatId]StateType
	txStates  map[chatId]Tx
	tplStates map[chatId]TemplateName
}

func NewStateHandler() *StateHandler {
	return &StateHandler{
		states:    map[chatId]StateType{},
		txStates:  map[chatId]Tx{},
		tplStates: map[chatId]TemplateName{},
	}
}

func (s *StateHandler) Clear(m *tb.Message) {
	delete(s.states, (chatId)(m.Chat.ID))
}

func (s *StateHandler) GetType(m *tb.Message) StateType {
	if st, exists := s.states[(chatId)(m.Chat.ID)]; exists {
		return st
	}
	return ST_NONE
}

func (s *StateHandler) GetTx(m *tb.Message) Tx {
	if s.states[(chatId)(m.Chat.ID)] == ST_TX {
		return s.txStates[(chatId)(m.Chat.ID)]
	}
	return nil
}

func (s *StateHandler) SimpleTx(m *tb.Message, suggestedCur string) (Tx, error) {
	tx, err := CreateSimpleTx(suggestedCur, TEMPLATE_SIMPLE_DEFAULT)
	if err != nil {
		return nil, err
	}
	// Set date:
	command := strings.Split(m.Text, " ")
	if len(command) >= 2 {
		date, err := ParseDate(command[1])
		if err != nil {
			return nil, err
		}
		tx.SetDate(date)
	}
	s.states[(chatId)(m.Chat.ID)] = ST_TX
	s.txStates[(chatId)(m.Chat.ID)] = tx
	return tx, nil
}

func (s *StateHandler) TemplateTx(m *tb.Message, template, suggestedCur, date string) (Tx, error) {
	tx, err := CreateSimpleTx(suggestedCur, template)
	if err != nil {
		return nil, err
	}
	s.states[(chatId)(m.Chat.ID)] = ST_TX
	s.txStates[(chatId)(m.Chat.ID)] = tx

	// set date
	if date != "" {
		date, err := ParseDate(date)
		if err != nil {
			return nil, err
		}
		return tx.SetDate(date)
	}
	return tx, nil
}

func (s *StateHandler) StartTpl(m *tb.Message, name string) {
	s.states[(chatId)(m.Chat.ID)] = ST_TPL
	s.tplStates[(chatId)(m.Chat.ID)] = TemplateName(name)
}

func (s *StateHandler) CountOpen() int {
	return len(s.states)
}
