// Package traktor provides MCP tools for Traktor Pro 3 library access.
package traktor

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Traktor tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "traktor"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Traktor Pro 3 library access for tracks, playlists, cue points, and export"
}

// Tools returns the Traktor tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_traktor_status",
				mcp.WithDescription("Get Traktor library status including track count and collection path"),
			),
			Handler:             handleTraktorStatus,
			Category:            "traktor",
			Subcategory:         "status",
			Tags:                []string{"traktor", "dj", "library", "status"},
			UseCases:            []string{"Check library status", "Verify collection path", "Get track count"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_library",
				mcp.WithDescription("Search tracks in Traktor library by title, artist, album, or genre"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query for title, artist, album, or genre")),
				mcp.WithNumber("limit", mcp.Description("Maximum results to return (default: 20)")),
			),
			Handler:             handleTraktorLibrary,
			Category:            "traktor",
			Subcategory:         "library",
			Tags:                []string{"traktor", "search", "tracks", "library"},
			UseCases:            []string{"Find tracks by name", "Search artist discography", "Browse by genre"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_track",
				mcp.WithDescription("Get detailed track information including cue points, loops, and metadata"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("File path of the track")),
			),
			Handler:             handleTraktorTrack,
			Category:            "traktor",
			Subcategory:         "library",
			Tags:                []string{"traktor", "track", "metadata", "cues"},
			UseCases:            []string{"Get track details", "View cue points", "Check BPM/key"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_playlists",
				mcp.WithDescription("List all playlists and folders in Traktor library"),
			),
			Handler:             handleTraktorPlaylists,
			Category:            "traktor",
			Subcategory:         "playlists",
			Tags:                []string{"traktor", "playlists", "folders"},
			UseCases:            []string{"Browse playlists", "View playlist hierarchy", "Find playlist by name"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_playlist",
				mcp.WithDescription("Get contents of a specific playlist with all tracks"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Playlist name (use folder/playlist for nested playlists)")),
			),
			Handler:             handleTraktorPlaylist,
			Category:            "traktor",
			Subcategory:         "playlists",
			Tags:                []string{"traktor", "playlist", "tracks"},
			UseCases:            []string{"View playlist tracks", "Export playlist", "Get playlist info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_history",
				mcp.WithDescription("Get recently played tracks from Traktor history"),
				mcp.WithNumber("limit", mcp.Description("Maximum tracks to return (default: 20)")),
			),
			Handler:             handleTraktorHistory,
			Category:            "traktor",
			Subcategory:         "history",
			Tags:                []string{"traktor", "history", "recent", "played"},
			UseCases:            []string{"View play history", "Find recently played", "Track usage stats"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_cues",
				mcp.WithDescription("Get cue points for a specific track"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("File path of the track")),
			),
			Handler:             handleTraktorCues,
			Category:            "traktor",
			Subcategory:         "cues",
			Tags:                []string{"traktor", "cues", "hotcues", "markers"},
			UseCases:            []string{"View cue points", "Check hotcues", "Get cue positions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_loops",
				mcp.WithDescription("Get saved loops for a specific track"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("File path of the track")),
			),
			Handler:             handleTraktorLoops,
			Category:            "traktor",
			Subcategory:         "loops",
			Tags:                []string{"traktor", "loops", "sections"},
			UseCases:            []string{"View saved loops", "Get loop positions", "Check loop lengths"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_export",
				mcp.WithDescription("Export Traktor collection to Rekordbox XML format for migration"),
				mcp.WithString("output_path", mcp.Required(), mcp.Description("Output path for the Rekordbox XML file")),
			),
			Handler:             handleTraktorExport,
			Category:            "traktor",
			Subcategory:         "export",
			Tags:                []string{"traktor", "export", "rekordbox", "migration"},
			UseCases:            []string{"Export to Rekordbox", "Migrate library", "Backup collection"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "traktor",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_traktor_health",
				mcp.WithDescription("Check Traktor library health and get troubleshooting recommendations"),
			),
			Handler:             handleTraktorHealth,
			Category:            "traktor",
			Subcategory:         "status",
			Tags:                []string{"traktor", "health", "diagnostics"},
			UseCases:            []string{"Check library health", "Diagnose issues", "Troubleshoot errors"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "traktor",
		},
	}
}

var getTraktorClient = tools.LazyClient(clients.NewTraktorClient)

func handleTraktorStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleTraktorLibrary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	tracks, err := client.SearchLibrary(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to search library: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"query":  query,
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleTraktorTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	track, err := client.GetTrack(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	return tools.JSONResult(track), nil
}

func handleTraktorPlaylists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	playlists, err := client.GetPlaylists(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get playlists: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"playlists": playlists,
		"count":     len(playlists),
	}), nil
}

func handleTraktorPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	playlist, err := client.GetPlaylist(ctx, name)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get playlist: %w", err)), nil
	}

	return tools.JSONResult(playlist), nil
}

func handleTraktorHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	tracks, err := client.GetHistory(ctx, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get history: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleTraktorCues(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	cues, err := client.GetCuePoints(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get cue points: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"file_path":  filePath,
		"cue_points": cues,
		"count":      len(cues),
	}), nil
}

func handleTraktorLoops(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	loops, err := client.GetLoops(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get loops: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"file_path": filePath,
		"loops":     loops,
		"count":     len(loops),
	}), nil
}

func handleTraktorExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	outputPath, errResult := tools.RequireStringParam(req, "output_path")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.ExportToRekordboxXML(ctx, outputPath); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to export: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":     true,
		"output_path": outputPath,
		"message":     "Collection exported to Rekordbox XML format",
	}), nil
}

func handleTraktorHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTraktorClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Traktor client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
