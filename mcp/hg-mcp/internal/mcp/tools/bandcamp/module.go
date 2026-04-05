// Package bandcamp provides MCP tools for Bandcamp music discovery and downloads.
package bandcamp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewBandcampClient)

// Module implements the Bandcamp tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "bandcamp"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Bandcamp music discovery, search, and download tools for DJ libraries"
}

// Tools returns the Bandcamp tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_bandcamp_status",
				mcp.WithDescription("Get Bandcamp connection status and available download tools"),
			),
			Handler:     handleBandcampStatus,
			Category:    "bandcamp",
			Subcategory: "status",
			Tags:        []string{"bandcamp", "music", "status"},
			UseCases:    []string{"Check Bandcamp connectivity", "Verify download tools"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_health",
				mcp.WithDescription("Check Bandcamp integration health with recommendations"),
			),
			Handler:     handleBandcampHealth,
			Category:    "bandcamp",
			Subcategory: "status",
			Tags:        []string{"bandcamp", "health", "diagnostics"},
			UseCases:    []string{"Diagnose Bandcamp issues", "Get setup recommendations"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_search",
				mcp.WithDescription("Search Bandcamp for artists, albums, or tracks"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("type", mcp.Description("Filter by type: all, artist, album, track (default: all)"), mcp.Enum("all", "artist", "album", "track")),
			),
			Handler:     handleBandcampSearch,
			Category:    "bandcamp",
			Subcategory: "search",
			Tags:        []string{"bandcamp", "search", "music", "discovery"},
			UseCases:    []string{"Find artists on Bandcamp", "Search for albums", "Discover new music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_artist",
				mcp.WithDescription("Get Bandcamp artist/label details"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Artist URL (e.g., https://artist.bandcamp.com or just 'artist')")),
			),
			Handler:     handleBandcampArtist,
			Category:    "bandcamp",
			Subcategory: "artists",
			Tags:        []string{"bandcamp", "artist", "label", "profile"},
			UseCases:    []string{"Get artist info", "View artist discography", "Check label releases"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_album",
				mcp.WithDescription("Get Bandcamp album details including track listing"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Album URL (e.g., https://artist.bandcamp.com/album/album-name)")),
			),
			Handler:     handleBandcampAlbum,
			Category:    "bandcamp",
			Subcategory: "albums",
			Tags:        []string{"bandcamp", "album", "tracks", "music"},
			UseCases:    []string{"Get album details", "View track listing", "Check pricing"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_download",
				mcp.WithDescription("Download album or track from Bandcamp (requires bandcamp-dl or yt-dlp)"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Bandcamp URL to download")),
				mcp.WithString("output_dir", mcp.Description("Output directory for downloaded files (optional)")),
			),
			Handler:     handleBandcampDownload,
			Category:    "bandcamp",
			Subcategory: "download",
			Tags:        []string{"bandcamp", "download", "music", "archive"},
			UseCases:    []string{"Download free/purchased music", "Build DJ library", "Archive releases"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_tags",
				mcp.WithDescription("Get popular Bandcamp genre tags for discovery"),
			),
			Handler:     handleBandcampTags,
			Category:    "bandcamp",
			Subcategory: "discovery",
			Tags:        []string{"bandcamp", "tags", "genres", "discovery"},
			UseCases:    []string{"Browse genres", "Discover new music", "Find genre releases"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_tag_releases",
				mcp.WithDescription("Get recent releases for a specific tag/genre"),
				mcp.WithString("tag", mcp.Required(), mcp.Description("Genre tag (e.g., electronic, ambient, techno)")),
				mcp.WithNumber("page", mcp.Description("Page number for pagination (default: 1)")),
			),
			Handler:     handleBandcampTagReleases,
			Category:    "bandcamp",
			Subcategory: "discovery",
			Tags:        []string{"bandcamp", "tags", "releases", "discovery", "genre"},
			UseCases:    []string{"Browse genre releases", "Find new music by tag", "Discover trending releases"},
			Complexity:  tools.ComplexitySimple,
		},
	}

	// Add pipeline tools
	allTools = append(allTools, pipelineTools()...)

	// Apply circuit breaker to all tools — scraping service, conservative rate limits
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "bandcamp"
	}

	return allTools
}

func handleBandcampStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleBandcampHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func handleBandcampSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	searchType := tools.OptionalStringParam(req, "type", "all")

	result, err := client.Search(ctx, query, searchType)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleBandcampArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	artistURL, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	artist, err := client.GetArtist(ctx, artistURL)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist: %w", err)), nil
	}

	return tools.JSONResult(artist), nil
}

func handleBandcampAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	albumURL, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	album, err := client.GetAlbum(ctx, albumURL)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get album: %w", err)), nil
	}

	return tools.JSONResult(album), nil
}

func handleBandcampDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	bcURL, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	outputDir := tools.GetStringParam(req, "output_dir")

	result, err := client.Download(ctx, bcURL, outputDir)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleBandcampTags(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	tags, err := client.GetTags(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tags: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
		"note":  "Use aftrs_bandcamp_tag_releases to browse releases for a specific tag",
	}), nil
}

func handleBandcampTagReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Bandcamp client: %w", err)), nil
	}

	tag, errResult := tools.RequireStringParam(req, "tag")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)
	if page < 1 {
		page = 1
	}

	albums, err := client.GetTagReleases(ctx, tag, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tag releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tag":      tag,
		"page":     page,
		"releases": albums,
		"count":    len(albums),
	}), nil
}
