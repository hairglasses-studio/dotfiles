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
	"strings"
	"sync"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	mixcloudAPIBase = "https://api.mixcloud.com"
)

// MixcloudClient provides access to Mixcloud API
type MixcloudClient struct {
	httpClient *http.Client
	mu         sync.RWMutex
}

// MixcloudUser represents a Mixcloud user
type MixcloudUser struct {
	Key            string   `json:"key"`
	Username       string   `json:"username"`
	Name           string   `json:"name"`
	URL            string   `json:"url"`
	Pictures       Pictures `json:"pictures,omitempty"`
	Bio            string   `json:"biog,omitempty"`
	City           string   `json:"city,omitempty"`
	Country        string   `json:"country,omitempty"`
	FollowerCount  int      `json:"follower_count"`
	FollowingCount int      `json:"following_count"`
	CloudcastCount int      `json:"cloudcast_count"`
	FavoriteCount  int      `json:"favorite_count"`
	ListenCount    int      `json:"listen_count"`
	IsFollowing    bool     `json:"is_following"`
	IsPro          bool     `json:"is_pro"`
}

// Pictures represents Mixcloud image URLs
type Pictures struct {
	Small        string `json:"small,omitempty"`
	Medium       string `json:"medium,omitempty"`
	MediumMobile string `json:"medium_mobile,omitempty"`
	Large        string `json:"large,omitempty"`
	ExtraLarge   string `json:"extra_large,omitempty"`
	Thumbnail    string `json:"thumbnail,omitempty"`
}

// MixcloudCloudcast represents a mix/show on Mixcloud
type MixcloudCloudcast struct {
	Key           string            `json:"key"`
	Name          string            `json:"name"`
	URL           string            `json:"url"`
	User          MixcloudUser      `json:"user"`
	Pictures      Pictures          `json:"pictures,omitempty"`
	Tags          []MixcloudTag     `json:"tags,omitempty"`
	CreatedTime   string            `json:"created_time"`
	UpdatedTime   string            `json:"updated_time,omitempty"`
	PlayCount     int               `json:"play_count"`
	FavoriteCount int               `json:"favorite_count"`
	ListenerCount int               `json:"listener_count"`
	CommentCount  int               `json:"comment_count"`
	RepostCount   int               `json:"repost_count"`
	AudioLength   int               `json:"audio_length"` // seconds
	Description   string            `json:"description,omitempty"`
	Slug          string            `json:"slug"`
	Sections      []MixcloudSection `json:"sections,omitempty"`
	IsExclusive   bool              `json:"is_exclusive"`
}

// MixcloudTag represents a genre/tag
type MixcloudTag struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// MixcloudSection represents a tracklist section
type MixcloudSection struct {
	TrackNumber int    `json:"track_number"`
	StartTime   int    `json:"start_time"` // seconds
	SectionType string `json:"section_type"`
	Track       struct {
		Key    string `json:"key"`
		Name   string `json:"name"`
		Artist struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"artist"`
	} `json:"track,omitempty"`
}

// MixcloudSearchResult represents search results
type MixcloudSearchResult struct {
	Query      string              `json:"query"`
	Type       string              `json:"type"`
	Cloudcasts []MixcloudCloudcast `json:"cloudcasts,omitempty"`
	Users      []MixcloudUser      `json:"users,omitempty"`
	Tags       []MixcloudTag       `json:"tags,omitempty"`
	Total      int                 `json:"total"`
}

// MixcloudStatus represents connection status
type MixcloudStatus struct {
	Connected  bool   `json:"connected"`
	APIVersion string `json:"api_version"`
	HasYtDlp   bool   `json:"has_yt_dlp"`
	YtDlpPath  string `json:"yt_dlp_path,omitempty"`
}

// MixcloudHealth represents health status
type MixcloudHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasDownloader   bool     `json:"has_downloader"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// MixcloudDownloadResult represents download result
type MixcloudDownloadResult struct {
	Success    bool   `json:"success"`
	URL        string `json:"url"`
	OutputPath string `json:"output_path,omitempty"`
	Title      string `json:"title,omitempty"`
	Error      string `json:"error,omitempty"`
}

// MixcloudPagination represents API pagination
type MixcloudPagination struct {
	Data   json.RawMessage `json:"data"`
	Paging struct {
		Next     string `json:"next,omitempty"`
		Previous string `json:"previous,omitempty"`
	} `json:"paging"`
}

// NewMixcloudClient creates a new Mixcloud client
func NewMixcloudClient() (*MixcloudClient, error) {
	return &MixcloudClient{
		httpClient: httpclient.Standard(),
	}, nil
}

// GetStatus returns the client status
func (c *MixcloudClient) GetStatus(ctx context.Context) (*MixcloudStatus, error) {
	status := &MixcloudStatus{
		APIVersion: "1.0",
	}

	// Check for yt-dlp
	if path, err := exec.LookPath("yt-dlp"); err == nil {
		status.HasYtDlp = true
		status.YtDlpPath = path
	}

	// Test API connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", mixcloudAPIBase+"/discover/", nil)
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
func (c *MixcloudClient) GetHealth(ctx context.Context) (*MixcloudHealth, error) {
	status, _ := c.GetStatus(ctx)

	health := &MixcloudHealth{
		Score:     100,
		Status:    "healthy",
		Connected: status.Connected,
	}

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Cannot connect to Mixcloud API")
		health.Recommendations = append(health.Recommendations, "Check internet connection")
	}

	if !status.HasYtDlp {
		health.Score -= 30
		health.HasDownloader = false
		health.Issues = append(health.Issues, "yt-dlp not available for downloads")
		health.Recommendations = append(health.Recommendations, "Install yt-dlp: pip install yt-dlp")
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

// Search searches Mixcloud for cloudcasts, users, or tags
func (c *MixcloudClient) Search(ctx context.Context, query string, searchType string) (*MixcloudSearchResult, error) {
	result := &MixcloudSearchResult{
		Query: query,
		Type:  searchType,
	}

	endpoint := "/search/"
	switch searchType {
	case "user":
		endpoint = "/search/?type=user"
	case "tag":
		endpoint = "/search/?type=tag"
	case "cloudcast", "show", "mix":
		endpoint = "/search/?type=cloudcast"
	default:
		endpoint = "/search/?type=cloudcast" // Default to mixes
	}

	searchURL := fmt.Sprintf("%s%s&q=%s", mixcloudAPIBase, endpoint, url.QueryEscape(query))

	body, err := c.doRequest(ctx, searchURL)
	if err != nil {
		return nil, err
	}

	var pagination MixcloudPagination
	if err := json.Unmarshal(body, &pagination); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	switch searchType {
	case "user":
		var users []MixcloudUser
		if err := json.Unmarshal(pagination.Data, &users); err == nil {
			result.Users = users
			result.Total = len(users)
		}
	case "tag":
		var tags []MixcloudTag
		if err := json.Unmarshal(pagination.Data, &tags); err == nil {
			result.Tags = tags
			result.Total = len(tags)
		}
	default:
		var cloudcasts []MixcloudCloudcast
		if err := json.Unmarshal(pagination.Data, &cloudcasts); err == nil {
			result.Cloudcasts = cloudcasts
			result.Total = len(cloudcasts)
		}
	}

	return result, nil
}

// GetUser gets user profile details
func (c *MixcloudClient) GetUser(ctx context.Context, username string) (*MixcloudUser, error) {
	userURL := fmt.Sprintf("%s/%s/", mixcloudAPIBase, username)

	body, err := c.doRequest(ctx, userURL)
	if err != nil {
		return nil, err
	}

	var user MixcloudUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}

	return &user, nil
}

// GetUserCloudcasts gets a user's uploaded mixes
func (c *MixcloudClient) GetUserCloudcasts(ctx context.Context, username string, limit int) ([]MixcloudCloudcast, error) {
	if limit <= 0 {
		limit = 20
	}

	cloudcastsURL := fmt.Sprintf("%s/%s/cloudcasts/?limit=%d", mixcloudAPIBase, username, limit)

	body, err := c.doRequest(ctx, cloudcastsURL)
	if err != nil {
		return nil, err
	}

	var pagination MixcloudPagination
	if err := json.Unmarshal(body, &pagination); err != nil {
		return nil, fmt.Errorf("failed to parse cloudcasts: %w", err)
	}

	var cloudcasts []MixcloudCloudcast
	if err := json.Unmarshal(pagination.Data, &cloudcasts); err != nil {
		return nil, fmt.Errorf("failed to parse cloudcasts data: %w", err)
	}

	return cloudcasts, nil
}

// GetUserFavorites gets a user's favorite mixes
func (c *MixcloudClient) GetUserFavorites(ctx context.Context, username string, limit int) ([]MixcloudCloudcast, error) {
	if limit <= 0 {
		limit = 20
	}

	favoritesURL := fmt.Sprintf("%s/%s/favorites/?limit=%d", mixcloudAPIBase, username, limit)

	body, err := c.doRequest(ctx, favoritesURL)
	if err != nil {
		return nil, err
	}

	var pagination MixcloudPagination
	if err := json.Unmarshal(body, &pagination); err != nil {
		return nil, fmt.Errorf("failed to parse favorites: %w", err)
	}

	var cloudcasts []MixcloudCloudcast
	if err := json.Unmarshal(pagination.Data, &cloudcasts); err != nil {
		return nil, fmt.Errorf("failed to parse favorites data: %w", err)
	}

	return cloudcasts, nil
}

// GetCloudcast gets cloudcast/mix details including tracklist
func (c *MixcloudClient) GetCloudcast(ctx context.Context, cloudcastKey string) (*MixcloudCloudcast, error) {
	// cloudcastKey format: /username/mix-name/
	if !strings.HasPrefix(cloudcastKey, "/") {
		cloudcastKey = "/" + cloudcastKey
	}
	if !strings.HasSuffix(cloudcastKey, "/") {
		cloudcastKey = cloudcastKey + "/"
	}

	cloudcastURL := mixcloudAPIBase + cloudcastKey

	body, err := c.doRequest(ctx, cloudcastURL)
	if err != nil {
		return nil, err
	}

	var cloudcast MixcloudCloudcast
	if err := json.Unmarshal(body, &cloudcast); err != nil {
		return nil, fmt.Errorf("failed to parse cloudcast: %w", err)
	}

	return &cloudcast, nil
}

// GetPopularTags returns popular Mixcloud tags/genres
func (c *MixcloudClient) GetPopularTags(ctx context.Context) ([]MixcloudTag, error) {
	// Return common DJ/music tags
	return []MixcloudTag{
		{Key: "/discover/house/", Name: "House", URL: "https://www.mixcloud.com/discover/house/"},
		{Key: "/discover/techno/", Name: "Techno", URL: "https://www.mixcloud.com/discover/techno/"},
		{Key: "/discover/deep-house/", Name: "Deep House", URL: "https://www.mixcloud.com/discover/deep-house/"},
		{Key: "/discover/drum-and-bass/", Name: "Drum and Bass", URL: "https://www.mixcloud.com/discover/drum-and-bass/"},
		{Key: "/discover/hip-hop/", Name: "Hip Hop", URL: "https://www.mixcloud.com/discover/hip-hop/"},
		{Key: "/discover/disco/", Name: "Disco", URL: "https://www.mixcloud.com/discover/disco/"},
		{Key: "/discover/funk/", Name: "Funk", URL: "https://www.mixcloud.com/discover/funk/"},
		{Key: "/discover/soul/", Name: "Soul", URL: "https://www.mixcloud.com/discover/soul/"},
		{Key: "/discover/ambient/", Name: "Ambient", URL: "https://www.mixcloud.com/discover/ambient/"},
		{Key: "/discover/electronica/", Name: "Electronica", URL: "https://www.mixcloud.com/discover/electronica/"},
		{Key: "/discover/trance/", Name: "Trance", URL: "https://www.mixcloud.com/discover/trance/"},
		{Key: "/discover/dub/", Name: "Dub", URL: "https://www.mixcloud.com/discover/dub/"},
		{Key: "/discover/reggae/", Name: "Reggae", URL: "https://www.mixcloud.com/discover/reggae/"},
		{Key: "/discover/jazz/", Name: "Jazz", URL: "https://www.mixcloud.com/discover/jazz/"},
		{Key: "/discover/world/", Name: "World", URL: "https://www.mixcloud.com/discover/world/"},
	}, nil
}

// GetTagCloudcasts gets mixes for a specific tag/genre
func (c *MixcloudClient) GetTagCloudcasts(ctx context.Context, tag string, limit int) ([]MixcloudCloudcast, error) {
	if limit <= 0 {
		limit = 20
	}

	// Clean tag name
	tag = strings.ToLower(strings.ReplaceAll(tag, " ", "-"))
	tagURL := fmt.Sprintf("%s/discover/%s/?limit=%d", mixcloudAPIBase, tag, limit)

	body, err := c.doRequest(ctx, tagURL)
	if err != nil {
		return nil, err
	}

	var pagination MixcloudPagination
	if err := json.Unmarshal(body, &pagination); err != nil {
		return nil, fmt.Errorf("failed to parse tag cloudcasts: %w", err)
	}

	var cloudcasts []MixcloudCloudcast
	if err := json.Unmarshal(pagination.Data, &cloudcasts); err != nil {
		return nil, fmt.Errorf("failed to parse cloudcasts data: %w", err)
	}

	return cloudcasts, nil
}

// Download downloads a mix using yt-dlp
func (c *MixcloudClient) Download(ctx context.Context, mixURL string, outputDir string) (*MixcloudDownloadResult, error) {
	result := &MixcloudDownloadResult{
		URL: mixURL,
	}

	// Check for yt-dlp
	ytdlpPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		return nil, fmt.Errorf("yt-dlp not found - install with: pip install yt-dlp")
	}

	args := []string{
		"-x", // Extract audio
		"--audio-format", "mp3",
		"--audio-quality", "0",
		"--embed-thumbnail",
		"--add-metadata",
	}

	if outputDir != "" {
		args = append(args, "-o", outputDir+"/%(uploader)s - %(title)s.%(ext)s")
		result.OutputPath = outputDir
	}

	args = append(args, mixURL)

	cmd := exec.CommandContext(ctx, ytdlpPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Success = false
		result.Error = string(output)
		return result, fmt.Errorf("download failed: %s", output)
	}

	result.Success = true

	// Try to extract title from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Destination:") {
			parts := strings.SplitN(line, "Destination:", 2)
			if len(parts) == 2 {
				result.Title = strings.TrimSpace(parts[1])
			}
		}
	}

	return result, nil
}

// doRequest performs an HTTP request and returns the body
func (c *MixcloudClient) doRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}
