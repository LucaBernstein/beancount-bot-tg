package migrations

import (
	"database/sql"
	"log"
)

func v3(db *sql.Tx) {
	v3CleanCache(db)
}

func v3CleanCache(db *sql.Tx) {
	sqlStatement := `
	DELETE FROM "bot::cache"
	WHERE "type" NOT IN ('accTo', 'accFrom', 'txDesc')
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
