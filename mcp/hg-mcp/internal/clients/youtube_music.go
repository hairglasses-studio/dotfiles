// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// YouTubeMusicClient provides access to YouTube Music via ytmusicapi
type YouTubeMusicClient struct {
	headersPath string // Path to headers_auth.json file
	httpClient  *http.Client
	mu          sync.RWMutex
}

// YouTubeMusicStatus represents API connection status
type YouTubeMusicStatus struct {
	Connected      bool   `json:"connected"`
	HasAuth        bool   `json:"has_auth"`
	AuthFile       string `json:"auth_file,omitempty"`
	YtDlpAvailable bool   `json:"ytdlp_available"`
}

// YouTubeMusicTrack represents a track
type YouTubeMusicTrack struct {
	VideoID     string   `json:"video_id"`
	Title       string   `json:"title"`
	Artists     []string `json:"artists"`
	Album       string   `json:"album,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	DurationSec int      `json:"duration_sec,omitempty"`
	Year        string   `json:"year,omitempty"`
	IsExplicit  bool     `json:"is_explicit"`
	Thumbnail   string   `json:"thumbnail,omitempty"`
	URL         string   `json:"url,omitempty"`
}

// YouTubeMusicArtist represents an artist
type YouTubeMusicArtist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Subscribers string `json:"subscribers,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	URL         string `json:"url,omitempty"`
}

// YouTubeMusicAlbum represents an album
type YouTubeMusicAlbum struct {
	BrowseID   string   `json:"browse_id"`
	Title      string   `json:"title"`
	Artists    []string `json:"artists"`
	Year       string   `json:"year,omitempty"`
	TrackCount int      `json:"track_count,omitempty"`
	Duration   string   `json:"duration,omitempty"`
	Thumbnail  string   `json:"thumbnail,omitempty"`
	URL        string   `json:"url,omitempty"`
	IsExplicit bool     `json:"is_explicit"`
}

// YouTubeMusicPlaylist represents a playlist
type YouTubeMusicPlaylist struct {
	PlaylistID  string `json:"playlist_id"`
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	TrackCount  int    `json:"track_count,omitempty"`
	Duration    string `json:"duration,omitempty"`
	Year        string `json:"year,omitempty"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// YouTubeMusicSearchResult represents search results
type YouTubeMusicSearchResult struct {
	Tracks    []YouTubeMusicTrack    `json:"tracks,omitempty"`
	Artists   []YouTubeMusicArtist   `json:"artists,omitempty"`
	Albums    []YouTubeMusicAlbum    `json:"albums,omitempty"`
	Playlists []YouTubeMusicPlaylist `json:"playlists,omitempty"`
	Videos    []YouTubeMusicTrack    `json:"videos,omitempty"`
}

// YouTubeMusicHealth represents health status
type YouTubeMusicHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasAuth         bool     `json:"has_auth"`
	YtDlpAvailable  bool     `json:"ytdlp_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// YouTubeMusicLyrics represents lyrics for a track
type YouTubeMusicLyrics struct {
	Source string `json:"source,omitempty"`
	Lyrics string `json:"lyrics,omitempty"`
}

// YouTubeMusicChartItem represents an item on a chart
type YouTubeMusicChartItem struct {
	Rank     int               `json:"rank"`
	Trend    string            `json:"trend,omitempty"` // UP, DOWN, SAME, NEW
	Track    YouTubeMusicTrack `json:"track"`
	Movement int               `json:"movement,omitempty"`
}

// YouTubeMusicChart represents a chart
type YouTubeMusicChart struct {
	Name    string                  `json:"name"`
	Country string                  `json:"country,omitempty"`
	Items   []YouTubeMusicChartItem `json:"items"`
}

const (
	ytMusicBaseURL = "https://music.youtube.com"
)

// NewYouTubeMusicClient creates a new YouTube Music client
func NewYouTubeMusicClient() (*YouTubeMusicClient, error) {
	headersPath := os.Getenv("YOUTUBE_MUSIC_HEADERS_PATH")
	if headersPath == "" {
		// Try default locations
		homeDir, _ := os.UserHomeDir()
		possiblePaths := []string{
			"headers_auth.json",
			homeDir + "/.config/ytmusicapi/headers_auth.json",
			homeDir + "/headers_auth.json",
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				headersPath = p
				break
			}
		}
	}

	return &YouTubeMusicClient{
		headersPath: headersPath,
		httpClient: httpclient.Standard(),
	}, nil
}

// checkYtDlp checks if yt-dlp is available
func (c *YouTubeMusicClient) checkYtDlp() bool {
	cmd := exec.Command("yt-dlp", "--version")
	return cmd.Run() == nil
}

// GetStatus returns connection status
func (c *YouTubeMusicClient) GetStatus(ctx context.Context) (*YouTubeMusicStatus, error) {
	status := &YouTubeMusicStatus{
		AuthFile:       c.headersPath,
		YtDlpAvailable: c.checkYtDlp(),
	}

	// Check if auth file exists
	if c.headersPath != "" {
		if _, err := os.Stat(c.headersPath); err == nil {
			status.HasAuth = true
			status.Connected = true
		}
	}

	// Without auth, we can still do some operations but not all
	if !status.HasAuth {
		status.Connected = true // Limited functionality without auth
	}

	return status, nil
}

// Health returns health check information
func (c *YouTubeMusicClient) Health(ctx context.Context) (*YouTubeMusicHealth, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return &YouTubeMusicHealth{
			Score:  0,
			Status: "error",
			Issues: []string{err.Error()},
		}, nil
	}

	health := &YouTubeMusicHealth{
		Connected:      status.Connected,
		HasAuth:        status.HasAuth,
		YtDlpAvailable: status.YtDlpAvailable,
	}

	health.Score = 50 // Base score for being accessible
	if status.HasAuth {
		health.Score += 30
	}
	if status.YtDlpAvailable {
		health.Score += 20
	}

	if health.Score >= 80 {
		health.Status = "healthy"
	} else if health.Score >= 50 {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	if !status.HasAuth {
		health.Issues = append(health.Issues, "No authentication headers configured")
		health.Recommendations = append(health.Recommendations, "Run ytmusicapi setup to create headers_auth.json")
	}
	if !status.YtDlpAvailable {
		health.Issues = append(health.Issues, "yt-dlp not available for downloads")
		health.Recommendations = append(health.Recommendations, "Install yt-dlp for download functionality")
	}

	return health, nil
}

// runYtDlp executes yt-dlp with given arguments and returns JSON output
func (c *YouTubeMusicClient) runYtDlp(ctx context.Context, args ...string) ([]byte, error) {
	fullArgs := append([]string{"--dump-json", "--no-download", "--flat-playlist"}, args...)
	cmd := exec.CommandContext(ctx, "yt-dlp", fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp error: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// Search searches YouTube Music
func (c *YouTubeMusicClient) Search(ctx context.Context, query string, filter string, limit int) (*YouTubeMusicSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	// Use yt-dlp to search YouTube Music
	searchURL := fmt.Sprintf("ytsearch%d:%s", limit, query)
	if filter != "" {
		// Modify search based on filter
		switch strings.ToLower(filter) {
		case "songs", "tracks":
			searchURL = fmt.Sprintf("https://music.youtube.com/search?q=%s#songs", query)
		case "albums":
			searchURL = fmt.Sprintf("https://music.youtube.com/search?q=%s#albums", query)
		case "artists":
			searchURL = fmt.Sprintf("https://music.youtube.com/search?q=%s#artists", query)
		case "playlists":
			searchURL = fmt.Sprintf("https://music.youtube.com/search?q=%s#community_playlists", query)
		}
	}

	output, err := c.runYtDlp(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	result := &YouTubeMusicSearchResult{}

	// Parse JSON lines output (each line is a separate JSON object)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID          string `json:"id"`
			Title       string `json:"title"`
			Uploader    string `json:"uploader"`
			Channel     string `json:"channel"`
			Duration    int    `json:"duration"`
			Description string `json:"description"`
			Thumbnail   string `json:"thumbnail"`
			ViewCount   int    `json:"view_count"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		track := YouTubeMusicTrack{
			VideoID:     item.ID,
			Title:       item.Title,
			Artists:     []string{item.Uploader},
			DurationSec: item.Duration,
			Duration:    formatDuration(item.Duration),
			Thumbnail:   item.Thumbnail,
			URL:         fmt.Sprintf("%s/watch?v=%s", ytMusicBaseURL, item.ID),
		}

		result.Tracks = append(result.Tracks, track)
	}

	return result, nil
}

// GetTrack retrieves information about a specific track/video
func (c *YouTubeMusicClient) GetTrack(ctx context.Context, videoID string) (*YouTubeMusicTrack, error) {
	url := fmt.Sprintf("https://music.youtube.com/watch?v=%s", videoID)

	output, err := c.runYtDlp(ctx, url)
	if err != nil {
		return nil, err
	}

	var item struct {
		ID          string   `json:"id"`
		Title       string   `json:"title"`
		Artist      string   `json:"artist"`
		Uploader    string   `json:"uploader"`
		Album       string   `json:"album"`
		Duration    int      `json:"duration"`
		ReleaseYear int      `json:"release_year"`
		Thumbnail   string   `json:"thumbnail"`
		Categories  []string `json:"categories"`
	}

	if err := json.Unmarshal(output, &item); err != nil {
		return nil, fmt.Errorf("failed to parse track info: %w", err)
	}

	artists := []string{}
	if item.Artist != "" {
		artists = append(artists, item.Artist)
	} else if item.Uploader != "" {
		artists = append(artists, item.Uploader)
	}

	year := ""
	if item.ReleaseYear > 0 {
		year = strconv.Itoa(item.ReleaseYear)
	}

	return &YouTubeMusicTrack{
		VideoID:     item.ID,
		Title:       item.Title,
		Artists:     artists,
		Album:       item.Album,
		DurationSec: item.Duration,
		Duration:    formatDuration(item.Duration),
		Year:        year,
		Thumbnail:   item.Thumbnail,
		URL:         url,
	}, nil
}

// GetAlbum retrieves information about an album
func (c *YouTubeMusicClient) GetAlbum(ctx context.Context, browseID string) (*YouTubeMusicAlbum, []YouTubeMusicTrack, error) {
	url := fmt.Sprintf("https://music.youtube.com/playlist?list=%s", browseID)

	output, err := c.runYtDlp(ctx, "--flat-playlist", url)
	if err != nil {
		return nil, nil, err
	}

	var tracks []YouTubeMusicTrack
	var albumTitle, albumArtist string
	var totalDuration int

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			Uploader      string `json:"uploader"`
			Duration      int    `json:"duration"`
			PlaylistTitle string `json:"playlist_title"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		if albumTitle == "" && item.PlaylistTitle != "" {
			albumTitle = item.PlaylistTitle
		}
		if albumArtist == "" && item.Uploader != "" {
			albumArtist = item.Uploader
		}

		totalDuration += item.Duration

		tracks = append(tracks, YouTubeMusicTrack{
			VideoID:     item.ID,
			Title:       item.Title,
			Artists:     []string{item.Uploader},
			DurationSec: item.Duration,
			Duration:    formatDuration(item.Duration),
			URL:         fmt.Sprintf("%s/watch?v=%s", ytMusicBaseURL, item.ID),
		})
	}

	album := &YouTubeMusicAlbum{
		BrowseID:   browseID,
		Title:      albumTitle,
		Artists:    []string{albumArtist},
		TrackCount: len(tracks),
		Duration:   formatDuration(totalDuration),
		URL:        url,
	}

	return album, tracks, nil
}

// GetPlaylist retrieves information about a playlist
func (c *YouTubeMusicClient) GetPlaylist(ctx context.Context, playlistID string) (*YouTubeMusicPlaylist, []YouTubeMusicTrack, error) {
	url := fmt.Sprintf("https://music.youtube.com/playlist?list=%s", playlistID)

	output, err := c.runYtDlp(ctx, "--flat-playlist", url)
	if err != nil {
		return nil, nil, err
	}

	var tracks []YouTubeMusicTrack
	var playlistTitle, playlistAuthor string
	var totalDuration int

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID               string `json:"id"`
			Title            string `json:"title"`
			Uploader         string `json:"uploader"`
			Duration         int    `json:"duration"`
			PlaylistTitle    string `json:"playlist_title"`
			PlaylistUploader string `json:"playlist_uploader"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		if playlistTitle == "" && item.PlaylistTitle != "" {
			playlistTitle = item.PlaylistTitle
		}
		if playlistAuthor == "" && item.PlaylistUploader != "" {
			playlistAuthor = item.PlaylistUploader
		}

		totalDuration += item.Duration

		tracks = append(tracks, YouTubeMusicTrack{
			VideoID:     item.ID,
			Title:       item.Title,
			Artists:     []string{item.Uploader},
			DurationSec: item.Duration,
			Duration:    formatDuration(item.Duration),
			URL:         fmt.Sprintf("%s/watch?v=%s", ytMusicBaseURL, item.ID),
		})
	}

	playlist := &YouTubeMusicPlaylist{
		PlaylistID: playlistID,
		Title:      playlistTitle,
		Author:     playlistAuthor,
		TrackCount: len(tracks),
		Duration:   formatDuration(totalDuration),
		URL:        url,
	}

	return playlist, tracks, nil
}

// GetArtist retrieves information about an artist
func (c *YouTubeMusicClient) GetArtist(ctx context.Context, channelID string) (*YouTubeMusicArtist, error) {
	url := fmt.Sprintf("https://music.youtube.com/channel/%s", channelID)

	// For artist pages, we'll use a simpler approach
	artist := &YouTubeMusicArtist{
		ID:  channelID,
		URL: url,
	}

	// Try to get artist info via yt-dlp
	output, err := c.runYtDlp(ctx, "--playlist-items", "1", url)
	if err == nil {
		var item struct {
			Channel   string `json:"channel"`
			ChannelID string `json:"channel_id"`
		}
		if json.Unmarshal(output, &item) == nil {
			artist.Name = item.Channel
		}
	}

	return artist, nil
}

// GetCharts retrieves music charts
func (c *YouTubeMusicClient) GetCharts(ctx context.Context, country string, limit int) (*YouTubeMusicChart, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Use YouTube Music charts URL
	url := "https://music.youtube.com/charts"
	if country != "" {
		url = fmt.Sprintf("%s?bp=%s", url, country)
	}

	output, err := c.runYtDlp(ctx, "--flat-playlist", fmt.Sprintf("--playlist-items=1-%d", limit), url)
	if err != nil {
		return nil, err
	}

	chart := &YouTubeMusicChart{
		Name:    "Top Songs",
		Country: country,
	}

	lines := strings.Split(string(output), "\n")
	rank := 1
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Uploader string `json:"uploader"`
			Duration int    `json:"duration"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		chart.Items = append(chart.Items, YouTubeMusicChartItem{
			Rank: rank,
			Track: YouTubeMusicTrack{
				VideoID:     item.ID,
				Title:       item.Title,
				Artists:     []string{item.Uploader},
				DurationSec: item.Duration,
				Duration:    formatDuration(item.Duration),
				URL:         fmt.Sprintf("%s/watch?v=%s", ytMusicBaseURL, item.ID),
			},
		})
		rank++
	}

	return chart, nil
}

// GetNewReleases retrieves new music releases
func (c *YouTubeMusicClient) GetNewReleases(ctx context.Context, limit int) ([]YouTubeMusicAlbum, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	// Use YouTube Music new releases explore page
	url := "https://music.youtube.com/explore/new_releases"

	output, err := c.runYtDlp(ctx, "--flat-playlist", fmt.Sprintf("--playlist-items=1-%d", limit), url)
	if err != nil {
		return nil, err
	}

	var albums []YouTubeMusicAlbum

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Uploader string `json:"uploader"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		albums = append(albums, YouTubeMusicAlbum{
			BrowseID: item.ID,
			Title:    item.Title,
			Artists:  []string{item.Uploader},
			URL:      fmt.Sprintf("%s/playlist?list=%s", ytMusicBaseURL, item.ID),
		})
	}

	return albums, nil
}

// GetMoods retrieves available moods/genres for browsing
func (c *YouTubeMusicClient) GetMoods(ctx context.Context) ([]string, error) {
	// Return a static list of common moods/genres available on YouTube Music
	return []string{
		"Chill", "Party", "Focus", "Workout", "Commute",
		"Romance", "Sad", "Feel Good", "Sleep", "Energize",
		"Pop", "Hip-Hop", "R&B", "Rock", "Electronic",
		"Country", "Latin", "K-Pop", "Indie", "Jazz",
		"Classical", "Metal", "Alternative", "Reggae", "Blues",
	}, nil
}

// GetRadio generates a radio mix based on a track
func (c *YouTubeMusicClient) GetRadio(ctx context.Context, videoID string, limit int) ([]YouTubeMusicTrack, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	// YouTube Music's radio feature via mix playlist
	url := fmt.Sprintf("https://music.youtube.com/watch?v=%s&list=RDAMVM%s", videoID, videoID)

	output, err := c.runYtDlp(ctx, "--flat-playlist", fmt.Sprintf("--playlist-items=1-%d", limit), url)
	if err != nil {
		return nil, err
	}

	var tracks []YouTubeMusicTrack

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var item struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Uploader string `json:"uploader"`
			Duration int    `json:"duration"`
		}

		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}

		tracks = append(tracks, YouTubeMusicTrack{
			VideoID:     item.ID,
			Title:       item.Title,
			Artists:     []string{item.Uploader},
			DurationSec: item.Duration,
			Duration:    formatDuration(item.Duration),
			URL:         fmt.Sprintf("%s/watch?v=%s", ytMusicBaseURL, item.ID),
		})
	}

	return tracks, nil
}

// ExtractVideoID extracts video ID from various URL formats
func ExtractVideoID(input string) string {
	// Already an ID (11 characters)
	if len(input) == 11 && !strings.Contains(input, "/") {
		return input
	}

	// YouTube Music URL patterns
	patterns := []string{
		`youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`,
		`youtu\.be/([a-zA-Z0-9_-]{11})`,
		`music\.youtube\.com/watch\?v=([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(input); len(matches) > 1 {
			return matches[1]
		}
	}

	return input
}

// formatDuration formats seconds into MM:SS or HH:MM:SS
func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "0:00"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}
