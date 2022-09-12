package helpers

import "strings"

const (
	FIELD_DATE        = "date"
	FIELD_DESCRIPTION = "description"
	FIELD_AMOUNT      = "amount"
	FIELD_ACCOUNT     = "account"
	FIELD_TAG         = "tag"

	FIELD_ACCOUNT_FROM = "from"
	FIELD_ACCOUNT_TO   = "to"

	DOT_INDENT = 47

	BEANCOUNT_DATE_FORMAT = "2006-01-02"

	USERSET_CUR        = "user.currency"
	USERSET_ADM        = "user.isAdmin"
	USERSET_TAG        = "user.vacationTag"
	USERSET_TZOFF      = "user.tzOffset"

	DEFAULT_CURRENCY = "EUR"

	TG_MAX_MSG_CHAR_LEN = 4096
)

func AllowedSuggestionTypes() []string {
	return []string{
		FIELD_DESCRIPTION,
		FIELD_ACCOUNT,
	}
}

func TypeCacheKey(key string) string {
	return strings.SplitN(key, ":", 2)[0]
}

func FqCacheKey(key string) string {
	if !strings.Contains(key, ":") {
		return key + ":"
	}
	splits := strings.SplitN(key, ":", 3)
	return splits[0] + ":" + splits[1]
}
