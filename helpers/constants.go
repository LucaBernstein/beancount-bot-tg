package helpers

const (
	STX_DESC = "txDesc"
	STX_DATE = "txDate"
	STX_ACCF = "accFrom"
	STX_AMTF = "amountFrom"
	STX_ACCT = "accTo"
)

func AllowedSuggestionTypes() []string {
	return []string{
		STX_ACCF,
		STX_ACCT,
		STX_DESC,
	}
}
