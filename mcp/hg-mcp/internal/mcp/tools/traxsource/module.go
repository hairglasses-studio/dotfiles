// Package traxsource provides MCP tools for Traxsource electronic music store.
package traxsource

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewTraxsourceClient)

// Module implements the Traxsource tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "traxsource"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Traxsource electronic music store for house, techno, and electronic tracks"
}

// Tools returns the Traxsource tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_traxsource_status",
				mcp.WithDescription("Get Traxsource connection status and availability"),
			),
			Handler:     handleTraxsourceStatus,
			Category:    "traxsource",
			Subcategory: "status",
			Tags:        []string{"traxsource", "electronic", "house", "techno", "status"},
			UseCases:    []string{"Check Traxsource availability", "Verify connection"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_health",
				mcp.WithDescription("Check Traxsource integration health with recommendations"),
			),
			Handler:     handleTraxsourceHealth,
			Category:    "traxsource",
			Subcategory: "status",
			Tags:        []string{"traxsource", "health", "diagnostics"},
			UseCases:    []string{"Diagnose Traxsource issues", "Get setup recommendations"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_search",
				mcp.WithDescription("Search Traxsource for tracks, releases, artists, or labels"),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query"),
				),
				mcp.WithString("type",
					mcp.Description("Search type: tracks, releases, artists, labels"),
					mcp.Enum("tracks", "releases", "artists", "labels"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceSearch,
			Category:    "traxsource",
			Subcategory: "search",
			Tags:        []string{"traxsource", "search", "electronic", "house", "techno"},
			UseCases:    []string{"Find tracks", "Search artists", "Discover labels"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_track",
				mcp.WithDescription("Get detailed track information including BPM and key"),
				mcp.WithNumber("track_id",
					mcp.Required(),
					mcp.Description("Traxsource track ID"),
				),
			),
			Handler:     handleTraxsourceTrack,
			Category:    "traxsource",
			Subcategory: "tracks",
			Tags:        []string{"traxsource", "track", "bpm", "key", "details"},
			UseCases:    []string{"Get track details", "Find BPM and key", "View track info"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_release",
				mcp.WithDescription("Get release/EP details with tracklist"),
				mcp.WithNumber("release_id",
					mcp.Required(),
					mcp.Description("Traxsource release ID"),
				),
			),
			Handler:     handleTraxsourceRelease,
			Category:    "traxsource",
			Subcategory: "releases",
			Tags:        []string{"traxsource", "release", "ep", "album", "tracklist"},
			UseCases:    []string{"Get release details", "View EP tracklist"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_artist",
				mcp.WithDescription("Get artist profile information"),
				mcp.WithNumber("artist_id",
					mcp.Required(),
					mcp.Description("Traxsource artist ID"),
				),
			),
			Handler:     handleTraxsourceArtist,
			Category:    "traxsource",
			Subcategory: "artists",
			Tags:        []string{"traxsource", "artist", "profile", "dj", "producer"},
			UseCases:    []string{"Get artist info", "View artist profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_artist_tracks",
				mcp.WithDescription("Get tracks by an artist"),
				mcp.WithNumber("artist_id",
					mcp.Required(),
					mcp.Description("Traxsource artist ID"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceArtistTracks,
			Category:    "traxsource",
			Subcategory: "artists",
			Tags:        []string{"traxsource", "artist", "tracks", "discography"},
			UseCases:    []string{"View artist's tracks", "Browse discography"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_label",
				mcp.WithDescription("Get record label information"),
				mcp.WithNumber("label_id",
					mcp.Required(),
					mcp.Description("Traxsource label ID"),
				),
			),
			Handler:     handleTraxsourceLabel,
			Category:    "traxsource",
			Subcategory: "labels",
			Tags:        []string{"traxsource", "label", "record-label"},
			UseCases:    []string{"Get label info", "View label profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_label_releases",
				mcp.WithDescription("Get releases from a record label"),
				mcp.WithNumber("label_id",
					mcp.Required(),
					mcp.Description("Traxsource label ID"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceLabelReleases,
			Category:    "traxsource",
			Subcategory: "labels",
			Tags:        []string{"traxsource", "label", "releases", "catalog"},
			UseCases:    []string{"Browse label releases", "View label catalog"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_genres",
				mcp.WithDescription("Get list of available electronic music genres"),
			),
			Handler:     handleTraxsourceGenres,
			Category:    "traxsource",
			Subcategory: "browse",
			Tags:        []string{"traxsource", "genres", "electronic", "house", "techno"},
			UseCases:    []string{"List genres", "Browse genre options"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_genre_tracks",
				mcp.WithDescription("Browse tracks by genre"),
				mcp.WithString("genre",
					mcp.Required(),
					mcp.Description("Genre slug (e.g., house, tech-house, deep-house, techno)"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceGenreTracks,
			Category:    "traxsource",
			Subcategory: "browse",
			Tags:        []string{"traxsource", "genre", "browse", "discover"},
			UseCases:    []string{"Browse genre tracks", "Discover new music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_top_charts",
				mcp.WithDescription("Get top charts, optionally filtered by genre"),
				mcp.WithString("genre",
					mcp.Description("Genre slug to filter (optional)"),
				),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceTopCharts,
			Category:    "traxsource",
			Subcategory: "charts",
			Tags:        []string{"traxsource", "charts", "top", "trending"},
			UseCases:    []string{"View top charts", "Find trending tracks"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_dj_charts",
				mcp.WithDescription("Get DJ charts and curated selections"),
				mcp.WithNumber("page",
					mcp.Description("Page number (default: 1)"),
				),
			),
			Handler:     handleTraxsourceDJCharts,
			Category:    "traxsource",
			Subcategory: "charts",
			Tags:        []string{"traxsource", "dj", "charts", "curated"},
			UseCases:    []string{"Browse DJ charts", "Find curated selections"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_traxsource_download",
				mcp.WithDescription("Download track preview using yt-dlp"),
				mcp.WithString("url",
					mcp.Required(),
					mcp.Description("Traxsource track URL"),
				),
				mcp.WithString("output_dir",
					mcp.Description("Output directory for downloaded file"),
				),
			),
			Handler:     handleTraxsourceDownload,
			Category:    "traxsource",
			Subcategory: "download",
			Tags:        []string{"traxsource", "download", "preview"},
			UseCases:    []string{"Download track preview", "Get sample audio"},
			Complexity:  tools.ComplexityModerate,
		},
	}

	// Apply circuit breaker to all tools — scraping service, conservative rate limits
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "traxsource"
	}

	return allTools
}

// Handler functions

func handleTraxsourceStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleTraxsourceHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleTraxsourceSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	searchType := tools.OptionalStringParam(req, "type", "tracks")

	page := tools.GetIntParam(req, "page", 1)

	results, err := client.Search(ctx, query, searchType, page, 25)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleTraxsourceTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	trackID, errResult := tools.RequireIntParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	return tools.JSONResult(track), nil
}

func handleTraxsourceRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	releaseID, errResult := tools.RequireIntParam(req, "release_id")
	if errResult != nil {
		return errResult, nil
	}

	release, err := client.GetRelease(ctx, releaseID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get release: %w", err)), nil
	}

	return tools.JSONResult(release), nil
}

func handleTraxsourceArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	artist, err := client.GetArtist(ctx, artistID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist: %w", err)), nil
	}

	return tools.JSONResult(artist), nil
}

func handleTraxsourceArtistTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	tracks, err := client.GetArtistTracks(ctx, artistID, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist tracks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"artist_id": artistID,
		"page":      page,
		"tracks":    tracks,
		"count":     len(tracks),
	}), nil
}

func handleTraxsourceLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	labelID, errResult := tools.RequireIntParam(req, "label_id")
	if errResult != nil {
		return errResult, nil
	}

	label, err := client.GetLabel(ctx, labelID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get label: %w", err)), nil
	}

	return tools.JSONResult(label), nil
}

func handleTraxsourceLabelReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	labelID, errResult := tools.RequireIntParam(req, "label_id")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetLabelReleases(ctx, labelID, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get label releases: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"label_id": labelID,
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleTraxsourceGenres(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleTraxsourceGenreTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre, errResult := tools.RequireStringParam(req, "genre")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	tracks, err := client.GetGenreTracks(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get genre tracks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":  genre,
		"page":   page,
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleTraxsourceTopCharts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre := tools.GetStringParam(req, "genre")
	page := tools.GetIntParam(req, "page", 1)

	tracks, err := client.GetTopCharts(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get top charts: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":  genre,
		"page":   page,
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleTraxsourceDJCharts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	page := tools.GetIntParam(req, "page", 1)

	charts, err := client.GetDJCharts(ctx, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get DJ charts: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"page":   page,
		"charts": charts,
		"count":  len(charts),
	}), nil
}

func handleTraxsourceDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
