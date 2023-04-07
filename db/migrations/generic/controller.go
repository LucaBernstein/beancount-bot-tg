package generic

import (
	"database/sql"
	"log"
	"strconv"
)

type MigrationFn func(*sql.Tx)
type MigrationFnTx func(*sql.DB)

type MigrationsWrapper struct {
	SchemaFn func(*sql.DB) int
}

func (m *MigrationsWrapper) Migrate(migrationFn MigrationFn, thisVersion int) MigrationFnTx {
	return func(db *sql.DB) {
		if m.SchemaFn(db) >= thisVersion {
			return
		}
		log.Printf("Performing schema upgrade to version %d", thisVersion)
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Could not start transaction: %v", err)
		}

		migrationFn(tx)

		SetSchemaVersion(tx, thisVersion)
		err = tx.Commit()
		if err != nil {
			log.Fatalf("Could not commit migration transaction to version %d: %s", thisVersion, err.Error())
		}
	}
}

func Schema(db *sql.DB) int {
	// Check schema version
	q := `
	SELECT value
	FROM "app::setting"
	WHERE key = 'schemaVersion';
	`
	// Version value in settings table should not be null.
	version := "0"
	err := db.QueryRow(q).Scan(&version)
	if err != nil {
		log.Fatal(err)
	}
	i, err := strconv.Atoi(version)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func SetSchemaVersion(db *sql.Tx, v int) {
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
