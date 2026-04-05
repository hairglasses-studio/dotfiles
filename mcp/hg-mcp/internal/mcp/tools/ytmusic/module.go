// Package ytmusic provides MCP tools for YouTube Music streaming service.
package ytmusic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

var getClient = tools.LazyClient(clients.NewYouTubeMusicClient)

// Module implements the YouTube Music tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "ytmusic"
}

// Description returns the module description
func (m *Module) Description() string {
	return "YouTube Music integration - search, browse, and discover music from the world's largest catalog"
}

// Tools returns all YouTube Music tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ytmusic_status",
				mcp.WithDescription("Get YouTube Music connection status including authentication and yt-dlp availability"),
			),
			Handler:             handleStatus,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "status", "connection"},
			UseCases:            []string{"Check YouTube Music connection", "Verify API access"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_health",
				mcp.WithDescription("Health check for YouTube Music integration with diagnostics and recommendations"),
			),
			Handler:             handleHealth,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "health", "diagnostics"},
			UseCases:            []string{"Diagnose YouTube Music issues", "Check integration health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_search",
				mcp.WithDescription("Search YouTube Music for songs, artists, albums, playlists, and videos"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("filter", mcp.Description("Filter by type: songs, artists, albums, playlists, videos (default: all)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 20)")),
			),
			Handler:             handleSearch,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "search", "songs"},
			UseCases:            []string{"Search YouTube Music", "Find music"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_track",
				mcp.WithDescription("Get detailed information about a YouTube Music track/video"),
				mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube video ID or full URL")),
			),
			Handler:             handleTrack,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "track", "video", "details"},
			UseCases:            []string{"Get track details", "View video info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_album",
				mcp.WithDescription("Get detailed information about a YouTube Music album including tracks"),
				mcp.WithString("browse_id", mcp.Required(), mcp.Description("Album browse ID or playlist ID")),
			),
			Handler:             handleAlbum,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "album", "tracks"},
			UseCases:            []string{"Get album details", "Browse album tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_artist",
				mcp.WithDescription("Get information about a YouTube Music artist/channel"),
				mcp.WithString("channel_id", mcp.Required(), mcp.Description("YouTube channel ID")),
			),
			Handler:             handleArtist,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "artist", "channel"},
			UseCases:            []string{"Get artist info", "Browse channel"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_playlist",
				mcp.WithDescription("Get detailed information about a YouTube Music playlist including tracks"),
				mcp.WithString("playlist_id", mcp.Required(), mcp.Description("YouTube playlist ID")),
			),
			Handler:             handlePlaylist,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "playlist", "tracks"},
			UseCases:            []string{"Get playlist details", "Browse playlist tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_charts",
				mcp.WithDescription("Get YouTube Music charts (Top Songs)"),
				mcp.WithString("country", mcp.Description("Country code for regional charts (e.g., US, GB, DE)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-100, default: 50)")),
			),
			Handler:             handleCharts,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "charts", "top", "trending"},
			UseCases:            []string{"Browse music charts", "Find trending songs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_new_releases",
				mcp.WithDescription("Get new music releases on YouTube Music"),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 25)")),
			),
			Handler:             handleNewReleases,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "new", "releases"},
			UseCases:            []string{"Discover new music", "Browse new releases"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_moods",
				mcp.WithDescription("Get available moods and genres for browsing YouTube Music"),
			),
			Handler:             handleMoods,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "moods", "genres", "browse"},
			UseCases:            []string{"Browse by mood", "Discover genres"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_radio",
				mcp.WithDescription("Generate a radio/mix playlist based on a track"),
				mcp.WithString("video_id", mcp.Required(), mcp.Description("YouTube video ID to base radio on")),
				mcp.WithNumber("limit", mcp.Description("Maximum tracks (1-100, default: 25)")),
			),
			Handler:             handleRadio,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "radio", "mix", "similar"},
			UseCases:            []string{"Generate radio playlist", "Find similar songs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_extract_id",
				mcp.WithDescription("Extract video ID from a YouTube/YouTube Music URL"),
				mcp.WithString("url", mcp.Required(), mcp.Description("YouTube URL to extract ID from")),
			),
			Handler:             handleExtractID,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "url", "parse", "id"},
			UseCases:            []string{"Extract video ID", "Parse YouTube URL"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_search_songs",
				mcp.WithDescription("Search YouTube Music specifically for songs/tracks"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 20)")),
			),
			Handler:             handleSearchSongs,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "search", "songs", "tracks"},
			UseCases:            []string{"Search for songs", "Find tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_search_albums",
				mcp.WithDescription("Search YouTube Music specifically for albums"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 20)")),
			),
			Handler:             handleSearchAlbums,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "search", "albums"},
			UseCases:            []string{"Search for albums", "Find releases"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_search_artists",
				mcp.WithDescription("Search YouTube Music specifically for artists"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 20)")),
			),
			Handler:             handleSearchArtists,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "search", "artists"},
			UseCases:            []string{"Search for artists", "Find musicians"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
		{
			Tool: mcp.NewTool("aftrs_ytmusic_search_playlists",
				mcp.WithDescription("Search YouTube Music specifically for playlists"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 20)")),
			),
			Handler:             handleSearchPlaylists,
			Category:            "music",
			Subcategory:         "ytmusic",
			Tags:                []string{"youtube", "music", "search", "playlists"},
			UseCases:            []string{"Search for playlists", "Find curated lists"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ytmusic",
		},
	}
}

// Handler implementations

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get status: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# YouTube Music Status\n\n")

	icon := "🔴"
	if status.Connected {
		icon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s **Connected:** %v\n", icon, status.Connected))
	sb.WriteString(fmt.Sprintf("- **Has Auth:** %v\n", status.HasAuth))
	if status.AuthFile != "" {
		sb.WriteString(fmt.Sprintf("- **Auth File:** %s\n", status.AuthFile))
	}

	ytdlpIcon := "🔴"
	if status.YtDlpAvailable {
		ytdlpIcon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s **yt-dlp Available:** %v\n", ytdlpIcon, status.YtDlpAvailable))

	return tools.TextResult(sb.String()), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("# YouTube Music Health\n\n🔴 **Status:** Error\n\n**Issue:** %v", err)), nil
	}

	health, err := client.Health(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get health: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# YouTube Music Health\n\n")

	icon := "✅"
	if health.Score < 50 {
		icon = "🔴"
	} else if health.Score < 80 {
		icon = "⚠️"
	}

	sb.WriteString(fmt.Sprintf("%s **Health Score:** %d/100\n", icon, health.Score))
	sb.WriteString(fmt.Sprintf("- **Status:** %s\n", health.Status))
	sb.WriteString(fmt.Sprintf("- **Connected:** %v\n", health.Connected))
	sb.WriteString(fmt.Sprintf("- **Has Auth:** %v\n", health.HasAuth))
	sb.WriteString(fmt.Sprintf("- **yt-dlp Available:** %v\n", health.YtDlpAvailable))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- 💡 %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	filter := tools.GetStringParam(req, "filter")
	limit := tools.GetIntParam(req, "limit", 20)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, filter, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# YouTube Music Search: \"%s\"\n\n", query))

	if len(results.Tracks) > 0 {
		sb.WriteString("## Songs\n\n")
		sb.WriteString("| Title | Artist | Duration |\n")
		sb.WriteString("|-------|--------|----------|\n")
		for _, t := range results.Tracks {
			artists := strings.Join(t.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s |\n",
				t.Title, t.URL, artists, t.Duration))
		}
		sb.WriteString("\n")
	}

	if len(results.Albums) > 0 {
		sb.WriteString("## Albums\n\n")
		for _, a := range results.Albums {
			artists := strings.Join(a.Artists, ", ")
			sb.WriteString(fmt.Sprintf("- [%s](%s) by %s\n", a.Title, a.URL, artists))
		}
		sb.WriteString("\n")
	}

	if len(results.Artists) > 0 {
		sb.WriteString("## Artists\n\n")
		for _, a := range results.Artists {
			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", a.Name, a.URL))
		}
		sb.WriteString("\n")
	}

	if len(results.Playlists) > 0 {
		sb.WriteString("## Playlists\n\n")
		for _, p := range results.Playlists {
			sb.WriteString(fmt.Sprintf("- [%s](%s) (%d tracks)\n", p.Title, p.URL, p.TrackCount))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoID, errResult := tools.RequireStringParam(req, "video_id")
	if errResult != nil {
		return errResult, nil
	}

	// Extract video ID from URL if needed
	videoID = clients.ExtractVideoID(videoID)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	track, err := client.GetTrack(ctx, videoID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get track: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", track.Title))

	sb.WriteString(fmt.Sprintf("**Artists:** %s\n", strings.Join(track.Artists, ", ")))
	if track.Album != "" {
		sb.WriteString(fmt.Sprintf("**Album:** %s\n", track.Album))
	}
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", track.Duration))
	if track.Year != "" {
		sb.WriteString(fmt.Sprintf("**Year:** %s\n", track.Year))
	}
	sb.WriteString(fmt.Sprintf("**Video ID:** %s\n", track.VideoID))
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", track.URL))

	return tools.TextResult(sb.String()), nil
}

func handleAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	browseID, errResult := tools.RequireStringParam(req, "browse_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	album, tracks, err := client.GetAlbum(ctx, browseID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get album: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", album.Title))

	sb.WriteString(fmt.Sprintf("**Artists:** %s\n", strings.Join(album.Artists, ", ")))
	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", album.TrackCount))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", album.Duration))
	if album.Year != "" {
		sb.WriteString(fmt.Sprintf("**Year:** %s\n", album.Year))
	}
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", album.URL))

	if len(tracks) > 0 {
		sb.WriteString("\n## Tracks\n\n")
		sb.WriteString("| # | Title | Duration |\n")
		sb.WriteString("|---|-------|----------|\n")
		for i, t := range tracks {
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s |\n",
				i+1, t.Title, t.URL, t.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID, errResult := tools.RequireStringParam(req, "channel_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	artist, err := client.GetArtist(ctx, channelID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get artist: %v", err)), nil
	}

	var sb strings.Builder
	if artist.Name != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", artist.Name))
	} else {
		sb.WriteString("# Artist\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Channel ID:** %s\n", artist.ID))
	if artist.Subscribers != "" {
		sb.WriteString(fmt.Sprintf("**Subscribers:** %s\n", artist.Subscribers))
	}
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", artist.URL))

	return tools.TextResult(sb.String()), nil
}

func handlePlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistID, errResult := tools.RequireStringParam(req, "playlist_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	playlist, tracks, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get playlist: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", playlist.Title))

	if playlist.Author != "" {
		sb.WriteString(fmt.Sprintf("**Author:** %s\n", playlist.Author))
	}
	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", playlist.TrackCount))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", playlist.Duration))
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", playlist.URL))

	if len(tracks) > 0 {
		sb.WriteString("\n## Tracks\n\n")
		sb.WriteString("| # | Title | Artist | Duration |\n")
		sb.WriteString("|---|-------|--------|----------|\n")
		for i, t := range tracks {
			artists := strings.Join(t.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
				i+1, t.Title, t.URL, artists, t.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleCharts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	country := tools.GetStringParam(req, "country")
	limit := tools.GetIntParam(req, "limit", 50)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	chart, err := client.GetCharts(ctx, country, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get charts: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# YouTube Music Charts - %s", chart.Name))
	if chart.Country != "" {
		sb.WriteString(fmt.Sprintf(" (%s)", chart.Country))
	}
	sb.WriteString("\n\n")

	if len(chart.Items) == 0 {
		sb.WriteString("No chart data available.\n")
	} else {
		sb.WriteString("| # | Song | Artist | Duration |\n")
		sb.WriteString("|---|------|--------|----------|\n")
		for _, item := range chart.Items {
			artists := strings.Join(item.Track.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
				item.Rank, item.Track.Title, item.Track.URL, artists, item.Track.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	albums, err := client.GetNewReleases(ctx, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get new releases: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# YouTube Music New Releases\n\n")

	if len(albums) == 0 {
		sb.WriteString("No new releases found.\n")
	} else {
		sb.WriteString("| Album | Artist |\n")
		sb.WriteString("|-------|--------|\n")
		for _, a := range albums {
			artists := strings.Join(a.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s |\n", a.Title, a.URL, artists))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleMoods(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	moods, err := client.GetMoods(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get moods: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# YouTube Music Moods & Genres\n\n")

	sb.WriteString("## Moods\n")
	for _, mood := range moods[:10] {
		sb.WriteString(fmt.Sprintf("- %s\n", mood))
	}

	sb.WriteString("\n## Genres\n")
	for _, genre := range moods[10:] {
		sb.WriteString(fmt.Sprintf("- %s\n", genre))
	}

	return tools.TextResult(sb.String()), nil
}

func handleRadio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	videoID, errResult := tools.RequireStringParam(req, "video_id")
	if errResult != nil {
		return errResult, nil
	}

	videoID = clients.ExtractVideoID(videoID)

	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	tracks, err := client.GetRadio(ctx, videoID, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get radio: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Radio Mix\n\n")
	sb.WriteString(fmt.Sprintf("*Based on video ID: %s*\n\n", videoID))

	if len(tracks) == 0 {
		sb.WriteString("No radio tracks found.\n")
	} else {
		sb.WriteString("| # | Song | Artist | Duration |\n")
		sb.WriteString("|---|------|--------|----------|\n")
		for i, t := range tracks {
			artists := strings.Join(t.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
				i+1, t.Title, t.URL, artists, t.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleExtractID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	videoID := clients.ExtractVideoID(url)

	var sb strings.Builder
	sb.WriteString("# Video ID Extraction\n\n")
	sb.WriteString(fmt.Sprintf("**Input:** %s\n", url))
	sb.WriteString(fmt.Sprintf("**Video ID:** %s\n", videoID))
	sb.WriteString(fmt.Sprintf("\n**YouTube Music URL:** https://music.youtube.com/watch?v=%s\n", videoID))
	sb.WriteString(fmt.Sprintf("**YouTube URL:** https://youtube.com/watch?v=%s\n", videoID))

	return tools.TextResult(sb.String()), nil
}

func handleSearchSongs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, "songs", limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Songs: \"%s\"\n\n", query))

	if len(results.Tracks) == 0 {
		sb.WriteString("No songs found.\n")
	} else {
		sb.WriteString("| Song | Artist | Duration |\n")
		sb.WriteString("|------|--------|----------|\n")
		for _, t := range results.Tracks {
			artists := strings.Join(t.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s |\n",
				t.Title, t.URL, artists, t.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSearchAlbums(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, "albums", limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Albums: \"%s\"\n\n", query))

	if len(results.Albums) == 0 && len(results.Tracks) == 0 {
		sb.WriteString("No albums found.\n")
	} else if len(results.Albums) > 0 {
		sb.WriteString("| Album | Artist | Year |\n")
		sb.WriteString("|-------|--------|------|\n")
		for _, a := range results.Albums {
			artists := strings.Join(a.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s |\n",
				a.Title, a.URL, artists, a.Year))
		}
	} else {
		// Fallback to tracks if no albums found
		sb.WriteString("| Track | Artist | Duration |\n")
		sb.WriteString("|-------|--------|----------|\n")
		for _, t := range results.Tracks {
			artists := strings.Join(t.Artists, ", ")
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s |\n",
				t.Title, t.URL, artists, t.Duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSearchArtists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, "artists", limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Artists: \"%s\"\n\n", query))

	if len(results.Artists) == 0 && len(results.Tracks) == 0 {
		sb.WriteString("No artists found.\n")
	} else if len(results.Artists) > 0 {
		for _, a := range results.Artists {
			sb.WriteString(fmt.Sprintf("- [%s](%s)", a.Name, a.URL))
			if a.Subscribers != "" {
				sb.WriteString(fmt.Sprintf(" (%s subscribers)", a.Subscribers))
			}
			sb.WriteString("\n")
		}
	} else {
		// Fallback to tracks with unique artists
		sb.WriteString("*Artists found in tracks:*\n\n")
		seen := make(map[string]bool)
		for _, t := range results.Tracks {
			for _, artist := range t.Artists {
				if !seen[artist] {
					seen[artist] = true
					sb.WriteString(fmt.Sprintf("- %s\n", artist))
				}
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSearchPlaylists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("YouTube Music client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, "playlists", limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Playlists: \"%s\"\n\n", query))

	if len(results.Playlists) == 0 {
		sb.WriteString("No playlists found.\n")
	} else {
		sb.WriteString("| Playlist | Author | Tracks |\n")
		sb.WriteString("|----------|--------|--------|\n")
		for _, p := range results.Playlists {
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %d |\n",
				p.Title, p.URL, p.Author, p.TrackCount))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
