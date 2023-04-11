package config_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/config"
	"github.com/LucaBernstein/beancount-bot-tg/api/helpers/apiTest"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func AllSettingsTypes() ([]string, error) {
	rows, err := db.Connection().Query(`SELECT "setting" FROM "bot::userSettingTypes"`)
	if err != nil {
		return nil, err
	}
	types := []string{}
	for rows.Next() {
		var next string
		err := rows.Scan(&next)
		if err != nil {
			return nil, err
		}
		types = append(types, next)
	}
	return types, nil
}

func TestFullConfigMap(t *testing.T) {
	token, mockBc, msg := apiTest.MockBcApiUser(t, 786)
	err := apiTest.PromoteAdmin(msg.Chat.ID, true)
	botTest.HandleErr(t, err)

	settings, err := AllSettingsTypes()
	botTest.HandleErr(t, err)

	r := gin.Default()
	g := r.Group("")
	config.NewRouter(mockBc).Hook(g)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)

	for _, setting := range settings {
		if strings.Contains(setting, "limitCache") {
			// Exclude transaction suggestion settings for now
			continue
		}
		assert.Contains(t, w.Body.String(), setting)
	}
}

// TODO: For setting, check, that user cannot elevate to admin!
func TestSetConfig(t *testing.T) {
	token, mockBc, _ := apiTest.MockBcApiUser(t, 427)

	r := gin.Default()
	g := r.Group("")
	config.NewRouter(mockBc).Hook(g)

	// Set tzOffset
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"setting":"user.tzOffset", "value":12}`))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)

	// Granting admin priviledges should not work
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(`{"setting":"user.isAdmin", "value":true}`))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Result().StatusCode)

	// No settings field should fail
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(`{"value":true}`))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Result().StatusCode)

	// Unsetting with null should work
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/", strings.NewReader(`{"setting":"user.currency", "value": null}`))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Result().StatusCode)

	// TODO: Enhance functionality coverage by looking at actual db states
}
