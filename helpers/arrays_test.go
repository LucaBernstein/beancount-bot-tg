package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
)

func TestArrayMatching(t *testing.T) {
	if helpers.ArrayContains([]string{"a", "b", "c"}, "d") {
		t.Error("'d' should not be found in array")
	}

	if !helpers.ArrayContains([]string{"a", "b", "c"}, "c") {
		t.Error("'c' should be found in array")
	}

	if !helpers.ArrayContainsC([]string{"a", "b", "big"}, "BIG", false) {
		t.Error("'BIG' (ci) should be found in array")
	}

	if helpers.ArraysEqual([]string{"a"}, []string{"b"}) {
		t.Error("ArraysEqual should fail for different arrays")
	}
}
