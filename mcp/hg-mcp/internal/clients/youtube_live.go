// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// YouTubeLiveClient provides access to YouTube Live Streaming API
type YouTubeLiveClient struct {
	apiKey     string
	channelID  string
	httpClient *http.Client
}

// YouTubeLiveStatus represents broadcast and connection status
type YouTubeLiveStatus struct {
	Connected       bool   `json:"connected"`
	IsLive          bool   `json:"is_live"`
	BroadcastID     string `json:"broadcast_id,omitempty"`
	BroadcastTitle  string `json:"broadcast_title,omitempty"`
	BroadcastStatus string `json:"broadcast_status,omitempty"`
	ViewerCount     int64  `json:"viewer_count,omitempty"`
	LikeCount       int64  `json:"like_count,omitempty"`
	StartTime       string `json:"start_time,omitempty"`
	ChannelID       string `json:"channel_id"`
}

// YouTubeBroadcast represents a live broadcast
type YouTubeBroadcast struct {
	ID                 string                  `json:"id"`
	Title              string                  `json:"title"`
	Description        string                  `json:"description"`
	ScheduledStartTime string                  `json:"scheduled_start_time"`
	ActualStartTime    string                  `json:"actual_start_time,omitempty"`
	ActualEndTime      string                  `json:"actual_end_time,omitempty"`
	Status             YouTubeBroadcastStatus  `json:"status"`
	Statistics         YouTubeBroadcastStats   `json:"statistics,omitempty"`
	ContentDetails     YouTubeBroadcastContent `json:"content_details,omitempty"`
}

// YouTubeBroadcastStatus represents broadcast status
type YouTubeBroadcastStatus struct {
	LifeCycleStatus         string `json:"life_cycle_status"`
	PrivacyStatus           string `json:"privacy_status"`
	RecordingStatus         string `json:"recording_status"`
	MadeForKids             bool   `json:"made_for_kids"`
	SelfDeclaredMadeForKids bool   `json:"self_declared_made_for_kids"`
}

// YouTubeBroadcastStats represents broadcast statistics
type YouTubeBroadcastStats struct {
	TotalChatCount int64 `json:"total_chat_count"`
}

// YouTubeBroadcastContent represents broadcast content details
type YouTubeBroadcastContent struct {
	BoundStreamID           string `json:"bound_stream_id"`
	EnableDvr               bool   `json:"enable_dvr"`
	EnableContentEncryption bool   `json:"enable_content_encryption"`
	EnableEmbed             bool   `json:"enable_embed"`
	RecordFromStart         bool   `json:"record_from_start"`
	StartWithSlate          bool   `json:"start_with_slate"`
}

// YouTubeLiveChatMessage represents a live chat message
type YouTubeLiveChatMessage struct {
	ID              string `json:"id"`
	AuthorChannelID string `json:"author_channel_id"`
	AuthorName      string `json:"author_name"`
	AuthorImageURL  string `json:"author_image_url"`
	Message         string `json:"message"`
	PublishedAt     string `json:"published_at"`
	IsChatOwner     bool   `json:"is_chat_owner"`
	IsChatModerator bool   `json:"is_chat_moderator"`
	IsChatSponsor   bool   `json:"is_chat_sponsor"`
}

// YouTubeVideo represents video/stream details
type YouTubeVideo struct {
	ID           string              `json:"id"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	ChannelID    string              `json:"channel_id"`
	ChannelTitle string              `json:"channel_title"`
	PublishedAt  string              `json:"published_at"`
	ThumbnailURL string              `json:"thumbnail_url"`
	LiveDetails  *YouTubeLiveDetails `json:"live_details,omitempty"`
	Statistics   *YouTubeVideoStats  `json:"statistics,omitempty"`
}

// YouTubeLiveDetails represents live streaming details
type YouTubeLiveDetails struct {
	ActualStartTime    string `json:"actual_start_time"`
	ActualEndTime      string `json:"actual_end_time,omitempty"`
	ScheduledStartTime string `json:"scheduled_start_time"`
	ConcurrentViewers  int64  `json:"concurrent_viewers"`
	ActiveLiveChatID   string `json:"active_live_chat_id"`
}

// YouTubeVideoStats represents video statistics
type YouTubeVideoStats struct {
	ViewCount    int64 `json:"view_count"`
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
}

// YouTubeLiveHealth represents API health status
type YouTubeLiveHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasAPIKey       bool     `json:"has_api_key"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewYouTubeLiveClient creates a new YouTube Live client
func NewYouTubeLiveClient() (*YouTubeLiveClient, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("YOUTUBE_API_KEY environment variable required")
	}

	channelID := os.Getenv("YOUTUBE_CHANNEL_ID")

	return &YouTubeLiveClient{
		apiKey:    apiKey,
		channelID: channelID,
		httpClient: httpclient.Standard(),
	}, nil
}

// ChannelID returns the configured channel ID
func (c *YouTubeLiveClient) ChannelID() string {
	return c.channelID
}

// doRequest performs an API request
func (c *YouTubeLiveClient) doRequest(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	baseURL := "https://www.googleapis.com/youtube/v3"

	// Add API key to URL
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	fullURL := baseURL + endpoint + sep + "key=" + c.apiKey

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus returns broadcast and connection status
func (c *YouTubeLiveClient) GetStatus(ctx context.Context) (*YouTubeLiveStatus, error) {
	status := &YouTubeLiveStatus{
		Connected: false,
		ChannelID: c.channelID,
	}

	// Check API connection with a simple request
	broadcasts, err := c.GetBroadcasts(ctx, "active")
	if err != nil {
		return status, nil
	}
	status.Connected = true

	// Check for active broadcast
	for _, broadcast := range broadcasts {
		if broadcast.Status.LifeCycleStatus == "live" {
			status.IsLive = true
			status.BroadcastID = broadcast.ID
			status.BroadcastTitle = broadcast.Title
			status.BroadcastStatus = broadcast.Status.LifeCycleStatus
			break
		}
	}

	return status, nil
}

// GetBroadcasts gets broadcasts by status (active, all, completed, upcoming)
func (c *YouTubeLiveClient) GetBroadcasts(ctx context.Context, broadcastStatus string) ([]YouTubeBroadcast, error) {
	endpoint := fmt.Sprintf("/liveBroadcasts?part=id,snippet,status,contentDetails&broadcastStatus=%s", broadcastStatus)
	if c.channelID != "" {
		endpoint += "&mine=true"
	}

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title              string `json:"title"`
				Description        string `json:"description"`
				ScheduledStartTime string `json:"scheduledStartTime"`
				ActualStartTime    string `json:"actualStartTime"`
				ActualEndTime      string `json:"actualEndTime"`
			} `json:"snippet"`
			Status struct {
				LifeCycleStatus         string `json:"lifeCycleStatus"`
				PrivacyStatus           string `json:"privacyStatus"`
				RecordingStatus         string `json:"recordingStatus"`
				MadeForKids             bool   `json:"madeForKids"`
				SelfDeclaredMadeForKids bool   `json:"selfDeclaredMadeForKids"`
			} `json:"status"`
			ContentDetails struct {
				BoundStreamID           string `json:"boundStreamId"`
				EnableDvr               bool   `json:"enableDvr"`
				EnableContentEncryption bool   `json:"enableContentEncryption"`
				EnableEmbed             bool   `json:"enableEmbed"`
				RecordFromStart         bool   `json:"recordFromStart"`
				StartWithSlate          bool   `json:"startWithSlate"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding broadcasts: %w", err)
	}

	broadcasts := make([]YouTubeBroadcast, len(resp.Items))
	for i, item := range resp.Items {
		broadcasts[i] = YouTubeBroadcast{
			ID:                 item.ID,
			Title:              item.Snippet.Title,
			Description:        item.Snippet.Description,
			ScheduledStartTime: item.Snippet.ScheduledStartTime,
			ActualStartTime:    item.Snippet.ActualStartTime,
			ActualEndTime:      item.Snippet.ActualEndTime,
			Status: YouTubeBroadcastStatus{
				LifeCycleStatus:         item.Status.LifeCycleStatus,
				PrivacyStatus:           item.Status.PrivacyStatus,
				RecordingStatus:         item.Status.RecordingStatus,
				MadeForKids:             item.Status.MadeForKids,
				SelfDeclaredMadeForKids: item.Status.SelfDeclaredMadeForKids,
			},
			ContentDetails: YouTubeBroadcastContent{
				BoundStreamID:           item.ContentDetails.BoundStreamID,
				EnableDvr:               item.ContentDetails.EnableDvr,
				EnableContentEncryption: item.ContentDetails.EnableContentEncryption,
				EnableEmbed:             item.ContentDetails.EnableEmbed,
				RecordFromStart:         item.ContentDetails.RecordFromStart,
				StartWithSlate:          item.ContentDetails.StartWithSlate,
			},
		}
	}

	return broadcasts, nil
}

// GetBroadcast gets a specific broadcast by ID
func (c *YouTubeLiveClient) GetBroadcast(ctx context.Context, broadcastID string) (*YouTubeBroadcast, error) {
	endpoint := fmt.Sprintf("/liveBroadcasts?part=id,snippet,status,contentDetails,statistics&id=%s", broadcastID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title              string `json:"title"`
				Description        string `json:"description"`
				ScheduledStartTime string `json:"scheduledStartTime"`
				ActualStartTime    string `json:"actualStartTime"`
				ActualEndTime      string `json:"actualEndTime"`
			} `json:"snippet"`
			Status struct {
				LifeCycleStatus string `json:"lifeCycleStatus"`
				PrivacyStatus   string `json:"privacyStatus"`
				RecordingStatus string `json:"recordingStatus"`
			} `json:"status"`
			Statistics struct {
				TotalChatCount int64 `json:"totalChatCount,string"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding broadcast: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("broadcast not found: %s", broadcastID)
	}

	item := resp.Items[0]
	return &YouTubeBroadcast{
		ID:                 item.ID,
		Title:              item.Snippet.Title,
		Description:        item.Snippet.Description,
		ScheduledStartTime: item.Snippet.ScheduledStartTime,
		ActualStartTime:    item.Snippet.ActualStartTime,
		ActualEndTime:      item.Snippet.ActualEndTime,
		Status: YouTubeBroadcastStatus{
			LifeCycleStatus: item.Status.LifeCycleStatus,
			PrivacyStatus:   item.Status.PrivacyStatus,
			RecordingStatus: item.Status.RecordingStatus,
		},
		Statistics: YouTubeBroadcastStats{
			TotalChatCount: item.Statistics.TotalChatCount,
		},
	}, nil
}

// GetVideo gets video/stream details including live stats
func (c *YouTubeLiveClient) GetVideo(ctx context.Context, videoID string) (*YouTubeVideo, error) {
	endpoint := fmt.Sprintf("/videos?part=id,snippet,liveStreamingDetails,statistics&id=%s", videoID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				ChannelID    string `json:"channelId"`
				ChannelTitle string `json:"channelTitle"`
				PublishedAt  string `json:"publishedAt"`
				Thumbnails   struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			LiveStreamingDetails struct {
				ActualStartTime    string `json:"actualStartTime"`
				ActualEndTime      string `json:"actualEndTime"`
				ScheduledStartTime string `json:"scheduledStartTime"`
				ConcurrentViewers  string `json:"concurrentViewers"`
				ActiveLiveChatID   string `json:"activeLiveChatId"`
			} `json:"liveStreamingDetails"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding video: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	item := resp.Items[0]
	video := &YouTubeVideo{
		ID:           item.ID,
		Title:        item.Snippet.Title,
		Description:  item.Snippet.Description,
		ChannelID:    item.Snippet.ChannelID,
		ChannelTitle: item.Snippet.ChannelTitle,
		PublishedAt:  item.Snippet.PublishedAt,
		ThumbnailURL: item.Snippet.Thumbnails.High.URL,
	}

	// Parse live streaming details
	if item.LiveStreamingDetails.ActiveLiveChatID != "" {
		var concurrentViewers int64
		fmt.Sscanf(item.LiveStreamingDetails.ConcurrentViewers, "%d", &concurrentViewers)

		video.LiveDetails = &YouTubeLiveDetails{
			ActualStartTime:    item.LiveStreamingDetails.ActualStartTime,
			ActualEndTime:      item.LiveStreamingDetails.ActualEndTime,
			ScheduledStartTime: item.LiveStreamingDetails.ScheduledStartTime,
			ConcurrentViewers:  concurrentViewers,
			ActiveLiveChatID:   item.LiveStreamingDetails.ActiveLiveChatID,
		}
	}

	// Parse statistics
	var viewCount, likeCount, commentCount int64
	fmt.Sscanf(item.Statistics.ViewCount, "%d", &viewCount)
	fmt.Sscanf(item.Statistics.LikeCount, "%d", &likeCount)
	fmt.Sscanf(item.Statistics.CommentCount, "%d", &commentCount)

	video.Statistics = &YouTubeVideoStats{
		ViewCount:    viewCount,
		LikeCount:    likeCount,
		CommentCount: commentCount,
	}

	return video, nil
}

// GetLiveChatMessages gets recent live chat messages
func (c *YouTubeLiveClient) GetLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]YouTubeLiveChatMessage, string, error) {
	endpoint := fmt.Sprintf("/liveChat/messages?part=id,snippet,authorDetails&liveChatId=%s", url.QueryEscape(liveChatID))
	if pageToken != "" {
		endpoint += "&pageToken=" + pageToken
	}

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	var resp struct {
		NextPageToken string `json:"nextPageToken"`
		Items         []struct {
			ID      string `json:"id"`
			Snippet struct {
				Type               string `json:"type"`
				LiveChatID         string `json:"liveChatId"`
				AuthorChannelID    string `json:"authorChannelId"`
				PublishedAt        string `json:"publishedAt"`
				HasDisplayContent  bool   `json:"hasDisplayContent"`
				DisplayMessage     string `json:"displayMessage"`
				TextMessageDetails struct {
					MessageText string `json:"messageText"`
				} `json:"textMessageDetails"`
			} `json:"snippet"`
			AuthorDetails struct {
				ChannelID       string `json:"channelId"`
				DisplayName     string `json:"displayName"`
				ProfileImageURL string `json:"profileImageUrl"`
				IsChatOwner     bool   `json:"isChatOwner"`
				IsChatModerator bool   `json:"isChatModerator"`
				IsChatSponsor   bool   `json:"isChatSponsor"`
			} `json:"authorDetails"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("decoding chat messages: %w", err)
	}

	messages := make([]YouTubeLiveChatMessage, len(resp.Items))
	for i, item := range resp.Items {
		message := item.Snippet.DisplayMessage
		if message == "" {
			message = item.Snippet.TextMessageDetails.MessageText
		}
		messages[i] = YouTubeLiveChatMessage{
			ID:              item.ID,
			AuthorChannelID: item.AuthorDetails.ChannelID,
			AuthorName:      item.AuthorDetails.DisplayName,
			AuthorImageURL:  item.AuthorDetails.ProfileImageURL,
			Message:         message,
			PublishedAt:     item.Snippet.PublishedAt,
			IsChatOwner:     item.AuthorDetails.IsChatOwner,
			IsChatModerator: item.AuthorDetails.IsChatModerator,
			IsChatSponsor:   item.AuthorDetails.IsChatSponsor,
		}
	}

	return messages, resp.NextPageToken, nil
}

// SearchLiveStreams searches for live streams
func (c *YouTubeLiveClient) SearchLiveStreams(ctx context.Context, query string, maxResults int) ([]YouTubeVideo, error) {
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 25
	}

	endpoint := fmt.Sprintf("/search?part=id,snippet&type=video&eventType=live&q=%s&maxResults=%d",
		url.QueryEscape(query), maxResults)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title        string `json:"title"`
				Description  string `json:"description"`
				ChannelID    string `json:"channelId"`
				ChannelTitle string `json:"channelTitle"`
				PublishedAt  string `json:"publishedAt"`
				Thumbnails   struct {
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding search results: %w", err)
	}

	videos := make([]YouTubeVideo, len(resp.Items))
	for i, item := range resp.Items {
		videos[i] = YouTubeVideo{
			ID:           item.ID.VideoID,
			Title:        item.Snippet.Title,
			Description:  item.Snippet.Description,
			ChannelID:    item.Snippet.ChannelID,
			ChannelTitle: item.Snippet.ChannelTitle,
			PublishedAt:  item.Snippet.PublishedAt,
			ThumbnailURL: item.Snippet.Thumbnails.High.URL,
		}
	}

	return videos, nil
}

// GetChannelLiveStream gets the current live stream for a channel
func (c *YouTubeLiveClient) GetChannelLiveStream(ctx context.Context, channelID string) (*YouTubeVideo, error) {
	if channelID == "" {
		channelID = c.channelID
	}
	if channelID == "" {
		return nil, fmt.Errorf("channel ID required")
	}

	endpoint := fmt.Sprintf("/search?part=id&type=video&eventType=live&channelId=%s&maxResults=1", channelID)

	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding search results: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, nil // Channel is not live
	}

	return c.GetVideo(ctx, resp.Items[0].ID.VideoID)
}

// GetHealth returns API health status
func (c *YouTubeLiveClient) GetHealth(ctx context.Context) (*YouTubeLiveHealth, error) {
	health := &YouTubeLiveHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check API key
	if c.apiKey == "" {
		health.Score -= 50
		health.HasAPIKey = false
		health.Issues = append(health.Issues, "No API key configured")
		health.Recommendations = append(health.Recommendations, "Set YOUTUBE_API_KEY environment variable")
	} else {
		health.HasAPIKey = true
	}

	// Try API request
	if health.HasAPIKey {
		_, err := c.doRequest(ctx, "GET", "/videos?part=id&chart=mostPopular&maxResults=1", nil)
		if err != nil {
			health.Score -= 30
			health.Issues = append(health.Issues, fmt.Sprintf("API request failed: %v", err))
			health.Recommendations = append(health.Recommendations, "Check API key validity and quota")
		} else {
			health.Connected = true
		}
	}

	// Check channel ID
	if c.channelID == "" {
		health.Score -= 10
		health.Issues = append(health.Issues, "No channel ID configured (optional)")
		health.Recommendations = append(health.Recommendations, "Set YOUTUBE_CHANNEL_ID for channel-specific operations")
	}

	// Determine status
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
