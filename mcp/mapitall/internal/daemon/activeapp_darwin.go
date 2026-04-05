//go:build darwin

package daemon

import (
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// startActiveAppTracker polls for the frontmost application bundle ID on macOS.
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

// detectActiveApp uses lsappinfo to get the frontmost application's bundle ID.
func detectActiveApp() string {
	out, err := exec.Command("lsappinfo", "info", "-only", "bundleid", "-app", "front").Output()
	if err != nil {
		return ""
	}
	// Output: "bundleid"="com.example.App"
	s := strings.TrimSpace(string(out))
	if idx := strings.Index(s, "=\""); idx >= 0 {
		s = s[idx+2:]
		if end := strings.Index(s, "\""); end >= 0 {
			return s[:end]
		}
	}
	return ""
}
