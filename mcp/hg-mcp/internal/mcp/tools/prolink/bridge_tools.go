// bridge_tools.go provides MCP tools for the Pro DJ Link to Showkontrol bridge.
// This enables real-time synchronization between XDJ/CDJ decks and Showkontrol
// for beat-synced lighting, track-triggered cues, and genre-based scene changes.
package prolink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/bridge"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// BridgeTools returns the MCP tools for the prolink-showkontrol bridge
func BridgeTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_prolink_bridge_start",
				mcp.WithDescription("Start the Pro DJ Link to Showkontrol bridge for real-time beat sync and cue triggering. Connects XDJ/CDJ playback to lighting control."),
				mcp.WithString("mode",
					mcp.Description("Sync mode: 'master' (follow master deck), 'on_air' (follow mixer channel), 'all' (track all decks)"),
					mcp.DefaultString("master"),
				),
				mcp.WithString("beat_cue",
					mcp.Description("Showkontrol cue ID to fire on beat 1 (downbeat)"),
				),
				mcp.WithString("track_change_cue",
					mcp.Description("Showkontrol cue ID to fire when track changes"),
				),
				mcp.WithNumber("poll_interval_ms",
					mcp.Description("Polling interval in milliseconds (default 200, min 50)"),
					mcp.DefaultNumber(200),
				),
			),
			Handler:     handleBridgeStart,
			Category:    "prolink",
			Subcategory: "bridge",
			Tags:        []string{"prolink", "showkontrol", "bridge", "sync", "lighting", "beat"},
			UseCases:    []string{"Start beat-synced lighting", "Enable track-triggered cues", "Live DJ show sync"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_bridge_stop",
				mcp.WithDescription("Stop the Pro DJ Link to Showkontrol bridge."),
			),
			Handler:     handleBridgeStop,
			Category:    "prolink",
			Subcategory: "bridge",
			Tags:        []string{"prolink", "showkontrol", "bridge", "stop"},
			UseCases:    []string{"Stop sync", "Disable bridge"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_bridge_status",
				mcp.WithDescription("Get current status of the Pro DJ Link to Showkontrol bridge including sync stats and current track info."),
			),
			Handler:     handleBridgeStatus,
			Category:    "prolink",
			Subcategory: "bridge",
			Tags:        []string{"prolink", "showkontrol", "bridge", "status"},
			UseCases:    []string{"Check bridge status", "View sync stats", "Monitor live performance"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_bridge_configure",
				mcp.WithDescription("Configure the Pro DJ Link to Showkontrol bridge. Can update settings while running."),
				mcp.WithString("mode",
					mcp.Description("Sync mode: 'master', 'on_air', or 'all'"),
				),
				mcp.WithString("beat_cue",
					mcp.Description("Showkontrol cue ID to fire on beat 1 (empty to disable)"),
				),
				mcp.WithString("track_change_cue",
					mcp.Description("Showkontrol cue ID to fire when track changes (empty to disable)"),
				),
				mcp.WithString("genre_cue_mapping",
					mcp.Description("JSON object mapping genres to cue IDs, e.g. {\"House\": \"cue_1\", \"Techno\": \"cue_2\"}"),
				),
				mcp.WithBoolean("key_color_mapping",
					mcp.Description("Enable mapping musical key to lighting color"),
				),
			),
			Handler:     handleBridgeConfigure,
			Category:    "prolink",
			Subcategory: "bridge",
			Tags:        []string{"prolink", "showkontrol", "bridge", "configure", "settings"},
			UseCases:    []string{"Update bridge config", "Set genre cues", "Enable key colors"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_key_color",
				mcp.WithDescription("Get the recommended lighting color for a musical key based on the Camelot wheel."),
				mcp.WithString("key",
					mcp.Description("Musical key (e.g. 'Am', 'C', 'F#m', 'Bb')"),
					mcp.Required(),
				),
			),
			Handler:     handleKeyColor,
			Category:    "prolink",
			Subcategory: "bridge",
			Tags:        []string{"prolink", "key", "color", "lighting", "camelot"},
			UseCases:    []string{"Map key to color", "Harmonic lighting", "Camelot wheel colors"},
			Complexity:  tools.ComplexitySimple,
		},
	}
}

func handleBridgeStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetBridge()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	if b.IsRunning() {
		return tools.TextResult("⚠️ Bridge is already running. Use `aftrs_prolink_bridge_status` to check status or `aftrs_prolink_bridge_stop` to restart."), nil
	}

	// Parse configuration
	config := &bridge.BridgeConfig{
		PollIntervalMs: tools.GetIntParam(req, "poll_interval_ms", 200),
	}

	modeStr := tools.GetStringParam(req, "mode")
	switch modeStr {
	case "on_air":
		config.Mode = bridge.BridgeModeOnAir
	case "all":
		config.Mode = bridge.BridgeModeAll
	default:
		config.Mode = bridge.BridgeModeMaster
	}

	config.BeatCue = tools.GetStringParam(req, "beat_cue")
	config.TrackChangeCue = tools.GetStringParam(req, "track_change_cue")

	// Apply configuration
	b.Configure(config)

	// Start the bridge
	if err := b.Start(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start bridge: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link → Showkontrol Bridge Started\n\n")
	sb.WriteString("✅ Bridge is now running\n\n")
	sb.WriteString("## Configuration\n\n")
	sb.WriteString(fmt.Sprintf("- **Mode:** %s\n", config.Mode))
	sb.WriteString(fmt.Sprintf("- **Poll Interval:** %d ms\n", config.PollIntervalMs))

	if config.BeatCue != "" {
		sb.WriteString(fmt.Sprintf("- **Beat Cue:** `%s` (fires on beat 1)\n", config.BeatCue))
	}
	if config.TrackChangeCue != "" {
		sb.WriteString(fmt.Sprintf("- **Track Change Cue:** `%s`\n", config.TrackChangeCue))
	}

	sb.WriteString("\n## What's Happening\n\n")
	sb.WriteString("The bridge is now:\n")
	sb.WriteString("1. Monitoring Pro DJ Link for beat and track data\n")
	sb.WriteString("2. Following the **")
	switch config.Mode {
	case bridge.BridgeModeMaster:
		sb.WriteString("master deck")
	case bridge.BridgeModeOnAir:
		sb.WriteString("on-air deck")
	case bridge.BridgeModeAll:
		sb.WriteString("all decks")
	}
	sb.WriteString("**\n")
	if config.BeatCue != "" || config.TrackChangeCue != "" {
		sb.WriteString("3. Firing Showkontrol cues based on events\n")
	}

	sb.WriteString("\n**Use `aftrs_prolink_bridge_status` to monitor progress.**\n")

	return tools.TextResult(sb.String()), nil
}

func handleBridgeStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetBridge()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	if !b.IsRunning() {
		return tools.TextResult("⚠️ Bridge is not running."), nil
	}

	if err := b.Stop(); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to stop bridge: %w", err)), nil
	}

	status := b.GetStatus()

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link → Showkontrol Bridge Stopped\n\n")
	sb.WriteString("✅ Bridge has been stopped\n\n")
	sb.WriteString("## Session Stats\n\n")
	sb.WriteString(fmt.Sprintf("- **Beats Synced:** %d\n", status.BeatsSynced))
	sb.WriteString(fmt.Sprintf("- **Track Changes:** %d\n", status.TrackChanges))

	if status.LastCueFired != "" {
		sb.WriteString(fmt.Sprintf("- **Last Cue Fired:** `%s` at %s\n", status.LastCueFired, status.LastCueTime.Format("15:04:05")))
	}

	return tools.TextResult(sb.String()), nil
}

func handleBridgeStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetBridge()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	status := b.GetStatus()

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link → Showkontrol Bridge Status\n\n")

	if status.Running {
		sb.WriteString("✅ **Bridge is RUNNING**\n\n")
	} else {
		sb.WriteString("⏸️ **Bridge is STOPPED**\n\n")
		sb.WriteString("Use `aftrs_prolink_bridge_start` to begin synchronization.\n\n")

		// Still show config
		if status.Config != nil {
			sb.WriteString("## Configured Settings\n\n")
			configJSON, _ := json.MarshalIndent(status.Config, "", "  ")
			sb.WriteString("```json\n")
			sb.WriteString(string(configJSON))
			sb.WriteString("\n```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Current State\n\n")
	sb.WriteString(fmt.Sprintf("- **Mode:** %s\n", status.Mode))
	sb.WriteString(fmt.Sprintf("- **Active Player:** %d\n", status.ActivePlayerID))
	sb.WriteString(fmt.Sprintf("- **Current BPM:** %.1f\n", status.CurrentBPM))
	sb.WriteString(fmt.Sprintf("- **Current Beat:** %d/4\n", status.CurrentBeat))

	if status.CurrentTrack != "" {
		sb.WriteString(fmt.Sprintf("- **Current Track:** %s\n", status.CurrentTrack))
	}
	if status.CurrentKey != "" {
		color := bridge.KeyToColor(status.CurrentKey)
		sb.WriteString(fmt.Sprintf("- **Key:** %s (color: %s)\n", status.CurrentKey, color))
	}
	if status.CurrentGenre != "" {
		sb.WriteString(fmt.Sprintf("- **Genre:** %s\n", status.CurrentGenre))
	}

	sb.WriteString("\n## Session Stats\n\n")
	sb.WriteString(fmt.Sprintf("- **Beats Synced:** %d\n", status.BeatsSynced))
	sb.WriteString(fmt.Sprintf("- **Track Changes:** %d\n", status.TrackChanges))

	if status.LastCueFired != "" {
		sb.WriteString(fmt.Sprintf("- **Last Cue Fired:** `%s` at %s\n", status.LastCueFired, status.LastCueTime.Format("15:04:05")))
	}

	if !status.StartedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("- **Running Since:** %s\n", status.StartedAt.Format("15:04:05")))
	}

	if len(status.Errors) > 0 {
		sb.WriteString("\n## Recent Errors\n\n")
		for _, err := range status.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	// Configuration
	if status.Config != nil {
		sb.WriteString("\n## Configuration\n\n")
		sb.WriteString(fmt.Sprintf("- **Poll Interval:** %d ms\n", status.Config.PollIntervalMs))
		if status.Config.BeatCue != "" {
			sb.WriteString(fmt.Sprintf("- **Beat Cue:** `%s`\n", status.Config.BeatCue))
		}
		if status.Config.TrackChangeCue != "" {
			sb.WriteString(fmt.Sprintf("- **Track Change Cue:** `%s`\n", status.Config.TrackChangeCue))
		}
		if status.Config.GenreCueMapping != nil && len(status.Config.GenreCueMapping) > 0 {
			sb.WriteString("- **Genre Cue Mapping:**\n")
			for genre, cue := range status.Config.GenreCueMapping {
				sb.WriteString(fmt.Sprintf("  - %s → `%s`\n", genre, cue))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleBridgeConfigure(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetBridge()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	// Get current config as base
	currentConfig := b.GetConfig()
	if currentConfig == nil {
		currentConfig = &bridge.BridgeConfig{
			Mode:           bridge.BridgeModeMaster,
			PollIntervalMs: 200,
		}
	}

	// Apply updates
	if modeStr := tools.GetStringParam(req, "mode"); modeStr != "" {
		switch modeStr {
		case "master":
			currentConfig.Mode = bridge.BridgeModeMaster
		case "on_air":
			currentConfig.Mode = bridge.BridgeModeOnAir
		case "all":
			currentConfig.Mode = bridge.BridgeModeAll
		}
	}

	// Handle string params that can be explicitly cleared
	if tools.HasParam(req, "beat_cue") {
		currentConfig.BeatCue = tools.GetStringParam(req, "beat_cue")
	}
	if tools.HasParam(req, "track_change_cue") {
		currentConfig.TrackChangeCue = tools.GetStringParam(req, "track_change_cue")
	}

	// Parse genre cue mapping
	if genreMapping := tools.GetStringParam(req, "genre_cue_mapping"); genreMapping != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(genreMapping), &mapping); err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid genre_cue_mapping JSON: %w", err)), nil
		}
		currentConfig.GenreCueMapping = mapping
	}

	// Key color mapping
	if tools.HasParam(req, "key_color_mapping") {
		currentConfig.KeyColorMapping = tools.GetBoolParam(req, "key_color_mapping", false)
	}

	// Apply configuration
	b.Configure(currentConfig)

	var sb strings.Builder
	sb.WriteString("# Bridge Configuration Updated\n\n")
	sb.WriteString("✅ Configuration applied")
	if b.IsRunning() {
		sb.WriteString(" (bridge is running)\n\n")
	} else {
		sb.WriteString(" (bridge is stopped)\n\n")
	}

	sb.WriteString("## Current Configuration\n\n")
	configJSON, _ := json.MarshalIndent(currentConfig, "", "  ")
	sb.WriteString("```json\n")
	sb.WriteString(string(configJSON))
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

func handleKeyColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	color := bridge.KeyToColor(key)

	var sb strings.Builder
	sb.WriteString("# Key to Color Mapping\n\n")
	sb.WriteString(fmt.Sprintf("**Key:** %s\n", key))
	sb.WriteString(fmt.Sprintf("**Color:** %s\n\n", color))

	// Show the full Camelot wheel
	sb.WriteString("## Camelot Wheel Colors\n\n")
	sb.WriteString("### Minor Keys (A)\n")
	sb.WriteString("| Key | Color |\n")
	sb.WriteString("|-----|-------|\n")
	minorKeys := []string{"Abm", "Am", "Bbm", "Bm", "Cm", "C#m", "Dm", "Ebm", "Em", "Fm", "F#m", "Gm"}
	for _, k := range minorKeys {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", k, bridge.KeyToColor(k)))
	}

	sb.WriteString("\n### Major Keys (B)\n")
	sb.WriteString("| Key | Color |\n")
	sb.WriteString("|-----|-------|\n")
	majorKeys := []string{"Ab", "A", "Bb", "B", "C", "C#", "D", "Eb", "E", "F", "F#", "G"}
	for _, k := range majorKeys {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", k, bridge.KeyToColor(k)))
	}

	return tools.TextResult(sb.String()), nil
}
