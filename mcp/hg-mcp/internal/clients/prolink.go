package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.evanpurkhiser.com/prolink"
)

// ProlinkDevice represents a device on the Pro DJ Link network
type ProlinkDevice struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	IP         string `json:"ip"`
	MAC        string `json:"mac"`
	LastActive string `json:"last_active"`
}

// ProlinkStatus represents the current status of a CDJ/XDJ deck
type ProlinkStatus struct {
	PlayerID       int     `json:"player_id"`
	PlayState      string  `json:"play_state"`
	TrackID        uint32  `json:"track_id,omitempty"`
	TrackDevice    int     `json:"track_device,omitempty"`
	TrackSlot      string  `json:"track_slot,omitempty"`
	TrackType      string  `json:"track_type,omitempty"`
	BPM            float32 `json:"bpm"`
	EffectiveBPM   float32 `json:"effective_bpm"`
	EffectivePitch float32 `json:"effective_pitch"`
	SliderPitch    float32 `json:"slider_pitch"`
	BeatInMeasure  uint8   `json:"beat_in_measure"`
	BeatsUntilCue  uint16  `json:"beats_until_cue"`
	Beat           uint32  `json:"beat"`
	PacketNum      uint32  `json:"packet_num"`
	IsOnAir        bool    `json:"is_on_air"`
	IsMaster       bool    `json:"is_master"`
	IsSync         bool    `json:"is_sync"`
	IsPlaying      bool    `json:"is_playing"`
	MsPerBeat      float64 `json:"ms_per_beat"`
}

// ProlinkTrack represents track metadata from RemoteDB
type ProlinkTrack struct {
	ID         uint32  `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Album      string  `json:"album"`
	Genre      string  `json:"genre"`
	Label      string  `json:"label"`
	Key        string  `json:"key"`
	Length     float64 `json:"length_seconds"`
	Comment    string  `json:"comment,omitempty"`
	Path       string  `json:"path,omitempty"`
	DateAdded  string  `json:"date_added,omitempty"`
	HasArtwork bool    `json:"has_artwork"`
}

// ProlinkNowPlaying combines status and track info for a deck
type ProlinkNowPlaying struct {
	PlayerID int            `json:"player_id"`
	Status   *ProlinkStatus `json:"status"`
	Track    *ProlinkTrack  `json:"track,omitempty"`
}

// ProlinkFullData contains all available prolink data for comprehensive output
type ProlinkFullData struct {
	Timestamp   string                        `json:"timestamp"`
	Connected   bool                          `json:"connected"`
	VirtualCDJ  int                           `json:"virtual_cdj_id"`
	Devices     []ProlinkDevice               `json:"devices"`
	Players     map[string]*ProlinkNowPlaying `json:"players"`
	MasterID    int                           `json:"master_player_id"`
	NetworkInfo *ProlinkNetworkInfo           `json:"network_info,omitempty"`
}

// ProlinkNetworkInfo contains network-level information
type ProlinkNetworkInfo struct {
	Interface     string `json:"interface,omitempty"`
	LocalIP       string `json:"local_ip,omitempty"`
	BroadcastAddr string `json:"broadcast_addr,omitempty"`
}

// ProlinkClient manages connections to the Pro DJ Link network
type ProlinkClient struct {
	network *prolink.Network
	dm      *prolink.DeviceManager
	st      *prolink.CDJStatusMonitor
	rdb     *prolink.RemoteDB

	mu         sync.RWMutex
	connected  bool
	lastStatus map[prolink.DeviceID]*prolink.CDJStatus
	lastTracks map[prolink.DeviceID]*prolink.Track

	// Configuration
	virtualCDJID prolink.DeviceID
	timeout      time.Duration
}

var (
	prolinkClient  *ProlinkClient
	prolinkOnce    sync.Once
	prolinkInitErr error
)

// GetProlinkClient returns the singleton prolink client instance
func GetProlinkClient() (*ProlinkClient, error) {
	prolinkOnce.Do(func() {
		prolinkClient, prolinkInitErr = newProlinkClient()
	})
	return prolinkClient, prolinkInitErr
}

func newProlinkClient() (*ProlinkClient, error) {
	client := &ProlinkClient{
		lastStatus: make(map[prolink.DeviceID]*prolink.CDJStatus),
		lastTracks: make(map[prolink.DeviceID]*prolink.Track),
		timeout:    10 * time.Second,
	}
	return client, nil
}

// Connect establishes a connection to the Pro DJ Link network
func (c *ProlinkClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Connect to the network
	network, err := prolink.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to Pro DJ Link network: %w", err)
	}

	// Auto-configure interface and virtual CDJ ID
	if err := network.AutoConfigure(c.timeout); err != nil {
		return fmt.Errorf("failed to auto-configure: %w", err)
	}

	c.network = network
	c.dm = network.DeviceManager()
	c.st = network.CDJStatusMonitor()
	c.rdb = network.RemoteDB()

	// Start status monitoring using AddStatusHandler
	c.st.AddStatusHandler(prolink.StatusHandlerFunc(func(status *prolink.CDJStatus) {
		c.mu.Lock()
		c.lastStatus[status.PlayerID] = status
		c.mu.Unlock()

		// Fetch track metadata if playing
		if status.PlayState == prolink.PlayStatePlaying {
			trackKey := status.TrackKey()
			if trackKey != nil {
				if track, err := c.rdb.GetTrack(trackKey); err == nil {
					c.mu.Lock()
					c.lastTracks[status.PlayerID] = track
					c.mu.Unlock()
				}
			}
		}
	}))

	c.connected = true
	return nil
}

// Disconnect closes the connection to the Pro DJ Link network
func (c *ProlinkClient) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// Clear state
	c.lastStatus = make(map[prolink.DeviceID]*prolink.CDJStatus)
	c.lastTracks = make(map[prolink.DeviceID]*prolink.Track)
	c.connected = false
	c.network = nil
	c.dm = nil
	c.st = nil
	c.rdb = nil

	return nil
}

// IsConnected returns whether the client is connected
func (c *ProlinkClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetDevices returns all devices on the Pro DJ Link network
func (c *ProlinkClient) GetDevices(ctx context.Context) ([]ProlinkDevice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pro DJ Link network")
	}

	devices := c.dm.ActiveDevices()
	result := make([]ProlinkDevice, 0, len(devices))

	for _, dev := range devices {
		deviceType := dev.Type.String()

		result = append(result, ProlinkDevice{
			ID:         int(dev.ID),
			Name:       dev.Name,
			Type:       deviceType,
			IP:         dev.IP.String(),
			MAC:        dev.MacAddr.String(),
			LastActive: dev.LastActive.Format(time.RFC3339),
		})
	}

	return result, nil
}

// GetStatus returns the current status of a specific player or all players
func (c *ProlinkClient) GetStatus(ctx context.Context, playerID int) ([]ProlinkStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pro DJ Link network")
	}

	result := make([]ProlinkStatus, 0)

	for id, status := range c.lastStatus {
		if playerID != 0 && int(id) != playerID {
			continue
		}

		result = append(result, c.buildStatus(status))
	}

	return result, nil
}

// buildStatus creates a ProlinkStatus from raw prolink status
func (c *ProlinkClient) buildStatus(status *prolink.CDJStatus) ProlinkStatus {
	playState := status.PlayState.String()

	effectiveBPM := status.TrackBPM
	if status.SliderPitch != 0 {
		effectiveBPM = status.TrackBPM * (1 + status.SliderPitch/100)
	}

	msPerBeat := float64(0)
	if effectiveBPM > 0 {
		msPerBeat = 60000.0 / float64(effectiveBPM)
	}

	trackSlot := status.TrackSlot.String()
	trackType := status.TrackType.String()

	return ProlinkStatus{
		PlayerID:       int(status.PlayerID),
		PlayState:      playState,
		TrackID:        status.TrackID,
		TrackDevice:    int(status.TrackDevice),
		TrackSlot:      trackSlot,
		TrackType:      trackType,
		BPM:            status.TrackBPM,
		EffectiveBPM:   effectiveBPM,
		EffectivePitch: status.EffectivePitch,
		SliderPitch:    status.SliderPitch,
		BeatInMeasure:  status.BeatInMeasure,
		BeatsUntilCue:  status.BeatsUntilCue,
		Beat:           status.Beat,
		PacketNum:      status.PacketNum,
		IsOnAir:        status.IsOnAir,
		IsMaster:       status.IsMaster,
		IsSync:         status.IsSync,
		IsPlaying:      status.PlayState == prolink.PlayStatePlaying,
		MsPerBeat:      msPerBeat,
	}
}

// GetNowPlaying returns the current track playing on each deck
func (c *ProlinkClient) GetNowPlaying(ctx context.Context, playerID int) ([]ProlinkNowPlaying, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pro DJ Link network")
	}

	result := make([]ProlinkNowPlaying, 0)

	for id, status := range c.lastStatus {
		if playerID != 0 && int(id) != playerID {
			continue
		}

		builtStatus := c.buildStatus(status)
		np := ProlinkNowPlaying{
			PlayerID: int(id),
			Status:   &builtStatus,
		}

		// Add track info if available
		if track, ok := c.lastTracks[id]; ok && track != nil {
			np.Track = c.buildTrack(track)
		}

		result = append(result, np)
	}

	return result, nil
}

// buildTrack creates a ProlinkTrack from raw prolink track
func (c *ProlinkClient) buildTrack(track *prolink.Track) *ProlinkTrack {
	if track == nil {
		return nil
	}

	dateAdded := ""
	if !track.DateAdded.IsZero() {
		dateAdded = track.DateAdded.Format(time.RFC3339)
	}

	return &ProlinkTrack{
		ID:         track.ID,
		Title:      track.Title,
		Artist:     track.Artist,
		Album:      track.Album,
		Genre:      track.Genre,
		Label:      track.Label,
		Key:        track.Key,
		Length:     track.Length.Seconds(),
		Comment:    track.Comment,
		Path:       track.Path,
		DateAdded:  dateAdded,
		HasArtwork: len(track.Artwork) > 0,
	}
}

// GetTrack retrieves track metadata from a device
func (c *ProlinkClient) GetTrack(ctx context.Context, playerID int, trackID uint32) (*ProlinkTrack, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to Pro DJ Link network")
	}

	// Get the status to find the track slot and type
	status, ok := c.lastStatus[prolink.DeviceID(playerID)]
	if !ok {
		return nil, fmt.Errorf("player %d not found", playerID)
	}

	trackKey := status.TrackKey()
	if trackKey == nil {
		return nil, fmt.Errorf("no track loaded on player %d", playerID)
	}

	// Override track ID if specified
	if trackID != 0 {
		trackKey.TrackID = trackID
	}

	track, err := c.rdb.GetTrack(trackKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get track: %w", err)
	}

	return c.buildTrack(track), nil
}

// GetMasterPlayer returns the current master player ID
func (c *ProlinkClient) GetMasterPlayer(ctx context.Context) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected {
		return 0, fmt.Errorf("not connected to Pro DJ Link network")
	}

	for id, status := range c.lastStatus {
		if status.IsMaster {
			return int(id), nil
		}
	}

	return 0, fmt.Errorf("no master player found")
}

// GetFullData returns all available prolink data in a single comprehensive structure
func (c *ProlinkClient) GetFullData(ctx context.Context) (*ProlinkFullData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data := &ProlinkFullData{
		Timestamp: time.Now().Format(time.RFC3339Nano),
		Connected: c.connected,
		Players:   make(map[string]*ProlinkNowPlaying),
	}

	if !c.connected {
		return data, nil
	}

	// Get devices
	devices := c.dm.ActiveDevices()
	data.Devices = make([]ProlinkDevice, 0, len(devices))
	for _, dev := range devices {
		data.Devices = append(data.Devices, ProlinkDevice{
			ID:         int(dev.ID),
			Name:       dev.Name,
			Type:       dev.Type.String(),
			IP:         dev.IP.String(),
			MAC:        dev.MacAddr.String(),
			LastActive: dev.LastActive.Format(time.RFC3339),
		})
	}

	// Get all player status and track info
	for id, status := range c.lastStatus {
		builtStatus := c.buildStatus(status)
		np := &ProlinkNowPlaying{
			PlayerID: int(id),
			Status:   &builtStatus,
		}

		if track, ok := c.lastTracks[id]; ok && track != nil {
			np.Track = c.buildTrack(track)
		}

		data.Players[fmt.Sprintf("player_%d", id)] = np

		// Check for master
		if status.IsMaster {
			data.MasterID = int(id)
		}
	}

	return data, nil
}

// ProlinkDevicesToJSON converts a slice of devices to JSON
func ProlinkDevicesToJSON(d []ProlinkDevice) (string, error) {
	data, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// String returns a human-readable status
func (s *ProlinkStatus) String() string {
	return fmt.Sprintf("Player %d: %s (%.1f BPM, beat %d/%d)",
		s.PlayerID, s.PlayState, s.EffectiveBPM, s.BeatInMeasure, s.Beat)
}

// String returns a human-readable now playing
func (np *ProlinkNowPlaying) String() string {
	if np.Track != nil {
		return fmt.Sprintf("Player %d: %s - %s (%.1f BPM)",
			np.PlayerID, np.Track.Artist, np.Track.Title, np.Status.EffectiveBPM)
	}
	return fmt.Sprintf("Player %d: %s (%.1f BPM)",
		np.PlayerID, np.Status.PlayState, np.Status.EffectiveBPM)
}
