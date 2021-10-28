package helpers

import "testing"

func TestExpect(t *testing.T, e1 interface{}, e2 interface{}, msg string) {
	if e1 != e2 {
		errorMsg := "Expected %v to be %v."
		if msg != "" {
			errorMsg = "%s. " + errorMsg
		}
		t.Errorf(errorMsg, msg, e1, e2)
	}
}
