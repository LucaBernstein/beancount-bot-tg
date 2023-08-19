package health_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/bot"
	"github.com/LucaBernstein/beancount-bot-tg/v2/db"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	"gopkg.in/telebot.v3"
)

func TestHealthEndpointRecentlyActiveUsers(t *testing.T) {
	// Given I have a bot
	conn := db.Connection()
	ctrl := bot.NewBotController(conn)
	repo := ctrl.Repo
	var chatId int64 = -17
	var err error
	msg := &telebot.Message{Chat: &telebot.Chat{ID: chatId}, Sender: &telebot.User{ID: chatId}}
	err = repo.EnrichUserData(msg)
	if err != nil {
		t.Errorf("Error encountered while issuing command: %e", err)
	}

	// When I create some log messages
	ctrl.Logf(helpers.ERROR, msg, "Test issued some log")
	ctrl.Logf(helpers.ERROR, msg, "Test issued some log")
	ctrl.Logf(helpers.ERROR, msg, "Test issued some log")
	ctrl.Logf(helpers.ERROR, msg, "Test issued some log")

	// Then user should be counted active
	count, err := repo.HealthGetUsersActiveCounts(1)
	if err != nil {
		t.Errorf("Getting active users should not error: %e", err)
	}
	if count < 1 {
		t.Errorf("At least 1 user should have been active recently...: %d", count)
	}
	users, err := repo.HealthGetUserCount()
	if err != nil {
		t.Errorf("Getting users should not error: %e", err)
	}
	if users < 1 {
		t.Errorf("At least 1 user should exist...: %d", users)
	}
}
