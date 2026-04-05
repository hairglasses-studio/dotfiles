// Package avsync provides MCP tools for audio-visual synchronization between music
// metadata (BPM, key, genre, energy) and visual systems (Resolume, TouchDesigner).
package avsync

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for AV sync tools
type Module struct{}

func (m *Module) Name() string {
	return "avsync"
}

func (m *Module) Description() string {
	return "Audio-visual sync tools for bridging music metadata (BPM, key, genre, energy) to visual systems (Resolume, TouchDesigner). Automates visual parameter control based on track characteristics."
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Status
		{
			Tool: mcp.NewTool("aftrs_avsync_status",
				mcp.WithDescription("Get AV bridge connection status including Resolume and TouchDesigner connectivity, live mode state, and loaded mapping count"),
			),
			Handler:             handleStatus,
			Category:            "avsync",
			Subcategory:         "status",
			Tags:                []string{"av", "sync", "status", "resolume", "touchdesigner"},
			UseCases:            []string{"Check visual system connectivity", "Verify AV bridge status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// BPM to Resolume
		{
			Tool: mcp.NewTool("aftrs_avsync_bpm_to_resolume",
				mcp.WithDescription("Push BPM to Resolume's tempo controller. Syncs visual playback speed with music tempo."),
				mcp.WithNumber("bpm",
					mcp.Required(),
					mcp.Description("BPM value (20-999)"),
				),
			),
			Handler:             handleBPMToResolume,
			Category:            "avsync",
			Subcategory:         "bpm",
			Tags:                []string{"av", "sync", "bpm", "resolume", "tempo"},
			UseCases:            []string{"Sync Resolume to track BPM", "Match visual tempo to music"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// BPM to TouchDesigner
		{
			Tool: mcp.NewTool("aftrs_avsync_bpm_to_td",
				mcp.WithDescription("Push BPM to TouchDesigner project variables. Sets current_bpm variable for project-wide tempo sync."),
				mcp.WithNumber("bpm",
					mcp.Required(),
					mcp.Description("BPM value"),
				),
			),
			Handler:             handleBPMToTD,
			Category:            "avsync",
			Subcategory:         "bpm",
			Tags:                []string{"av", "sync", "bpm", "touchdesigner", "tempo"},
			UseCases:            []string{"Sync TouchDesigner to track BPM", "Update TD tempo variable"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// Key to Color
		{
			Tool: mcp.NewTool("aftrs_avsync_key_to_color",
				mcp.WithDescription("Map musical key (Camelot notation) to color palette and push to visual systems. Uses harmonic color theory for key-to-color mapping."),
				mcp.WithString("key",
					mcp.Required(),
					mcp.Description("Musical key in Camelot notation (1A-12B)"),
				),
			),
			Handler:             handleKeyToColor,
			Category:            "avsync",
			Subcategory:         "color",
			Tags:                []string{"av", "sync", "key", "color", "camelot", "palette"},
			UseCases:            []string{"Color visuals by key", "Harmonic color matching", "Key-based color schemes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// Genre Preset
		{
			Tool: mcp.NewTool("aftrs_avsync_genre_preset",
				mcp.WithDescription("Load a genre-appropriate visual preset. Maps genres (techno, house, trance, etc.) to visual style presets."),
				mcp.WithString("genre",
					mcp.Required(),
					mcp.Description("Music genre (e.g., techno, house, trance, drum-and-bass, ambient)"),
				),
			),
			Handler:             handleGenrePreset,
			Category:            "avsync",
			Subcategory:         "preset",
			Tags:                []string{"av", "sync", "genre", "preset", "style"},
			UseCases:            []string{"Auto-select visuals by genre", "Genre-matched visual styles"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// Energy Intensity
		{
			Tool: mcp.NewTool("aftrs_avsync_energy_intensity",
				mcp.WithDescription("Map track energy level (0.0-1.0) to visual effect intensities. Higher energy increases effect parameters, strobes, and brightness."),
				mcp.WithNumber("energy",
					mcp.Required(),
					mcp.Description("Energy level from 0.0 (calm) to 1.0 (intense)"),
				),
			),
			Handler:             handleEnergyIntensity,
			Category:            "avsync",
			Subcategory:         "energy",
			Tags:                []string{"av", "sync", "energy", "intensity", "effects"},
			UseCases:            []string{"Auto-intensity from Spotify energy", "Energy-reactive visuals"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// Track Cue
		{
			Tool: mcp.NewTool("aftrs_avsync_track_cue",
				mcp.WithDescription("Trigger complete visual sync on track change. Syncs BPM, key color, genre preset, energy, and now-playing info to all connected visual systems."),
				mcp.WithString("artist",
					mcp.Required(),
					mcp.Description("Track artist name"),
				),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Track title"),
				),
				mcp.WithNumber("bpm",
					mcp.Description("Track BPM"),
				),
				mcp.WithString("key",
					mcp.Description("Musical key in Camelot notation (e.g., 8A, 11B)"),
				),
				mcp.WithString("genre",
					mcp.Description("Music genre"),
				),
				mcp.WithNumber("energy",
					mcp.Description("Energy level 0.0-1.0"),
				),
			),
			Handler:             handleTrackCue,
			Category:            "avsync",
			Subcategory:         "cue",
			Tags:                []string{"av", "sync", "track", "cue", "trigger", "complete"},
			UseCases:            []string{"Full visual sync on track change", "Automated VJ cueing"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "avsync",
		},

		// Setlist Visuals
		{
			Tool: mcp.NewTool("aftrs_avsync_setlist_visuals",
				mcp.WithDescription("Pre-load visual configurations for an entire setlist. Returns color schemes and presets mapped to each track for preparation."),
				mcp.WithString("tracks",
					mcp.Required(),
					mcp.Description("JSON array of tracks with artist, title, bpm, key, genre fields"),
				),
			),
			Handler:             handleSetlistVisuals,
			Category:            "avsync",
			Subcategory:         "setlist",
			Tags:                []string{"av", "sync", "setlist", "preload", "prepare"},
			UseCases:            []string{"Prepare visuals for DJ set", "Pre-map setlist to visuals"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "avsync",
		},

		// Live Mode
		{
			Tool: mcp.NewTool("aftrs_avsync_live_mode",
				mcp.WithDescription("Enable or disable live sync mode for real-time track-to-visual synchronization."),
				mcp.WithBoolean("enable",
					mcp.Required(),
					mcp.Description("true to enable live mode, false to disable"),
				),
			),
			Handler:             handleLiveMode,
			Category:            "avsync",
			Subcategory:         "live",
			Tags:                []string{"av", "sync", "live", "realtime", "mode"},
			UseCases:            []string{"Enable real-time sync", "Toggle live VJ mode"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "avsync",
		},

		// Record Mapping
		{
			Tool: mcp.NewTool("aftrs_avsync_record_mapping",
				mcp.WithDescription("Record a custom audio-visual mapping. Creates a user-defined mapping between music metadata and visual parameters."),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Name for this mapping"),
				),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source type: bpm, key, genre, or energy"),
				),
				mcp.WithString("target",
					mcp.Required(),
					mcp.Description("Target system: resolume or touchdesigner"),
				),
				mcp.WithString("parameter",
					mcp.Required(),
					mcp.Description("Target parameter path (e.g., /composition/layers/1/video/opacity)"),
				),
				mcp.WithString("transform",
					mcp.Description("Transform type: linear (default), exponential, or step"),
				),
			),
			Handler:             handleRecordMapping,
			Category:            "avsync",
			Subcategory:         "mapping",
			Tags:                []string{"av", "sync", "mapping", "custom", "record"},
			UseCases:            []string{"Create custom AV mappings", "Learn new parameter mappings"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "avsync",
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Handler implementations

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleBPMToResolume(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	bpm := tools.GetFloatParam(req, "bpm", 0)
	if bpm == 0 {
		return tools.ErrorResult(fmt.Errorf("bpm parameter is required")), nil
	}

	if err := client.SyncBPMToResolume(ctx, bpm); err != nil {
		return tools.ErrorResult(fmt.Errorf("sync BPM to Resolume: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Set Resolume BPM to %.1f", bpm)), nil
}

func handleBPMToTD(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	bpm := tools.GetFloatParam(req, "bpm", 0)
	if bpm == 0 {
		return tools.ErrorResult(fmt.Errorf("bpm parameter is required")), nil
	}

	if err := client.SyncBPMToTouchDesigner(ctx, bpm); err != nil {
		return tools.ErrorResult(fmt.Errorf("sync BPM to TouchDesigner: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Set TouchDesigner BPM to %.1f", bpm)), nil
}

func handleKeyToColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	// Normalize key format (e.g., "8a" -> "8A")
	key = strings.ToUpper(key)

	result, err := client.SyncKeyToColor(ctx, key)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync key to color: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Key %s mapped to color %s\n\n", key, result.ColorSet))
	sb.WriteString("Parameters Updated:\n")
	for k, v := range result.ParamsUpdated {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}

	return tools.TextResult(sb.String()), nil
}

func handleGenrePreset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	genre, errResult := tools.RequireStringParam(req, "genre")
	if errResult != nil {
		return errResult, nil
	}

	// Normalize genre (lowercase, hyphens for spaces)
	genre = strings.ToLower(strings.ReplaceAll(genre, " ", "-"))

	result, err := client.SyncGenrePreset(ctx, genre)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync genre preset: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Genre '%s' mapped to preset '%s'\n\n", genre, result.PresetLoaded))
	sb.WriteString("Parameters Updated:\n")
	for k, v := range result.ParamsUpdated {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
	}

	return tools.TextResult(sb.String()), nil
}

func handleEnergyIntensity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	energy := tools.GetFloatParam(req, "energy", -1)
	if energy < 0 {
		return tools.ErrorResult(fmt.Errorf("energy parameter is required")), nil
	}

	result, err := client.SyncEnergyIntensity(ctx, energy)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync energy intensity: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Energy %.2f applied to visual intensities\n\n", energy))

	if len(result.ParamsUpdated) > 0 {
		sb.WriteString("Parameters Updated:\n")
		for k, v := range result.ParamsUpdated {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if len(result.Errors) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleTrackCue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	artist, errResult := tools.RequireStringParam(req, "artist")
	if errResult != nil {
		return errResult, nil
	}
	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	track := &clients.TrackMetadata{
		Artist: artist,
		Title:  title,
		BPM:    tools.GetFloatParam(req, "bpm", 0),
		Key:    strings.ToUpper(tools.GetStringParam(req, "key")),
		Genre:  strings.ToLower(strings.ReplaceAll(tools.GetStringParam(req, "genre"), " ", "-")),
		Energy: tools.GetFloatParam(req, "energy", 0),
	}

	result, err := client.SyncTrackCue(ctx, track)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync track cue: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Track Cue: %s - %s\n\n", artist, title))
	sb.WriteString("Sync Results:\n")
	sb.WriteString(fmt.Sprintf("  Resolume: %v\n", result.ResolumeSync))
	sb.WriteString(fmt.Sprintf("  TouchDesigner: %v\n", result.TDSync))

	if result.BPMSet > 0 {
		sb.WriteString(fmt.Sprintf("  BPM: %.1f\n", result.BPMSet))
	}
	if result.ColorSet != "" {
		sb.WriteString(fmt.Sprintf("  Color: %s\n", result.ColorSet))
	}
	if result.PresetLoaded != "" {
		sb.WriteString(fmt.Sprintf("  Preset: %s\n", result.PresetLoaded))
	}

	if len(result.ParamsUpdated) > 0 {
		sb.WriteString("\nAll Parameters:\n")
		for k, v := range result.ParamsUpdated {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if len(result.Errors) > 0 {
		sb.WriteString("\nErrors:\n")
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSetlistVisuals(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	tracksJSON, errResult := tools.RequireStringParam(req, "tracks")
	if errResult != nil {
		return errResult, nil
	}

	var tracks []clients.TrackMetadata
	if err := json.Unmarshal([]byte(tracksJSON), &tracks); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid tracks JSON: %w", err)), nil
	}

	visuals, err := client.LoadSetlistVisuals(ctx, tracks)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("load setlist visuals: %w", err)), nil
	}

	return tools.JSONResult(visuals), nil
}

func handleLiveMode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	enable := tools.GetBoolParam(req, "enable", false)

	if enable {
		if err := client.StartLiveMode(ctx); err != nil {
			return tools.ErrorResult(fmt.Errorf("start live mode: %w", err)), nil
		}
		return tools.TextResult("Live mode enabled. Real-time track-to-visual sync is now active."), nil
	} else {
		if err := client.StopLiveMode(); err != nil {
			return tools.ErrorResult(fmt.Errorf("stop live mode: %w", err)), nil
		}
		return tools.TextResult("Live mode disabled."), nil
	}
}

func handleRecordMapping(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetAVBridgeClient()

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}
	target, errResult := tools.RequireStringParam(req, "target")
	if errResult != nil {
		return errResult, nil
	}
	parameter, errResult := tools.RequireStringParam(req, "parameter")
	if errResult != nil {
		return errResult, nil
	}

	// Validate source
	validSources := []string{"bpm", "key", "genre", "energy"}
	sourceValid := false
	for _, s := range validSources {
		if source == s {
			sourceValid = true
			break
		}
	}
	if !sourceValid {
		return tools.ErrorResult(fmt.Errorf("source must be one of: bpm, key, genre, energy")), nil
	}

	// Validate target
	if target != "resolume" && target != "touchdesigner" {
		return tools.ErrorResult(fmt.Errorf("target must be 'resolume' or 'touchdesigner'")), nil
	}

	transform := tools.OptionalStringParam(req, "transform", "linear")

	mapping := clients.CustomMapping{
		Name:      name,
		Source:    source,
		Target:    target,
		Parameter: parameter,
		Transform: transform,
		Enabled:   true,
	}

	if err := client.AddCustomMapping(mapping); err != nil {
		return tools.ErrorResult(fmt.Errorf("add mapping: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Custom Mapping Created: %s\n\n", name))
	sb.WriteString(fmt.Sprintf("  Source: %s\n", source))
	sb.WriteString(fmt.Sprintf("  Target: %s\n", target))
	sb.WriteString(fmt.Sprintf("  Parameter: %s\n", parameter))
	sb.WriteString(fmt.Sprintf("  Transform: %s\n", transform))
	sb.WriteString(fmt.Sprintf("  Enabled: true\n"))

	return tools.TextResult(sb.String()), nil
}
