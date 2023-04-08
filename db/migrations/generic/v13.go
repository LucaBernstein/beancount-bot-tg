package generic

import (
	"database/sql"
	"log"
)

func V13AddSettingEnableApi(db *sql.Tx) {
	sqlStatement := `
	INSERT INTO "bot::userSettingTypes" ("setting", "description") VALUES
		('user.enableApi', 'enable API/UI support');
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
