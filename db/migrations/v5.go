package migrations

import (
	"database/sql"
	"log"
)

func v5(db *sql.Tx) {
	v5TxTags(db)
}

func v5TxTags(db *sql.Tx) {
	sqlStatement := `
	ALTER TABLE "auth::user"
	ADD "tag" TEXT NULL DEFAULT NULL;
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
