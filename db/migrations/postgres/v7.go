package postgres

import (
	"database/sql"
	"log"
)

func (c *Controller) V7(db *sql.Tx) {
	v7AddLoggingTable(db)
}

func v7AddLoggingTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "app::log" (
		"id" SERIAL PRIMARY KEY,
		"created" TIMESTAMP NOT NULL DEFAULT NOW(),
		"chat" TEXT,
		"level" NUMERIC NOT NULL,
		"message" TEXT NOT NULL
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
