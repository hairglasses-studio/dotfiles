// Package retrogaming provides PS2 and emulator tools for hg-mcp.
package retrogaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewRetroGamingClient)

// Module implements the ToolModule interface for retro gaming
type Module struct{}

func (m *Module) Name() string {
	return "retrogaming"
}

func (m *Module) Description() string {
	return "PS2 emulation and retro gaming support"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ps2_status",
				mcp.WithDescription("Get PCSX2 emulator status including running state and current game."),
			),
			Handler:             handlePS2Status,
			Category:            "retrogaming",
			Subcategory:         "ps2",
			Tags:                []string{"ps2", "pcsx2", "emulator", "status"},
			UseCases:            []string{"Check if emulator is running", "Get current game info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "retrogaming",
		},
		{
			Tool: mcp.NewTool("aftrs_ps2_games",
				mcp.WithDescription("List available PS2 games in the configured games directory."),
				mcp.WithString("search",
					mcp.Description("Filter games by name"),
				),
			),
			Handler:             handlePS2Games,
			Category:            "retrogaming",
			Subcategory:         "ps2",
			Tags:                []string{"ps2", "games", "library", "list"},
			UseCases:            []string{"Browse game library", "Find specific games"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "retrogaming",
		},
		{
			Tool: mcp.NewTool("aftrs_ps2_savestate",
				mcp.WithDescription("List save states for PS2 games."),
				mcp.WithString("game",
					mcp.Description("Filter by game name (optional)"),
				),
			),
			Handler:             handlePS2SaveState,
			Category:            "retrogaming",
			Subcategory:         "ps2",
			Tags:                []string{"ps2", "savestate", "save", "load"},
			UseCases:            []string{"Find save states", "Manage game saves"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "retrogaming",
		},
		{
			Tool: mcp.NewTool("aftrs_capture_status",
				mcp.WithDescription("Get video capture device status for game capture."),
			),
			Handler:             handleCaptureStatus,
			Category:            "retrogaming",
			Subcategory:         "capture",
			Tags:                []string{"capture", "video", "recording", "streaming"},
			UseCases:            []string{"Check capture card status", "Verify recording setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "retrogaming",
		},
		{
			Tool: mcp.NewTool("aftrs_retro_visualizer",
				mcp.WithDescription("Control audio visualizer for retro gaming streams (placeholder)."),
				mcp.WithString("action",
					mcp.Description("Action: 'start', 'stop', 'status'"),
				),
			),
			Handler:             handleRetroVisualizer,
			Category:            "retrogaming",
			Subcategory:         "visualizer",
			Tags:                []string{"visualizer", "audio", "effects", "streaming"},
			UseCases:            []string{"Control stream visuals", "Audio reactive effects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "retrogaming",
			IsWrite:             true,
		},
	}
}

// handlePS2Status handles the aftrs_ps2_status tool
func handlePS2Status(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetPS2Status(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# PCSX2 Emulator Status\n\n")

	if status.Running {
		sb.WriteString("**Status:** 🟢 Running\n\n")
		if status.GameLoaded != "" {
			sb.WriteString(fmt.Sprintf("**Current Game:** %s\n", status.GameLoaded))
		}
		if status.FPS > 0 {
			sb.WriteString(fmt.Sprintf("**FPS:** %.1f\n", status.FPS))
		}
		if status.Speed != "" {
			sb.WriteString(fmt.Sprintf("**Speed:** %s\n", status.Speed))
		}
		if status.Resolution != "" {
			sb.WriteString(fmt.Sprintf("**Resolution:** %s\n", status.Resolution))
		}
	} else {
		sb.WriteString("**Status:** ⚫ Not Running\n\n")
		if client.PCSX2Path() != "" {
			sb.WriteString(fmt.Sprintf("**PCSX2 Path:** %s\n", client.PCSX2Path()))
		} else {
			sb.WriteString("**PCSX2 Path:** Not configured\n\n")
			sb.WriteString("## Setup\n\n")
			sb.WriteString("Set environment variable:\n")
			sb.WriteString("```bash\n")
			sb.WriteString("export PCSX2_PATH=/path/to/pcsx2\n")
			sb.WriteString("```\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handlePS2Games handles the aftrs_ps2_games tool
func handlePS2Games(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	search := tools.GetStringParam(req, "search")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	games, err := client.ListPS2Games(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Filter by search term
	if search != "" {
		searchLower := strings.ToLower(search)
		filtered := []clients.PS2Game{}
		for _, game := range games {
			if strings.Contains(strings.ToLower(game.Name), searchLower) {
				filtered = append(filtered, game)
			}
		}
		games = filtered
	}

	var sb strings.Builder
	sb.WriteString("# PS2 Game Library\n\n")
	sb.WriteString(fmt.Sprintf("**Games Path:** %s\n\n", client.GamesPath()))

	if len(games) == 0 {
		sb.WriteString("No games found.\n\n")
		if search != "" {
			sb.WriteString(fmt.Sprintf("No games matching '%s'.\n", search))
		} else {
			sb.WriteString("## Setup\n\n")
			sb.WriteString("Set environment variable:\n")
			sb.WriteString("```bash\n")
			sb.WriteString("export PS2_GAMES_PATH=/path/to/ps2/games\n")
			sb.WriteString("```\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** games:\n\n", len(games)))
	sb.WriteString("| Game | Format | Size |\n")
	sb.WriteString("|------|--------|------|\n")

	for _, game := range games {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", game.Name, game.Format, game.Size))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePS2SaveState handles the aftrs_ps2_savestate tool
func handlePS2SaveState(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gameFilter := tools.GetStringParam(req, "game")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	states, err := client.ListSaveStates(ctx, gameFilter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Filter by game if specified
	if gameFilter != "" {
		filterLower := strings.ToLower(gameFilter)
		filtered := []clients.SaveState{}
		for _, state := range states {
			if strings.Contains(strings.ToLower(state.Game), filterLower) {
				filtered = append(filtered, state)
			}
		}
		states = filtered
	}

	var sb strings.Builder
	sb.WriteString("# PS2 Save States\n\n")

	if len(states) == 0 {
		sb.WriteString("No save states found.\n")
		if gameFilter != "" {
			sb.WriteString(fmt.Sprintf("\nNo states for game matching '%s'.\n", gameFilter))
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** save states:\n\n", len(states)))
	sb.WriteString("| Game | Slot | Date |\n")
	sb.WriteString("|------|------|------|\n")

	for _, state := range states {
		date := state.Timestamp.Format("2006-01-02 15:04")
		sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", state.Game, state.Slot, date))
	}

	return tools.TextResult(sb.String()), nil
}

// handleCaptureStatus handles the aftrs_capture_status tool
func handleCaptureStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	devices, err := client.GetCaptureDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video Capture Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No capture devices detected.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Count connected devices
	connected := 0
	for _, d := range devices {
		if d.Connected {
			connected++
		}
	}

	sb.WriteString(fmt.Sprintf("Found **%d** devices (%d connected):\n\n", len(devices), connected))
	sb.WriteString("| Device | Path | Status |\n")
	sb.WriteString("|--------|------|--------|\n")

	for _, device := range devices {
		status := "🔴 Disconnected"
		if device.Connected {
			status = "🟢 Connected"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", device.Name, device.Path, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleRetroVisualizer handles the aftrs_retro_visualizer tool
func handleRetroVisualizer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "status")

	var sb strings.Builder
	sb.WriteString("# Retro Visualizer\n\n")
	sb.WriteString("**Status:** ⚠️ Not Implemented\n\n")
	sb.WriteString("Audio visualizer control requires integration with:\n")
	sb.WriteString("- TouchDesigner audio reactive components\n")
	sb.WriteString("- OBS scene switching\n")
	sb.WriteString("- Custom visualizer application\n\n")
	sb.WriteString(fmt.Sprintf("**Requested Action:** %s\n", action))

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
