// Package spotify provides MCP tools for Spotify Web API integration.
package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the Spotify tools module
type Module struct{}

var getSpotifyClient = tools.LazyClient(clients.GetSpotifyClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "spotify"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Spotify Web API integration for music discovery and track metadata"
}

// Tools returns all Spotify tools
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_spotify_status",
				mcp.WithDescription("Get Spotify API connection status and token info"),
			),
			Handler:             handleSpotifyStatus,
			Category:            "spotify",
			Subcategory:         "status",
			Tags:                []string{"spotify", "music", "status", "api"},
			UseCases:            []string{"Check Spotify API connectivity", "Verify authentication"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_search",
				mcp.WithDescription("Search Spotify for tracks, artists, or albums"),
				mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
				mcp.WithString("type", mcp.Description("Search type: track, artist, album, or comma-separated (default: track)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 20, max: 50)")),
			),
			Handler:             handleSpotifySearch,
			Category:            "spotify",
			Subcategory:         "search",
			Tags:                []string{"spotify", "music", "search", "discovery"},
			UseCases:            []string{"Find tracks by name", "Search for artists", "Discover new music"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_track",
				mcp.WithDescription("Get detailed information about a Spotify track"),
				mcp.WithString("track_id", mcp.Description("Spotify track ID"), mcp.Required()),
			),
			Handler:             handleSpotifyTrack,
			Category:            "spotify",
			Subcategory:         "tracks",
			Tags:                []string{"spotify", "music", "track", "metadata"},
			UseCases:            []string{"Get track details", "View track metadata"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_features",
				mcp.WithDescription("Get audio features for a track (BPM, key, energy, danceability)"),
				mcp.WithString("track_id", mcp.Description("Spotify track ID"), mcp.Required()),
			),
			Handler:             handleSpotifyFeatures,
			Category:            "spotify",
			Subcategory:         "analysis",
			Tags:                []string{"spotify", "music", "audio-features", "bpm", "key"},
			UseCases:            []string{"Get BPM and key", "Analyze track mood", "DJ preparation"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_artist",
				mcp.WithDescription("Get detailed information about a Spotify artist"),
				mcp.WithString("artist_id", mcp.Description("Spotify artist ID"), mcp.Required()),
			),
			Handler:             handleSpotifyArtist,
			Category:            "spotify",
			Subcategory:         "artists",
			Tags:                []string{"spotify", "music", "artist", "metadata"},
			UseCases:            []string{"Get artist details", "View artist genres"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_artist_top",
				mcp.WithDescription("Get an artist's top tracks"),
				mcp.WithString("artist_id", mcp.Description("Spotify artist ID"), mcp.Required()),
			),
			Handler:             handleSpotifyArtistTop,
			Category:            "spotify",
			Subcategory:         "artists",
			Tags:                []string{"spotify", "music", "artist", "top-tracks"},
			UseCases:            []string{"Discover artist's best tracks", "Build playlists"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_related",
				mcp.WithDescription("Get artists related to a given artist"),
				mcp.WithString("artist_id", mcp.Description("Spotify artist ID"), mcp.Required()),
			),
			Handler:             handleSpotifyRelated,
			Category:            "spotify",
			Subcategory:         "discovery",
			Tags:                []string{"spotify", "music", "discovery", "related-artists"},
			UseCases:            []string{"Discover similar artists", "Expand music library"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_album",
				mcp.WithDescription("Get detailed information about a Spotify album"),
				mcp.WithString("album_id", mcp.Description("Spotify album ID"), mcp.Required()),
			),
			Handler:             handleSpotifyAlbum,
			Category:            "spotify",
			Subcategory:         "albums",
			Tags:                []string{"spotify", "music", "album", "metadata"},
			UseCases:            []string{"Get album details", "View album artwork"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_new_releases",
				mcp.WithDescription("Get new album releases"),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 20, max: 50)")),
			),
			Handler:             handleSpotifyNewReleases,
			Category:            "spotify",
			Subcategory:         "discovery",
			Tags:                []string{"spotify", "music", "new-releases", "discovery"},
			UseCases:            []string{"Discover new music", "Stay current with releases"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_recommendations",
				mcp.WithDescription("Get track recommendations based on seed tracks or artists"),
				mcp.WithString("seed_tracks", mcp.Description("Comma-separated Spotify track IDs (up to 5)")),
				mcp.WithString("seed_artists", mcp.Description("Comma-separated Spotify artist IDs (up to 5)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 20, max: 100)")),
			),
			Handler:             handleSpotifyRecommendations,
			Category:            "spotify",
			Subcategory:         "discovery",
			Tags:                []string{"spotify", "music", "recommendations", "discovery"},
			UseCases:            []string{"Get similar tracks", "Build playlists", "Discover new music"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_genres",
				mcp.WithDescription("Get available genre seeds for recommendations"),
			),
			Handler:             handleSpotifyGenres,
			Category:            "spotify",
			Subcategory:         "discovery",
			Tags:                []string{"spotify", "music", "genres", "discovery"},
			UseCases:            []string{"List available genres", "Find genre categories"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_track_analysis",
				mcp.WithDescription("Get combined track details and audio features in one call"),
				mcp.WithString("track_id", mcp.Description("Spotify track ID"), mcp.Required()),
			),
			Handler:             handleSpotifyTrackAnalysis,
			Category:            "spotify",
			Subcategory:         "analysis",
			Tags:                []string{"spotify", "music", "analysis", "bpm", "key"},
			UseCases:            []string{"Full track analysis", "DJ preparation", "Sync visuals to music"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_key",
				mcp.WithDescription("Convert Spotify key number to musical notation"),
				mcp.WithNumber("key", mcp.Description("Spotify key number (0-11)"), mcp.Required()),
				mcp.WithNumber("mode", mcp.Description("Mode (0 = minor, 1 = major)"), mcp.Required()),
			),
			Handler:             handleSpotifyKey,
			Category:            "spotify",
			Subcategory:         "utility",
			Tags:                []string{"spotify", "music", "key", "utility"},
			UseCases:            []string{"Convert key notation", "DJ preparation"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_health",
				mcp.WithDescription("Check Spotify API health and configuration"),
			),
			Handler:             handleSpotifyHealth,
			Category:            "spotify",
			Subcategory:         "status",
			Tags:                []string{"spotify", "health", "diagnostics"},
			UseCases:            []string{"Diagnose API issues", "Check configuration"},
			Complexity:          "simple",
			CircuitBreakerGroup: "spotify",
		},
	}
	// Add pipeline tools
	allTools = append(allTools, pipelineTools()...)
	return allTools
}

func handleSpotifyStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleSpotifySearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	typeStr := tools.GetStringParam(req, "type")
	var types []string
	if typeStr != "" {
		types = strings.Split(typeStr, ",")
	}

	limit := tools.GetIntParam(req, "limit", 20)

	results, err := client.Search(ctx, query, types, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

func handleSpotifyTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	trackID, errResult := tools.RequireStringParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get track: %w", err)), nil
	}

	return tools.JSONResult(track), nil
}

func handleSpotifyFeatures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	trackID, errResult := tools.RequireStringParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	features, err := client.GetAudioFeatures(ctx, trackID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get audio features: %w", err)), nil
	}

	// Add human-readable key
	result := map[string]any{
		"id":               features.ID,
		"tempo":            features.Tempo,
		"key":              features.Key,
		"key_name":         keyToName(features.Key, features.Mode),
		"mode":             features.Mode,
		"mode_name":        modeName(features.Mode),
		"time_signature":   features.TimeSignature,
		"danceability":     features.Danceability,
		"energy":           features.Energy,
		"valence":          features.Valence,
		"loudness":         features.Loudness,
		"speechiness":      features.Speechiness,
		"acousticness":     features.Acousticness,
		"instrumentalness": features.Instrumentalness,
		"liveness":         features.Liveness,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	artistID, errResult := tools.RequireStringParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	artist, err := client.GetArtist(ctx, artistID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get artist: %w", err)), nil
	}

	return tools.JSONResult(artist), nil
}

func handleSpotifyArtistTop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	artistID, errResult := tools.RequireStringParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	tracks, err := client.GetArtistTopTracks(ctx, artistID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get top tracks: %w", err)), nil
	}

	result := map[string]any{
		"artist_id": artistID,
		"count":     len(tracks),
		"tracks":    tracks,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyRelated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	artistID, errResult := tools.RequireStringParam(req, "artist_id")
	if errResult != nil {
		return errResult, nil
	}

	artists, err := client.GetRelatedArtists(ctx, artistID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get related artists: %w", err)), nil
	}

	result := map[string]any{
		"artist_id": artistID,
		"count":     len(artists),
		"artists":   artists,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	albumID, errResult := tools.RequireStringParam(req, "album_id")
	if errResult != nil {
		return errResult, nil
	}

	album, err := client.GetAlbum(ctx, albumID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get album: %w", err)), nil
	}

	return tools.JSONResult(album), nil
}

func handleSpotifyNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	albums, err := client.GetNewReleases(ctx, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get new releases: %w", err)), nil
	}

	result := map[string]any{
		"count":  len(albums),
		"albums": albums,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyRecommendations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	seedTracksStr := tools.GetStringParam(req, "seed_tracks")
	seedArtistsStr := tools.GetStringParam(req, "seed_artists")
	limit := tools.GetIntParam(req, "limit", 20)

	var seedTracks, seedArtists []string
	if seedTracksStr != "" {
		seedTracks = strings.Split(seedTracksStr, ",")
	}
	if seedArtistsStr != "" {
		seedArtists = strings.Split(seedArtistsStr, ",")
	}

	if len(seedTracks) == 0 && len(seedArtists) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("at least one seed_tracks or seed_artists is required")), nil
	}

	tracks, err := client.GetRecommendations(ctx, seedTracks, seedArtists, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get recommendations: %w", err)), nil
	}

	result := map[string]any{
		"seed_tracks":  seedTracks,
		"seed_artists": seedArtists,
		"count":        len(tracks),
		"tracks":       tracks,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyGenres(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	genres, err := client.GetGenres(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get genres: %w", err)), nil
	}

	result := map[string]any{
		"count":  len(genres),
		"genres": genres,
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyTrackAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Spotify client: %w", err)), nil
	}

	trackID, errResult := tools.RequireStringParam(req, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get track: %w", err)), nil
	}

	features, err := client.GetAudioFeatures(ctx, trackID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get audio features: %w", err)), nil
	}

	// Format duration
	durationSec := track.DurationMS / 1000
	durationStr := fmt.Sprintf("%d:%02d", durationSec/60, durationSec%60)

	result := map[string]any{
		"track": map[string]any{
			"id":          track.ID,
			"name":        track.Name,
			"artists":     track.Artists,
			"album":       track.Album,
			"duration":    durationStr,
			"duration_ms": track.DurationMS,
			"popularity":  track.Popularity,
			"preview_url": track.PreviewURL,
			"uri":         track.URI,
		},
		"audio_features": map[string]any{
			"bpm":              features.Tempo,
			"key":              keyToName(features.Key, features.Mode),
			"key_number":       features.Key,
			"mode":             modeName(features.Mode),
			"time_signature":   features.TimeSignature,
			"danceability":     features.Danceability,
			"energy":           features.Energy,
			"valence":          features.Valence,
			"loudness":         features.Loudness,
			"speechiness":      features.Speechiness,
			"acousticness":     features.Acousticness,
			"instrumentalness": features.Instrumentalness,
			"liveness":         features.Liveness,
		},
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyKey(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key := tools.GetIntParam(req, "key", -1)
	mode := tools.GetIntParam(req, "mode", -1)

	if key < 0 || key > 11 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("key must be 0-11")), nil
	}

	if mode < 0 || mode > 1 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("mode must be 0 (minor) or 1 (major)")), nil
	}

	result := map[string]any{
		"key":       key,
		"mode":      mode,
		"key_name":  keyToName(key, mode),
		"mode_name": modeName(mode),
		"camelot":   keyToCamelot(key, mode),
	}

	return tools.JSONResult(result), nil
}

func handleSpotifyHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSpotifyClient()
	if err != nil {
		// Return health info even if client creation fails
		health := &clients.SpotifyHealth{
			Score:          50,
			Status:         "degraded",
			Connected:      false,
			HasCredentials: false,
			Issues:         []string{err.Error()},
			Recommendations: []string{
				"Set SPOTIFY_CLIENT_ID environment variable",
				"Set SPOTIFY_CLIENT_SECRET environment variable",
			},
		}
		return tools.JSONResult(health), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

// keyToName converts Spotify key number to musical notation
func keyToName(key, mode int) string {
	keys := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	if key < 0 || key > 11 {
		return "Unknown"
	}
	suffix := "m"
	if mode == 1 {
		suffix = ""
	}
	return keys[key] + suffix
}

// modeName returns the mode name
func modeName(mode int) string {
	if mode == 1 {
		return "major"
	}
	return "minor"
}

// keyToCamelot converts Spotify key to Camelot notation (used by DJs)
func keyToCamelot(key, mode int) string {
	// Major keys
	majorCamelot := []string{"8B", "3B", "10B", "5B", "12B", "7B", "2B", "9B", "4B", "11B", "6B", "1B"}
	// Minor keys
	minorCamelot := []string{"5A", "12A", "7A", "2A", "9A", "4A", "11A", "6A", "1A", "8A", "3A", "10A"}

	if key < 0 || key > 11 {
		return "Unknown"
	}

	if mode == 1 {
		return majorCamelot[key]
	}
	return minorCamelot[key]
}
