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

const junoBaseURL = "https://www.junodownload.com"

// JunoClient provides access to Juno Download electronic music store
type JunoClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
}

// JunoTrack represents a track on Juno Download
type JunoTrack struct {
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
	CatalogNum  string   `json:"catalog_number,omitempty"`
	Format      string   `json:"format,omitempty"` // WAV, MP3, FLAC
}

// JunoRelease represents a release/EP on Juno Download
type JunoRelease struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Artists     []string    `json:"artists"`
	Label       string      `json:"label"`
	Genre       string      `json:"genre"`
	CatalogNum  string      `json:"catalog_number,omitempty"`
	ReleaseDate string      `json:"release_date,omitempty"`
	Tracks      []JunoTrack `json:"tracks,omitempty"`
	URL         string      `json:"url"`
	ImageURL    string      `json:"image_url,omitempty"`
	Price       string      `json:"price,omitempty"`
	Description string      `json:"description,omitempty"`
}

// JunoArtist represents an artist on Juno Download
type JunoArtist struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	ImageURL   string   `json:"image_url,omitempty"`
	Bio        string   `json:"bio,omitempty"`
	Genres     []string `json:"genres,omitempty"`
	TrackCount int      `json:"track_count,omitempty"`
}

// JunoLabel represents a record label on Juno Download
type JunoLabel struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	ImageURL     string   `json:"image_url,omitempty"`
	Description  string   `json:"description,omitempty"`
	Genres       []string `json:"genres,omitempty"`
	ReleaseCount int      `json:"release_count,omitempty"`
}

// JunoChart represents a featured chart
type JunoChart struct {
	ID       int         `json:"id"`
	Title    string      `json:"title"`
	DJ       string      `json:"dj,omitempty"`
	Date     string      `json:"date,omitempty"`
	Genre    string      `json:"genre,omitempty"`
	Tracks   []JunoTrack `json:"tracks,omitempty"`
	URL      string      `json:"url"`
	ImageURL string      `json:"image_url,omitempty"`
}

// JunoSearchResults contains search results
type JunoSearchResults struct {
	Tracks   []JunoTrack   `json:"tracks,omitempty"`
	Releases []JunoRelease `json:"releases,omitempty"`
	Artists  []JunoArtist  `json:"artists,omitempty"`
	Labels   []JunoLabel   `json:"labels,omitempty"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PerPage  int           `json:"per_page"`
}

// JunoStatus represents connection status
type JunoStatus struct {
	Available      bool   `json:"available"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	YtDlpAvailable bool   `json:"yt_dlp_available"`
	Message        string `json:"message,omitempty"`
}

// JunoHealth represents health check results
type JunoHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	SiteReachable   bool     `json:"site_reachable"`
	YtDlpAvailable  bool     `json:"yt_dlp_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// JunoGenre represents a music genre
type JunoGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

// NewJunoClient creates a new Juno Download client
func NewJunoClient() (*JunoClient, error) {
	return &JunoClient{
		httpClient: httpclient.Standard(),
	}, nil
}

// GetStatus returns the current connection status
func (c *JunoClient) GetStatus(ctx context.Context) (*JunoStatus, error) {
	status := &JunoStatus{
		Available: false,
	}

	// Check site availability
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", junoBaseURL, nil)
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
		status.Message = "Juno Download is accessible"
	} else {
		status.Message = fmt.Sprintf("Site returned status %d", resp.StatusCode)
	}

	return status, nil
}

// GetHealth performs a health check
func (c *JunoClient) GetHealth(ctx context.Context) (*JunoHealth, error) {
	health := &JunoHealth{
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
		health.Issues = append(health.Issues, "Juno Download site is not reachable")
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
func (c *JunoClient) Search(ctx context.Context, query string, searchType string, page int, perPage int) (*JunoSearchResults, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	// Juno uses different search endpoints
	// Main search: /search/?q=query&facet=type
	var searchURL string
	switch searchType {
	case "track", "tracks":
		searchURL = fmt.Sprintf("%s/search/?q=%s&facet=track&page=%d", junoBaseURL, url.QueryEscape(query), page)
	case "release", "releases":
		searchURL = fmt.Sprintf("%s/search/?q=%s&facet=release&page=%d", junoBaseURL, url.QueryEscape(query), page)
	case "artist", "artists":
		searchURL = fmt.Sprintf("%s/search/?q=%s&facet=artist&page=%d", junoBaseURL, url.QueryEscape(query), page)
	case "label", "labels":
		searchURL = fmt.Sprintf("%s/search/?q=%s&facet=label&page=%d", junoBaseURL, url.QueryEscape(query), page)
	default:
		// Default to all results
		searchURL = fmt.Sprintf("%s/search/?q=%s&page=%d", junoBaseURL, url.QueryEscape(query), page)
	}

	body, err := c.fetchPage(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results := &JunoSearchResults{
		Page:    page,
		PerPage: perPage,
	}

	// Parse results based on type
	switch searchType {
	case "track", "tracks":
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
	default:
		// Parse all types for general search
		results.Tracks = c.parseTracksFromHTML(body)
		results.Releases = c.parseReleasesFromHTML(body)
		results.Artists = c.parseArtistsFromHTML(body)
		results.Labels = c.parseLabelsFromHTML(body)
		results.Total = len(results.Tracks) + len(results.Releases) + len(results.Artists) + len(results.Labels)
	}

	return results, nil
}

// GetTrack gets detailed information about a specific track
func (c *JunoClient) GetTrack(ctx context.Context, trackID int) (*JunoTrack, error) {
	// Juno uses product pages that include all tracks
	// Track URLs are like: /products/artist-name-title/123456-01/
	trackURL := fmt.Sprintf("%s/products/?track_id=%d", junoBaseURL, trackID)
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
func (c *JunoClient) GetRelease(ctx context.Context, releaseID int) (*JunoRelease, error) {
	// Juno release URLs: /products/artist-name-title/123456/
	releaseURL := fmt.Sprintf("%s/products/%d/", junoBaseURL, releaseID)
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
func (c *JunoClient) GetArtist(ctx context.Context, artistSlug string) (*JunoArtist, error) {
	artistURL := fmt.Sprintf("%s/artists/%s/", junoBaseURL, url.PathEscape(artistSlug))
	body, err := c.fetchPage(ctx, artistURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist: %w", err)
	}

	artist := c.parseArtistDetailsFromHTML(body, artistSlug)
	if artist == nil {
		return nil, fmt.Errorf("artist not found")
	}

	return artist, nil
}

// GetArtistReleases gets releases by an artist
func (c *JunoClient) GetArtistReleases(ctx context.Context, artistSlug string, page int) ([]JunoRelease, error) {
	if page < 1 {
		page = 1
	}
	releasesURL := fmt.Sprintf("%s/artists/%s/releases/?page=%d", junoBaseURL, url.PathEscape(artistSlug), page)
	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetLabel gets information about a record label
func (c *JunoClient) GetLabel(ctx context.Context, labelSlug string) (*JunoLabel, error) {
	labelURL := fmt.Sprintf("%s/labels/%s/", junoBaseURL, url.PathEscape(labelSlug))
	body, err := c.fetchPage(ctx, labelURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch label: %w", err)
	}

	label := c.parseLabelDetailsFromHTML(body, labelSlug)
	if label == nil {
		return nil, fmt.Errorf("label not found")
	}

	return label, nil
}

// GetLabelReleases gets releases from a label
func (c *JunoClient) GetLabelReleases(ctx context.Context, labelSlug string, page int) ([]JunoRelease, error) {
	if page < 1 {
		page = 1
	}
	releasesURL := fmt.Sprintf("%s/labels/%s/releases/?page=%d", junoBaseURL, url.PathEscape(labelSlug), page)
	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch label releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetNewReleases gets new releases, optionally filtered by genre
func (c *JunoClient) GetNewReleases(ctx context.Context, genre string, page int) ([]JunoRelease, error) {
	if page < 1 {
		page = 1
	}

	var releasesURL string
	if genre != "" {
		releasesURL = fmt.Sprintf("%s/%s/this-week/releases/?page=%d", junoBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		releasesURL = fmt.Sprintf("%s/this-week/releases/?page=%d", junoBaseURL, page)
	}

	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetTopSellers gets top selling tracks
func (c *JunoClient) GetTopSellers(ctx context.Context, genre string, page int) ([]JunoTrack, error) {
	if page < 1 {
		page = 1
	}

	var chartsURL string
	if genre != "" {
		chartsURL = fmt.Sprintf("%s/%s/bestsellers/?page=%d", junoBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		chartsURL = fmt.Sprintf("%s/bestsellers/?page=%d", junoBaseURL, page)
	}

	body, err := c.fetchPage(ctx, chartsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top sellers: %w", err)
	}

	return c.parseTracksFromHTML(body), nil
}

// GetStaffPicks gets staff picks / featured releases
func (c *JunoClient) GetStaffPicks(ctx context.Context, genre string, page int) ([]JunoRelease, error) {
	if page < 1 {
		page = 1
	}

	var picksURL string
	if genre != "" {
		picksURL = fmt.Sprintf("%s/%s/staff-picks/?page=%d", junoBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		picksURL = fmt.Sprintf("%s/staff-picks/?page=%d", junoBaseURL, page)
	}

	body, err := c.fetchPage(ctx, picksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staff picks: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetDJCharts gets DJ charts
func (c *JunoClient) GetDJCharts(ctx context.Context, page int) ([]JunoChart, error) {
	if page < 1 {
		page = 1
	}
	chartsURL := fmt.Sprintf("%s/dj-charts/?page=%d", junoBaseURL, page)
	body, err := c.fetchPage(ctx, chartsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch DJ charts: %w", err)
	}

	return c.parseChartsFromHTML(body), nil
}

// GetGenres returns available genres
func (c *JunoClient) GetGenres(ctx context.Context) ([]JunoGenre, error) {
	// Juno Download genres - these are relatively static
	genres := []JunoGenre{
		{ID: 1, Name: "House", Slug: "house", URL: junoBaseURL + "/house/"},
		{ID: 2, Name: "Deep House", Slug: "deep-house", URL: junoBaseURL + "/deep-house/"},
		{ID: 3, Name: "Tech House", Slug: "tech-house", URL: junoBaseURL + "/tech-house/"},
		{ID: 4, Name: "Techno", Slug: "techno", URL: junoBaseURL + "/techno/"},
		{ID: 5, Name: "Minimal / Deep Tech", Slug: "minimal-deep-tech", URL: junoBaseURL + "/minimal-deep-tech/"},
		{ID: 6, Name: "Electro House", Slug: "electro-house", URL: junoBaseURL + "/electro-house/"},
		{ID: 7, Name: "Progressive House", Slug: "progressive-house", URL: junoBaseURL + "/progressive-house/"},
		{ID: 8, Name: "Trance", Slug: "trance", URL: junoBaseURL + "/trance/"},
		{ID: 9, Name: "Drum & Bass", Slug: "drum-and-bass", URL: junoBaseURL + "/drum-and-bass/"},
		{ID: 10, Name: "Breaks / Breakbeat", Slug: "breaks-breakbeat", URL: junoBaseURL + "/breaks-breakbeat/"},
		{ID: 11, Name: "Dubstep / Grime", Slug: "dubstep-grime", URL: junoBaseURL + "/dubstep-grime/"},
		{ID: 12, Name: "Garage / Bassline", Slug: "garage-bassline", URL: junoBaseURL + "/garage-bassline/"},
		{ID: 13, Name: "Disco / Nu-Disco", Slug: "disco-nu-disco", URL: junoBaseURL + "/disco-nu-disco/"},
		{ID: 14, Name: "Funk / Soul", Slug: "funk-soul", URL: junoBaseURL + "/funk-soul/"},
		{ID: 15, Name: "Downtempo / Chill Out", Slug: "downtempo-chill-out", URL: junoBaseURL + "/downtempo-chill-out/"},
		{ID: 16, Name: "Ambient / Electronica", Slug: "ambient-electronica", URL: junoBaseURL + "/ambient-electronica/"},
		{ID: 17, Name: "Hard Dance / Hardcore", Slug: "hard-dance-hardcore", URL: junoBaseURL + "/hard-dance-hardcore/"},
		{ID: 18, Name: "Hip Hop / R&B", Slug: "hip-hop-rnb", URL: junoBaseURL + "/hip-hop-rnb/"},
		{ID: 19, Name: "Indie / Alternative", Slug: "indie-alternative", URL: junoBaseURL + "/indie-alternative/"},
		{ID: 20, Name: "Reggae / Dancehall", Slug: "reggae-dancehall", URL: junoBaseURL + "/reggae-dancehall/"},
	}
	return genres, nil
}

// GetGenreTracks gets tracks for a specific genre
func (c *JunoClient) GetGenreTracks(ctx context.Context, genreSlug string, page int) ([]JunoTrack, error) {
	if page < 1 {
		page = 1
	}
	genreURL := fmt.Sprintf("%s/%s/?page=%d", junoBaseURL, url.PathEscape(genreSlug), page)
	body, err := c.fetchPage(ctx, genreURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genre tracks: %w", err)
	}

	return c.parseTracksFromHTML(body), nil
}

// DownloadPreview downloads a track preview using yt-dlp
func (c *JunoClient) DownloadPreview(ctx context.Context, trackURL string, outputDir string) (string, error) {
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
func (c *JunoClient) fetchPage(ctx context.Context, pageURL string) (string, error) {
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
func (c *JunoClient) parseTracksFromHTML(html string) []JunoTrack {
	var tracks []JunoTrack

	// Pattern to find track data in JSON-LD or structured data
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

	// Parse product items for tracks
	// Juno uses data-pid and data-track attributes
	productPattern := regexp.MustCompile(`data-pid="(\d+)"[^>]*data-track="([^"]*)"`)
	productMatches := productPattern.FindAllStringSubmatch(html, -1)

	for _, m := range productMatches {
		if len(m) > 2 {
			id, _ := strconv.Atoi(m[1])
			track := JunoTrack{
				ID:    id,
				Title: m[2],
				URL:   fmt.Sprintf("%s/products/%d/", junoBaseURL, id),
			}
			tracks = append(tracks, track)
		}
	}

	// Try to parse track list items with more detail
	titlePattern := regexp.MustCompile(`class="[^"]*product-title[^"]*"[^>]*>([^<]+)`)
	artistPattern := regexp.MustCompile(`class="[^"]*product-artist[^"]*"[^>]*>([^<]+)`)
	labelPattern := regexp.MustCompile(`class="[^"]*product-label[^"]*"[^>]*>([^<]+)`)

	if len(tracks) == 0 {
		titles := titlePattern.FindAllStringSubmatch(html, 50)
		artists := artistPattern.FindAllStringSubmatch(html, 50)
		labels := labelPattern.FindAllStringSubmatch(html, 50)

		for i, t := range titles {
			if len(t) > 1 {
				track := JunoTrack{
					Title: strings.TrimSpace(t[1]),
				}
				if i < len(artists) && len(artists[i]) > 1 {
					track.Artists = []string{strings.TrimSpace(artists[i][1])}
				}
				if i < len(labels) && len(labels[i]) > 1 {
					track.Label = strings.TrimSpace(labels[i][1])
				}
				tracks = append(tracks, track)
			}
		}
	}

	return tracks
}

// extractTrackFromJSONLD extracts track info from JSON-LD data
func (c *JunoClient) extractTrackFromJSONLD(data map[string]interface{}) *JunoTrack {
	schemaType, _ := data["@type"].(string)
	if schemaType != "MusicRecording" && schemaType != "Product" && schemaType != "MusicAlbum" {
		return nil
	}

	track := &JunoTrack{}

	if name, ok := data["name"].(string); ok {
		track.Title = name
	}
	if urlStr, ok := data["url"].(string); ok {
		track.URL = urlStr
		// Extract ID from URL
		idPattern := regexp.MustCompile(`/products/[^/]*/(\d+)`)
		if m := idPattern.FindStringSubmatch(urlStr); len(m) > 1 {
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
	if offers, ok := data["offers"].(map[string]interface{}); ok {
		if price, ok := offers["price"].(string); ok {
			track.Price = price
		}
	}

	return track
}

// parseReleasesFromHTML extracts releases from HTML content
func (c *JunoClient) parseReleasesFromHTML(html string) []JunoRelease {
	var releases []JunoRelease

	// Pattern to find product items
	productPattern := regexp.MustCompile(`data-pid="(\d+)"`)
	matches := productPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[int]bool)
	for _, m := range matches {
		if len(m) > 1 {
			id, _ := strconv.Atoi(m[1])
			if !seen[id] {
				seen[id] = true
				release := JunoRelease{
					ID:  id,
					URL: fmt.Sprintf("%s/products/%d/", junoBaseURL, id),
				}
				releases = append(releases, release)
			}
		}
	}

	return releases
}

// parseArtistsFromHTML extracts artists from HTML content
func (c *JunoClient) parseArtistsFromHTML(html string) []JunoArtist {
	var artists []JunoArtist

	// Pattern: /artists/artist-name/
	artistPattern := regexp.MustCompile(`<a[^>]*href="/artists/([^/"]+)/"[^>]*>([^<]+)`)
	matches := artistPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) > 2 {
			slug := m[1]
			if !seen[slug] {
				seen[slug] = true
				artist := JunoArtist{
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/artists/%s/", junoBaseURL, slug),
				}
				artists = append(artists, artist)
			}
		}
	}

	return artists
}

// parseLabelsFromHTML extracts labels from HTML content
func (c *JunoClient) parseLabelsFromHTML(html string) []JunoLabel {
	var labels []JunoLabel

	// Pattern: /labels/label-name/
	labelPattern := regexp.MustCompile(`<a[^>]*href="/labels/([^/"]+)/"[^>]*>([^<]+)`)
	matches := labelPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) > 2 {
			slug := m[1]
			if !seen[slug] {
				seen[slug] = true
				label := JunoLabel{
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/labels/%s/", junoBaseURL, slug),
				}
				labels = append(labels, label)
			}
		}
	}

	return labels
}

// parseChartsFromHTML extracts DJ charts from HTML
func (c *JunoClient) parseChartsFromHTML(html string) []JunoChart {
	var charts []JunoChart

	// Pattern for chart items
	chartPattern := regexp.MustCompile(`<a[^>]*href="/dj-charts/([^/"]+)/"[^>]*class="[^"]*chart[^"]*"[^>]*>`)
	djPattern := regexp.MustCompile(`class="[^"]*chart-dj[^"]*"[^>]*>([^<]+)`)
	titlePattern := regexp.MustCompile(`class="[^"]*chart-title[^"]*"[^>]*>([^<]+)`)

	chartMatches := chartPattern.FindAllStringSubmatch(html, -1)
	djMatches := djPattern.FindAllStringSubmatch(html, -1)
	titleMatches := titlePattern.FindAllStringSubmatch(html, -1)

	for i, m := range chartMatches {
		if len(m) > 1 {
			chart := JunoChart{
				URL: fmt.Sprintf("%s/dj-charts/%s/", junoBaseURL, m[1]),
			}
			if i < len(djMatches) && len(djMatches[i]) > 1 {
				chart.DJ = strings.TrimSpace(djMatches[i][1])
			}
			if i < len(titleMatches) && len(titleMatches[i]) > 1 {
				chart.Title = strings.TrimSpace(titleMatches[i][1])
			}
			charts = append(charts, chart)
		}
	}

	return charts
}

// parseTrackDetailsFromHTML extracts detailed track info from a track page
func (c *JunoClient) parseTrackDetailsFromHTML(html string, trackID int) *JunoTrack {
	track := &JunoTrack{
		ID:  trackID,
		URL: fmt.Sprintf("%s/products/%d/", junoBaseURL, trackID),
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
	bpmPattern := regexp.MustCompile(`(\d+)\s*BPM`)
	if m := bpmPattern.FindStringSubmatch(html); len(m) > 1 {
		track.BPM, _ = strconv.Atoi(m[1])
	}

	// Extract key (Juno format: Amin, Cmaj, etc.)
	keyPattern := regexp.MustCompile(`([A-G][#b]?(?:min|maj))\b`)
	if m := keyPattern.FindStringSubmatch(html); len(m) > 1 {
		track.Key = m[1]
	}

	// Extract genre
	genrePattern := regexp.MustCompile(`class="[^"]*product-genre[^"]*"[^>]*>([^<]+)`)
	if m := genrePattern.FindStringSubmatch(html); len(m) > 1 {
		track.Genre = strings.TrimSpace(m[1])
	}

	// Extract label
	labelPattern := regexp.MustCompile(`class="[^"]*product-label[^"]*"[^>]*><a[^>]*>([^<]+)`)
	if m := labelPattern.FindStringSubmatch(html); len(m) > 1 {
		track.Label = strings.TrimSpace(m[1])
	}

	// Extract catalog number
	catPattern := regexp.MustCompile(`Cat:\s*([A-Z0-9-]+)`)
	if m := catPattern.FindStringSubmatch(html); len(m) > 1 {
		track.CatalogNum = m[1]
	}

	// Extract format (WAV, MP3, FLAC)
	formatPattern := regexp.MustCompile(`(WAV|MP3|FLAC|AIFF)`)
	if m := formatPattern.FindStringSubmatch(html); len(m) > 1 {
		track.Format = m[1]
	}

	return track
}

// parseReleaseDetailsFromHTML extracts detailed release info from a release page
func (c *JunoClient) parseReleaseDetailsFromHTML(html string, releaseID int) *JunoRelease {
	release := &JunoRelease{
		ID:  releaseID,
		URL: fmt.Sprintf("%s/products/%d/", junoBaseURL, releaseID),
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
			if desc, ok := data["description"].(string); ok {
				release.Description = desc
			}
		}
	}

	// Extract tracks from the release page
	release.Tracks = c.parseTracksFromHTML(html)

	// Extract catalog number
	catPattern := regexp.MustCompile(`Cat:\s*([A-Z0-9-]+)`)
	if m := catPattern.FindStringSubmatch(html); len(m) > 1 {
		release.CatalogNum = m[1]
	}

	// Extract label
	labelPattern := regexp.MustCompile(`class="[^"]*product-label[^"]*"[^>]*><a[^>]*>([^<]+)`)
	if m := labelPattern.FindStringSubmatch(html); len(m) > 1 {
		release.Label = strings.TrimSpace(m[1])
	}

	return release
}

// parseArtistDetailsFromHTML extracts detailed artist info from an artist page
func (c *JunoClient) parseArtistDetailsFromHTML(html string, artistSlug string) *JunoArtist {
	artist := &JunoArtist{
		URL: fmt.Sprintf("%s/artists/%s/", junoBaseURL, artistSlug),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Name = strings.TrimSpace(m[1])
	}

	// Extract bio
	bioPattern := regexp.MustCompile(`class="[^"]*artist-bio[^"]*"[^>]*>(.*?)</div>`)
	if m := bioPattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Bio = strings.TrimSpace(junoStripHTMLTags(m[1]))
	}

	return artist
}

// parseLabelDetailsFromHTML extracts detailed label info from a label page
func (c *JunoClient) parseLabelDetailsFromHTML(html string, labelSlug string) *JunoLabel {
	label := &JunoLabel{
		URL: fmt.Sprintf("%s/labels/%s/", junoBaseURL, labelSlug),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		label.Name = strings.TrimSpace(m[1])
	}

	// Extract description
	descPattern := regexp.MustCompile(`class="[^"]*label-description[^"]*"[^>]*>(.*?)</div>`)
	if m := descPattern.FindStringSubmatch(html); len(m) > 1 {
		label.Description = strings.TrimSpace(junoStripHTMLTags(m[1]))
	}

	return label
}

// junoStripHTMLTags removes HTML tags from a string
func junoStripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}
