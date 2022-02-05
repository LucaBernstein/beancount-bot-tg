package helpers

const (
	STX_DESC = "txDesc"
	STX_DATE = "txDate"
	STX_ACCF = "accFrom"
	STX_AMTF = "amountFrom"
	STX_ACCT = "accTo"

	DOT_INDENT = 47

	BEANCOUNT_DATE_FORMAT = "2006-01-02"

	USERSET_CUR        = "user.currency"
	USERSET_ADM        = "user.isAdmin"
	USERSET_TAG        = "user.vacationTag"
	USERSET_LIM_PREFIX = "user.limitCache."
)

func AllowedSuggestionTypes() []string {
	return []string{
		STX_ACCF,
		STX_ACCT,
		STX_DESC,
	}
}
