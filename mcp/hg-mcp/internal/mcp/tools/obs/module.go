// Package obs provides OBS Studio control tools for hg-mcp.
package obs

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for OBS integration
type Module struct{}

func (m *Module) Name() string {
	return "obs"
}

func (m *Module) Description() string {
	return "OBS Studio streaming and recording control via WebSocket"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_obs_status",
				mcp.WithDescription("Get OBS Studio status including streaming, recording, and performance metrics."),
			),
			Handler:             handleStatus,
			Category:            "obs",
			Subcategory:         "status",
			Tags:                []string{"obs", "status", "streaming", "recording"},
			UseCases:            []string{"Check OBS status", "Monitor stream health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_scenes",
				mcp.WithDescription("List all scenes in OBS."),
			),
			Handler:             handleScenes,
			Category:            "obs",
			Subcategory:         "scenes",
			Tags:                []string{"obs", "scenes", "list"},
			UseCases:            []string{"View available scenes", "Check scene order"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_scene_switch",
				mcp.WithDescription("Switch to a different scene."),
				mcp.WithString("scene", mcp.Required(), mcp.Description("Scene name to switch to")),
				mcp.WithBoolean("preview", mcp.Description("Switch preview instead of program (studio mode)")),
			),
			Handler:             handleSceneSwitch,
			Category:            "obs",
			Subcategory:         "scenes",
			Tags:                []string{"obs", "scene", "switch", "transition"},
			UseCases:            []string{"Change active scene", "Queue preview scene"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_sources",
				mcp.WithDescription("List sources in a scene or all sources."),
				mcp.WithString("scene", mcp.Description("Scene name (omit to list all sources)")),
			),
			Handler:             handleSources,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "sources", "list"},
			UseCases:            []string{"View scene sources", "Check source configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_source_visibility",
				mcp.WithDescription("Show or hide a source in a scene."),
				mcp.WithString("scene", mcp.Required(), mcp.Description("Scene containing the source")),
				mcp.WithString("source", mcp.Required(), mcp.Description("Source name")),
				mcp.WithBoolean("visible", mcp.Required(), mcp.Description("True to show, false to hide")),
			),
			Handler:             handleSourceVisibility,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "source", "visibility", "toggle"},
			UseCases:            []string{"Toggle source visibility", "Show/hide overlays"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_stream",
				mcp.WithDescription("Control streaming (start, stop, toggle)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, or toggle")),
			),
			Handler:             handleStream,
			Category:            "obs",
			Subcategory:         "streaming",
			Tags:                []string{"obs", "stream", "live", "broadcast"},
			UseCases:            []string{"Start streaming", "Stop streaming"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_record",
				mcp.WithDescription("Control recording (start, stop, pause, resume)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, pause, resume")),
			),
			Handler:             handleRecord,
			Category:            "obs",
			Subcategory:         "recording",
			Tags:                []string{"obs", "record", "capture", "video"},
			UseCases:            []string{"Start recording", "Stop recording"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_virtualcam",
				mcp.WithDescription("Control virtual camera (start, stop)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start or stop")),
			),
			Handler:             handleVirtualCam,
			Category:            "obs",
			Subcategory:         "virtualcam",
			Tags:                []string{"obs", "virtualcam", "camera", "video"},
			UseCases:            []string{"Start virtual camera", "Use OBS output in other apps"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_replay",
				mcp.WithDescription("Control replay buffer (start, stop, save)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, or save")),
			),
			Handler:             handleReplay,
			Category:            "obs",
			Subcategory:         "replay",
			Tags:                []string{"obs", "replay", "buffer", "instant"},
			UseCases:            []string{"Save instant replay", "Capture highlights"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_audio",
				mcp.WithDescription("List audio sources with volume and mute status."),
			),
			Handler:             handleAudio,
			Category:            "obs",
			Subcategory:         "audio",
			Tags:                []string{"obs", "audio", "volume", "mixer"},
			UseCases:            []string{"View audio sources", "Check audio levels"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_mute",
				mcp.WithDescription("Mute or unmute an audio source."),
				mcp.WithString("source", mcp.Required(), mcp.Description("Audio source name")),
				mcp.WithBoolean("mute", mcp.Description("True to mute, false to unmute (default: toggle)")),
			),
			Handler:             handleMute,
			Category:            "obs",
			Subcategory:         "audio",
			Tags:                []string{"obs", "audio", "mute", "toggle"},
			UseCases:            []string{"Mute microphone", "Toggle audio"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_volume",
				mcp.WithDescription("Set volume for an audio source."),
				mcp.WithString("source", mcp.Required(), mcp.Description("Audio source name")),
				mcp.WithNumber("volume", mcp.Required(), mcp.Description("Volume in dB (-100 to 26)")),
			),
			Handler:             handleVolume,
			Category:            "obs",
			Subcategory:         "audio",
			Tags:                []string{"obs", "audio", "volume", "level"},
			UseCases:            []string{"Adjust audio levels", "Set microphone volume"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_studio_mode",
				mcp.WithDescription("Control studio mode (enable, disable, transition)."),
				mcp.WithString("action", mcp.Description("Action: enable, disable, transition, or status")),
			),
			Handler:             handleStudioMode,
			Category:            "obs",
			Subcategory:         "studio",
			Tags:                []string{"obs", "studio", "preview", "transition"},
			UseCases:            []string{"Enable studio mode", "Trigger transition"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_settings",
				mcp.WithDescription("View streaming or recording settings."),
				mcp.WithString("type", mcp.Description("Settings type: stream or record (default: both)")),
			),
			Handler:             handleSettings,
			Category:            "obs",
			Subcategory:         "settings",
			Tags:                []string{"obs", "settings", "config", "output"},
			UseCases:            []string{"Check stream settings", "View recording format"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_health",
				mcp.WithDescription("Get OBS system health and recommendations."),
			),
			Handler:             handleHealth,
			Category:            "obs",
			Subcategory:         "health",
			Tags:                []string{"obs", "health", "monitoring", "status"},
			UseCases:            []string{"Check system health", "Monitor performance"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_scene_item_transform",
				mcp.WithDescription("Get or set position, rotation, and scale of a scene item."),
				mcp.WithString("scene", mcp.Required(), mcp.Description("Scene name")),
				mcp.WithString("source", mcp.Required(), mcp.Description("Source/item name within the scene")),
				mcp.WithNumber("position_x", mcp.Description("X position in pixels")),
				mcp.WithNumber("position_y", mcp.Description("Y position in pixels")),
				mcp.WithNumber("rotation", mcp.Description("Rotation in degrees (0-360)")),
				mcp.WithNumber("scale_x", mcp.Description("Horizontal scale factor (1.0 = original)")),
				mcp.WithNumber("scale_y", mcp.Description("Vertical scale factor (1.0 = original)")),
			),
			Handler:             handleSceneItemTransform,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "scene", "item", "transform", "position", "scale"},
			UseCases:            []string{"Move scene items", "Resize overlays", "Rotate sources"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_filter_toggle",
				mcp.WithDescription("Enable or disable a source filter, or list filters on a source."),
				mcp.WithString("source", mcp.Required(), mcp.Description("Source name")),
				mcp.WithString("filter", mcp.Description("Filter name (omit to list all filters)")),
				mcp.WithBoolean("enabled", mcp.Description("True to enable, false to disable")),
			),
			Handler:             handleFilterToggle,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "filter", "effect", "toggle"},
			UseCases:            []string{"Toggle source filters", "List filters on a source"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_media_control",
				mcp.WithDescription("Control media source playback (play, pause, stop, restart, seek)."),
				mcp.WithString("source", mcp.Required(), mcp.Description("Media source name")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: play, pause, stop, restart, or seek")),
				mcp.WithNumber("position", mcp.Description("Seek position in milliseconds (required for seek action)")),
			),
			Handler:             handleMediaControl,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "media", "playback", "video", "audio"},
			UseCases:            []string{"Play media sources", "Seek video position", "Pause playback"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_transition_settings",
				mcp.WithDescription("Get or set the current scene transition type and duration."),
				mcp.WithString("name", mcp.Description("Transition name to set (e.g., Fade, Cut, Slide). Omit to get current.")),
				mcp.WithNumber("duration", mcp.Description("Transition duration in milliseconds")),
			),
			Handler:             handleTransitionSettings,
			Category:            "obs",
			Subcategory:         "settings",
			Tags:                []string{"obs", "transition", "settings", "scene"},
			UseCases:            []string{"Change transition type", "Set transition duration"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "obs",
		},
		{
			Tool: mcp.NewTool("aftrs_obs_screenshot",
				mcp.WithDescription("Capture a screenshot of a source or the program output."),
				mcp.WithString("source", mcp.Description("Source name (omit for program output)")),
				mcp.WithString("format", mcp.Description("Image format: png (default) or jpg")),
			),
			Handler:             handleScreenshot,
			Category:            "obs",
			Subcategory:         "sources",
			Tags:                []string{"obs", "screenshot", "capture", "image"},
			UseCases:            []string{"Capture source screenshot", "Take output snapshot"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "obs",
		},
	}
}

var getClient = tools.LazyClient(clients.GetOBSClient)

// handleStatus handles the aftrs_obs_status tool
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
	sb.WriteString("# OBS Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s:%d\n\n", client.Host(), client.Port()))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open OBS Studio\n")
		sb.WriteString("2. Go to Tools → WebSocket Server Settings\n")
		sb.WriteString("3. Enable WebSocket server on port 4455\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export OBS_HOST=localhost\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n\n")
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", status.Version))

	// Streaming status
	streamStatus := "Offline"
	if status.Streaming {
		streamStatus = fmt.Sprintf("Live (%s)", status.StreamTime)
	}
	sb.WriteString(fmt.Sprintf("**Streaming:** %s\n", streamStatus))

	// Recording status
	recordStatus := "Stopped"
	if status.Recording {
		recordStatus = fmt.Sprintf("Recording (%s)", status.RecordTime)
	}
	sb.WriteString(fmt.Sprintf("**Recording:** %s\n", recordStatus))

	// Virtual cam
	vcamStatus := "Off"
	if status.VirtualCam {
		vcamStatus = "On"
	}
	sb.WriteString(fmt.Sprintf("**Virtual Cam:** %s\n", vcamStatus))

	if status.CurrentScene != "" {
		sb.WriteString(fmt.Sprintf("**Current Scene:** %s\n", status.CurrentScene))
	}

	if status.Streaming {
		sb.WriteString("\n## Performance\n\n")
		sb.WriteString("| Metric | Value |\n")
		sb.WriteString("|--------|-------|\n")
		sb.WriteString(fmt.Sprintf("| Bitrate | %d kbps |\n", status.OutputBitrate))
		sb.WriteString(fmt.Sprintf("| Dropped Frames | %d / %d |\n", status.DroppedFrames, status.TotalFrames))
		sb.WriteString(fmt.Sprintf("| CPU Usage | %.1f%% |\n", status.CPUUsage))
	}

	return tools.TextResult(sb.String()), nil
}

// handleScenes handles the aftrs_obs_scenes tool
func handleScenes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	scenes, err := client.GetScenes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	currentScene, _ := client.GetCurrentScene(ctx)

	var sb strings.Builder
	sb.WriteString("# OBS Scenes\n\n")

	if len(scenes) == 0 {
		sb.WriteString("No scenes found.\n\n")
		sb.WriteString("*Note: Requires OBS WebSocket connection to list scenes.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** scenes:\n\n", len(scenes)))
	sb.WriteString("| # | Name | Sources | Status |\n")
	sb.WriteString("|---|------|---------|--------|\n")

	for _, scene := range scenes {
		status := ""
		if scene.Name == currentScene {
			status = "Active"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s |\n", scene.Index, scene.Name, len(scene.Sources), status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSceneSwitch handles the aftrs_obs_scene_switch tool
func handleSceneSwitch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene, errResult := tools.RequireStringParam(req, "scene")
	if errResult != nil {
		return errResult, nil
	}
	preview := tools.GetBoolParam(req, "preview", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if preview {
		err = client.SetPreviewScene(ctx, scene)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set preview to: %s", scene)), nil
	}

	err = client.SetCurrentScene(ctx, scene)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(fmt.Sprintf("Switched to scene: %s", scene)), nil
}

// handleSources handles the aftrs_obs_sources tool
func handleSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene := tools.GetStringParam(req, "scene")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder

	if scene != "" {
		sources, err := client.GetSceneSources(ctx, scene)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString(fmt.Sprintf("# Sources in %s\n\n", scene))

		if len(sources) == 0 {
			sb.WriteString("No sources found.\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString(fmt.Sprintf("Found **%d** sources:\n\n", len(sources)))
		sb.WriteString("| Name | Type | Visible | Locked |\n")
		sb.WriteString("|------|------|---------|--------|\n")

		for _, src := range sources {
			visible := "Yes"
			if !src.Visible {
				visible = "No"
			}
			locked := "No"
			if src.Locked {
				locked = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", src.Name, src.Type, visible, locked))
		}
	} else {
		sources, err := client.GetSources(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString("# All OBS Sources\n\n")

		if len(sources) == 0 {
			sb.WriteString("No sources found.\n\n")
			sb.WriteString("*Note: Requires OBS WebSocket connection to list sources.*\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString(fmt.Sprintf("Found **%d** sources:\n\n", len(sources)))
		sb.WriteString("| Name | Type | Kind |\n")
		sb.WriteString("|------|------|------|\n")

		for _, src := range sources {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", src.Name, src.Type, src.Kind))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSourceVisibility handles the aftrs_obs_source_visibility tool
func handleSourceVisibility(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene, errResult := tools.RequireStringParam(req, "scene")
	if errResult != nil {
		return errResult, nil
	}
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}
	visible := tools.GetBoolParam(req, "visible", true)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetSourceVisibility(ctx, scene, source, visible)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "visible"
	if !visible {
		status = "hidden"
	}
	return tools.TextResult(fmt.Sprintf("%s is now %s in %s", source, status, scene)), nil
}

// handleStream handles the aftrs_obs_stream tool
func handleStream(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartStreaming(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Streaming started"), nil

	case "stop":
		err = client.StopStreaming(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Streaming stopped"), nil

	case "toggle":
		status, _ := client.GetStatus(ctx)
		if status.Streaming {
			err = client.StopStreaming(ctx)
			return tools.TextResult("Streaming stopped"), nil
		}
		err = client.StartStreaming(ctx)
		return tools.TextResult("Streaming started"), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use start, stop, or toggle)", action)), nil
	}
}

// handleRecord handles the aftrs_obs_record tool
func handleRecord(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Recording started"), nil

	case "stop":
		err = client.StopRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Recording stopped"), nil

	case "pause":
		err = client.PauseRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Recording paused"), nil

	case "resume":
		err = client.ResumeRecording(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Recording resumed"), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use start, stop, pause, or resume)", action)), nil
	}
}

// handleVirtualCam handles the aftrs_obs_virtualcam tool
func handleVirtualCam(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartVirtualCam(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Virtual camera started"), nil

	case "stop":
		err = client.StopVirtualCam(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Virtual camera stopped"), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use start or stop)", action)), nil
	}
}

// handleReplay handles the aftrs_obs_replay tool
func handleReplay(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartReplayBuffer(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Replay buffer started"), nil

	case "stop":
		err = client.StopReplayBuffer(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Replay buffer stopped"), nil

	case "save":
		err = client.SaveReplayBuffer(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Replay saved"), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use start, stop, or save)", action)), nil
	}
}

// handleAudio handles the aftrs_obs_audio tool
func handleAudio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sources, err := client.GetAudioSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# OBS Audio Sources\n\n")

	if len(sources) == 0 {
		sb.WriteString("No audio sources found.\n\n")
		sb.WriteString("*Note: Requires OBS WebSocket connection to list audio sources.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** audio sources:\n\n", len(sources)))
	sb.WriteString("| Name | Type | Volume | Muted | Monitor |\n")
	sb.WriteString("|------|------|--------|-------|--------|\n")

	for _, src := range sources {
		muted := "No"
		if src.Muted {
			muted = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %.1f dB | %s | %s |\n", src.Name, src.Type, src.Volume, muted, src.MonitorType))
	}

	return tools.TextResult(sb.String()), nil
}

// handleMute handles the aftrs_obs_mute tool
func handleMute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}
	mute := tools.GetBoolParam(req, "mute", true)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetSourceMute(ctx, source, mute)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "muted"
	if !mute {
		status = "unmuted"
	}
	return tools.TextResult(fmt.Sprintf("%s is now %s", source, status)), nil
}

// handleVolume handles the aftrs_obs_volume tool
func handleVolume(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}
	volume := float64(tools.GetIntParam(req, "volume", 0))

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetSourceVolume(ctx, source, volume)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set %s volume to %.1f dB", source, volume)), nil
}

// handleStudioMode handles the aftrs_obs_studio_mode tool
func handleStudioMode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "status")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "enable":
		err = client.SetStudioModeEnabled(ctx, true)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Studio mode enabled"), nil

	case "disable":
		err = client.SetStudioModeEnabled(ctx, false)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Studio mode disabled"), nil

	case "transition":
		err = client.TriggerStudioModeTransition(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transition triggered"), nil

	case "status":
		enabled, err := client.GetStudioModeEnabled(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		status := "disabled"
		if enabled {
			status = "enabled"
		}
		return tools.TextResult(fmt.Sprintf("Studio mode is %s", status)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use enable, disable, transition, or status)", action)), nil
	}
}

// handleSettings handles the aftrs_obs_settings tool
func handleSettings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	settingsType := tools.GetStringParam(req, "type")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder

	if settingsType == "" || settingsType == "stream" {
		streamSettings, err := client.GetStreamSettings(ctx)
		if err == nil {
			sb.WriteString("# Stream Settings\n\n")
			sb.WriteString("| Setting | Value |\n")
			sb.WriteString("|---------|-------|\n")
			sb.WriteString(fmt.Sprintf("| Service | %s |\n", streamSettings.Service))
			sb.WriteString(fmt.Sprintf("| Protocol | %s |\n", streamSettings.Protocol))
			sb.WriteString(fmt.Sprintf("| Bitrate | %d kbps |\n", streamSettings.Bitrate))
			sb.WriteString(fmt.Sprintf("| Encoder | %s |\n", streamSettings.Encoder))
			sb.WriteString("\n")
		}
	}

	if settingsType == "" || settingsType == "record" {
		recordSettings, err := client.GetRecordSettings(ctx)
		if err == nil {
			sb.WriteString("# Recording Settings\n\n")
			sb.WriteString("| Setting | Value |\n")
			sb.WriteString("|---------|-------|\n")
			sb.WriteString(fmt.Sprintf("| Path | %s |\n", recordSettings.Path))
			sb.WriteString(fmt.Sprintf("| Format | %s |\n", recordSettings.Format))
			sb.WriteString(fmt.Sprintf("| Quality | %s |\n", recordSettings.Quality))
			sb.WriteString(fmt.Sprintf("| Encoder | %s |\n", recordSettings.Encoder))
		}
	}

	if sb.Len() == 0 {
		sb.WriteString("No settings available.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleHealth handles the aftrs_obs_health tool
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
	sb.WriteString("# OBS Health\n\n")

	// Status emoji
	statusEmoji := ""
	if health.Status == "degraded" {
		statusEmoji = ""
	} else if health.Status == "critical" {
		statusEmoji = ""
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Scenes | %d |\n", health.SceneCount))
	sb.WriteString(fmt.Sprintf("| Sources | %d |\n", health.SourceCount))
	sb.WriteString(fmt.Sprintf("| CPU Usage | %.1f%% |\n", health.CPUUsage))
	sb.WriteString(fmt.Sprintf("| Dropped Frames | %.2f%% |\n", health.DroppedPercent))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
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

// handleSceneItemTransform handles the aftrs_obs_scene_item_transform tool
func handleSceneItemTransform(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene := tools.GetStringParam(req, "scene")
	source := tools.GetStringParam(req, "source")
	if scene == "" || source == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("scene and source are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Check if any transform params were provided — if so, set; otherwise, get
	hasSet := false
	transform := &clients.OBSSceneItemTransform{ScaleX: 1.0, ScaleY: 1.0}

	if v := tools.GetFloatParam(req, "position_x", -999999); v != -999999 {
		transform.PositionX = v
		hasSet = true
	}
	if v := tools.GetFloatParam(req, "position_y", -999999); v != -999999 {
		transform.PositionY = v
		hasSet = true
	}
	if v := tools.GetFloatParam(req, "rotation", -999999); v != -999999 {
		transform.Rotation = v
		hasSet = true
	}
	if v := tools.GetFloatParam(req, "scale_x", -999999); v != -999999 {
		transform.ScaleX = v
		hasSet = true
	}
	if v := tools.GetFloatParam(req, "scale_y", -999999); v != -999999 {
		transform.ScaleY = v
		hasSet = true
	}

	if hasSet {
		err = client.SetSceneItemTransform(ctx, scene, source, transform)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Updated transform for %s in %s", source, scene)), nil
	}

	// Get current transform
	current, err := client.GetSceneItemTransform(ctx, scene, source)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(current), nil
}

// handleFilterToggle handles the aftrs_obs_filter_toggle tool
func handleFilterToggle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	filter := tools.GetStringParam(req, "filter")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If no filter specified, list all filters
	if filter == "" {
		filters, err := client.GetSourceFilters(ctx, source)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Filters on %s\n\n", source))

		if len(filters) == 0 {
			sb.WriteString("No filters found.\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString("| # | Name | Kind | Enabled |\n")
		sb.WriteString("|---|------|------|---------|\n")
		for _, f := range filters {
			enabled := "Yes"
			if !f.Enabled {
				enabled = "No"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", f.Index, f.Name, f.Kind, enabled))
		}
		return tools.TextResult(sb.String()), nil
	}

	// Toggle filter
	enabled := tools.GetBoolParam(req, "enabled", true)
	err = client.SetSourceFilterEnabled(ctx, source, filter, enabled)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status := "enabled"
	if !enabled {
		status = "disabled"
	}
	return tools.TextResult(fmt.Sprintf("Filter %s on %s is now %s", filter, source, status)), nil
}

// handleMediaControl handles the aftrs_obs_media_control tool
func handleMediaControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "play", "pause", "stop", "restart":
		err = client.ControlMedia(ctx, source, action)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Media %s: %s", source, action)), nil

	case "seek":
		position := tools.GetIntParam(req, "position", 0)
		err = client.SeekMedia(ctx, source, position)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Media %s: seeked to %dms", source, position)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use play, pause, stop, restart, or seek)", action)), nil
	}
}

// handleTransitionSettings handles the aftrs_obs_transition_settings tool
func handleTransitionSettings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If no name, get current transition
	if name == "" {
		info, err := client.GetCurrentTransition(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(info), nil
	}

	// Set transition
	duration := tools.GetIntParam(req, "duration", 0)
	err = client.SetTransition(ctx, name, duration)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	msg := fmt.Sprintf("Transition set to: %s", name)
	if duration > 0 {
		msg += fmt.Sprintf(" (%dms)", duration)
	}
	return tools.TextResult(msg), nil
}

// handleScreenshot handles the aftrs_obs_screenshot tool
func handleScreenshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	format := tools.OptionalStringParam(req, "format", "png")

	if format != "png" && format != "jpg" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid format: %s (use png or jpg)", format)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	screenshot, err := client.TakeScreenshot(ctx, source, format)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	target := "program output"
	if source != "" {
		target = source
	}

	result := map[string]interface{}{
		"source":     target,
		"format":     screenshot.Format,
		"width":      screenshot.Width,
		"height":     screenshot.Height,
		"has_data":   screenshot.ImageData != "",
		"image_data": screenshot.ImageData,
	}
	return tools.JSONResult(result), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
