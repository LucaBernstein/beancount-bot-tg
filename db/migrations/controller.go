package migrations

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func Migrate(db *sql.DB) {
	helpers.LogLocalf(helpers.INFO, nil, "DB schema before migrations: %d", schema(db))

	migrationWrapper(v1, 1)(db)
	migrationWrapper(v2, 2)(db)
	migrationWrapper(v3, 3)(db)
	migrationWrapper(v4, 4)(db)
	migrationWrapper(v5, 5)(db)
	migrationWrapper(v6, 6)(db)
	migrationWrapper(v7, 7)(db)

	helpers.LogLocalf(helpers.INFO, nil, "Migrations ran through. Schema version: %d", schema(db))
}

type MigrationFn func(*sql.Tx)
type MigrationFnTx func(*sql.DB)

func migrationWrapper(migrationFn MigrationFn, thisVersion int) MigrationFnTx {
	return func(db *sql.DB) {
		if schema(db) >= thisVersion {
			return
		}
		helpers.LogLocalf(helpers.INFO, nil, "Performing schema upgrade to version %d", thisVersion)
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Could not start transaction: %v", err)
		}

		migrationFn(tx)

		setSchemaVersion(tx, thisVersion)
		err = tx.Commit()
		if err != nil {
			log.Fatalf("Could not commit migration transaction to version %d: %s", thisVersion, err.Error())
		}
	}
}

func schema(db *sql.DB) int {
	// Check if settings table holding schema version exists
	q := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE  table_schema = 'public'
		AND    table_name   = 'app::setting'
	);
	`
	exists := false
	err := db.QueryRow(q).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		return 0
	}

	// Check schema version
	q = `
	SELECT value
	FROM "app::setting"
	WHERE key = 'schemaVersion';
	`
	// Version value in settings table should not be null.
	version := "0"
	err = db.QueryRow(q).Scan(&version)
	if err != nil {
		log.Fatal(err)
	}
	i, err := strconv.Atoi(version)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func setSchemaVersion(db *sql.Tx, v int) {
	sqlStatement := `
	UPDATE "app::setting"
	SET value = $1
	WHERE key = 'schemaVersion';
	`
	_, err := db.Exec(sqlStatement, v)
	if err != nil {
		log.Fatal(err)
	}
}
