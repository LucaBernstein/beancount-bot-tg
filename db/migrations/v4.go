package migrations

import (
	"database/sql"
	"log"
)

func v4(db *sql.Tx) {
	v4UserAdmin(db)
}

func v4UserAdmin(db *sql.Tx) {
	sqlStatement := `
	ALTER TABLE "auth::user"
	ADD "isAdmin" BOOLEAN DEFAULT FALSE NOT NULL;
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
