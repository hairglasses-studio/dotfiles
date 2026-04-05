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

const boomkatBaseURL = "https://boomkat.com"

// BoomkatClient provides access to Boomkat electronic music store
type BoomkatClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
}

// BoomkatTrack represents a track on Boomkat
type BoomkatTrack struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Artists     []string `json:"artists"`
	Duration    string   `json:"duration,omitempty"`
	TrackNumber int      `json:"track_number,omitempty"`
	PreviewURL  string   `json:"preview_url,omitempty"`
}

// BoomkatRelease represents a release on Boomkat
type BoomkatRelease struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Artists     []string       `json:"artists"`
	Label       string         `json:"label"`
	Genre       string         `json:"genre,omitempty"`
	Subgenre    string         `json:"subgenre,omitempty"`
	CatalogNum  string         `json:"catalog_number,omitempty"`
	ReleaseDate string         `json:"release_date,omitempty"`
	Format      string         `json:"format,omitempty"` // Vinyl, CD, Digital
	Price       string         `json:"price,omitempty"`
	URL         string         `json:"url"`
	ImageURL    string         `json:"image_url,omitempty"`
	Description string         `json:"description,omitempty"`
	Tracks      []BoomkatTrack `json:"tracks,omitempty"`
	Rating      string         `json:"rating,omitempty"` // Essential, Recommended, etc.
}

// BoomkatArtist represents an artist on Boomkat
type BoomkatArtist struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Bio          string `json:"bio,omitempty"`
	ReleaseCount int    `json:"release_count,omitempty"`
}

// BoomkatLabel represents a record label on Boomkat
type BoomkatLabel struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	Description  string `json:"description,omitempty"`
	ReleaseCount int    `json:"release_count,omitempty"`
}

// BoomkatSearchResults contains search results
type BoomkatSearchResults struct {
	Releases []BoomkatRelease `json:"releases,omitempty"`
	Artists  []BoomkatArtist  `json:"artists,omitempty"`
	Labels   []BoomkatLabel   `json:"labels,omitempty"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PerPage  int              `json:"per_page"`
}

// BoomkatStatus represents connection status
type BoomkatStatus struct {
	Available      bool   `json:"available"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	YtDlpAvailable bool   `json:"yt_dlp_available"`
	Message        string `json:"message,omitempty"`
}

// BoomkatHealth represents health check results
type BoomkatHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	SiteReachable   bool     `json:"site_reachable"`
	YtDlpAvailable  bool     `json:"yt_dlp_available"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// BoomkatGenre represents a music genre/category
type BoomkatGenre struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

// NewBoomkatClient creates a new Boomkat client
func NewBoomkatClient() (*BoomkatClient, error) {
	return &BoomkatClient{
		httpClient: httpclient.Standard(),
	}, nil
}

// GetStatus returns the current connection status
func (c *BoomkatClient) GetStatus(ctx context.Context) (*BoomkatStatus, error) {
	status := &BoomkatStatus{
		Available: false,
	}

	// Check site availability
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", boomkatBaseURL, nil)
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
		status.Message = "Boomkat is accessible"
	} else {
		status.Message = fmt.Sprintf("Site returned status %d", resp.StatusCode)
	}

	return status, nil
}

// GetHealth performs a health check
func (c *BoomkatClient) GetHealth(ctx context.Context) (*BoomkatHealth, error) {
	health := &BoomkatHealth{
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
		health.Issues = append(health.Issues, "Boomkat site is not reachable")
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

// Search searches for releases, artists, or labels
func (c *BoomkatClient) Search(ctx context.Context, query string, searchType string, page int, perPage int) (*BoomkatSearchResults, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	// Boomkat search URL: /search?q=query
	searchURL := fmt.Sprintf("%s/search?q=%s&page=%d", boomkatBaseURL, url.QueryEscape(query), page)

	body, err := c.fetchPage(ctx, searchURL)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	results := &BoomkatSearchResults{
		Page:    page,
		PerPage: perPage,
	}

	// Parse results
	results.Releases = c.parseReleasesFromHTML(body)
	results.Artists = c.parseArtistsFromHTML(body)
	results.Labels = c.parseLabelsFromHTML(body)
	results.Total = len(results.Releases) + len(results.Artists) + len(results.Labels)

	return results, nil
}

// GetRelease gets detailed information about a release
func (c *BoomkatClient) GetRelease(ctx context.Context, releaseSlug string) (*BoomkatRelease, error) {
	// Boomkat release URLs: /products/artist-name-title
	releaseURL := fmt.Sprintf("%s/products/%s", boomkatBaseURL, url.PathEscape(releaseSlug))
	body, err := c.fetchPage(ctx, releaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}

	release := c.parseReleaseDetailsFromHTML(body, releaseSlug)
	if release == nil {
		return nil, fmt.Errorf("release not found")
	}

	return release, nil
}

// GetArtist gets information about an artist
func (c *BoomkatClient) GetArtist(ctx context.Context, artistSlug string) (*BoomkatArtist, error) {
	artistURL := fmt.Sprintf("%s/artists/%s", boomkatBaseURL, url.PathEscape(artistSlug))
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
func (c *BoomkatClient) GetArtistReleases(ctx context.Context, artistSlug string, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}
	releasesURL := fmt.Sprintf("%s/artists/%s?page=%d", boomkatBaseURL, url.PathEscape(artistSlug), page)
	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetLabel gets information about a record label
func (c *BoomkatClient) GetLabel(ctx context.Context, labelSlug string) (*BoomkatLabel, error) {
	labelURL := fmt.Sprintf("%s/labels/%s", boomkatBaseURL, url.PathEscape(labelSlug))
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
func (c *BoomkatClient) GetLabelReleases(ctx context.Context, labelSlug string, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}
	releasesURL := fmt.Sprintf("%s/labels/%s?page=%d", boomkatBaseURL, url.PathEscape(labelSlug), page)
	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch label releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetNewReleases gets new releases, optionally filtered by genre
func (c *BoomkatClient) GetNewReleases(ctx context.Context, genre string, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}

	var releasesURL string
	if genre != "" {
		releasesURL = fmt.Sprintf("%s/downloads/%s?page=%d", boomkatBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		releasesURL = fmt.Sprintf("%s/downloads?page=%d", boomkatBaseURL, page)
	}

	body, err := c.fetchPage(ctx, releasesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetBestsellers gets bestselling releases
func (c *BoomkatClient) GetBestsellers(ctx context.Context, genre string, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}

	var chartsURL string
	if genre != "" {
		chartsURL = fmt.Sprintf("%s/bestsellers/%s?page=%d", boomkatBaseURL, url.PathEscape(strings.ToLower(genre)), page)
	} else {
		chartsURL = fmt.Sprintf("%s/bestsellers?page=%d", boomkatBaseURL, page)
	}

	body, err := c.fetchPage(ctx, chartsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bestsellers: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetRecommended gets Boomkat recommended releases (curated picks)
func (c *BoomkatClient) GetRecommended(ctx context.Context, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}

	recURL := fmt.Sprintf("%s/recommended?page=%d", boomkatBaseURL, page)
	body, err := c.fetchPage(ctx, recURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recommended: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetEssential gets Boomkat essential releases (highest rated)
func (c *BoomkatClient) GetEssential(ctx context.Context, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}

	essURL := fmt.Sprintf("%s/downloads?rating=essential&page=%d", boomkatBaseURL, page)
	body, err := c.fetchPage(ctx, essURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch essential: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// GetGenres returns available genres
func (c *BoomkatClient) GetGenres(ctx context.Context) ([]BoomkatGenre, error) {
	// Boomkat genres - curated selection of electronic/experimental genres
	genres := []BoomkatGenre{
		{ID: "1", Name: "Techno", Slug: "techno", URL: boomkatBaseURL + "/downloads/techno"},
		{ID: "2", Name: "House", Slug: "house", URL: boomkatBaseURL + "/downloads/house"},
		{ID: "3", Name: "Electro", Slug: "electro", URL: boomkatBaseURL + "/downloads/electro"},
		{ID: "4", Name: "Ambient", Slug: "ambient", URL: boomkatBaseURL + "/downloads/ambient"},
		{ID: "5", Name: "Experimental", Slug: "experimental", URL: boomkatBaseURL + "/downloads/experimental"},
		{ID: "6", Name: "Electronic", Slug: "electronic", URL: boomkatBaseURL + "/downloads/electronic"},
		{ID: "7", Name: "Industrial", Slug: "industrial", URL: boomkatBaseURL + "/downloads/industrial"},
		{ID: "8", Name: "Noise", Slug: "noise", URL: boomkatBaseURL + "/downloads/noise"},
		{ID: "9", Name: "Drone", Slug: "drone", URL: boomkatBaseURL + "/downloads/drone"},
		{ID: "10", Name: "IDM", Slug: "idm", URL: boomkatBaseURL + "/downloads/idm"},
		{ID: "11", Name: "Bass", Slug: "bass", URL: boomkatBaseURL + "/downloads/bass"},
		{ID: "12", Name: "Dub", Slug: "dub", URL: boomkatBaseURL + "/downloads/dub"},
		{ID: "13", Name: "Disco", Slug: "disco", URL: boomkatBaseURL + "/downloads/disco"},
		{ID: "14", Name: "Leftfield", Slug: "leftfield", URL: boomkatBaseURL + "/downloads/leftfield"},
		{ID: "15", Name: "Modern Classical", Slug: "modern-classical", URL: boomkatBaseURL + "/downloads/modern-classical"},
		{ID: "16", Name: "Synth", Slug: "synth", URL: boomkatBaseURL + "/downloads/synth"},
		{ID: "17", Name: "Wave", Slug: "wave", URL: boomkatBaseURL + "/downloads/wave"},
		{ID: "18", Name: "EBM", Slug: "ebm", URL: boomkatBaseURL + "/downloads/ebm"},
	}
	return genres, nil
}

// GetGenreReleases gets releases for a specific genre
func (c *BoomkatClient) GetGenreReleases(ctx context.Context, genreSlug string, page int) ([]BoomkatRelease, error) {
	if page < 1 {
		page = 1
	}
	genreURL := fmt.Sprintf("%s/downloads/%s?page=%d", boomkatBaseURL, url.PathEscape(genreSlug), page)
	body, err := c.fetchPage(ctx, genreURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch genre releases: %w", err)
	}

	return c.parseReleasesFromHTML(body), nil
}

// DownloadPreview downloads a release preview using yt-dlp
func (c *BoomkatClient) DownloadPreview(ctx context.Context, releaseURL string, outputDir string) (string, error) {
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
		releaseURL,
	}

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w - %s", err, string(output))
	}

	return string(output), nil
}

// fetchPage fetches a page and returns the body
func (c *BoomkatClient) fetchPage(ctx context.Context, pageURL string) (string, error) {
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

// parseReleasesFromHTML extracts releases from HTML content
func (c *BoomkatClient) parseReleasesFromHTML(html string) []BoomkatRelease {
	var releases []BoomkatRelease

	// Pattern to find product data in JSON-LD or structured data
	jsonLDPattern := regexp.MustCompile(`<script type="application/ld\+json">(.*?)</script>`)
	matches := jsonLDPattern.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 1 {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &data); err == nil {
				if release := c.extractReleaseFromJSONLD(data); release != nil {
					releases = append(releases, *release)
				}
			}
		}
	}

	// Parse product cards
	// Boomkat uses class patterns like "product-card" or "release-item"
	productPattern := regexp.MustCompile(`<a[^>]*href="/products/([^"]+)"[^>]*class="[^"]*product[^"]*"`)
	titlePattern := regexp.MustCompile(`class="[^"]*product-title[^"]*"[^>]*>([^<]+)`)
	artistPattern := regexp.MustCompile(`class="[^"]*product-artist[^"]*"[^>]*>([^<]+)`)
	labelPattern := regexp.MustCompile(`class="[^"]*product-label[^"]*"[^>]*>([^<]+)`)

	productMatches := productPattern.FindAllStringSubmatch(html, -1)
	for _, m := range productMatches {
		if len(m) > 1 {
			release := BoomkatRelease{
				ID:  m[1],
				URL: fmt.Sprintf("%s/products/%s", boomkatBaseURL, m[1]),
			}
			releases = append(releases, release)
		}
	}

	// Extract more details if JSON-LD didn't work
	if len(releases) == 0 {
		titles := titlePattern.FindAllStringSubmatch(html, 50)
		artists := artistPattern.FindAllStringSubmatch(html, 50)
		labels := labelPattern.FindAllStringSubmatch(html, 50)

		for i, t := range titles {
			if len(t) > 1 {
				release := BoomkatRelease{
					Title: strings.TrimSpace(t[1]),
				}
				if i < len(artists) && len(artists[i]) > 1 {
					release.Artists = []string{strings.TrimSpace(artists[i][1])}
				}
				if i < len(labels) && len(labels[i]) > 1 {
					release.Label = strings.TrimSpace(labels[i][1])
				}
				releases = append(releases, release)
			}
		}
	}

	return releases
}

// extractReleaseFromJSONLD extracts release info from JSON-LD data
func (c *BoomkatClient) extractReleaseFromJSONLD(data map[string]interface{}) *BoomkatRelease {
	schemaType, _ := data["@type"].(string)
	if schemaType != "MusicAlbum" && schemaType != "Product" && schemaType != "MusicRecording" {
		return nil
	}

	release := &BoomkatRelease{}

	if name, ok := data["name"].(string); ok {
		release.Title = name
	}
	if urlStr, ok := data["url"].(string); ok {
		release.URL = urlStr
		// Extract ID from URL
		idPattern := regexp.MustCompile(`/products/([^/]+)`)
		if m := idPattern.FindStringSubmatch(urlStr); len(m) > 1 {
			release.ID = m[1]
		}
	}
	if image, ok := data["image"].(string); ok {
		release.ImageURL = image
	}
	if byArtist, ok := data["byArtist"].(map[string]interface{}); ok {
		if name, ok := byArtist["name"].(string); ok {
			release.Artists = []string{name}
		}
	}
	if offers, ok := data["offers"].(map[string]interface{}); ok {
		if price, ok := offers["price"].(string); ok {
			release.Price = price
		} else if priceNum, ok := offers["price"].(float64); ok {
			release.Price = fmt.Sprintf("%.2f", priceNum)
		}
	}
	if desc, ok := data["description"].(string); ok {
		release.Description = desc
	}

	return release
}

// parseArtistsFromHTML extracts artists from HTML content
func (c *BoomkatClient) parseArtistsFromHTML(html string) []BoomkatArtist {
	var artists []BoomkatArtist

	// Pattern: /artists/artist-slug
	artistPattern := regexp.MustCompile(`<a[^>]*href="/artists/([^/"]+)"[^>]*>([^<]+)`)
	matches := artistPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) > 2 {
			slug := m[1]
			if !seen[slug] {
				seen[slug] = true
				artist := BoomkatArtist{
					ID:   slug,
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/artists/%s", boomkatBaseURL, slug),
				}
				artists = append(artists, artist)
			}
		}
	}

	return artists
}

// parseLabelsFromHTML extracts labels from HTML content
func (c *BoomkatClient) parseLabelsFromHTML(html string) []BoomkatLabel {
	var labels []BoomkatLabel

	// Pattern: /labels/label-slug
	labelPattern := regexp.MustCompile(`<a[^>]*href="/labels/([^/"]+)"[^>]*>([^<]+)`)
	matches := labelPattern.FindAllStringSubmatch(html, -1)

	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) > 2 {
			slug := m[1]
			if !seen[slug] {
				seen[slug] = true
				label := BoomkatLabel{
					ID:   slug,
					Name: strings.TrimSpace(m[2]),
					URL:  fmt.Sprintf("%s/labels/%s", boomkatBaseURL, slug),
				}
				labels = append(labels, label)
			}
		}
	}

	return labels
}

// parseReleaseDetailsFromHTML extracts detailed release info from a release page
func (c *BoomkatClient) parseReleaseDetailsFromHTML(html string, releaseSlug string) *BoomkatRelease {
	release := &BoomkatRelease{
		ID:  releaseSlug,
		URL: fmt.Sprintf("%s/products/%s", boomkatBaseURL, releaseSlug),
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

	// Extract title from h1
	titlePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := titlePattern.FindStringSubmatch(html); len(m) > 1 {
		release.Title = strings.TrimSpace(m[1])
	}

	// Extract genre
	genrePattern := regexp.MustCompile(`class="[^"]*genre[^"]*"[^>]*>([^<]+)`)
	if m := genrePattern.FindStringSubmatch(html); len(m) > 1 {
		release.Genre = strings.TrimSpace(m[1])
	}

	// Extract label
	labelPattern := regexp.MustCompile(`class="[^"]*label[^"]*"[^>]*><a[^>]*>([^<]+)`)
	if m := labelPattern.FindStringSubmatch(html); len(m) > 1 {
		release.Label = strings.TrimSpace(m[1])
	}

	// Extract catalog number
	catPattern := regexp.MustCompile(`Cat(?:alog)?[:\s#]*([A-Z0-9-]+)`)
	if m := catPattern.FindStringSubmatch(html); len(m) > 1 {
		release.CatalogNum = m[1]
	}

	// Extract format
	formatPattern := regexp.MustCompile(`(?i)(vinyl|cd|digital|lp|12"|10")`)
	if m := formatPattern.FindStringSubmatch(html); len(m) > 1 {
		release.Format = m[1]
	}

	// Extract rating (Essential, Recommended)
	ratingPattern := regexp.MustCompile(`class="[^"]*rating[^"]*"[^>]*>([^<]+)`)
	if m := ratingPattern.FindStringSubmatch(html); len(m) > 1 {
		release.Rating = strings.TrimSpace(m[1])
	}

	// Parse tracks from tracklist
	release.Tracks = c.parseTracksFromHTML(html)

	return release
}

// parseTracksFromHTML extracts tracks from a release page
func (c *BoomkatClient) parseTracksFromHTML(html string) []BoomkatTrack {
	var tracks []BoomkatTrack

	// Look for tracklist items
	trackPattern := regexp.MustCompile(`class="[^"]*track[^"]*"[^>]*>.*?(\d+)[.\s]*([^<]+)`)
	matches := trackPattern.FindAllStringSubmatch(html, -1)

	for _, m := range matches {
		if len(m) > 2 {
			trackNum, _ := strconv.Atoi(m[1])
			track := BoomkatTrack{
				TrackNumber: trackNum,
				Title:       strings.TrimSpace(m[2]),
			}
			tracks = append(tracks, track)
		}
	}

	return tracks
}

// parseArtistDetailsFromHTML extracts detailed artist info from an artist page
func (c *BoomkatClient) parseArtistDetailsFromHTML(html string, artistSlug string) *BoomkatArtist {
	artist := &BoomkatArtist{
		ID:  artistSlug,
		URL: fmt.Sprintf("%s/artists/%s", boomkatBaseURL, artistSlug),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Name = strings.TrimSpace(m[1])
	}

	// Extract bio
	bioPattern := regexp.MustCompile(`class="[^"]*artist-bio[^"]*"[^>]*>(.*?)</div>`)
	if m := bioPattern.FindStringSubmatch(html); len(m) > 1 {
		artist.Bio = strings.TrimSpace(boomkatStripHTMLTags(m[1]))
	}

	return artist
}

// parseLabelDetailsFromHTML extracts detailed label info from a label page
func (c *BoomkatClient) parseLabelDetailsFromHTML(html string, labelSlug string) *BoomkatLabel {
	label := &BoomkatLabel{
		ID:  labelSlug,
		URL: fmt.Sprintf("%s/labels/%s", boomkatBaseURL, labelSlug),
	}

	// Extract name from title or header
	namePattern := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	if m := namePattern.FindStringSubmatch(html); len(m) > 1 {
		label.Name = strings.TrimSpace(m[1])
	}

	// Extract description
	descPattern := regexp.MustCompile(`class="[^"]*label-description[^"]*"[^>]*>(.*?)</div>`)
	if m := descPattern.FindStringSubmatch(html); len(m) > 1 {
		label.Description = strings.TrimSpace(boomkatStripHTMLTags(m[1]))
	}

	return label
}

// boomkatStripHTMLTags removes HTML tags from a string
func boomkatStripHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}
