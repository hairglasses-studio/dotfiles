package daemon

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

// Config holds the daemon configuration.
type Config struct {
	ProfileDirs  []string `toml:"profile_dirs"`
	SocketPath   string   `toml:"socket_path"`
	LogLevel     string   `toml:"log_level"`
	DeviceFilter []string `toml:"device_filter"` // glob patterns to include
}

// LoadConfig reads a TOML config file. If path is empty, uses defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := &Config{
		ProfileDirs: defaultProfileDirs(),
		SocketPath:  defaultSocketPath(),
		LogLevel:    "info",
	}

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultProfileDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "makima"),
	}
}

func defaultSocketPath() string {
	switch runtime.GOOS {
	case "linux":
		if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
			return filepath.Join(dir, "mapitall.sock")
		}
		return "/tmp/mapitall.sock"
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Caches", "mapitall", "mapitall.sock")
	default:
		return "/tmp/mapitall.sock"
	}
}
