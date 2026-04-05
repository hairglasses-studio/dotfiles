// Package bridge provides integration between different systems.
// prolink_showkontrol.go bridges XDJ/CDJ Pro DJ Link data to Showkontrol for
// beat-synchronized lighting and cue triggering.
package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

// BridgeMode defines how the bridge syncs to the decks
type BridgeMode string

const (
	// BridgeModeMaster follows only the master deck
	BridgeModeMaster BridgeMode = "master"
	// BridgeModeOnAir follows the deck that's on air (via DJM mixer)
	BridgeModeOnAir BridgeMode = "on_air"
	// BridgeModeAll tracks all decks
	BridgeModeAll BridgeMode = "all"
)

// BridgeConfig configures the prolink-showkontrol bridge
type BridgeConfig struct {
	Mode            BridgeMode        `json:"mode"`
	BeatCue         string            `json:"beat_cue,omitempty"`          // Cue to fire on beat 1
	TrackChangeCue  string            `json:"track_change_cue,omitempty"`  // Cue to fire on track change
	AutoFireOnDrop  bool              `json:"auto_fire_on_drop,omitempty"` // Fire cue on detected drop
	GenreCueMapping map[string]string `json:"genre_cue_mapping,omitempty"` // Genre -> cue ID mapping
	KeyColorMapping bool              `json:"key_color_mapping,omitempty"` // Map musical key to color
	PollIntervalMs  int               `json:"poll_interval_ms,omitempty"`  // Polling interval in ms (default 200)
}

// BridgeStatus represents the current state of the bridge
type BridgeStatus struct {
	Running        bool          `json:"running"`
	Mode           BridgeMode    `json:"mode"`
	ActivePlayerID int           `json:"active_player_id"`
	CurrentBPM     float32       `json:"current_bpm"`
	CurrentBeat    uint8         `json:"current_beat"`
	CurrentKey     string        `json:"current_key,omitempty"`
	CurrentGenre   string        `json:"current_genre,omitempty"`
	CurrentTrack   string        `json:"current_track,omitempty"`
	LastCueFired   string        `json:"last_cue_fired,omitempty"`
	LastCueTime    time.Time     `json:"last_cue_time,omitempty"`
	BeatsSynced    uint64        `json:"beats_synced"`
	TrackChanges   uint64        `json:"track_changes"`
	Errors         []string      `json:"errors,omitempty"`
	StartedAt      time.Time     `json:"started_at,omitempty"`
	Config         *BridgeConfig `json:"config"`
}

// BeatEvent represents a beat event for subscribers
type BeatEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	PlayerID      int       `json:"player_id"`
	BPM           float32   `json:"bpm"`
	EffectiveBPM  float32   `json:"effective_bpm"`
	BeatInMeasure uint8     `json:"beat_in_measure"`
	TotalBeat     uint32    `json:"total_beat"`
	MsPerBeat     float64   `json:"ms_per_beat"`
	IsDownbeat    bool      `json:"is_downbeat"` // Beat 1 of measure
}

// TrackChangeEvent represents a track change event
type TrackChangeEvent struct {
	Timestamp     time.Time             `json:"timestamp"`
	PlayerID      int                   `json:"player_id"`
	PreviousTrack *clients.ProlinkTrack `json:"previous_track,omitempty"`
	NewTrack      *clients.ProlinkTrack `json:"new_track"`
}

// ProlinkShowkontrolBridge bridges Pro DJ Link data to Showkontrol
type ProlinkShowkontrolBridge struct {
	prolinkClient     *clients.ProlinkClient
	showkontrolClient *clients.ShowkontrolClient

	config *BridgeConfig
	status *BridgeStatus

	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}

	// Callbacks for external subscribers
	onBeat        func(event BeatEvent)
	onTrackChange func(event TrackChangeEvent)

	// Track state for change detection
	lastTrackIDs   map[int]uint32
	lastBeatCounts map[int]uint32
}

var (
	bridgeInstance *ProlinkShowkontrolBridge
	bridgeOnce     sync.Once
)

// GetBridge returns the singleton bridge instance
func GetBridge() (*ProlinkShowkontrolBridge, error) {
	var initErr error
	bridgeOnce.Do(func() {
		bridgeInstance, initErr = newBridge()
	})
	return bridgeInstance, initErr
}

func newBridge() (*ProlinkShowkontrolBridge, error) {
	prolinkClient, err := clients.GetProlinkClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get prolink client: %w", err)
	}

	showkontrolClient, err := clients.NewShowkontrolClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create showkontrol client: %w", err)
	}

	return &ProlinkShowkontrolBridge{
		prolinkClient:     prolinkClient,
		showkontrolClient: showkontrolClient,
		config: &BridgeConfig{
			Mode:           BridgeModeMaster,
			PollIntervalMs: 200,
		},
		status: &BridgeStatus{
			Running: false,
			Mode:    BridgeModeMaster,
		},
		lastTrackIDs:   make(map[int]uint32),
		lastBeatCounts: make(map[int]uint32),
	}, nil
}

// Configure updates the bridge configuration
func (b *ProlinkShowkontrolBridge) Configure(config *BridgeConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.config = config
	b.status.Config = config
	b.status.Mode = config.Mode
}

// GetConfig returns the current configuration
func (b *ProlinkShowkontrolBridge) GetConfig() *BridgeConfig {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.config
}

// SetOnBeat sets the callback for beat events
func (b *ProlinkShowkontrolBridge) SetOnBeat(callback func(event BeatEvent)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onBeat = callback
}

// SetOnTrackChange sets the callback for track change events
func (b *ProlinkShowkontrolBridge) SetOnTrackChange(callback func(event TrackChangeEvent)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onTrackChange = callback
}

// Start begins the bridge synchronization
func (b *ProlinkShowkontrolBridge) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("bridge already running")
	}

	// Ensure prolink is connected
	if !b.prolinkClient.IsConnected() {
		if err := b.prolinkClient.Connect(ctx); err != nil {
			b.mu.Unlock()
			return fmt.Errorf("failed to connect to Pro DJ Link: %w", err)
		}
	}

	b.running = true
	b.stopCh = make(chan struct{})
	b.status.Running = true
	b.status.StartedAt = time.Now()
	b.status.Errors = nil
	b.mu.Unlock()

	// Start the sync loop
	go b.syncLoop(ctx)

	return nil
}

// Stop halts the bridge synchronization
func (b *ProlinkShowkontrolBridge) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return fmt.Errorf("bridge not running")
	}

	close(b.stopCh)
	b.running = false
	b.status.Running = false

	return nil
}

// IsRunning returns whether the bridge is active
func (b *ProlinkShowkontrolBridge) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// GetStatus returns the current bridge status
func (b *ProlinkShowkontrolBridge) GetStatus() *BridgeStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	statusCopy := *b.status
	if b.config != nil {
		configCopy := *b.config
		statusCopy.Config = &configCopy
	}
	return &statusCopy
}

// syncLoop is the main synchronization loop
func (b *ProlinkShowkontrolBridge) syncLoop(ctx context.Context) {
	pollInterval := time.Duration(b.config.PollIntervalMs) * time.Millisecond
	if pollInterval < 50*time.Millisecond {
		pollInterval = 50 * time.Millisecond
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ctx.Done():
			b.Stop()
			return
		case <-ticker.C:
			b.syncTick(ctx)
		}
	}
}

// syncTick processes one sync cycle
func (b *ProlinkShowkontrolBridge) syncTick(ctx context.Context) {
	// Get full prolink data
	fullData, err := b.prolinkClient.GetFullData(ctx)
	if err != nil {
		b.addError(fmt.Sprintf("failed to get prolink data: %v", err))
		return
	}

	if !fullData.Connected {
		return
	}

	// Find the active player based on mode
	activePlayer := b.findActivePlayer(fullData)
	if activePlayer == nil {
		return
	}

	b.mu.Lock()
	b.status.ActivePlayerID = activePlayer.PlayerID
	if activePlayer.Status != nil {
		b.status.CurrentBPM = activePlayer.Status.EffectiveBPM
		b.status.CurrentBeat = activePlayer.Status.BeatInMeasure
	}
	if activePlayer.Track != nil {
		b.status.CurrentTrack = fmt.Sprintf("%s - %s", activePlayer.Track.Artist, activePlayer.Track.Title)
		b.status.CurrentKey = activePlayer.Track.Key
		b.status.CurrentGenre = activePlayer.Track.Genre
	}
	b.mu.Unlock()

	// Process beat events
	if activePlayer.Status != nil {
		b.processBeatEvent(activePlayer)
	}

	// Process track changes
	if activePlayer.Track != nil {
		b.processTrackChange(ctx, activePlayer)
	}
}

// findActivePlayer returns the player to sync to based on mode
func (b *ProlinkShowkontrolBridge) findActivePlayer(data *clients.ProlinkFullData) *clients.ProlinkNowPlaying {
	b.mu.RLock()
	mode := b.config.Mode
	b.mu.RUnlock()

	for _, player := range data.Players {
		if player.Status == nil {
			continue
		}

		switch mode {
		case BridgeModeMaster:
			if player.Status.IsMaster {
				return player
			}
		case BridgeModeOnAir:
			if player.Status.IsOnAir {
				return player
			}
		case BridgeModeAll:
			// Return first playing player
			if player.Status.IsPlaying {
				return player
			}
		}
	}

	// Fallback: return first player with status
	for _, player := range data.Players {
		if player.Status != nil {
			return player
		}
	}

	return nil
}

// processBeatEvent handles beat synchronization
func (b *ProlinkShowkontrolBridge) processBeatEvent(player *clients.ProlinkNowPlaying) {
	if player.Status == nil {
		return
	}

	playerID := player.PlayerID
	currentBeat := player.Status.Beat

	b.mu.Lock()
	lastBeat := b.lastBeatCounts[playerID]
	b.lastBeatCounts[playerID] = currentBeat

	// Detect new beat
	if currentBeat <= lastBeat {
		b.mu.Unlock()
		return
	}

	b.status.BeatsSynced++
	onBeat := b.onBeat
	config := b.config
	b.mu.Unlock()

	// Create beat event
	event := BeatEvent{
		Timestamp:     time.Now(),
		PlayerID:      playerID,
		BPM:           player.Status.BPM,
		EffectiveBPM:  player.Status.EffectiveBPM,
		BeatInMeasure: player.Status.BeatInMeasure,
		TotalBeat:     player.Status.Beat,
		MsPerBeat:     player.Status.MsPerBeat,
		IsDownbeat:    player.Status.BeatInMeasure == 1,
	}

	// Fire callback
	if onBeat != nil {
		onBeat(event)
	}

	// Fire downbeat cue if configured
	if event.IsDownbeat && config.BeatCue != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		if err := b.showkontrolClient.FireCue(ctx, config.BeatCue); err != nil {
			b.addError(fmt.Sprintf("failed to fire beat cue: %v", err))
		} else {
			b.mu.Lock()
			b.status.LastCueFired = config.BeatCue
			b.status.LastCueTime = time.Now()
			b.mu.Unlock()
		}
		cancel()
	}
}

// processTrackChange handles track change detection and cue firing
func (b *ProlinkShowkontrolBridge) processTrackChange(ctx context.Context, player *clients.ProlinkNowPlaying) {
	if player.Track == nil {
		return
	}

	playerID := player.PlayerID
	currentTrackID := player.Track.ID

	b.mu.Lock()
	lastTrackID := b.lastTrackIDs[playerID]

	// No change
	if currentTrackID == lastTrackID {
		b.mu.Unlock()
		return
	}

	// Track changed
	b.lastTrackIDs[playerID] = currentTrackID
	b.status.TrackChanges++

	onTrackChange := b.onTrackChange
	config := b.config
	b.mu.Unlock()

	// Create track change event
	event := TrackChangeEvent{
		Timestamp: time.Now(),
		PlayerID:  playerID,
		NewTrack:  player.Track,
	}

	// Fire callback
	if onTrackChange != nil {
		onTrackChange(event)
	}

	// Fire track change cue if configured
	if config.TrackChangeCue != "" {
		timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		if err := b.showkontrolClient.FireCue(timeoutCtx, config.TrackChangeCue); err != nil {
			b.addError(fmt.Sprintf("failed to fire track change cue: %v", err))
		} else {
			b.mu.Lock()
			b.status.LastCueFired = config.TrackChangeCue
			b.status.LastCueTime = time.Now()
			b.mu.Unlock()
		}
		cancel()
	}

	// Fire genre-specific cue if configured
	if config.GenreCueMapping != nil && player.Track.Genre != "" {
		if cueID, ok := config.GenreCueMapping[player.Track.Genre]; ok {
			timeoutCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			if err := b.showkontrolClient.FireCue(timeoutCtx, cueID); err != nil {
				b.addError(fmt.Sprintf("failed to fire genre cue: %v", err))
			} else {
				b.mu.Lock()
				b.status.LastCueFired = cueID
				b.status.LastCueTime = time.Now()
				b.mu.Unlock()
			}
			cancel()
		}
	}
}

// addError adds an error to the status (keeps last 10)
func (b *ProlinkShowkontrolBridge) addError(err string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.status.Errors = append(b.status.Errors, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), err))
	if len(b.status.Errors) > 10 {
		b.status.Errors = b.status.Errors[len(b.status.Errors)-10:]
	}
}

// GetBridgeStatusJSON returns status as JSON string
func (b *ProlinkShowkontrolBridge) GetBridgeStatusJSON() (string, error) {
	status := b.GetStatus()
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// KeyToColor maps musical keys to colors for lighting
// Based on the Camelot wheel - harmonically related keys get similar colors
func KeyToColor(key string) string {
	// Camelot wheel color mapping
	keyColors := map[string]string{
		// Minor keys (A)
		"Abm": "#FF0000", "Am": "#FF4400", "Bbm": "#FF8800",
		"Bm": "#FFCC00", "Cm": "#FFFF00", "C#m": "#88FF00",
		"Dm": "#00FF00", "Ebm": "#00FF88", "Em": "#00FFFF",
		"Fm": "#0088FF", "F#m": "#0000FF", "Gm": "#8800FF",

		// Major keys (B)
		"Ab": "#FF0088", "A": "#FF0044", "Bb": "#FF4400",
		"B": "#FF8800", "C": "#FFCC00", "C#": "#FFFF00",
		"D": "#88FF00", "Eb": "#00FF00", "E": "#00FF88",
		"F": "#00FFFF", "F#": "#0088FF", "G": "#0000FF",
	}

	if color, ok := keyColors[key]; ok {
		return color
	}
	return "#FFFFFF" // White for unknown keys
}

// BPMToTempo categorizes BPM into tempo categories for scene selection
func BPMToTempo(bpm float32) string {
	switch {
	case bpm < 90:
		return "slow"
	case bpm < 120:
		return "medium"
	case bpm < 140:
		return "fast"
	case bpm < 160:
		return "very_fast"
	default:
		return "extreme"
	}
}
