package helpers

import "strings"

func ArrayContains(s []string, e string) bool {
	return ArrayContainsC(s, e, true)
}

func ArrayContainsC(s []string, e string, caseSensitive bool) bool {
	if !caseSensitive {
		e = strings.ToLower(e)
	}
	for _, a := range s {
		if !caseSensitive {
			a = strings.ToLower(a)
		}
		if a == e {
			return true
		}
	}
	return false
}
