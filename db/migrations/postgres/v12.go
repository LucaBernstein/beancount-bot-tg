package postgres

import (
	"database/sql"
	"log"
)

func (c *Controller) V12(db *sql.Tx) {
	v12AddLeadingSlashSetting(db)
}

func v12AddLeadingSlashSetting(db *sql.Tx) {
	sqlStatement := `
	INSERT INTO "bot::userSettingTypes" ("setting", "description") VALUES
		('user.omitCommandSlash', 'make leading slash for commands optional');
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
