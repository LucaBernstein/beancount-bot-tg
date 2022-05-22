package health

import (
	"fmt"
	"net/http"
	"os"

	"github.com/LucaBernstein/beancount-bot-tg/bot"
	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

type MonitoringResult struct {
	logs_daily_warning int
	logs_daily_error   int

	transactions_count_open     int
	transactions_count_archived int

	users_count int

	users_active_last_1d int
	users_active_last_7d int

	cache_entries_accTo   int
	cache_entries_accFrom int
	cache_entries_txDesc  int

	version string
}

// TODO: Use package?
// https://pkg.go.dev/github.com/prometheus/client_golang/prometheus?utm_source=godoc#pkg-overview
func MonitoringEndpoint(bc *bot.BotController) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m := gatherMetrics(bc)
		fmt.Fprintf(w, `
# HELP bc_bot_logs_daily Count of logs of specified type in the previous 24h
# TYPE bc_bot_logs_daily gauge
bc_bot_logs_daily{level="error"} %d
bc_bot_logs_daily{level="warning"} %d

# HELP bc_bot_transactions_count Count of transactions in the database per status
# TYPE bc_bot_transactions_count gauge
bc_bot_transactions_count{type="open"} %d
bc_bot_transactions_count{type="archived"} %d

# HELP bc_bot_users_count Count of unique users by user ID in the database
# TYPE bc_bot_users_count gauge
bc_bot_users_count %d

# HELP users_active_last Count of unique users by user ID active in the previous timeframe
# TYPE users_active_last gauge
users_active_last{timeframe="1d"} %d
users_active_last{timeframe="7d"} %d

# HELP bc_bot_cache_entries Count of cache entries per type
# TYPE bc_bot_cache_entries gauge
bc_bot_cache_entries{type="accTo"} %d
bc_bot_cache_entries{type="accFrom"} %d
bc_bot_cache_entries{type="txDesc"} %d

# HELP bc_bot_version_information
# TYPE bc_bot_version_information gauge
bc_bot_version_information{version="%s"} 1
		`,
			m.logs_daily_error,
			m.logs_daily_warning,

			m.transactions_count_open,
			m.transactions_count_archived,

			m.users_count,

			m.users_active_last_1d,
			m.users_active_last_7d,

			m.cache_entries_accTo,
			m.cache_entries_accFrom,
			m.cache_entries_txDesc,

			m.version,
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

	active_1d, err := bc.Repo.HealthGetUsersActiveCounts(1)
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health users: %s", err.Error())
	}
	result.users_active_last_1d = active_1d
	active_7d, err := bc.Repo.HealthGetUsersActiveCounts(7)
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health users: %s", err.Error())
	}
	result.users_active_last_7d = active_7d

	accTo, accFrom, txDesc, err := bc.Repo.HealthGetCacheStats()
	if err != nil {
		bc.Logf(helpers.ERROR, nil, "Error getting health transactions: %s", err.Error())
	}
	result.cache_entries_accTo = accTo
	result.cache_entries_accFrom = accFrom
	result.cache_entries_txDesc = txDesc

	result.version = os.Getenv("VERSION")

	return
}
