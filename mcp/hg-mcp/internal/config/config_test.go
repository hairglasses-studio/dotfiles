package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars to test defaults
	t.Setenv("HOME", "/tmp/test-home")
	t.Setenv("USER", "testuser")
	t.Setenv("AFTRS_VAULT_PATH", "")
	t.Setenv("AFTRS_DATA_DIR", "")
	t.Setenv("PORT", "")
	t.Setenv("MCP_MODE", "")
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("BEATPORT_USERNAME", "")
	t.Setenv("BEATPORT_PASSWORD", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")

	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	if cfg.Home != "/tmp/test-home" {
		t.Errorf("Home = %q, want /tmp/test-home", cfg.Home)
	}
	if cfg.User != "testuser" {
		t.Errorf("User = %q, want testuser", cfg.User)
	}
	if want := filepath.Join("/tmp/test-home", "aftrs-vault"); cfg.AftrsVaultPath != want {
		t.Errorf("AftrsVaultPath = %q, want %q", cfg.AftrsVaultPath, want)
	}
	if want := filepath.Join("/tmp/test-home", ".hg-mcp"); cfg.AftrsDataDir != want {
		t.Errorf("AftrsDataDir = %q, want %q", cfg.AftrsDataDir, want)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.MCPMode != "stdio" {
		t.Errorf("MCPMode = %q, want stdio", cfg.MCPMode)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want text", cfg.LogFormat)
	}
}

func TestLoadExplicitValues(t *testing.T) {
	t.Setenv("HOME", "/home/studio")
	t.Setenv("USER", "dj")
	t.Setenv("AFTRS_VAULT_PATH", "/data/vault")
	t.Setenv("AFTRS_USER", "operator1")
	t.Setenv("AFTRS_DATA_DIR", "/data/hg-mcp")
	t.Setenv("BEATPORT_USERNAME", "bp-user")
	t.Setenv("BEATPORT_PASSWORD", "bp-pass")
	t.Setenv("BEATPORT_ACCESS_TOKEN", "at-123")
	t.Setenv("BEATPORT_REFRESH_TOKEN", "rt-456")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/creds.json")
	t.Setenv("GOOGLE_API_KEY", "AIza-test")
	t.Setenv("MCP_MODE", "sse")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_FORMAT", "json")

	cfg := Load()

	if cfg.Home != "/home/studio" {
		t.Errorf("Home = %q, want /home/studio", cfg.Home)
	}
	if cfg.AftrsVaultPath != "/data/vault" {
		t.Errorf("AftrsVaultPath = %q, want /data/vault", cfg.AftrsVaultPath)
	}
	if cfg.AftrsUser != "operator1" {
		t.Errorf("AftrsUser = %q, want operator1", cfg.AftrsUser)
	}
	if cfg.AftrsDataDir != "/data/hg-mcp" {
		t.Errorf("AftrsDataDir = %q, want /data/hg-mcp", cfg.AftrsDataDir)
	}
	if cfg.BeatportUsername != "bp-user" {
		t.Errorf("BeatportUsername = %q, want bp-user", cfg.BeatportUsername)
	}
	if cfg.BeatportPassword != "bp-pass" {
		t.Errorf("BeatportPassword = %q, want bp-pass", cfg.BeatportPassword)
	}
	if cfg.BeatportAccessToken != "at-123" {
		t.Errorf("BeatportAccessToken = %q, want at-123", cfg.BeatportAccessToken)
	}
	if cfg.BeatportRefreshToken != "rt-456" {
		t.Errorf("BeatportRefreshToken = %q, want rt-456", cfg.BeatportRefreshToken)
	}
	if cfg.GoogleApplicationCredentials != "/path/to/creds.json" {
		t.Errorf("GoogleApplicationCredentials = %q, want /path/to/creds.json", cfg.GoogleApplicationCredentials)
	}
	if cfg.GoogleAPIKey != "AIza-test" {
		t.Errorf("GoogleAPIKey = %q, want AIza-test", cfg.GoogleAPIKey)
	}
	if cfg.MCPMode != "sse" {
		t.Errorf("MCPMode = %q, want sse", cfg.MCPMode)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat = %q, want json", cfg.LogFormat)
	}
}

func TestGetReturnsSingleton(t *testing.T) {
	t.Setenv("HOME", "/tmp/singleton-test")
	t.Setenv("AFTRS_VAULT_PATH", "")
	t.Setenv("AFTRS_DATA_DIR", "")

	loaded := Load()
	got := Get()

	if got != loaded {
		t.Error("Get() did not return the same pointer as Load()")
	}
}

func TestGetBeforeLoad(t *testing.T) {
	// Reset global to nil to test pre-Load behavior
	mu.Lock()
	saved := global
	global = nil
	mu.Unlock()

	defer func() {
		mu.Lock()
		global = saved
		mu.Unlock()
	}()

	got := Get()
	if got != nil {
		t.Error("Get() should return nil before Load()")
	}
}

func TestLoadOverwritesPrevious(t *testing.T) {
	t.Setenv("HOME", "/first")
	t.Setenv("AFTRS_VAULT_PATH", "")
	t.Setenv("AFTRS_DATA_DIR", "")
	Load()

	t.Setenv("HOME", "/second")
	cfg := Load()

	if cfg.Home != "/second" {
		t.Errorf("Home = %q after second Load(), want /second", cfg.Home)
	}

	got := Get()
	if got.Home != "/second" {
		t.Errorf("Get().Home = %q after second Load(), want /second", got.Home)
	}
}

func TestLoadFallsBackToUserHomeDir(t *testing.T) {
	// When HOME is empty, Load should fall back to os.UserHomeDir()
	t.Setenv("HOME", "")
	t.Setenv("AFTRS_VAULT_PATH", "")
	t.Setenv("AFTRS_DATA_DIR", "")

	cfg := Load()

	expected, err := os.UserHomeDir()
	if err != nil {
		t.Skip("os.UserHomeDir() failed, cannot test fallback")
	}

	if cfg.Home != expected {
		t.Errorf("Home = %q with empty $HOME, want %q from UserHomeDir()", cfg.Home, expected)
	}
}
