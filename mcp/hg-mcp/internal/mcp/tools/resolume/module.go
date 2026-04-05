// Package resolume provides Resolume Arena/Avenue control tools for hg-mcp.
package resolume

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewResolumeClient)

// Module implements the ToolModule interface for Resolume integration
type Module struct{}

func (m *Module) Name() string {
	return "resolume"
}

func (m *Module) Description() string {
	return "Resolume Arena/Avenue VJ software control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	baseTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_resolume_status",
				mcp.WithDescription("Get Resolume Arena/Avenue status including connection state and BPM."),
			),
			Handler:     handleStatus,
			Category:    "resolume",
			Subcategory: "status",
			Tags:        []string{"resolume", "status", "vj", "connection"},
			UseCases:    []string{"Check Resolume connection", "View current BPM"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_layers",
				mcp.WithDescription("List all layers with opacity and active clip information."),
			),
			Handler:     handleLayers,
			Category:    "resolume",
			Subcategory: "layers",
			Tags:        []string{"resolume", "layers", "opacity", "clips"},
			UseCases:    []string{"View layer status", "Check active clips"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clips",
				mcp.WithDescription("List clips in a specific layer's clip bank."),
				mcp.WithNumber("layer",
					mcp.Required(),
					mcp.Description("Layer number (1-based)"),
				),
			),
			Handler:     handleClips,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clips", "bank", "media"},
			UseCases:    []string{"Browse clip bank", "Find specific clips"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_deck",
				mcp.WithDescription("Get information about decks in the composition."),
			),
			Handler:     handleDeck,
			Category:    "resolume",
			Subcategory: "deck",
			Tags:        []string{"resolume", "deck", "composition"},
			UseCases:    []string{"View deck configuration", "Check composition structure"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_bpm",
				mcp.WithDescription("Get or set the BPM (beats per minute)."),
				mcp.WithNumber("bpm",
					mcp.Description("BPM to set (20-999). Omit to just read current BPM."),
				),
				mcp.WithBoolean("tap",
					mcp.Description("If true, sends a tap tempo signal instead of setting BPM."),
				),
			),
			Handler:     handleBPM,
			Category:    "resolume",
			Subcategory: "tempo",
			Tags:        []string{"resolume", "bpm", "tempo", "sync"},
			UseCases:    []string{"Adjust tempo", "Sync to music"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_trigger",
				mcp.WithDescription("Trigger a clip or column in Resolume."),
				mcp.WithNumber("layer",
					mcp.Description("Layer number (1-based). Required for clip trigger."),
				),
				mcp.WithNumber("column",
					mcp.Required(),
					mcp.Description("Column number (1-based)."),
				),
			),
			Handler:     handleTrigger,
			Category:    "resolume",
			Subcategory: "control",
			Tags:        []string{"resolume", "trigger", "clip", "column"},
			UseCases:    []string{"Trigger clips", "Fire columns"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_effects",
				mcp.WithDescription("List effects on a layer or the master."),
				mcp.WithNumber("layer",
					mcp.Description("Layer number (1-based). Omit for master effects."),
				),
			),
			Handler:     handleEffects,
			Category:    "resolume",
			Subcategory: "effects",
			Tags:        []string{"resolume", "effects", "fx", "processing"},
			UseCases:    []string{"View effect chains", "Check effect status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_output",
				mcp.WithDescription("Get output routing and display configuration."),
			),
			Handler:     handleOutput,
			Category:    "resolume",
			Subcategory: "output",
			Tags:        []string{"resolume", "output", "display", "routing"},
			UseCases:    []string{"Check output config", "View display setup"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_record",
				mcp.WithDescription("Control recording in Resolume."),
				mcp.WithString("action",
					mcp.Required(),
					mcp.Description("Action: 'start', 'stop', or 'status'"),
				),
			),
			Handler:     handleRecord,
			Category:    "resolume",
			Subcategory: "recording",
			Tags:        []string{"resolume", "record", "capture", "export"},
			UseCases:    []string{"Record performance", "Capture output"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_health",
				mcp.WithDescription("Get overall Resolume system health and recommendations."),
			),
			Handler:     handleHealth,
			Category:    "resolume",
			Subcategory: "health",
			Tags:        []string{"resolume", "health", "monitoring", "status"},
			UseCases:    []string{"Check system health", "Troubleshoot issues"},
			Complexity:  tools.ComplexityModerate,
		},
		// Additional tools for expanded coverage
		{
			Tool: mcp.NewTool("aftrs_resolume_layer_opacity",
				mcp.WithDescription("Set opacity for a specific layer."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("opacity", mcp.Required(), mcp.Description("Opacity value (0-100)")),
			),
			Handler:     handleLayerOpacity,
			Category:    "resolume",
			Subcategory: "layers",
			Tags:        []string{"resolume", "layer", "opacity", "control"},
			UseCases:    []string{"Adjust layer visibility", "Fade layers"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_bypass",
				mcp.WithDescription("Bypass or enable a layer."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithBoolean("bypass", mcp.Description("True to bypass, false to enable (default: toggle)")),
			),
			Handler:     handleBypass,
			Category:    "resolume",
			Subcategory: "layers",
			Tags:        []string{"resolume", "layer", "bypass", "mute"},
			UseCases:    []string{"Mute layer output", "Toggle layer visibility"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_solo",
				mcp.WithDescription("Solo a specific layer."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithBoolean("solo", mcp.Description("True to solo, false to unsolo (default: toggle)")),
			),
			Handler:     handleSolo,
			Category:    "resolume",
			Subcategory: "layers",
			Tags:        []string{"resolume", "layer", "solo", "isolate"},
			UseCases:    []string{"Isolate layer for preview", "Focus on single layer"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clear",
				mcp.WithDescription("Clear all layers or a specific layer."),
				mcp.WithNumber("layer", mcp.Description("Layer to clear (omit to clear all)")),
			),
			Handler:     handleClear,
			Category:    "resolume",
			Subcategory: "control",
			Tags:        []string{"resolume", "clear", "disconnect", "reset"},
			UseCases:    []string{"Clear all clips", "Reset composition"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_columns",
				mcp.WithDescription("List columns in the composition."),
			),
			Handler:     handleColumns,
			Category:    "resolume",
			Subcategory: "composition",
			Tags:        []string{"resolume", "columns", "scenes", "composition"},
			UseCases:    []string{"View column layout", "List scenes"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_crossfade",
				mcp.WithDescription("Crossfade between decks (A/B mixing)."),
				mcp.WithNumber("position", mcp.Required(), mcp.Description("Crossfade position (0=Deck A, 100=Deck B)")),
			),
			Handler:     handleCrossfade,
			Category:    "resolume",
			Subcategory: "deck",
			Tags:        []string{"resolume", "crossfade", "deck", "mix"},
			UseCases:    []string{"Mix between decks", "Transition compositions"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_effect_toggle",
				mcp.WithDescription("Toggle an effect on or off."),
				mcp.WithNumber("layer", mcp.Description("Layer number (omit for master)")),
				mcp.WithNumber("effect", mcp.Required(), mcp.Description("Effect index (1-based)")),
				mcp.WithBoolean("enabled", mcp.Description("True to enable, false to disable")),
			),
			Handler:     handleEffectToggle,
			Category:    "resolume",
			Subcategory: "effects",
			Tags:        []string{"resolume", "effect", "toggle", "bypass"},
			UseCases:    []string{"Enable/disable effects", "Control effect chain"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_master",
				mcp.WithDescription("Get or set the master output level."),
				mcp.WithNumber("level", mcp.Description("Master level (0-100). Omit to read current.")),
			),
			Handler:     handleMaster,
			Category:    "resolume",
			Subcategory: "output",
			Tags:        []string{"resolume", "master", "level", "output"},
			UseCases:    []string{"Adjust master output", "Fade to black"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_autopilot",
				mcp.WithDescription("Control autopilot/random mode."),
				mcp.WithBoolean("enabled", mcp.Description("Enable or disable autopilot")),
				mcp.WithString("mode", mcp.Description("Mode: random, sequential, bpm")),
				mcp.WithNumber("interval", mcp.Description("Interval in seconds between changes")),
			),
			Handler:     handleAutopilot,
			Category:    "resolume",
			Subcategory: "automation",
			Tags:        []string{"resolume", "autopilot", "random", "automation"},
			UseCases:    []string{"Auto-VJ mode", "Random clip triggering"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_speed",
				mcp.WithDescription("Set playback speed for a clip."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
				mcp.WithNumber("speed", mcp.Required(), mcp.Description("Speed multiplier (e.g., 0.5=half, 2.0=double)")),
			),
			Handler:     handleClipSpeed,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "speed", "playback"},
			UseCases:    []string{"Slow motion", "Speed up playback"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		// VJ Clip Management
		{
			Tool: mcp.NewTool("aftrs_resolume_local_clips",
				mcp.WithDescription("List locally synced VJ clips available for loading into Resolume."),
				mcp.WithString("pack", mcp.Description("Filter by pack name (e.g., 'hackerglasses', 'hairglasses')")),
			),
			Handler:     handleLocalClips,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clips", "local", "library", "vj"},
			UseCases:    []string{"Browse local VJ clips", "Find clips to load"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_show_info",
				mcp.WithDescription("Get comprehensive show status: connection, layers, clips, sync status, and recommendations."),
			),
			Handler:     handleShowInfo,
			Category:    "resolume",
			Subcategory: "show",
			Tags:        []string{"resolume", "show", "status", "overview", "vj"},
			UseCases:    []string{"Pre-show check", "VJ status overview"},
			Complexity:  tools.ComplexityModerate,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_quick_setup",
				mcp.WithDescription("Quick VJ setup: clear layers, set BPM, and prepare for show."),
				mcp.WithNumber("bpm", mcp.Description("BPM to set (default: 128)")),
				mcp.WithBoolean("clear", mcp.Description("Clear all layers first (default: true)")),
			),
			Handler:     handleQuickSetup,
			Category:    "resolume",
			Subcategory: "show",
			Tags:        []string{"resolume", "setup", "show", "quick", "vj"},
			UseCases:    []string{"Quick show prep", "Reset for new set"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_random_trigger",
				mcp.WithDescription("Trigger a random clip from a layer or all layers."),
				mcp.WithNumber("layer", mcp.Description("Layer to trigger random clip on (omit for random layer)")),
			),
			Handler:     handleRandomTrigger,
			Category:    "resolume",
			Subcategory: "control",
			Tags:        []string{"resolume", "random", "trigger", "vj", "auto"},
			UseCases:    []string{"Random VJ mode", "Auto-trigger clips"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},

		// ============================================================================
		// Phase 1: Effect Parameter Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_resolume_effect_params",
				mcp.WithDescription("Get all parameters for an effect with their current values and ranges."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("effect", mcp.Required(), mcp.Description("Effect index (1-based)")),
			),
			Handler:     handleEffectParams,
			Category:    "resolume",
			Subcategory: "effects",
			Tags:        []string{"resolume", "effect", "params", "control"},
			UseCases:    []string{"Explore effect parameters", "Find parameter IDs"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_effect_set",
				mcp.WithDescription("Set an effect parameter by ID or name."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("effect", mcp.Required(), mcp.Description("Effect index (1-based)")),
				mcp.WithNumber("param_id", mcp.Description("Parameter ID (use aftrs_resolume_effect_params to find)")),
				mcp.WithString("param_name", mcp.Description("Parameter name (alternative to ID)")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Value to set (0-100 for percentages)")),
			),
			Handler:     handleEffectSet,
			Category:    "resolume",
			Subcategory: "effects",
			Tags:        []string{"resolume", "effect", "param", "set", "control"},
			UseCases:    []string{"Adjust effect parameters", "Fine-tune effects"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_effect_mix",
				mcp.WithDescription("Set the mix/intensity of an effect."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("effect", mcp.Required(), mcp.Description("Effect index (1-based)")),
				mcp.WithNumber("mix", mcp.Required(), mcp.Description("Mix value (0-100)")),
			),
			Handler:     handleEffectMix,
			Category:    "resolume",
			Subcategory: "effects",
			Tags:        []string{"resolume", "effect", "mix", "intensity"},
			UseCases:    []string{"Adjust effect intensity", "Fade effects"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},

		// ============================================================================
		// Phase 1: Clip Management Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_info",
				mcp.WithDescription("Get detailed information about a clip including resolution, duration, and properties."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
			),
			Handler:     handleClipInfo,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "info", "details"},
			UseCases:    []string{"View clip details", "Check clip properties"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_load",
				mcp.WithDescription("Load a video file into a clip slot."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
				mcp.WithString("file", mcp.Required(), mcp.Description("Path to video file")),
			),
			Handler:     handleClipLoad,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "load", "import"},
			UseCases:    []string{"Load new clips", "Add media"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_properties",
				mcp.WithDescription("Set clip playback properties: trigger style, beat snap, direction."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
				mcp.WithString("trigger_style", mcp.Description("Trigger style: toggle, gate, retrigger")),
				mcp.WithString("beat_snap", mcp.Description("Beat snap: off, beat, bar, 4bars")),
				mcp.WithString("direction", mcp.Description("Playback direction: forward, backward, pingpong")),
			),
			Handler:     handleClipProperties,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "properties", "playback"},
			UseCases:    []string{"Configure clip playback", "Set trigger mode"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_transport",
				mcp.WithDescription("Control clip transport: seek to position."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
				mcp.WithNumber("position", mcp.Description("Seek position (0-100 as percentage)")),
			),
			Handler:     handleClipTransport,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "transport", "seek"},
			UseCases:    []string{"Seek clip position", "Jump to time"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_clip_thumbnail",
				mcp.WithDescription("Get clip thumbnail as base64 encoded PNG."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("column", mcp.Required(), mcp.Description("Column number (1-based)")),
			),
			Handler:     handleClipThumbnail,
			Category:    "resolume",
			Subcategory: "clips",
			Tags:        []string{"resolume", "clip", "thumbnail", "preview"},
			UseCases:    []string{"Preview clip", "Get clip image"},
			Complexity:  tools.ComplexitySimple,
		},

		// ============================================================================
		// Phase 1: Layer Group Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_resolume_groups",
				mcp.WithDescription("List all layer groups in the composition."),
			),
			Handler:     handleGroups,
			Category:    "resolume",
			Subcategory: "groups",
			Tags:        []string{"resolume", "groups", "layers", "organization"},
			UseCases:    []string{"View layer groups", "Check group status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_group_control",
				mcp.WithDescription("Control a layer group: opacity, bypass, solo."),
				mcp.WithNumber("group", mcp.Required(), mcp.Description("Group number (1-based)")),
				mcp.WithNumber("opacity", mcp.Description("Opacity (0-100)")),
				mcp.WithBoolean("bypass", mcp.Description("Bypass the group")),
				mcp.WithBoolean("solo", mcp.Description("Solo the group")),
			),
			Handler:     handleGroupControl,
			Category:    "resolume",
			Subcategory: "groups",
			Tags:        []string{"resolume", "group", "control", "opacity"},
			UseCases:    []string{"Control layer groups", "Group fading"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},

		// ============================================================================
		// Phase 1: Audio Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_resolume_audio_tracks",
				mcp.WithDescription("List audio tracks/layers with volume and mute status."),
			),
			Handler:     handleAudioTracks,
			Category:    "resolume",
			Subcategory: "audio",
			Tags:        []string{"resolume", "audio", "tracks", "volume"},
			UseCases:    []string{"View audio status", "Check volumes"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_audio_volume",
				mcp.WithDescription("Set audio volume for a layer or master."),
				mcp.WithNumber("layer", mcp.Description("Layer number (omit for master)")),
				mcp.WithNumber("volume", mcp.Required(), mcp.Description("Volume (0-100)")),
			),
			Handler:     handleAudioVolume,
			Category:    "resolume",
			Subcategory: "audio",
			Tags:        []string{"resolume", "audio", "volume", "level"},
			UseCases:    []string{"Adjust audio levels", "Control volume"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_audio_mute",
				mcp.WithDescription("Mute or unmute audio for a layer or master."),
				mcp.WithNumber("layer", mcp.Description("Layer number (omit for master)")),
				mcp.WithBoolean("mute", mcp.Required(), mcp.Description("True to mute, false to unmute")),
			),
			Handler:     handleAudioMute,
			Category:    "resolume",
			Subcategory: "audio",
			Tags:        []string{"resolume", "audio", "mute", "silence"},
			UseCases:    []string{"Mute audio", "Toggle audio"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_audio_pan",
				mcp.WithDescription("Set audio pan for a layer."),
				mcp.WithNumber("layer", mcp.Required(), mcp.Description("Layer number (1-based)")),
				mcp.WithNumber("pan", mcp.Required(), mcp.Description("Pan position (-100=left, 0=center, 100=right)")),
			),
			Handler:     handleAudioPan,
			Category:    "resolume",
			Subcategory: "audio",
			Tags:        []string{"resolume", "audio", "pan", "stereo"},
			UseCases:    []string{"Pan audio", "Stereo positioning"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
	}

	// Batch operation tools
	batchTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_resolume_batch_trigger",
				mcp.WithDescription("Trigger multiple clips or columns in a single call. Reduces latency for multi-layer updates."),
				mcp.WithString("triggers",
					mcp.Required(),
					mcp.Description("JSON array of triggers: [{\"layer\":1,\"column\":2},{\"layer\":3,\"column\":2},{\"column\":5}]. Omit layer to trigger entire column."),
				),
			),
			Handler:     handleBatchTrigger,
			Category:    "vj",
			Subcategory: "resolume",
			Tags:        []string{"resolume", "trigger", "batch", "clips", "columns"},
			UseCases:    []string{"Trigger multiple clips at once", "Batch scene changes"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
	}
	baseTools = append(baseTools, batchTools...)

	// Add display tools for track info sync
	baseTools = append(baseTools, DisplayTools()...)

	// Apply circuit breaker to all tools — network-dependent (OSC)
	for i := range baseTools {
		baseTools[i].CircuitBreakerGroup = "resolume"
	}

	return baseTools
}

// getClient creates a new Resolume client

// handleStatus handles the aftrs_resolume_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** ❌ Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**OSC Target:** %s:%d\n\n", client.OSCHost(), client.OSCPort()))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open Resolume Arena or Avenue\n")
		sb.WriteString("2. Go to Preferences → OSC\n")
		sb.WriteString("3. Enable OSC input on port 7000\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export RESOLUME_OSC_HOST=127.0.0.1\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** ✅ Connected\n\n")
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", status.Version))
	if status.Composition != "" {
		sb.WriteString(fmt.Sprintf("**Composition:** %s\n", status.Composition))
	}
	sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", status.BPM))
	sb.WriteString(fmt.Sprintf("**Master Level:** %.0f%%\n", status.MasterLevel*100))

	playStatus := "⏸️ Stopped"
	if status.Playing {
		playStatus = "▶️ Playing"
	}
	sb.WriteString(fmt.Sprintf("**Playback:** %s\n", playStatus))

	return tools.TextResult(sb.String()), nil
}

// handleLayers handles the aftrs_resolume_layers tool
func handleLayers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	layers, err := client.GetLayers(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Layers\n\n")

	if len(layers) == 0 {
		sb.WriteString("No layers found.\n\n")
		sb.WriteString("*Note: Requires Resolume OSC/API connection to list layers.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** layers:\n\n", len(layers)))
	sb.WriteString("| # | Name | Opacity | Active Clip | Status |\n")
	sb.WriteString("|---|------|---------|-------------|--------|\n")

	for _, layer := range layers {
		status := "✅"
		if layer.Bypassed {
			status = "⏸️ Bypassed"
		} else if layer.Solo {
			status = "🎯 Solo"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %d | %s |\n",
			layer.Index, layer.Name, layer.Opacity*100, layer.ActiveClip, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleClips handles the aftrs_resolume_clips tool
func handleClips(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layerNum, errResult := tools.RequireIntParam(req, "layer")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	clips, err := client.GetClips(ctx, layerNum)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Clips in Layer %d\n\n", int(layerNum)))

	if len(clips) == 0 {
		sb.WriteString("No clips found.\n\n")
		sb.WriteString("*Note: Requires Resolume API connection to list clips.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** clips:\n\n", len(clips)))
	sb.WriteString("| Column | Name | Duration | Status |\n")
	sb.WriteString("|--------|------|----------|--------|\n")

	for _, clip := range clips {
		status := "⚫"
		if clip.Playing {
			status = "▶️ Playing"
		} else if clip.Connected {
			status = "🔗 Connected"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.1fs | %s |\n",
			clip.Column, clip.Name, clip.Duration, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDeck handles the aftrs_resolume_deck tool
func handleDeck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	decks, err := client.GetDecks(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Decks\n\n")

	if len(decks) == 0 {
		sb.WriteString("No decks found.\n\n")
		sb.WriteString("*Note: Requires Resolume API connection to list decks.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** decks:\n\n", len(decks)))
	sb.WriteString("| # | Name | Layers | Clips | Status |\n")
	sb.WriteString("|---|------|--------|-------|--------|\n")

	for _, deck := range decks {
		status := "⚫ Inactive"
		if deck.Active {
			status = "✅ Active"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %d | %d | %s |\n",
			deck.Index, deck.Name, deck.LayerCount, deck.ClipCount, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBPM handles the aftrs_resolume_bpm tool
func handleBPM(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	bpm := float64(tools.GetIntParam(req, "bpm", 0))
	tap := tools.GetBoolParam(req, "tap", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Handle tap tempo
	if tap {
		err := client.TapTempo(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("✅ Tap tempo sent"), nil
	}

	// If BPM provided, set it
	if bpm > 0 {
		err := client.SetBPM(ctx, bpm)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ BPM set to %.1f", bpm)), nil
	}

	// Otherwise, read current BPM
	currentBPM, err := client.GetBPM(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Current BPM:** %.1f", currentBPM)), nil
}

// handleTrigger handles the aftrs_resolume_trigger tool
func handleTrigger(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column, errResult := tools.RequireIntParam(req, "column")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If layer specified, trigger specific clip
	if layer > 0 {
		err := client.TriggerClip(ctx, layer, column)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Triggered clip at layer %d, column %d", layer, column)), nil
	}

	// Otherwise, trigger entire column
	err = client.TriggerColumn(ctx, column)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("✅ Triggered column %d", column)), nil
}

// handleEffects handles the aftrs_resolume_effects tool
func handleEffects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	effects, err := client.GetEffects(ctx, layer)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	if layer > 0 {
		sb.WriteString(fmt.Sprintf("# Effects on Layer %d\n\n", layer))
	} else {
		sb.WriteString("# Master Effects\n\n")
	}

	if len(effects) == 0 {
		sb.WriteString("No effects found.\n\n")
		sb.WriteString("*Note: Requires Resolume API connection to list effects.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** effects:\n\n", len(effects)))
	sb.WriteString("| Name | Type | Mix | Status |\n")
	sb.WriteString("|------|------|-----|--------|\n")

	for _, fx := range effects {
		status := "⚫ Disabled"
		if fx.Enabled {
			status = "✅ Enabled"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %.0f%% | %s |\n",
			fx.Name, fx.Type, fx.Mix*100, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleOutput handles the aftrs_resolume_output tool
func handleOutput(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	outputs, err := client.GetOutputs(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Outputs\n\n")

	if len(outputs) == 0 {
		sb.WriteString("No outputs configured.\n\n")
		sb.WriteString("*Note: Requires Resolume API connection to list outputs.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** outputs:\n\n", len(outputs)))
	sb.WriteString("| # | Name | Resolution | Fullscreen | Status |\n")
	sb.WriteString("|---|------|------------|------------|--------|\n")

	for _, out := range outputs {
		status := "⚫ Disabled"
		if out.Enabled {
			status = "✅ Enabled"
		}
		fs := "No"
		if out.Fullscreen {
			fs = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %dx%d | %s | %s |\n",
			out.Index, out.Name, out.Width, out.Height, fs, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleRecord handles the aftrs_resolume_record tool
func handleRecord(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "status")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err := client.StartRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("✅ Recording started"), nil

	case "stop":
		err := client.StopRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("✅ Recording stopped"), nil

	case "status":
		return tools.TextResult("**Recording Status:** Check Resolume directly for recording state."), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use start, stop, or status)", action)), nil
	}
}

// handleHealth handles the aftrs_resolume_health tool
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Health\n\n")

	// Status emoji
	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Layers | %d |\n", health.LayerCount))
	sb.WriteString(fmt.Sprintf("| Clips | %d |\n", health.ClipCount))
	sb.WriteString(fmt.Sprintf("| Outputs | %d |\n", health.OutputCount))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleLayerOpacity handles the aftrs_resolume_layer_opacity tool
func handleLayerOpacity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	opacity := tools.GetIntParam(req, "opacity", -1)

	if layer == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer is required")), nil
	}
	if opacity < 0 || opacity > 100 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("opacity must be between 0 and 100")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetLayerOpacity(ctx, layer, float64(opacity)/100.0)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Set layer %d opacity to %d%%", layer, opacity)), nil
}

// handleBypass handles the aftrs_resolume_bypass tool
func handleBypass(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	bypass := tools.GetBoolParam(req, "bypass", true)

	if layer == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetLayerBypass(ctx, layer, bypass)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "bypassed"
	if !bypass {
		status = "enabled"
	}
	return tools.TextResult(fmt.Sprintf("✅ Layer %d %s", layer, status)), nil
}

// handleSolo handles the aftrs_resolume_solo tool
func handleSolo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	solo := tools.GetBoolParam(req, "solo", true)

	if layer == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetLayerSolo(ctx, layer, solo)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "soloed"
	if !solo {
		status = "unsoloed"
	}
	return tools.TextResult(fmt.Sprintf("✅ Layer %d %s", layer, status)), nil
}

// handleClear handles the aftrs_resolume_clear tool
func handleClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if layer > 0 {
		err = client.ClearLayer(ctx, layer)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Cleared layer %d", layer)), nil
	}

	err = client.ClearAll(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult("✅ Cleared all layers"), nil
}

// handleColumns handles the aftrs_resolume_columns tool
func handleColumns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	columns, err := client.GetColumns(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Columns\n\n")

	if len(columns) == 0 {
		sb.WriteString("No columns found.\n\n")
		sb.WriteString("*Note: Requires Resolume API connection to list columns.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** columns:\n\n", len(columns)))
	sb.WriteString("| # | Name | Clips | Status |\n")
	sb.WriteString("|---|------|-------|--------|\n")

	for _, col := range columns {
		status := "⚫"
		if col.Connected {
			status = "✅ Active"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s |\n", col.Index, col.Name, col.ClipsLoaded, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleCrossfade handles the aftrs_resolume_crossfade tool
func handleCrossfade(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	position := tools.GetIntParam(req, "position", -1)

	if position < 0 || position > 100 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("position must be between 0 and 100")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.CrossfadeDecks(ctx, float64(position)/100.0)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	deckLabel := "A"
	if position >= 50 {
		deckLabel = "B"
	}
	return tools.TextResult(fmt.Sprintf("✅ Crossfade set to %d%% (toward Deck %s)", position, deckLabel)), nil
}

// handleEffectToggle handles the aftrs_resolume_effect_toggle tool
func handleEffectToggle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	effect := tools.GetIntParam(req, "effect", 0)
	enabled := tools.GetBoolParam(req, "enabled", true)

	if effect == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("effect index is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetEffectEnabled(ctx, layer, effect, enabled)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	target := "master"
	if layer > 0 {
		target = fmt.Sprintf("layer %d", layer)
	}
	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	return tools.TextResult(fmt.Sprintf("✅ Effect %d on %s %s", effect, target, status)), nil
}

// handleMaster handles the aftrs_resolume_master tool
func handleMaster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	level := tools.GetIntParam(req, "level", -1)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If level provided, set it
	if level >= 0 {
		if level > 100 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("level must be between 0 and 100")), nil
		}
		err = client.SetMasterLevel(ctx, float64(level)/100.0)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Master level set to %d%%", level)), nil
	}

	// Otherwise read current level
	currentLevel, err := client.GetMasterLevel(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Master Level:** %.0f%%", currentLevel*100)), nil
}

// handleAutopilot handles the aftrs_resolume_autopilot tool
func handleAutopilot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	enabled := tools.GetBoolParam(req, "enabled", false)
	mode := tools.GetStringParam(req, "mode")
	interval := float64(tools.GetIntParam(req, "interval", 8))

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If any params provided, configure autopilot
	if mode != "" || enabled {
		if mode == "" {
			mode = "random"
		}
		err = client.SetAutopilot(ctx, enabled, mode, interval)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		status := "disabled"
		if enabled {
			status = fmt.Sprintf("enabled (%s, %.0fs)", mode, interval)
		}
		return tools.TextResult(fmt.Sprintf("✅ Autopilot %s", status)), nil
	}

	// Otherwise get current status
	autopilot, err := client.GetAutopilot(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Autopilot Status\n\n")

	status := "❌ Disabled"
	if autopilot.Enabled {
		status = "✅ Enabled"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", status))
	sb.WriteString(fmt.Sprintf("**Mode:** %s\n", autopilot.Mode))
	sb.WriteString(fmt.Sprintf("**Interval:** %.1f seconds\n", autopilot.Interval))

	return tools.TextResult(sb.String()), nil
}

// handleClipSpeed handles the aftrs_resolume_clip_speed tool
func handleClipSpeed(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)
	speed := float64(tools.GetIntParam(req, "speed", 100)) / 100.0

	if layer == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer is required")), nil
	}
	if column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("column is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetClipSpeed(ctx, layer, column, speed)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Set clip [%d,%d] speed to %.2fx", layer, column, speed)), nil
}

// VJ Pack definitions matching vj-sync.sh
var vjPacks = map[string]string{
	"hackerglasses": "HACKERGLASSES",
	"hairglasses":   "Hairglasses Visuals",
	"fetz":          "Fetz VJ Storage",
	"masks":         "Masks",
	"algorave":      "Algorave",
	"relic":         "Relic VJ Clips",
	"mantissa":      "Mantissa",
	"tricky":        "TrickyFM Visuals",
	"church":        "Hairglasses at Church",
	"footage":       "Hairglasses Footage",
}

// handleLocalClips lists locally synced VJ clips
func handleLocalClips(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	packFilter := tools.GetStringParam(req, "pack")

	home, _ := os.UserHomeDir()
	mediaPath := filepath.Join(home, "Documents", "Resolume Arena", "Media")

	var sb strings.Builder
	sb.WriteString("# Local VJ Clips\n\n")
	sb.WriteString(fmt.Sprintf("**Media Folder:** `%s`\n\n", mediaPath))

	totalClips := 0
	totalSize := int64(0)

	for packKey, packFolder := range vjPacks {
		// Apply filter if specified
		if packFilter != "" && packKey != packFilter {
			continue
		}

		packPath := filepath.Join(mediaPath, packFolder)
		info, err := os.Stat(packPath)
		if err != nil || !info.IsDir() {
			continue
		}

		// Count clips and size
		clips := []string{}
		packSize := int64(0)
		filepath.Walk(packPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mov" || ext == ".mp4" || ext == ".avi" {
				clips = append(clips, info.Name())
				packSize += info.Size()
			}
			return nil
		})

		if len(clips) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s (%s)\n", packFolder, packKey))
		sb.WriteString(fmt.Sprintf("**Clips:** %d | **Size:** %.1f GB\n\n", len(clips), float64(packSize)/(1024*1024*1024)))

		// Show first 10 clips
		for i, clip := range clips {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(clips)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", clip))
		}
		sb.WriteString("\n")

		totalClips += len(clips)
		totalSize += packSize
	}

	if totalClips == 0 {
		sb.WriteString("*No clips found. Run `vj-sync.sh auto` to sync clips.*\n")
	} else {
		sb.WriteString("---\n")
		sb.WriteString(fmt.Sprintf("**Total:** %d clips (%.1f GB)\n", totalClips, float64(totalSize)/(1024*1024*1024)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleShowInfo provides comprehensive show status
func handleShowInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# 🎬 VJ Show Status\n\n")

	// Connection status
	status, err := client.GetStatus(ctx)
	if err != nil {
		sb.WriteString("## Connection\n❌ **Not connected to Resolume**\n\n")
		sb.WriteString("Start Resolume Arena and enable the webserver.\n")
		return tools.TextResult(sb.String()), nil
	}

	if !status.Connected {
		sb.WriteString("## Connection\n❌ **Resolume not responding**\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Connection\n")
	sb.WriteString(fmt.Sprintf("✅ **Connected** - %s\n", status.Version))
	sb.WriteString(fmt.Sprintf("📝 **Composition:** %s\n", status.Composition))
	sb.WriteString(fmt.Sprintf("🎵 **BPM:** %.1f\n", status.BPM))
	sb.WriteString(fmt.Sprintf("🔊 **Master:** %.0f%%\n\n", status.MasterLevel*100))

	// Layers
	layers, _ := client.GetLayers(ctx)
	sb.WriteString("## Layers\n")
	if len(layers) == 0 {
		sb.WriteString("*No layers*\n\n")
	} else {
		sb.WriteString("| # | Name | Opacity | Status |\n")
		sb.WriteString("|---|------|---------|--------|\n")
		for _, layer := range layers {
			status := "✅"
			if layer.Bypassed {
				status = "⏸️ Bypassed"
			} else if layer.Solo {
				status = "🎯 Solo"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %s |\n",
				layer.Index, layer.Name, layer.Opacity*100, status))
		}
		sb.WriteString("\n")
	}

	// Local clips summary
	home, _ := os.UserHomeDir()
	mediaPath := filepath.Join(home, "Documents", "Resolume Arena", "Media")
	clipCount := 0
	for _, packFolder := range vjPacks {
		packPath := filepath.Join(mediaPath, packFolder)
		filepath.Walk(packPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".mov" || ext == ".mp4" {
				clipCount++
			}
			return nil
		})
	}

	sb.WriteString("## Local Clips\n")
	sb.WriteString(fmt.Sprintf("📁 **Available:** %d clips synced\n", clipCount))
	sb.WriteString(fmt.Sprintf("📂 **Path:** `%s`\n\n", mediaPath))

	// Recommendations
	sb.WriteString("## Quick Actions\n")
	sb.WriteString("- `aftrs_resolume_quick_setup bpm=128` - Reset for new set\n")
	sb.WriteString("- `aftrs_resolume_trigger layer=1 column=1` - Trigger clip\n")
	sb.WriteString("- `aftrs_resolume_random_trigger` - Random clip\n")

	return tools.TextResult(sb.String()), nil
}

// handleQuickSetup provides quick VJ setup
func handleQuickSetup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	bpm := float64(tools.GetIntParam(req, "bpm", 128))
	clear := tools.GetBoolParam(req, "clear", true)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# 🎬 Quick Setup\n\n")

	// Clear if requested
	if clear {
		err := client.ClearAll(ctx)
		if err != nil {
			sb.WriteString(fmt.Sprintf("⚠️ Clear failed: %v\n", err))
		} else {
			sb.WriteString("✅ Cleared all layers\n")
		}
	}

	// Set BPM
	err = client.SetBPM(ctx, bpm)
	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ BPM set failed: %v\n", err))
	} else {
		sb.WriteString(fmt.Sprintf("✅ BPM set to %.0f\n", bpm))
	}

	// Set master to 100%
	err = client.SetMasterLevel(ctx, 1.0)
	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Master level failed: %v\n", err))
	} else {
		sb.WriteString("✅ Master level set to 100%\n")
	}

	sb.WriteString("\n**Ready to VJ!** 🎉\n")

	return tools.TextResult(sb.String()), nil
}

// handleRandomTrigger triggers a random clip
func handleRandomTrigger(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layerNum := tools.GetIntParam(req, "layer", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get layers info
	layers, err := client.GetLayers(ctx)
	if err != nil || len(layers) == 0 {
		return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("no layers available")), nil
	}

	// Select layer
	targetLayer := layerNum
	if targetLayer == 0 {
		// Random layer
		targetLayer = 1 + (int(time.Now().UnixNano()) % len(layers))
	}

	if targetLayer < 1 || targetLayer > len(layers) {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid layer: %d (have %d layers)", targetLayer, len(layers))), nil
	}

	// Get clips for this layer
	clips, err := client.GetClips(ctx, targetLayer)
	if err != nil || len(clips) == 0 {
		return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("no clips in layer %d", targetLayer)), nil
	}

	// Pick random clip
	randomClip := 1 + (int(time.Now().UnixNano()) % len(clips))

	// Trigger it
	err = client.TriggerClip(ctx, targetLayer, randomClip)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	clipName := fmt.Sprintf("Clip %d", randomClip)
	if randomClip <= len(clips) {
		clipName = clips[randomClip-1].Name
	}

	return tools.TextResult(fmt.Sprintf("🎲 Triggered **%s** on layer %d, column %d", clipName, targetLayer, randomClip)), nil
}

// ============================================================================
// Phase 1: Effect Parameter Handlers
// ============================================================================

// handleEffectParams lists all parameters for an effect
func handleEffectParams(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	effect := tools.GetIntParam(req, "effect", 0)

	if layer == 0 || effect == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and effect are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	ext, err := client.GetEffectExtended(ctx, layer, effect)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Effect: %s\n\n", ext.Name))
	sb.WriteString(fmt.Sprintf("**Layer:** %d | **Index:** %d\n", layer, effect))
	sb.WriteString(fmt.Sprintf("**Enabled:** %v | **Mix:** %.0f%%\n\n", ext.Enabled, ext.Mix*100))

	if len(ext.Parameters) == 0 {
		sb.WriteString("*No parameters available*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Parameters\n\n")
	sb.WriteString("| ID | Name | Value | Range | Type |\n")
	sb.WriteString("|----|------|-------|-------|------|\n")

	for _, p := range ext.Parameters {
		valueStr := fmt.Sprintf("%v", p.Value)
		rangeStr := "-"
		if p.Min != 0 || p.Max != 0 {
			rangeStr = fmt.Sprintf("%.1f - %.1f", p.Min, p.Max)
		}
		if len(p.Options) > 0 {
			rangeStr = fmt.Sprintf("[%s]", strings.Join(p.Options, ", "))
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			p.ID, p.Name, valueStr, rangeStr, p.Type))
	}

	return tools.TextResult(sb.String()), nil
}

// handleEffectSet sets an effect parameter
func handleEffectSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	effect := tools.GetIntParam(req, "effect", 0)
	paramID := tools.GetIntParam(req, "param_id", 0)
	paramName := tools.GetStringParam(req, "param_name")
	value := float64(tools.GetIntParam(req, "value", -1))

	if layer == 0 || effect == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and effect are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If param_name provided, find the ID
	if paramID == 0 && paramName != "" {
		ext, err := client.GetEffectExtended(ctx, layer, effect)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		for _, p := range ext.Parameters {
			if strings.EqualFold(p.Name, paramName) {
				paramID = p.ID
				break
			}
		}
		if paramID == 0 {
			return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("parameter '%s' not found", paramName)), nil
		}
	}

	if paramID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("param_id or param_name required")), nil
	}

	// Normalize value (assume 0-100 input, convert to 0-1 for common params)
	normalizedValue := value / 100.0

	err = client.SetParameterByID(ctx, paramID, normalizedValue)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Set parameter %d to %.1f", paramID, value)), nil
}

// handleEffectMix sets an effect's mix/intensity
func handleEffectMix(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	effect := tools.GetIntParam(req, "effect", 0)
	mix := tools.GetIntParam(req, "mix", -1)

	if layer == 0 || effect == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and effect are required")), nil
	}
	if mix < 0 || mix > 100 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("mix must be between 0 and 100")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetEffectMix(ctx, layer, effect, float64(mix)/100.0)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Set effect %d mix to %d%%", effect, mix)), nil
}

// ============================================================================
// Phase 1: Clip Management Handlers
// ============================================================================

// handleClipInfo gets detailed clip information
func handleClipInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)

	if layer == 0 || column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and column are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	details, err := client.GetClipDetails(ctx, layer, column)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Clip: %s\n\n", details.Name))
	sb.WriteString(fmt.Sprintf("**Position:** Layer %d, Column %d\n", details.Layer, details.Column))
	sb.WriteString(fmt.Sprintf("**ID:** %d\n\n", details.ID))

	sb.WriteString("## Properties\n\n")
	sb.WriteString("| Property | Value |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Resolution | %dx%d |\n", details.Width, details.Height))
	sb.WriteString(fmt.Sprintf("| Duration | %.1fs |\n", details.Duration))
	sb.WriteString(fmt.Sprintf("| Framerate | %.1f fps |\n", details.Framerate))
	sb.WriteString(fmt.Sprintf("| Speed | %.2fx |\n", details.Speed))
	sb.WriteString(fmt.Sprintf("| Trigger Style | %s |\n", details.TriggerStyle))
	sb.WriteString(fmt.Sprintf("| Beat Snap | %s |\n", details.BeatSnap))
	sb.WriteString(fmt.Sprintf("| Direction | %s |\n", details.Direction))

	sb.WriteString("\n## Status\n\n")
	connStatus := "⚫ Not connected"
	if details.Connected {
		connStatus = "🔗 Connected"
	}
	playStatus := "⏸️ Paused"
	if details.Playing {
		playStatus = "▶️ Playing"
	}
	sb.WriteString(fmt.Sprintf("- %s\n", connStatus))
	sb.WriteString(fmt.Sprintf("- %s\n", playStatus))

	if details.FilePath != "" {
		sb.WriteString(fmt.Sprintf("\n**File:** `%s`\n", details.FilePath))
	}

	return tools.TextResult(sb.String()), nil
}

// handleClipLoad loads a video file into a clip slot
func handleClipLoad(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)
	filePath := tools.GetStringParam(req, "file")

	if layer == 0 || column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and column are required")), nil
	}
	if filePath == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("file path is required")), nil
	}

	// Expand home directory
	if strings.HasPrefix(filePath, "~/") {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, filePath[2:])
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.LoadClip(ctx, layer, column, filePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Loaded clip into layer %d, column %d", layer, column)), nil
}

// handleClipProperties sets clip playback properties
func handleClipProperties(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)
	triggerStyle := tools.GetStringParam(req, "trigger_style")
	beatSnap := tools.GetStringParam(req, "beat_snap")
	direction := tools.GetStringParam(req, "direction")

	if layer == 0 || column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and column are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var changes []string

	if triggerStyle != "" {
		err = client.SetClipTriggerStyle(ctx, layer, column, triggerStyle)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("trigger=%s", triggerStyle))
	}

	if beatSnap != "" {
		err = client.SetClipBeatSnap(ctx, layer, column, beatSnap)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("beatsnap=%s", beatSnap))
	}

	if direction != "" {
		err = client.SetClipDirection(ctx, layer, column, direction)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("direction=%s", direction))
	}

	if len(changes) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("no properties specified")), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Updated clip [%d,%d]: %s", layer, column, strings.Join(changes, ", "))), nil
}

// handleClipTransport controls clip transport
func handleClipTransport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)
	position := tools.GetIntParam(req, "position", -1)

	if layer == 0 || column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and column are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if position >= 0 {
		if position > 100 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("position must be between 0 and 100")), nil
		}
		err = client.SeekClip(ctx, layer, column, float64(position)/100.0)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Seeked clip [%d,%d] to %d%%", layer, column, position)), nil
	}

	return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("position required")), nil
}

// handleClipThumbnail gets clip thumbnail
func handleClipThumbnail(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	column := tools.GetIntParam(req, "column", 0)

	if layer == 0 || column == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer and column are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	data, err := client.GetClipThumbnail(ctx, layer, column)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Return as base64 encoded
	encoded := base64.StdEncoding.EncodeToString(data)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Clip Thumbnail [%d,%d]\n\n", layer, column))
	sb.WriteString(fmt.Sprintf("**Size:** %d bytes\n\n", len(data)))
	sb.WriteString("**Base64 PNG:**\n```\n")
	// Truncate for display
	if len(encoded) > 500 {
		sb.WriteString(encoded[:500])
		sb.WriteString("...")
	} else {
		sb.WriteString(encoded)
	}
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// ============================================================================
// Phase 1: Layer Group Handlers
// ============================================================================

// handleGroups lists all layer groups
func handleGroups(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	groups, err := client.GetLayerGroups(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Layer Groups\n\n")

	if len(groups) == 0 {
		sb.WriteString("*No layer groups configured*\n\n")
		sb.WriteString("Layer groups allow you to organize and control multiple layers together.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| # | Name | Opacity | Layers | Status |\n")
	sb.WriteString("|---|------|---------|--------|--------|\n")

	for _, g := range groups {
		status := "✅"
		if g.Bypassed {
			status = "⏸️ Bypassed"
		} else if g.Solo {
			status = "🎯 Solo"
		}
		layerList := ""
		if len(g.Layers) > 0 {
			layerStrs := make([]string, len(g.Layers))
			for i, l := range g.Layers {
				layerStrs[i] = fmt.Sprintf("%d", l)
			}
			layerList = strings.Join(layerStrs, ", ")
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %s | %s |\n",
			g.Index, g.Name, g.Opacity*100, layerList, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleGroupControl controls a layer group
func handleGroupControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	group := tools.GetIntParam(req, "group", 0)
	opacity := tools.GetIntParam(req, "opacity", -1)
	bypass := tools.GetBoolParam(req, "bypass", false)
	solo := tools.GetBoolParam(req, "solo", false)

	if group == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("group number is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var changes []string

	if opacity >= 0 && opacity <= 100 {
		err = client.SetLayerGroupOpacity(ctx, group, float64(opacity)/100.0)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("opacity=%d%%", opacity))
	}

	// Check if bypass was explicitly set in the request
	if tools.HasParam(req, "bypass") {
		err = client.SetLayerGroupBypass(ctx, group, bypass)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("bypass=%v", bypass))
	}

	// Check if solo was explicitly set in the request
	if tools.HasParam(req, "solo") {
		err = client.SetLayerGroupSolo(ctx, group, solo)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		changes = append(changes, fmt.Sprintf("solo=%v", solo))
	}

	if len(changes) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("no properties specified")), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Updated group %d: %s", group, strings.Join(changes, ", "))), nil
}

// ============================================================================
// Phase 1: Audio Handlers
// ============================================================================

// handleAudioTracks lists audio tracks
func handleAudioTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	tracks, err := client.GetAudioTracks(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get master audio
	master, _ := client.GetMasterAudio(ctx)

	var sb strings.Builder
	sb.WriteString("# Audio Tracks\n\n")

	// Master
	sb.WriteString("## Master\n")
	masterMuted := ""
	if master != nil && master.Muted {
		masterMuted = " 🔇 MUTED"
	}
	masterVol := 100.0
	if master != nil {
		masterVol = master.Volume * 100
	}
	sb.WriteString(fmt.Sprintf("**Volume:** %.0f%%%s\n\n", masterVol, masterMuted))

	// Layers
	sb.WriteString("## Layers\n\n")
	sb.WriteString("| # | Name | Volume | Pan | Clip | Status |\n")
	sb.WriteString("|---|------|--------|-----|------|--------|\n")

	for _, t := range tracks {
		status := "✅"
		if t.Muted {
			status = "🔇 Muted"
		} else if t.Solo {
			status = "🎯 Solo"
		}
		clip := "-"
		if t.HasClip {
			clip = t.ClipName
			if clip == "" {
				clip = "Active"
			}
		}
		panStr := "C"
		if t.Pan < -0.1 {
			panStr = fmt.Sprintf("L%.0f", -t.Pan*100)
		} else if t.Pan > 0.1 {
			panStr = fmt.Sprintf("R%.0f", t.Pan*100)
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %s | %s | %s |\n",
			t.Layer, t.Name, t.Volume*100, panStr, clip, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleAudioVolume sets audio volume
func handleAudioVolume(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	volume := tools.GetIntParam(req, "volume", -1)

	if volume < 0 || volume > 100 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("volume must be between 0 and 100")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if layer > 0 {
		err = client.SetLayerAudioVolume(ctx, layer, float64(volume)/100.0)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Set layer %d volume to %d%%", layer, volume)), nil
	}

	// Master volume
	err = client.SetMasterAudioVolume(ctx, float64(volume)/100.0)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("✅ Set master audio volume to %d%%", volume)), nil
}

// handleAudioMute mutes/unmutes audio
func handleAudioMute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	mute := tools.GetBoolParam(req, "mute", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if layer > 0 {
		err = client.SetLayerAudioMute(ctx, layer, mute)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		muteStatus := "unmuted"
		if mute {
			muteStatus = "muted"
		}
		return tools.TextResult(fmt.Sprintf("✅ Layer %d audio %s", layer, muteStatus)), nil
	}

	// Master mute
	err = client.SetMasterAudioMute(ctx, mute)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	muteStatus := "unmuted"
	if mute {
		muteStatus = "muted"
	}
	return tools.TextResult(fmt.Sprintf("✅ Master audio %s", muteStatus)), nil
}

// handleAudioPan sets audio pan
func handleAudioPan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	layer := tools.GetIntParam(req, "layer", 0)
	pan := tools.GetIntParam(req, "pan", 0)

	if layer == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("layer is required")), nil
	}
	if pan < -100 || pan > 100 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("pan must be between -100 and 100")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetLayerAudioPan(ctx, layer, float64(pan)/100.0)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	panLabel := "center"
	if pan < 0 {
		panLabel = fmt.Sprintf("%d%% left", -pan)
	} else if pan > 0 {
		panLabel = fmt.Sprintf("%d%% right", pan)
	}
	return tools.TextResult(fmt.Sprintf("✅ Set layer %d pan to %s", layer, panLabel)), nil
}

// handleBatchTrigger triggers multiple clips or columns in one call
func handleBatchTrigger(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	triggersStr, errResult := tools.RequireStringParam(req, "triggers")
	if errResult != nil {
		return errResult, nil
	}

	var triggers []struct {
		Layer  int `json:"layer"`
		Column int `json:"column"`
	}
	if err := json.Unmarshal([]byte(triggersStr), &triggers); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid JSON: %w", err)), nil
	}
	if len(triggers) == 0 {
		return tools.ErrorResult(fmt.Errorf("triggers array is empty")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Batch Trigger\n\n")

	triggered := 0
	var errors []string
	for i, t := range triggers {
		if t.Column == 0 {
			errors = append(errors, fmt.Sprintf("trigger %d: column is required", i))
			continue
		}

		if t.Layer > 0 {
			if err := client.TriggerClip(ctx, t.Layer, t.Column); err != nil {
				errors = append(errors, fmt.Sprintf("trigger %d: %v", i, err))
				continue
			}
			sb.WriteString(fmt.Sprintf("- Triggered clip layer %d, column %d\n", t.Layer, t.Column))
		} else {
			if err := client.TriggerColumn(ctx, t.Column); err != nil {
				errors = append(errors, fmt.Sprintf("trigger %d: %v", i, err))
				continue
			}
			sb.WriteString(fmt.Sprintf("- Triggered column %d\n", t.Column))
		}
		triggered++
	}

	sb.WriteString(fmt.Sprintf("\n**Triggered:** %d/%d\n", triggered, len(triggers)))
	if len(errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n**Errors:** %d\n", len(errors)))
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
