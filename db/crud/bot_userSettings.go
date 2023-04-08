package crud

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	tb "gopkg.in/telebot.v3"
)

func (r *Repo) GetUserSetting(setting string, tgChatId int64) (exists bool, val string, err error) {
	rows, err := r.db.Query(`
		SELECT "value"
		FROM "bot::userSetting"
		WHERE "tgChatId" = $1 AND "setting" = $2
		`, tgChatId, setting)
	if err != nil {
		return
	}
	defer rows.Close()
	var value sql.NullString
	if rows.Next() {
		exists = true
		err = rows.Scan(&value)
		if err != nil {
			return
		}
		if value.Valid {
			val = value.String
		}
	}
	return
}

func (r *Repo) SetUserSetting(setting string, value string, tgChatId int64) (err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not create db tx for userSetting: %s", err.Error())
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM "bot::userSetting" WHERE "tgChatId" = $1 AND "setting" = $2`, tgChatId, setting)
	if err != nil {
		return fmt.Errorf("could not delete setting: %s", err.Error())
	}
	if value != "" {
		_, err = tx.Exec(`INSERT INTO "bot::userSetting" ("tgChatId", "setting", "value") VALUES ($1, $2, $3)`, tgChatId, setting, value)
		if err != nil {
			return fmt.Errorf("could not insert setting: %s", err.Error())
		}
	}
	err = tx.Commit()
	if err != nil {
		LogDbf(r, helpers.ERROR, nil, "Could not commit userSetting tx: %s", err.Error())
	}
	return
}

func (r *Repo) DeleteAllUserSettings(tgChatId int64) error {
	_, err := r.db.Exec(`DELETE FROM "bot::userSetting" WHERE "tgChatId" = $1`, tgChatId)
	if err != nil {
		return fmt.Errorf("could not delete settings: %s", err.Error())
	}
	return nil
}

// Currency

func (r *Repo) UserGetCurrency(m *tb.Message) string {
	currencyCacheKey := helpers.USERSET_CUR
	exists, value, err := r.GetUserSetting(currencyCacheKey, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user currency: %s", err.Error())
	}
	if exists && value != "" {
		return value
	}
	return helpers.DEFAULT_CURRENCY
}

func (r *Repo) UserSetCurrency(m *tb.Message, currency string) error {
	return r.SetUserSetting(helpers.USERSET_CUR, currency, m.Chat.ID)
}

// Tag

func (r *Repo) UserGetTag(m *tb.Message) string {
	_, value, err := r.GetUserSetting(helpers.USERSET_TAG, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Could not get tag: %s", err.Error())
		return ""
	}
	return value
}

func (r *Repo) UserSetTag(m *tb.Message, tag string) error {
	return r.SetUserSetting(helpers.USERSET_TAG, tag, m.Chat.ID)
}

// TzOffset

func (r *Repo) UserGetTzOffset(m *tb.Message) (tzOffset int) {
	exists, value, err := r.GetUserSetting(helpers.USERSET_TZOFF, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Could not get tzOffset: %s", err.Error())
		return
	}
	if exists && value != "" {
		tzOffset, err = strconv.Atoi(value)
		if err != nil {
			LogDbf(r, helpers.ERROR, m, "Could not parse tzOffset: %s", err.Error())
			return 0
		}
	}
	return
}

func (r *Repo) UserSetTzOffset(m *tb.Message, timezoneOffset int) error {
	tzOffsetS := ""
	if timezoneOffset != 0 {
		tzOffsetS = strconv.Itoa(timezoneOffset)
	}
	return r.SetUserSetting(helpers.USERSET_TZOFF, tzOffsetS, m.Chat.ID)
}

// Admin

func (r *Repo) UserIsAdmin(m *tb.Message) (isAdmin bool) {
	adminCacheKey := helpers.USERSET_ADM
	exists, value, err := r.GetUserSetting(adminCacheKey, m.Chat.ID)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while getting user isAdmin flag: %s", err.Error())
	}
	if !exists || value == "" {
		return
	}
	isAdmin, err = strconv.ParseBool(value)
	if err != nil {
		LogDbf(r, helpers.ERROR, m, "Encountered error while parsing isAdmin setting value: %s", err.Error())
		return false
	}
	return
}
