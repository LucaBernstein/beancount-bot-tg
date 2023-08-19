package helpers_test

import (
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
)

func TestFqCacheKey(t *testing.T) {
	helpers.TestExpect(t, helpers.FqCacheKey("desc"), "desc:", "")
	helpers.TestExpect(t, helpers.FqCacheKey("desc:"), "desc:", "")
	helpers.TestExpect(t, helpers.FqCacheKey("desc:test"), "desc:test", "")
	helpers.TestExpect(t, helpers.FqCacheKey("desc:test:"), "desc:test", "")
	helpers.TestExpect(t, helpers.FqCacheKey("desc:test:abc"), "desc:test", "")
}
