// Package tidal provides MCP tools for Tidal Hi-Fi music streaming service.
package tidal

import (
	"context"
	"fmt"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

var getClient = tools.LazyClient(clients.NewTidalClient)

// Module implements the tidal tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "tidal"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Tidal Hi-Fi music streaming integration - search, browse, and discover high-quality audio content"
}

// Tools returns all tidal tools
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_tidal_status",
				mcp.WithDescription("Get Tidal API connection status including authentication state and country code"),
			),
			Handler:             handleStatus,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "status", "connection", "hifi"},
			UseCases:            []string{"Check Tidal connection", "Verify API access"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_health",
				mcp.WithDescription("Health check for Tidal integration with diagnostics and recommendations"),
			),
			Handler:             handleHealth,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Tidal issues", "Check API health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_search",
				mcp.WithDescription("Search Tidal for tracks, artists, albums, and playlists"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("types", mcp.Description("Comma-separated types to search: TRACKS,ARTISTS,ALBUMS,PLAYLISTS (default: all)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results per type (1-50, default: 10)")),
			),
			Handler:             handleSearch,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "search", "tracks", "artists", "albums"},
			UseCases:            []string{"Search Tidal catalog", "Find music"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_track",
				mcp.WithDescription("Get detailed information about a specific Tidal track including audio quality"),
				mcp.WithNumber("track_id", mcp.Required(), mcp.Description("Tidal track ID")),
			),
			Handler:             handleTrack,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "track", "details", "quality"},
			UseCases:            []string{"Get track details", "Check audio quality"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_album",
				mcp.WithDescription("Get detailed information about a Tidal album including tracks"),
				mcp.WithNumber("album_id", mcp.Required(), mcp.Description("Tidal album ID")),
				mcp.WithBoolean("include_tracks", mcp.Description("Include album tracks (default: true)")),
			),
			Handler:             handleAlbum,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "album", "details", "tracks"},
			UseCases:            []string{"Get album details", "Browse album tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_artist",
				mcp.WithDescription("Get detailed information about a Tidal artist including top tracks and albums"),
				mcp.WithNumber("artist_id", mcp.Required(), mcp.Description("Tidal artist ID")),
				mcp.WithBoolean("include_top_tracks", mcp.Description("Include top tracks (default: true)")),
				mcp.WithBoolean("include_albums", mcp.Description("Include albums (default: true)")),
			),
			Handler:             handleArtist,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "artist", "details", "discography"},
			UseCases:            []string{"Get artist details", "Browse discography"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_playlist",
				mcp.WithDescription("Get detailed information about a Tidal playlist including tracks"),
				mcp.WithString("playlist_uuid", mcp.Required(), mcp.Description("Tidal playlist UUID")),
				mcp.WithBoolean("include_tracks", mcp.Description("Include playlist tracks (default: true)")),
				mcp.WithNumber("track_limit", mcp.Description("Maximum tracks to return (1-100, default: 50)")),
			),
			Handler:             handlePlaylist,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "playlist", "tracks", "curated"},
			UseCases:            []string{"Get playlist details", "Browse playlist tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_genres",
				mcp.WithDescription("Get available Tidal genres/categories for browsing"),
			),
			Handler:             handleGenres,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "genres", "categories", "browse"},
			UseCases:            []string{"Browse genres", "Discover categories"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_new_releases",
				mcp.WithDescription("Get new album releases on Tidal, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Filter by genre path (use aftrs_tidal_genres to see options)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 25)")),
			),
			Handler:             handleNewReleases,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "new", "releases", "discover"},
			UseCases:            []string{"Discover new music", "Browse new releases"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_bestsellers",
				mcp.WithDescription("Get bestselling/top albums on Tidal, optionally filtered by genre"),
				mcp.WithString("genre", mcp.Description("Filter by genre path (use aftrs_tidal_genres to see options)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 25)")),
			),
			Handler:             handleBestsellers,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "top", "bestsellers", "charts"},
			UseCases:            []string{"Browse top albums", "Discover popular music"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_quality_info",
				mcp.WithDescription("Get detailed audio quality information for a track (MQA, Dolby Atmos, Hi-Res)"),
				mcp.WithNumber("track_id", mcp.Required(), mcp.Description("Tidal track ID")),
			),
			Handler:             handleQualityInfo,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "quality", "mqa", "dolby", "hires"},
			UseCases:            []string{"Check audio quality", "Find Hi-Res tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_similar_artists",
				mcp.WithDescription("Get artists similar to a given artist"),
				mcp.WithNumber("artist_id", mcp.Required(), mcp.Description("Tidal artist ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 10)")),
			),
			Handler:             handleSimilarArtists,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "similar", "artists", "discover"},
			UseCases:            []string{"Discover similar artists", "Expand music taste"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_artist_top_tracks",
				mcp.WithDescription("Get top/popular tracks from an artist"),
				mcp.WithNumber("artist_id", mcp.Required(), mcp.Description("Tidal artist ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 10)")),
			),
			Handler:             handleArtistTopTracks,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "artist", "top", "tracks", "popular"},
			UseCases:            []string{"Get artist's popular tracks", "Quick artist overview"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_artist_albums",
				mcp.WithDescription("Get all albums from an artist"),
				mcp.WithNumber("artist_id", mcp.Required(), mcp.Description("Tidal artist ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 25)")),
			),
			Handler:             handleArtistAlbums,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "artist", "albums", "discography"},
			UseCases:            []string{"Browse artist discography", "Find albums"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_mixes",
				mcp.WithDescription("Get curated mixes and playlists from Tidal"),
				mcp.WithNumber("limit", mcp.Description("Maximum results (1-50, default: 25)")),
			),
			Handler:             handleMixes,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "mixes", "curated", "playlists"},
			UseCases:            []string{"Discover curated playlists", "Find DJ mixes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_album_tracks",
				mcp.WithDescription("Get all tracks from an album"),
				mcp.WithNumber("album_id", mcp.Required(), mcp.Description("Tidal album ID")),
			),
			Handler:             handleAlbumTracks,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "album", "tracks", "tracklist"},
			UseCases:            []string{"Get album tracklist", "Browse album"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
	}
	// Add pipeline tools
	allTools = append(allTools, pipelineTools()...)
	return allTools
}

// Handler implementations

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get status: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Status\n\n")

	icon := "🔴"
	if status.Connected {
		icon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s **Connected:** %v\n", icon, status.Connected))
	sb.WriteString(fmt.Sprintf("- **Has Token:** %v\n", status.HasToken))
	if status.TokenExpiry != "" {
		sb.WriteString(fmt.Sprintf("- **Token Expiry:** %s\n", status.TokenExpiry))
	}
	sb.WriteString(fmt.Sprintf("- **Country Code:** %s\n", status.CountryCode))

	return tools.TextResult(sb.String()), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("# Tidal Health\n\n🔴 **Status:** Error\n\n**Issue:** %v\n\n**Recommendation:** Set TIDAL_CLIENT_ID and TIDAL_CLIENT_SECRET environment variables", err)), nil
	}

	health, err := client.Health(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get health: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Health\n\n")

	icon := "✅"
	if health.Score < 50 {
		icon = "🔴"
	} else if health.Score < 80 {
		icon = "⚠️"
	}

	sb.WriteString(fmt.Sprintf("%s **Health Score:** %d/100\n", icon, health.Score))
	sb.WriteString(fmt.Sprintf("- **Status:** %s\n", health.Status))
	sb.WriteString(fmt.Sprintf("- **Connected:** %v\n", health.Connected))
	sb.WriteString(fmt.Sprintf("- **Has Credentials:** %v\n", health.HasCredentials))

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

	typesStr := tools.GetStringParam(req, "types")
	var types []string
	if typesStr != "" {
		types = strings.Split(strings.ToUpper(typesStr), ",")
	}

	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	results, err := client.Search(ctx, query, types, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Tidal Search: \"%s\"\n\n", query))

	if len(results.Tracks) > 0 {
		sb.WriteString("## Tracks\n\n")
		sb.WriteString("| Title | Artist | Album | Quality | Duration |\n")
		sb.WriteString("|-------|--------|-------|---------|----------|\n")
		for _, t := range results.Tracks {
			artists := make([]string, len(t.Artists))
			for i, a := range t.Artists {
				artists[i] = a.Name
			}
			album := ""
			if t.Album != nil {
				album = t.Album.Title
			}
			duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %s | %s |\n",
				t.Title, t.URL, strings.Join(artists, ", "), album, t.AudioQuality, duration))
		}
		sb.WriteString("\n")
	}

	if len(results.Artists) > 0 {
		sb.WriteString("## Artists\n\n")
		for _, a := range results.Artists {
			sb.WriteString(fmt.Sprintf("- [%s](%s) (ID: %d)\n", a.Name, a.URL, a.ID))
		}
		sb.WriteString("\n")
	}

	if len(results.Albums) > 0 {
		sb.WriteString("## Albums\n\n")
		sb.WriteString("| Album | Artist | Tracks | Quality |\n")
		sb.WriteString("|-------|--------|--------|--------|\n")
		for _, a := range results.Albums {
			artists := make([]string, len(a.Artists))
			for i, art := range a.Artists {
				artists[i] = art.Name
			}
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %d | %s |\n",
				a.Title, a.URL, strings.Join(artists, ", "), a.NumberOfTracks, a.AudioQuality))
		}
		sb.WriteString("\n")
	}

	if len(results.Playlists) > 0 {
		sb.WriteString("## Playlists\n\n")
		for _, p := range results.Playlists {
			sb.WriteString(fmt.Sprintf("- [%s](%s) (%d tracks)\n", p.Title, p.URL, p.NumberOfTracks))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID, errResult := tools.RequireIntParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get track: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", track.Title))

	artists := make([]string, len(track.Artists))
	for i, a := range track.Artists {
		artists[i] = a.Name
	}
	sb.WriteString(fmt.Sprintf("**Artists:** %s\n", strings.Join(artists, ", ")))

	if track.Album != nil {
		sb.WriteString(fmt.Sprintf("**Album:** %s\n", track.Album.Title))
	}

	sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", track.Duration/60, track.Duration%60))
	sb.WriteString(fmt.Sprintf("**Audio Quality:** %s\n", track.AudioQuality))
	if len(track.AudioModes) > 0 {
		sb.WriteString(fmt.Sprintf("**Audio Modes:** %s\n", strings.Join(track.AudioModes, ", ")))
	}
	sb.WriteString(fmt.Sprintf("**Explicit:** %v\n", track.Explicit))
	sb.WriteString(fmt.Sprintf("**Popularity:** %d\n", track.Popularity))
	if track.ISRC != "" {
		sb.WriteString(fmt.Sprintf("**ISRC:** %s\n", track.ISRC))
	}
	sb.WriteString(fmt.Sprintf("**Stream Ready:** %v\n", track.StreamReady))
	sb.WriteString(fmt.Sprintf("**Premium Only:** %v\n", track.PremiumOnly))
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", track.URL))

	return tools.TextResult(sb.String()), nil
}

func handleAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	albumID, errResult := tools.RequireIntParam(req, "album_id")
	if errResult != nil {
		return errResult, nil
	}

	includeTracks := tools.GetBoolParam(req, "include_tracks", true)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	album, err := client.GetAlbum(ctx, albumID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get album: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", album.Title))

	artists := make([]string, len(album.Artists))
	for i, a := range album.Artists {
		artists[i] = a.Name
	}
	sb.WriteString(fmt.Sprintf("**Artists:** %s\n", strings.Join(artists, ", ")))
	sb.WriteString(fmt.Sprintf("**Release Date:** %s\n", album.ReleaseDate))
	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", album.NumberOfTracks))
	sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", album.Duration/60, album.Duration%60))
	sb.WriteString(fmt.Sprintf("**Audio Quality:** %s\n", album.AudioQuality))
	if len(album.AudioModes) > 0 {
		sb.WriteString(fmt.Sprintf("**Audio Modes:** %s\n", strings.Join(album.AudioModes, ", ")))
	}
	sb.WriteString(fmt.Sprintf("**Explicit:** %v\n", album.Explicit))
	if album.UPC != "" {
		sb.WriteString(fmt.Sprintf("**UPC:** %s\n", album.UPC))
	}
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", album.URL))

	if includeTracks {
		tracks, err := client.GetAlbumTracks(ctx, int(albumID))
		if err == nil && len(tracks) > 0 {
			sb.WriteString("\n## Tracks\n\n")
			sb.WriteString("| # | Title | Duration | Quality |\n")
			sb.WriteString("|---|-------|----------|--------|\n")
			for _, t := range tracks {
				duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
				sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
					t.TrackNumber, t.Title, t.URL, duration, t.AudioQuality))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	includeTopTracks := tools.GetBoolParam(req, "include_top_tracks", true)
	includeAlbums := tools.GetBoolParam(req, "include_albums", true)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	artist, err := client.GetArtist(ctx, artistID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get artist: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", artist.Name))
	sb.WriteString(fmt.Sprintf("**ID:** %d\n", artist.ID))
	sb.WriteString(fmt.Sprintf("**Popularity:** %d\n", artist.Popularity))
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", artist.URL))

	if includeTopTracks {
		tracks, err := client.GetArtistTopTracks(ctx, int(artistID), 10)
		if err == nil && len(tracks) > 0 {
			sb.WriteString("\n## Top Tracks\n\n")
			sb.WriteString("| Title | Album | Duration | Quality |\n")
			sb.WriteString("|-------|-------|----------|--------|\n")
			for _, t := range tracks {
				album := ""
				if t.Album != nil {
					album = t.Album.Title
				}
				duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
				sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %s |\n",
					t.Title, t.URL, album, duration, t.AudioQuality))
			}
		}
	}

	if includeAlbums {
		albums, err := client.GetArtistAlbums(ctx, int(artistID), 10)
		if err == nil && len(albums) > 0 {
			sb.WriteString("\n## Albums\n\n")
			sb.WriteString("| Album | Release Date | Tracks | Quality |\n")
			sb.WriteString("|-------|--------------|--------|--------|\n")
			for _, a := range albums {
				sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %d | %s |\n",
					a.Title, a.URL, a.ReleaseDate, a.NumberOfTracks, a.AudioQuality))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handlePlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistUUID, errResult := tools.RequireStringParam(req, "playlist_uuid")
	if errResult != nil {
		return errResult, nil
	}

	includeTracks := tools.GetBoolParam(req, "include_tracks", true)
	trackLimit := tools.GetIntParam(req, "track_limit", 50)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	playlist, err := client.GetPlaylist(ctx, playlistUUID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get playlist: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", playlist.Title))

	if playlist.Description != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n\n", playlist.Description))
	}

	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", playlist.NumberOfTracks))
	sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", playlist.Duration/60, playlist.Duration%60))
	if playlist.Creator != nil {
		sb.WriteString(fmt.Sprintf("**Creator:** %s\n", playlist.Creator.Name))
	}
	sb.WriteString(fmt.Sprintf("**Public:** %v\n", playlist.Public))
	if playlist.LastUpdated != "" {
		sb.WriteString(fmt.Sprintf("**Last Updated:** %s\n", playlist.LastUpdated))
	}
	sb.WriteString(fmt.Sprintf("\n**URL:** %s\n", playlist.URL))

	if includeTracks {
		tracks, err := client.GetPlaylistTracks(ctx, playlistUUID, trackLimit)
		if err == nil && len(tracks) > 0 {
			sb.WriteString("\n## Tracks\n\n")
			sb.WriteString("| # | Title | Artist | Duration |\n")
			sb.WriteString("|---|-------|--------|----------|\n")
			for i, t := range tracks {
				artists := make([]string, len(t.Artists))
				for j, a := range t.Artists {
					artists[j] = a.Name
				}
				duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
				sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
					i+1, t.Title, t.URL, strings.Join(artists, ", "), duration))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleGenres(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	genres, err := client.GetGenres(ctx)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get genres: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Genres\n\n")

	if len(genres) == 0 {
		sb.WriteString("No genres available.\n")
	} else {
		sb.WriteString("| Genre | Path |\n")
		sb.WriteString("|-------|------|\n")
		for _, g := range genres {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", g.Name, g.Path))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	genre := tools.GetStringParam(req, "genre")
	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	releases, err := client.GetNewReleases(ctx, genre, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get new releases: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal New Releases")
	if genre != "" {
		sb.WriteString(fmt.Sprintf(" - %s", genre))
	}
	sb.WriteString("\n\n")

	if len(releases) == 0 {
		sb.WriteString("No new releases found.\n")
	} else {
		sb.WriteString("| Album | Artist | Release Date | Quality |\n")
		sb.WriteString("|-------|--------|--------------|--------|\n")
		for _, r := range releases {
			artists := make([]string, len(r.Artists))
			for i, a := range r.Artists {
				artists[i] = a.Name
			}
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %s |\n",
				r.Title, r.URL, strings.Join(artists, ", "), r.ReleaseDate, r.AudioQuality))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleBestsellers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	genre := tools.GetStringParam(req, "genre")
	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	albums, err := client.GetBestsellers(ctx, genre, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get bestsellers: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Bestsellers")
	if genre != "" {
		sb.WriteString(fmt.Sprintf(" - %s", genre))
	}
	sb.WriteString("\n\n")

	if len(albums) == 0 {
		sb.WriteString("No bestsellers found.\n")
	} else {
		sb.WriteString("| # | Album | Artist | Tracks | Quality |\n")
		sb.WriteString("|---|-------|--------|--------|--------|\n")
		for i, a := range albums {
			artists := make([]string, len(a.Artists))
			for j, art := range a.Artists {
				artists[j] = art.Name
			}
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %d | %s |\n",
				i+1, a.Title, a.URL, strings.Join(artists, ", "), a.NumberOfTracks, a.AudioQuality))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleQualityInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID, errResult := tools.RequireIntParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	info, err := client.GetQualityInfo(ctx, trackID)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get quality info: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Audio Quality Info\n\n")

	sb.WriteString(fmt.Sprintf("**Audio Quality:** %s\n", info.AudioQuality))
	if len(info.AudioModes) > 0 {
		sb.WriteString(fmt.Sprintf("**Audio Modes:** %s\n", strings.Join(info.AudioModes, ", ")))
	}

	sb.WriteString("\n## Features\n\n")
	mqaIcon := "❌"
	if info.HasMQA {
		mqaIcon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s MQA (Master Quality Authenticated)\n", mqaIcon))

	dolbyIcon := "❌"
	if info.HasDolby {
		dolbyIcon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s Dolby Atmos\n", dolbyIcon))

	sony360Icon := "❌"
	if info.Has360 {
		sony360Icon = "✅"
	}
	sb.WriteString(fmt.Sprintf("- %s Sony 360 Reality Audio\n", sony360Icon))

	if info.BitDepth > 0 {
		sb.WriteString(fmt.Sprintf("\n**Bit Depth:** %d-bit\n", info.BitDepth))
	}
	if info.SampleRate > 0 {
		sb.WriteString(fmt.Sprintf("**Sample Rate:** %d Hz\n", info.SampleRate))
	}

	return tools.TextResult(sb.String()), nil
}

func handleSimilarArtists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	artists, err := client.GetSimilarArtists(ctx, int(artistID), limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get similar artists: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Similar Artists\n\n")

	if len(artists) == 0 {
		sb.WriteString("No similar artists found.\n")
	} else {
		sb.WriteString("| Artist | Popularity |\n")
		sb.WriteString("|--------|------------|\n")
		for _, a := range artists {
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %d |\n", a.Name, a.URL, a.Popularity))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleArtistTopTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	tracks, err := client.GetArtistTopTracks(ctx, int(artistID), limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get top tracks: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Artist Top Tracks\n\n")

	if len(tracks) == 0 {
		sb.WriteString("No tracks found.\n")
	} else {
		sb.WriteString("| # | Title | Album | Duration | Quality |\n")
		sb.WriteString("|---|-------|-------|----------|--------|\n")
		for i, t := range tracks {
			album := ""
			if t.Album != nil {
				album = t.Album.Title
			}
			duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s | %s |\n",
				i+1, t.Title, t.URL, album, duration, t.AudioQuality))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleArtistAlbums(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artistID, errResult := tools.RequireIntParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	albums, err := client.GetArtistAlbums(ctx, int(artistID), limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get albums: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Artist Albums\n\n")

	if len(albums) == 0 {
		sb.WriteString("No albums found.\n")
	} else {
		sb.WriteString("| Album | Release Date | Tracks | Quality |\n")
		sb.WriteString("|-------|--------------|--------|--------|\n")
		for _, a := range albums {
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %d | %s |\n",
				a.Title, a.URL, a.ReleaseDate, a.NumberOfTracks, a.AudioQuality))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleMixes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	mixes, err := client.GetMixes(ctx, limit)
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get mixes: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Curated Mixes\n\n")

	if len(mixes) == 0 {
		sb.WriteString("No mixes found.\n")
	} else {
		sb.WriteString("| Playlist | Tracks | Duration |\n")
		sb.WriteString("|----------|--------|----------|\n")
		for _, m := range mixes {
			duration := fmt.Sprintf("%d:%02d", m.Duration/60, m.Duration%60)
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %d | %s |\n",
				m.Title, m.URL, m.NumberOfTracks, duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleAlbumTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	albumID, errResult := tools.RequireIntParam(req, "album_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Tidal client error: %v", err)), nil
	}

	tracks, err := client.GetAlbumTracks(ctx, int(albumID))
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Failed to get album tracks: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Album Tracks\n\n")

	if len(tracks) == 0 {
		sb.WriteString("No tracks found.\n")
	} else {
		sb.WriteString("| # | Title | Duration | Quality |\n")
		sb.WriteString("|---|-------|----------|--------|\n")
		for _, t := range tracks {
			duration := fmt.Sprintf("%d:%02d", t.Duration/60, t.Duration%60)
			sb.WriteString(fmt.Sprintf("| %d | [%s](%s) | %s | %s |\n",
				t.TrackNumber, t.Title, t.URL, duration, t.AudioQuality))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
