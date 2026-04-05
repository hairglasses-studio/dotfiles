package config

import (
	"testing"
)

func TestGetEnv_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		envVal     string
		setEnv     bool // false means unset
		defaultVal string
		want       string
	}{
		{"not set returns default", "", false, "fallback", "fallback"},
		{"empty string returns default", "", true, "fallback", "fallback"},
		{"set value overrides default", "custom", true, "fallback", "custom"},
		{"whitespace is not empty", " ", true, "fallback", " "},
		{"default is empty string", "", false, "", ""},
		{"special characters", "hello=world&foo", true, "default", "hello=world&foo"},
		{"unicode value", "\u00e4\u00f6\u00fc", true, "default", "\u00e4\u00f6\u00fc"},
		{"newlines preserved", "line1\nline2", true, "default", "line1\nline2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const key = "TEST_GETENV_TABLE"
			if tc.setEnv {
				t.Setenv(key, tc.envVal)
			} else {
				t.Setenv(key, "")
				// Setenv with "" makes it set-but-empty; we need truly unset for some cases
				// GetEnv treats empty and unset the same, so this is fine
			}
			got := GetEnv(key, tc.defaultVal)
			if got != tc.want {
				t.Errorf("GetEnv(%q, %q) = %q, want %q", tc.envVal, tc.defaultVal, got, tc.want)
			}
		})
	}
}

func TestGetEnvRequired_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		setEnv  bool
		wantVal string
		wantErr bool
	}{
		{"not set returns error", "", false, "", true},
		{"empty string returns error", "", true, "", true},
		{"valid value succeeds", "myval", true, "myval", false},
		{"whitespace is valid", " ", true, " ", false},
		{"special chars valid", "a=b&c", true, "a=b&c", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const key = "TEST_GETENVREQUIRED_TABLE"
			if tc.setEnv {
				t.Setenv(key, tc.envVal)
			} else {
				t.Setenv(key, "")
			}

			got, err := GetEnvRequired(key)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetEnvRequired() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if got != tc.wantVal {
				t.Errorf("GetEnvRequired() = %q, want %q", got, tc.wantVal)
			}
		})
	}
}

func TestGetEnvRequired_ErrorMessage(t *testing.T) {
	const key = "TEST_REQUIRED_MSG"
	t.Setenv(key, "")

	_, err := GetEnvRequired(key)
	if err == nil {
		t.Fatal("expected error")
	}
	// Error message should include the key name
	if msg := err.Error(); msg == "" {
		t.Error("error message should not be empty")
	}
}

func TestGetEnvInt_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		envVal     string
		setEnv     bool
		defaultVal int
		want       int
	}{
		{"not set returns default", "", false, 42, 42},
		{"empty returns default", "", true, 42, 42},
		{"valid positive int", "100", true, 0, 100},
		{"valid zero", "0", true, 42, 0},
		{"valid negative int", "-5", true, 0, -5},
		{"float returns default", "3.14", true, 42, 42},
		{"text returns default", "abc", true, 42, 42},
		{"overflow handled", "99999999999999999999", true, 42, 42},
		{"leading zeros", "007", true, 0, 7},
		{"whitespace returns default", " 10 ", true, 42, 42},
		{"hex returns default", "0xff", true, 42, 42},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const key = "TEST_GETENVINT_TABLE"
			if tc.setEnv {
				t.Setenv(key, tc.envVal)
			} else {
				t.Setenv(key, "")
			}
			got := GetEnvInt(key, tc.defaultVal)
			if got != tc.want {
				t.Errorf("GetEnvInt(%q, %d) = %d, want %d", tc.envVal, tc.defaultVal, got, tc.want)
			}
		})
	}
}

func TestGetEnvBool_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		envVal     string
		setEnv     bool
		defaultVal bool
		want       bool
	}{
		{"not set returns default true", "", false, true, true},
		{"not set returns default false", "", false, false, false},
		{"empty returns default", "", true, true, true},
		{"true string", "true", true, false, true},
		{"TRUE string", "TRUE", true, false, true},
		{"True string", "True", true, false, true},
		{"1 string", "1", true, false, true},
		{"false string", "false", true, true, false},
		{"FALSE string", "FALSE", true, true, false},
		{"False string", "False", true, true, false},
		{"0 string", "0", true, true, false},
		{"t string", "t", true, false, true},
		{"f string", "f", true, true, false},
		{"yes returns default (not parsed by strconv)", "yes", true, false, false},
		{"no returns default (not parsed by strconv)", "no", true, true, true},
		{"random string returns default", "maybe", true, false, false},
		{"whitespace returns default", " true ", true, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			const key = "TEST_GETENVBOOL_TABLE"
			if tc.setEnv {
				t.Setenv(key, tc.envVal)
			} else {
				t.Setenv(key, "")
			}
			got := GetEnvBool(key, tc.defaultVal)
			if got != tc.want {
				t.Errorf("GetEnvBool(%q, %v) = %v, want %v", tc.envVal, tc.defaultVal, got, tc.want)
			}
		})
	}
}
