package crud

import "github.com/LucaBernstein/beancount-bot-tg/helpers"

func (r *Repo) HealthGetLogs(lastHours int) (errors int, warnings int, err error) {
	rows, err := r.db.Query(`
		SELECT "level", COUNT(*) "c"
		FROM "app::log"
		WHERE "level" IN ($1, $2)
			AND "created" + INTERVAL '1 hour' * $3 > NOW()
		GROUP BY "level"`,
		helpers.ERROR, helpers.WARN, lastHours)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		level int
		count int
	)
	for rows.Next() {
		rows.Scan(&level, &count)
		switch level {
		case int(helpers.ERROR):
			errors = count
		case int(helpers.WARN):
			warnings = count
		}
	}
	return
}

func (r *Repo) HealthGetTransactions() (open int, archived int, err error) {
	rows, err := r.db.Query(`
		SELECT "archived", COUNT(*) "c"
		FROM "bot::transaction"
		GROUP BY "archived"`)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		isArchived bool
		count      int
	)
	for rows.Next() {
		rows.Scan(&isArchived, &count)
		if isArchived {
			archived = count
		} else {
			open = count
		}
	}
	return
}

func (r *Repo) HealthGetUserCount() (count int, err error) {
	rows, err := r.db.Query(`
		SELECT COUNT(*) "c"
		FROM "auth::user"`)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&count)
	}
	return
}

func (r *Repo) HealthGetUsersActiveCounts(maxDiffHours int) (count int, err error) {
	rows, err := r.db.Query(`
		SELECT COUNT("difference") from (
			SELECT DISTINCT EXTRACT(EPOCH FROM (NOW() - MAX(l."created"))) / 3600 AS difference
			FROM "app::log" l JOIN "auth::user" u ON CONCAT('C', u."tgChatId", '/U', u."tgUserId") = l.chat
			GROUP BY l."chat"
		) q
		WHERE "difference" < $1`, maxDiffHours)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&count)
	}
	return
}

func (r *Repo) HealthGetCacheStats() (accTo int, accFrom int, txDesc int, err error) {
	rows, err := r.db.Query(`
		SELECT "type", COUNT(*) "c"
		FROM "bot::cache"
		GROUP BY "type"`)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		t     string
		count int
	)
	for rows.Next() {
		rows.Scan(&t, &count)
		switch t {
		case helpers.STX_ACCT:
			accTo = count
		case helpers.STX_ACCF:
			accFrom = count
		case helpers.STX_DESC:
			txDesc = count
		}
	}
	return
}
