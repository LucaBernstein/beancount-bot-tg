package migrations

import (
	"database/sql"
	"log"
)

func v11(db *sql.Tx) {
	v11ExtendAccountAndDescriptionHintsInSuggestions(db)
	v11ExtendAccountHintsInExistingTemplates(db)
	v11RemoveUserSettingTypesLimits(db)
}

func v11ExtendAccountAndDescriptionHintsInSuggestions(db *sql.Tx) {
	sqlStatement := `
	UPDATE "bot::cache"
	SET "type" = 'account:from'
	WHERE "type" = 'accFrom';
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement = `
	UPDATE "bot::cache"
	SET "type" = 'account:to'
	WHERE "type" = 'accTo';
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement = `
	UPDATE "bot::cache"
	SET "type" = 'description:'
	WHERE "type" = 'txDesc';
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}

func v11ExtendAccountHintsInExistingTemplates(db *sql.Tx) {
	sqlStatement := `
	UPDATE "bot::template"
	SET "template" = REPLACE("template", '${from}', '${account:from}')
	WHERE "template" LIKE '%${from}%';
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement = `
	UPDATE "bot::template"
	SET "template" = REPLACE("template", '${to}', '${account:to}')
	WHERE "template" LIKE '%${to}%';
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}

func v11RemoveUserSettingTypesLimits(db *sql.Tx) {
	sqlStatement := `
	DELETE FROM "bot::userSetting"
	WHERE "setting" LIKE 'user.limitCache.%';
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement = `
	DELETE FROM "bot::userSettingTypes"
	WHERE "setting" LIKE 'user.limitCache.%';
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	sqlStatement = `
	INSERT INTO "bot::userSettingTypes"
		("setting", "description")
	VALUES ('user.limitCache', 'limit cached value count for transactions. Array by type.');;
	`
	_, err = db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
