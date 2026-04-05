// Package ableton provides MCP tools for Ableton Live control via AbletonOSC.
package ableton

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Ableton tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "ableton"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Ableton Live control via AbletonOSC for transport, clips, tracks, and devices"
}

// Tools returns the Ableton tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ableton_status",
				mcp.WithDescription("Get Ableton Live status including tempo, playing state, and track count"),
			),
			Handler:     handleAbletonStatus,
			Category:    "ableton",
			Subcategory: "status",
			Tags:        []string{"ableton", "live", "daw", "status"},
			UseCases:    []string{"Check Live status", "Get current tempo", "View session info"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_transport",
				mcp.WithDescription("Control Ableton Live transport (play, stop, record)"),
				mcp.WithString("action", mcp.Required(), mcp.Description("Transport action to perform"), mcp.Enum("play", "stop", "continue", "record_on", "record_off")),
			),
			Handler:     handleAbletonTransport,
			Category:    "ableton",
			Subcategory: "transport",
			Tags:        []string{"ableton", "transport", "play", "stop", "record"},
			UseCases:    []string{"Start playback", "Stop playback", "Toggle recording"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_tempo",
				mcp.WithDescription("Get or set Ableton Live tempo"),
				mcp.WithNumber("bpm", mcp.Description("Tempo to set (20-999 BPM). Omit to get current tempo.")),
			),
			Handler:     handleAbletonTempo,
			Category:    "ableton",
			Subcategory: "transport",
			Tags:        []string{"ableton", "tempo", "bpm"},
			UseCases:    []string{"Get current tempo", "Set tempo", "Sync BPM"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_tracks",
				mcp.WithDescription("List all tracks in Ableton Live session"),
			),
			Handler:     handleAbletonTracks,
			Category:    "ableton",
			Subcategory: "tracks",
			Tags:        []string{"ableton", "tracks", "list"},
			UseCases:    []string{"List all tracks", "View track properties", "Get track names"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_track",
				mcp.WithDescription("Get or modify a specific track (mute, solo, volume, pan)"),
				mcp.WithNumber("index", mcp.Required(), mcp.Description("Track index (0-based)")),
				mcp.WithBoolean("mute", mcp.Description("Set mute state")),
				mcp.WithBoolean("solo", mcp.Description("Set solo state")),
				mcp.WithBoolean("arm", mcp.Description("Set arm state")),
				mcp.WithNumber("volume", mcp.Description("Set volume (0.0-1.0)")),
				mcp.WithNumber("pan", mcp.Description("Set pan (-1.0 to 1.0)")),
			),
			Handler:     handleAbletonTrack,
			Category:    "ableton",
			Subcategory: "tracks",
			Tags:        []string{"ableton", "track", "mute", "solo", "volume"},
			UseCases:    []string{"Mute/solo tracks", "Adjust volume", "Arm for recording"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_clips",
				mcp.WithDescription("List clips in a track"),
				mcp.WithNumber("track_index", mcp.Required(), mcp.Description("Track index (0-based)")),
			),
			Handler:     handleAbletonClips,
			Category:    "ableton",
			Subcategory: "clips",
			Tags:        []string{"ableton", "clips", "list"},
			UseCases:    []string{"List track clips", "View clip slots", "Check clip status"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_clip_fire",
				mcp.WithDescription("Trigger a specific clip"),
				mcp.WithNumber("track_index", mcp.Required(), mcp.Description("Track index (0-based)")),
				mcp.WithNumber("slot_index", mcp.Required(), mcp.Description("Clip slot index (0-based)")),
			),
			Handler:     handleAbletonClipFire,
			Category:    "ableton",
			Subcategory: "clips",
			Tags:        []string{"ableton", "clip", "fire", "trigger"},
			UseCases:    []string{"Trigger clip", "Launch clip", "Fire specific slot"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_scene_fire",
				mcp.WithDescription("Trigger a scene (horizontal row of clips)"),
				mcp.WithNumber("scene_index", mcp.Required(), mcp.Description("Scene index (0-based)")),
			),
			Handler:     handleAbletonSceneFire,
			Category:    "ableton",
			Subcategory: "scenes",
			Tags:        []string{"ableton", "scene", "fire", "trigger"},
			UseCases:    []string{"Trigger scene", "Launch scene row", "Fire all clips in row"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_devices",
				mcp.WithDescription("List devices on a track"),
				mcp.WithNumber("track_index", mcp.Required(), mcp.Description("Track index (0-based)")),
			),
			Handler:     handleAbletonDevices,
			Category:    "ableton",
			Subcategory: "devices",
			Tags:        []string{"ableton", "devices", "instruments", "effects"},
			UseCases:    []string{"List track devices", "View effects chain", "Find instruments"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_device",
				mcp.WithDescription("Get or set device parameters"),
				mcp.WithNumber("track_index", mcp.Required(), mcp.Description("Track index (0-based)")),
				mcp.WithNumber("device_index", mcp.Required(), mcp.Description("Device index (0-based)")),
				mcp.WithNumber("param_index", mcp.Description("Parameter index to set (optional)")),
				mcp.WithNumber("value", mcp.Description("Parameter value to set (0.0-1.0)")),
			),
			Handler:     handleAbletonDevice,
			Category:    "ableton",
			Subcategory: "devices",
			Tags:        []string{"ableton", "device", "parameters", "automation"},
			UseCases:    []string{"Get device parameters", "Automate effects", "Control instruments"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_cue_points",
				mcp.WithDescription("Get or jump to arrangement cue points"),
				mcp.WithNumber("jump_to", mcp.Description("Cue point index to jump to (optional)")),
			),
			Handler:     handleAbletonCuePoints,
			Category:    "ableton",
			Subcategory: "arrangement",
			Tags:        []string{"ableton", "cue", "arrangement", "markers"},
			UseCases:    []string{"List cue points", "Jump to marker", "Navigate arrangement"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_ableton_health",
				mcp.WithDescription("Check Ableton Live connection health and AbletonOSC status"),
			),
			Handler:     handleAbletonHealth,
			Category:    "ableton",
			Subcategory: "status",
			Tags:        []string{"ableton", "health", "diagnostics", "troubleshooting"},
			UseCases:    []string{"Check connection", "Diagnose issues", "Verify AbletonOSC"},
			Complexity:  tools.ComplexitySimple,
		},
	}
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "ableton"
	}
	return allTools
}

var getAbletonClient = tools.LazyClient(clients.NewAbletonClient)

func handleAbletonStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleAbletonTransport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	var actionErr error
	switch action {
	case "play":
		actionErr = client.Play(ctx)
	case "stop":
		actionErr = client.Stop(ctx)
	case "continue":
		actionErr = client.Continue(ctx)
	case "record_on":
		actionErr = client.Record(ctx, true)
	case "record_off":
		actionErr = client.Record(ctx, false)
	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s", action)), nil
	}

	if actionErr != nil {
		return tools.ErrorResult(fmt.Errorf("failed to %s: %w", action, actionErr)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"action":  action,
		"message": fmt.Sprintf("Transport action '%s' executed", action),
	}), nil
}

func handleAbletonTempo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	bpm := tools.GetFloatParam(req, "bpm", 0)

	if bpm > 0 {
		if err := client.SetTempo(ctx, bpm); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set tempo: %w", err)), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"success": true,
			"tempo":   bpm,
			"message": fmt.Sprintf("Tempo set to %.2f BPM", bpm),
		}), nil
	}

	tempo, err := client.GetTempo(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tempo: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tempo": tempo,
	}), nil
}

func handleAbletonTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	tracks, err := client.GetTracks(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tracks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleAbletonTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	index := tools.GetIntParam(req, "index", -1)
	if index < 0 {
		return tools.ErrorResult(fmt.Errorf("index is required")), nil
	}

	args, _ := req.Params.Arguments.(map[string]interface{})
	changes := make(map[string]interface{})

	if mute, ok := args["mute"].(bool); ok {
		if err := client.SetTrackMute(ctx, index, mute); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set mute: %w", err)), nil
		}
		changes["mute"] = mute
	}

	if solo, ok := args["solo"].(bool); ok {
		if err := client.SetTrackSolo(ctx, index, solo); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set solo: %w", err)), nil
		}
		changes["solo"] = solo
	}

	if arm, ok := args["arm"].(bool); ok {
		if err := client.SetTrackArm(ctx, index, arm); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set arm: %w", err)), nil
		}
		changes["arm"] = arm
	}

	if volume, ok := args["volume"].(float64); ok {
		if err := client.SetTrackVolume(ctx, index, volume); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set volume: %w", err)), nil
		}
		changes["volume"] = volume
	}

	if pan, ok := args["pan"].(float64); ok {
		if err := client.SetTrackPan(ctx, index, pan); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set pan: %w", err)), nil
		}
		changes["pan"] = pan
	}

	track, err := client.GetTrack(ctx, index)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"track":   track,
		"changes": changes,
	}), nil
}

func handleAbletonClips(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	trackIndex := tools.GetIntParam(req, "track_index", -1)
	if trackIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("track_index is required")), nil
	}

	clips, err := client.GetClips(ctx, trackIndex)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get clips: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"track_index": trackIndex,
		"clips":       clips,
		"count":       len(clips),
	}), nil
}

func handleAbletonClipFire(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	trackIndex := tools.GetIntParam(req, "track_index", -1)
	if trackIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("track_index is required")), nil
	}

	slotIndex := tools.GetIntParam(req, "slot_index", -1)
	if slotIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("slot_index is required")), nil
	}

	if err := client.FireClip(ctx, trackIndex, slotIndex); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fire clip: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":     true,
		"track_index": trackIndex,
		"slot_index":  slotIndex,
		"message":     fmt.Sprintf("Clip fired: track %d, slot %d", trackIndex, slotIndex),
	}), nil
}

func handleAbletonSceneFire(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	sceneIndex := tools.GetIntParam(req, "scene_index", -1)
	if sceneIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("scene_index is required")), nil
	}

	if err := client.FireScene(ctx, sceneIndex); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fire scene: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":     true,
		"scene_index": sceneIndex,
		"message":     fmt.Sprintf("Scene %d fired", sceneIndex),
	}), nil
}

func handleAbletonDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	trackIndex := tools.GetIntParam(req, "track_index", -1)
	if trackIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("track_index is required")), nil
	}

	devices, err := client.GetDevices(ctx, trackIndex)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get devices: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"track_index": trackIndex,
		"devices":     devices,
		"count":       len(devices),
	}), nil
}

func handleAbletonDevice(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	trackIndex := tools.GetIntParam(req, "track_index", -1)
	if trackIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("track_index is required")), nil
	}

	deviceIndex := tools.GetIntParam(req, "device_index", -1)
	if deviceIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("device_index is required")), nil
	}

	paramIndex := tools.GetIntParam(req, "param_index", -1)
	value := tools.GetFloatParam(req, "value", -1)

	if paramIndex >= 0 && value >= 0 {
		if err := client.SetDeviceParameter(ctx, trackIndex, deviceIndex, paramIndex, value); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set parameter: %w", err)), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"success":      true,
			"track_index":  trackIndex,
			"device_index": deviceIndex,
			"param_index":  paramIndex,
			"value":        value,
			"message":      "Parameter set successfully",
		}), nil
	}

	params, err := client.GetDeviceParameters(ctx, trackIndex, deviceIndex)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get parameters: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"track_index":  trackIndex,
		"device_index": deviceIndex,
		"parameters":   params,
		"count":        len(params),
	}), nil
}

func handleAbletonCuePoints(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	jumpTo := tools.GetIntParam(req, "jump_to", -1)

	if jumpTo >= 0 {
		if err := client.JumpToCue(ctx, jumpTo); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to jump to cue: %w", err)), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"success":   true,
			"cue_index": jumpTo,
			"message":   fmt.Sprintf("Jumped to cue point %d", jumpTo),
		}), nil
	}

	cues, err := client.GetCuePoints(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get cue points: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"cue_points": cues,
		"count":      len(cues),
	}), nil
}

func handleAbletonHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getAbletonClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ableton client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
