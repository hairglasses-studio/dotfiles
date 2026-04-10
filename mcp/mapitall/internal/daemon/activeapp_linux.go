//go:build linux

package daemon

import (
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// startActiveAppTracker polls for the active window class and updates EngineState.
func (d *Daemon) startActiveAppTracker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		var lastApp string
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				app := detectActiveApp()
				if app != "" && app != lastApp {
					lastApp = app
					d.state.SetActiveApp(app)
					slog.Debug("active app changed", "app", app)
				}
			}
		}
	}()
}

// detectActiveApp tries hyprctl, then swaymsg, then xdotool.
func detectActiveApp() string {
	// Hyprland
	if out, err := exec.Command("hyprctl", "activewindow", "-j").Output(); err == nil {
		var result struct {
			Class string `json:"class"`
		}
		if json.Unmarshal(out, &result) == nil && result.Class != "" {
			return result.Class
		}
	}

	// Sway / wlroots
	if out, err := exec.Command("swaymsg", "-t", "get_tree").Output(); err == nil {
		if cls := parseSwaymsgFocused(out); cls != "" {
			return cls
		}
	}

	// X11
	if out, err := exec.Command("xdotool", "getactivewindow", "getwindowclassname").Output(); err == nil {
		cls := strings.TrimSpace(string(out))
		if cls != "" {
			return cls
		}
	}

	return ""
}

// parseSwaymsgFocused walks the sway tree JSON to find the focused node's app_id.
func parseSwaymsgFocused(data []byte) string {
	var tree struct {
		Nodes []swayNode `json:"nodes"`
	}
	if json.Unmarshal(data, &tree) != nil {
		return ""
	}
	return findFocused(tree.Nodes)
}

type swayNode struct {
	Focused  bool       `json:"focused"`
	AppID    string     `json:"app_id"`
	Nodes    []swayNode `json:"nodes"`
	Floating []swayNode `json:"floating_nodes"`
}

func findFocused(nodes []swayNode) string {
	for _, n := range nodes {
		if n.Focused && n.AppID != "" {
			return n.AppID
		}
		if cls := findFocused(n.Nodes); cls != "" {
			return cls
		}
		if cls := findFocused(n.Floating); cls != "" {
			return cls
		}
	}
	return ""
}
