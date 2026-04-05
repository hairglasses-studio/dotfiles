// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// ShowkontrolClient provides control of Showkontrol for timecode and cue management
type ShowkontrolClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// ShowkontrolStatus represents system status
type ShowkontrolStatus struct {
	Connected   bool             `json:"connected"`
	Version     string           `json:"version,omitempty"`
	CurrentShow *ShowkontrolShow `json:"current_show,omitempty"`
	Timecode    *TimecodeStatus  `json:"timecode,omitempty"`
	BaseURL     string           `json:"base_url"`
}

// ShowkontrolShow represents a show
type ShowkontrolShow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CueCount    int       `json:"cue_count"`
	Duration    float64   `json:"duration,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// ShowkontrolCue represents a cue in a show
type ShowkontrolCue struct {
	ID          string      `json:"id"`
	Number      string      `json:"number"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Time        float64     `json:"time"` // Timecode position in seconds
	Duration    float64     `json:"duration,omitempty"`
	CueType     string      `json:"cue_type"` // go, stop, pause, etc.
	Actions     []CueAction `json:"actions,omitempty"`
	Fired       bool        `json:"fired"`
	LastFired   time.Time   `json:"last_fired,omitempty"`
}

// CueAction represents an action within a cue
type CueAction struct {
	Type       string                 `json:"type"` // osc, midi, http, script
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Delay      float64                `json:"delay,omitempty"`
}

// TimecodeStatus represents timecode status
type TimecodeStatus struct {
	Running    bool    `json:"running"`
	Position   float64 `json:"position"`    // Current position in seconds
	PositionTC string  `json:"position_tc"` // Formatted timecode
	Format     string  `json:"format"`      // SMPTE format (25, 29.97df, 30, etc.)
	Source     string  `json:"source"`      // internal, ltc, mtc
	FrameRate  float64 `json:"frame_rate"`
}

// ShowkontrolHealth represents health status
type ShowkontrolHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	APIReachable    bool     `json:"api_reachable"`
	TimecodeRunning bool     `json:"timecode_running"`
	ShowLoaded      bool     `json:"show_loaded"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewShowkontrolClient creates a new Showkontrol client
func NewShowkontrolClient() (*ShowkontrolClient, error) {
	host := os.Getenv("SHOWKONTROL_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("SHOWKONTROL_PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)

	apiKey := os.Getenv("SHOWKONTROL_API_KEY")

	return &ShowkontrolClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: httpclient.Standard(),
	}, nil
}

// doRequest performs an HTTP request
func (c *ShowkontrolClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetStatus returns system status
func (c *ShowkontrolClient) GetStatus(ctx context.Context) (*ShowkontrolStatus, error) {
	status := &ShowkontrolStatus{
		BaseURL: c.baseURL,
	}

	body, err := c.doRequest(ctx, "GET", "/api/status", nil)
	if err != nil {
		status.Connected = false
		return status, nil
	}

	status.Connected = true

	var apiStatus struct {
		Version     string `json:"version"`
		CurrentShow struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			CueCount int    `json:"cue_count"`
		} `json:"current_show"`
		Timecode struct {
			Running   bool    `json:"running"`
			Position  float64 `json:"position"`
			Format    string  `json:"format"`
			FrameRate float64 `json:"frame_rate"`
		} `json:"timecode"`
	}

	if err := json.Unmarshal(body, &apiStatus); err == nil {
		status.Version = apiStatus.Version

		if apiStatus.CurrentShow.ID != "" {
			status.CurrentShow = &ShowkontrolShow{
				ID:       apiStatus.CurrentShow.ID,
				Name:     apiStatus.CurrentShow.Name,
				CueCount: apiStatus.CurrentShow.CueCount,
			}
		}

		status.Timecode = &TimecodeStatus{
			Running:    apiStatus.Timecode.Running,
			Position:   apiStatus.Timecode.Position,
			Format:     apiStatus.Timecode.Format,
			FrameRate:  apiStatus.Timecode.FrameRate,
			PositionTC: formatTimecode(apiStatus.Timecode.Position, apiStatus.Timecode.FrameRate),
		}
	}

	return status, nil
}

// formatTimecode formats seconds as timecode
func formatTimecode(seconds, frameRate float64) string {
	if frameRate == 0 {
		frameRate = 25
	}

	totalFrames := int(seconds * frameRate)
	frames := totalFrames % int(frameRate)
	totalSeconds := totalFrames / int(frameRate)
	secs := totalSeconds % 60
	mins := (totalSeconds / 60) % 60
	hours := totalSeconds / 3600

	return fmt.Sprintf("%02d:%02d:%02d:%02d", hours, mins, secs, frames)
}

// parseTimecode parses timecode string to seconds
func parseTimecode(tc string, frameRate float64) float64 {
	if frameRate == 0 {
		frameRate = 25
	}

	var hours, mins, secs, frames int
	fmt.Sscanf(tc, "%d:%d:%d:%d", &hours, &mins, &secs, &frames)

	totalSeconds := float64(hours*3600 + mins*60 + secs)
	totalSeconds += float64(frames) / frameRate

	return totalSeconds
}

// GetShows returns all shows
func (c *ShowkontrolClient) GetShows(ctx context.Context) ([]ShowkontrolShow, error) {
	body, err := c.doRequest(ctx, "GET", "/api/shows", nil)
	if err != nil {
		return nil, err
	}

	var shows []ShowkontrolShow
	if err := json.Unmarshal(body, &shows); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return shows, nil
}

// GetShow returns a specific show
func (c *ShowkontrolClient) GetShow(ctx context.Context, showID string) (*ShowkontrolShow, error) {
	body, err := c.doRequest(ctx, "GET", "/api/shows/"+showID, nil)
	if err != nil {
		return nil, err
	}

	var show ShowkontrolShow
	if err := json.Unmarshal(body, &show); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &show, nil
}

// LoadShow loads a show
func (c *ShowkontrolClient) LoadShow(ctx context.Context, showID string) error {
	_, err := c.doRequest(ctx, "POST", "/api/shows/"+showID+"/load", nil)
	return err
}

// GetCues returns cues for the current show
func (c *ShowkontrolClient) GetCues(ctx context.Context) ([]ShowkontrolCue, error) {
	body, err := c.doRequest(ctx, "GET", "/api/cues", nil)
	if err != nil {
		return nil, err
	}

	var cues []ShowkontrolCue
	if err := json.Unmarshal(body, &cues); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return cues, nil
}

// GetCue returns a specific cue
func (c *ShowkontrolClient) GetCue(ctx context.Context, cueID string) (*ShowkontrolCue, error) {
	body, err := c.doRequest(ctx, "GET", "/api/cues/"+cueID, nil)
	if err != nil {
		return nil, err
	}

	var cue ShowkontrolCue
	if err := json.Unmarshal(body, &cue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &cue, nil
}

// FireCue fires a specific cue
func (c *ShowkontrolClient) FireCue(ctx context.Context, cueID string) error {
	_, err := c.doRequest(ctx, "POST", "/api/cues/"+cueID+"/go", nil)
	return err
}

// Go fires the next cue
func (c *ShowkontrolClient) Go(ctx context.Context) error {
	_, err := c.doRequest(ctx, "POST", "/api/go", nil)
	return err
}

// Stop stops the current cue
func (c *ShowkontrolClient) Stop(ctx context.Context) error {
	_, err := c.doRequest(ctx, "POST", "/api/stop", nil)
	return err
}

// Pause pauses the current cue
func (c *ShowkontrolClient) Pause(ctx context.Context) error {
	_, err := c.doRequest(ctx, "POST", "/api/pause", nil)
	return err
}

// GetTimecodeStatus returns timecode status
func (c *ShowkontrolClient) GetTimecodeStatus(ctx context.Context) (*TimecodeStatus, error) {
	body, err := c.doRequest(ctx, "GET", "/api/timecode/status", nil)
	if err != nil {
		return nil, err
	}

	var status TimecodeStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	status.PositionTC = formatTimecode(status.Position, status.FrameRate)

	return &status, nil
}

// StartTimecode starts timecode playback
func (c *ShowkontrolClient) StartTimecode(ctx context.Context) error {
	_, err := c.doRequest(ctx, "POST", "/api/timecode/start", nil)
	return err
}

// StopTimecode stops timecode playback
func (c *ShowkontrolClient) StopTimecode(ctx context.Context) error {
	_, err := c.doRequest(ctx, "POST", "/api/timecode/stop", nil)
	return err
}

// GotoTimecode jumps to a specific timecode position
func (c *ShowkontrolClient) GotoTimecode(ctx context.Context, position string) error {
	// Position can be in seconds or timecode format
	var posSeconds float64

	if strings.Contains(position, ":") {
		// Parse timecode format
		posSeconds = parseTimecode(position, 25) // Default 25fps
	} else {
		fmt.Sscanf(position, "%f", &posSeconds)
	}

	payload := fmt.Sprintf(`{"position": %f}`, posSeconds)
	_, err := c.doRequest(ctx, "POST", "/api/timecode/goto", strings.NewReader(payload))
	return err
}

// SetTimecodePosition sets timecode position in seconds
func (c *ShowkontrolClient) SetTimecodePosition(ctx context.Context, seconds float64) error {
	payload := fmt.Sprintf(`{"position": %f}`, seconds)
	_, err := c.doRequest(ctx, "POST", "/api/timecode/goto", strings.NewReader(payload))
	return err
}

// GetHealth returns health status
func (c *ShowkontrolClient) GetHealth(ctx context.Context) (*ShowkontrolHealth, error) {
	health := &ShowkontrolHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check API reachability
	status, err := c.GetStatus(ctx)
	if err != nil || !status.Connected {
		health.APIReachable = false
		health.Score -= 50
		health.Issues = append(health.Issues, "Showkontrol API not reachable")
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Ensure Showkontrol is running on %s", c.baseURL))
	} else {
		health.APIReachable = true

		// Check if show is loaded
		if status.CurrentShow != nil {
			health.ShowLoaded = true
		} else {
			health.Score -= 20
			health.ShowLoaded = false
			health.Issues = append(health.Issues, "No show loaded")
			health.Recommendations = append(health.Recommendations,
				"Load a show using aftrs_showkontrol_show")
		}

		// Check timecode status
		if status.Timecode != nil {
			health.TimecodeRunning = status.Timecode.Running
		}
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
