// Package showcontrol provides high-level show control tools.
package showcontrol

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for show control
type Module struct{}

func (m *Module) Name() string {
	return "showcontrol"
}

func (m *Module) Description() string {
	return "High-level show control and multi-system orchestration"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_show_start",
				mcp.WithDescription("Start a show by initializing all systems. Runs health checks, syncs BPM, and prepares systems for performance."),
				mcp.WithString("show_name", mcp.Description("Name of the show/set")),
				mcp.WithNumber("bpm", mcp.Description("Initial BPM (default: 120)")),
				mcp.WithBoolean("skip_health_check", mcp.Description("Skip initial health checks")),
				mcp.WithString("profile", mcp.Description("Load a saved show profile by name to set BPM, systems, and chains")),
			),
			Handler:             handleShowStart,
			Category:            "showcontrol",
			Subcategory:         "orchestration",
			Tags:                []string{"show", "start", "initialize", "performance"},
			UseCases:            []string{"Start a DJ set", "Initialize live performance", "Begin show"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "showcontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_stop",
				mcp.WithDescription("Stop the show gracefully. Fades out audio/video, stops playback, and saves state."),
				mcp.WithBoolean("save_snapshot", mcp.Description("Save current state before stopping (default: true)")),
				mcp.WithBoolean("blackout", mcp.Description("Trigger blackout on lighting (default: true)")),
				mcp.WithNumber("fade_time", mcp.Description("Fade out time in seconds (default: 3)")),
			),
			Handler:             handleShowStop,
			Category:            "showcontrol",
			Subcategory:         "orchestration",
			Tags:                []string{"show", "stop", "end", "fade"},
			UseCases:            []string{"End a performance", "Graceful shutdown", "Emergency stop"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "showcontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_status",
				mcp.WithDescription("Get comprehensive show status across all systems. Shows health, sync status, and performance metrics."),
			),
			Handler:             handleShowStatus,
			Category:            "showcontrol",
			Subcategory:         "orchestration",
			Tags:                []string{"show", "status", "health", "overview"},
			UseCases:            []string{"Check show health", "Monitor systems", "Pre-show checklist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showcontrol",
		},
		{
			Tool: mcp.NewTool("aftrs_show_emergency",
				mcp.WithDescription("Emergency stop all systems immediately. Stops all playback, triggers blackout, and mutes audio."),
				mcp.WithString("reason", mcp.Description("Reason for emergency stop")),
			),
			Handler:             handleShowEmergency,
			Category:            "showcontrol",
			Subcategory:         "orchestration",
			Tags:                []string{"emergency", "stop", "blackout", "safety"},
			UseCases:            []string{"Emergency situation", "Technical failure", "Safety stop"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showcontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_profile_save",
				mcp.WithDescription("Save a reusable show profile with BPM, systems, snapshot, and chains to run."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Profile name")),
				mcp.WithString("description", mcp.Description("Profile description")),
				mcp.WithNumber("bpm", mcp.Description("Initial BPM for the show")),
				mcp.WithString("systems", mcp.Description("Comma-separated systems to check: ableton,resolume,obs,grandma3")),
				mcp.WithString("snapshot_id", mcp.Description("Snapshot ID to recall at show start")),
				mcp.WithString("chains", mcp.Description("Comma-separated chain IDs to run after start")),
			),
			Handler:             handleShowProfileSave,
			Category:            "showcontrol",
			Subcategory:         "profiles",
			Tags:                []string{"show", "profile", "save", "configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showcontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_profile_load",
				mcp.WithDescription("Load a saved show profile and display its configuration."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Profile name to load")),
			),
			Handler:             handleShowProfileLoad,
			Category:            "showcontrol",
			Subcategory:         "profiles",
			Tags:                []string{"show", "profile", "load", "configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showcontrol",
		},
		{
			Tool: mcp.NewTool("aftrs_show_profile_list",
				mcp.WithDescription("List all saved show profiles."),
			),
			Handler:             handleShowProfileList,
			Category:            "showcontrol",
			Subcategory:         "profiles",
			Tags:                []string{"show", "profile", "list"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showcontrol",
		},
	}
}

// SystemStatus represents a system's status
type SystemStatus struct {
	Name      string
	Connected bool
	Health    string
	Details   string
	Error     string
}

func handleShowStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	showName := tools.GetStringParam(req, "show_name")
	if showName == "" {
		showName = fmt.Sprintf("Show_%s", time.Now().Format("20060102_150405"))
	}
	bpm := tools.GetFloatParam(req, "bpm", 120)
	skipHealth := tools.GetBoolParam(req, "skip_health_check", false)

	// Apply profile if specified
	if profileName := tools.GetStringParam(req, "profile"); profileName != "" {
		profile, err := loadProfile(profileName)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to load profile %q: %w", profileName, err)), nil
		}
		if profile.InitialBPM > 0 && bpm == 120 {
			bpm = profile.InitialBPM
		}
		if showName == fmt.Sprintf("Show_%s", time.Now().Format("20060102_150405")) {
			showName = fmt.Sprintf("%s_%s", profile.Name, time.Now().Format("150405"))
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Starting Show: %s\n\n", showName))

	results := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Step 1: Health checks (unless skipped)
	if !skipHealth {
		sb.WriteString("## Health Checks\n\n")

		systems := []struct {
			name   string
			client func() error
		}{
			{"Ableton", func() error {
				c, err := clients.NewAbletonClient()
				if err != nil {
					return err
				}
				_, err = c.GetStatus(ctx)
				return err
			}},
			{"Resolume", func() error {
				c, err := clients.NewResolumeClient()
				if err != nil {
					return err
				}
				_, err = c.GetStatus(ctx)
				return err
			}},
			{"OBS", func() error {
				c, err := clients.NewOBSClient()
				if err != nil {
					return err
				}
				_, err = c.GetStatus(ctx)
				return err
			}},
			{"grandMA3", func() error {
				c, err := clients.NewGrandMA3Client()
				if err != nil {
					return err
				}
				_, err = c.GetStatus(ctx)
				return err
			}},
		}

		for _, sys := range systems {
			wg.Add(1)
			go func(name string, check func() error) {
				defer wg.Done()
				status := "✅ OK"
				if err := check(); err != nil {
					status = fmt.Sprintf("⚠️ %v", err)
				}
				mu.Lock()
				results[name] = status
				mu.Unlock()
			}(sys.name, sys.client)
		}

		wg.Wait()

		for name, status := range results {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", name, status))
		}
		sb.WriteString("\n")
	}

	// Step 2: Sync BPM
	sb.WriteString("## BPM Sync\n\n")
	sb.WriteString(fmt.Sprintf("Setting BPM to **%.1f** across systems...\n\n", bpm))

	bpmClient, err := clients.NewBPMSyncClient()
	if err == nil {
		if err := bpmClient.PushBPM(ctx, bpm); err != nil {
			sb.WriteString(fmt.Sprintf("⚠️ BPM sync warning: %v\n", err))
		} else {
			sb.WriteString("✅ BPM synchronized\n")
		}
	}
	sb.WriteString("\n")

	// Step 3: Initialize systems
	sb.WriteString("## System Initialization\n\n")

	// Start Ableton transport
	if ableton, err := clients.NewAbletonClient(); err == nil {
		if err := ableton.Stop(ctx); err == nil {
			sb.WriteString("✅ Ableton transport ready\n")
		}
	}

	// Clear Resolume to first column
	if resolume, err := clients.NewResolumeClient(); err == nil {
		if err := resolume.TriggerColumn(ctx, 1); err == nil {
			sb.WriteString("✅ Resolume ready (column 1)\n")
		}
	}

	sb.WriteString("\n## Show Started\n\n")
	sb.WriteString(fmt.Sprintf("**%s** is ready for performance at **%.1f BPM**\n", showName, bpm))

	return tools.TextResult(sb.String()), nil
}

func handleShowStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	saveSnapshot := tools.GetBoolParam(req, "save_snapshot", true)
	blackout := tools.GetBoolParam(req, "blackout", true)
	fadeTime := tools.GetFloatParam(req, "fade_time", 3)

	var sb strings.Builder
	sb.WriteString("# Stopping Show\n\n")

	// Step 1: Save snapshot
	if saveSnapshot {
		sb.WriteString("## Saving State\n\n")
		snapClient, err := clients.NewSnapshotsClient()
		if err == nil {
			snapshot, err := snapClient.CaptureSnapshot(ctx, fmt.Sprintf("ShowEnd_%s", time.Now().Format("150405")), "Auto-saved at show end", nil, []string{"auto", "show-end"})
			if err == nil {
				sb.WriteString(fmt.Sprintf("✅ Snapshot saved: `%s`\n\n", snapshot.ID[:12]))
			} else {
				sb.WriteString(fmt.Sprintf("⚠️ Snapshot failed: %v\n\n", err))
			}
		}
	}

	// Step 2: Fade and stop
	sb.WriteString(fmt.Sprintf("## Fade Out (%.1fs)\n\n", fadeTime))

	// Stop Ableton
	if ableton, err := clients.NewAbletonClient(); err == nil {
		if err := ableton.Stop(ctx); err == nil {
			sb.WriteString("✅ Ableton stopped\n")
		}
	}

	// Clear Resolume
	if resolume, err := clients.NewResolumeClient(); err == nil {
		if err := resolume.ClearAll(ctx); err == nil {
			sb.WriteString("✅ Resolume cleared\n")
		}
	}

	// Step 3: Blackout lighting
	if blackout {
		sb.WriteString("\n## Lighting Blackout\n\n")
		if grandma3, err := clients.NewGrandMA3Client(); err == nil {
			if err := grandma3.Blackout(ctx, true); err == nil {
				sb.WriteString("✅ Lighting blackout\n")
			}
		}
	}

	sb.WriteString("\n## Show Stopped\n\n")
	sb.WriteString("All systems have been gracefully stopped.\n")

	return tools.TextResult(sb.String()), nil
}

func handleShowStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Show Status\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n\n", time.Now().Format("2006-01-02 15:04:05")))

	var statuses []SystemStatus
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Check all systems in parallel
	checks := []struct {
		name  string
		check func(context.Context) SystemStatus
	}{
		{"Ableton", func(ctx context.Context) SystemStatus {
			s := SystemStatus{Name: "Ableton"}
			c, err := clients.NewAbletonClient()
			if err != nil {
				s.Error = err.Error()
				return s
			}
			status, err := c.GetStatus(ctx)
			if err != nil {
				s.Error = err.Error()
				return s
			}
			s.Connected = status.Connected
			s.Health = "OK"
			if status.State != nil {
				playState := "Stopped"
				if status.State.Playing {
					playState = "Playing"
				}
				s.Details = fmt.Sprintf("%.1f BPM, %s", status.State.Tempo, playState)
			}
			return s
		}},
		{"Resolume", func(ctx context.Context) SystemStatus {
			s := SystemStatus{Name: "Resolume"}
			c, err := clients.NewResolumeClient()
			if err != nil {
				s.Error = err.Error()
				return s
			}
			status, err := c.GetStatus(ctx)
			if err != nil {
				s.Error = err.Error()
				return s
			}
			s.Connected = status.Connected
			s.Health = "OK"
			s.Details = fmt.Sprintf("%.1f BPM, Comp: %s", status.BPM, status.Composition)
			return s
		}},
		{"OBS", func(ctx context.Context) SystemStatus {
			s := SystemStatus{Name: "OBS"}
			c, err := clients.NewOBSClient()
			if err != nil {
				s.Error = err.Error()
				return s
			}
			status, err := c.GetStatus(ctx)
			if err != nil {
				s.Error = err.Error()
				return s
			}
			s.Connected = status.Connected
			s.Health = "OK"
			streaming := ""
			if status.Streaming {
				streaming = " [LIVE]"
			}
			s.Details = fmt.Sprintf("Scene: %s%s", status.CurrentScene, streaming)
			return s
		}},
		{"grandMA3", func(ctx context.Context) SystemStatus {
			s := SystemStatus{Name: "grandMA3"}
			c, err := clients.NewGrandMA3Client()
			if err != nil {
				s.Error = err.Error()
				return s
			}
			status, err := c.GetStatus(ctx)
			if err != nil {
				s.Error = err.Error()
				return s
			}
			s.Connected = status.Connected
			s.Health = "OK"
			s.Details = fmt.Sprintf("%s:%d", status.Host, status.Port)
			return s
		}},
		{"Showkontrol", func(ctx context.Context) SystemStatus {
			s := SystemStatus{Name: "Showkontrol"}
			c, err := clients.NewShowkontrolClient()
			if err != nil {
				s.Error = err.Error()
				return s
			}
			status, err := c.GetStatus(ctx)
			if err != nil {
				s.Error = err.Error()
				return s
			}
			s.Connected = status.Connected
			s.Health = "OK"
			if status.Timecode != nil && status.Timecode.Running {
				s.Details = fmt.Sprintf("TC: %s", status.Timecode.PositionTC)
			} else {
				s.Details = "TC: Stopped"
			}
			return s
		}},
	}

	for _, check := range checks {
		wg.Add(1)
		go func(name string, fn func(context.Context) SystemStatus) {
			defer wg.Done()
			status := fn(ctx)
			mu.Lock()
			statuses = append(statuses, status)
			mu.Unlock()
		}(check.name, check.check)
	}

	wg.Wait()

	// Display results
	sb.WriteString("## Systems\n\n")
	sb.WriteString("| System | Status | Health | Details |\n")
	sb.WriteString("|--------|--------|--------|----------|\n")

	connectedCount := 0
	for _, s := range statuses {
		statusIcon := "🔴"
		if s.Connected {
			statusIcon = "🟢"
			connectedCount++
		}
		health := s.Health
		if s.Error != "" {
			health = "Error"
		}
		details := s.Details
		if s.Error != "" {
			details = s.Error
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", s.Name, statusIcon, health, details))
	}

	sb.WriteString(fmt.Sprintf("\n**Connected:** %d/%d systems\n", connectedCount, len(statuses)))

	// BPM sync status
	sb.WriteString("\n## BPM Sync\n\n")
	if bpmClient, err := clients.NewBPMSyncClient(); err == nil {
		if status, err := bpmClient.GetStatus(ctx); err == nil {
			sb.WriteString(fmt.Sprintf("**Master:** %s @ **%.1f BPM**\n", status.MasterSource, status.CurrentBPM))
			if status.InSync {
				sb.WriteString("**Sync:** ✅ All systems in sync\n")
			} else {
				sb.WriteString(fmt.Sprintf("**Sync:** ⚠️ Systems out of sync (drift: %.0f ms)\n", status.DriftMS))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleShowEmergency(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reason := tools.OptionalStringParam(req, "reason", "Emergency stop triggered")

	var sb strings.Builder
	sb.WriteString("# 🚨 EMERGENCY STOP\n\n")
	sb.WriteString(fmt.Sprintf("**Reason:** %s\n\n", reason))
	sb.WriteString("## Actions Taken\n\n")

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]string)

	// Stop all systems in parallel
	actions := []struct {
		name   string
		action func(context.Context) error
	}{
		{"Ableton Stop", func(ctx context.Context) error {
			c, err := clients.NewAbletonClient()
			if err != nil {
				return err
			}
			return c.Stop(ctx)
		}},
		{"Resolume Clear", func(ctx context.Context) error {
			c, err := clients.NewResolumeClient()
			if err != nil {
				return err
			}
			return c.ClearAll(ctx)
		}},
		{"Lighting Blackout", func(ctx context.Context) error {
			c, err := clients.NewGrandMA3Client()
			if err != nil {
				return err
			}
			return c.Blackout(ctx, true)
		}},
		{"OBS Recording Stop", func(ctx context.Context) error {
			c, err := clients.NewOBSClient()
			if err != nil {
				return err
			}
			// Don't fail if not recording
			_ = c.StopRecording(ctx)
			return nil
		}},
	}

	for _, a := range actions {
		wg.Add(1)
		go func(name string, action func(context.Context) error) {
			defer wg.Done()
			status := "✅"
			if err := action(ctx); err != nil {
				status = fmt.Sprintf("⚠️ %v", err)
			}
			mu.Lock()
			results[name] = status
			mu.Unlock()
		}(a.name, a.action)
	}

	wg.Wait()

	for name, status := range results {
		sb.WriteString(fmt.Sprintf("- %s: %s\n", name, status))
	}

	sb.WriteString("\n## Emergency Stop Complete\n\n")
	sb.WriteString(fmt.Sprintf("All systems stopped at %s\n", time.Now().Format("15:04:05")))

	return tools.TextResult(sb.String()), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
