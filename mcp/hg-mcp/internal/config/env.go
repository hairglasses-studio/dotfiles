package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnv returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or empty.
func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// GetEnvRequired returns the value of the environment variable named by key.
// It returns an error if the variable is not set or empty.
func GetEnvRequired(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return v, nil
}

// GetEnvInt returns the value of the environment variable named by key as an int,
// or defaultVal if the variable is not set, empty, or not a valid integer.
func GetEnvInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// GetEnvBool returns the value of the environment variable named by key as a bool,
// or defaultVal if the variable is not set, empty, or not a valid boolean.
// Accepts "true", "1", "yes" (case-insensitive) as true; "false", "0", "no" as false.
func GetEnvBool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultVal
	}
	return b
}
