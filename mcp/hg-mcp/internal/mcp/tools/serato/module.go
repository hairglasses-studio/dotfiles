// Package serato provides MCP tools for Serato DJ Pro integration.
package serato

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the Serato DJ tools module
type Module struct{}

var getSeratoClient = tools.LazyClient(clients.NewSeratoClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "serato"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Serato DJ Pro library integration for track info, crates, and history"
}

// Tools returns all Serato tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_serato_status",
				mcp.WithDescription("Get Serato DJ library status including connection state, crate count, and history sessions"),
			),
			Handler:             handleSeratoStatus,
			Category:            "serato",
			Subcategory:         "status",
			Tags:                []string{"serato", "dj", "status", "library"},
			UseCases:            []string{"Check if Serato library is accessible", "Get library overview"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_now_playing",
				mcp.WithDescription("Get the currently playing track from the most recent Serato session"),
			),
			Handler:             handleSeratoNowPlaying,
			Category:            "serato",
			Subcategory:         "playback",
			Tags:                []string{"serato", "dj", "now-playing", "track"},
			UseCases:            []string{"Get current track info", "Display now playing", "Sync visuals to music"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_history",
				mcp.WithDescription("Get recent play history sessions from Serato"),
				mcp.WithNumber("limit",
					mcp.Description("Maximum number of sessions to return (default: 10)"),
				),
			),
			Handler:             handleSeratoHistory,
			Category:            "serato",
			Subcategory:         "history",
			Tags:                []string{"serato", "dj", "history", "sessions"},
			UseCases:            []string{"Review past DJ sets", "Export setlists", "Track play statistics"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_session",
				mcp.WithDescription("Get detailed track list from a specific history session"),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Session name (filename without .session extension)"),
				),
			),
			Handler:             handleSeratoSession,
			Category:            "serato",
			Subcategory:         "history",
			Tags:                []string{"serato", "dj", "history", "session", "tracks"},
			UseCases:            []string{"Get full setlist from a session", "Export track list"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_crates",
				mcp.WithDescription("List all crates in the Serato library"),
			),
			Handler:             handleSeratoCrates,
			Category:            "serato",
			Subcategory:         "library",
			Tags:                []string{"serato", "dj", "crates", "library"},
			UseCases:            []string{"Browse music collection", "List available crates"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_crate",
				mcp.WithDescription("Get tracks from a specific crate"),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Crate name (use / for nested crates)"),
				),
			),
			Handler:             handleSeratoCrate,
			Category:            "serato",
			Subcategory:         "library",
			Tags:                []string{"serato", "dj", "crate", "tracks"},
			UseCases:            []string{"Browse crate contents", "Get track list from crate"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_search",
				mcp.WithDescription("Search for tracks in the Serato library"),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query (matches filename)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum results (default: 50)"),
				),
			),
			Handler:             handleSeratoSearch,
			Category:            "serato",
			Subcategory:         "library",
			Tags:                []string{"serato", "dj", "search", "tracks"},
			UseCases:            []string{"Find tracks by name", "Search music library"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_track",
				mcp.WithDescription("Get detailed information about a specific track"),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Full path to the track file"),
				),
			),
			Handler:             handleSeratoTrack,
			Category:            "serato",
			Subcategory:         "library",
			Tags:                []string{"serato", "dj", "track", "metadata"},
			UseCases:            []string{"Get track details", "View BPM and key information"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_library_path",
				mcp.WithDescription("Get the configured Serato library path"),
			),
			Handler:             handleSeratoLibraryPath,
			Category:            "serato",
			Subcategory:         "config",
			Tags:                []string{"serato", "dj", "config", "library"},
			UseCases:            []string{"Check library location", "Verify configuration"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_stats",
				mcp.WithDescription("Get library statistics including track counts and crate info"),
			),
			Handler:             handleSeratoStats,
			Category:            "serato",
			Subcategory:         "library",
			Tags:                []string{"serato", "dj", "stats", "library"},
			UseCases:            []string{"Get library overview", "Track collection statistics"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_recent_tracks",
				mcp.WithDescription("Get recently played tracks across all sessions"),
				mcp.WithNumber("limit",
					mcp.Description("Maximum tracks to return (default: 20)"),
				),
			),
			Handler:             handleSeratoRecentTracks,
			Category:            "serato",
			Subcategory:         "history",
			Tags:                []string{"serato", "dj", "recent", "tracks"},
			UseCases:            []string{"Get recently played tracks", "Quick history overview"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
		{
			Tool: mcp.NewTool("aftrs_serato_health",
				mcp.WithDescription("Check Serato library health and get recommendations"),
			),
			Handler:             handleSeratoHealth,
			Category:            "serato",
			Subcategory:         "status",
			Tags:                []string{"serato", "dj", "health", "diagnostics"},
			UseCases:            []string{"Diagnose library issues", "Check configuration"},
			Complexity:          "simple",
			CircuitBreakerGroup: "serato",
		},
	}
}

func handleSeratoStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleSeratoNowPlaying(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	track, err := client.GetNowPlaying(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get now playing: %w", err)), nil
	}

	return tools.JSONResult(track), nil
}

func handleSeratoHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	sessions, err := client.GetHistory(ctx, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get history: %w", err)), nil
	}

	return tools.JSONResult(sessions), nil
}

func handleSeratoSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	session, err := client.GetHistorySession(ctx, name)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get session: %w", err)), nil
	}

	return tools.JSONResult(session), nil
}

func handleSeratoCrates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	crates, err := client.GetCrates(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get crates: %w", err)), nil
	}

	return tools.JSONResult(crates), nil
}

func handleSeratoCrate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	crate, err := client.GetCrate(ctx, name)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get crate: %w", err)), nil
	}

	return tools.JSONResult(crate), nil
}

func handleSeratoSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	tracks, err := client.SearchLibrary(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to search: %w", err)), nil
	}

	result := map[string]any{
		"query":   query,
		"count":   len(tracks),
		"results": tracks,
	}

	return tools.JSONResult(result), nil
}

func handleSeratoTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	// Return basic track info from path
	track := clients.SeratoTrack{
		Path:     path,
		Filename: path[len(path)-minInt(len(path), 50):],
	}

	return tools.JSONResult(track), nil
}

func handleSeratoLibraryPath(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	result := map[string]string{
		"library_path": client.LibraryPath(),
	}

	return tools.JSONResult(result), nil
}

func handleSeratoStats(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	crates, _ := client.GetCrates(ctx)

	totalTracks := 0
	for _, c := range crates {
		totalTracks += c.TrackCount
	}

	stats := map[string]any{
		"library_path":  status.LibraryPath,
		"connected":     status.Connected,
		"has_database":  status.HasDatabase,
		"crate_count":   status.CrateCount,
		"history_count": status.HistoryCount,
		"total_tracks":  totalTracks,
		"unique_crates": len(crates),
	}

	return tools.JSONResult(stats), nil
}

func handleSeratoRecentTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	// Get recent sessions
	sessions, err := client.GetHistory(ctx, 5)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get history: %w", err)), nil
	}

	var recentTracks []clients.SeratoTrack
	for _, session := range sessions {
		fullSession, err := client.GetHistorySession(ctx, session.Name)
		if err != nil {
			continue
		}
		for _, track := range fullSession.Tracks {
			recentTracks = append(recentTracks, track)
			if len(recentTracks) >= limit {
				break
			}
		}
		if len(recentTracks) >= limit {
			break
		}
	}

	result := map[string]any{
		"count":  len(recentTracks),
		"tracks": recentTracks,
	}

	return tools.JSONResult(result), nil
}

func handleSeratoHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSeratoClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Serato client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
