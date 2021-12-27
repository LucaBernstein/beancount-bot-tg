package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

func TestTestHelpers(t *testing.T) {
	// Just coverage for now
	helpers.TestExpect(&testing.T{}, "this", "that", "should be the same but are different")
	helpers.TestExpectArrEq(&testing.T{}, []string{"this"}, nil, "arrays are different so test fails")
}
