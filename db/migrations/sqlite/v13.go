package sqlite

import (
	"database/sql"

	"github.com/LucaBernstein/beancount-bot-tg/v2/db/migrations/generic"
)

func (c *Controller) V13(db *sql.Tx) {
	generic.V13AddSettingEnableApi(db)
}
