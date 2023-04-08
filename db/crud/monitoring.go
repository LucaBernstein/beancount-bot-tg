package crud

import (
	"strings"
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func (r *Repo) HealthGetLogs(lastHours int) (errors int, warnings int, err error) {
	var query string
	if strings.ToUpper(helpers.Env("DB_TYPE")) == "POSTGRES" {
		query = `
			SELECT "level", COUNT(*) "c"
			FROM "app::log"
			WHERE "level" IN ($1, $2)
				AND "created" + INTERVAL '1 hour' * $3 > NOW()
			GROUP BY "level"
		`
	} else {
		query = `
			SELECT "level", COUNT(*) "c"
			FROM "app::log"
			WHERE "level" IN ($1, $2)
				AND datetime("created", '+'||$3||' hour') > datetime()
			GROUP BY "level"
		`
	}
	rows, err := r.db.Query(query, helpers.ERROR, helpers.WARN, lastHours)
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
	offsetTimestamp := time.Now().UTC().Add(-time.Duration(maxDiffHours) * time.Hour)
	var query string
	if strings.ToUpper(helpers.Env("DB_TYPE")) == "POSTGRES" {
		query = `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT SPLIT_PART(l."chat", '/', 1) AS "chat", MAX("created") as "last" 
			FROM "app::log" l
			WHERE "chat" IS NOT NULL
			GROUP BY l."chat"
		) AS "logs" WHERE "last" >= $1
		`
	} else {
		query = `
		SELECT COUNT(*) FROM (
			SELECT DISTINCT SUBSTR(l."chat", 1, INSTR(l."chat", '/')-1) AS "chat", MAX("created") as "last" 
			FROM "app::log" l
			WHERE "chat" IS NOT NULL
		) WHERE "last" >= $1
		`
	}
	rows, err := r.db.Query(query, offsetTimestamp)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&count)
	}
	return
}

func (r *Repo) HealthGetCacheStats() (accTo, accFrom, txDesc, other int, err error) {
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
		// TODO: Refactor: Unify all into account
		case helpers.FqCacheKey(helpers.FIELD_ACCOUNT + ":" + helpers.FIELD_ACCOUNT_TO):
			accTo = count
		case helpers.FqCacheKey(helpers.FIELD_ACCOUNT + ":" + helpers.FIELD_ACCOUNT_FROM):
			accFrom = count
		case helpers.FqCacheKey(helpers.FIELD_DESCRIPTION):
			txDesc = count
		default:
			other += count
		}
	}
	return
}
