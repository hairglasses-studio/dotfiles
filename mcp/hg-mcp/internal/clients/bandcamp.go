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
	"strings"
	"sync"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	bandcampBaseURL = "https://bandcamp.com"
)

// BandcampClient provides access to Bandcamp data
type BandcampClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]interface{}
}

// BandcampArtist represents an artist/label
type BandcampArtist struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	URL        string            `json:"url"`
	Location   string            `json:"location,omitempty"`
	Bio        string            `json:"bio,omitempty"`
	ImageURL   string            `json:"image_url,omitempty"`
	BannerURL  string            `json:"banner_url,omitempty"`
	AlbumCount int               `json:"album_count"`
	TrackCount int               `json:"track_count"`
	Links      map[string]string `json:"links,omitempty"`
	IsLabel    bool              `json:"is_label"`
	Subdomain  string            `json:"subdomain"`
}

// BandcampAlbum represents an album or EP
type BandcampAlbum struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	URL         string          `json:"url"`
	Artist      string          `json:"artist"`
	ArtistURL   string          `json:"artist_url"`
	ArtworkURL  string          `json:"artwork_url,omitempty"`
	ReleaseDate string          `json:"release_date,omitempty"`
	TrackCount  int             `json:"track_count"`
	Tracks      []BandcampTrack `json:"tracks,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price,omitempty"`
	Currency    string          `json:"currency,omitempty"`
	IsFree      bool            `json:"is_free"`
	About       string          `json:"about,omitempty"`
}

// BandcampTrack represents a single track
type BandcampTrack struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
	Duration   int     `json:"duration"` // seconds
	TrackNum   int     `json:"track_num"`
	Album      string  `json:"album,omitempty"`
	Artist     string  `json:"artist"`
	ArtworkURL string  `json:"artwork_url,omitempty"`
	StreamURL  string  `json:"stream_url,omitempty"`
	Lyrics     string  `json:"lyrics,omitempty"`
	Price      float64 `json:"price,omitempty"`
	IsFree     bool    `json:"is_free"`
}

// BandcampSearchResult represents search results
type BandcampSearchResult struct {
	Query   string           `json:"query"`
	Type    string           `json:"type"` // all, artist, album, track
	Artists []BandcampArtist `json:"artists,omitempty"`
	Albums  []BandcampAlbum  `json:"albums,omitempty"`
	Tracks  []BandcampTrack  `json:"tracks,omitempty"`
	Total   int              `json:"total"`
}

// BandcampStatus represents connection status
type BandcampStatus struct {
	Connected      bool   `json:"connected"`
	HasBandcampDL  bool   `json:"has_bandcamp_dl"`
	HasYtDlp       bool   `json:"has_yt_dlp"`
	BandcampDLPath string `json:"bandcamp_dl_path,omitempty"`
	YtDlpPath      string `json:"yt_dlp_path,omitempty"`
}

// BandcampHealth represents health status
type BandcampHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasDownloader   bool     `json:"has_downloader"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// BandcampDownloadResult represents download result
type BandcampDownloadResult struct {
	Success    bool     `json:"success"`
	URL        string   `json:"url"`
	OutputPath string   `json:"output_path,omitempty"`
	Files      []string `json:"files,omitempty"`
	Error      string   `json:"error,omitempty"`
	Tool       string   `json:"tool"` // bandcamp-dl or yt-dlp
}

// NewBandcampClient creates a new Bandcamp client
func NewBandcampClient() (*BandcampClient, error) {
	return &BandcampClient{
		httpClient: httpclient.Standard(),
		cache: make(map[string]interface{}),
	}, nil
}

// GetStatus returns the client status
func (c *BandcampClient) GetStatus(ctx context.Context) (*BandcampStatus, error) {
	status := &BandcampStatus{
		Connected: true,
	}

	// Check for bandcamp-dl
	if path, err := exec.LookPath("bandcamp-dl"); err == nil {
		status.HasBandcampDL = true
		status.BandcampDLPath = path
	}

	// Check for yt-dlp
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		status.HasYtDlp = true
		status.YtDlpPath = path
	}

	// Test connectivity
	req, err := http.NewRequestWithContext(ctx, "HEAD", bandcampBaseURL, nil)
	if err != nil {
		status.Connected = false
		return status, nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		status.Connected = false
		return status, nil
	}
	defer resp.Body.Close()

	status.Connected = resp.StatusCode == http.StatusOK

	return status, nil
}

// GetHealth returns health status with recommendations
func (c *BandcampClient) GetHealth(ctx context.Context) (*BandcampHealth, error) {
	status, _ := c.GetStatus(ctx)

	health := &BandcampHealth{
		Score:     100,
		Status:    "healthy",
		Connected: status.Connected,
	}

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Cannot connect to Bandcamp")
		health.Recommendations = append(health.Recommendations, "Check internet connection")
	}

	if !status.HasBandcampDL && !status.HasYtDlp {
		health.Score -= 30
		health.HasDownloader = false
		health.Issues = append(health.Issues, "No download tools available")
		health.Recommendations = append(health.Recommendations, "Install bandcamp-dl: pip install bandcamp-dl")
		health.Recommendations = append(health.Recommendations, "Or install yt-dlp: pip install yt-dlp")
	} else {
		health.HasDownloader = true
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

// Search searches Bandcamp for artists, albums, or tracks
func (c *BandcampClient) Search(ctx context.Context, query string, searchType string) (*BandcampSearchResult, error) {
	result := &BandcampSearchResult{
		Query: query,
		Type:  searchType,
	}

	// Build search URL
	searchURL := fmt.Sprintf("%s/search?q=%s", bandcampBaseURL, url.QueryEscape(query))
	if searchType != "" && searchType != "all" {
		searchURL += fmt.Sprintf("&item_type=%s", searchType)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse search results from HTML
	result.Artists, result.Albums, result.Tracks = c.parseSearchResults(string(body), searchType)
	result.Total = len(result.Artists) + len(result.Albums) + len(result.Tracks)

	return result, nil
}

// parseSearchResults extracts search results from HTML
func (c *BandcampClient) parseSearchResults(html string, searchType string) ([]BandcampArtist, []BandcampAlbum, []BandcampTrack) {
	var artists []BandcampArtist
	var albums []BandcampAlbum
	var tracks []BandcampTrack

	// Pattern for search result items
	itemPattern := regexp.MustCompile(`<li class="searchresult[^"]*"[^>]*data-search="([^"]*)"[^>]*>`)

	// Pattern for result type
	typePattern := regexp.MustCompile(`<div class="result-info">\s*<div class="itemtype">(\w+)</div>`)

	// Pattern for heading/title
	headingPattern := regexp.MustCompile(`<div class="heading">\s*<a href="([^"]+)"[^>]*>([^<]+)</a>`)

	// Pattern for subhead (artist name for albums/tracks)
	subheadPattern := regexp.MustCompile(`<div class="subhead">\s*(?:by\s*)?<a href="([^"]+)"[^>]*>([^<]+)</a>`)

	items := itemPattern.FindAllStringSubmatchIndex(html, -1)

	for i := 0; i < len(items); i++ {
		startIdx := items[i][0]
		endIdx := len(html)
		if i+1 < len(items) {
			endIdx = items[i+1][0]
		}

		itemHTML := html[startIdx:endIdx]

		// Get item type
		typeMatch := typePattern.FindStringSubmatch(itemHTML)
		if len(typeMatch) < 2 {
			continue
		}
		itemType := strings.ToLower(typeMatch[1])

		// Get heading (title/name and URL)
		headingMatch := headingPattern.FindStringSubmatch(itemHTML)
		if len(headingMatch) < 3 {
			continue
		}
		itemURL := headingMatch[1]
		itemName := strings.TrimSpace(headingMatch[2])

		switch itemType {
		case "artist", "label":
			if searchType != "" && searchType != "all" && searchType != "artist" {
				continue
			}
			artist := BandcampArtist{
				ID:      extractSubdomain(itemURL),
				Name:    itemName,
				URL:     itemURL,
				IsLabel: itemType == "label",
			}
			artists = append(artists, artist)

		case "album":
			if searchType != "" && searchType != "all" && searchType != "album" {
				continue
			}
			album := BandcampAlbum{
				ID:    extractAlbumID(itemURL),
				Title: itemName,
				URL:   itemURL,
			}
			// Get artist from subhead
			if subheadMatch := subheadPattern.FindStringSubmatch(itemHTML); len(subheadMatch) >= 3 {
				album.ArtistURL = subheadMatch[1]
				album.Artist = strings.TrimSpace(subheadMatch[2])
			}
			albums = append(albums, album)

		case "track":
			if searchType != "" && searchType != "all" && searchType != "track" {
				continue
			}
			track := BandcampTrack{
				ID:    extractTrackID(itemURL),
				Title: itemName,
				URL:   itemURL,
			}
			// Get artist from subhead
			if subheadMatch := subheadPattern.FindStringSubmatch(itemHTML); len(subheadMatch) >= 3 {
				track.Artist = strings.TrimSpace(subheadMatch[2])
			}
			tracks = append(tracks, track)
		}
	}

	return artists, albums, tracks
}

// GetArtist gets artist details
func (c *BandcampClient) GetArtist(ctx context.Context, artistURL string) (*BandcampArtist, error) {
	// Normalize URL
	if !strings.HasPrefix(artistURL, "http") {
		artistURL = "https://" + artistURL + ".bandcamp.com"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", artistURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return c.parseArtistPage(string(body), artistURL)
}

// parseArtistPage extracts artist info from HTML
func (c *BandcampClient) parseArtistPage(html string, artistURL string) (*BandcampArtist, error) {
	artist := &BandcampArtist{
		URL:       artistURL,
		Subdomain: extractSubdomain(artistURL),
	}

	// Extract name from title or band-name element
	if nameMatch := regexp.MustCompile(`<p id="band-name-location"[^>]*>\s*<span class="title">([^<]+)</span>`).FindStringSubmatch(html); len(nameMatch) >= 2 {
		artist.Name = strings.TrimSpace(nameMatch[1])
	} else if titleMatch := regexp.MustCompile(`<title>([^|<]+)`).FindStringSubmatch(html); len(titleMatch) >= 2 {
		artist.Name = strings.TrimSpace(titleMatch[1])
	}

	// Extract location
	if locMatch := regexp.MustCompile(`<span class="location[^"]*">([^<]+)</span>`).FindStringSubmatch(html); len(locMatch) >= 2 {
		artist.Location = strings.TrimSpace(locMatch[1])
	}

	// Extract bio
	if bioMatch := regexp.MustCompile(`<meta name="description" content="([^"]+)"`).FindStringSubmatch(html); len(bioMatch) >= 2 {
		artist.Bio = strings.TrimSpace(bioMatch[1])
	}

	// Extract image
	if imgMatch := regexp.MustCompile(`<img[^>]*class="[^"]*band-photo[^"]*"[^>]*src="([^"]+)"`).FindStringSubmatch(html); len(imgMatch) >= 2 {
		artist.ImageURL = imgMatch[1]
	}

	// Count albums
	albumMatches := regexp.MustCompile(`<li class="music-grid-item[^"]*album[^"]*"`).FindAllString(html, -1)
	artist.AlbumCount = len(albumMatches)

	// Count tracks (singles)
	trackMatches := regexp.MustCompile(`<li class="music-grid-item[^"]*track[^"]*"`).FindAllString(html, -1)
	artist.TrackCount = len(trackMatches)

	artist.ID = artist.Subdomain

	return artist, nil
}

// GetAlbum gets album details including tracks
func (c *BandcampClient) GetAlbum(ctx context.Context, albumURL string) (*BandcampAlbum, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", albumURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch album page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return c.parseAlbumPage(string(body), albumURL)
}

// parseAlbumPage extracts album info from HTML
func (c *BandcampClient) parseAlbumPage(html string, albumURL string) (*BandcampAlbum, error) {
	album := &BandcampAlbum{
		URL: albumURL,
		ID:  extractAlbumID(albumURL),
	}

	// Try to extract TralbumData JSON from the page
	tralbumPattern := regexp.MustCompile(`data-tralbum="([^"]+)"`)
	if match := tralbumPattern.FindStringSubmatch(html); len(match) >= 2 {
		// Decode HTML entities and parse JSON
		jsonStr := strings.ReplaceAll(match[1], "&quot;", "\"")
		jsonStr = strings.ReplaceAll(jsonStr, "&amp;", "&")

		var tralbumData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &tralbumData); err == nil {
			if current, ok := tralbumData["current"].(map[string]interface{}); ok {
				if title, ok := current["title"].(string); ok {
					album.Title = title
				}
				if artist, ok := current["artist"].(string); ok {
					album.Artist = artist
				}
				if about, ok := current["about"].(string); ok {
					album.About = about
				}
				if releaseDate, ok := current["release_date"].(string); ok {
					album.ReleaseDate = releaseDate
				}
				if minPrice, ok := current["minimum_price"].(float64); ok {
					album.Price = minPrice
					album.IsFree = minPrice == 0
				}
			}

			// Parse tracks
			if trackInfo, ok := tralbumData["trackinfo"].([]interface{}); ok {
				for i, t := range trackInfo {
					if track, ok := t.(map[string]interface{}); ok {
						bcTrack := BandcampTrack{
							TrackNum: i + 1,
							Artist:   album.Artist,
						}
						if title, ok := track["title"].(string); ok {
							bcTrack.Title = title
						}
						if duration, ok := track["duration"].(float64); ok {
							bcTrack.Duration = int(duration)
						}
						if streamURL, ok := track["file"].(map[string]interface{}); ok {
							if mp3URL, ok := streamURL["mp3-128"].(string); ok {
								bcTrack.StreamURL = mp3URL
							}
						}
						if trackID, ok := track["track_id"].(float64); ok {
							bcTrack.ID = fmt.Sprintf("%d", int(trackID))
						}
						album.Tracks = append(album.Tracks, bcTrack)
					}
				}
				album.TrackCount = len(album.Tracks)
			}
		}
	}

	// Fallback to HTML parsing if JSON not found
	if album.Title == "" {
		if titleMatch := regexp.MustCompile(`<h2 class="trackTitle"[^>]*>([^<]+)</h2>`).FindStringSubmatch(html); len(titleMatch) >= 2 {
			album.Title = strings.TrimSpace(titleMatch[1])
		}
	}

	if album.Artist == "" {
		if artistMatch := regexp.MustCompile(`<span itemprop="byArtist"[^>]*>\s*<a[^>]*>([^<]+)</a>`).FindStringSubmatch(html); len(artistMatch) >= 2 {
			album.Artist = strings.TrimSpace(artistMatch[1])
		}
	}

	// Extract artwork
	if artMatch := regexp.MustCompile(`<a class="popupImage"[^>]*href="([^"]+)"`).FindStringSubmatch(html); len(artMatch) >= 2 {
		album.ArtworkURL = artMatch[1]
	}

	// Extract tags
	tagPattern := regexp.MustCompile(`<a class="tag"[^>]*>([^<]+)</a>`)
	tagMatches := tagPattern.FindAllStringSubmatch(html, -1)
	for _, match := range tagMatches {
		if len(match) >= 2 {
			album.Tags = append(album.Tags, strings.TrimSpace(match[1]))
		}
	}

	return album, nil
}

// Download downloads an album or track using bandcamp-dl or yt-dlp
func (c *BandcampClient) Download(ctx context.Context, bcURL string, outputDir string) (*BandcampDownloadResult, error) {
	result := &BandcampDownloadResult{
		URL:        bcURL,
		OutputPath: outputDir,
	}

	// Try bandcamp-dl first
	if path, err := exec.LookPath("bandcamp-dl"); err == nil {
		result.Tool = "bandcamp-dl"
		args := []string{bcURL}
		if outputDir != "" {
			args = append(args, "--base-dir", outputDir)
		}

		cmd := exec.CommandContext(ctx, path, args...)
		output, err := cmd.CombinedOutput()
		if err == nil {
			result.Success = true
			// Parse output for downloaded files
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Saved:") || strings.HasSuffix(line, ".mp3") || strings.HasSuffix(line, ".flac") {
					result.Files = append(result.Files, strings.TrimSpace(line))
				}
			}
			return result, nil
		}
		// If bandcamp-dl fails, try yt-dlp
		result.Error = string(output)
	}

	// Fallback to yt-dlp
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		result.Tool = "yt-dlp"
		args := []string{
			"-x", // Extract audio
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"--embed-thumbnail",
			"--add-metadata",
		}
		if outputDir != "" {
			args = append(args, "-o", outputDir+"/%(artist)s - %(title)s.%(ext)s")
		}
		args = append(args, bcURL)

		cmd := exec.CommandContext(ctx, path, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			result.Success = false
			result.Error = string(output)
			return result, fmt.Errorf("download failed: %s", output)
		}

		result.Success = true
		// Parse output for downloaded files
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Destination:") {
				parts := strings.SplitN(line, "Destination:", 2)
				if len(parts) == 2 {
					result.Files = append(result.Files, strings.TrimSpace(parts[1]))
				}
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("no download tools available (need bandcamp-dl or yt-dlp)")
}

// GetTags returns popular Bandcamp tags
func (c *BandcampClient) GetTags(ctx context.Context) ([]string, error) {
	return []string{
		"electronic", "ambient", "experimental", "hip-hop-rap",
		"rock", "punk", "metal", "jazz", "soul", "r-b-soul",
		"world", "folk", "acoustic", "indie", "pop",
		"soundtrack", "classical", "dance", "house", "techno",
		"drum-bass", "dubstep", "trap", "lo-fi", "noise",
		"shoegaze", "post-punk", "new-wave", "synth", "vaporwave",
	}, nil
}

// GetTagReleases gets releases for a specific tag
func (c *BandcampClient) GetTagReleases(ctx context.Context, tag string, page int) ([]BandcampAlbum, error) {
	tagURL := fmt.Sprintf("%s/tag/%s?page=%d&sort_field=date", bandcampBaseURL, tag, page)

	req, err := http.NewRequestWithContext(ctx, "GET", tagURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tag page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return c.parseTagPage(string(body), tag)
}

// parseTagPage extracts albums from a tag browse page
func (c *BandcampClient) parseTagPage(html string, tag string) ([]BandcampAlbum, error) {
	var albums []BandcampAlbum

	// Pattern for album items on tag pages
	itemPattern := regexp.MustCompile(`<li class="item"[^>]*>\s*<a href="([^"]+)"[^>]*>\s*<div class="art">\s*<img[^>]*src="([^"]+)"[^>]*>\s*</div>\s*<div class="info">\s*<div class="heading">([^<]+)</div>\s*<div class="subhead">[^<]*by\s*([^<]+)</div>`)

	matches := itemPattern.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) >= 5 {
			album := BandcampAlbum{
				URL:        match[1],
				ArtworkURL: match[2],
				Title:      strings.TrimSpace(match[3]),
				Artist:     strings.TrimSpace(match[4]),
				Tags:       []string{tag},
				ID:         extractAlbumID(match[1]),
			}
			albums = append(albums, album)
		}
	}

	return albums, nil
}

// Helper functions

func extractSubdomain(bcURL string) string {
	u, err := url.Parse(bcURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(u.Host, ".")
	if len(parts) > 0 && parts[0] != "www" {
		return parts[0]
	}
	return ""
}

func extractAlbumID(bcURL string) string {
	u, err := url.Parse(bcURL)
	if err != nil {
		return ""
	}
	// Format: https://artist.bandcamp.com/album/album-name
	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "album" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return u.Path
}

func extractTrackID(bcURL string) string {
	u, err := url.Parse(bcURL)
	if err != nil {
		return ""
	}
	// Format: https://artist.bandcamp.com/track/track-name
	parts := strings.Split(u.Path, "/")
	for i, part := range parts {
		if part == "track" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return u.Path
}
