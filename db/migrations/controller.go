package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

func Migrate(db *sql.DB) {
	fmt.Println("DB schema before migrations:", schema(db))

	migrationWrapper(v1, 1)(db)
	migrationWrapper(v2, 2)(db)
	migrationWrapper(v3, 3)(db)

	fmt.Println("Migrations ran through. Schema version:", schema(db))
}

type MigrationFn func(*sql.Tx)
type MigrationFnTx func(*sql.DB)

func migrationWrapper(migrationFn MigrationFn, thisVersion int) MigrationFnTx {
	return func(db *sql.DB) {
		if schema(db) >= thisVersion {
			return
		}
		fmt.Println("Performing schema upgrade to version", thisVersion)
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Could not start transaction: %v", err)
		}

		migrationFn(tx)

		setSchemaVersion(tx, thisVersion)
		err = tx.Commit()
		if err != nil {
			log.Fatalf("Could not commit migration transaction to version %v: %v", thisVersion, err)
		}
	}
}

func schema(db *sql.DB) int {
	// Check if settings table holding schema version exists
	sql := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE  table_schema = 'public'
		AND    table_name   = 'app::setting'
	);
	`
	exists := false
	err := db.QueryRow(sql).Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		return 0
	}

	// Check schema version
	sql = `
	SELECT value
	FROM "app::setting"
	WHERE key = 'schemaVersion';
	`
	version := "0"
	err = db.QueryRow(sql).Scan(&version)
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
