// Package showkontrol provides MCP tools for Showkontrol timecode and cue management.
package showkontrol

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Showkontrol tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "showkontrol"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Showkontrol timecode and cue management for synchronized live performances"
}

// Tools returns the Showkontrol tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_showkontrol_status",
				mcp.WithDescription("Get Showkontrol system status including timecode and current show"),
			),
			Handler:             handleShowkontrolStatus,
			Category:            "showkontrol",
			Subcategory:         "status",
			Tags:                []string{"showkontrol", "timecode", "status", "cue"},
			UseCases:            []string{"Check system status", "View timecode", "Get current show"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_shows",
				mcp.WithDescription("List available shows in Showkontrol"),
			),
			Handler:             handleShowkontrolShows,
			Category:            "showkontrol",
			Subcategory:         "shows",
			Tags:                []string{"showkontrol", "shows", "list"},
			UseCases:            []string{"List shows", "Find show IDs", "Browse available shows"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_show",
				mcp.WithDescription("Get or load a specific show"),
				mcp.WithString("show_id", mcp.Required(), mcp.Description("Show ID to get or load")),
				mcp.WithBoolean("load", mcp.Description("If true, load the show as current")),
			),
			Handler:             handleShowkontrolShow,
			Category:            "showkontrol",
			Subcategory:         "shows",
			Tags:                []string{"showkontrol", "show", "load"},
			UseCases:            []string{"Load show", "Get show details", "Switch shows"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_cues",
				mcp.WithDescription("List cues in the current show"),
			),
			Handler:             handleShowkontrolCues,
			Category:            "showkontrol",
			Subcategory:         "cues",
			Tags:                []string{"showkontrol", "cues", "list"},
			UseCases:            []string{"List cues", "View cue list", "Check cue timings"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_cue_go",
				mcp.WithDescription("Fire a specific cue by ID"),
				mcp.WithString("cue_id", mcp.Required(), mcp.Description("Cue ID to fire")),
			),
			Handler:             handleShowkontrolCueGo,
			Category:            "showkontrol",
			Subcategory:         "cues",
			Tags:                []string{"showkontrol", "cue", "fire", "go"},
			UseCases:            []string{"Fire cue", "Trigger specific cue", "Manual cue execution"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_go",
				mcp.WithDescription("Fire the next cue in sequence"),
			),
			Handler:             handleShowkontrolGo,
			Category:            "showkontrol",
			Subcategory:         "cues",
			Tags:                []string{"showkontrol", "go", "next", "cue"},
			UseCases:            []string{"Fire next cue", "Advance show", "Continue sequence"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_timecode",
				mcp.WithDescription("Get or set timecode position"),
				mcp.WithString("position", mcp.Description("Position to jump to (HH:MM:SS:FF or seconds). Omit to get current position.")),
			),
			Handler:             handleShowkontrolTimecode,
			Category:            "showkontrol",
			Subcategory:         "timecode",
			Tags:                []string{"showkontrol", "timecode", "position"},
			UseCases:            []string{"Get timecode", "Jump to position", "Set playhead"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_timecode_start",
				mcp.WithDescription("Start timecode playback"),
			),
			Handler:             handleShowkontrolTimecodeStart,
			Category:            "showkontrol",
			Subcategory:         "timecode",
			Tags:                []string{"showkontrol", "timecode", "start", "play"},
			UseCases:            []string{"Start timecode", "Begin playback", "Run show"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_timecode_stop",
				mcp.WithDescription("Stop timecode playback"),
			),
			Handler:             handleShowkontrolTimecodeStop,
			Category:            "showkontrol",
			Subcategory:         "timecode",
			Tags:                []string{"showkontrol", "timecode", "stop"},
			UseCases:            []string{"Stop timecode", "Pause show", "Halt playback"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_showkontrol_health",
				mcp.WithDescription("Check Showkontrol connection health and get troubleshooting recommendations"),
			),
			Handler:             handleShowkontrolHealth,
			Category:            "showkontrol",
			Subcategory:         "status",
			Tags:                []string{"showkontrol", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Check connection", "Diagnose issues", "Verify setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "showkontrol",
		},
	}
}

var getShowkontrolClient = tools.LazyClient(clients.NewShowkontrolClient)

func handleShowkontrolStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleShowkontrolShows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	shows, err := client.GetShows(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get shows: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"shows": shows,
		"count": len(shows),
	}), nil
}

func handleShowkontrolShow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	showID, errResult := tools.RequireStringParam(req, "show_id")
	if errResult != nil {
		return errResult, nil
	}

	load := tools.GetBoolParam(req, "load", false)

	if load {
		if err := client.LoadShow(ctx, showID); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to load show: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"success": true,
			"show_id": showID,
			"message": fmt.Sprintf("Show '%s' loaded", showID),
		}), nil
	}

	show, err := client.GetShow(ctx, showID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get show: %w", err)), nil
	}

	return tools.JSONResult(show), nil
}

func handleShowkontrolCues(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	cues, err := client.GetCues(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get cues: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"cues":  cues,
		"count": len(cues),
	}), nil
}

func handleShowkontrolCueGo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	cueID, errResult := tools.RequireStringParam(req, "cue_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.FireCue(ctx, cueID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fire cue: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"cue_id":  cueID,
		"message": fmt.Sprintf("Cue '%s' fired", cueID),
	}), nil
}

func handleShowkontrolGo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	if err := client.Go(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to fire next cue: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Next cue fired",
	}), nil
}

func handleShowkontrolTimecode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	position := tools.GetStringParam(req, "position")

	if position != "" {
		if err := client.GotoTimecode(ctx, position); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to set timecode: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"success":  true,
			"position": position,
			"message":  fmt.Sprintf("Timecode set to %s", position),
		}), nil
	}

	tcStatus, err := client.GetTimecodeStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get timecode: %w", err)), nil
	}

	return tools.JSONResult(tcStatus), nil
}

func handleShowkontrolTimecodeStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	if err := client.StartTimecode(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start timecode: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Timecode playback started",
	}), nil
}

func handleShowkontrolTimecodeStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	if err := client.StopTimecode(ctx); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to stop timecode: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Timecode playback stopped",
	}), nil
}

func handleShowkontrolHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getShowkontrolClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Showkontrol client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
