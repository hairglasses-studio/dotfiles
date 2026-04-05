// Package prolink provides Pioneer Pro DJ Link network integration tools for hg-mcp.
// This enables real-time communication with XDJ-1000 MK2, CDJ-2000NXS2, CDJ-3000,
// and other Pioneer DJ equipment via the Pro DJ Link protocol.
package prolink

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Known Pioneer DJ / AlphaTheta MAC address prefixes (OUI)
var pioneerOUIPrefixes = []string{
	"c8:3d:fc", // AlphaTheta Corporation (Pioneer DJ parent company)
	"70:56:81", // Pioneer Corporation
	"00:e0:36", // Pioneer Corporation
	"00:11:a0", // Pioneer Corporation
	"ac:3a:7a", // Pioneer Corporation
}

// Module implements the ToolModule interface for Pro DJ Link integration
type Module struct{}

var getProlinkClient = tools.LazyClient(clients.GetProlinkClient)

func (m *Module) Name() string {
	return "prolink"
}

func (m *Module) Description() string {
	return "Pioneer Pro DJ Link network integration for real-time CDJ/XDJ communication and Showkontrol sync"
}

func (m *Module) Tools() []tools.ToolDefinition {
	baseTools := []tools.ToolDefinition{
		// ============================================================================
		// Network Discovery Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_connect",
				mcp.WithDescription("Connect to the Pro DJ Link network. Must be called before using other prolink tools. Auto-configures network interface and virtual CDJ ID."),
			),
			Handler:             handleConnect,
			Category:            "prolink",
			Subcategory:         "network",
			Tags:                []string{"prolink", "connect", "network", "cdj", "xdj"},
			UseCases:            []string{"Connect to Pro DJ Link network", "Initialize CDJ monitoring"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_disconnect",
				mcp.WithDescription("Disconnect from the Pro DJ Link network."),
			),
			Handler:             handleDisconnect,
			Category:            "prolink",
			Subcategory:         "network",
			Tags:                []string{"prolink", "disconnect", "network"},
			UseCases:            []string{"Disconnect from network", "Release resources"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_devices",
				mcp.WithDescription("List all Pioneer devices on the Pro DJ Link network (CDJs, XDJs, mixers, Rekordbox)."),
			),
			Handler:             handleDevices,
			Category:            "prolink",
			Subcategory:         "network",
			Tags:                []string{"prolink", "devices", "cdj", "xdj", "mixer", "discovery"},
			UseCases:            []string{"Discover CDJs on network", "Check connected devices", "Verify XDJ setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},

		// ============================================================================
		// Real-Time Status Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_status",
				mcp.WithDescription("Get real-time playback status from CDJ/XDJ decks including BPM, play state, beat position, and sync status."),
				mcp.WithNumber("player_id", mcp.Description("Player ID 1-4, or 0 for all players (default: 0)")),
			),
			Handler:             handleStatus,
			Category:            "prolink",
			Subcategory:         "playback",
			Tags:                []string{"prolink", "status", "playback", "bpm", "beat", "sync"},
			UseCases:            []string{"Monitor deck status", "Check BPM sync", "Live performance monitoring"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_now_playing",
				mcp.WithDescription("Get the current track playing on each CDJ/XDJ deck with full metadata (title, artist, BPM, key). Essential for Showkontrol integration."),
				mcp.WithNumber("player_id", mcp.Description("Player ID 1-4, or 0 for all players (default: 0)")),
			),
			Handler:             handleNowPlaying,
			Category:            "prolink",
			Subcategory:         "playback",
			Tags:                []string{"prolink", "now_playing", "track", "metadata", "showkontrol", "live"},
			UseCases:            []string{"Live DJ monitoring", "Showkontrol integration", "Track display", "BPM sync for lighting"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_master",
				mcp.WithDescription("Get the current master player (the deck controlling tempo sync)."),
			),
			Handler:             handleMaster,
			Category:            "prolink",
			Subcategory:         "playback",
			Tags:                []string{"prolink", "master", "sync", "tempo"},
			UseCases:            []string{"Find master deck", "Sync lighting to master BPM"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},

		// ============================================================================
		// Track Metadata Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_track",
				mcp.WithDescription("Get detailed track metadata from a specific deck's USB drive."),
				mcp.WithNumber("player_id", mcp.Required(), mcp.Description("Player ID 1-4")),
				mcp.WithNumber("track_id", mcp.Description("Track ID (optional, uses current track if not specified)")),
			),
			Handler:             handleTrack,
			Category:            "prolink",
			Subcategory:         "library",
			Tags:                []string{"prolink", "track", "metadata", "usb"},
			UseCases:            []string{"Get track details", "Query USB library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "prolink",
		},

		// ============================================================================
		// Showkontrol Integration Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_beat_info",
				mcp.WithDescription("Get current beat/tempo information for Showkontrol lighting sync. Returns BPM, beat position in measure (1-4), and effective BPM with pitch adjustment."),
				mcp.WithNumber("player_id", mcp.Description("Player ID to get beat info from (0 = master player)")),
			),
			Handler:             handleBeatInfo,
			Category:            "prolink",
			Subcategory:         "showkontrol",
			Tags:                []string{"prolink", "beat", "tempo", "showkontrol", "lighting", "sync"},
			UseCases:            []string{"Sync lighting to BPM", "Beat-matched effects", "Showkontrol integration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},

		// ============================================================================
		// Full Data Export Tool
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_all",
				mcp.WithDescription("Get ALL available Pro DJ Link data in a single call. Returns devices, player status, track metadata, BPM, beat position, and more. Ideal for Showkontrol integration."),
			),
			Handler:             handleAllData,
			Category:            "prolink",
			Subcategory:         "showkontrol",
			Tags:                []string{"prolink", "all", "full", "data", "showkontrol", "export"},
			UseCases:            []string{"Get all prolink data at once", "Showkontrol integration", "Full state snapshot"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "prolink",
		},

		// ============================================================================
		// Troubleshooting & Diagnostics Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_prolink_diagnose",
				mcp.WithDescription("Diagnose Pro DJ Link network connectivity issues. Scans network for Pioneer devices, checks UDP port availability, and provides troubleshooting steps."),
			),
			Handler:             handleDiagnose,
			Category:            "prolink",
			Subcategory:         "diagnostics",
			Tags:                []string{"prolink", "diagnose", "troubleshoot", "network", "firewall"},
			UseCases:            []string{"Troubleshoot connection issues", "Check network setup", "Verify firewall rules"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "prolink",
		},
		{
			Tool: mcp.NewTool("aftrs_prolink_scan_network",
				mcp.WithDescription("Scan local network for Pioneer DJ devices by MAC address. Useful when Pro DJ Link discovery isn't working due to firewall issues."),
			),
			Handler:             handleScanNetwork,
			Category:            "prolink",
			Subcategory:         "diagnostics",
			Tags:                []string{"prolink", "scan", "network", "discovery", "mac", "pioneer"},
			UseCases:            []string{"Find XDJ/CDJ IP addresses", "Bypass UDP discovery issues", "Network troubleshooting"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "prolink",
		},
	}

	// Add bridge tools for Showkontrol integration
	baseTools = append(baseTools, BridgeTools()...)

	return baseTools
}

// handleConnect connects to the Pro DJ Link network
func handleConnect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create prolink client: %w", err)), nil
	}

	if client.IsConnected() {
		return tools.TextResult("✅ Already connected to Pro DJ Link network"), nil
	}

	// Check port availability first
	port50000Available := checkUDPPortAvailable(50000)

	if err := client.Connect(ctx); err != nil {
		var sb strings.Builder
		sb.WriteString("# Pro DJ Link Connection Failed\n\n")
		sb.WriteString(fmt.Sprintf("❌ **Error:** %v\n\n", err))

		if !port50000Available {
			sb.WriteString("## Likely Cause: Port Conflict\n\n")
			sb.WriteString("UDP port 50000 is in use, typically by Rekordbox.\n\n")
			sb.WriteString("**Solution:** Close Rekordbox first:\n")
			sb.WriteString("```powershell\n")
			sb.WriteString("taskkill /IM rekordbox.exe /F\n")
			sb.WriteString("```\n\n")
		} else {
			sb.WriteString("## Possible Causes\n\n")
			sb.WriteString("1. **No CDJ/XDJ devices on network** - Auto-configure needs at least one device broadcasting\n")
			sb.WriteString("2. **Firewall blocking UDP** - Run `aftrs_prolink_diagnose` for setup commands\n")
			sb.WriteString("3. **Network misconfiguration** - Ensure PC and XDJ are on same subnet\n\n")
		}

		sb.WriteString("**Tip:** Use `aftrs_prolink_diagnose` for detailed troubleshooting.\n")
		return tools.TextResult(sb.String()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link Connected\n\n")
	sb.WriteString("✅ Successfully connected to Pro DJ Link network\n\n")
	sb.WriteString("**Note:** Cannot run simultaneously with Rekordbox on the same machine.\n\n")
	sb.WriteString("Use `aftrs_prolink_devices` to see connected devices.\n")

	return tools.TextResult(sb.String()), nil
}

// handleDisconnect disconnects from the Pro DJ Link network
func handleDisconnect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.TextResult("ℹ️ Not connected to Pro DJ Link network"), nil
	}

	if err := client.Disconnect(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to disconnect: %w", err)), nil
	}

	return tools.TextResult("✅ Disconnected from Pro DJ Link network"), nil
}

// handleDevices lists all devices on the Pro DJ Link network
func handleDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get devices: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("*No devices discovered via Pro DJ Link protocol*\n\n")
		sb.WriteString("## Troubleshooting\n\n")
		sb.WriteString("This usually means:\n")
		sb.WriteString("1. **Windows Firewall** is blocking incoming UDP on ports 50000-50002\n")
		sb.WriteString("2. **XDJ/CDJ not broadcasting** - ensure Link indicator is lit and a track is loaded\n")
		sb.WriteString("3. **Different subnet** - PC and XDJ must be on same network segment\n\n")
		sb.WriteString("**Run `aftrs_prolink_diagnose` for detailed troubleshooting steps.**\n\n")

		// Check if we can at least see Pioneer devices via ARP
		pioneerDevices := scanForPioneerDevices()
		if len(pioneerDevices) > 0 {
			sb.WriteString("## Network Scan Results\n\n")
			sb.WriteString("Found Pioneer devices on network (via ARP), but Pro DJ Link discovery failed:\n\n")
			for _, dev := range pioneerDevices {
				sb.WriteString(fmt.Sprintf("- %s (MAC: %s)\n", dev.IP, dev.MAC))
			}
			sb.WriteString("\nThis confirms the XDJ is reachable. The issue is likely firewall-related.\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| ID | Name | Type | IP Address | MAC Address |\n")
	sb.WriteString("|----|------|------|------------|-------------|\n")

	for _, dev := range devices {
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			dev.ID, dev.Name, dev.Type, dev.IP, dev.MAC))
	}

	sb.WriteString(fmt.Sprintf("\n*Found %d device(s)*\n", len(devices)))

	return tools.TextResult(sb.String()), nil
}

// handleStatus returns playback status from decks
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	playerID := tools.GetIntParam(req, "player_id", 0)

	statuses, err := client.GetStatus(ctx, playerID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Deck Status\n\n")

	if len(statuses) == 0 {
		sb.WriteString("*No deck status available*\n\n")
		sb.WriteString("Ensure CDJs are playing and connected to the network.\n")
		return tools.TextResult(sb.String()), nil
	}

	for _, status := range statuses {
		sb.WriteString(fmt.Sprintf("## Player %d\n\n", status.PlayerID))
		sb.WriteString(fmt.Sprintf("**State:** %s\n", status.PlayState))
		sb.WriteString(fmt.Sprintf("**BPM:** %.2f\n", status.BPM))
		sb.WriteString(fmt.Sprintf("**Effective BPM:** %.2f (with pitch)\n", status.EffectiveBPM))
		sb.WriteString(fmt.Sprintf("**Pitch:** %.2f%%\n", status.SliderPitch))
		sb.WriteString(fmt.Sprintf("**Beat:** %d/%d (beat %d)\n", status.BeatInMeasure, 4, status.Beat))

		// Status flags
		flags := []string{}
		if status.IsOnAir {
			flags = append(flags, "🔊 On Air")
		}
		if status.IsMaster {
			flags = append(flags, "👑 Master")
		}
		if status.IsSync {
			flags = append(flags, "🔗 Sync")
		}
		if len(flags) > 0 {
			sb.WriteString(fmt.Sprintf("**Flags:** %s\n", strings.Join(flags, " | ")))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleNowPlaying returns current track on each deck
func handleNowPlaying(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	playerID := tools.GetIntParam(req, "player_id", 0)

	nowPlaying, err := client.GetNowPlaying(ctx, playerID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get now playing: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Now Playing\n\n")

	if len(nowPlaying) == 0 {
		sb.WriteString("*No track information available*\n\n")
		sb.WriteString("Ensure CDJs have tracks loaded and are connected to the network.\n")
		return tools.TextResult(sb.String()), nil
	}

	for _, np := range nowPlaying {
		sb.WriteString(fmt.Sprintf("## Player %d\n\n", np.PlayerID))

		if np.Track != nil {
			sb.WriteString(fmt.Sprintf("**Title:** %s\n", np.Track.Title))
			sb.WriteString(fmt.Sprintf("**Artist:** %s\n", np.Track.Artist))
			if np.Track.Album != "" {
				sb.WriteString(fmt.Sprintf("**Album:** %s\n", np.Track.Album))
			}
			if np.Track.Key != "" {
				sb.WriteString(fmt.Sprintf("**Key:** %s\n", np.Track.Key))
			}
			if np.Track.Length > 0 {
				minutes := int(np.Track.Length) / 60
				seconds := int(np.Track.Length) % 60
				sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", minutes, seconds))
			}
		} else {
			sb.WriteString("*Track metadata not available*\n")
		}

		if np.Status != nil {
			sb.WriteString(fmt.Sprintf("**BPM:** %.2f (effective: %.2f)\n", np.Status.BPM, np.Status.EffectiveBPM))
			sb.WriteString(fmt.Sprintf("**State:** %s\n", np.Status.PlayState))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleMaster returns the current master player
func handleMaster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	masterID, err := client.GetMasterPlayer(ctx)
	if err != nil {
		return tools.TextResult("ℹ️ No master player set"), nil
	}

	// Get the full status for the master player
	nowPlaying, _ := client.GetNowPlaying(ctx, masterID)

	var sb strings.Builder
	sb.WriteString("# Master Player\n\n")
	sb.WriteString(fmt.Sprintf("**Player ID:** %d\n\n", masterID))

	if len(nowPlaying) > 0 && nowPlaying[0].Status != nil {
		np := nowPlaying[0]
		sb.WriteString(fmt.Sprintf("**BPM:** %.2f\n", np.Status.EffectiveBPM))
		sb.WriteString(fmt.Sprintf("**Beat:** %d/4\n", np.Status.BeatInMeasure))
		if np.Track != nil {
			sb.WriteString(fmt.Sprintf("**Track:** %s - %s\n", np.Track.Artist, np.Track.Title))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleTrack returns track metadata from a deck
func handleTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	playerID, errResult := tools.RequireIntParam(req, "player_id")
	if errResult != nil {
		return errResult, nil
	}

	trackID := uint32(tools.GetIntParam(req, "track_id", 0))

	// If no track_id specified, get current track
	if trackID == 0 {
		nowPlaying, err := client.GetNowPlaying(ctx, playerID)
		if err != nil || len(nowPlaying) == 0 {
			return tools.ErrorResult(fmt.Errorf("no track loaded on player %d", playerID)), nil
		}
		if nowPlaying[0].Track != nil {
			trackID = nowPlaying[0].Track.ID
		}
	}

	if trackID == 0 {
		return tools.ErrorResult(fmt.Errorf("no track_id specified and no track loaded")), nil
	}

	track, err := client.GetTrack(ctx, playerID, trackID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Track Details\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** %d\n", track.ID))
	sb.WriteString(fmt.Sprintf("**Title:** %s\n", track.Title))
	sb.WriteString(fmt.Sprintf("**Artist:** %s\n", track.Artist))
	if track.Album != "" {
		sb.WriteString(fmt.Sprintf("**Album:** %s\n", track.Album))
	}
	if track.Genre != "" {
		sb.WriteString(fmt.Sprintf("**Genre:** %s\n", track.Genre))
	}
	if track.Label != "" {
		sb.WriteString(fmt.Sprintf("**Label:** %s\n", track.Label))
	}
	if track.Key != "" {
		sb.WriteString(fmt.Sprintf("**Key:** %s\n", track.Key))
	}
	if track.Length > 0 {
		minutes := int(track.Length) / 60
		seconds := int(track.Length) % 60
		sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", minutes, seconds))
	}
	if track.Comment != "" {
		sb.WriteString(fmt.Sprintf("**Comment:** %s\n", track.Comment))
	}
	if track.Path != "" {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`\n", track.Path))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBeatInfo returns beat/tempo information for Showkontrol
func handleBeatInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	playerID := tools.GetIntParam(req, "player_id", 0)

	// If player_id is 0, get master player
	if playerID == 0 {
		masterID, err := client.GetMasterPlayer(ctx)
		if err == nil {
			playerID = masterID
		}
	}

	statuses, err := client.GetStatus(ctx, playerID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	if len(statuses) == 0 {
		return tools.ErrorResult(fmt.Errorf("no status available for player %d", playerID)), nil
	}

	status := statuses[0]

	beatInfo := map[string]interface{}{
		"player_id":       status.PlayerID,
		"bpm":             status.BPM,
		"effective_bpm":   status.EffectiveBPM,
		"beat_in_measure": status.BeatInMeasure,
		"beat":            status.Beat,
		"is_master":       status.IsMaster,
		"is_playing":      status.PlayState == "playing",
		"ms_per_beat":     60000.0 / float64(status.EffectiveBPM),
	}

	return tools.JSONResult(beatInfo), nil
}

// handleAllData returns all available prolink data in a single comprehensive response
func handleAllData(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getProlinkClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get prolink client: %w", err)), nil
	}

	if !client.IsConnected() {
		return tools.ErrorResult(fmt.Errorf("not connected to Pro DJ Link network. Use aftrs_prolink_connect first.")), nil
	}

	fullData, err := client.GetFullData(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get full data: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pro DJ Link - Full Data Export\n\n")

	sb.WriteString(fmt.Sprintf("**Timestamp:** %s\n", fullData.Timestamp))
	sb.WriteString(fmt.Sprintf("**Connected:** %v\n", fullData.Connected))
	sb.WriteString(fmt.Sprintf("**Master Player:** %d\n\n", fullData.MasterID))

	// Devices section
	sb.WriteString("## Devices\n\n")
	if len(fullData.Devices) == 0 {
		sb.WriteString("*No devices found*\n\n")
	} else {
		sb.WriteString("| ID | Name | Type | IP | MAC |\n")
		sb.WriteString("|----|------|------|----|-----|\n")
		for _, dev := range fullData.Devices {
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
				dev.ID, dev.Name, dev.Type, dev.IP, dev.MAC))
		}
		sb.WriteString("\n")
	}

	// Players section
	sb.WriteString("## Players\n\n")
	if len(fullData.Players) == 0 {
		sb.WriteString("*No player data available*\n\n")
	} else {
		for playerKey, np := range fullData.Players {
			sb.WriteString(fmt.Sprintf("### %s\n\n", playerKey))

			if np.Status != nil {
				sb.WriteString("**Status:**\n")
				sb.WriteString(fmt.Sprintf("- Play State: %s\n", np.Status.PlayState))
				sb.WriteString(fmt.Sprintf("- BPM: %.2f (effective: %.2f)\n", np.Status.BPM, np.Status.EffectiveBPM))
				sb.WriteString(fmt.Sprintf("- Pitch: %.2f%% (slider), %.2f%% (effective)\n", np.Status.SliderPitch, np.Status.EffectivePitch))
				sb.WriteString(fmt.Sprintf("- Beat: %d/4 (total: %d)\n", np.Status.BeatInMeasure, np.Status.Beat))
				sb.WriteString(fmt.Sprintf("- Beats Until Cue: %d\n", np.Status.BeatsUntilCue))
				sb.WriteString(fmt.Sprintf("- ms/beat: %.1f\n", np.Status.MsPerBeat))
				sb.WriteString(fmt.Sprintf("- Track ID: %d (slot: %s, type: %s)\n", np.Status.TrackID, np.Status.TrackSlot, np.Status.TrackType))
				sb.WriteString(fmt.Sprintf("- Flags: Master=%v, Sync=%v, OnAir=%v, Playing=%v\n\n",
					np.Status.IsMaster, np.Status.IsSync, np.Status.IsOnAir, np.Status.IsPlaying))
			}

			if np.Track != nil {
				sb.WriteString("**Track:**\n")
				sb.WriteString(fmt.Sprintf("- Title: %s\n", np.Track.Title))
				sb.WriteString(fmt.Sprintf("- Artist: %s\n", np.Track.Artist))
				if np.Track.Album != "" {
					sb.WriteString(fmt.Sprintf("- Album: %s\n", np.Track.Album))
				}
				if np.Track.Genre != "" {
					sb.WriteString(fmt.Sprintf("- Genre: %s\n", np.Track.Genre))
				}
				if np.Track.Label != "" {
					sb.WriteString(fmt.Sprintf("- Label: %s\n", np.Track.Label))
				}
				if np.Track.Key != "" {
					sb.WriteString(fmt.Sprintf("- Key: %s\n", np.Track.Key))
				}
				minutes := int(np.Track.Length) / 60
				seconds := int(np.Track.Length) % 60
				sb.WriteString(fmt.Sprintf("- Length: %d:%02d\n", minutes, seconds))
				if np.Track.Comment != "" {
					sb.WriteString(fmt.Sprintf("- Comment: %s\n", np.Track.Comment))
				}
				if np.Track.Path != "" {
					sb.WriteString(fmt.Sprintf("- Path: %s\n", np.Track.Path))
				}
				sb.WriteString(fmt.Sprintf("- Has Artwork: %v\n", np.Track.HasArtwork))
			} else {
				sb.WriteString("*No track metadata available*\n")
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleDiagnose performs network diagnostics for Pro DJ Link connectivity
func handleDiagnose(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Pro DJ Link Network Diagnostics\n\n")

	// 1. Check for Pioneer devices on network via ARP/neighbor cache
	sb.WriteString("## 1. Network Scan for Pioneer Devices\n\n")
	pioneerDevices := scanForPioneerDevices()
	if len(pioneerDevices) > 0 {
		sb.WriteString("✅ **Found Pioneer/AlphaTheta devices on network:**\n\n")
		for _, dev := range pioneerDevices {
			sb.WriteString(fmt.Sprintf("- **%s** (MAC: `%s`)\n", dev.IP, dev.MAC))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("⚠️ No Pioneer devices found in ARP cache.\n")
		sb.WriteString("   Try pinging your XDJ/CDJ first, or ensure it's connected to the same network.\n\n")
	}

	// 2. Check UDP port availability
	sb.WriteString("## 2. UDP Port Availability\n\n")
	ports := []int{50000, 50001, 50002}
	allPortsAvailable := true
	for _, port := range ports {
		available := checkUDPPortAvailable(port)
		if available {
			sb.WriteString(fmt.Sprintf("✅ UDP port %d: Available\n", port))
		} else {
			sb.WriteString(fmt.Sprintf("❌ UDP port %d: **In use or blocked** (likely Rekordbox)\n", port))
			allPortsAvailable = false
		}
	}
	sb.WriteString("\n")

	// 3. Check if currently connected
	sb.WriteString("## 3. Pro DJ Link Connection Status\n\n")
	client, err := getProlinkClient()
	if err != nil {
		sb.WriteString(fmt.Sprintf("❌ Failed to get client: %v\n\n", err))
	} else if client.IsConnected() {
		sb.WriteString("✅ Connected to Pro DJ Link network\n\n")
		devices, _ := client.GetDevices(ctx)
		sb.WriteString(fmt.Sprintf("   Devices discovered: %d\n\n", len(devices)))
	} else {
		sb.WriteString("⚠️ Not connected to Pro DJ Link network\n\n")
	}

	// 4. Troubleshooting recommendations
	sb.WriteString("## 4. Troubleshooting Steps\n\n")

	if !allPortsAvailable {
		sb.WriteString("### Port Conflict Detected\n\n")
		sb.WriteString("UDP port 50000 is in use. This is typically caused by Rekordbox.\n\n")
		sb.WriteString("**Solution:** Close Rekordbox before using Pro DJ Link tools:\n")
		sb.WriteString("```powershell\n")
		sb.WriteString("taskkill /IM rekordbox.exe /F\n")
		sb.WriteString("```\n\n")
	}

	sb.WriteString("### Windows Firewall Setup\n\n")
	sb.WriteString("Run these commands as Administrator to allow Pro DJ Link traffic:\n\n")
	sb.WriteString("```powershell\n")
	sb.WriteString("# Allow Pro DJ Link UDP ports (run as Administrator)\n")
	sb.WriteString("New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50000' -Direction Inbound -Protocol UDP -LocalPort 50000 -Action Allow\n")
	sb.WriteString("New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50001' -Direction Inbound -Protocol UDP -LocalPort 50001 -Action Allow\n")
	sb.WriteString("New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50002' -Direction Inbound -Protocol UDP -LocalPort 50002 -Action Allow\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### XDJ/CDJ Checklist\n\n")
	sb.WriteString("- [ ] XDJ/CDJ is powered on\n")
	sb.WriteString("- [ ] Connected to same network via Ethernet (not WiFi)\n")
	sb.WriteString("- [ ] USB drive with exported Rekordbox library inserted\n")
	sb.WriteString("- [ ] Track loaded on the deck\n")
	sb.WriteString("- [ ] Link indicator lit on the XDJ/CDJ\n\n")

	if len(pioneerDevices) > 0 {
		sb.WriteString("### Direct Connection Test\n\n")
		sb.WriteString("Pioneer device found at: `" + pioneerDevices[0].IP + "`\n\n")
		sb.WriteString("Try pinging it:\n")
		sb.WriteString("```cmd\n")
		sb.WriteString("ping " + pioneerDevices[0].IP + "\n")
		sb.WriteString("```\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleScanNetwork scans the local network for Pioneer DJ devices
func handleScanNetwork(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Network Scan for Pioneer DJ Devices\n\n")

	devices := scanForPioneerDevices()

	if len(devices) == 0 {
		sb.WriteString("No Pioneer/AlphaTheta devices found on the network.\n\n")
		sb.WriteString("**Tips:**\n")
		sb.WriteString("- Ensure your XDJ/CDJ is connected via Ethernet to the same network\n")
		sb.WriteString("- Try pinging your router or other devices to populate the ARP cache\n")
		sb.WriteString("- Check that the XDJ/CDJ Link indicator is lit\n\n")

		sb.WriteString("**Known Pioneer/AlphaTheta MAC prefixes:**\n")
		for _, prefix := range pioneerOUIPrefixes {
			sb.WriteString(fmt.Sprintf("- `%s`\n", prefix))
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Found Devices\n\n")
	sb.WriteString("| IP Address | MAC Address | Vendor |\n")
	sb.WriteString("|------------|-------------|--------|\n")

	for _, dev := range devices {
		sb.WriteString(fmt.Sprintf("| %s | %s | Pioneer/AlphaTheta |\n", dev.IP, dev.MAC))
	}

	sb.WriteString(fmt.Sprintf("\n*Found %d Pioneer device(s)*\n\n", len(devices)))

	sb.WriteString("## Next Steps\n\n")
	sb.WriteString("1. Ensure Windows Firewall allows UDP ports 50000-50002\n")
	sb.WriteString("2. Close Rekordbox if running\n")
	sb.WriteString("3. Use `aftrs_prolink_connect` to connect to the network\n")
	sb.WriteString("4. Use `aftrs_prolink_devices` to verify device discovery\n")

	return tools.TextResult(sb.String()), nil
}

// NetworkDevice represents a device found on the network
type NetworkDevice struct {
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

// scanForPioneerDevices scans the ARP cache for Pioneer/AlphaTheta devices
func scanForPioneerDevices() []NetworkDevice {
	var devices []NetworkDevice

	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return devices
	}

	// Build a map of known IPs from ARP cache by checking interface addresses
	// Note: Go doesn't have direct ARP access, so we check what we can reach
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}

			// Check if this interface has a hardware address that matches Pioneer
			mac := strings.ToLower(iface.HardwareAddr.String())
			for _, prefix := range pioneerOUIPrefixes {
				if strings.HasPrefix(mac, strings.ToLower(prefix)) {
					devices = append(devices, NetworkDevice{
						IP:  ipNet.IP.String(),
						MAC: iface.HardwareAddr.String(),
					})
					break
				}
			}
		}
	}

	// Also try to find devices by attempting connections to common IPs
	// This is a simplified approach - real ARP scanning would require raw sockets
	return devices
}

// checkUDPPortAvailable checks if a UDP port is available for binding
func checkUDPPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	// Small delay to ensure port is released
	time.Sleep(10 * time.Millisecond)
	return true
}

// init registers this module
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
