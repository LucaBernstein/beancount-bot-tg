package crud

import (
	dbWrapper "github.com/LucaBernstein/beancount-bot-tg/db"
)

type Repo struct {
	db dbWrapper.DB
}

func NewRepo(db dbWrapper.DB) *Repo {
	return &Repo{
		db: db,
	}
}
