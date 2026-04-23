package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hairglasses-studio/mcpkit/registry"
)

func setupNotificationControlFixture(t *testing.T, scriptBody string) (string, string) {
	t.Helper()
	root := t.TempDir()
	home := filepath.Join(root, "home")
	dotfilesRoot := filepath.Join(root, "dotfiles")
	scriptDir := filepath.Join(dotfilesRoot, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatalf("mkdir script dir: %v", err)
	}
	scriptPath := filepath.Join(scriptDir, "notification-control.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptBody), 0o755); err != nil {
		t.Fatalf("write notification-control stub: %v", err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(root, "state"))
	t.Setenv("DOTFILES_DIR", dotfilesRoot)
	t.Setenv("DOTFILES_NOTIFICATION_CONTROL", scriptPath)
	return root, scriptPath
}

func callNotificationTool(t *testing.T, name string, args map[string]any) string {
	t.Helper()
	td := findModuleTool(t, &NotificationModule{}, name)
	req := registry.CallToolRequest{}
	req.Params.Arguments = args
	result, err := td.Handler(context.Background(), req)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if result == nil || result.IsError {
		t.Fatalf("unexpected error result: %#v", result)
	}
	return extractTextFromResult(t, result)
}

func TestNotificationHistoryUsesControlStatus(t *testing.T) {
	setupNotificationControlFixture(t, `#!/usr/bin/env bash
if [[ "$1" == "status" ]]; then
  printf '{"notificationCount":7,"dnd":true}\n'
  exit 0
fi
exit 1
`)

	text := callNotificationTool(t, "notify_history", map[string]any{})
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("unmarshal notify_history: %v; text=%s", err, text)
	}
	if payload["backend"] != "notification-control" {
		t.Fatalf("backend = %v, want notification-control", payload["backend"])
	}
	if payload["count"] != float64(7) {
		t.Fatalf("count = %v, want 7", payload["count"])
	}
	if payload["dnd"] != true {
		t.Fatalf("dnd = %v, want true", payload["dnd"])
	}
}

func TestNotificationActionsUseControlWrapper(t *testing.T) {
	root, _ := setupNotificationControlFixture(t, `#!/usr/bin/env bash
printf '%s\n' "$*" >> "$DOTFILES_NOTIFICATION_LOG"
case "$1" in
  status) printf '{"notificationCount":0,"dnd":false}\n' ;;
esac
`)
	logPath := filepath.Join(root, "notification-control.log")
	t.Setenv("DOTFILES_NOTIFICATION_LOG", logPath)

	_ = callNotificationTool(t, "notify_dnd", map[string]any{"action": "on"})
	_ = callNotificationTool(t, "notify_panel", map[string]any{"action": "open"})
	_ = callNotificationTool(t, "notify_dismiss", map[string]any{"scope": "all"})

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read control log: %v", err)
	}
	log := string(data)
	for _, want := range []string{"dnd-on", "show-center", "dismiss-all"} {
		if !strings.Contains(log, want) {
			t.Fatalf("control log missing %q:\n%s", want, log)
		}
	}
}
