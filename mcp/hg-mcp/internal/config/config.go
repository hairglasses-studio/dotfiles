// Package config provides centralized configuration for hg-mcp.
//
// All environment variables are read once at startup via Load() and accessed
// through the singleton returned by Get(). This replaces scattered os.Getenv
// calls throughout the codebase with typed, documented, default-bearing fields.
package config

import (
	"os"
	"path/filepath"
	"sync"
)

// Config holds all environment-driven configuration for hg-mcp.
// Loaded once at startup via Load(). Fields are read-only after init.
type Config struct {
	// ── System paths ──

	// Home is the user's home directory ($HOME).
	Home string

	// User is the current OS username ($USER).
	User string

	// AftrsVaultPath is the Obsidian vault root ($AFTRS_VAULT_PATH, default: $HOME/aftrs-vault).
	AftrsVaultPath string

	// AftrsUser is the studio operator identity ($AFTRS_USER).
	AftrsUser string

	// AftrsDataDir is the data directory for chains/tasks ($AFTRS_DATA_DIR, default: $HOME/.hg-mcp).
	AftrsDataDir string

	// ── Beatport credentials ──

	// BeatportUsername is the Beatport account username ($BEATPORT_USERNAME).
	BeatportUsername string

	// BeatportPassword is the Beatport account password ($BEATPORT_PASSWORD).
	BeatportPassword string

	// BeatportAccessToken is a pre-existing OAuth access token ($BEATPORT_ACCESS_TOKEN).
	BeatportAccessToken string

	// BeatportRefreshToken is a pre-existing OAuth refresh token ($BEATPORT_REFRESH_TOKEN).
	BeatportRefreshToken string

	// ── Google credentials ──

	// GoogleApplicationCredentials is the path to a service account JSON file
	// ($GOOGLE_APPLICATION_CREDENTIALS).
	GoogleApplicationCredentials string

	// GoogleAPIKey is a Google API key for read-only access ($GOOGLE_API_KEY).
	GoogleAPIKey string

	// ── Server ──

	// MCPMode is the transport mode: stdio, sse, web ($MCP_MODE, default: stdio).
	MCPMode string

	// Port is the HTTP port for non-stdio modes ($PORT, default: 8080).
	Port string

	// LogFormat selects the log handler: "json" or "text" ($LOG_FORMAT, default: text).
	LogFormat string
}

var (
	global *Config
	mu     sync.RWMutex
)

// Get returns the global config singleton. Returns nil if Load() has not been called.
// Callers that run before main() (e.g., init() functions) should use os.Getenv directly
// or the helper functions in env.go.
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return global
}

// GetOrLoad returns the global config singleton, loading it from the current
// environment if needed.
func GetOrLoad() *Config {
	if cfg := Get(); cfg != nil {
		return cfg
	}
	return Load()
}

// Load reads all environment variables and populates the global config.
// Call once at the start of main(). Safe to call multiple times (last write wins).
func Load() *Config {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}

	vaultPath := os.Getenv("AFTRS_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = filepath.Join(home, "aftrs-vault")
	}

	dataDir := os.Getenv("AFTRS_DATA_DIR")
	if dataDir == "" {
		dataDir = filepath.Join(home, ".hg-mcp")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := &Config{
		// System paths
		Home:           home,
		User:           os.Getenv("USER"),
		AftrsVaultPath: vaultPath,
		AftrsUser:      os.Getenv("AFTRS_USER"),
		AftrsDataDir:   dataDir,

		// Beatport
		BeatportUsername:     os.Getenv("BEATPORT_USERNAME"),
		BeatportPassword:     os.Getenv("BEATPORT_PASSWORD"),
		BeatportAccessToken:  os.Getenv("BEATPORT_ACCESS_TOKEN"),
		BeatportRefreshToken: os.Getenv("BEATPORT_REFRESH_TOKEN"),

		// Google
		GoogleApplicationCredentials: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		GoogleAPIKey:                 os.Getenv("GOOGLE_API_KEY"),

		// Server
		MCPMode:   GetEnv("MCP_MODE", "stdio"),
		Port:      port,
		LogFormat: GetEnv("LOG_FORMAT", "text"),
	}

	mu.Lock()
	global = cfg
	mu.Unlock()

	return cfg
}
