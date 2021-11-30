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

	CREATE TABLE "bot::notificationSchedule" (
		"tgChatId" NUMERIC REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"delayHours" NUMERIC NOT NULL,
		"notificationHour" NUMERIC NOT NULL
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
