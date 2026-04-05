// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	soundCloudAPIv1 = "https://api.soundcloud.com"
	soundCloudAPIv2 = "https://api-v2.soundcloud.com"
)

// SoundCloudClient provides access to SoundCloud API
type SoundCloudClient struct {
	clientID     string
	clientSecret string
	accessToken  string
	tokenExpiry  time.Time
	httpClient   *http.Client
	mu           sync.RWMutex
}

// SoundCloudStatus represents API connection status
type SoundCloudStatus struct {
	Connected   bool   `json:"connected"`
	HasToken    bool   `json:"has_token"`
	TokenExpiry string `json:"token_expiry,omitempty"`
	APIVersion  string `json:"api_version"`
}

// SoundCloudUser represents a user
type SoundCloudUser struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Permalink     string `json:"permalink"`
	FullName      string `json:"full_name,omitempty"`
	Description   string `json:"description,omitempty"`
	City          string `json:"city,omitempty"`
	Country       string `json:"country,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	FollowerCount int    `json:"followers_count"`
	TrackCount    int    `json:"track_count"`
	PlaylistCount int    `json:"playlist_count"`
	URI           string `json:"uri"`
	PermalinkURL  string `json:"permalink_url"`
}

// SoundCloudTrack represents a track
type SoundCloudTrack struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description,omitempty"`
	Duration      int            `json:"duration"` // milliseconds
	Genre         string         `json:"genre,omitempty"`
	TagList       string         `json:"tag_list,omitempty"`
	BPM           float64        `json:"bpm,omitempty"`
	KeySignature  string         `json:"key_signature,omitempty"`
	Waveform      string         `json:"waveform_url,omitempty"`
	StreamURL     string         `json:"stream_url,omitempty"`
	DownloadURL   string         `json:"download_url,omitempty"`
	Downloadable  bool           `json:"downloadable"`
	ArtworkURL    string         `json:"artwork_url,omitempty"`
	User          SoundCloudUser `json:"user"`
	PlaybackCount int            `json:"playback_count"`
	LikesCount    int            `json:"likes_count"`
	RepostsCount  int            `json:"reposts_count"`
	CommentCount  int            `json:"comment_count"`
	CreatedAt     string         `json:"created_at"`
	PermalinkURL  string         `json:"permalink_url"`
	URI           string         `json:"uri"`
	LabelName     string         `json:"label_name,omitempty"`
	Purchase      string         `json:"purchase_url,omitempty"`
	ReleaseDate   string         `json:"release_date,omitempty"`
	License       string         `json:"license,omitempty"`
}

// SoundCloudPlaylist represents a playlist
type SoundCloudPlaylist struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description,omitempty"`
	Duration     int               `json:"duration"` // milliseconds
	TrackCount   int               `json:"track_count"`
	User         SoundCloudUser    `json:"user"`
	Tracks       []SoundCloudTrack `json:"tracks,omitempty"`
	ArtworkURL   string            `json:"artwork_url,omitempty"`
	LikesCount   int               `json:"likes_count"`
	CreatedAt    string            `json:"created_at"`
	PermalinkURL string            `json:"permalink_url"`
	URI          string            `json:"uri"`
	IsAlbum      bool              `json:"is_album"`
	Genre        string            `json:"genre,omitempty"`
}

// SoundCloudComment represents a comment on a track
type SoundCloudComment struct {
	ID        int64          `json:"id"`
	Body      string         `json:"body"`
	Timestamp int            `json:"timestamp"` // position in track (ms)
	User      SoundCloudUser `json:"user"`
	CreatedAt string         `json:"created_at"`
}

// SoundCloudSearchResult represents search results
type SoundCloudSearchResult struct {
	Tracks    []SoundCloudTrack    `json:"tracks,omitempty"`
	Users     []SoundCloudUser     `json:"users,omitempty"`
	Playlists []SoundCloudPlaylist `json:"playlists,omitempty"`
}

// SoundCloudHealth represents health status
type SoundCloudHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasCredentials  bool     `json:"has_credentials"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	soundcloudClientSingleton *SoundCloudClient
	soundcloudClientOnce      sync.Once
	soundcloudClientErr       error

	// TestOverrideSoundCloudClient, when non-nil, is returned by GetSoundCloudClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideSoundCloudClient *SoundCloudClient
)

// GetSoundCloudClient returns the singleton SoundCloud client.
func GetSoundCloudClient() (*SoundCloudClient, error) {
	if TestOverrideSoundCloudClient != nil {
		return TestOverrideSoundCloudClient, nil
	}
	soundcloudClientOnce.Do(func() {
		soundcloudClientSingleton, soundcloudClientErr = NewSoundCloudClient()
	})
	return soundcloudClientSingleton, soundcloudClientErr
}

// NewTestSoundCloudClient creates an in-memory test client.
func NewTestSoundCloudClient() *SoundCloudClient {
	return &SoundCloudClient{
		clientID:   "test-client-id",
		httpClient: httpclient.Fast(),
	}
}

// NewSoundCloudClient creates a new SoundCloud client
func NewSoundCloudClient() (*SoundCloudClient, error) {
	clientID := os.Getenv("SOUNDCLOUD_CLIENT_ID")
	clientSecret := os.Getenv("SOUNDCLOUD_CLIENT_SECRET")

	if clientID == "" {
		return nil, fmt.Errorf("SOUNDCLOUD_CLIENT_ID required")
	}

	return &SoundCloudClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   httpclient.Standard(),
	}, nil
}

// authenticate gets an access token using client credentials flow
func (c *SoundCloudClient) authenticate(ctx context.Context) error {
	if c.clientSecret == "" {
		return nil // Use client_id only for public endpoints
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", soundCloudAPIv1+"/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// doRequest makes an authenticated API request
func (c *SoundCloudClient) doRequest(ctx context.Context, method, endpoint string, useV2 bool) (*http.Response, error) {
	_ = c.authenticate(ctx) // Ignore error, fall back to client_id

	baseURL := soundCloudAPIv1
	if useV2 {
		baseURL = soundCloudAPIv2
	}

	fullURL := baseURL + endpoint
	if strings.Contains(endpoint, "?") {
		fullURL += "&client_id=" + c.clientID
	} else {
		fullURL += "?client_id=" + c.clientID
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	token := c.accessToken
	c.mu.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "OAuth "+token)
	}
	req.Header.Set("Accept", "application/json")

	return c.httpClient.Do(req)
}

// GetStatus returns API connection status
func (c *SoundCloudClient) GetStatus(ctx context.Context) (*SoundCloudStatus, error) {
	status := &SoundCloudStatus{
		Connected:  false,
		HasToken:   false,
		APIVersion: "v1",
	}

	if err := c.authenticate(ctx); err == nil && c.accessToken != "" {
		c.mu.RLock()
		status.HasToken = true
		if !c.tokenExpiry.IsZero() {
			status.TokenExpiry = c.tokenExpiry.Format(time.RFC3339)
		}
		c.mu.RUnlock()
	}

	// Test API connection with a simple resolve call
	resp, err := c.doRequest(ctx, "GET", "/resolve?url=https://soundcloud.com/soundcloud", false)
	if err != nil {
		return status, nil
	}
	defer resp.Body.Close()

	status.Connected = resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound

	return status, nil
}

// Resolve resolves a SoundCloud URL to a resource
func (c *SoundCloudClient) Resolve(ctx context.Context, scURL string) (interface{}, string, error) {
	resp, err := c.doRequest(ctx, "GET", "/resolve?url="+url.QueryEscape(scURL), false)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("resolve failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Detect resource type from response
	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return nil, "", err
	}

	switch probe.Kind {
	case "track":
		var track SoundCloudTrack
		if err := json.Unmarshal(body, &track); err != nil {
			return nil, "", err
		}
		return &track, "track", nil
	case "user":
		var user SoundCloudUser
		if err := json.Unmarshal(body, &user); err != nil {
			return nil, "", err
		}
		return &user, "user", nil
	case "playlist":
		var playlist SoundCloudPlaylist
		if err := json.Unmarshal(body, &playlist); err != nil {
			return nil, "", err
		}
		return &playlist, "playlist", nil
	default:
		return nil, probe.Kind, fmt.Errorf("unknown resource type: %s", probe.Kind)
	}
}

// GetUser gets user by ID
func (c *SoundCloudClient) GetUser(ctx context.Context, userID int64) (*SoundCloudUser, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d", userID), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user failed: %s", resp.Status)
	}

	var user SoundCloudUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// GetTrack gets track by ID
func (c *SoundCloudClient) GetTrack(ctx context.Context, trackID int64) (*SoundCloudTrack, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tracks/%d", trackID), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get track failed: %s", resp.Status)
	}

	var track SoundCloudTrack
	if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
		return nil, err
	}

	return &track, nil
}

// GetUserTracks gets a user's tracks
func (c *SoundCloudClient) GetUserTracks(ctx context.Context, userID int64, limit int) ([]SoundCloudTrack, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/tracks?limit=%d", userID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user tracks failed: %s", resp.Status)
	}

	var tracks []SoundCloudTrack
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, err
	}

	return tracks, nil
}

// GetUserLikes gets a user's liked tracks (v2 API)
func (c *SoundCloudClient) GetUserLikes(ctx context.Context, userID int64, limit int) ([]SoundCloudTrack, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/likes?limit=%d", userID, limit), true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Fall back to v1
		return c.getUserLikesV1(ctx, userID, limit)
	}

	var result struct {
		Collection []struct {
			Track SoundCloudTrack `json:"track"`
		} `json:"collection"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var tracks []SoundCloudTrack
	for _, item := range result.Collection {
		if item.Track.ID != 0 {
			tracks = append(tracks, item.Track)
		}
	}

	return tracks, nil
}

func (c *SoundCloudClient) getUserLikesV1(ctx context.Context, userID int64, limit int) ([]SoundCloudTrack, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/favorites?limit=%d", userID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user likes failed: %s", resp.Status)
	}

	var tracks []SoundCloudTrack
	if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
		return nil, err
	}

	return tracks, nil
}

// GetPlaylist gets playlist by ID
func (c *SoundCloudClient) GetPlaylist(ctx context.Context, playlistID int64) (*SoundCloudPlaylist, error) {
	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/playlists/%d", playlistID), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get playlist failed: %s", resp.Status)
	}

	var playlist SoundCloudPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	return &playlist, nil
}

// GetUserPlaylists gets a user's playlists
func (c *SoundCloudClient) GetUserPlaylists(ctx context.Context, userID int64, limit int) ([]SoundCloudPlaylist, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/playlists?limit=%d", userID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get user playlists failed: %s", resp.Status)
	}

	var playlists []SoundCloudPlaylist
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return nil, err
	}

	return playlists, nil
}

// GetTrackComments gets comments on a track
func (c *SoundCloudClient) GetTrackComments(ctx context.Context, trackID int64, limit int) ([]SoundCloudComment, error) {
	if limit <= 0 {
		limit = 50
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tracks/%d/comments?limit=%d", trackID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get track comments failed: %s", resp.Status)
	}

	var comments []SoundCloudComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// Search searches for tracks, users, or playlists
func (c *SoundCloudClient) Search(ctx context.Context, query string, searchType string, limit int) (*SoundCloudSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	result := &SoundCloudSearchResult{}

	switch searchType {
	case "track", "tracks", "":
		resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tracks?q=%s&limit=%d", url.QueryEscape(query), limit), false)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("search tracks failed: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result.Tracks); err != nil {
			return nil, err
		}

	case "user", "users":
		resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users?q=%s&limit=%d", url.QueryEscape(query), limit), false)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("search users failed: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result.Users); err != nil {
			return nil, err
		}

	case "playlist", "playlists":
		resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/playlists?q=%s&limit=%d", url.QueryEscape(query), limit), false)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("search playlists failed: %s", resp.Status)
		}

		if err := json.NewDecoder(resp.Body).Decode(&result.Playlists); err != nil {
			return nil, err
		}

	case "all":
		// Search all types in parallel
		var wg sync.WaitGroup
		var mu sync.Mutex
		var searchErr error

		wg.Add(3)

		go func() {
			defer wg.Done()
			resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/tracks?q=%s&limit=%d", url.QueryEscape(query), limit), false)
			if err != nil {
				mu.Lock()
				searchErr = err
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var tracks []SoundCloudTrack
				if err := json.NewDecoder(resp.Body).Decode(&tracks); err == nil {
					mu.Lock()
					result.Tracks = tracks
					mu.Unlock()
				}
			}
		}()

		go func() {
			defer wg.Done()
			resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users?q=%s&limit=%d", url.QueryEscape(query), limit), false)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var users []SoundCloudUser
				if err := json.NewDecoder(resp.Body).Decode(&users); err == nil {
					mu.Lock()
					result.Users = users
					mu.Unlock()
				}
			}
		}()

		go func() {
			defer wg.Done()
			resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/playlists?q=%s&limit=%d", url.QueryEscape(query), limit), false)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var playlists []SoundCloudPlaylist
				if err := json.NewDecoder(resp.Body).Decode(&playlists); err == nil {
					mu.Lock()
					result.Playlists = playlists
					mu.Unlock()
				}
			}
		}()

		wg.Wait()

		if searchErr != nil && len(result.Tracks) == 0 && len(result.Users) == 0 && len(result.Playlists) == 0 {
			return nil, searchErr
		}

	default:
		return nil, fmt.Errorf("invalid search type: %s (use track, user, playlist, or all)", searchType)
	}

	return result, nil
}

// GetUserFollowers gets a user's followers
func (c *SoundCloudClient) GetUserFollowers(ctx context.Context, userID int64, limit int) ([]SoundCloudUser, error) {
	if limit <= 0 {
		limit = 50
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/followers?limit=%d", userID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get followers failed: %s", resp.Status)
	}

	var users []SoundCloudUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserFollowings gets users that a user follows
func (c *SoundCloudClient) GetUserFollowings(ctx context.Context, userID int64, limit int) ([]SoundCloudUser, error) {
	if limit <= 0 {
		limit = 50
	}

	resp, err := c.doRequest(ctx, "GET", fmt.Sprintf("/users/%d/followings?limit=%d", userID, limit), false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get followings failed: %s", resp.Status)
	}

	var users []SoundCloudUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

// GetHealth returns health status
func (c *SoundCloudClient) GetHealth(ctx context.Context) (*SoundCloudHealth, error) {
	health := &SoundCloudHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check credentials
	if c.clientID == "" {
		health.Score -= 50
		health.HasCredentials = false
		health.Issues = append(health.Issues, "Missing SoundCloud client ID")
		health.Recommendations = append(health.Recommendations, "Set SOUNDCLOUD_CLIENT_ID environment variable")
	} else {
		health.HasCredentials = true
	}

	// Test connection
	status, _ := c.GetStatus(ctx)
	health.Connected = status != nil && status.Connected

	if !health.Connected && health.HasCredentials {
		health.Score -= 30
		health.Issues = append(health.Issues, "Cannot connect to SoundCloud API")
		health.Recommendations = append(health.Recommendations, "Check client ID is valid and not rate-limited")
	}

	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// FormatDuration formats duration in ms to human-readable string
func FormatDuration(ms int) string {
	s := ms / 1000
	m := s / 60
	s = s % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
