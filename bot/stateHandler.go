package bot

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type chatId int64

type StateHandler struct {
	states map[chatId]Tx
}

func NewStateHandler() *StateHandler {
	return &StateHandler{
		states: map[chatId]Tx{},
	}
}

func (s *StateHandler) Clear(m *tb.Message) {
	delete(s.states, (chatId)(m.Chat.ID))
}

func (s *StateHandler) Get(m *tb.Message) Tx {
	return s.states[(chatId)(m.Chat.ID)]
}

func (s *StateHandler) SimpleTx(m *tb.Message) Tx {
	tx := CreateSimpleTx()
	s.states[(chatId)(m.Chat.ID)] = tx
	return tx
}
