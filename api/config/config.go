package config

import (
	"net/http"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	"github.com/gin-gonic/gin"
	"gopkg.in/telebot.v3"
)

func (r *Router) ReadConfig(c *gin.Context) {
	tgChatId := c.GetInt64("tgChatId")
	settings := map[string]interface{}{}
	// String settings
	for _, setting := range []string{helpers.USERSET_CUR, helpers.USERSET_TAG} {
		exists, val, err := r.bc.Repo.GetUserSetting(setting, tgChatId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"setting": setting,
			})
			return
		}
		if exists {
			settings[setting] = val
		} else {
			settings[setting] = nil
		}
	}
	// Boolean settings
	for _, setting := range []string{helpers.USERSET_ENABLEAPI, helpers.USERSET_OMITCMDSLASH, helpers.USERSET_ADM} {
		exists, val, err := r.bc.Repo.GetUserSetting(setting, tgChatId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"setting": setting,
			})
			return
		}
		if !(exists && strings.ToUpper(val) == "TRUE") && setting == helpers.USERSET_ADM {
			// Admin state should not be visible for non-admins.
			continue
		}
		settings[setting] = exists && strings.ToUpper(val) == "TRUE"
	}
	// Timezone offset setting
	m := &telebot.Message{Chat: &telebot.Chat{ID: tgChatId}}
	offset := r.bc.Repo.UserGetTzOffset(m)
	settings[helpers.USERSET_TZOFF] = offset

	c.JSON(http.StatusOK, settings)
}
