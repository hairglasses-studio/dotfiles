// Package juno provides MCP tools for Juno Download electronic music store.
package juno

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewJunoClient)

// Module implements the Juno Download tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "juno"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Juno Download electronic music store for house, techno, drum & bass, and more"
}

// Tools returns the Juno Download tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_juno_status",
				mcp.WithDescription("Get Juno Download connection status and availability"),
			),
			Handler:     handleJunoStatus,
			Category:    "juno",
			Subcategory: "status",
			Tags:        []string{"juno", "electronic", "house", "techno", "dnb", "status"},
			UseCases:    []string{"Check Juno Download availability", "Verify connection"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_health",
				mcp.WithDescription("Check Juno Download integration health with recommendations"),
			),
			Handler:     handleJunoHealth,
			Category:    "juno",
			Subcategory: "status",
			Tags:        []string{"juno", "health", "diagnostics"},
			UseCases:    []string{"Diagnose Juno issues", "Get setup recommendations"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_search",
				mcp.WithDescription("Search Juno Download for tracks, releases, artists, or labels"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("type", mcp.Description("Search type: tracks, releases, artists, labels"), mcp.Enum("tracks", "releases", "artists", "labels")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoSearch,
			Category:    "juno",
			Subcategory: "search",
			Tags:        []string{"juno", "search", "electronic", "house", "techno", "dnb"},
			UseCases:    []string{"Find tracks", "Search artists", "Discover labels"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_track",
				mcp.WithDescription("Get detailed track information including BPM, key, and format options"),
				mcp.WithNumber("track_id", mcp.Required(), mcp.Description("Juno track ID")),
			),
			Handler:     handleJunoTrack,
			Category:    "juno",
			Subcategory: "tracks",
			Tags:        []string{"juno", "track", "bpm", "key", "wav", "flac", "details"},
			UseCases:    []string{"Get track details", "Find BPM and key", "Check format options"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_release",
				mcp.WithDescription("Get release/EP details with tracklist"),
				mcp.WithNumber("release_id", mcp.Required(), mcp.Description("Juno release ID")),
			),
			Handler:     handleJunoRelease,
			Category:    "juno",
			Subcategory: "releases",
			Tags:        []string{"juno", "release", "ep", "album", "tracklist"},
			UseCases:    []string{"Get release details", "View EP tracklist"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_artist",
				mcp.WithDescription("Get artist profile information"),
				mcp.WithString("artist_slug", mcp.Required(), mcp.Description("Juno artist slug (URL-friendly name)")),
			),
			Handler:     handleJunoArtist,
			Category:    "juno",
			Subcategory: "artists",
			Tags:        []string{"juno", "artist", "profile", "dj", "producer"},
			UseCases:    []string{"Get artist info", "View artist profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_artist_releases",
				mcp.WithDescription("Get releases by an artist"),
				mcp.WithString("artist_slug", mcp.Required(), mcp.Description("Juno artist slug")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoArtistReleases,
			Category:    "juno",
			Subcategory: "artists",
			Tags:        []string{"juno", "artist", "releases", "discography"},
			UseCases:    []string{"View artist's releases", "Browse discography"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_label",
				mcp.WithDescription("Get record label information"),
				mcp.WithString("label_slug", mcp.Required(), mcp.Description("Juno label slug (URL-friendly name)")),
			),
			Handler:     handleJunoLabel,
			Category:    "juno",
			Subcategory: "labels",
			Tags:        []string{"juno", "label", "record-label"},
			UseCases:    []string{"Get label info", "View label profile"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_label_releases",
				mcp.WithDescription("Get releases from a record label"),
				mcp.WithString("label_slug", mcp.Required(), mcp.Description("Juno label slug")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoLabelReleases,
			Category:    "juno",
			Subcategory: "labels",
			Tags:        []string{"juno", "label", "releases", "catalog"},
			UseCases:    []string{"Browse label releases", "View label catalog"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_genres",
				mcp.WithDescription("Get list of available electronic music genres"),
			),
			Handler:     handleJunoGenres,
			Category:    "juno",
			Subcategory: "browse",
			Tags:        []string{"juno", "genres", "electronic", "house", "techno", "dnb"},
			UseCases:    []string{"List genres", "Browse genre options"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_genre_tracks",
				mcp.WithDescription("Browse tracks by genre"),
				mcp.WithString("genre", mcp.Required(), mcp.Description("Genre slug (e.g., house, tech-house, techno, drum-and-bass)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoGenreTracks,
			Category:    "juno",
			Subcategory: "browse",
			Tags:        []string{"juno", "genre", "browse", "discover"},
			UseCases:    []string{"Browse genre tracks", "Discover new music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_new_releases",
				mcp.WithDescription("Get this week's new releases, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Genre slug to filter (optional)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoNewReleases,
			Category:    "juno",
			Subcategory: "browse",
			Tags:        []string{"juno", "new", "releases", "this-week"},
			UseCases:    []string{"Find new releases", "Browse latest music"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_bestsellers",
				mcp.WithDescription("Get bestselling tracks, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Genre slug to filter (optional)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoBestsellers,
			Category:    "juno",
			Subcategory: "charts",
			Tags:        []string{"juno", "bestsellers", "top", "popular"},
			UseCases:    []string{"View bestsellers", "Find popular tracks"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_staff_picks",
				mcp.WithDescription("Get staff picks and featured releases"),
				mcp.WithString("genre", mcp.Description("Genre slug to filter (optional)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoStaffPicks,
			Category:    "juno",
			Subcategory: "curated",
			Tags:        []string{"juno", "staff-picks", "featured", "curated"},
			UseCases:    []string{"Browse staff picks", "Find featured releases"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_dj_charts",
				mcp.WithDescription("Get DJ charts and curated selections"),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:     handleJunoDJCharts,
			Category:    "juno",
			Subcategory: "charts",
			Tags:        []string{"juno", "dj", "charts", "curated"},
			UseCases:    []string{"Browse DJ charts", "Find curated selections"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_juno_download",
				mcp.WithDescription("Download track preview using yt-dlp"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Juno Download track URL")),
				mcp.WithString("output_dir", mcp.Description("Output directory for downloaded file")),
			),
			Handler:     handleJunoDownload,
			Category:    "juno",
			Subcategory: "download",
			Tags:        []string{"juno", "download", "preview"},
			UseCases:    []string{"Download track preview", "Get sample audio"},
			Complexity:  tools.ComplexityModerate,
		},
	}

	// Apply circuit breaker to all tools — scraping service, conservative rate limits
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "juno"
	}

	return allTools
}

// Handler functions

func handleJunoStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoArtistReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoLabelReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoGenres(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoGenreTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoBestsellers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre := tools.GetStringParam(req, "genre")
	page := tools.GetIntParam(req, "page", 1)

	tracks, err := client.GetTopSellers(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get bestsellers: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":  genre,
		"page":   page,
		"tracks": tracks,
		"count":  len(tracks),
	}), nil
}

func handleJunoStaffPicks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to initialize client: %w", err)), nil
	}

	genre := tools.GetStringParam(req, "genre")
	page := tools.GetIntParam(req, "page", 1)

	releases, err := client.GetStaffPicks(ctx, genre, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get staff picks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":    genre,
		"page":     page,
		"releases": releases,
		"count":    len(releases),
	}), nil
}

func handleJunoDJCharts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

func handleJunoDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
