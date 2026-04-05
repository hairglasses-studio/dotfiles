package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	const key = "TEST_GET_ENV_KEY"
	defer os.Unsetenv(key)

	// Not set — returns default
	os.Unsetenv(key)
	if got := GetEnv(key, "default"); got != "default" {
		t.Errorf("GetEnv(%q) = %q, want %q", key, got, "default")
	}

	// Set — returns value
	os.Setenv(key, "custom")
	if got := GetEnv(key, "default"); got != "custom" {
		t.Errorf("GetEnv(%q) = %q, want %q", key, got, "custom")
	}

	// Empty string — returns default
	os.Setenv(key, "")
	if got := GetEnv(key, "fallback"); got != "fallback" {
		t.Errorf("GetEnv(%q) with empty = %q, want %q", key, got, "fallback")
	}
}

func TestGetEnvRequired(t *testing.T) {
	const key = "TEST_GET_ENV_REQUIRED_KEY"
	defer os.Unsetenv(key)

	// Not set — returns error
	os.Unsetenv(key)
	_, err := GetEnvRequired(key)
	if err == nil {
		t.Error("GetEnvRequired should return error when not set")
	}

	// Set — returns value
	os.Setenv(key, "value")
	v, err := GetEnvRequired(key)
	if err != nil {
		t.Errorf("GetEnvRequired returned unexpected error: %v", err)
	}
	if v != "value" {
		t.Errorf("GetEnvRequired = %q, want %q", v, "value")
	}

	// Empty — returns error
	os.Setenv(key, "")
	_, err = GetEnvRequired(key)
	if err == nil {
		t.Error("GetEnvRequired should return error when empty")
	}
}

func TestGetEnvInt(t *testing.T) {
	const key = "TEST_GET_ENV_INT_KEY"
	defer os.Unsetenv(key)

	// Not set — returns default
	os.Unsetenv(key)
	if got := GetEnvInt(key, 42); got != 42 {
		t.Errorf("GetEnvInt(%q) = %d, want %d", key, got, 42)
	}

	// Valid int
	os.Setenv(key, "8080")
	if got := GetEnvInt(key, 42); got != 8080 {
		t.Errorf("GetEnvInt(%q) = %d, want %d", key, got, 8080)
	}

	// Invalid int — returns default
	os.Setenv(key, "not-a-number")
	if got := GetEnvInt(key, 42); got != 42 {
		t.Errorf("GetEnvInt(%q) with invalid = %d, want %d", key, got, 42)
	}
}

func TestGetEnvBool(t *testing.T) {
	const key = "TEST_GET_ENV_BOOL_KEY"
	defer os.Unsetenv(key)

	// Not set — returns default
	os.Unsetenv(key)
	if got := GetEnvBool(key, true); got != true {
		t.Errorf("GetEnvBool(%q) = %v, want %v", key, got, true)
	}

	// "true"
	os.Setenv(key, "true")
	if got := GetEnvBool(key, false); got != true {
		t.Errorf("GetEnvBool(%q, 'true') = %v, want %v", key, got, true)
	}

	// "1"
	os.Setenv(key, "1")
	if got := GetEnvBool(key, false); got != true {
		t.Errorf("GetEnvBool(%q, '1') = %v, want %v", key, got, true)
	}

	// "false"
	os.Setenv(key, "false")
	if got := GetEnvBool(key, true); got != false {
		t.Errorf("GetEnvBool(%q, 'false') = %v, want %v", key, got, false)
	}

	// "0"
	os.Setenv(key, "0")
	if got := GetEnvBool(key, true); got != false {
		t.Errorf("GetEnvBool(%q, '0') = %v, want %v", key, got, false)
	}

	// Invalid — returns default
	os.Setenv(key, "maybe")
	if got := GetEnvBool(key, false); got != false {
		t.Errorf("GetEnvBool(%q, 'maybe') = %v, want %v", key, got, false)
	}
}
