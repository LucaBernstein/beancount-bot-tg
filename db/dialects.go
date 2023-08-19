package db

import (
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
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

func Now() string {
	switch DbType() {
	case "SQLITE":
		return "CURRENT_TIMESTAMP"
	case "POSTGRES":
		return "NOW()"
	}
	return ""
}
