package botTest

import "testing"

func HandleErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %e", err)
	}
}
