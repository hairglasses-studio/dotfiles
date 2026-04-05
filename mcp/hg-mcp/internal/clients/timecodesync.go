// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TimecodeSyncClient provides unified timecode synchronization across systems
type TimecodeSyncClient struct {
	mu            sync.RWMutex
	masterSource  string
	linkedSystems map[string]bool
	currentTC     Timecode
	lastSync      time.Time
	format        TimecodeFormat

	// Clients
	grandma3Client    *GrandMA3Client
	showkontrolClient *ShowkontrolClient
	abletonClient     *AbletonClient
}

// Timecode represents a timecode position
type Timecode struct {
	Hours        int     `json:"hours"`
	Minutes      int     `json:"minutes"`
	Seconds      int     `json:"seconds"`
	Frames       int     `json:"frames"`
	FrameRate    float64 `json:"frame_rate"`
	DropFrame    bool    `json:"drop_frame"`
	TotalFrames  int64   `json:"total_frames"`
	TotalSeconds float64 `json:"total_seconds"`
}

// TimecodeFormat represents the timecode format
type TimecodeFormat string

const (
	TimecodeFormatSMPTE TimecodeFormat = "smpte" // SMPTE timecode
	TimecodeFormatMTC   TimecodeFormat = "mtc"   // MIDI Time Code
	TimecodeFormatLTC   TimecodeFormat = "ltc"   // Linear Time Code (audio)
	TimecodeFormatMS    TimecodeFormat = "ms"    // Milliseconds
)

// TimecodeSystem represents a system capable of timecode sync
type TimecodeSystem struct {
	Name             string           `json:"name"`
	Type             string           `json:"type"` // master, slave, bidirectional
	Connected        bool             `json:"connected"`
	CurrentTC        *Timecode        `json:"current_tc,omitempty"`
	SupportedFormats []TimecodeFormat `json:"supported_formats"`
	CanRead          bool             `json:"can_read"`
	CanWrite         bool             `json:"can_write"`
	Linked           bool             `json:"linked"`
	Offset           int64            `json:"offset_frames"` // Frame offset from master
}

// TimecodeSyncStatus represents the current sync state
type TimecodeSyncStatus struct {
	MasterSource string           `json:"master_source"`
	CurrentTC    Timecode         `json:"current_tc"`
	Format       TimecodeFormat   `json:"format"`
	LastSync     time.Time        `json:"last_sync"`
	Systems      []TimecodeSystem `json:"systems"`
	InSync       bool             `json:"in_sync"`
	DriftFrames  int64            `json:"drift_frames,omitempty"`
	IsRunning    bool             `json:"is_running"`
}

// TimecodeSyncHealth represents health status
type TimecodeSyncHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	MasterConnected bool     `json:"master_connected"`
	LinkedCount     int      `json:"linked_count"`
	ConnectedCount  int      `json:"connected_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewTimecodeSyncClient creates a new timecode sync client
func NewTimecodeSyncClient() (*TimecodeSyncClient, error) {
	return &TimecodeSyncClient{
		masterSource:  "showkontrol",
		linkedSystems: make(map[string]bool),
		format:        TimecodeFormatSMPTE,
		currentTC: Timecode{
			FrameRate: 30,
		},
	}, nil
}

// getGrandMA3Client lazily initializes the grandMA3 client
func (c *TimecodeSyncClient) getGrandMA3Client() (*GrandMA3Client, error) {
	if c.grandma3Client == nil {
		client, err := NewGrandMA3Client()
		if err != nil {
			return nil, err
		}
		c.grandma3Client = client
	}
	return c.grandma3Client, nil
}

// getShowkontrolClient lazily initializes the Showkontrol client
func (c *TimecodeSyncClient) getShowkontrolClient() (*ShowkontrolClient, error) {
	if c.showkontrolClient == nil {
		client, err := NewShowkontrolClient()
		if err != nil {
			return nil, err
		}
		c.showkontrolClient = client
	}
	return c.showkontrolClient, nil
}

// getAbletonClient lazily initializes the Ableton client
func (c *TimecodeSyncClient) getAbletonClient() (*AbletonClient, error) {
	if c.abletonClient == nil {
		client, err := NewAbletonClient()
		if err != nil {
			return nil, err
		}
		c.abletonClient = client
	}
	return c.abletonClient, nil
}

// GetStatus returns the current timecode sync status across all systems
func (c *TimecodeSyncClient) GetStatus(ctx context.Context) (*TimecodeSyncStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &TimecodeSyncStatus{
		MasterSource: c.masterSource,
		CurrentTC:    c.currentTC,
		Format:       c.format,
		LastSync:     c.lastSync,
		Systems:      make([]TimecodeSystem, 0),
		InSync:       true,
	}

	// Check Showkontrol (primary timecode source)
	sys := TimecodeSystem{
		Name:             "showkontrol",
		Type:             "master",
		SupportedFormats: []TimecodeFormat{TimecodeFormatSMPTE, TimecodeFormatMTC, TimecodeFormatLTC},
		CanRead:          true,
		CanWrite:         true,
		Linked:           c.linkedSystems["showkontrol"],
	}
	if showkontrol, err := c.getShowkontrolClient(); err == nil {
		if showkontrolStatus, err := showkontrol.GetStatus(ctx); err == nil {
			sys.Connected = showkontrolStatus.Connected
			if showkontrolStatus.Timecode != nil {
				// Convert TimecodeStatus to Timecode
				tc := msToTimecode(showkontrolStatus.Timecode.Position*1000, showkontrolStatus.Timecode.FrameRate)
				sys.CurrentTC = tc
			}
		}
	}
	status.Systems = append(status.Systems, sys)

	// Check grandMA3
	sys = TimecodeSystem{
		Name:             "grandma3",
		Type:             "slave",
		SupportedFormats: []TimecodeFormat{TimecodeFormatSMPTE, TimecodeFormatMTC},
		CanRead:          true,
		CanWrite:         true,
		Linked:           c.linkedSystems["grandma3"],
	}
	if grandma3, err := c.getGrandMA3Client(); err == nil {
		if grandma3Status, err := grandma3.GetStatus(ctx); err == nil {
			sys.Connected = grandma3Status.Connected
		}
	}
	status.Systems = append(status.Systems, sys)

	// Check Ableton (arrangement position as timecode)
	sys = TimecodeSystem{
		Name:             "ableton",
		Type:             "bidirectional",
		SupportedFormats: []TimecodeFormat{TimecodeFormatMS},
		CanRead:          true,
		CanWrite:         true,
		Linked:           c.linkedSystems["ableton"],
	}
	if ableton, err := c.getAbletonClient(); err == nil {
		if abletonStatus, err := ableton.GetStatus(ctx); err == nil {
			sys.Connected = abletonStatus.Connected
			if abletonStatus.State != nil {
				// Convert SongTime (seconds) to timecode
				positionMs := abletonStatus.State.SongTime * 1000
				sys.CurrentTC = msToTimecode(positionMs, c.currentTC.FrameRate)
			}
		}
	}
	status.Systems = append(status.Systems, sys)

	// Calculate drift and sync status
	var maxDrift int64
	for _, sys := range status.Systems {
		if sys.Connected && sys.CurrentTC != nil {
			drift := abs64(sys.CurrentTC.TotalFrames - c.currentTC.TotalFrames)
			if drift > maxDrift {
				maxDrift = drift
			}
			if drift > 2 { // More than 2 frames drift
				status.InSync = false
			}
		}
	}
	status.DriftFrames = maxDrift

	return status, nil
}

// SetMaster sets the master timecode source
func (c *TimecodeSyncClient) SetMaster(ctx context.Context, source string) error {
	validSources := map[string]bool{
		"showkontrol": true,
		"grandma3":    true,
		"ableton":     true,
		"manual":      true,
	}

	if !validSources[source] {
		return fmt.Errorf("invalid master source: %s (valid: showkontrol, grandma3, ableton, manual)", source)
	}

	c.mu.Lock()
	c.masterSource = source
	c.mu.Unlock()

	// Read current timecode from new master
	if source != "manual" {
		tc, err := c.readTimecodeFromSource(ctx, source)
		if err != nil {
			return fmt.Errorf("failed to read timecode from new master: %w", err)
		}
		c.mu.Lock()
		c.currentTC = *tc
		c.mu.Unlock()
	}

	return nil
}

// readTimecodeFromSource reads timecode from a specific source
func (c *TimecodeSyncClient) readTimecodeFromSource(ctx context.Context, source string) (*Timecode, error) {
	switch source {
	case "showkontrol":
		if showkontrol, err := c.getShowkontrolClient(); err == nil {
			tcStatus, err := showkontrol.GetTimecodeStatus(ctx)
			if err != nil {
				return nil, err
			}
			// Parse the timecode string from Showkontrol
			if tcStatus.PositionTC != "" {
				return ParseTimecode(tcStatus.PositionTC, tcStatus.FrameRate, false)
			}
			// Fallback to converting seconds
			return msToTimecode(tcStatus.Position*1000, tcStatus.FrameRate), nil
		}
		return nil, fmt.Errorf("showkontrol client not available")

	case "grandma3":
		// grandMA3 typically receives timecode, doesn't generate it
		return &c.currentTC, nil

	case "ableton":
		if ableton, err := c.getAbletonClient(); err == nil {
			status, err := ableton.GetStatus(ctx)
			if err != nil {
				return nil, err
			}
			if status.State != nil {
				positionMs := status.State.SongTime * 1000
				tc := msToTimecode(positionMs, c.currentTC.FrameRate)
				return tc, nil
			}
		}
		return nil, fmt.Errorf("ableton client not available")

	default:
		return &c.currentTC, nil
	}
}

// LinkSystem links a system to receive timecode from master
func (c *TimecodeSyncClient) LinkSystem(ctx context.Context, system string) error {
	validSystems := map[string]bool{
		"showkontrol": true,
		"grandma3":    true,
		"ableton":     true,
	}

	if !validSystems[system] {
		return fmt.Errorf("invalid system: %s", system)
	}

	c.mu.Lock()
	c.linkedSystems[system] = true
	c.mu.Unlock()

	return nil
}

// UnlinkSystem removes a system from timecode sync
func (c *TimecodeSyncClient) UnlinkSystem(ctx context.Context, system string) error {
	c.mu.Lock()
	delete(c.linkedSystems, system)
	c.mu.Unlock()
	return nil
}

// GotoTimecode jumps all linked systems to a specific timecode position
func (c *TimecodeSyncClient) GotoTimecode(ctx context.Context, tc *Timecode) error {
	c.mu.Lock()
	c.currentTC = *tc
	c.lastSync = time.Now()
	linkedSystems := make(map[string]bool)
	for k, v := range c.linkedSystems {
		linkedSystems[k] = v
	}
	c.mu.Unlock()

	var errs []error
	for system, linked := range linkedSystems {
		if linked {
			if err := c.pushTimecodeToSystem(ctx, system, tc); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", system, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some systems failed to sync: %v", errs)
	}

	return nil
}

// GotoPosition jumps to a position specified in HH:MM:SS:FF format
func (c *TimecodeSyncClient) GotoPosition(ctx context.Context, position string) error {
	tc, err := ParseTimecode(position, c.currentTC.FrameRate, c.currentTC.DropFrame)
	if err != nil {
		return err
	}
	return c.GotoTimecode(ctx, tc)
}

// pushTimecodeToSystem pushes timecode to a specific system
func (c *TimecodeSyncClient) pushTimecodeToSystem(ctx context.Context, system string, tc *Timecode) error {
	switch system {
	case "showkontrol":
		if showkontrol, err := c.getShowkontrolClient(); err == nil {
			// Showkontrol takes timecode as string
			return showkontrol.GotoTimecode(ctx, tc.String())
		}
		return fmt.Errorf("showkontrol client not available")

	case "grandma3":
		if grandma3, err := c.getGrandMA3Client(); err == nil {
			// grandMA3 uses command to go to timecode position
			cmd := fmt.Sprintf("Go Timecode %02d:%02d:%02d:%02d", tc.Hours, tc.Minutes, tc.Seconds, tc.Frames)
			return grandma3.SendCommand(ctx, cmd)
		}
		return fmt.Errorf("grandma3 client not available")

	case "ableton":
		if ableton, err := c.getAbletonClient(); err == nil {
			// Convert timecode to beats for Ableton (assuming 120 BPM default)
			// JumpToTime takes beats, so convert seconds to beats
			beats := tc.TotalSeconds * 2.0 // Approximate: 2 beats per second at 120 BPM
			return ableton.JumpToTime(ctx, beats)
		}
		return fmt.Errorf("ableton client not available")

	default:
		return fmt.Errorf("unknown system: %s", system)
	}
}

// SyncFromMaster reads timecode from master and pushes to all linked systems
func (c *TimecodeSyncClient) SyncFromMaster(ctx context.Context) error {
	c.mu.RLock()
	master := c.masterSource
	c.mu.RUnlock()

	tc, err := c.readTimecodeFromSource(ctx, master)
	if err != nil {
		return fmt.Errorf("failed to read from master: %w", err)
	}

	return c.GotoTimecode(ctx, tc)
}

// SetFormat sets the timecode format
func (c *TimecodeSyncClient) SetFormat(ctx context.Context, format TimecodeFormat, frameRate float64, dropFrame bool) error {
	validFormats := map[TimecodeFormat]bool{
		TimecodeFormatSMPTE: true,
		TimecodeFormatMTC:   true,
		TimecodeFormatLTC:   true,
		TimecodeFormatMS:    true,
	}

	if !validFormats[format] {
		return fmt.Errorf("invalid format: %s", format)
	}

	// Validate frame rate
	validRates := map[float64]bool{23.976: true, 24: true, 25: true, 29.97: true, 30: true, 50: true, 59.94: true, 60: true}
	if !validRates[frameRate] {
		return fmt.Errorf("invalid frame rate: %f", frameRate)
	}

	c.mu.Lock()
	c.format = format
	c.currentTC.FrameRate = frameRate
	c.currentTC.DropFrame = dropFrame
	c.mu.Unlock()

	return nil
}

// GetHealth returns health status
func (c *TimecodeSyncClient) GetHealth(ctx context.Context) (*TimecodeSyncHealth, error) {
	health := &TimecodeSyncHealth{
		Score:  100,
		Status: "healthy",
	}

	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Score -= 50
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get status: %v", err))
	} else {
		// Check master connection
		for _, sys := range status.Systems {
			if sys.Name == status.MasterSource {
				health.MasterConnected = sys.Connected
				if !sys.Connected {
					health.Score -= 30
					health.Issues = append(health.Issues, fmt.Sprintf("Master source '%s' not connected", status.MasterSource))
					health.Recommendations = append(health.Recommendations, fmt.Sprintf("Ensure %s is running and accessible", status.MasterSource))
				}
			}
			if sys.Connected {
				health.ConnectedCount++
			}
			if sys.Linked {
				health.LinkedCount++
			}
		}

		// Check sync status
		if !status.InSync {
			health.Score -= 20
			health.Issues = append(health.Issues, fmt.Sprintf("Systems out of sync (drift: %d frames)", status.DriftFrames))
			health.Recommendations = append(health.Recommendations, "Run aftrs_timecode_sync to resync all systems")
		}

		// Check if any systems are linked
		if health.LinkedCount == 0 {
			health.Score -= 10
			health.Issues = append(health.Issues, "No systems linked for sync")
			health.Recommendations = append(health.Recommendations, "Link systems using aftrs_timecode_link")
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

// Helper functions

// ParseTimecode parses a timecode string in HH:MM:SS:FF format
func ParseTimecode(s string, frameRate float64, dropFrame bool) (*Timecode, error) {
	var h, m, sec, f int
	n, err := fmt.Sscanf(s, "%d:%d:%d:%d", &h, &m, &sec, &f)
	if err != nil || n != 4 {
		// Try alternative formats
		n, err = fmt.Sscanf(s, "%d:%d:%d;%d", &h, &m, &sec, &f)
		if err != nil || n != 4 {
			return nil, fmt.Errorf("invalid timecode format: %s (expected HH:MM:SS:FF)", s)
		}
		dropFrame = true
	}

	if h < 0 || h > 23 {
		return nil, fmt.Errorf("hours must be 0-23")
	}
	if m < 0 || m > 59 {
		return nil, fmt.Errorf("minutes must be 0-59")
	}
	if sec < 0 || sec > 59 {
		return nil, fmt.Errorf("seconds must be 0-59")
	}
	maxFrames := int(frameRate)
	if f < 0 || f >= maxFrames {
		return nil, fmt.Errorf("frames must be 0-%d for %.2f fps", maxFrames-1, frameRate)
	}

	totalFrames := int64(h)*3600*int64(frameRate) + int64(m)*60*int64(frameRate) + int64(sec)*int64(frameRate) + int64(f)
	totalSeconds := float64(totalFrames) / frameRate

	return &Timecode{
		Hours:        h,
		Minutes:      m,
		Seconds:      sec,
		Frames:       f,
		FrameRate:    frameRate,
		DropFrame:    dropFrame,
		TotalFrames:  totalFrames,
		TotalSeconds: totalSeconds,
	}, nil
}

// FormatTimecode formats a timecode to HH:MM:SS:FF string
func (tc *Timecode) String() string {
	sep := ":"
	if tc.DropFrame {
		sep = ";"
	}
	return fmt.Sprintf("%02d:%02d:%02d%s%02d", tc.Hours, tc.Minutes, tc.Seconds, sep, tc.Frames)
}

// msToTimecode converts milliseconds to timecode
func msToTimecode(ms float64, frameRate float64) *Timecode {
	totalSeconds := ms / 1000.0
	totalFrames := int64(totalSeconds * frameRate)

	framesPerSecond := int64(frameRate)
	framesPerMinute := framesPerSecond * 60
	framesPerHour := framesPerMinute * 60

	hours := int(totalFrames / framesPerHour)
	remaining := totalFrames % framesPerHour
	minutes := int(remaining / framesPerMinute)
	remaining = remaining % framesPerMinute
	seconds := int(remaining / framesPerSecond)
	frames := int(remaining % framesPerSecond)

	return &Timecode{
		Hours:        hours,
		Minutes:      minutes,
		Seconds:      seconds,
		Frames:       frames,
		FrameRate:    frameRate,
		TotalFrames:  totalFrames,
		TotalSeconds: totalSeconds,
	}
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
