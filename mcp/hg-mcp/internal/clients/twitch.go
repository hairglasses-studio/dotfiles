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
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// TwitchClient provides access to Twitch Helix API
type TwitchClient struct {
	clientID      string
	clientSecret  string
	accessToken   string
	tokenExpiry   time.Time
	broadcasterID string
	httpClient    *http.Client
}

// TwitchStatus represents stream and connection status
type TwitchStatus struct {
	Connected     bool   `json:"connected"`
	IsLive        bool   `json:"is_live"`
	StreamTitle   string `json:"stream_title,omitempty"`
	GameName      string `json:"game_name,omitempty"`
	ViewerCount   int    `json:"viewer_count,omitempty"`
	StartedAt     string `json:"started_at,omitempty"`
	ThumbnailURL  string `json:"thumbnail_url,omitempty"`
	BroadcasterID string `json:"broadcaster_id"`
}

// TwitchStream represents stream information
type TwitchStream struct {
	ID           string   `json:"id"`
	UserID       string   `json:"user_id"`
	UserLogin    string   `json:"user_login"`
	UserName     string   `json:"user_name"`
	GameID       string   `json:"game_id"`
	GameName     string   `json:"game_name"`
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	ViewerCount  int      `json:"viewer_count"`
	StartedAt    string   `json:"started_at"`
	Language     string   `json:"language"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Tags         []string `json:"tags"`
	IsMature     bool     `json:"is_mature"`
}

// TwitchChatMessage represents a chat message
type TwitchChatMessage struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	Message      string `json:"message"`
	Timestamp    string `json:"timestamp"`
	Color        string `json:"color,omitempty"`
	IsMod        bool   `json:"is_mod"`
	IsSubscriber bool   `json:"is_subscriber"`
}

// TwitchPoll represents a poll
type TwitchPoll struct {
	ID            string             `json:"id"`
	BroadcasterID string             `json:"broadcaster_id"`
	Title         string             `json:"title"`
	Choices       []TwitchPollChoice `json:"choices"`
	Status        string             `json:"status"`
	Duration      int                `json:"duration"`
	StartedAt     string             `json:"started_at"`
	EndedAt       string             `json:"ended_at,omitempty"`
}

// TwitchPollChoice represents a poll choice
type TwitchPollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Votes              int    `json:"votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
}

// TwitchPrediction represents a prediction
type TwitchPrediction struct {
	ID               string                    `json:"id"`
	BroadcasterID    string                    `json:"broadcaster_id"`
	Title            string                    `json:"title"`
	Outcomes         []TwitchPredictionOutcome `json:"outcomes"`
	PredictionWindow int                       `json:"prediction_window"`
	Status           string                    `json:"status"`
	CreatedAt        string                    `json:"created_at"`
	EndedAt          string                    `json:"ended_at,omitempty"`
	LockedAt         string                    `json:"locked_at,omitempty"`
}

// TwitchPredictionOutcome represents a prediction outcome
type TwitchPredictionOutcome struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Users         int    `json:"users"`
	ChannelPoints int    `json:"channel_points"`
	Color         string `json:"color"`
}

// TwitchClip represents a clip
type TwitchClip struct {
	ID            string  `json:"id"`
	URL           string  `json:"url"`
	EmbedURL      string  `json:"embed_url"`
	BroadcasterID string  `json:"broadcaster_id"`
	CreatorID     string  `json:"creator_id"`
	VideoID       string  `json:"video_id"`
	GameID        string  `json:"game_id"`
	Title         string  `json:"title"`
	ViewCount     int     `json:"view_count"`
	CreatedAt     string  `json:"created_at"`
	ThumbnailURL  string  `json:"thumbnail_url"`
	Duration      float64 `json:"duration"`
}

// TwitchMarker represents a stream marker
type TwitchMarker struct {
	ID              string `json:"id"`
	CreatedAt       string `json:"created_at"`
	Description     string `json:"description"`
	PositionSeconds int    `json:"position_seconds"`
}

// TwitchUser represents a user for ban/timeout
type TwitchUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

// TwitchHealth represents API health status
type TwitchHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	HasValidToken   bool     `json:"has_valid_token"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewTwitchClient creates a new Twitch client
func NewTwitchClient() (*TwitchClient, error) {
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_ID environment variable required")
	}

	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("TWITCH_CLIENT_SECRET environment variable required")
	}

	broadcasterID := os.Getenv("TWITCH_BROADCASTER_ID")
	if broadcasterID == "" {
		broadcasterID = os.Getenv("TWITCH_CHANNEL_ID") // Alias
	}

	return &TwitchClient{
		clientID:      clientID,
		clientSecret:  clientSecret,
		broadcasterID: broadcasterID,
		httpClient: httpclient.Standard(),
	}, nil
}

// BroadcasterID returns the configured broadcaster ID
func (c *TwitchClient) BroadcasterID() string {
	return c.broadcasterID
}

// ensureToken ensures we have a valid access token
func (c *TwitchClient) ensureToken(ctx context.Context) error {
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// Get app access token using client credentials flow
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token request failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return nil
}

// doRequest performs an authenticated API request
func (c *TwitchClient) doRequest(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	url := "https://api.twitch.tv/helix" + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Client-Id", c.clientID)
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

// GetStatus returns stream and connection status
func (c *TwitchClient) GetStatus(ctx context.Context) (*TwitchStatus, error) {
	status := &TwitchStatus{
		Connected:     false,
		BroadcasterID: c.broadcasterID,
	}

	if err := c.ensureToken(ctx); err != nil {
		return status, nil
	}
	status.Connected = true

	// Get stream info
	stream, err := c.GetStream(ctx, c.broadcasterID)
	if err != nil {
		return status, nil
	}

	if stream != nil {
		status.IsLive = true
		status.StreamTitle = stream.Title
		status.GameName = stream.GameName
		status.ViewerCount = stream.ViewerCount
		status.StartedAt = stream.StartedAt
		status.ThumbnailURL = stream.ThumbnailURL
	}

	return status, nil
}

// GetStream gets current stream info for a broadcaster
func (c *TwitchClient) GetStream(ctx context.Context, broadcasterID string) (*TwitchStream, error) {
	if broadcasterID == "" {
		broadcasterID = c.broadcasterID
	}

	data, err := c.doRequest(ctx, "GET", "/streams?user_id="+broadcasterID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchStream `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding stream: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, nil // Not live
	}

	return &resp.Data[0], nil
}

// UpdateStreamInfo updates stream title and/or game
func (c *TwitchClient) UpdateStreamInfo(ctx context.Context, title, gameID string) error {
	payload := map[string]interface{}{
		"broadcaster_id": c.broadcasterID,
	}
	if title != "" {
		payload["title"] = title
	}
	if gameID != "" {
		payload["game_id"] = gameID
	}

	body, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PATCH", "/channels?broadcaster_id="+c.broadcasterID, strings.NewReader(string(body)))
	return err
}

// SearchGames searches for games by name
func (c *TwitchClient) SearchGames(ctx context.Context, query string) ([]map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/search/categories?query="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// SendChatMessage sends a message to chat (requires user access token)
func (c *TwitchClient) SendChatMessage(ctx context.Context, message string) error {
	payload := map[string]string{
		"broadcaster_id": c.broadcasterID,
		"sender_id":      c.broadcasterID,
		"message":        message,
	}
	body, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "POST", "/chat/messages", strings.NewReader(string(body)))
	return err
}

// BanUser bans a user from chat
func (c *TwitchClient) BanUser(ctx context.Context, userID, reason string) error {
	payload := map[string]interface{}{
		"data": map[string]string{
			"user_id": userID,
			"reason":  reason,
		},
	}
	body, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "POST", "/moderation/bans?broadcaster_id="+c.broadcasterID+"&moderator_id="+c.broadcasterID, strings.NewReader(string(body)))
	return err
}

// TimeoutUser times out a user from chat
func (c *TwitchClient) TimeoutUser(ctx context.Context, userID string, duration int, reason string) error {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"user_id":  userID,
			"duration": duration,
			"reason":   reason,
		},
	}
	body, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "POST", "/moderation/bans?broadcaster_id="+c.broadcasterID+"&moderator_id="+c.broadcasterID, strings.NewReader(string(body)))
	return err
}

// UnbanUser unbans a user from chat
func (c *TwitchClient) UnbanUser(ctx context.Context, userID string) error {
	_, err := c.doRequest(ctx, "DELETE", "/moderation/bans?broadcaster_id="+c.broadcasterID+"&moderator_id="+c.broadcasterID+"&user_id="+userID, nil)
	return err
}

// CreatePoll creates a new poll
func (c *TwitchClient) CreatePoll(ctx context.Context, title string, choices []string, duration int) (*TwitchPoll, error) {
	choiceData := make([]map[string]string, len(choices))
	for i, choice := range choices {
		choiceData[i] = map[string]string{"title": choice}
	}

	payload := map[string]interface{}{
		"broadcaster_id": c.broadcasterID,
		"title":          title,
		"choices":        choiceData,
		"duration":       duration,
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "POST", "/polls", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPoll `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no poll returned")
	}

	return &resp.Data[0], nil
}

// EndPoll ends a poll
func (c *TwitchClient) EndPoll(ctx context.Context, pollID string, showResults bool) (*TwitchPoll, error) {
	status := "TERMINATED"
	if showResults {
		status = "ARCHIVED"
	}

	payload := map[string]interface{}{
		"broadcaster_id": c.broadcasterID,
		"id":             pollID,
		"status":         status,
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "PATCH", "/polls", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPoll `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no poll returned")
	}

	return &resp.Data[0], nil
}

// GetPolls gets active or recent polls
func (c *TwitchClient) GetPolls(ctx context.Context) ([]TwitchPoll, error) {
	data, err := c.doRequest(ctx, "GET", "/polls?broadcaster_id="+c.broadcasterID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPoll `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// CreatePrediction creates a new prediction
func (c *TwitchClient) CreatePrediction(ctx context.Context, title string, outcomes []string, duration int) (*TwitchPrediction, error) {
	outcomeData := make([]map[string]string, len(outcomes))
	for i, outcome := range outcomes {
		outcomeData[i] = map[string]string{"title": outcome}
	}

	payload := map[string]interface{}{
		"broadcaster_id":    c.broadcasterID,
		"title":             title,
		"outcomes":          outcomeData,
		"prediction_window": duration,
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "POST", "/predictions", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPrediction `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no prediction returned")
	}

	return &resp.Data[0], nil
}

// ResolvePrediction resolves a prediction with a winning outcome
func (c *TwitchClient) ResolvePrediction(ctx context.Context, predictionID, winningOutcomeID string) (*TwitchPrediction, error) {
	payload := map[string]interface{}{
		"broadcaster_id":     c.broadcasterID,
		"id":                 predictionID,
		"status":             "RESOLVED",
		"winning_outcome_id": winningOutcomeID,
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "PATCH", "/predictions", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPrediction `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no prediction returned")
	}

	return &resp.Data[0], nil
}

// CancelPrediction cancels a prediction
func (c *TwitchClient) CancelPrediction(ctx context.Context, predictionID string) (*TwitchPrediction, error) {
	payload := map[string]interface{}{
		"broadcaster_id": c.broadcasterID,
		"id":             predictionID,
		"status":         "CANCELED",
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "PATCH", "/predictions", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPrediction `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no prediction returned")
	}

	return &resp.Data[0], nil
}

// GetPredictions gets active or recent predictions
func (c *TwitchClient) GetPredictions(ctx context.Context) ([]TwitchPrediction, error) {
	data, err := c.doRequest(ctx, "GET", "/predictions?broadcaster_id="+c.broadcasterID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchPrediction `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// CreateClip creates a clip from the current stream
func (c *TwitchClient) CreateClip(ctx context.Context) (*TwitchClip, error) {
	data, err := c.doRequest(ctx, "POST", "/clips?broadcaster_id="+c.broadcasterID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []struct {
			ID      string `json:"id"`
			EditURL string `json:"edit_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no clip created")
	}

	// Get full clip info
	return c.GetClip(ctx, resp.Data[0].ID)
}

// GetClip gets clip details
func (c *TwitchClient) GetClip(ctx context.Context, clipID string) (*TwitchClip, error) {
	data, err := c.doRequest(ctx, "GET", "/clips?id="+clipID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchClip `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("clip not found")
	}

	return &resp.Data[0], nil
}

// CreateStreamMarker creates a marker in the stream
func (c *TwitchClient) CreateStreamMarker(ctx context.Context, description string) (*TwitchMarker, error) {
	payload := map[string]string{
		"user_id":     c.broadcasterID,
		"description": description,
	}
	body, _ := json.Marshal(payload)

	data, err := c.doRequest(ctx, "POST", "/streams/markers", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchMarker `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no marker created")
	}

	return &resp.Data[0], nil
}

// StartRaid starts a raid to another channel
func (c *TwitchClient) StartRaid(ctx context.Context, toUserID string) error {
	_, err := c.doRequest(ctx, "POST", "/raids?from_broadcaster_id="+c.broadcasterID+"&to_broadcaster_id="+toUserID, nil)
	return err
}

// CancelRaid cancels an active raid
func (c *TwitchClient) CancelRaid(ctx context.Context) error {
	_, err := c.doRequest(ctx, "DELETE", "/raids?broadcaster_id="+c.broadcasterID, nil)
	return err
}

// GetUserByLogin gets user info by login name
func (c *TwitchClient) GetUserByLogin(ctx context.Context, login string) (*TwitchUser, error) {
	data, err := c.doRequest(ctx, "GET", "/users?login="+url.QueryEscape(login), nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data []TwitchUser `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("user not found: %s", login)
	}

	return &resp.Data[0], nil
}

// GetHealth returns API health status
func (c *TwitchClient) GetHealth(ctx context.Context) (*TwitchHealth, error) {
	health := &TwitchHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check token
	if err := c.ensureToken(ctx); err != nil {
		health.Score -= 50
		health.HasValidToken = false
		health.Issues = append(health.Issues, fmt.Sprintf("Token error: %v", err))
		health.Recommendations = append(health.Recommendations, "Check TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET")
	} else {
		health.Connected = true
		health.HasValidToken = true
	}

	// Check broadcaster ID
	if c.broadcasterID == "" {
		health.Score -= 20
		health.Issues = append(health.Issues, "No broadcaster ID configured")
		health.Recommendations = append(health.Recommendations, "Set TWITCH_BROADCASTER_ID or TWITCH_CHANNEL_ID")
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
