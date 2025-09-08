package env

import (
	"os"
	"strconv"
)

func GetString(key, fallback string) string {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return env
}

func GetInt(key string, fallback int) int {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	i, err := strconv.Atoi(env)
	if err != nil {
		return fallback
	}
	return i
}

func GetBool(key string, fallback bool) bool {
	env, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(env)
	if err != nil {
		return fallback
	}
	return b
}
