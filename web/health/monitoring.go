package health

import (
	"fmt"
	"net/http"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

type MonitoringResult struct {
	logs_daily_warning int
	logs_daily_error   int

	transactions_count_open     int
	transactions_count_archived int

	users_count int

	cache_entries_accTo   int
	cache_entries_accFrom int
	cache_entries_txDesc  int
}

// TODO: Use package?
// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus?utm_source=godoc#pkg-overview
func MonitoringEndpoint(bc *bot.BotController) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m := gatherMetrics(bc)
		fmt.Fprintf(w, `
# HELP logs_daily Count of logs of specified type in the previous 24h
# TYPE logs_daily gauge
logs_daily{level="error"} %d
logs_daily{level="warning"} %d

# HELP transactions_count Count of transactions in the database per status
# TYPE transactions_count gauge
transactions_count{type="open"} %d
transactions_count{type="archived"} %d

# HELP users_count Count of unique users by user ID in the database
# TYPE users_count gauge
users_count %d

# HELP cache_entries Count of cache entries per type
# TYPE cache_entries gauge
cache_entries{type="accTo"} %d
cache_entries{type="accFrom"} %d
cache_entries{type="txDesc"} %d
		`,
			m.logs_daily_error,
			m.logs_daily_warning,

			m.transactions_count_open,
			m.transactions_count_archived,

			m.users_count,

			m.cache_entries_accTo,
			m.cache_entries_accFrom,
			m.cache_entries_txDesc,
		)
	}
}

func gatherMetrics(bc *bot.BotController) (result *MonitoringResult) {
	result = &MonitoringResult{}

	errors, warnings, err := bc.Repo.HealthGetLogs(24)
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health logs: %s", err.Error())
	}
	result.logs_daily_error = errors
	result.logs_daily_warning = warnings

	open, archived, err := bc.Repo.HealthGetTransactions()
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health transactions: %s", err.Error())
	}
	result.transactions_count_open = open
	result.transactions_count_archived = archived

	count, err := bc.Repo.HealthGetUserCount()
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health users: %s", err.Error())
	}
	result.users_count = count

	accTo, accFrom, txDesc, err := bc.Repo.HealthGetCacheStats()
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health transactions: %s", err.Error())
	}
	result.cache_entries_accTo = accTo
	result.cache_entries_accFrom = accFrom
	result.cache_entries_txDesc = txDesc

	return
}
