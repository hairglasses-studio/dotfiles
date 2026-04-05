// Package boomkat provides MCP tools for Boomkat electronic/experimental music store.
package boomkat

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewBoomkatClient)

// Module implements the Boomkat tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "boomkat"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Boomkat electronic and experimental music store for curated releases"
}

// Tools returns the Boomkat tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_boomkat_status",
				mcp.WithDescription("Get Boomkat connection status and availability"),
			),
			Handler:     handleBoomkatStatus,
			Category:    "boomkat",
			Subcategory: "status",
			Tags:        []string{"boomkat", "electronic", "experimental", "ambient", "status"},
			UseCases:    []string{"Check Boomkat availability", "Verify connection"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_health",
				mcp.WithDescription("Check Boomkat integration health with recommendations"),
			),
			Handler:     handleBoomkatHealth,
			Category:    "boomkat",
			Subcategory: "status",
			Tags:        []string{"boomkat", "health", "diagnostics"},
			UseCases:    []string{"Diagnose Boomkat issues", "Get setup recommendations"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_search",
				mcp.WithDescription("Search Boomkat for releases, artists, or labels"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatSearch,
			Category:    "boomkat",
			Subcategory: "search",
			Tags:        []string{"boomkat", "search", "electronic", "experimental"},
			UseCases:    []string{"Find releases", "Search artists", "Discover labels"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_release",
				mcp.WithDescription("Get detailed release information with tracklist"),
				mcp.WithString("release_slug", mcp.Required(), mcp.Description("Boomkat release slug (from URL)")),
			),
			Handler:     handleBoomkatRelease,
			Category:    "boomkat",
			Subcategory: "releases",
			Tags:        []string{"boomkat", "release", "album", "tracklist"},
			UseCases:    []string{"Get release details", "View tracklist"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_artist",
				mcp.WithDescription("Get artist profile information"),
				mcp.WithString("artist_slug", mcp.Required(), mcp.Description("Boomkat artist slug (URL-friendly name)")),
			),
			Handler:     handleBoomkatArtist,
			Category:    "boomkat",
			Subcategory: "artists",
			Tags:        []string{"boomkat", "artist", "profile"},
			UseCases:    []string{"Get artist info", "View artist profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_artist_releases",
				mcp.WithDescription("Get releases by an artist"),
				mcp.WithString("artist_slug", mcp.Required(), mcp.Description("Boomkat artist slug")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatArtistReleases,
			Category:    "boomkat",
			Subcategory: "artists",
			Tags:        []string{"boomkat", "artist", "releases", "discography"},
			UseCases:    []string{"View artist's releases", "Browse discography"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_label",
				mcp.WithDescription("Get record label information"),
				mcp.WithString("label_slug", mcp.Required(), mcp.Description("Boomkat label slug (URL-friendly name)")),
			),
			Handler:     handleBoomkatLabel,
			Category:    "boomkat",
			Subcategory: "labels",
			Tags:        []string{"boomkat", "label", "record-label"},
			UseCases:    []string{"Get label info", "View label profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_label_releases",
				mcp.WithDescription("Get releases from a record label"),
				mcp.WithString("label_slug", mcp.Required(), mcp.Description("Boomkat label slug")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatLabelReleases,
			Category:    "boomkat",
			Subcategory: "labels",
			Tags:        []string{"boomkat", "label", "releases", "catalog"},
			UseCases:    []string{"Browse label releases", "View label catalog"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_genres",
				mcp.WithDescription("Get list of available music genres/categories"),
			),
			Handler:     handleBoomkatGenres,
			Category:    "boomkat",
			Subcategory: "browse",
			Tags:        []string{"boomkat", "genres", "categories", "electronic", "experimental"},
			UseCases:    []string{"List genres", "Browse genre options"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_genre_releases",
				mcp.WithDescription("Browse releases by genre"),
				mcp.WithString("genre", mcp.Required(), mcp.Description("Genre slug (e.g., techno, ambient, experimental, idm, industrial)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatGenreReleases,
			Category:    "boomkat",
			Subcategory: "browse",
			Tags:        []string{"boomkat", "genre", "browse", "discover"},
			UseCases:    []string{"Browse genre releases", "Discover new music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_new_releases",
				mcp.WithDescription("Get new releases, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Genre slug to filter (optional)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatNewReleases,
			Category:    "boomkat",
			Subcategory: "browse",
			Tags:        []string{"boomkat", "new", "releases", "latest"},
			UseCases:    []string{"Find new releases", "Browse latest music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_bestsellers",
				mcp.WithDescription("Get bestselling releases, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Genre slug to filter (optional)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatBestsellers,
			Category:    "boomkat",
			Subcategory: "charts",
			Tags:        []string{"boomkat", "bestsellers", "top", "popular"},
			UseCases:    []string{"View bestsellers", "Find popular releases"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_recommended",
				mcp.WithDescription("Get Boomkat recommended releases (staff picks)"),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatRecommended,
			Category:    "boomkat",
			Subcategory: "curated",
			Tags:        []string{"boomkat", "recommended", "staff-picks", "curated"},
			UseCases:    []string{"Browse recommended releases", "Find curated picks"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_essential",
				mcp.WithDescription("Get Boomkat essential releases (highest rated)"),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)"), mcp.Min(1)),
			),
			Handler:     handleBoomkatEssential,
			Category:    "boomkat",
			Subcategory: "curated",
			Tags:        []string{"boomkat", "essential", "must-have", "best"},
			UseCases:    []string{"Find essential releases", "View best-rated music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_boomkat_download",
				mcp.WithDescription("Download release preview using yt-dlp"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Boomkat release URL")),
				mcp.WithString("output_dir", mcp.Description("Output directory for downloaded file")),
			),
			Handler:     handleBoomkatDownload,
			Category:    "boomkat",
			Subcategory: "download",
			Tags:        []string{"boomkat", "download", "preview"},
			UseCases:    []string{"Download release preview", "Get sample audio"},
			Complexity:  tools.ComplexityModerate,
		},
	}

	// Apply circuit breaker to all tools — scraping service, conservative rate limits
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "boomkat"
	}

	return allTools
}

// Handler functions

func handleBoomkatStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleBoomkatHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func handleBoomkatSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	results, err := client.Search(ctx, query, "", page, 25)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleBoomkatRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	releaseSlug, errResult := tools.RequireStringParam(req, "release_slug")
	if errResult != nil {
		return errResult, nil
	}

	release, err := client.GetRelease(ctx, releaseSlug)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get release: %w", err)), nil
	}

	return tools.JSONResult(release), nil
}

func handleBoomkatArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	artistSlug, errResult := tools.RequireStringParam(req, "artist_slug")
	if errResult != nil {
		return errResult, nil
	}

	artist, err := client.GetArtist(ctx, artistSlug)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist: %w", err)), nil
	}

	return tools.JSONResult(artist), nil
}

func handleBoomkatArtistReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	artistSlug, errResult := tools.RequireStringParam(req, "artist_slug")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetArtistReleases(ctx, artistSlug, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"artist_slug": artistSlug,
		"page":        page,
		"releases":    releases,
		"count":       len(releases),
	}), nil
}

func handleBoomkatLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	labelSlug, errResult := tools.RequireStringParam(req, "label_slug")
	if errResult != nil {
		return errResult, nil
	}

	label, err := client.GetLabel(ctx, labelSlug)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get label: %w", err)), nil
	}

	return tools.JSONResult(label), nil
}

func handleBoomkatLabelReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	labelSlug, errResult := tools.RequireStringParam(req, "label_slug")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetLabelReleases(ctx, labelSlug, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get label releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"label_slug": labelSlug,
		"page":       page,
		"releases":   releases,
		"count":      len(releases),
	}), nil
}

func handleBoomkatGenres(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genres, err := client.GetGenres(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get genres: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genres": genres,
		"count":  len(genres),
	}), nil
}

func handleBoomkatGenreReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre, errResult := tools.RequireStringParam(req, "genre")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetGenreReleases(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get genre releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":    genre,
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleBoomkatNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre := tools.GetStringParam(req, "genre")
	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetNewReleases(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get new releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":    genre,
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleBoomkatBestsellers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre := tools.GetStringParam(req, "genre")
	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetBestsellers(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bestsellers: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":    genre,
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleBoomkatRecommended(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetRecommended(ctx, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get recommended: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleBoomkatEssential(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetEssential(ctx, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get essential: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleBoomkatDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	url, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	outputDir := tools.OptionalStringParam(req, "output_dir", ".")

	output, err := client.DownloadPreview(ctx, url, outputDir)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":    true,
		"url":        url,
		"output_dir": outputDir,
		"output":     output,
	}), nil
}
