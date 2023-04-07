package sqlite

import (
	"database/sql"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/generic"
)

type Controller struct{}

func (s *Controller) Schema(db *sql.DB) int {
	// Check if settings table holding schema version exists
	q := `
	SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='app::setting';
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
