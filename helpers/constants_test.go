package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func TestAllowedSuggestionTypes(t *testing.T) {
	types := helpers.AllowedSuggestionTypes()
	if !helpers.ArrayContains(types, helpers.STX_ACCT) {
		t.Errorf("Allowed suggestion types did not contain %s", helpers.STX_ACCT)
	}
}
