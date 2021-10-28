package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
	_ "github.com/lib/pq"
)

func PostgresConnection() *sql.DB {
	host := helpers.EnvOrFb("POSTGRES_HOST", "database")
	port := 5432
	user := helpers.EnvOrFb("POSTGRES_USER", "postgres")
	dbname := user
	password := helpers.EnvOrFb("POSTGRES_PASSWORD", "")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Print("Opening db (postgres) connection...")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Pinging database (postgres)...")
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Successfully connected to database (postgres)")

	migrations.Migrate(db)

	return db
}
