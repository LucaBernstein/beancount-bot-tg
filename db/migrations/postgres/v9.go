package postgres

import (
	"database/sql"
	"log"
)

func (c *Controller) V9(db *sql.Tx) {
	v9AddTable(db)
}

func v9AddTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "bot::template" (
		"tgChatId" NUMERIC REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"name" TEXT NOT NULL,
		"template" TEXT NOT NULL,

		PRIMARY KEY ("tgChatId", "name")
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
