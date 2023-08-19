package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
)

func TestTestHelpers(t *testing.T) {
	// Just coverage for now
	helpers.TestExpect(&testing.T{}, "this", "this", "")
	helpers.TestExpect(&testing.T{}, "this", "that", "should be the same but are different")
	helpers.TestExpect(&testing.T{}, "this", "that", "")

	helpers.TestExpectArrEq(&testing.T{}, []string{"this"}, nil, "arrays are different so test fails")

	helpers.TestStringContains(&testing.T{}, "this", "is", "")
	helpers.TestStringContains(&testing.T{}, "this", "are", "")
	helpers.TestStringContains(&testing.T{}, "this", "are", "nope")
}
