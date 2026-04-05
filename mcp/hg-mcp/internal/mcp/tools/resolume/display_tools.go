// Package resolume provides MCP tools for Resolume track display.
package resolume

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/bridge"
	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// DisplayTools returns the Resolume display tool definitions
func DisplayTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_resolume_display_start",
				mcp.WithDescription("Start syncing XDJ/CDJ track info to Resolume for visual display"),
				mcp.WithString("mode", mcp.Description("Display mode: 'separate' (artist/title on different strings), 'combined' (single 'Artist - Title'), or 'full' (all metadata)"), mcp.Enum("separate", "combined", "full")),
				mcp.WithNumber("poll_interval_ms", mcp.Description("How often to poll for track changes (default 500ms)")),
			),
			Handler:     handleDisplayStart,
			Category:    "resolume",
			Subcategory: "display",
			Tags:        []string{"resolume", "display", "track", "sync", "prolink"},
			UseCases:    []string{"Start track display sync", "Show now playing in Resolume"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_display_stop",
				mcp.WithDescription("Stop syncing track info to Resolume"),
			),
			Handler:     handleDisplayStop,
			Category:    "resolume",
			Subcategory: "display",
			Tags:        []string{"resolume", "display", "stop"},
			UseCases:    []string{"Stop track display sync"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_display_status",
				mcp.WithDescription("Get the status of the Resolume track display sync"),
			),
			Handler:     handleDisplayStatus,
			Category:    "resolume",
			Subcategory: "display",
			Tags:        []string{"resolume", "display", "status"},
			UseCases:    []string{"Check display sync status", "View current track"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_display_set",
				mcp.WithDescription("Manually set the track display in Resolume (bypasses XDJ sync)"),
				mcp.WithString("artist", mcp.Description("Artist name to display")),
				mcp.WithString("title", mcp.Description("Track title to display")),
				mcp.WithString("key", mcp.Description("Musical key (e.g., '8A', 'Cm')")),
				mcp.WithNumber("bpm", mcp.Description("Beats per minute")),
				mcp.WithString("genre", mcp.Description("Genre/style")),
			),
			Handler:     handleDisplaySet,
			Category:    "resolume",
			Subcategory: "display",
			Tags:        []string{"resolume", "display", "manual", "set"},
			UseCases:    []string{"Manually set track display", "Test display"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_display_clear",
				mcp.WithDescription("Clear the track display in Resolume"),
			),
			Handler:     handleDisplayClear,
			Category:    "resolume",
			Subcategory: "display",
			Tags:        []string{"resolume", "display", "clear"},
			UseCases:    []string{"Clear track display", "Reset display"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
	}
}

func handleDisplayStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetResolumeDisplayBridge()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	// Configure if options provided
	mode := tools.GetStringParam(req, "mode")
	pollInterval := tools.GetIntParam(req, "poll_interval_ms", 0)

	if mode != "" || pollInterval > 0 {
		config := &bridge.ResolumeConfig{
			UpdateOnLoad: true,
		}
		if mode != "" {
			config.DisplayMode = bridge.ResolumeDisplayMode(mode)
		}
		if pollInterval > 0 {
			config.PollIntervalMs = pollInterval
		}
		b.Configure(config)
	}

	if err := b.Start(ctx); err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to start display sync: %w", err)), nil
	}

	status := b.GetStatus()
	return tools.JSONResult(map[string]interface{}{
		"success":      true,
		"message":      "Track display sync started",
		"display_mode": status.DisplayMode,
	}), nil
}

func handleDisplayStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetResolumeDisplayBridge()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	b.Stop()

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Track display sync stopped",
	}), nil
}

func handleDisplayStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	b, err := bridge.GetResolumeDisplayBridge()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to get bridge: %w", err)), nil
	}

	return tools.JSONResult(b.GetStatus()), nil
}

func handleDisplaySet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resolume, err := clients.NewResolumeClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create resolume client: %w", err)), nil
	}

	artist := tools.GetStringParam(req, "artist")
	title := tools.GetStringParam(req, "title")
	key := tools.GetStringParam(req, "key")
	bpm := tools.GetFloatParam(req, "bpm", 0)
	genre := tools.GetStringParam(req, "genre")

	// If only artist and title, use simple method
	if key == "" && bpm == 0 && genre == "" {
		if err := resolume.SetNowPlaying(artist, title); err != nil {
			return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to set display: %w", err)), nil
		}
	} else {
		// Use full track info
		if err := resolume.SetTrackInfo(clients.TrackDisplay{
			Artist: artist,
			Title:  title,
			Key:    key,
			BPM:    bpm,
			Genre:  genre,
		}); err != nil {
			return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to set display: %w", err)), nil
		}
	}

	result := map[string]interface{}{
		"success": true,
		"message": "Track display updated",
		"artist":  artist,
		"title":   title,
	}
	if key != "" {
		result["key"] = key
	}
	if bpm > 0 {
		result["bpm"] = bpm
	}
	if genre != "" {
		result["genre"] = genre
	}
	return tools.JSONResult(result), nil
}

func handleDisplayClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resolume, err := clients.NewResolumeClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create resolume client: %w", err)), nil
	}

	if err := resolume.ClearTrackDisplay(); err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to clear display: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success": true,
		"message": "Track display cleared",
	}), nil
}
