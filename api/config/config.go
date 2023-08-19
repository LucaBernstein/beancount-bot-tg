package config

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
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

type SettingsPost struct {
	Setting string      `json:"setting"`
	Value   interface{} `json:"value"`
}

func (r *Router) SetConfig(c *gin.Context) {
	tgChatId := c.GetInt64("tgChatId")
	var setting SettingsPost
	err := c.ShouldBindJSON(&setting)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if setting.Setting == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "could not read setting to update from body",
		})
		return
	}
	if setting.Setting == helpers.USERSET_ADM {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "granting admin privileges via API is not allowed",
		})
		return
	}
	if setting.Value == nil {
		setting.Value = ""
	}
	log.Printf("Setting for user %d %s=%v", tgChatId, setting.Setting, setting.Value)
	// TODO: Value assertions (int, string, bool)
	err = r.bc.Repo.SetUserSetting(setting.Setting, fmt.Sprintf("%v", setting.Value), tgChatId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Status(http.StatusOK)
}
