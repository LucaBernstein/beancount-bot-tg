package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/LucaBernstein/beancount-bot-tg/v2/db/migrations"
	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var dbConnection *sql.DB

func Connection() *sql.DB {
	if dbConnection != nil && dbConnection.Ping() == nil {
		return dbConnection
	}
	switch strings.ToUpper(helpers.Env("DB_TYPE")) {
	case "SQLITE":
		log.Print("Using: sqlite...")
		dbConnection = sqliteConnection()
		return dbConnection
	case "POSTGRES":
		log.Print("Using: postgres...")
		dbConnection = postgresConnection()
		return dbConnection
	}
	log.Fatal("Please provide the ENV var DB_TYPE with database type specified")
	return nil
}

func postgresConnection() *sql.DB {
	host := helpers.EnvOrFb("POSTGRES_HOST", "database")
	port := 5432
	user := helpers.EnvOrFb("POSTGRES_USER", "postgres")
	dbname := user
	password := helpers.EnvOrFb("POSTGRES_PASSWORD", "")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Print("Opening postgres db connection...")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Pinging postgres database...")
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Successfully connected to postgres database.")

	migrations.MigratePostgres(db)

	return db
}

func sqliteConnection() *sql.DB {
	filename := helpers.EnvOrFb("SQLITE_FILE", "beancount-bot-tg.sqlite")
	connection := fmt.Sprintf("file:%s?cache=shared", filename)
	db, err := sql.Open("sqlite", connection)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`PRAGMA foreign_keys=ON`)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	migrations.MigrateSqlite(db)

	return db
}
