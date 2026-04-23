// mod_notification.go — Notification control via Quickshell/swaync wrappers
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hairglasses-studio/mcpkit/handler"
	"github.com/hairglasses-studio/mcpkit/registry"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const swayncCmd = "swaync-client"
const swayncTimeout = 5 * time.Second
const notificationControlTimeout = 5 * time.Second

func notificationControlPath() string {
	if path := os.Getenv("DOTFILES_NOTIFICATION_CONTROL"); path != "" {
		return path
	}
	return filepath.Join(dotfilesDir(), "scripts", "notification-control.sh")
}

func notificationControlAvailable() bool {
	info, err := os.Stat(notificationControlPath())
	return err == nil && !info.IsDir()
}

func notificationControlRun(args ...string) (string, error) {
	path := notificationControlPath()
	ctx, cancel := context.WithTimeout(context.Background(), notificationControlTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = append(os.Environ(), "QUICKSHELL_IPC_TIMEOUT=0.8")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("%s %s failed: %w: %s", path, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func notificationControlStatus() (map[string]any, error) {
	if !notificationControlAvailable() {
		return nil, fmt.Errorf("notification-control.sh not available")
	}
	out, err := notificationControlRun("status")
	if err != nil {
		return nil, err
	}
	var status map[string]any
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return nil, fmt.Errorf("decode notification-control status: %w", err)
	}
	return status, nil
}

func notificationStatusBool(status map[string]any, key string) (bool, bool) {
	value, ok := status[key]
	if !ok || value == nil {
		return false, false
	}
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "on":
			return true, true
		case "false", "0", "no", "off":
			return false, true
		}
	}
	return false, false
}

func notificationStatusInt(status map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		value, ok := status[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed), true
		case int:
			return typed, true
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(typed)); err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func notificationControlAction(action string) error {
	if !notificationControlAvailable() {
		return fmt.Errorf("notification-control.sh not available")
	}
	_, err := notificationControlRun(action)
	return err
}

// swayncCheckTool checks if swaync-client is available on PATH.
func swayncCheckTool() error {
	_, err := exec.LookPath(swayncCmd)
	if err != nil {
		return fmt.Errorf("%s not found on PATH — install swaync (e.g. pacman -S swaync)", swayncCmd)
	}
	return nil
}

// swayncRunCmd executes a swaync-client command with a timeout and returns
// trimmed stdout. Returns a descriptive error if the process fails.
func swayncRunCmd(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), swayncTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, swayncCmd, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return strings.TrimSpace(string(out)), fmt.Errorf("%s %s failed: %w: %s", swayncCmd, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// ---------------------------------------------------------------------------
// Input types
// ---------------------------------------------------------------------------

// NotifyHistoryInput defines the input for the notify_history tool.
type NotifyHistoryInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"description=Maximum number of notifications to return (0 or omit for all)"`
}

type NotifyHistoryEntriesInput struct {
	Limit            int  `json:"limit,omitempty" jsonschema:"description=Maximum number of entries to return. Defaults to 25."`
	IncludeDismissed bool `json:"include_dismissed,omitempty" jsonschema:"description=Include entries that have already been dismissed or cleared."`
}

type NotifyHistoryEntriesOutput struct {
	Entries       []notificationHistoryEntry `json:"entries"`
	Total         int                        `json:"total"`
	Visible       int                        `json:"visible"`
	BackendReady  bool                       `json:"backend_ready"`
	LogPath       string                     `json:"log_path"`
	ListenerAlive bool                       `json:"listener_alive"`
}

type NotifyHistoryClearInput struct {
	Purge bool `json:"purge,omitempty" jsonschema:"description=When true, remove the stored history entirely instead of marking entries dismissed."`
}

type NotifyHistoryClearOutput struct {
	Cleared int  `json:"cleared"`
	Purged  bool `json:"purged"`
}

// NotifyDNDInput defines the input for the notify_dnd tool.
type NotifyDNDInput struct {
	Action string `json:"action" jsonschema:"required,description=DND action: get (query state) / toggle / on / off,enum=get,enum=toggle,enum=on,enum=off"`
}

// NotifyDismissInput defines the input for the notify_dismiss tool.
type NotifyDismissInput struct {
	Scope string `json:"scope" jsonschema:"required,description=Which notifications to dismiss: all or latest,enum=all,enum=latest"`
}

// NotifyPanelInput defines the input for the notify_panel tool.
type NotifyPanelInput struct {
	Action string `json:"action" jsonschema:"required,description=Panel action: toggle / open / close,enum=toggle,enum=open,enum=close"`
}

// NotifyResult wraps a string result so the MCP response is a JSON object.
type NotifyResult struct {
	Result string `json:"result"`
}

// ---------------------------------------------------------------------------
// Module
// ---------------------------------------------------------------------------

// NotificationModule provides notification management tools via the repo
// notification-control wrapper, with swaync-client as rollback fallback.
type NotificationModule struct{}

func (m *NotificationModule) Name() string { return "notification" }
func (m *NotificationModule) Description() string {
	return "Notification control via Quickshell/swaync wrappers"
}

func (m *NotificationModule) Tools() []registry.ToolDefinition {
	// ── notify_history ─────────────────────────────────────
	notifyHistory := handler.TypedHandler[NotifyHistoryInput, any](
		"notify_history",
		"Get notification status: count, DND state, local tracked history, and active wrapper backend.",
		func(_ context.Context, _ NotifyHistoryInput) (any, error) {
			entries, _ := readNotificationHistoryEntries()
			tracked := len(entries)
			visible := 0
			for _, entry := range entries {
				if entry.Visible && !entry.Dismissed {
					visible++
				}
			}

			backend := "history"
			count := visible
			dnd := false
			if status, err := notificationControlStatus(); err == nil {
				backend = "notification-control"
				if parsed, ok := notificationStatusInt(status, "notificationCount", "daemon_count", "visible"); ok {
					count = parsed
				}
				if parsed, ok := notificationStatusBool(status, "dnd"); ok {
					dnd = parsed
				}
			} else if swayncCheckTool() == nil {
				backend = "swaync-client"
				countRaw, _ := swayncRunCmd("-c")
				dndRaw, _ := swayncRunCmd("-D")
				if countRaw != "" {
					_, _ = fmt.Sscanf(countRaw, "%d", &count)
				}
				dnd = strings.TrimSpace(dndRaw) == "true"
			}

			return map[string]any{
				"count":                  count,
				"dnd":                    dnd,
				"backend":                backend,
				"tracked_entries":        tracked,
				"tracked_visible":        visible,
				"history_log_path":       dotfilesNotificationHistoryLogPath(),
				"history_listener_alive": notificationHistoryListenerRunning(),
				"note":                   "Detailed history is provided by the local desktop-control log; actions route through notification-control when available.",
			}, nil
		},
	)

	notifyHistoryEntries := handler.TypedHandler[NotifyHistoryEntriesInput, NotifyHistoryEntriesOutput](
		"notify_history_entries",
		"Get the locally logged notification history entries with app, summary, body, urgency, timestamp, and dismissal state.",
		func(_ context.Context, input NotifyHistoryEntriesInput) (NotifyHistoryEntriesOutput, error) {
			entries, err := readNotificationHistoryEntries()
			if err != nil {
				return NotifyHistoryEntriesOutput{}, err
			}

			visible := 0
			filtered := make([]notificationHistoryEntry, 0, len(entries))
			for i := len(entries) - 1; i >= 0; i-- {
				entry := entries[i]
				if entry.Visible && !entry.Dismissed {
					visible++
				}
				if !input.IncludeDismissed && (entry.Dismissed || !entry.Visible) {
					continue
				}
				filtered = append(filtered, entry)
			}

			limit := input.Limit
			if limit <= 0 {
				limit = 25
			}
			if len(filtered) > limit {
				filtered = filtered[:limit]
			}

			return NotifyHistoryEntriesOutput{
				Entries:       filtered,
				Total:         len(entries),
				Visible:       visible,
				BackendReady:  pathExists(dotfilesNotificationHistoryListenerPath()) && hasCmd("python3") && hasCmd("dbus-monitor"),
				LogPath:       dotfilesNotificationHistoryLogPath(),
				ListenerAlive: notificationHistoryListenerRunning(),
			}, nil
		},
	)

	notifyHistoryClear := handler.TypedHandler[NotifyHistoryClearInput, NotifyHistoryClearOutput](
		"notify_history_clear",
		"DESTRUCTIVE: Clear the locally logged notification history. By default entries are marked dismissed; set purge=true to remove them entirely.",
		func(_ context.Context, input NotifyHistoryClearInput) (NotifyHistoryClearOutput, error) {
			cleared, err := clearNotificationHistory(input.Purge)
			if err != nil {
				return NotifyHistoryClearOutput{}, err
			}
			return NotifyHistoryClearOutput{
				Cleared: cleared,
				Purged:  input.Purge,
			}, nil
		},
	)
	notifyHistoryClear.IsWrite = true

	// ── notify_dnd ─────────────────────────────────────────
	notifyDND := handler.TypedHandler[NotifyDNDInput, NotifyResult](
		"notify_dnd",
		"Manage Do Not Disturb mode through Quickshell/swaync wrappers. Actions: get, toggle, on, off.",
		func(_ context.Context, input NotifyDNDInput) (NotifyResult, error) {
			switch input.Action {
			case "get":
				if status, err := notificationControlStatus(); err == nil {
					enabled, _ := notificationStatusBool(status, "dnd")
					if enabled {
						return NotifyResult{Result: "DND is enabled"}, nil
					}
					return NotifyResult{Result: "DND is disabled"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				out, err := swayncRunCmd("-D")
				if err != nil {
					return NotifyResult{}, fmt.Errorf("failed to get DND state: %w", err)
				}
				enabled := strings.ToLower(out) == "true"
				if enabled {
					return NotifyResult{Result: "DND is enabled"}, nil
				}
				return NotifyResult{Result: "DND is disabled"}, nil

			case "toggle":
				if err := notificationControlAction("toggle-dnd"); err == nil {
					return NotifyResult{Result: "DND toggled"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				out, err := swayncRunCmd("-d")
				if err != nil {
					return NotifyResult{}, fmt.Errorf("failed to toggle DND: %w", err)
				}
				return NotifyResult{Result: fmt.Sprintf("DND toggled (new state: %s)", out)}, nil

			case "on":
				if err := notificationControlAction("dnd-on"); err == nil {
					return NotifyResult{Result: "DND enabled"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-dn"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to enable DND: %w", err)
				}
				return NotifyResult{Result: "DND enabled"}, nil

			case "off":
				if err := notificationControlAction("dnd-off"); err == nil {
					return NotifyResult{Result: "DND disabled"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-df"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to disable DND: %w", err)
				}
				return NotifyResult{Result: "DND disabled"}, nil

			default:
				return NotifyResult{}, fmt.Errorf("[%s] action must be one of: get, toggle, on, off", handler.ErrInvalidParam)
			}
		},
	)
	notifyDND.IsWrite = true

	// ── notify_dismiss ─────────────────────────────────────
	notifyDismiss := handler.TypedHandler[NotifyDismissInput, NotifyResult](
		"notify_dismiss",
		"DESTRUCTIVE: Dismiss notifications through Quickshell/swaync wrappers. Scope: all or latest.",
		func(_ context.Context, input NotifyDismissInput) (NotifyResult, error) {
			switch input.Scope {
			case "all":
				if err := notificationControlAction("dismiss-all"); err == nil {
					_, _ = markNotificationHistoryDismissed(0)
					return NotifyResult{Result: "All notifications dismissed"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-C"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to dismiss all notifications: %w", err)
				}
				_, _ = markNotificationHistoryDismissed(0)
				return NotifyResult{Result: "All notifications dismissed"}, nil

			case "latest":
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("--close-latest"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to dismiss latest notification: %w", err)
				}
				_, _ = markNotificationHistoryDismissed(1)
				return NotifyResult{Result: "Latest notification dismissed"}, nil

			default:
				return NotifyResult{}, fmt.Errorf("[%s] scope must be one of: all, latest", handler.ErrInvalidParam)
			}
		},
	)
	notifyDismiss.IsWrite = true

	// ── notify_count ───────────────────────────────────────
	notifyCount := handler.TypedHandler[struct{}, NotifyResult](
		"notify_count",
		"Get the current notification count from notification-control or swaync fallback.",
		func(_ context.Context, _ struct{}) (NotifyResult, error) {
			if status, err := notificationControlStatus(); err == nil {
				if count, ok := notificationStatusInt(status, "notificationCount", "daemon_count", "visible"); ok {
					return NotifyResult{Result: fmt.Sprintf("%d", count)}, nil
				}
			}
			if err := swayncCheckTool(); err != nil {
				return NotifyResult{}, err
			}

			out, err := swayncRunCmd("-c")
			if err != nil {
				return NotifyResult{}, fmt.Errorf("failed to get notification count: %w", err)
			}

			// Validate it's actually a number.
			count, err := strconv.Atoi(out)
			if err != nil {
				return NotifyResult{}, fmt.Errorf("unexpected output from swaync-client -c: %q", out)
			}

			return NotifyResult{Result: fmt.Sprintf("%d", count)}, nil
		},
	)

	// ── notify_panel ───────────────────────────────────────
	notifyPanel := handler.TypedHandler[NotifyPanelInput, NotifyResult](
		"notify_panel",
		"DESTRUCTIVE: Control the notification panel visibility through Quickshell/swaync wrappers. Actions: toggle, open, close.",
		func(_ context.Context, input NotifyPanelInput) (NotifyResult, error) {
			switch input.Action {
			case "toggle":
				if err := notificationControlAction("toggle-center"); err == nil {
					return NotifyResult{Result: "Notification panel toggled"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-t"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to toggle panel: %w", err)
				}
				return NotifyResult{Result: "Notification panel toggled"}, nil

			case "open":
				if err := notificationControlAction("show-center"); err == nil {
					return NotifyResult{Result: "Notification panel opened"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-op"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to open panel: %w", err)
				}
				return NotifyResult{Result: "Notification panel opened"}, nil

			case "close":
				if err := notificationControlAction("hide-center"); err == nil {
					return NotifyResult{Result: "Notification panel closed"}, nil
				}
				if err := swayncCheckTool(); err != nil {
					return NotifyResult{}, err
				}
				if _, err := swayncRunCmd("-cp"); err != nil {
					return NotifyResult{}, fmt.Errorf("failed to close panel: %w", err)
				}
				return NotifyResult{Result: "Notification panel closed"}, nil

			default:
				return NotifyResult{}, fmt.Errorf("[%s] action must be one of: toggle, open, close", handler.ErrInvalidParam)
			}
		},
	)
	notifyPanel.IsWrite = true

	return []registry.ToolDefinition{
		notifyHistory,
		notifyHistoryEntries,
		notifyHistoryClear,
		notifyDND,
		notifyDismiss,
		notifyCount,
		notifyPanel,
	}
}
