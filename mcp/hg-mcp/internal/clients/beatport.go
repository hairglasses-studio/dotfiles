// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

const (
	beatportAPIBase  = "https://api.beatport.com/v4"
	beatportAuthURL  = "https://api.beatport.com/v4/auth"
	beatportDocsURL  = "https://api.beatport.com/v4/docs/"
	beatportLoginURL = "https://api.beatport.com/v4/auth/login/"
	// Hardcoded client_id from Beatport (same as beatportdl uses)
	beatportClientID = "ryZ8LuyQVPqbK2mBX2Hwt4qSMtnWuTYSqBPO92yQ"
)

// BeatportClient provides access to Beatport API
type BeatportClient struct {
	httpClient     *http.Client
	downloadClient *http.Client // Separate client with longer timeout for file downloads
	accessToken    string
	refreshToken   string
	clientID       string
	tokenExpiry    time.Time
	username       string
	password       string
	mu             sync.RWMutex
}

// BeatportTrack represents a track from Beatport API
type BeatportTrack struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	MixName     string           `json:"mix_name"`
	BPM         float64          `json:"bpm"`
	Key         *BeatportKey     `json:"key"`
	Genre       *BeatportGenre   `json:"genre"`
	SubGenre    *BeatportGenre   `json:"sub_genre"`
	Label       *BeatportLabel   `json:"label"`
	Release     *BeatportRelease `json:"release"`
	Artists     []BeatportArtist `json:"artists"`
	Remixers    []BeatportArtist `json:"remixers"`
	ISRC        string           `json:"isrc"`
	PublishDate string           `json:"publish_date"`
	LengthMS    int              `json:"length_ms"`
	Image       *BeatportImage   `json:"image"`
	URL         string           `json:"url"`
	Exclusive   bool             `json:"exclusive"`
}

// BeatportKey represents musical key information
type BeatportKey struct {
	ID      int              `json:"id"`
	Name    string           `json:"name"` // "Eb Major", "A Minor"
	Camelot *BeatportCamelot `json:"camelot"`
}

// BeatportCamelot represents Camelot wheel notation
type BeatportCamelot struct {
	Key string `json:"key"` // "5A", "8B"
}

// BeatportGenre represents a genre
type BeatportGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// BeatportLabel represents a record label
type BeatportLabel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// BeatportRelease represents a release/album
type BeatportRelease struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	ReleaseDate string         `json:"release_date"`
	Label       *BeatportLabel `json:"label"`
	CatalogNum  string         `json:"catalog_number"`
}

// BeatportArtist represents an artist
type BeatportArtist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role,omitempty"` // "primary", "remixer", etc.
}

// BeatportImage represents artwork
type BeatportImage struct {
	ID         int    `json:"id"`
	URI        string `json:"uri"`
	DynamicURI string `json:"dynamic_uri"` // Template with {w}x{h}
}

// BeatportChart represents a chart
type BeatportChart struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Genre       *BeatportGenre `json:"genre"`
	URL         string         `json:"url"`
}

// BeatportSearchResponse represents search API response
type BeatportSearchResponse struct {
	Tracks   []BeatportTrack   `json:"tracks"`
	Releases []BeatportRelease `json:"releases"`
	Artists  []BeatportArtist  `json:"artists"`
	Labels   []BeatportLabel   `json:"labels"`
}

// BeatportTokenResponse represents OAuth token response
type BeatportTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// NewBeatportClient creates a new Beatport client
func NewBeatportClient() (*BeatportClient, error) {
	cfg := config.GetOrLoad()
	username := cfg.BeatportUsername
	password := cfg.BeatportPassword

	// Check for pre-existing tokens
	accessToken := cfg.BeatportAccessToken
	refreshToken := cfg.BeatportRefreshToken

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &BeatportClient{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: httpclient.Transport(),
			Jar:       jar,
		},
		downloadClient: httpclient.WithTimeout(10 * time.Minute),
		username:       username,
		password:       password,
		accessToken:    accessToken,
		refreshToken:   refreshToken,
	}

	return client, nil
}

// NewBeatportClientWithCredentials creates a new Beatport client with credentials
func NewBeatportClientWithCredentials(username, password string) (*BeatportClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &BeatportClient{
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: httpclient.Transport(),
			Jar:       jar,
		},
		downloadClient: httpclient.WithTimeout(10 * time.Minute),
		username:       username,
		password:       password,
	}, nil
}

// SetTokens sets existing OAuth tokens
func (c *BeatportClient) SetTokens(accessToken, refreshToken string, expiry time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	c.tokenExpiry = expiry
}

// GetTokens returns current OAuth tokens
func (c *BeatportClient) GetTokens() (string, string, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, c.refreshToken, c.tokenExpiry
}

// EnsureAuthenticated ensures we have valid tokens
func (c *BeatportClient) EnsureAuthenticated(ctx context.Context) error {
	c.mu.RLock()
	hasToken := c.accessToken != ""
	expired := time.Now().After(c.tokenExpiry)
	c.mu.RUnlock()

	if hasToken && !expired {
		return nil
	}

	// Try to refresh if we have a refresh token
	c.mu.RLock()
	hasRefresh := c.refreshToken != ""
	c.mu.RUnlock()

	if hasRefresh {
		if err := c.RefreshTokens(ctx); err == nil {
			return nil
		}
	}

	// Full authentication
	return c.Authenticate(ctx)
}

// GetPublicClientID returns the Beatport client ID
func (c *BeatportClient) GetPublicClientID(ctx context.Context) (string, error) {
	if c.clientID != "" {
		return c.clientID, nil
	}
	// Use hardcoded client_id (same as beatportdl)
	c.clientID = beatportClientID
	return c.clientID, nil
}

// Authenticate performs full OAuth authentication flow
func (c *BeatportClient) Authenticate(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("BEATPORT_USERNAME and BEATPORT_PASSWORD environment variables required")
	}

	// Get public client ID
	clientID, err := c.GetPublicClientID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client ID: %w", err)
	}

	// Step 1: Login to get session (uses JSON, not form data)
	loginPayload := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	loginBody, err := json.Marshal(loginPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal login payload: %w", err)
	}

	loginReq, err := http.NewRequestWithContext(ctx, "POST", beatportLoginURL, strings.NewReader(string(loginBody)))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("User-Agent", "CR8-MCP/1.0")

	loginResp, err := c.httpClient.Do(loginReq)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode >= 400 {
		body, _ := io.ReadAll(loginResp.Body)
		return fmt.Errorf("login failed (status %d): %s", loginResp.StatusCode, string(body))
	}

	// Step 2: Get authorization code (without following redirects)
	authParams := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
	}

	authURL := fmt.Sprintf("%s/o/authorize/?%s", beatportAuthURL, authParams.Encode())
	authReq, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	authReq.Header.Set("User-Agent", "CR8-MCP/1.0")

	// Use a client that doesn't follow redirects to capture the Location header
	noRedirectClient := &http.Client{
		Transport: httpclient.Transport(),
		Jar:       c.httpClient.Jar, // Share cookies
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	authResp, err := noRedirectClient.Do(authReq)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer authResp.Body.Close()

	// Extract authorization code from Location header
	var authCode string
	location := authResp.Header.Get("Location")
	if location != "" {
		if locURL, err := url.Parse(location); err == nil {
			authCode = locURL.Query().Get("code")
		}
	}

	if authCode == "" {
		return fmt.Errorf("failed to obtain authorization code (status: %d)", authResp.StatusCode)
	}

	// Step 3: Exchange code for tokens
	tokenData := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {authCode},
		"client_id":  {clientID},
	}

	tokenReq, err := http.NewRequestWithContext(ctx, "POST", beatportAuthURL+"/o/token/", strings.NewReader(tokenData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("User-Agent", "CR8-MCP/1.0")

	tokenResp, err := c.httpClient.Do(tokenReq)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode >= 400 {
		body, _ := io.ReadAll(tokenResp.Body)
		return fmt.Errorf("token exchange failed (status %d): %s", tokenResp.StatusCode, string(body))
	}

	var tokens BeatportTokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.mu.Lock()
	c.accessToken = tokens.AccessToken
	c.refreshToken = tokens.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	c.mu.Unlock()

	return nil
}

// RefreshTokens refreshes the access token using refresh token
func (c *BeatportClient) RefreshTokens(ctx context.Context) error {
	c.mu.RLock()
	refreshToken := c.refreshToken
	clientID := c.clientID
	c.mu.RUnlock()

	if refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	if clientID == "" {
		var err error
		clientID, err = c.GetPublicClientID(ctx)
		if err != nil {
			return err
		}
	}

	tokenData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", beatportAuthURL+"/o/token/", strings.NewReader(tokenData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "CR8-MCP/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("token refresh failed (status %d)", resp.StatusCode)
	}

	var tokens BeatportTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	c.mu.Lock()
	c.accessToken = tokens.AccessToken
	if tokens.RefreshToken != "" {
		c.refreshToken = tokens.RefreshToken
	}
	c.tokenExpiry = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	c.mu.Unlock()

	return nil
}

// doAuthenticatedRequest performs an authenticated API request
func (c *BeatportClient) doAuthenticatedRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	if err := c.EnsureAuthenticated(ctx); err != nil {
		return nil, err
	}

	url := beatportAPIBase + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CR8-MCP/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		// Token expired, try refresh and retry
		if err := c.RefreshTokens(ctx); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		return c.doAuthenticatedRequest(ctx, method, endpoint)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// SearchTracks searches Beatport for tracks
func (c *BeatportClient) SearchTracks(ctx context.Context, query string, limit int) ([]BeatportTrack, error) {
	if limit <= 0 {
		limit = 10
	}

	params := url.Values{
		"q":        {query},
		"type":     {"tracks"},
		"per_page": {fmt.Sprintf("%d", limit)},
	}

	endpoint := "/catalog/search/?" + params.Encode()

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var response BeatportSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	return response.Tracks, nil
}

// GetTrack gets a single track by ID
func (c *BeatportClient) GetTrack(ctx context.Context, trackID int) (*BeatportTrack, error) {
	endpoint := fmt.Sprintf("/catalog/tracks/%d/", trackID)

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var track BeatportTrack
	if err := json.Unmarshal(body, &track); err != nil {
		return nil, fmt.Errorf("failed to parse track response: %w", err)
	}

	return &track, nil
}

// GetCharts gets Beatport charts, optionally filtered by genre
func (c *BeatportClient) GetCharts(ctx context.Context, genreSlug string) ([]BeatportChart, error) {
	endpoint := "/catalog/charts/"
	if genreSlug != "" {
		endpoint += "?genre=" + url.QueryEscape(genreSlug)
	}

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []BeatportChart `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse charts response: %w", err)
	}

	return response.Results, nil
}

// GetChartTracks gets tracks from a specific chart
func (c *BeatportClient) GetChartTracks(ctx context.Context, chartID int) ([]BeatportTrack, error) {
	endpoint := fmt.Sprintf("/catalog/charts/%d/tracks/", chartID)

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []BeatportTrack `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse chart tracks response: %w", err)
	}

	return response.Results, nil
}

// GetGenres gets available genres
func (c *BeatportClient) GetGenres(ctx context.Context) ([]BeatportGenre, error) {
	body, err := c.doAuthenticatedRequest(ctx, "GET", "/catalog/genres/")
	if err != nil {
		return nil, err
	}

	var response struct {
		Results []BeatportGenre `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse genres response: %w", err)
	}

	return response.Results, nil
}

// IsAuthenticated returns whether the client has valid tokens
func (c *BeatportClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken != "" && time.Now().Before(c.tokenExpiry)
}

// BeatportDownload represents a track download response
type BeatportDownload struct {
	Location      string `json:"location"`
	StreamQuality string `json:"stream_quality"`
}

// BeatportStream represents a track stream response
type BeatportStream struct {
	URL           string `json:"stream_url"`
	SampleStartMS int    `json:"sample_start_ms"`
	SampleEndMS   int    `json:"sample_end_ms"`
}

// DownloadQuality constants
const (
	QualityFLAC   = "lossless" // 44.1kHz FLAC (Professional)
	QualityAAC256 = "medium"   // 256kbps AAC (Professional)
	QualityAAC128 = "low"      // 128kbps AAC (Essential)
)

// GetTrackDownload gets a download URL for a track
func (c *BeatportClient) GetTrackDownload(ctx context.Context, trackID int, quality string) (*BeatportDownload, error) {
	if quality == "" {
		quality = QualityFLAC
	}

	endpoint := fmt.Sprintf("/catalog/tracks/%d/download/?quality=%s", trackID, url.QueryEscape(quality))

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var download BeatportDownload
	if err := json.Unmarshal(body, &download); err != nil {
		return nil, fmt.Errorf("failed to parse download response: %w", err)
	}

	return &download, nil
}

// GetTrackStream gets a stream URL for a track (for preview/streaming)
func (c *BeatportClient) GetTrackStream(ctx context.Context, trackID int) (*BeatportStream, error) {
	endpoint := fmt.Sprintf("/catalog/tracks/%d/stream/", trackID)

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var stream BeatportStream
	if err := json.Unmarshal(body, &stream); err != nil {
		return nil, fmt.Errorf("failed to parse stream response: %w", err)
	}

	return &stream, nil
}

// GetPlaylist gets a playlist by ID
func (c *BeatportClient) GetPlaylist(ctx context.Context, playlistID int) (*BeatportPlaylist, error) {
	endpoint := fmt.Sprintf("/catalog/playlists/%d/", playlistID)

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var playlist BeatportPlaylist
	if err := json.Unmarshal(body, &playlist); err != nil {
		return nil, fmt.Errorf("failed to parse playlist response: %w", err)
	}

	return &playlist, nil
}

// GetPlaylistTracks gets tracks from a playlist with pagination
func (c *BeatportClient) GetPlaylistTracks(ctx context.Context, playlistID int, page int) (*BeatportPlaylistTracksResponse, error) {
	if page < 1 {
		page = 1
	}

	endpoint := fmt.Sprintf("/catalog/playlists/%d/tracks/?page=%d&per_page=100", playlistID, page)

	body, err := c.doAuthenticatedRequest(ctx, "GET", endpoint)
	if err != nil {
		return nil, err
	}

	var response BeatportPlaylistTracksResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse playlist tracks response: %w", err)
	}

	return &response, nil
}

// GetAllPlaylistTracks fetches all tracks from a playlist (handles pagination)
func (c *BeatportClient) GetAllPlaylistTracks(ctx context.Context, playlistID int) ([]BeatportTrack, error) {
	var allTracks []BeatportTrack
	page := 1

	for {
		response, err := c.GetPlaylistTracks(ctx, playlistID, page)
		if err != nil {
			return nil, err
		}

		for _, item := range response.Results {
			allTracks = append(allTracks, item.Track)
		}

		if response.Next == "" || len(response.Results) == 0 {
			break
		}
		page++
	}

	return allTracks, nil
}

// BeatportPlaylist represents a playlist
type BeatportPlaylist struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	TrackCount  int      `json:"track_count"`
	Genres      []string `json:"genres"`
	BPMRange    []int    `json:"bpm_range"`
	LengthMS    int      `json:"length_ms"`
	CreatedDate string   `json:"created_date"`
	UpdatedDate string   `json:"updated_date"`
}

// BeatportPlaylistItem represents a track in a playlist
type BeatportPlaylistItem struct {
	ID       int           `json:"id"`
	Position int           `json:"position"`
	Track    BeatportTrack `json:"track"`
}

// BeatportPlaylistTracksResponse represents paginated playlist tracks
type BeatportPlaylistTracksResponse struct {
	Count    int                    `json:"count"`
	Next     string                 `json:"next"`
	Previous string                 `json:"previous"`
	Results  []BeatportPlaylistItem `json:"results"`
}

// DownloadTrackToFile downloads a track to a local file
func (c *BeatportClient) DownloadTrackToFile(ctx context.Context, trackID int, quality, destPath string) error {
	download, err := c.GetTrackDownload(ctx, trackID, quality)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	if download.Location == "" {
		return fmt.Errorf("no download location returned")
	}

	// Download the file using dedicated download client (longer timeout for large files)
	req, err := http.NewRequestWithContext(ctx, "GET", download.Location, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	req.Header.Set("User-Agent", "CR8-MCP/1.0")

	resp, err := c.downloadClient.Do(req)
	if err != nil {
		return fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// TrackToEnrichmentData converts a BeatportTrack to a map for CR8 enrichment
func TrackToEnrichmentData(track *BeatportTrack) map[string]interface{} {
	data := map[string]interface{}{
		"beatport_id":  track.ID,
		"beatport_url": track.URL,
		"bpm":          track.BPM,
		"source":       "beatport",
	}

	if track.Key != nil {
		data["key"] = track.Key.Name
		if track.Key.Camelot != nil {
			data["camelot_key"] = track.Key.Camelot.Key
		}
	}

	if track.Genre != nil {
		data["genre"] = track.Genre.Name
	}

	if track.SubGenre != nil {
		data["sub_genre"] = track.SubGenre.Name
	}

	if track.Label != nil {
		data["label"] = track.Label.Name
	}

	if track.Release != nil {
		if track.Release.ReleaseDate != "" {
			data["release_date"] = track.Release.ReleaseDate
		}
		if track.Release.CatalogNum != "" {
			data["catalog_number"] = track.Release.CatalogNum
		}
	}

	if track.MixName != "" && track.MixName != "Original Mix" {
		data["mix_name"] = track.MixName
	}

	if len(track.Remixers) > 0 {
		names := make([]string, len(track.Remixers))
		for i, r := range track.Remixers {
			names[i] = r.Name
		}
		data["remixer"] = strings.Join(names, ", ")
	}

	if track.ISRC != "" {
		data["isrc"] = track.ISRC
	}

	if track.Image != nil {
		if track.Image.DynamicURI != "" {
			// Use 500x500 artwork
			data["artwork_url"] = strings.ReplaceAll(strings.ReplaceAll(track.Image.DynamicURI, "{w}", "500"), "{h}", "500")
		} else if track.Image.URI != "" {
			data["artwork_url"] = track.Image.URI
		}
	}

	return data
}
