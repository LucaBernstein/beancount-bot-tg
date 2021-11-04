package migrations

import (
	"database/sql"
	"log"
)

func v2(db *sql.Tx) {
	v2AddCurrencyToUserTable(db)
}

func v2AddCurrencyToUserTable(db *sql.Tx) {
	sqlStatement := `
	ALTER TABLE "auth::user"
	ADD "currency" text NULL;
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
