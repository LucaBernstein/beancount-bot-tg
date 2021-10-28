package helpers

import (
	"os"
)

func Env(k string) string {
	return EnvOrFb(k, "")
}

func EnvOrFb(k string, fb string) string {
	v := os.Getenv(k)
	if v == "" {
		return fb
	}
	return v
}
