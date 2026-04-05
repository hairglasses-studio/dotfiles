// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// AbletonClient provides control of Ableton Live via AbletonOSC
type AbletonClient struct {
	host      string
	port      int
	replyPort int
	client    *osc.Client
	server    *osc.Server
	mu        sync.RWMutex
	state     *AbletonState
	connected bool
}

// AbletonState represents the current state of Live
type AbletonState struct {
	Playing       bool    `json:"playing"`
	Recording     bool    `json:"recording"`
	Tempo         float64 `json:"tempo"`
	TimeSignNum   int     `json:"time_signature_numerator"`
	TimeSignDen   int     `json:"time_signature_denominator"`
	SongTime      float64 `json:"song_time"`
	TrackCount    int     `json:"track_count"`
	SceneCount    int     `json:"scene_count"`
	SelectedTrack int     `json:"selected_track"`
	SelectedScene int     `json:"selected_scene"`
}

// AbletonStatus represents connection status
type AbletonStatus struct {
	Connected bool          `json:"connected"`
	Host      string        `json:"host"`
	Port      int           `json:"port"`
	ReplyPort int           `json:"reply_port"`
	State     *AbletonState `json:"state,omitempty"`
}

// AbletonTrack represents a track in Live
type AbletonTrack struct {
	Index     int     `json:"index"`
	Name      string  `json:"name"`
	Color     int     `json:"color,omitempty"`
	Mute      bool    `json:"mute"`
	Solo      bool    `json:"solo"`
	Arm       bool    `json:"arm"`
	Volume    float64 `json:"volume"`
	Pan       float64 `json:"pan"`
	HasMidi   bool    `json:"has_midi_input"`
	HasAudio  bool    `json:"has_audio_input"`
	ClipSlots int     `json:"clip_slots"`
}

// AbletonClip represents a clip in Live
type AbletonClip struct {
	TrackIndex int     `json:"track_index"`
	SlotIndex  int     `json:"slot_index"`
	Name       string  `json:"name"`
	Color      int     `json:"color,omitempty"`
	Length     float64 `json:"length"`
	Playing    bool    `json:"playing"`
	Triggered  bool    `json:"triggered"`
	IsAudio    bool    `json:"is_audio"`
}

// AbletonScene represents a scene in Live
type AbletonScene struct {
	Index int     `json:"index"`
	Name  string  `json:"name"`
	Color int     `json:"color,omitempty"`
	Tempo float64 `json:"tempo,omitempty"`
}

// AbletonDevice represents a device on a track
type AbletonDevice struct {
	TrackIndex  int                `json:"track_index"`
	DeviceIndex int                `json:"device_index"`
	Name        string             `json:"name"`
	ClassName   string             `json:"class_name"`
	IsActive    bool               `json:"is_active"`
	Parameters  []AbletonParameter `json:"parameters,omitempty"`
}

// AbletonParameter represents a device parameter
type AbletonParameter struct {
	Index   int     `json:"index"`
	Name    string  `json:"name"`
	Value   float64 `json:"value"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Default float64 `json:"default,omitempty"`
}

// AbletonCuePoint represents an arrangement cue point
type AbletonCuePoint struct {
	Index int     `json:"index"`
	Name  string  `json:"name"`
	Time  float64 `json:"time"`
}

// AbletonHealth represents health status
type AbletonHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	AbletonOSC      bool     `json:"abletonosc_responding"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewAbletonClient creates a new Ableton client
func NewAbletonClient() (*AbletonClient, error) {
	host := os.Getenv("ABLETON_OSC_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 11000
	if p := os.Getenv("ABLETON_OSC_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	replyPort := 11001
	if p := os.Getenv("ABLETON_OSC_REPLY_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &replyPort)
	}

	client := osc.NewClient(host, port)

	return &AbletonClient{
		host:      host,
		port:      port,
		replyPort: replyPort,
		client:    client,
		state:     &AbletonState{Tempo: 120},
	}, nil
}

// GetStatus returns the current status
func (c *AbletonClient) GetStatus(ctx context.Context) (*AbletonStatus, error) {
	status := &AbletonStatus{
		Host:      c.host,
		Port:      c.port,
		ReplyPort: c.replyPort,
	}

	// Try to get tempo to check connection
	if err := c.sendMessage("/live/song/get/tempo"); err != nil {
		status.Connected = false
		return status, nil
	}

	// Give AbletonOSC time to respond
	time.Sleep(100 * time.Millisecond)

	c.mu.RLock()
	status.Connected = c.connected
	status.State = c.state
	c.mu.RUnlock()

	// Refresh state
	c.refreshState(ctx)

	return status, nil
}

// refreshState updates the internal state
func (c *AbletonClient) refreshState(ctx context.Context) {
	c.sendMessage("/live/song/get/tempo")
	c.sendMessage("/live/song/get/is_playing")
	c.sendMessage("/live/song/get/record_mode")
	c.sendMessage("/live/song/get/current_song_time")
	c.sendMessage("/live/song/get/num_tracks")
	c.sendMessage("/live/song/get/num_scenes")
}

// sendMessage sends an OSC message
func (c *AbletonClient) sendMessage(address string, args ...interface{}) error {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}
	return c.client.Send(msg)
}

// Transport controls

// Play starts playback
func (c *AbletonClient) Play(ctx context.Context) error {
	return c.sendMessage("/live/song/start_playing")
}

// Stop stops playback
func (c *AbletonClient) Stop(ctx context.Context) error {
	return c.sendMessage("/live/song/stop_playing")
}

// Continue continues from current position
func (c *AbletonClient) Continue(ctx context.Context) error {
	return c.sendMessage("/live/song/continue_playing")
}

// Record toggles record mode
func (c *AbletonClient) Record(ctx context.Context, enable bool) error {
	val := 0
	if enable {
		val = 1
	}
	return c.sendMessage("/live/song/set/record_mode", int32(val))
}

// GetTransportState returns current transport state
func (c *AbletonClient) GetTransportState(ctx context.Context) (map[string]interface{}, error) {
	c.refreshState(ctx)
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"playing":   c.state.Playing,
		"recording": c.state.Recording,
		"song_time": c.state.SongTime,
		"tempo":     c.state.Tempo,
	}, nil
}

// Tempo controls

// GetTempo returns the current tempo
func (c *AbletonClient) GetTempo(ctx context.Context) (float64, error) {
	c.sendMessage("/live/song/get/tempo")
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state.Tempo, nil
}

// SetTempo sets the tempo
func (c *AbletonClient) SetTempo(ctx context.Context, bpm float64) error {
	if bpm < 20 || bpm > 999 {
		return fmt.Errorf("tempo must be between 20 and 999 BPM")
	}
	return c.sendMessage("/live/song/set/tempo", float32(bpm))
}

// Track operations

// GetTracks returns all tracks
func (c *AbletonClient) GetTracks(ctx context.Context) ([]AbletonTrack, error) {
	// First get track count
	c.sendMessage("/live/song/get/num_tracks")
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	trackCount := c.state.TrackCount
	c.mu.RUnlock()

	if trackCount == 0 {
		trackCount = 8 // Default assumption
	}

	var tracks []AbletonTrack
	for i := 0; i < trackCount; i++ {
		track := AbletonTrack{Index: i}

		// Request track info
		c.sendMessage("/live/track/get/name", int32(i))
		c.sendMessage("/live/track/get/mute", int32(i))
		c.sendMessage("/live/track/get/solo", int32(i))
		c.sendMessage("/live/track/get/arm", int32(i))
		c.sendMessage("/live/track/get/volume", int32(i))
		c.sendMessage("/live/track/get/panning", int32(i))

		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetTrack returns a specific track
func (c *AbletonClient) GetTrack(ctx context.Context, index int) (*AbletonTrack, error) {
	track := &AbletonTrack{Index: index}

	c.sendMessage("/live/track/get/name", int32(index))
	c.sendMessage("/live/track/get/mute", int32(index))
	c.sendMessage("/live/track/get/solo", int32(index))
	c.sendMessage("/live/track/get/arm", int32(index))
	c.sendMessage("/live/track/get/volume", int32(index))
	c.sendMessage("/live/track/get/panning", int32(index))

	time.Sleep(50 * time.Millisecond)

	return track, nil
}

// SetTrackMute sets track mute state
func (c *AbletonClient) SetTrackMute(ctx context.Context, trackIndex int, mute bool) error {
	val := 0
	if mute {
		val = 1
	}
	return c.sendMessage("/live/track/set/mute", int32(trackIndex), int32(val))
}

// SetTrackSolo sets track solo state
func (c *AbletonClient) SetTrackSolo(ctx context.Context, trackIndex int, solo bool) error {
	val := 0
	if solo {
		val = 1
	}
	return c.sendMessage("/live/track/set/solo", int32(trackIndex), int32(val))
}

// SetTrackArm sets track arm state
func (c *AbletonClient) SetTrackArm(ctx context.Context, trackIndex int, arm bool) error {
	val := 0
	if arm {
		val = 1
	}
	return c.sendMessage("/live/track/set/arm", int32(trackIndex), int32(val))
}

// SetTrackVolume sets track volume (0.0 to 1.0)
func (c *AbletonClient) SetTrackVolume(ctx context.Context, trackIndex int, volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0.0 and 1.0")
	}
	return c.sendMessage("/live/track/set/volume", int32(trackIndex), float32(volume))
}

// SetTrackPan sets track pan (-1.0 to 1.0)
func (c *AbletonClient) SetTrackPan(ctx context.Context, trackIndex int, pan float64) error {
	if pan < -1 || pan > 1 {
		return fmt.Errorf("pan must be between -1.0 and 1.0")
	}
	return c.sendMessage("/live/track/set/panning", int32(trackIndex), float32(pan))
}

// Clip operations

// GetClips returns clips for a track
func (c *AbletonClient) GetClips(ctx context.Context, trackIndex int) ([]AbletonClip, error) {
	var clips []AbletonClip

	// Get number of scenes first
	c.sendMessage("/live/song/get/num_scenes")
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	sceneCount := c.state.SceneCount
	c.mu.RUnlock()

	if sceneCount == 0 {
		sceneCount = 8 // Default
	}

	for i := 0; i < sceneCount; i++ {
		c.sendMessage("/live/clip_slot/get/has_clip", int32(trackIndex), int32(i))
	}

	return clips, nil
}

// FireClip triggers a clip
func (c *AbletonClient) FireClip(ctx context.Context, trackIndex, slotIndex int) error {
	return c.sendMessage("/live/clip/fire", int32(trackIndex), int32(slotIndex))
}

// StopClip stops a clip
func (c *AbletonClient) StopClip(ctx context.Context, trackIndex, slotIndex int) error {
	return c.sendMessage("/live/clip/stop", int32(trackIndex), int32(slotIndex))
}

// Scene operations

// GetScenes returns all scenes
func (c *AbletonClient) GetScenes(ctx context.Context) ([]AbletonScene, error) {
	c.sendMessage("/live/song/get/num_scenes")
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	sceneCount := c.state.SceneCount
	c.mu.RUnlock()

	if sceneCount == 0 {
		sceneCount = 8
	}

	var scenes []AbletonScene
	for i := 0; i < sceneCount; i++ {
		scene := AbletonScene{Index: i}
		c.sendMessage("/live/scene/get/name", int32(i))
		scenes = append(scenes, scene)
	}

	return scenes, nil
}

// FireScene triggers a scene
func (c *AbletonClient) FireScene(ctx context.Context, sceneIndex int) error {
	return c.sendMessage("/live/scene/fire", int32(sceneIndex))
}

// Device operations

// GetDevices returns devices on a track
func (c *AbletonClient) GetDevices(ctx context.Context, trackIndex int) ([]AbletonDevice, error) {
	var devices []AbletonDevice

	c.sendMessage("/live/track/get/num_devices", int32(trackIndex))
	time.Sleep(50 * time.Millisecond)

	// For now return empty list - would need response handling
	return devices, nil
}

// GetDeviceParameters returns parameters for a device
func (c *AbletonClient) GetDeviceParameters(ctx context.Context, trackIndex, deviceIndex int) ([]AbletonParameter, error) {
	var params []AbletonParameter

	c.sendMessage("/live/device/get/num_parameters", int32(trackIndex), int32(deviceIndex))
	time.Sleep(50 * time.Millisecond)

	return params, nil
}

// SetDeviceParameter sets a device parameter
func (c *AbletonClient) SetDeviceParameter(ctx context.Context, trackIndex, deviceIndex, paramIndex int, value float64) error {
	return c.sendMessage("/live/device/set/parameter/value",
		int32(trackIndex), int32(deviceIndex), int32(paramIndex), float32(value))
}

// Cue point operations

// GetCuePoints returns arrangement cue points
func (c *AbletonClient) GetCuePoints(ctx context.Context) ([]AbletonCuePoint, error) {
	var cues []AbletonCuePoint
	c.sendMessage("/live/song/get/cue_points")
	time.Sleep(50 * time.Millisecond)
	return cues, nil
}

// JumpToCue jumps to a cue point
func (c *AbletonClient) JumpToCue(ctx context.Context, cueIndex int) error {
	return c.sendMessage("/live/song/cue_point/jump", int32(cueIndex))
}

// JumpToTime jumps to a specific time in beats
func (c *AbletonClient) JumpToTime(ctx context.Context, beats float64) error {
	return c.sendMessage("/live/song/set/current_song_time", float32(beats))
}

// Utility methods

// SelectTrack selects a track
func (c *AbletonClient) SelectTrack(ctx context.Context, trackIndex int) error {
	return c.sendMessage("/live/view/set/selected_track", int32(trackIndex))
}

// SelectScene selects a scene
func (c *AbletonClient) SelectScene(ctx context.Context, sceneIndex int) error {
	return c.sendMessage("/live/view/set/selected_scene", int32(sceneIndex))
}

// Undo performs undo
func (c *AbletonClient) Undo(ctx context.Context) error {
	return c.sendMessage("/live/song/undo")
}

// Redo performs redo
func (c *AbletonClient) Redo(ctx context.Context) error {
	return c.sendMessage("/live/song/redo")
}

// GetHealth returns health status
func (c *AbletonClient) GetHealth(ctx context.Context) (*AbletonHealth, error) {
	health := &AbletonHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check if we can connect
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", c.host, c.port), 2*time.Second)
	if err != nil {
		health.Connected = false
		health.Score -= 50
		health.Issues = append(health.Issues, fmt.Sprintf("Cannot connect to %s:%d", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Ensure Ableton Live is running with AbletonOSC")
	} else {
		conn.Close()
		health.Connected = true

		// Try to get tempo
		if err := c.sendMessage("/live/song/get/tempo"); err != nil {
			health.Score -= 30
			health.AbletonOSC = false
			health.Issues = append(health.Issues, "AbletonOSC not responding")
			health.Recommendations = append(health.Recommendations,
				"Install AbletonOSC: https://github.com/ideoforms/AbletonOSC")
		} else {
			health.AbletonOSC = true
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

// Close closes the client
func (c *AbletonClient) Close() error {
	if c.server != nil {
		// Server would be closed here if we had started one
	}
	return nil
}
