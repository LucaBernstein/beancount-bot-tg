package config_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/api/config"
	"github.com/LucaBernstein/beancount-bot-tg/bot/botTest"
	"github.com/LucaBernstein/beancount-bot-tg/db"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func PromoteAdmin(chatId int64) error {
	_, err := db.Connection().Exec(`DELETE FROM "bot::userSetting" WHERE "tgChatId" = $1 AND "setting" = $2`, chatId, helpers.USERSET_ADM)
	if err != nil {
		return err
	}
	_, err = db.Connection().Exec(`INSERT INTO "bot::userSetting" ("tgChatId", "setting", "value") VALUES ($1, $2, $3)`, chatId, helpers.USERSET_ADM, "true")
	return err
}

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
	token, mockBc, msg := botTest.MockBcApiUser(t, 786)
	err := PromoteAdmin(msg.Chat.ID)
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
