package sqlite

import (
	"database/sql"
	"log"
)

func (c *Controller) V14(db *sql.Tx) {
	v14ApiTokens(db)
}

func v14ApiTokens(db *sql.Tx) {
	_, err := db.Exec(`
	CREATE TABLE "app::apiToken" (
		"token"		VARCHAR(255) PRIMARY KEY NOT NULL,
		"nonce" 	VARCHAR(16),
		"tgChatId"	INTEGER REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"createdOn" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
}
