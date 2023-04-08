package sqlite

import (
	"database/sql"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/generic"
)

func (c *Controller) V1(db *sql.Tx) {
	generic.V1CreateSettingsTable(db)
}
func (c *Controller) V2(db *sql.Tx)  {}
func (c *Controller) V3(db *sql.Tx)  {}
func (c *Controller) V4(db *sql.Tx)  {}
func (c *Controller) V5(db *sql.Tx)  {}
func (c *Controller) V6(db *sql.Tx)  {}
func (c *Controller) V7(db *sql.Tx)  {}
func (c *Controller) V8(db *sql.Tx)  {}
func (c *Controller) V9(db *sql.Tx)  {}
func (c *Controller) V10(db *sql.Tx) {}
func (c *Controller) V11(db *sql.Tx) {}
func (c *Controller) V12(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "auth::user" (
		"tgChatId" INTEGER PRIMARY KEY,
		"tgUsername" TEXT
	);

	CREATE TABLE "bot::notificationSchedule" (
		"tgChatId" INTEGER REFERENCES "auth::user" ("tgChatId") ON DELETE CASCADE NOT NULL,
		"delayHours" INTEGER NOT NULL,
		"notificationHour" INTEGER NOT NULL
	);

	CREATE TABLE "bot::transaction" (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"tgChatId" INTEGER REFERENCES "auth::user" ("tgChatId") ON DELETE CASCADE NOT NULL,
		"created" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"value" TEXT NOT NULL,
		"archived" BOOLEAN DEFAULT FALSE NOT NULL
	);

	CREATE TABLE "bot::cache" (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"tgChatId" INTEGER REFERENCES "auth::user" ("tgChatId") ON DELETE CASCADE NOT NULL,
		"lastUsed" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"type" TEXT NOT NULL,
		"value" TEXT NOT NULL
	);

	CREATE TABLE "app::log" (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"created" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"chat" TEXT,
		"level" INTEGER NOT NULL,
		"message" TEXT NOT NULL
	);

	CREATE TABLE "bot::userSettingTypes" (
		"setting" TEXT UNIQUE NOT NULL,
		"description" TEXT
	);

	CREATE TABLE "bot::userSetting" (
		"tgChatId" INTEGER REFERENCES "auth::user" ("tgChatId") ON DELETE CASCADE NOT NULL,
		"setting" TEXT REFERENCES "bot::userSettingTypes" ("setting") ON DELETE CASCADE NOT NULL,
		"value" TEXT,

		PRIMARY KEY ("tgChatId", "setting")
	);

	CREATE TABLE "bot::template" (
		"tgChatId" INTEGER REFERENCES "auth::user" ("tgChatId") ON DELETE CASCADE NOT NULL,
		"name" TEXT NOT NULL,
		"template" TEXT NOT NULL,

		PRIMARY KEY ("tgChatId", "name")
	);

	-- Fill "bot::userSettingTypes" 
	INSERT INTO "bot::userSettingTypes" ("setting", "description") VALUES
		('user.currency', 'currency to use by default for simple transactions'),
		('user.isAdmin', 'grants admin priviledges to a user. use with caution!'),
		('user.vacationTag', 'add a default tag to simple transactions'),

		('user.limitCache.txDesc', 'limit caching of simple transaction descriptions'),
		('user.limitCache.accFrom', 'limit caching of simple transaction from accounts'),
		('user.limitCache.accTo', 'limit caching of simple transaction to accounts'),
		
		('user.tzOffset', 'set timezone offset for automatic transaction dates and notifications'),

		('user.omitCommandSlash', 'make leading slash for commands optional');
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
