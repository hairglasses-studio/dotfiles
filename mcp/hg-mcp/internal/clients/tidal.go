// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// TidalClient provides access to Tidal API
type TidalClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
	countryCode  string
	mu           sync.RWMutex
}

// TidalStatus represents API connection status
type TidalStatus struct {
	Connected    bool   `json:"connected"`
	HasToken     bool   `json:"has_token"`
	TokenExpiry  string `json:"token_expiry,omitempty"`
	CountryCode  string `json:"country_code"`
	Subscription string `json:"subscription,omitempty"`
}

// TidalTrack represents a track
type TidalTrack struct {
	ID             int           `json:"id"`
	Title          string        `json:"title"`
	Duration       int           `json:"duration"` // seconds
	TrackNumber    int           `json:"trackNumber"`
	VolumeNumber   int           `json:"volumeNumber"`
	Version        string        `json:"version,omitempty"`
	ISRC           string        `json:"isrc,omitempty"`
	Explicit       bool          `json:"explicit"`
	AudioQuality   string        `json:"audioQuality"` // LOW, HIGH, LOSSLESS, HI_RES, HI_RES_LOSSLESS
	AudioModes     []string      `json:"audioModes,omitempty"`
	Copyright      string        `json:"copyright,omitempty"`
	Artists        []TidalArtist `json:"artists"`
	Album          *TidalAlbum   `json:"album,omitempty"`
	URL            string        `json:"url,omitempty"`
	Popularity     int           `json:"popularity"`
	AllowStreaming bool          `json:"allowStreaming"`
	StreamReady    bool          `json:"streamReady"`
	PremiumOnly    bool          `json:"premiumStreamingOnly"`
}

// TidalArtist represents an artist
type TidalArtist struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	Picture    string `json:"picture,omitempty"`
	Popularity int    `json:"popularity,omitempty"`
	URL        string `json:"url,omitempty"`
}

// TidalAlbum represents an album
type TidalAlbum struct {
	ID              int           `json:"id"`
	Title           string        `json:"title"`
	Cover           string        `json:"cover,omitempty"`
	Duration        int           `json:"duration,omitempty"`
	NumberOfTracks  int           `json:"numberOfTracks,omitempty"`
	NumberOfVolumes int           `json:"numberOfVolumes,omitempty"`
	ReleaseDate     string        `json:"releaseDate,omitempty"`
	Copyright       string        `json:"copyright,omitempty"`
	AudioQuality    string        `json:"audioQuality,omitempty"`
	AudioModes      []string      `json:"audioModes,omitempty"`
	Artists         []TidalArtist `json:"artists,omitempty"`
	URL             string        `json:"url,omitempty"`
	Explicit        bool          `json:"explicit"`
	UPC             string        `json:"upc,omitempty"`
}

// TidalPlaylist represents a playlist
type TidalPlaylist struct {
	UUID           string     `json:"uuid"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	Duration       int        `json:"duration"`
	NumberOfTracks int        `json:"numberOfTracks"`
	Created        string     `json:"created,omitempty"`
	LastUpdated    string     `json:"lastUpdated,omitempty"`
	Image          string     `json:"image,omitempty"`
	Creator        *TidalUser `json:"creator,omitempty"`
	URL            string     `json:"url,omitempty"`
	Public         bool       `json:"publicPlaylist"`
}

// TidalUser represents a user
type TidalUser struct {
	ID   int    `json:"id"`
	Name string `json:"name,omitempty"`
}

// TidalGenre represents a genre/category
type TidalGenre struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Image string `json:"image,omitempty"`
}

// TidalSearchResult represents search results
type TidalSearchResult struct {
	Tracks    []TidalTrack    `json:"tracks,omitempty"`
	Artists   []TidalArtist   `json:"artists,omitempty"`
	Albums    []TidalAlbum    `json:"albums,omitempty"`
	Playlists []TidalPlaylist `json:"playlists,omitempty"`
	TopHit    interface{}     `json:"topHit,omitempty"`
}

// TidalQualityInfo represents audio quality information
type TidalQualityInfo struct {
	AudioQuality string   `json:"audio_quality"`
	AudioModes   []string `json:"audio_modes"`
	HasMQA       bool     `json:"has_mqa"`
	HasDolby     bool     `json:"has_dolby_atmos"`
	Has360       bool     `json:"has_360_reality"`
	BitDepth     int      `json:"bit_depth,omitempty"`
	SampleRate   int      `json:"sample_rate,omitempty"`
}

// TidalHealth represents health status
type TidalHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasCredentials  bool     `json:"has_credentials"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// TidalNewRelease represents a new release item
type TidalNewRelease struct {
	ID           int           `json:"id"`
	Title        string        `json:"title"`
	Artists      []TidalArtist `json:"artists"`
	ReleaseDate  string        `json:"release_date"`
	Cover        string        `json:"cover,omitempty"`
	AudioQuality string        `json:"audio_quality"`
	Type         string        `json:"type"` // ALBUM, EP, SINGLE
	URL          string        `json:"url,omitempty"`
}

const (
	tidalAPIBase  = "https://api.tidal.com/v1"
	tidalAuthBase = "https://auth.tidal.com/v1/oauth2"
	tidalWebBase  = "https://tidal.com"
)

// NewTidalClient creates a new Tidal client
func NewTidalClient() (*TidalClient, error) {
	clientID := os.Getenv("TIDAL_CLIENT_ID")
	clientSecret := os.Getenv("TIDAL_CLIENT_SECRET")
	countryCode := os.Getenv("TIDAL_COUNTRY_CODE")

	if countryCode == "" {
		countryCode = "US"
	}

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("TIDAL_CLIENT_ID and TIDAL_CLIENT_SECRET required")
	}

	return &TidalClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		countryCode:  countryCode,
		httpClient: httpclient.Standard(),
	}, nil
}

// authenticate gets or refreshes OAuth token
func (c *TidalClient) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if token is still valid
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-60*time.Second)) {
		return nil
	}

	// Request new token using client credentials
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tidalAuthBase+"/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status: %d", resp.StatusCode)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	c.accessToken = authResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return nil
}

// doRequest performs an authenticated API request
func (c *TidalClient) doRequest(ctx context.Context, method, endpoint string, params url.Values) (*http.Response, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, err
	}

	// Add country code to all requests
	if params == nil {
		params = url.Values{}
	}
	params.Set("countryCode", c.countryCode)

	fullURL := tidalAPIBase + endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.mu.RUnlock()
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// GetStatus returns connection status
func (c *TidalClient) GetStatus(ctx context.Context) (*TidalStatus, error) {
	status := &TidalStatus{
		CountryCode: c.countryCode,
	}

	if err := c.authenticate(ctx); err != nil {
		status.Connected = false
		return status, nil
	}

	c.mu.RLock()
	status.Connected = c.accessToken != ""
	status.HasToken = c.accessToken != ""
	if !c.tokenExpiry.IsZero() {
		status.TokenExpiry = c.tokenExpiry.Format(time.RFC3339)
	}
	c.mu.RUnlock()

	return status, nil
}

// Health returns health check information
func (c *TidalClient) Health(ctx context.Context) (*TidalHealth, error) {
	health := &TidalHealth{
		HasCredentials: c.clientID != "" && c.clientSecret != "",
	}

	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Score = 0
		health.Status = "error"
		health.Issues = append(health.Issues, err.Error())
		return health, nil
	}

	health.Connected = status.Connected
	if health.Connected {
		health.Score = 100
		health.Status = "healthy"
	} else {
		health.Score = 0
		health.Status = "disconnected"
		health.Issues = append(health.Issues, "Not connected to Tidal API")
		if !health.HasCredentials {
			health.Recommendations = append(health.Recommendations, "Set TIDAL_CLIENT_ID and TIDAL_CLIENT_SECRET environment variables")
		}
	}

	return health, nil
}

// Search searches Tidal for tracks, artists, albums, and playlists
func (c *TidalClient) Search(ctx context.Context, query string, types []string, limit int) (*TidalSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	if len(types) > 0 {
		params.Set("types", strings.Join(types, ","))
	}

	resp, err := c.doRequest(ctx, "GET", "/search", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	// Parse the nested response structure
	var apiResp struct {
		Tracks struct {
			Items []TidalTrack `json:"items"`
		} `json:"tracks"`
		Artists struct {
			Items []TidalArtist `json:"items"`
		} `json:"artists"`
		Albums struct {
			Items []TidalAlbum `json:"items"`
		} `json:"albums"`
		Playlists struct {
			Items []TidalPlaylist `json:"items"`
		} `json:"playlists"`
		TopHit interface{} `json:"topHit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Add URLs to tracks
	for i := range apiResp.Tracks.Items {
		apiResp.Tracks.Items[i].URL = fmt.Sprintf("%s/track/%d", tidalWebBase, apiResp.Tracks.Items[i].ID)
	}
	for i := range apiResp.Artists.Items {
		apiResp.Artists.Items[i].URL = fmt.Sprintf("%s/artist/%d", tidalWebBase, apiResp.Artists.Items[i].ID)
	}
	for i := range apiResp.Albums.Items {
		apiResp.Albums.Items[i].URL = fmt.Sprintf("%s/album/%d", tidalWebBase, apiResp.Albums.Items[i].ID)
	}

	return &TidalSearchResult{
		Tracks:    apiResp.Tracks.Items,
		Artists:   apiResp.Artists.Items,
		Albums:    apiResp.Albums.Items,
		Playlists: apiResp.Playlists.Items,
		TopHit:    apiResp.TopHit,
	}, nil
}

// GetTrack retrieves a specific track
func (c *TidalClient) GetTrack(ctx context.Context, trackID int) (*TidalTrack, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tracks/%d", trackID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get track failed with status: %d", resp.StatusCode)
	}

	var track TidalTrack
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, fmt.Errorf("failed to decode track: %w", err)
	}

	track.URL = fmt.Sprintf("%s/track/%d", tidalWebBase, track.ID)
	return &track, nil
}

// GetAlbum retrieves a specific album
func (c *TidalClient) GetAlbum(ctx context.Context, albumID int) (*TidalAlbum, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/albums/%d", albumID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get album failed with status: %d", resp.StatusCode)
	}

	var album TidalAlbum
	if err := json.NewDecoder(resp.Body).Decode(&album); err != nil {
		return nil, fmt.Errorf("failed to decode album: %w", err)
	}

	album.URL = fmt.Sprintf("%s/album/%d", tidalWebBase, album.ID)
	return &album, nil
}

// GetAlbumTracks retrieves tracks from an album
func (c *TidalClient) GetAlbumTracks(ctx context.Context, albumID int) ([]TidalTrack, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/albums/%d/tracks", albumID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get album tracks failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalTrack `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode album tracks: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/track/%d", tidalWebBase, result.Items[i].ID)
	}

	return result.Items, nil
}

// GetArtist retrieves a specific artist
func (c *TidalClient) GetArtist(ctx context.Context, artistID int) (*TidalArtist, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/artists/%d", artistID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get artist failed with status: %d", resp.StatusCode)
	}

	var artist TidalArtist
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		return nil, fmt.Errorf("failed to decode artist: %w", err)
	}

	artist.URL = fmt.Sprintf("%s/artist/%d", tidalWebBase, artist.ID)
	return &artist, nil
}

// GetArtistAlbums retrieves albums from an artist
func (c *TidalClient) GetArtistAlbums(ctx context.Context, artistID int, limit int) ([]TidalAlbum, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/artists/%d/albums", artistID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get artist albums failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalAlbum `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode artist albums: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/album/%d", tidalWebBase, result.Items[i].ID)
	}

	return result.Items, nil
}

// GetArtistTopTracks retrieves top tracks from an artist
func (c *TidalClient) GetArtistTopTracks(ctx context.Context, artistID int, limit int) ([]TidalTrack, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/artists/%d/toptracks", artistID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get artist top tracks failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalTrack `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode artist top tracks: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/track/%d", tidalWebBase, result.Items[i].ID)
	}

	return result.Items, nil
}

// GetPlaylist retrieves a specific playlist
func (c *TidalClient) GetPlaylist(ctx context.Context, playlistUUID string) (*TidalPlaylist, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/playlists/%s", playlistUUID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get playlist failed with status: %d", resp.StatusCode)
	}

	var playlist TidalPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, fmt.Errorf("failed to decode playlist: %w", err)
	}

	playlist.URL = fmt.Sprintf("%s/playlist/%s", tidalWebBase, playlist.UUID)
	return &playlist, nil
}

// GetPlaylistTracks retrieves tracks from a playlist
func (c *TidalClient) GetPlaylistTracks(ctx context.Context, playlistUUID string, limit int) ([]TidalTrack, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/playlists/%s/tracks", playlistUUID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get playlist tracks failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []struct {
			Item TidalTrack `json:"item"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode playlist tracks: %w", err)
	}

	tracks := make([]TidalTrack, len(result.Items))
	for i, item := range result.Items {
		tracks[i] = item.Item
		tracks[i].URL = fmt.Sprintf("%s/track/%d", tidalWebBase, tracks[i].ID)
	}

	return tracks, nil
}

// GetGenres retrieves available genres
func (c *TidalClient) GetGenres(ctx context.Context) ([]TidalGenre, error) {
	resp, err := c.doRequest(ctx, "GET", "/genres", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get genres failed with status: %d", resp.StatusCode)
	}

	var genres []TidalGenre
	if err := json.NewDecoder(resp.Body).Decode(&genres); err != nil {
		return nil, fmt.Errorf("failed to decode genres: %w", err)
	}

	return genres, nil
}

// GetNewReleases retrieves new releases
func (c *TidalClient) GetNewReleases(ctx context.Context, genre string, limit int) ([]TidalNewRelease, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	endpoint := "/featured/new"
	if genre != "" {
		endpoint = fmt.Sprintf("/genres/%s/albums", genre)
	}

	resp, err := c.doRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get new releases failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalAlbum `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode new releases: %w", err)
	}

	releases := make([]TidalNewRelease, len(result.Items))
	for i, album := range result.Items {
		releases[i] = TidalNewRelease{
			ID:           album.ID,
			Title:        album.Title,
			Artists:      album.Artists,
			ReleaseDate:  album.ReleaseDate,
			Cover:        album.Cover,
			AudioQuality: album.AudioQuality,
			Type:         "ALBUM",
			URL:          fmt.Sprintf("%s/album/%d", tidalWebBase, album.ID),
		}
	}

	return releases, nil
}

// GetBestsellers retrieves bestselling albums
func (c *TidalClient) GetBestsellers(ctx context.Context, genre string, limit int) ([]TidalAlbum, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	endpoint := "/featured/top"
	if genre != "" {
		endpoint = fmt.Sprintf("/genres/%s/top", genre)
	}

	resp, err := c.doRequest(ctx, "GET", endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get bestsellers failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalAlbum `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode bestsellers: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/album/%d", tidalWebBase, result.Items[i].ID)
	}

	return result.Items, nil
}

// GetQualityInfo retrieves quality information for a track
func (c *TidalClient) GetQualityInfo(ctx context.Context, trackID int) (*TidalQualityInfo, error) {
	track, err := c.GetTrack(ctx, trackID)
	if err != nil {
		return nil, err
	}

	info := &TidalQualityInfo{
		AudioQuality: track.AudioQuality,
		AudioModes:   track.AudioModes,
	}

	// Determine quality features from audio modes and quality
	for _, mode := range track.AudioModes {
		switch mode {
		case "DOLBY_ATMOS":
			info.HasDolby = true
		case "SONY_360RA":
			info.Has360 = true
		}
	}

	// MQA is typically indicated in HI_RES quality
	if track.AudioQuality == "HI_RES" || track.AudioQuality == "HI_RES_LOSSLESS" {
		info.HasMQA = true
		info.BitDepth = 24
		info.SampleRate = 96000
	} else if track.AudioQuality == "LOSSLESS" {
		info.BitDepth = 16
		info.SampleRate = 44100
	}

	return info, nil
}

// GetSimilarArtists retrieves similar artists
func (c *TidalClient) GetSimilarArtists(ctx context.Context, artistID int, limit int) ([]TidalArtist, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/artists/%d/similar", artistID), params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get similar artists failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalArtist `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode similar artists: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/artist/%d", tidalWebBase, result.Items[i].ID)
	}

	return result.Items, nil
}

// GetMixes retrieves curated mixes/playlists
func (c *TidalClient) GetMixes(ctx context.Context, limit int) ([]TidalPlaylist, error) {
	if limit <= 0 || limit > 50 {
		limit = 25
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))

	resp, err := c.doRequest(ctx, "GET", "/featured/mixes", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get mixes failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Items []TidalPlaylist `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mixes: %w", err)
	}

	for i := range result.Items {
		result.Items[i].URL = fmt.Sprintf("%s/playlist/%s", tidalWebBase, result.Items[i].UUID)
	}

	return result.Items, nil
}
