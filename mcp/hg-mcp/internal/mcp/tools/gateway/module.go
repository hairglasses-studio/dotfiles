// Package gateway provides unified entry points for domain operations.
// Gateway tools reduce token usage by consolidating related tools into single entry points.
package gateway

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for gateway tools
type Module struct{}

func (m *Module) Name() string {
	return "gateway"
}

func (m *Module) Description() string {
	return "Unified gateway tools for domain operations - saves tokens by consolidating related tools"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// DJ Gateway - Serato, Rekordbox, Traktor
		{
			Tool: mcp.NewTool("aftrs_dj",
				mcp.WithDescription("Unified DJ operations for Serato, Rekordbox, and Traktor. Actions: status, search, playlists, track_info, history, crates, export, sync"),
				mcp.WithString("software", mcp.Required(), mcp.Description("Target software: serato, rekordbox, traktor")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Operation: status, search, playlists, track_info, history, crates, export, sync")),
				mcp.WithString("query", mcp.Description("Search query or track/playlist ID (for search, track_info)")),
				mcp.WithNumber("limit", mcp.Description("Max results to return (default: 20)")),
			),
			Handler:             handleDJGateway,
			Category:            "gateway",
			Subcategory:         "dj",
			Tags:                []string{"dj", "serato", "rekordbox", "traktor", "music", "library"},
			UseCases:            []string{"Query DJ library", "Search tracks", "Manage playlists"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gateway",
		},
		// AV Gateway - Resolume, TouchDesigner, OBS
		{
			Tool: mcp.NewTool("aftrs_av",
				mcp.WithDescription("Unified AV control for Resolume, TouchDesigner, and OBS. Actions: status, health, layers, clips, trigger, effects, output, scenes"),
				mcp.WithString("software", mcp.Required(), mcp.Description("Target software: resolume, touchdesigner, obs")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Operation: status, health, layers, clips, trigger, effects, output, scenes, sources")),
				mcp.WithString("target", mcp.Description("Target layer/clip/scene name or ID")),
				mcp.WithString("value", mcp.Description("Value to set (for trigger, effects)")),
			),
			Handler:             handleAVGateway,
			Category:            "gateway",
			Subcategory:         "av",
			Tags:                []string{"av", "video", "resolume", "touchdesigner", "obs", "visuals"},
			UseCases:            []string{"Control visuals", "Switch scenes", "Trigger clips"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gateway",
			IsWrite:             true,
		},
		// Lighting Gateway - grandMA3, DMX, WLED
		{
			Tool: mcp.NewTool("aftrs_lighting",
				mcp.WithDescription("Unified lighting control for grandMA3, DMX/ArtNet, and WLED. Actions: status, health, fixtures, scenes, presets, blackout, color, effect"),
				mcp.WithString("system", mcp.Required(), mcp.Description("Target system: grandma3, dmx, wled")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Operation: status, health, fixtures, scenes, presets, blackout, color, effect, intensity")),
				mcp.WithString("target", mcp.Description("Target fixture/scene/preset name or ID")),
				mcp.WithString("value", mcp.Description("Value to set (color hex, intensity 0-100, effect name)")),
			),
			Handler:             handleLightingGateway,
			Category:            "gateway",
			Subcategory:         "lighting",
			Tags:                []string{"lighting", "dmx", "grandma3", "wled", "artnet", "fixtures"},
			UseCases:            []string{"Control lights", "Recall scenes", "Set colors"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gateway",
			IsWrite:             true,
		},
		// Audio Gateway - Ableton, Dante, MIDI
		{
			Tool: mcp.NewTool("aftrs_audio",
				mcp.WithDescription("Unified audio control for Ableton Live, Dante, and MIDI. Actions: status, transport, tracks, devices, routing, bpm"),
				mcp.WithString("system", mcp.Required(), mcp.Description("Target system: ableton, dante, midi")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Operation: status, transport, tracks, devices, routing, bpm, play, stop, record")),
				mcp.WithString("target", mcp.Description("Target track/device/route name or ID")),
				mcp.WithString("value", mcp.Description("Value to set (bpm, volume, etc)")),
			),
			Handler:             handleAudioGateway,
			Category:            "gateway",
			Subcategory:         "audio",
			Tags:                []string{"audio", "ableton", "dante", "midi", "music", "daw"},
			UseCases:            []string{"Control DAW", "Manage audio routing", "Set BPM"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gateway",
			IsWrite:             true,
		},
		// Streaming Gateway - Twitch, YouTube, NDI
		{
			Tool: mcp.NewTool("aftrs_streaming",
				mcp.WithDescription("Unified streaming control for Twitch, YouTube Live, and NDI. Actions: status, go_live, end_stream, sources, chat, viewers"),
				mcp.WithString("platform", mcp.Required(), mcp.Description("Target platform: twitch, youtube, ndi")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Operation: status, go_live, end_stream, sources, chat, viewers, title, game")),
				mcp.WithString("value", mcp.Description("Value to set (title, game, message)")),
			),
			Handler:             handleStreamingGateway,
			Category:            "gateway",
			Subcategory:         "streaming",
			Tags:                []string{"streaming", "twitch", "youtube", "ndi", "broadcast", "live"},
			UseCases:            []string{"Go live", "Monitor stream", "Update stream info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gateway",
			IsWrite:             true,
		},
	}
}

// handleDJGateway handles the aftrs_dj gateway tool
func handleDJGateway(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	software := strings.ToLower(tools.GetStringParam(req, "software"))
	action := strings.ToLower(tools.GetStringParam(req, "action"))
	query := tools.GetStringParam(req, "query")
	limit := tools.GetIntParam(req, "limit", 20)

	var sb strings.Builder

	switch software {
	case "serato":
		client, err := clients.NewSeratoClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("serato client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Serato Status\n\n")
			sb.WriteString(fmt.Sprintf("**Crates:** %d\n", status.CrateCount))
			sb.WriteString(fmt.Sprintf("**History Sessions:** %d\n", status.HistoryCount))
			sb.WriteString(fmt.Sprintf("**Library Path:** %s\n", status.LibraryPath))

		case "search":
			if query == "" {
				return tools.ErrorResult(fmt.Errorf("query required for search")), nil
			}
			results, err := client.SearchLibrary(ctx, query, limit)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("# Serato Search: %s\n\n", query))
			sb.WriteString(fmt.Sprintf("Found **%d** tracks\n\n", len(results)))
			sb.WriteString("| Title | Artist | BPM | Key |\n")
			sb.WriteString("|-------|--------|-----|-----|\n")
			for _, t := range results {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", t.Title, t.Artist, t.BPM, t.Key))
			}

		case "crates", "playlists":
			crates, err := client.GetCrates(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Serato Crates\n\n")
			sb.WriteString(fmt.Sprintf("**Total:** %d crates\n\n", len(crates)))
			for _, c := range crates {
				sb.WriteString(fmt.Sprintf("- %s (%d tracks)\n", c.Name, c.TrackCount))
			}

		case "history":
			history, err := client.GetHistory(ctx, limit)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Serato Play History\n\n")
			for i, h := range history {
				if i >= limit {
					break
				}
				sb.WriteString(fmt.Sprintf("- %s (%d tracks)\n", h.Name, h.TrackCount))
			}

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Serato Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown serato action: %s (valid: status, search, crates, history, health)", action)), nil
		}

	case "traktor":
		client, err := clients.NewTraktorClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("traktor client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Traktor Status\n\n")
			sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", status.TrackCount))
			sb.WriteString(fmt.Sprintf("**Playlists:** %d\n", status.PlaylistCount))
			sb.WriteString(fmt.Sprintf("**Collection Path:** %s\n", status.CollectionPath))

		case "search":
			if query == "" {
				return tools.ErrorResult(fmt.Errorf("query required for search")), nil
			}
			results, err := client.SearchLibrary(ctx, query, limit)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("# Traktor Search: %s\n\n", query))
			sb.WriteString(fmt.Sprintf("Found **%d** tracks\n\n", len(results)))
			sb.WriteString("| Title | Artist | BPM | Key |\n")
			sb.WriteString("|-------|--------|-----|-----|\n")
			for _, t := range results {
				sb.WriteString(fmt.Sprintf("| %s | %s | %.1f | %s |\n", t.Title, t.Artist, t.BPM, t.Key))
			}

		case "playlists":
			playlists, err := client.GetPlaylists(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Traktor Playlists\n\n")
			sb.WriteString(fmt.Sprintf("**Total:** %d playlists\n\n", len(playlists)))
			for _, p := range playlists {
				sb.WriteString(fmt.Sprintf("- %s (%d tracks)\n", p.Name, p.TrackCount))
			}

		case "history":
			history, err := client.GetHistory(ctx, limit)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Traktor History\n\n")
			for i, t := range history {
				if i >= limit {
					break
				}
				sb.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, t.Artist, t.Title))
			}

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Traktor Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown traktor action: %s (valid: status, search, playlists, history, health)", action)), nil
		}

	case "rekordbox":
		// Rekordbox uses OneLibrary API - route to existing tool via registry
		// For now, return a message pointing to the specific tools
		sb.WriteString("# Rekordbox Gateway\n\n")
		sb.WriteString("Use specific rekordbox tools:\n")
		sb.WriteString("- `aftrs_rekordbox_stats` - Library statistics\n")
		sb.WriteString("- `aftrs_rekordbox_search` - Search tracks\n")
		sb.WriteString("- `aftrs_rekordbox_playlists` - List playlists\n")
		sb.WriteString("- `aftrs_rekordbox_track_info` - Get track details\n")
		sb.WriteString("- `aftrs_rekordbox_recent` - Recent tracks\n")
		sb.WriteString("\n*Note: Rekordbox gateway integration coming soon*\n")

	default:
		return tools.ErrorResult(fmt.Errorf("unknown DJ software: %s (valid: serato, traktor, rekordbox)", software)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// handleAVGateway handles the aftrs_av gateway tool
func handleAVGateway(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	software := strings.ToLower(tools.GetStringParam(req, "software"))
	action := strings.ToLower(tools.GetStringParam(req, "action"))
	target := tools.GetStringParam(req, "target")
	value := tools.GetStringParam(req, "value")

	var sb strings.Builder

	switch software {
	case "resolume":
		client, err := clients.NewResolumeClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("resolume client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Resolume Status\n\n")
			sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status.Connected))
			sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", status.BPM))
			sb.WriteString(fmt.Sprintf("**Master Level:** %.0f%%\n", status.MasterLevel*100))

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Resolume Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))
			if len(health.Issues) > 0 {
				sb.WriteString("\n**Issues:**\n")
				for _, issue := range health.Issues {
					sb.WriteString(fmt.Sprintf("- %s\n", issue))
				}
			}

		case "layers":
			layers, err := client.GetLayers(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Resolume Layers\n\n")
			sb.WriteString("| Layer | Name | Opacity | Solo |\n")
			sb.WriteString("|-------|------|---------|------|\n")
			for _, l := range layers {
				sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %v |\n",
					l.Index, l.Name, l.Opacity*100, l.Solo))
			}

		case "clips":
			layer := 0
			if target != "" {
				fmt.Sscanf(target, "%d", &layer)
			}
			clips, err := client.GetClips(ctx, layer)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Resolume Clips\n\n")
			sb.WriteString(fmt.Sprintf("**Total:** %d clips\n\n", len(clips)))
			for _, c := range clips {
				status := " "
				if c.Connected {
					status = ">"
				}
				sb.WriteString(fmt.Sprintf("[%s] %s\n", status, c.Name))
			}

		case "trigger":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (layer/column like '1/3') required for trigger")), nil
			}
			var layer, column int
			_, err := fmt.Sscanf(target, "%d/%d", &layer, &column)
			if err != nil {
				return tools.ErrorResult(fmt.Errorf("invalid target format '%s', use layer/column (e.g. '1/3')", target)), nil
			}
			err = client.TriggerClip(ctx, layer, column)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Triggered clip at layer %d, column %d\n", layer, column))

		case "effects":
			layer := 0
			if target != "" {
				fmt.Sscanf(target, "%d", &layer)
			}
			effects, err := client.GetEffects(ctx, layer)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Resolume Effects\n\n")
			for _, e := range effects {
				status := " "
				if e.Enabled {
					status := "ON"
					_ = status
				}
				sb.WriteString(fmt.Sprintf("[%s] %s (%.0f%%)\n", status, e.Name, e.Mix*100))
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown resolume action: %s (valid: status, health, layers, clips, trigger, effects)", action)), nil
		}

	case "touchdesigner", "td":
		client, err := clients.NewTouchDesignerClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("touchdesigner client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# TouchDesigner Status\n\n")
			sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status.Connected))
			sb.WriteString(fmt.Sprintf("**FPS:** %.1f\n", status.FPS))
			sb.WriteString(fmt.Sprintf("**Cook Time:** %.2f ms\n", status.CookTime))
			sb.WriteString(fmt.Sprintf("**GPU Memory:** %s\n", status.GPUMemory))
			sb.WriteString(fmt.Sprintf("**Errors:** %d\n", status.ErrorCount))

		case "health":
			health, err := client.GetNetworkHealth(ctx, "")
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# TouchDesigner Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))
			if len(health.Recommendations) > 0 {
				sb.WriteString("\n**Recommendations:**\n")
				for _, r := range health.Recommendations {
					sb.WriteString(fmt.Sprintf("- %s\n", r))
				}
			}

		case "operators", "ops":
			path := "/project1"
			if target != "" {
				path = target
			}
			operators, err := client.GetOperators(ctx, path)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# TouchDesigner Operators\n\n")
			sb.WriteString(fmt.Sprintf("**Path:** %s\n", path))
			sb.WriteString(fmt.Sprintf("**Total:** %d operators\n\n", len(operators)))
			for _, op := range operators {
				sb.WriteString(fmt.Sprintf("- %s (%s)\n", op.Name, op.Type))
			}

		case "trigger", "pulse":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (operator/parameter like '/project1/button1:Pulse') required for trigger")), nil
			}
			// Parse operator path and parameter name from target
			parts := strings.Split(target, ":")
			operatorPath := parts[0]
			paramName := "Pulse"
			if len(parts) > 1 {
				paramName = parts[1]
			}
			err := client.PulseParameter(ctx, operatorPath, paramName)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Pulsed parameter: %s on %s\n", paramName, operatorPath))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown touchdesigner action: %s (valid: status, health, operators, trigger)", action)), nil
		}

	case "obs":
		client, err := clients.NewOBSClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("obs client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# OBS Status\n\n")
			sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status.Connected))
			sb.WriteString(fmt.Sprintf("**Streaming:** %v\n", status.Streaming))
			sb.WriteString(fmt.Sprintf("**Recording:** %v\n", status.Recording))
			sb.WriteString(fmt.Sprintf("**Current Scene:** %s\n", status.CurrentScene))

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# OBS Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

		case "scenes":
			scenes, err := client.GetScenes(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			status, _ := client.GetStatus(ctx)
			currentScene := ""
			if status != nil {
				currentScene = status.CurrentScene
			}
			sb.WriteString("# OBS Scenes\n\n")
			for _, s := range scenes {
				marker := ""
				if s.Name == currentScene {
					marker = " <-- current"
				}
				sb.WriteString(fmt.Sprintf("- %s%s\n", s.Name, marker))
			}

		case "sources":
			sources, err := client.GetSources(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# OBS Sources\n\n")
			for _, s := range sources {
				sb.WriteString(fmt.Sprintf("- %s (%s)\n", s.Name, s.Type))
			}

		case "trigger", "scene_switch":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (scene name) required")), nil
			}
			err := client.SetCurrentScene(ctx, target)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Switched to scene: %s\n", target))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown obs action: %s (valid: status, health, scenes, sources, trigger)", action)), nil
		}

	default:
		return tools.ErrorResult(fmt.Errorf("unknown AV software: %s (valid: resolume, touchdesigner, obs)", software)), nil
	}

	_ = value // silence unused warning - value used in some actions
	return tools.TextResult(sb.String()), nil
}

// handleLightingGateway handles the aftrs_lighting gateway tool
func handleLightingGateway(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	system := strings.ToLower(tools.GetStringParam(req, "system"))
	action := strings.ToLower(tools.GetStringParam(req, "action"))
	target := tools.GetStringParam(req, "target")
	value := tools.GetStringParam(req, "value")

	var sb strings.Builder

	switch system {
	case "grandma3", "gma3", "grandma":
		client, err := clients.NewGrandMA3Client()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("grandma3 client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# grandMA3 Status\n\n")
			sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status.Connected))
			sb.WriteString(fmt.Sprintf("**Host:** %s:%d\n", status.Host, status.Port))
			sb.WriteString(fmt.Sprintf("**Protocol:** %s\n", status.Protocol))

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# grandMA3 Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

		case "command":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (command) required")), nil
			}
			err := client.SendCommand(ctx, target)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Sent command: %s\n", target))

		case "blackout":
			err := client.Blackout(ctx, true)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Blackout activated\n")

		case "clear":
			err := client.ClearProgrammer(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Programmer cleared\n")

		case "go":
			// Go next cue on sequence 1
			seq := 1
			if target != "" {
				fmt.Sscanf(target, "%d", &seq)
			}
			err := client.GoNextCue(ctx, seq)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Sequence %d: GO (next cue)\n", seq))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown grandma3 action: %s (valid: status, health, command, blackout, clear, go)", action)), nil
		}

	case "dmx", "artnet":
		client, err := clients.NewLightingClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("dmx client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetDMXStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# DMX Status\n\n")
			sb.WriteString(fmt.Sprintf("**Active:** %v\n", status.Active))
			sb.WriteString(fmt.Sprintf("**Universe:** %d\n", status.Universe))
			sb.WriteString(fmt.Sprintf("**Channels:** %d\n", status.Channels))
			sb.WriteString(fmt.Sprintf("**Source:** %s\n", status.Source))

		case "health":
			health, err := client.GetHealth(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Lighting Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

		case "fixtures":
			fixtures, err := client.ListFixtures(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# DMX Fixtures\n\n")
			sb.WriteString(fmt.Sprintf("**Total:** %d fixtures\n\n", len(fixtures)))
			for _, f := range fixtures {
				sb.WriteString(fmt.Sprintf("- %s @ Ch %d (%s)\n", f.Name, f.StartChannel, f.Type))
			}

		case "scenes":
			scenes, err := client.ListScenes(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Lighting Scenes\n\n")
			for _, s := range scenes {
				sb.WriteString(fmt.Sprintf("- %s\n", s.Name))
			}

		case "blackout":
			err := client.Blackout(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Blackout activated\n")

		case "intensity":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (fixture name) required")), nil
			}
			intensity := 100
			if value != "" {
				fmt.Sscanf(value, "%d", &intensity)
			}
			err := client.SetFixtureDimmer(ctx, target, intensity)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Set %s to %d%%\n", target, intensity))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown dmx action: %s", action)), nil
		}

	case "wled":
		client, err := clients.NewWLEDClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("wled client: %w", err)), nil
		}

		switch action {
		case "status":
			devices := client.ListDevices(ctx)
			sb.WriteString("# WLED Status\n\n")
			sb.WriteString(fmt.Sprintf("**Devices Found:** %d\n\n", len(devices)))
			for _, d := range devices {
				status := "OFF"
				if d.On {
					status = "ON"
				}
				sb.WriteString(fmt.Sprintf("- %s (%s) - %s, Brightness: %d%%\n", d.Name, d.IP, status, d.Brightness))
			}

		case "discover":
			subnet := target
			if subnet == "" {
				subnet = "192.168.1"
			}
			devices, err := client.DiscoverDevices(ctx, subnet)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# WLED Discovery\n\n")
			sb.WriteString(fmt.Sprintf("**Devices Found:** %d\n\n", len(devices)))
			for _, d := range devices {
				sb.WriteString(fmt.Sprintf("- %s (%s)\n", d.Name, d.IP))
			}

		case "power":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (device IP) required")), nil
			}
			on := value == "on" || value == "1" || value == "true"
			err := client.SetPower(ctx, target, on)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Set %s power: %v\n", target, on))

		case "brightness":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (device IP) required")), nil
			}
			brightness := 100
			if value != "" {
				fmt.Sscanf(value, "%d", &brightness)
			}
			err := client.SetBrightness(ctx, target, brightness)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Set %s brightness: %d%%\n", target, brightness))

		case "color":
			if target == "" || value == "" {
				return tools.ErrorResult(fmt.Errorf("target (device IP) and value (r,g,b) required")), nil
			}
			var r, g, b int
			_, err := fmt.Sscanf(value, "%d,%d,%d", &r, &g, &b)
			if err != nil {
				return tools.ErrorResult(fmt.Errorf("invalid color format '%s', use r,g,b (e.g. '255,0,0')", value)), nil
			}
			err = client.SetColor(ctx, target, r, g, b)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Set %s color: rgb(%d,%d,%d)\n", target, r, g, b))

		case "effect":
			if target == "" || value == "" {
				return tools.ErrorResult(fmt.Errorf("target (device IP) and value (effect ID) required")), nil
			}
			var effectID int
			fmt.Sscanf(value, "%d", &effectID)
			err := client.SetEffect(ctx, target, effectID)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("Set %s effect ID: %d\n", target, effectID))

		case "effects":
			if target == "" {
				return tools.ErrorResult(fmt.Errorf("target (device IP) required")), nil
			}
			effects, err := client.GetEffects(ctx, target)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# WLED Effects\n\n")
			for _, e := range effects {
				sb.WriteString(fmt.Sprintf("%d. %s\n", e.ID, e.Name))
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown wled action: %s (valid: status, discover, power, brightness, color, effect, effects)", action)), nil
		}

	default:
		return tools.ErrorResult(fmt.Errorf("unknown lighting system: %s (valid: grandma3, dmx, wled)", system)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// handleAudioGateway handles the aftrs_audio gateway tool
func handleAudioGateway(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	system := strings.ToLower(tools.GetStringParam(req, "system"))
	action := strings.ToLower(tools.GetStringParam(req, "action"))
	target := tools.GetStringParam(req, "target")
	value := tools.GetStringParam(req, "value")

	var sb strings.Builder

	switch system {
	case "ableton", "live":
		client, err := clients.NewAbletonClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("ableton client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Ableton Live Status\n\n")
			sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status.Connected))
			if status.State != nil {
				sb.WriteString(fmt.Sprintf("**Playing:** %v\n", status.State.Playing))
				sb.WriteString(fmt.Sprintf("**Recording:** %v\n", status.State.Recording))
				sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", status.State.Tempo))
				sb.WriteString(fmt.Sprintf("**Time Signature:** %d/%d\n", status.State.TimeSignNum, status.State.TimeSignDen))
				sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", status.State.TrackCount))
				sb.WriteString(fmt.Sprintf("**Scenes:** %d\n", status.State.SceneCount))
			}

		case "transport":
			transport, err := client.GetTransportState(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Ableton Transport\n\n")
			for k, v := range transport {
				sb.WriteString(fmt.Sprintf("**%s:** %v\n", k, v))
			}

		case "tracks":
			tracks, err := client.GetTracks(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Ableton Tracks\n\n")
			sb.WriteString("| Track | Name | Volume | Muted | Armed |\n")
			sb.WriteString("|-------|------|--------|-------|-------|\n")
			for _, t := range tracks {
				sb.WriteString(fmt.Sprintf("| %d | %s | %.0f%% | %v | %v |\n",
					t.Index, t.Name, t.Volume*100, t.Mute, t.Arm))
			}

		case "bpm":
			if value != "" {
				var bpm float64
				fmt.Sscanf(value, "%f", &bpm)
				err := client.SetTempo(ctx, bpm)
				if err != nil {
					return tools.ErrorResult(err), nil
				}
				sb.WriteString(fmt.Sprintf("Set BPM to %.1f\n", bpm))
			} else {
				bpm, err := client.GetTempo(ctx)
				if err != nil {
					return tools.ErrorResult(err), nil
				}
				sb.WriteString(fmt.Sprintf("**Current BPM:** %.1f\n", bpm))
			}

		case "play":
			err := client.Play(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Playback started\n")

		case "stop":
			err := client.Stop(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Playback stopped\n")

		case "record":
			err := client.Record(ctx, true)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("Recording started\n")

		default:
			return tools.ErrorResult(fmt.Errorf("unknown ableton action: %s", action)), nil
		}

	case "dante":
		client, err := clients.NewDanteClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("dante client: %w", err)), nil
		}

		switch action {
		case "status":
			devices, err := client.GetDevices(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Dante Network Status\n\n")
			sb.WriteString(fmt.Sprintf("**Devices Found:** %d\n\n", len(devices)))
			for _, d := range devices {
				sb.WriteString(fmt.Sprintf("- %s (%s) - %d Tx, %d Rx\n",
					d.Name, d.IPAddress, len(d.TxChannels), len(d.RxChannels)))
			}

		case "routing":
			routes, err := client.GetRoutes(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Dante Routing\n\n")
			sb.WriteString("| Source | Destination | Status |\n")
			sb.WriteString("|--------|-------------|--------|\n")
			for _, r := range routes {
				status := "Active"
				if !r.Active {
					status = "Inactive"
				}
				sb.WriteString(fmt.Sprintf("| %s:%d | %s:%d | %s |\n", r.TxDevice, r.TxChannel, r.RxDevice, r.RxChannel, status))
			}

		case "devices":
			devices, err := client.GetDevices(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Dante Devices\n\n")
			for _, d := range devices {
				sb.WriteString(fmt.Sprintf("## %s\n", d.Name))
				sb.WriteString(fmt.Sprintf("- IP: %s\n", d.IPAddress))
				sb.WriteString(fmt.Sprintf("- Tx Channels: %d\n", len(d.TxChannels)))
				sb.WriteString(fmt.Sprintf("- Rx Channels: %d\n", len(d.RxChannels)))
				sb.WriteString("\n")
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown dante action: %s (valid: status, routing, devices)", action)), nil
		}

	case "midi":
		client, err := clients.NewMIDIClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("midi client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# MIDI Status\n\n")
			sb.WriteString("## Input Devices\n\n")
			for _, d := range status.InputDevices {
				connected := ""
				if d.Connected {
					connected = " ✓"
				}
				sb.WriteString(fmt.Sprintf("- %s%s\n", d.Name, connected))
			}
			sb.WriteString("\n## Output Devices\n\n")
			for _, d := range status.OutputDevices {
				connected := ""
				if d.Connected {
					connected = " ✓"
				}
				sb.WriteString(fmt.Sprintf("- %s%s\n", d.Name, connected))
			}

		case "devices":
			devices, err := client.GetDevices(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# MIDI Devices\n\n")
			sb.WriteString(fmt.Sprintf("**Total Devices:** %d\n\n", len(devices)))
			for _, d := range devices {
				connected := ""
				if d.Connected {
					connected = " ✓"
				}
				sb.WriteString(fmt.Sprintf("- %s (%s)%s\n", d.Name, d.Type, connected))
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown midi action: %s (valid: status, devices)", action)), nil
		}

	default:
		return tools.ErrorResult(fmt.Errorf("unknown audio system: %s (valid: ableton, dante, midi)", system)), nil
	}

	_ = target // silence unused warning
	return tools.TextResult(sb.String()), nil
}

// handleStreamingGateway handles the aftrs_streaming gateway tool
func handleStreamingGateway(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	platform := strings.ToLower(tools.GetStringParam(req, "platform"))
	action := strings.ToLower(tools.GetStringParam(req, "action"))
	value := tools.GetStringParam(req, "value")

	var sb strings.Builder

	switch platform {
	case "twitch":
		client, err := clients.NewTwitchClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("twitch client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# Twitch Status\n\n")
			sb.WriteString(fmt.Sprintf("**Live:** %v\n", status.IsLive))
			sb.WriteString(fmt.Sprintf("**Title:** %s\n", status.StreamTitle))
			sb.WriteString(fmt.Sprintf("**Game:** %s\n", status.GameName))
			sb.WriteString(fmt.Sprintf("**Viewers:** %d\n", status.ViewerCount))

		case "viewers":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("**Current Viewers:** %d\n", status.ViewerCount))

		case "title":
			if value == "" {
				status, _ := client.GetStatus(ctx)
				sb.WriteString(fmt.Sprintf("**Current Title:** %s\n", status.StreamTitle))
			} else {
				err := client.UpdateStreamInfo(ctx, value, "")
				if err != nil {
					return tools.ErrorResult(err), nil
				}
				sb.WriteString(fmt.Sprintf("Updated title: %s\n", value))
			}

		case "game":
			if value == "" {
				status, _ := client.GetStatus(ctx)
				sb.WriteString(fmt.Sprintf("**Current Game:** %s\n", status.GameName))
			} else {
				err := client.UpdateStreamInfo(ctx, "", value)
				if err != nil {
					return tools.ErrorResult(err), nil
				}
				sb.WriteString(fmt.Sprintf("Updated game: %s\n", value))
			}

		case "chat":
			if value != "" {
				err := client.SendChatMessage(ctx, value)
				if err != nil {
					return tools.ErrorResult(err), nil
				}
				sb.WriteString(fmt.Sprintf("Sent message: %s\n", value))
			} else {
				sb.WriteString("Use 'value' parameter to send a chat message\n")
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown twitch action: %s (valid: status, viewers, title, game, chat)", action)), nil
		}

	case "youtube", "yt":
		client, err := clients.NewYouTubeLiveClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("youtube client: %w", err)), nil
		}

		switch action {
		case "status":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# YouTube Live Status\n\n")
			sb.WriteString(fmt.Sprintf("**Live:** %v\n", status.IsLive))
			sb.WriteString(fmt.Sprintf("**Title:** %s\n", status.BroadcastTitle))
			sb.WriteString(fmt.Sprintf("**Viewers:** %d\n", status.ViewerCount))
			sb.WriteString(fmt.Sprintf("**Likes:** %d\n", status.LikeCount))

		case "viewers":
			status, err := client.GetStatus(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString(fmt.Sprintf("**Current Viewers:** %d\n", status.ViewerCount))

		case "title":
			status, _ := client.GetStatus(ctx)
			sb.WriteString(fmt.Sprintf("**Current Title:** %s\n", status.BroadcastTitle))

		case "broadcasts":
			broadcasts, err := client.GetBroadcasts(ctx, "all")
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# YouTube Broadcasts\n\n")
			for _, b := range broadcasts {
				sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", b.Title, b.ID))
			}

		default:
			return tools.ErrorResult(fmt.Errorf("unknown youtube action: %s (valid: status, viewers, title, broadcasts)", action)), nil
		}

	case "ndi":
		client, err := clients.NewNDIClient()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("ndi client: %w", err)), nil
		}

		switch action {
		case "status", "sources":
			sources, err := client.DiscoverSources(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			sb.WriteString("# NDI Sources\n\n")
			sb.WriteString(fmt.Sprintf("**Found:** %d sources\n\n", len(sources)))
			sb.WriteString("| Name | Host | Status |\n")
			sb.WriteString("|------|------|--------|\n")
			for _, s := range sources {
				sb.WriteString(fmt.Sprintf("| %s | %s | Available |\n", s.Name, s.Host))
			}

		case "health":
			sources, err := client.DiscoverSources(ctx)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			score := 100
			status := "healthy"
			if len(sources) == 0 {
				score = 50
				status = "degraded"
			}
			sb.WriteString("# NDI Health\n\n")
			sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", score))
			sb.WriteString(fmt.Sprintf("**Status:** %s\n", status))
			sb.WriteString(fmt.Sprintf("**Sources:** %d\n", len(sources)))

		default:
			return tools.ErrorResult(fmt.Errorf("unknown ndi action: %s", action)), nil
		}

	default:
		return tools.ErrorResult(fmt.Errorf("unknown streaming platform: %s (valid: twitch, youtube, ndi)", platform)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
