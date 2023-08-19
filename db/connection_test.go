package db_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/db"
)

func TestBasicMigration(t *testing.T) {
	conn := db.Connection()
	defer conn.Close()
}
