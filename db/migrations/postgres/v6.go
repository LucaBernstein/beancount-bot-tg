package postgres

import (
	"database/sql"
	"log"
)

func (c *Controller) V6(db *sql.Tx) {
	v6CleanDbSchema(db)
}

func v6CleanDbSchema(db *sql.Tx) {
	sqlStatement := `
	ALTER TABLE "auth::user"
		ALTER "tgUserId" SET NOT NULL,
		DROP "email";

	ALTER TABLE "bot::cache"
		ALTER "type" SET NOT NULL,
		ALTER "value" SET NOT NULL;

	ALTER TABLE "bot::transaction"
		ALTER "value" SET NOT NULL;
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
