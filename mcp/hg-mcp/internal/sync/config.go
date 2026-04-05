package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// YAMLConfig represents the sync.yaml file structure
type YAMLConfig struct {
	Users    []YAMLUserConfig   `yaml:"users"`
	Services YAMLServicesConfig `yaml:"services"`
	Storage  YAMLStorageConfig  `yaml:"storage"`
	Logging  YAMLLoggingConfig  `yaml:"logging"`
}

// YAMLUserConfig represents user configuration in YAML
type YAMLUserConfig struct {
	Username    string   `yaml:"username"`
	DisplayName string   `yaml:"display_name"`
	SoundCloud  bool     `yaml:"soundcloud"`
	Beatport    bool     `yaml:"beatport"`
	Playlists   []string `yaml:"playlists"`
}

// YAMLServicesConfig represents service configurations
type YAMLServicesConfig struct {
	SoundCloud YAMLServiceConfig `yaml:"soundcloud"`
	Beatport   YAMLServiceConfig `yaml:"beatport"`
	Rekordbox  YAMLServiceConfig `yaml:"rekordbox"`
}

// YAMLServiceConfig represents individual service configuration
type YAMLServiceConfig struct {
	Enabled   bool                `yaml:"enabled"`
	Timeout   string              `yaml:"timeout"`
	Workers   int                 `yaml:"workers"`
	RateLimit YAMLRateLimitConfig `yaml:"rate_limit"`
	Retry     YAMLRetryConfig     `yaml:"retry"`
}

// YAMLRateLimitConfig represents rate limit configuration
type YAMLRateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	BurstSize         int `yaml:"burst_size"`
}

// YAMLRetryConfig represents retry configuration
type YAMLRetryConfig struct {
	MaxRetries   int    `yaml:"max_retries"`
	InitialDelay string `yaml:"initial_delay"`
	MaxDelay     string `yaml:"max_delay"`
}

// YAMLStorageConfig represents storage configuration
type YAMLStorageConfig struct {
	LocalRoot  string `yaml:"local_root"`
	S3Bucket   string `yaml:"s3_bucket"`
	AWSProfile string `yaml:"aws_profile"`
	StateFile  string `yaml:"state_file"`
}

// YAMLLoggingConfig represents logging configuration
type YAMLLoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfigFromYAML loads configuration from a YAML file
func LoadConfigFromYAML(path string) (*Config, error) {
	// Expand ~ in path
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("expand home dir: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var yamlConfig YAMLConfig
	if err := yaml.Unmarshal(data, &yamlConfig); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return convertYAMLToConfig(&yamlConfig)
}

// FindConfigFile looks for sync.yaml in standard locations
func FindConfigFile() string {
	// Check common locations
	locations := []string{
		"config/sync.yaml",
		"sync.yaml",
		"~/.config/aftrs/sync.yaml",
	}

	cwd, _ := os.Getwd()
	home, _ := os.UserHomeDir()

	for _, loc := range locations {
		path := loc
		if len(path) > 0 && path[0] == '~' {
			path = filepath.Join(home, path[1:])
		} else if !filepath.IsAbs(path) {
			path = filepath.Join(cwd, path)
		}

		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// LoadConfigOrDefault loads config from YAML if available, otherwise uses defaults
func LoadConfigOrDefault() *Config {
	configPath := FindConfigFile()
	if configPath != "" {
		config, err := LoadConfigFromYAML(configPath)
		if err == nil {
			return config
		}
		// Log warning but continue with defaults
		fmt.Printf("Warning: failed to load config from %s: %v\n", configPath, err)
	}
	return DefaultConfig()
}

// convertYAMLToConfig converts YAML config to internal Config struct
func convertYAMLToConfig(y *YAMLConfig) (*Config, error) {
	home, _ := os.UserHomeDir()

	config := &Config{
		Users:      make([]UserConfig, 0, len(y.Users)),
		LocalRoot:  expandPath(y.Storage.LocalRoot, home),
		S3Bucket:   y.Storage.S3Bucket,
		AWSProfile: y.Storage.AWSProfile,
		DryRun:     false,
	}

	// Set defaults if not specified
	if config.LocalRoot == "" {
		config.LocalRoot = filepath.Join(home, "Music")
	}
	if config.S3Bucket == "" {
		config.S3Bucket = "cr8-music-storage"
	}
	if config.AWSProfile == "" {
		config.AWSProfile = "cr8"
	}

	// Convert users
	for _, u := range y.Users {
		user := UserConfig{
			Username:            u.Username,
			DisplayName:         u.DisplayName,
			SoundCloud:          u.SoundCloud,
			Beatport:            u.Beatport,
			SoundCloudPlaylists: u.Playlists,
		}
		if user.DisplayName == "" {
			user.DisplayName = u.Username
		}
		config.Users = append(config.Users, user)
	}

	// Configure rate limiters from YAML
	if y.Services.SoundCloud.RateLimit.RequestsPerMinute > 0 {
		GlobalRateLimiters.Configure("soundcloud", RateLimitConfig{
			RequestsPerMinute: y.Services.SoundCloud.RateLimit.RequestsPerMinute,
			BurstSize:         y.Services.SoundCloud.RateLimit.BurstSize,
		})
	}

	if y.Services.Beatport.RateLimit.RequestsPerMinute > 0 {
		GlobalRateLimiters.Configure("beatport", RateLimitConfig{
			RequestsPerMinute: y.Services.Beatport.RateLimit.RequestsPerMinute,
			BurstSize:         y.Services.Beatport.RateLimit.BurstSize,
		})
	}

	return config, nil
}

// expandPath expands ~ and returns the path
func expandPath(path, home string) string {
	if path == "" {
		return ""
	}
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(home, path[1:])
	}
	return path
}

// parseDuration parses a duration string like "1s", "5m", "1h"
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}
