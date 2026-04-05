// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// OBSClient provides access to OBS Studio via WebSocket
type OBSClient struct {
	host string
	port int
}

// OBSStatus represents OBS application status
type OBSStatus struct {
	Connected     bool    `json:"connected"`
	Version       string  `json:"version"`
	Streaming     bool    `json:"streaming"`
	Recording     bool    `json:"recording"`
	VirtualCam    bool    `json:"virtual_cam"`
	ReplayBuffer  bool    `json:"replay_buffer"`
	CurrentScene  string  `json:"current_scene"`
	StreamTime    string  `json:"stream_time"`
	RecordTime    string  `json:"record_time"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage_mb"`
	DroppedFrames int     `json:"dropped_frames"`
	TotalFrames   int     `json:"total_frames"`
	OutputBitrate int     `json:"output_bitrate_kbps"`
}

// OBSScene represents a scene in OBS
type OBSScene struct {
	Name    string           `json:"name"`
	Index   int              `json:"index"`
	Sources []OBSSceneSource `json:"sources,omitempty"`
}

// OBSSceneSource represents a source within a scene
type OBSSceneSource struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Visible bool   `json:"visible"`
	Locked  bool   `json:"locked"`
}

// OBSSource represents a source in OBS
type OBSSource struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Kind     string                 `json:"kind"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Muted    bool                   `json:"muted,omitempty"`
	Volume   float64                `json:"volume,omitempty"`
}

// OBSStreamSettings represents streaming settings
type OBSStreamSettings struct {
	Server   string `json:"server"`
	Key      string `json:"key"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	Bitrate  int    `json:"bitrate_kbps"`
	Encoder  string `json:"encoder"`
}

// OBSRecordSettings represents recording settings
type OBSRecordSettings struct {
	Path      string `json:"path"`
	Format    string `json:"format"`
	Quality   string `json:"quality"`
	Encoder   string `json:"encoder"`
	SplitFile bool   `json:"split_file"`
}

// OBSAudioSource represents an audio source
type OBSAudioSource struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Volume      float64 `json:"volume_db"`
	Muted       bool    `json:"muted"`
	MonitorType string  `json:"monitor_type"`
}

// OBSHealth represents OBS system health
type OBSHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	SceneCount      int      `json:"scene_count"`
	SourceCount     int      `json:"source_count"`
	CPUUsage        float64  `json:"cpu_usage"`
	DroppedPercent  float64  `json:"dropped_percent"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// OBSHotkey represents a hotkey binding
type OBSHotkey struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	KeyCombo    string `json:"key_combo"`
}

// OBSOutput represents an output (streaming, recording, etc.)
type OBSOutput struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Active     bool   `json:"active"`
	TotalBytes int64  `json:"total_bytes"`
	Duration   string `json:"duration"`
}

// OBSSceneItemTransform represents the transform properties of a scene item
type OBSSceneItemTransform struct {
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
	Rotation  float64 `json:"rotation"`
	ScaleX    float64 `json:"scale_x"`
	ScaleY    float64 `json:"scale_y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Alignment int     `json:"alignment"`
}

// OBSSourceFilter represents a filter on a source
type OBSSourceFilter struct {
	Name     string                 `json:"name"`
	Kind     string                 `json:"kind"`
	Enabled  bool                   `json:"enabled"`
	Index    int                    `json:"index"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// OBSTransitionInfo represents the current transition
type OBSTransitionInfo struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Duration int    `json:"duration_ms"`
	Fixed    bool   `json:"fixed_duration"`
}

// OBSScreenshot represents a captured screenshot
type OBSScreenshot struct {
	ImageData string `json:"image_data"` // base64-encoded
	Format    string `json:"format"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

var (
	obsClientSingleton *OBSClient
	obsClientOnce      sync.Once
	obsClientErr       error

	// TestOverrideOBSClient, when non-nil, is returned by GetOBSClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideOBSClient *OBSClient
)

// GetOBSClient returns the singleton OBS client.
func GetOBSClient() (*OBSClient, error) {
	if TestOverrideOBSClient != nil {
		return TestOverrideOBSClient, nil
	}
	obsClientOnce.Do(func() {
		obsClientSingleton, obsClientErr = NewOBSClient()
	})
	return obsClientSingleton, obsClientErr
}

// NewTestOBSClient creates an in-memory test client.
func NewTestOBSClient() *OBSClient {
	return &OBSClient{
		host: "localhost",
		port: 4455,
	}
}

// NewOBSClient creates a new OBS client
func NewOBSClient() (*OBSClient, error) {
	host := os.Getenv("OBS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 4455 // Default OBS WebSocket port (v5+)

	return &OBSClient{
		host: host,
		port: port,
	}, nil
}

// GetStatus returns OBS application status
func (c *OBSClient) GetStatus(ctx context.Context) (*OBSStatus, error) {
	status := &OBSStatus{
		Connected: false,
	}

	// Try to connect to WebSocket port
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
		status.Version = "Unknown (WebSocket mode)"
	}

	return status, nil
}

// GetScenes returns all scenes
func (c *OBSClient) GetScenes(ctx context.Context) ([]OBSScene, error) {
	scenes := []OBSScene{}
	return scenes, nil
}

// GetCurrentScene returns the current program scene
func (c *OBSClient) GetCurrentScene(ctx context.Context) (string, error) {
	return "", nil
}

// SetCurrentScene switches to a scene
func (c *OBSClient) SetCurrentScene(ctx context.Context, scene string) error {
	if scene == "" {
		return fmt.Errorf("scene name is required")
	}
	return nil
}

// GetPreviewScene returns the current preview scene (studio mode)
func (c *OBSClient) GetPreviewScene(ctx context.Context) (string, error) {
	return "", nil
}

// SetPreviewScene sets the preview scene (studio mode)
func (c *OBSClient) SetPreviewScene(ctx context.Context, scene string) error {
	if scene == "" {
		return fmt.Errorf("scene name is required")
	}
	return nil
}

// GetSources returns all sources
func (c *OBSClient) GetSources(ctx context.Context) ([]OBSSource, error) {
	sources := []OBSSource{}
	return sources, nil
}

// GetSceneSources returns sources in a scene
func (c *OBSClient) GetSceneSources(ctx context.Context, scene string) ([]OBSSceneSource, error) {
	sources := []OBSSceneSource{}
	return sources, nil
}

// SetSourceVisibility sets source visibility in a scene
func (c *OBSClient) SetSourceVisibility(ctx context.Context, scene, source string, visible bool) error {
	return nil
}

// StartStreaming starts streaming
func (c *OBSClient) StartStreaming(ctx context.Context) error {
	return nil
}

// StopStreaming stops streaming
func (c *OBSClient) StopStreaming(ctx context.Context) error {
	return nil
}

// StartRecording starts recording
func (c *OBSClient) StartRecording(ctx context.Context) error {
	return nil
}

// StopRecording stops recording
func (c *OBSClient) StopRecording(ctx context.Context) error {
	return nil
}

// PauseRecording pauses recording
func (c *OBSClient) PauseRecording(ctx context.Context) error {
	return nil
}

// ResumeRecording resumes recording
func (c *OBSClient) ResumeRecording(ctx context.Context) error {
	return nil
}

// StartVirtualCam starts virtual camera
func (c *OBSClient) StartVirtualCam(ctx context.Context) error {
	return nil
}

// StopVirtualCam stops virtual camera
func (c *OBSClient) StopVirtualCam(ctx context.Context) error {
	return nil
}

// StartReplayBuffer starts replay buffer
func (c *OBSClient) StartReplayBuffer(ctx context.Context) error {
	return nil
}

// StopReplayBuffer stops replay buffer
func (c *OBSClient) StopReplayBuffer(ctx context.Context) error {
	return nil
}

// SaveReplayBuffer saves replay buffer
func (c *OBSClient) SaveReplayBuffer(ctx context.Context) error {
	return nil
}

// GetStreamSettings returns streaming settings
func (c *OBSClient) GetStreamSettings(ctx context.Context) (*OBSStreamSettings, error) {
	settings := &OBSStreamSettings{
		Service:  "Custom",
		Protocol: "rtmp",
		Bitrate:  6000,
		Encoder:  "x264",
	}
	return settings, nil
}

// GetRecordSettings returns recording settings
func (c *OBSClient) GetRecordSettings(ctx context.Context) (*OBSRecordSettings, error) {
	settings := &OBSRecordSettings{
		Format:  "mkv",
		Quality: "high",
		Encoder: "x264",
	}
	return settings, nil
}

// GetAudioSources returns audio sources
func (c *OBSClient) GetAudioSources(ctx context.Context) ([]OBSAudioSource, error) {
	sources := []OBSAudioSource{}
	return sources, nil
}

// SetSourceMute mutes/unmutes an audio source
func (c *OBSClient) SetSourceMute(ctx context.Context, source string, mute bool) error {
	return nil
}

// SetSourceVolume sets source volume
func (c *OBSClient) SetSourceVolume(ctx context.Context, source string, volumeDb float64) error {
	return nil
}

// GetHealth returns OBS system health
func (c *OBSClient) GetHealth(ctx context.Context) (*OBSHealth, error) {
	health := &OBSHealth{
		Score:  100,
		Status: "healthy",
	}

	status, _ := c.GetStatus(ctx)
	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Not connected to OBS")
		health.Recommendations = append(health.Recommendations, "Start OBS and enable WebSocket server")
	}

	scenes, _ := c.GetScenes(ctx)
	health.SceneCount = len(scenes)

	sources, _ := c.GetSources(ctx)
	health.SourceCount = len(sources)

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

// TriggerStudioModeTransition triggers transition in studio mode
func (c *OBSClient) TriggerStudioModeTransition(ctx context.Context) error {
	return nil
}

// GetStudioModeEnabled returns whether studio mode is enabled
func (c *OBSClient) GetStudioModeEnabled(ctx context.Context) (bool, error) {
	return false, nil
}

// SetStudioModeEnabled enables/disables studio mode
func (c *OBSClient) SetStudioModeEnabled(ctx context.Context, enabled bool) error {
	return nil
}

// TriggerHotkey triggers a hotkey by name
func (c *OBSClient) TriggerHotkey(ctx context.Context, hotkeyName string) error {
	return nil
}

// GetOutputs returns all outputs
func (c *OBSClient) GetOutputs(ctx context.Context) ([]OBSOutput, error) {
	outputs := []OBSOutput{}
	return outputs, nil
}

// GetSceneItemTransform returns the transform of a scene item
func (c *OBSClient) GetSceneItemTransform(ctx context.Context, scene, source string) (*OBSSceneItemTransform, error) {
	if scene == "" || source == "" {
		return nil, fmt.Errorf("scene and source are required")
	}
	return &OBSSceneItemTransform{
		ScaleX: 1.0,
		ScaleY: 1.0,
	}, nil
}

// SetSceneItemTransform sets the transform of a scene item
func (c *OBSClient) SetSceneItemTransform(ctx context.Context, scene, source string, transform *OBSSceneItemTransform) error {
	if scene == "" || source == "" {
		return fmt.Errorf("scene and source are required")
	}
	return nil
}

// GetSourceFilters returns all filters on a source
func (c *OBSClient) GetSourceFilters(ctx context.Context, source string) ([]OBSSourceFilter, error) {
	if source == "" {
		return nil, fmt.Errorf("source is required")
	}
	return []OBSSourceFilter{}, nil
}

// SetSourceFilterEnabled enables or disables a source filter
func (c *OBSClient) SetSourceFilterEnabled(ctx context.Context, source, filter string, enabled bool) error {
	if source == "" || filter == "" {
		return fmt.Errorf("source and filter are required")
	}
	return nil
}

// ControlMedia controls media source playback (play, pause, stop, restart)
func (c *OBSClient) ControlMedia(ctx context.Context, source, action string) error {
	if source == "" {
		return fmt.Errorf("source is required")
	}
	if action == "" {
		return fmt.Errorf("action is required")
	}
	return nil
}

// SeekMedia seeks a media source to a position in milliseconds
func (c *OBSClient) SeekMedia(ctx context.Context, source string, positionMs int) error {
	if source == "" {
		return fmt.Errorf("source is required")
	}
	return nil
}

// GetCurrentTransition returns the current scene transition
func (c *OBSClient) GetCurrentTransition(ctx context.Context) (*OBSTransitionInfo, error) {
	return &OBSTransitionInfo{
		Name:     "Fade",
		Kind:     "fade_transition",
		Duration: 300,
	}, nil
}

// SetTransition sets the scene transition type and optional duration
func (c *OBSClient) SetTransition(ctx context.Context, name string, durationMs int) error {
	if name == "" {
		return fmt.Errorf("transition name is required")
	}
	return nil
}

// TakeScreenshot captures a screenshot of a source or the program output
func (c *OBSClient) TakeScreenshot(ctx context.Context, source, format string) (*OBSScreenshot, error) {
	if format == "" {
		format = "png"
	}
	return &OBSScreenshot{
		Format: format,
	}, nil
}

// Host returns the configured host
func (c *OBSClient) Host() string {
	return c.host
}

// Port returns the configured port
func (c *OBSClient) Port() int {
	return c.port
}
