package postgres

import (
	"database/sql"

	"github.com/LucaBernstein/beancount-bot-tg/db/migrations/generic"
)

func (c *Controller) V13(db *sql.Tx) {
	generic.V13AddSettingEnableApi(db)
}
