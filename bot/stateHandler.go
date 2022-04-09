package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type chatId int64
type StateType string
type TemplateName string

const (
	ST_NONE = ""
	ST_TX   = "tx"
	ST_TPL  = "tpl"
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
	tx, err := CreateSimpleTx(m, suggestedCur)
	if err != nil {
		return nil, err
	}
	s.states[(chatId)(m.Chat.ID)] = ST_TX
	s.txStates[(chatId)(m.Chat.ID)] = tx
	return tx, nil
}
