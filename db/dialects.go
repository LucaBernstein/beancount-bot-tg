package db

import (
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func DbType() string {
	return strings.ToUpper(helpers.Env("DB_TYPE"))
}

func AutoIncValue() string {
	switch DbType() {
	case "SQLITE":
		return "NULL"
	case "POSTGRES":
		return "DEFAULT"
	}
	return ""
}
