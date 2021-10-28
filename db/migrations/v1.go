package migrations

import (
	"database/sql"
	"log"
)

func v1(db *sql.Tx) {
	v1CreateSettingsTable(db)
	v1CreateUserTable(db)
	v1CreateTransactionsTable(db)
	v1CreateValueCache(db)
}

func v1CreateSettingsTable(db *sql.Tx) {
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

func v1CreateUserTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "auth::user" (
		"tgChatId" NUMERIC PRIMARY KEY,
		"tgUserId" NUMERIC,
		"tgUsername" TEXT,
		"email" TEXT UNIQUE
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}

func v1CreateTransactionsTable(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "bot::transaction" (
		"id" SERIAL PRIMARY KEY,
		"tgChatId" NUMERIC REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"created" TIMESTAMP NOT NULL DEFAULT NOW(),
		"value" TEXT,
		"archived" BOOLEAN DEFAULT FALSE NOT NULL
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}

func v1CreateValueCache(db *sql.Tx) {
	sqlStatement := `
	CREATE TABLE "bot::cache" (
		"id" SERIAL PRIMARY KEY,
		"tgChatId" NUMERIC REFERENCES "auth::user" ("tgChatId") NOT NULL,
		"lastUsed" TIMESTAMP NOT NULL DEFAULT NOW(),
		"type" TEXT,
		"value" TEXT
	);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
}
