package migrations

import (
	"database/sql"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/generic"
	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/postgres"
	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/sqlite"
)

type MigrationProvider interface {
	Schema(*sql.DB) int
	V1(*sql.Tx)
	V2(*sql.Tx)
	V3(*sql.Tx)
	V4(*sql.Tx)
	V5(*sql.Tx)
	V6(*sql.Tx)
	V7(*sql.Tx)
	V8(*sql.Tx)
	V9(*sql.Tx)
	V10(*sql.Tx)
	V11(*sql.Tx)
	V12(*sql.Tx)
}

func migrate(db *sql.DB, m MigrationProvider) {
	originalSchema := m.Schema(db)
	log.Printf("DB schema before migrations: %d", originalSchema)

	migrationsWrapper := generic.MigrationsWrapper{
		SchemaFn: m.Schema,
	}

	// Insert all migrations here:
	migrationsWrapper.Migrate(m.V1, 1)(db)
	migrationsWrapper.Migrate(m.V2, 2)(db)
	migrationsWrapper.Migrate(m.V3, 3)(db)
	migrationsWrapper.Migrate(m.V4, 4)(db)
	migrationsWrapper.Migrate(m.V5, 5)(db)
	migrationsWrapper.Migrate(m.V6, 6)(db)
	migrationsWrapper.Migrate(m.V7, 7)(db)
	migrationsWrapper.Migrate(m.V8, 8)(db)
	migrationsWrapper.Migrate(m.V9, 9)(db)
	migrationsWrapper.Migrate(m.V10, 10)(db)
	migrationsWrapper.Migrate(m.V11, 11)(db)
	migrationsWrapper.Migrate(m.V12, 12)(db)

	log.Printf("Migrations ran through. Schema version: %d", m.Schema(db))
}

func MigratePostgres(db *sql.DB) {
	migrate(db, &postgres.Controller{})
}

func MigrateSqlite(db *sql.DB) {
	migrate(db, &sqlite.Controller{})
}
