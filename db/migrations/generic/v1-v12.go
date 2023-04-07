package generic

import (
	"database/sql"
	"log"
)

func V1CreateSettingsTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "app::setting" (
		"key" TEXT PRIMARY KEY,
		"value" TEXT
	);

	INSERT INTO "app::setting" ("key", "value")
		VALUES ('schemaVersion', '0');
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
