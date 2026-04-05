// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const traxsourceBaseURL = "https://www.traxsource.com"

// TraxsourceClient provides access to Traxsource electronic music store
type TraxsourceClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
}

// TraxsourceTrack represents a track on Traxsource
type TraxsourceTrack struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Mix         string   `json:"mix,omitempty"`
	Artists     []string `json:"artists"`
	Remixers    []string `json:"remixers,omitempty"`
	Label       string   `json:"label"`
	Genre       string   `json:"genre"`
	BPM         int      `json:"bpm,omitempty"`
	Key         string   `json:"key,omitempty"`
	Duration    string   `json:"duration,omitempty"`
	ReleaseDate string   `json:"release_date,omitempty"`
	Price       string   `json:"price,omitempty"`
	PreviewURL  string   `json:"preview_url,omitempty"`
	URL         string   `json:"url"`
	ImageURL    string   `json:"image_url,omitempty"`
	IsExclusive bool     `json:"is_exclusive,omitempty"`
}

// TraxsourceRelease represents a release/EP on Traxsource
type TraxsourceRelease struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Artists     []string          `json:"artists"`
	Label       string            `json:"label"`
	Genre       string            `json:"genre"`
	CatalogNum  string            `json:"catalog_number,omitempty"`
	ReleaseDate string            `json:"release_date,omitempty"`
	Tracks      []TraxsourceTrack `json:"tracks,omitempty"`
	URL         string            `json:"url"`
	ImageURL    string            `json:"image_url,omitempty"`
	Price       string            `json:"price,omitempty"`
}

// TraxsourceArtist represents an artist on Traxsource
type TraxsourceArtist struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	ImageURL   string   `json:"image_url,omitempty"`
	Bio        string   `json:"bio,omitempty"`
	Genres     []string `json:"genres,omitempty"`
	TrackCount int      `json:"track_count,omitempty"`
}

// TraxsourceLabel represents a record label on Traxsource
type TraxsourceLabel struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	ImageURL     string   `json:"image_url,omitempty"`
	Description  string   `json:"description,omitempty"`
	Genres       []string `json:"genres,omitempty"`
	ReleaseCount int      `json:"release_count,omitempty"`
}

// TraxsourceChart represents a DJ chart
type TraxsourceChart struct {
	ID       int               `json:"id"`
	Title    string            `json:"title"`
	DJ       string            `json:"dj"`
	Date     string            `json:"date,omitempty"`
	Genre    string            `json:"genre,omitempty"`
	Tracks   []TraxsourceTrack `json:"tracks,omitempty"`
	URL      string            `json:"url"`
	ImageURL string            `json:"image_url,omitempty"`
}

// TraxsourceSearchResults contains search results
type TraxsourceSearchResults struct {
	Tracks   []TraxsourceTrack   `json:"tracks,omitempty"`
	Releases []TraxsourceRelease `json:"releases,omitempty"`
	Artists  []TraxsourceArtist  `json:"artists,omitempty"`
	Labels   []TraxsourceLabel   `json:"labels,omitempty"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PerPage  int                 `json:"per_page"`
}

// TraxsourceStatus represents connection status
type TraxsourceStatus struct {
	Available      bool   `json:"available"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	YtDlpAvailable bool   `json:"yt_dlp_available"`
	Message        string `json:"message,omitempty"`
}

// TraxsourceHealth represents health check results
type TraxsourceHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	SiteReachable   bool     `json:"site_reachable"`
	YtDlpAvailable  bool     `json:"yt_dlp_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// TraxsourceGenre represents a music genre
type TraxsourceGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

// NewTraxsourceClient creates a new Traxsource client
func NewTraxsourceClient() (*TraxsourceClient, error) {
	return &TraxsourceClient{
		httpClient: httpclient.Standard(),
	}, nil
}

// GetStatus returns the current connection status
func (c *TraxsourceClient) GetStatus(ctx context.Context) (*TraxsourceStatus, error) {
	status := &TraxsourceStatus{
		Available: false,
	}

	// Check site availability
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", traxsourceBaseURL, nil)
	if err != nil {
		status.Message = fmt.Sprintf("Failed to create request: %v", err)
		return status, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.Message = fmt.Sprintf("Site unreachable: %v", err)
		return status, nil
	}
	defer resp.Body.Close()

	status.ResponseTimeMs = time.Since(start).Milliseconds()
	status.Available = resp.StatusCode == 200

	// Check yt-dlp availability
	if _, err := exec.LookPath("yt-dlp"); err == nil {
		status.YtDlpAvailable = true
	}

	if status.Available {
		status.Message = "Traxsource is accessible"
	} else {
		status.Message = fmt.Sprintf("Site returned status %d", resp.StatusCode)
	}

	return status, nil
}

// GetHealth performs a health check
func (c *TraxsourceClient) GetHealth(ctx context.Context) (*TraxsourceHealth, error) {
	health := &TraxsourceHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check site reachability
	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Score = 0
		health.Status = "critical"
		health.Issues = append(health.Issues, fmt.Sprintf("Status check failed: %v", err))
		return health, nil
	}

	health.SiteReachable = status.Available
	health.YtDlpAvailable = status.YtDlpAvailable

	if !status.Available {
		health.Score -= 50
		health.Issues = append(health.Issues, "Traxsource site is not reachable")
		health.Recommendations = append(health.Recommendations, "Check internet connectivity")
	}

	if !status.YtDlpAvailable {
		health.Score -= 30
		health.Issues = append(health.Issues, "yt-dlp not available for preview downloads")
		health.Recommendations = append(health.Recommendations, "Install yt-dlp: pip install yt-dlp")
	}

	if health.Score >= 80 {
		health.Status = "healthy"
	} else if health.Score >= 50 {
		health.Status = "degraded"
	} else {
		health.Status = "critical"
	}

	return health, nil
}

// Search searches for tracks, releases, artists, or labels
func (c *TraxsourceClient) Search(ctx context.Context, query string, searchType string, page int, perPage int) (*TraxsourceSearchResults, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	// Build search URL based on type
	var searchURL string
	switch searchType {
	case "track", "tracks":
		searchURL = fmt.Sprintf("%s/search/tracks?q=%s&page=%d", traxsourceBaseURL, url.QueryEscape(query), page)
	case "release", "releases":
		searchURL = fmt.Sprintf("%s/search/releases?q=%s&page=%d", traxsourceBaseURL, url.QueryEscape(query), page)
	case "artist", "artists":
		searchURL = fmt.Sprintf("%s/search/artists?q=%s&page=%d", traxsourceBaseURL, url.QueryEscape(query), page)
	case "label", "labels":
		searchURL = fmt.Sprintf("%s/search/labels?q=%s&page=%d", traxsourceBaseURL, url.QueryEscape(query), page)
	default:
		searchURL = fmt.Sprintf("%s/search/tracks?q=%s&page=%d", traxsourceBaseURL, url.QueryEscape(query), page)
	}

	body, err := c.fetchPage(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results := &TraxsourceSearchResults{
		Page:    page,
		PerPage: perPage,
	}

	// Parse results based on type
	switch searchType {
	case "track", "tracks", "":
		results.Tracks = c.parseTracksFromHTML(body)
		results.Total = len(results.Tracks)
	case "release", "releases":
		results.Releases = c.parseReleasesFromHTML(body)
		results.Total = len(results.Releases)
	case "artist", "artists":
		results.Artists = c.parseArtistsFromHTML(body)
		results.Total = len(results.Artists)
	case "label", "labels":
		results.Labels = c.parseLabelsFromHTML(body)
		results.Total = len(results.Labels)
	}

	return results, nil
}

// GetTrack gets detailed information about a specific track
func (c *TraxsourceClient) GetTrack(ctx context.Context, trackID int) (*TraxsourceTrack, error) {
	trackURL := fmt.Sprintf("%s/track/%d", traxsourceBaseURL, trackID)
	body, err := c.fetchPage(ctx, trackURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch track: %w", err)
	}

	track := c.parseTrackDetailsFromHTML(body, trackID)
	if track == nil {
		return nil, fmt.Errorf("track not found")
	}

	return track, nil
}

// GetRelease gets detailed information about a release
func (c *TraxsourceClient) GetRelease(ctx context.Context, releaseID int) (*TraxsourceRelease, error) {
	releaseURL := fmt.Sprintf("%s/title/%d", traxsourceBaseURL, releaseID)
	body, err := c.fetchPage(ctx, releaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}

	release := c.parseReleaseDetailsFromHTML(body, releaseID)
	if release == nil {
		return nil, fmt.Errorf("release not found")
	}

	return release, nil
}

// GetArtist gets information about an artist
func (c *TraxsourceClient) GetArtist(ctx context.Context, artistID int) (*TraxsourceArtist, error) {
	artistURL := fmt.Sprintf("%s/artist/%d", traxsourceBaseURL, artistID)
	body, err := c.fetchPage(ctx, artistURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist: %w", err)
	}

	artist := c.parseArtistDetailsFromHTML(body, artistID)
	if artist == nil {
		return nil, fmt.Errorf("artist not found")
	}

	return artist, nil
}

// GetArtistTracks gets tracks by an artist
func (c *TraxsourceClient) GetArtistTracks(ctx context.Context, artistID int, page int) ([]TraxsourceTrack, error) {
	if page < 1 {
		page = 1
	}
	tracksURL := fmt.Sprintf("%s/artist/%d/tracks?page=%d", traxsourceBaseURL, artistID, page)
	body, err := c.fetchPage(ctx, tracksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist tracks: %w", err)
	}

	return c.parseTracksFromHTML(body), nil
}

// GetLabel gets information about a record label
func (c *TraxsourceClient) GetLabel(ctx context.Context, labelID int) (*TraxsourceLabel, error) {
	labelURL := fmt.Sprintf("%s/label/%d", traxsourceBaseURL, labelID)
	body, err := c.fetchPage(ctx, labelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch label: %w", err)
	}

	label := c.parseLabelDetailsFromHTML(body, labelID)
	if label == nil {
		return nil, fmt.Errorf("label not found")
	}

	return label, nil
}

// GetLabelReleases gets releases from a label
func (c *TraxsourceClient) GetLabelReleases(ctx context.Context, labelID int, page int) ([]TraxsourceRelease, error) {
	if page < 1 {
		page = 1
	}
	releasesURL := fmt.Sprintf("%s/label/%d/releases?page=%d", traxsourceBaseURL, labelID, page)
	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch label releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetTopCharts gets current top charts
func (c *TraxsourceClient) GetTopCharts(ctx context.Context, genre string, page int) ([]TraxsourceTrack, error) {
	if page < 1 {
		page = 1
	}

	var chartsURL string
	if genre != "" {
		chartsURL = fmt.Sprintf("%s/genre/%s/top?page=%d", traxsourceBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		chartsURL = fmt.Sprintf("%s/top?page=%d", traxsourceBaseURL, page)
	}

	body, err := c.fetchPage(ctx, chartsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch charts: %w", err)
	}

	return c.parseTracksFromHTML(body), nil
}

// GetDJCharts gets DJ charts
func (c *TraxsourceClient) GetDJCharts(ctx context.Context, page int) ([]TraxsourceChart, error) {
	if page < 1 {
		page = 1
	}
	chartsURL := fmt.Sprintf("%s/dj-charts?page=%d", traxsourceBaseURL, page)
	body, err := c.fetchPage(ctx, chartsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DJ charts: %w", err)
	}

	return c.parseChartsFromHTML(body), nil
}

// GetGenres returns available genres
func (c *TraxsourceClient) GetGenres(ctx context.Context) ([]TraxsourceGenre, error) {
	// Traxsource genres are relatively static
	genres := []TraxsourceGenre{
		{ID: 1, Name: "House", Slug: "house", URL: traxsourceBaseURL + "/genre/house"},
		{ID: 2, Name: "Deep House", Slug: "deep-house", URL: traxsourceBaseURL + "/genre/deep-house"},
		{ID: 3, Name: "Tech House", Slug: "tech-house", URL: traxsourceBaseURL + "/genre/tech-house"},
		{ID: 4, Name: "Techno", Slug: "techno", URL: traxsourceBaseURL + "/genre/techno"},
		{ID: 5, Name: "Afro House", Slug: "afro-house", URL: traxsourceBaseURL + "/genre/afro-house"},
		{ID: 6, Name: "Melodic House & Techno", Slug: "melodic-house-techno", URL: traxsourceBaseURL + "/genre/melodic-house-techno"},
		{ID: 7, Name: "Nu Disco / Disco", Slug: "nu-disco-disco", URL: traxsourceBaseURL + "/genre/nu-disco-disco"},
		{ID: 8, Name: "Funky House", Slug: "funky-house", URL: traxsourceBaseURL + "/genre/funky-house"},
		{ID: 9, Name: "Progressive House", Slug: "progressive-house", URL: traxsourceBaseURL + "/genre/progressive-house"},
		{ID: 10, Name: "Jackin House", Slug: "jackin-house", URL: traxsourceBaseURL + "/genre/jackin-house"},
		{ID: 11, Name: "Soulful House", Slug: "soulful-house", URL: traxsourceBaseURL + "/genre/soulful-house"},
		{ID: 12, Name: "Minimal / Deep Tech", Slug: "minimal-deep-tech", URL: traxsourceBaseURL + "/genre/minimal-deep-tech"},
		{ID: 13, Name: "Organic House / Downtempo", Slug: "organic-house-downtempo", URL: traxsourceBaseURL + "/genre/organic-house-downtempo"},
		{ID: 14, Name: "Indie Dance", Slug: "indie-dance", URL: traxsourceBaseURL + "/genre/indie-dance"},
		{ID: 15, Name: "Lounge / Chill Out", Slug: "lounge-chill-out", URL: traxsourceBaseURL + "/genre/lounge-chill-out"},
	}
	return genres, nil
}

// GetGenreTracks gets tracks for a specific genre
func (c *TraxsourceClient) GetGenreTracks(ctx context.Context, genreSlug string, page int) ([]TraxsourceTrack, error) {
	if page < 1 {
		page = 1
	}
	genreURL := fmt.Sprintf("%s/genre/%s?page=%d", traxsourceBaseURL, url.PathEscape(genreSlug), page)
	body, err := c.fetchPage(ctx, genreURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genre tracks: %w", err)
	}

	return c.parseTracksFromHTML(body), nil
}

// DownloadPreview downloads a track preview using yt-dlp
func (c *TraxsourceClient) DownloadPreview(ctx context.Context, trackURL string, outputDir string) (string, error) {
	// Check if yt-dlp is available
	ytdlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return "", fmt.Errorf("yt-dlp not found: %w", err)
	}

	// Build output template
	outputTemplate := fmt.Sprintf("%s/%%(title)s.%%(ext)s", outputDir)

	// Build yt-dlp command
	args := []string{
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"-o", outputTemplate,
		trackURL,
	}

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w - %s", err, string(output))
	}

	return string(output), nil
}

// fetchPage fetches a page and returns the body
func (c *TraxsourceClient) fetchPage(ctx context.Context, pageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseTracksFromHTML extracts tracks from HTML content
func (c *TraxsourceClient) parseTracksFromHTML(html string) []TraxsourceTrack {
	var tracks []TraxsourceTrack

	// Pattern to find track data in JSON-LD or data attributes
	// Traxsource often embeds track data in script tags
	jsonLDPattern := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	matches := jsonLDPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &data); err == nil {
				if track := c.extractTrackFromJSONLD(data); track != nil {
					tracks = append(tracks, *track)
				}
			}
		}
	}

	// Also try to parse track list items
	trackPattern := regexp.MustCompile(`class="track-item"[^>]*data-track-id="(\d+)"`)
	titlePattern := regexp.MustCompile(`class="title"[^>]*>([^<]+)`)
	artistPattern := regexp.MustCompile(`class="artist"[^>]*>([^<]+)`)

	trackMatches := trackPattern.FindAllStringSubmatch(html, -1)
	for _, m := range trackMatches {
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			track := TraxsourceTrack{
				ID:  id,
				URL: fmt.Sprintf("%s/track/%d", traxsourceBaseURL, id),
			}
			tracks = append(tracks, track)
		}
	}

	// Extract basic info if JSON-LD didn't work
	if len(tracks) == 0 {
		titles := titlePattern.FindAllStringSubmatch(html, 50)
		artists := artistPattern.FindAllStringSubmatch(html, 50)

		for i, t := range titles {
			if len(t) > 1 {
				track := TraxsourceTrack{
					Title: strings.TrimSpace(t[1]),
				}
				if i < len(artists) && len(artists[i]) > 1 {
					track.Artists = []string{strings.TrimSpace(artists[i][1])}
				}
				tracks = append(tracks, track)
			}
		}
	}

	return tracks
}

// extractTrackFromJSONLD extracts track info from JSON-LD data
func (c *TraxsourceClient) extractTrackFromJSONLD(data map[string]interface{}) *TraxsourceTrack {
	schemaType, _ := data["@type"].(string)
	if schemaType != "MusicRecording" && schemaType != "Product" {
		return nil
	}

	track := &TraxsourceTrack{}

	if name, ok := data["name"].(string); ok {
		track.Title = name
	}
	if url, ok := data["url"].(string); ok {
		track.URL = url
		// Extract ID from URL
		idPattern := regexp.MustCompile(`/track/(\d+)`)
		if m := idPattern.FindStringSubmatch(url); len(m) > 1 {
			track.ID, _ = strconv.Atoi(m[1])
		}
	}
	if image, ok := data["image"].(string); ok {
		track.ImageURL = image
	}
	if byArtist, ok := data["byArtist"].(map[string]interface{}); ok {
		if name, ok := byArtist["name"].(string); ok {
			track.Artists = []string{name}
		}
	}

	return track
}

// parseReleasesFromHTML extracts releases from HTML content
func (c *TraxsourceClient) parseReleasesFromHTML(html string) []TraxsourceRelease {
	var releases []TraxsourceRelease

	// Pattern to find release items
	releasePattern := regexp.MustCompile(`class="release-item"[^>]*data-release-id="(\d+)"`)
	matches := releasePattern.FindAllStringSubmatch(html, -1)

	for _, m := range matches {
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			release := TraxsourceRelease{
				ID:  id,
				URL: fmt.Sprintf("%s/title/%d", traxsourceBaseURL, id),
			}
			releases = append(releases, release)
		}
	}

	return releases
}

// parseArtistsFromHTML extracts artists from HTML content
func (c *TraxsourceClient) parseArtistsFromHTML(html string) []TraxsourceArtist {
	var artists []TraxsourceArtist

	artistPattern := regexp.MustCompile(`<a[^>]*href="/artist/(\d+)/[^"]*"[^>]*class="[^"]*artist[^"]*"[^>]*>([^<]+)`)
	matches := artistPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[int]bool)
	for _, m := range matches {
		if len(m) > 2 {
			id, _ := strconv.Atoi(m[1])
			if !seen[id] {
				seen[id] = true
				artist := TraxsourceArtist{
					ID:   id,
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/artist/%d", traxsourceBaseURL, id),
				}
				artists = append(artists, artist)
			}
		}
	}

	return artists
}

// parseLabelsFromHTML extracts labels from HTML content
func (c *TraxsourceClient) parseLabelsFromHTML(html string) []TraxsourceLabel {
	var labels []TraxsourceLabel

	labelPattern := regexp.MustCompile(`<a[^>]*href="/label/(\d+)/[^"]*"[^>]*>([^<]+)`)
	matches := labelPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[int]bool)
	for _, m := range matches {
		if len(m) > 2 {
			id, _ := strconv.Atoi(m[1])
			if !seen[id] {
				seen[id] = true
				label := TraxsourceLabel{
					ID:   id,
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/label/%d", traxsourceBaseURL, id),
				}
				labels = append(labels, label)
			}
		}
	}

	return labels
}

// parseChartsFromHTML extracts DJ charts from HTML
func (c *TraxsourceClient) parseChartsFromHTML(html string) []TraxsourceChart {
	var charts []TraxsourceChart

	chartPattern := regexp.MustCompile(`class="chart-item"[^>]*data-chart-id="(\d+)"`)
	matches := chartPattern.FindAllStringSubmatch(html, -1)

	for _, m := range matches {
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			chart := TraxsourceChart{
				ID:  id,
				URL: fmt.Sprintf("%s/chart/%d", traxsourceBaseURL, id),
			}
			charts = append(charts, chart)
		}
	}

	return charts
}

// parseTrackDetailsFromHTML extracts detailed track info from a track page
func (c *TraxsourceClient) parseTrackDetailsFromHTML(html string, trackID int) *TraxsourceTrack {
	track := &TraxsourceTrack{
		ID:  trackID,
		URL: fmt.Sprintf("%s/track/%d", traxsourceBaseURL, trackID),
	}

	// Try JSON-LD first
	jsonLDPattern := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	if matches := jsonLDPattern.FindStringSubmatch(html); len(matches) > 1 {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			if name, ok := data["name"].(string); ok {
				track.Title = name
			}
			if image, ok := data["image"].(string); ok {
				track.ImageURL = image
			}
		}
	}

	// Extract BPM
	bpmPattern := regexp.MustCompile(`BPM[:\s]*(\d+)`)
	if m := bpmPattern.FindStringSubmatch(html); len(m) > 1 {
		track.BPM, _ = strconv.Atoi(m[1])
	}

	// Extract key
	keyPattern := regexp.MustCompile(`Key[:\s]*([A-G][#b]?(?:m|min|maj)?)\b`)
	if m := keyPattern.FindStringSubmatch(html); len(m) > 1 {
		track.Key = m[1]
	}

	// Extract genre
	genrePattern := regexp.MustCompile(`class="genre"[^>]*>([^<]+)`)
	if m := genrePattern.FindStringSubmatch(html); len(m) > 1 {
		track.Genre = strings.TrimSpace(m[1])
	}

	// Extract label
	labelPattern := regexp.MustCompile(`class="label"[^>]*>([^<]+)`)
	if m := labelPattern.FindStringSubmatch(html); len(m) > 1 {
		track.Label = strings.TrimSpace(m[1])
	}

	return track
}

// parseReleaseDetailsFromHTML extracts detailed release info from a release page
func (c *TraxsourceClient) parseReleaseDetailsFromHTML(html string, releaseID int) *TraxsourceRelease {
	release := &TraxsourceRelease{
		ID:  releaseID,
		URL: fmt.Sprintf("%s/title/%d", traxsourceBaseURL, releaseID),
	}

	// Try JSON-LD first
	jsonLDPattern := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	if matches := jsonLDPattern.FindStringSubmatch(html); len(matches) > 1 {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(matches[1]), &data); err == nil {
			if name, ok := data["name"].(string); ok {
				release.Title = name
			}
			if image, ok := data["image"].(string); ok {
				release.ImageURL = image
			}
		}
	}

	// Extract tracks from the release page
	release.Tracks = c.parseTracksFromHTML(html)

	return release
}

// parseArtistDetailsFromHTML extracts detailed artist info from an artist page
func (c *TraxsourceClient) parseArtistDetailsFromHTML(html string, artistID int) *TraxsourceArtist {
	artist := &TraxsourceArtist{
		ID:  artistID,
		URL: fmt.Sprintf("%s/artist/%d", traxsourceBaseURL, artistID),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Name = strings.TrimSpace(m[1])
	}

	// Extract bio
	bioPattern := regexp.MustCompile(`class="bio"[^>]*>(.*?)</div>`)
	if m := bioPattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Bio = strings.TrimSpace(stripHTMLTags(m[1]))
	}

	return artist
}

// parseLabelDetailsFromHTML extracts detailed label info from a label page
func (c *TraxsourceClient) parseLabelDetailsFromHTML(html string, labelID int) *TraxsourceLabel {
	label := &TraxsourceLabel{
		ID:  labelID,
		URL: fmt.Sprintf("%s/label/%d", traxsourceBaseURL, labelID),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		label.Name = strings.TrimSpace(m[1])
	}

	// Extract description
	descPattern := regexp.MustCompile(`class="description"[^>]*>(.*?)</div>`)
	if m := descPattern.FindStringSubmatch(html); len(m) > 1 {
		label.Description = strings.TrimSpace(stripHTMLTags(m[1]))
	}

	return label
}

// stripHTMLTags removes HTML tags from a string
func stripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}
