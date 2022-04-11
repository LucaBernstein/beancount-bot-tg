package helpers

import (
	"strings"
	"testing"
)

func TestExpect(t *testing.T, e1 interface{}, e2 interface{}, msg string) {
	if e1 != e2 {
		errorMsg := "Expected '%v' to be '%v'"
		if msg != "" {
			errorMsg = "%s -> " + errorMsg
		} else {
			errorMsg = "%s" + errorMsg
		}
		t.Errorf(errorMsg, msg, e1, e2)
	}
}

func TestStringContains(t *testing.T, s, substring, msg string) {
	if !strings.Contains(s, substring) {
		errorMsg := "Expected '%s' to contain '%s'"
		if msg != "" {
			errorMsg = "%s -> " + errorMsg
		} else {
			errorMsg = "%s" + errorMsg
		}
		t.Errorf(errorMsg, msg, s, substring)
	}
}

func TestExpectArrEq(t *testing.T, e1, e2 []string, msg string) {
	if !ArraysEqual(e1, e2) {
		errorMsg := "Expected '%v' to be '%v'"
		if msg != "" {
			errorMsg = "%s -> " + errorMsg
		}
		t.Errorf(errorMsg, msg, e1, e2)
	}
}
