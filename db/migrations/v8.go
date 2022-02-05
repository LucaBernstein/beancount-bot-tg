package migrations

import (
	"database/sql"
	"log"
)

func v8(db *sql.Tx) {
	v8AddTable(db)
}

func v8AddTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "bot::userSettingTypes" (
		"setting" TEXT UNIQUE NOT NULL,
		"description" TEXT
	);

	CREATE TABLE "bot::userSetting" (
		"tgChatId" NUMERIC REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"setting" TEXT REFERENCES "bot::userSettingTypes" ("setting") NOT NULL,
		"value" TEXT,

		PRIMARY KEY ("tgChatId", "setting")
	);

	-- Fill "bot::userSettingTypes" 

	INSERT INTO "bot::userSettingTypes" ("setting", "description") VALUES
		('user.currency', 'currency to use by default for simple transactions'),
		('user.isAdmin', 'grants admin priviledges to a user. use with caution!'),
		('user.vacationTag', 'add a default tag to simple transactions'),

		('user.limitCache.txDesc', 'limit caching of simple transaction descriptions'),
		('user.limitCache.accFrom', 'limit caching of simple transaction from accounts'),
		('user.limitCache.accTo', 'limit caching of simple transaction to accounts');
		
		('user.tzOffset', 'set timezone offset for automatic transaction dates and notifications');

	-- Migrate values from "auth::user" to userSettings

	INSERT INTO "bot::userSetting" (
		SELECT "tgChatId", 'user.currency', "currency" FROM "auth::user" WHERE "currency" IS NOT NULL
	);

	INSERT INTO "bot::userSetting" (
		SELECT "tgChatId", 'user.isAdmin', "isAdmin" FROM "auth::user" WHERE "isAdmin" = TRUE
	);

	INSERT INTO "bot::userSetting" (
		SELECT "tgChatId", 'user.vacationTag', "tag" FROM "auth::user" WHERE "tag" IS NOT NULL
	);

	-- Delete migrated columns from "auth::user" table

	ALTER TABLE "auth::user"
	DROP COLUMN "currency";

	ALTER TABLE "auth::user"
	DROP COLUMN "isAdmin";

	ALTER TABLE "auth::user"
	DROP COLUMN "tag";

	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
