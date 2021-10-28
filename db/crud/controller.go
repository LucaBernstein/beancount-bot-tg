package crud

import (
	"database/sql"

	tb "gopkg.in/tucnak/telebot.v2"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) EnrichUserData(u tb.User) {
	// TODO: Implement
}
