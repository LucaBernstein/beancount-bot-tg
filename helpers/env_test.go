package helpers_test

import (
	"os"
	"testing"

	"github.com/LucaBernstein/beancount-bot-tg/helpers"
)

const TEST_ENV_KEY = "BEANCOUNT_BOT_TG_TEST_ENV_KEY"

func TestEnvFb(t *testing.T) {
	os.Setenv(TEST_ENV_KEY, "123")
	result := helpers.Env(TEST_ENV_KEY)
	if result != "123" {
		t.Errorf("Unexpected ENV value: %s != %s", "123", result)
	}

	result = helpers.EnvOrFb(TEST_ENV_KEY+"_FALSE", "NotSet")
	if result != "NotSet" {
		t.Errorf("Unexpected ENV value: %s != %s", "NotSet", result)
	}
}
