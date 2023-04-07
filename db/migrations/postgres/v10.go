package postgres

import (
	"database/sql"
	"log"
)

func (c *Controller) V10(db *sql.Tx) {
	v10CleanUserTable(db)
}

func v10CleanUserTable(db *sql.Tx) {
	sqlStatement := `
	ALTER TABLE "auth::user"
		DROP "tgUserId";
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
