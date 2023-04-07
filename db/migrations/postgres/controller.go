package postgres

import (
	"database/sql"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/generic"
)

type Controller struct{}

func (c *Controller) Schema(db *sql.DB) int {
	// Check if settings table holding schema version exists
	q := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE  table_schema = 'public'
		AND    table_name   = 'app::setting'
	);
	`
	exists := false
	err := db.QueryRow(q).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		return 0
	}

	return generic.Schema(db)
}
