package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// AutoSnapshotConfig defines the configuration for automatic snapshots.
type AutoSnapshotConfig struct {
	Enabled    bool     `json:"enabled"`
	Triggers   []string `json:"triggers"`    // "pre_show", "post_show", "interval"
	IntervalMs int      `json:"interval_ms"` // for "interval" trigger, in milliseconds
	Systems    []string `json:"systems"`     // systems to capture, empty = all
	MaxKeep    int      `json:"max_keep"`    // max auto-snapshots to retain, 0 = unlimited
	NamePrefix string   `json:"name_prefix"` // prefix for auto-snapshot names
}

var (
	autoConfig   *AutoSnapshotConfig
	autoConfigMu sync.RWMutex
)

func init() {
	autoConfig = &AutoSnapshotConfig{
		Enabled:    false,
		MaxKeep:    10,
		NamePrefix: "auto",
	}
	// Load persisted config if it exists
	if cfg, err := loadAutoConfig(); err == nil {
		autoConfig = cfg
	}
}

func autoConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".hg-mcp", "auto_snapshots.json")
}

func loadAutoConfig() (*AutoSnapshotConfig, error) {
	data, err := os.ReadFile(autoConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg AutoSnapshotConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveAutoConfig(cfg *AutoSnapshotConfig) error {
	path := autoConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// handleAutoSnapshotConfigure handles the aftrs_snapshot_auto_configure tool.
func handleAutoSnapshotConfigure(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "status")

	switch action {
	case "status":
		autoConfigMu.RLock()
		cfg := *autoConfig
		autoConfigMu.RUnlock()
		return tools.JSONResult(cfg), nil

	case "enable":
		triggers := tools.GetStringParam(req, "triggers")
		intervalMs := tools.GetIntParam(req, "interval_ms", 0)
		systems := tools.GetStringParam(req, "systems")
		maxKeep := tools.GetIntParam(req, "max_keep", 10)
		namePrefix := tools.GetStringParam(req, "name_prefix")

		autoConfigMu.Lock()
		autoConfig.Enabled = true
		if triggers != "" {
			autoConfig.Triggers = splitAndTrim(triggers)
		}
		if intervalMs > 0 {
			autoConfig.IntervalMs = intervalMs
		}
		if systems != "" {
			autoConfig.Systems = splitAndTrim(systems)
		}
		if maxKeep > 0 {
			autoConfig.MaxKeep = maxKeep
		}
		if namePrefix != "" {
			autoConfig.NamePrefix = namePrefix
		}
		cfg := *autoConfig
		autoConfigMu.Unlock()

		if err := saveAutoConfig(&cfg); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to save config: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"message": "Auto-snapshot enabled",
			"config":  cfg,
		}), nil

	case "disable":
		autoConfigMu.Lock()
		autoConfig.Enabled = false
		cfg := *autoConfig
		autoConfigMu.Unlock()

		if err := saveAutoConfig(&cfg); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to save config: %w", err)), nil
		}

		return tools.TextResult("Auto-snapshot disabled"), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use status, enable, or disable)", action)), nil
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
