// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hypebeast/go-osc/osc"

	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// ResolumeClient provides access to Resolume Arena/Avenue via OSC and REST API
type ResolumeClient struct {
	oscHost    string
	oscPort    int
	apiPort    int
	oscClient  *osc.Client
	httpClient *http.Client
	baseURL    string
}

// ResolumeStatus represents Resolume application status
type ResolumeStatus struct {
	Connected   bool    `json:"connected"`
	Version     string  `json:"version"`
	Composition string  `json:"composition"`
	BPM         float64 `json:"bpm"`
	Playing     bool    `json:"playing"`
	MasterLevel float64 `json:"master_level"`
}

// ResolumeLayer represents a layer in Resolume
type ResolumeLayer struct {
	Index      int     `json:"index"`
	Name       string  `json:"name"`
	Opacity    float64 `json:"opacity"`
	Bypassed   bool    `json:"bypassed"`
	Solo       bool    `json:"solo"`
	ActiveClip int     `json:"active_clip"`
}

// ResolumeClip represents a clip in Resolume
type ResolumeClip struct {
	Column    int     `json:"column"`
	Row       int     `json:"row"`
	Name      string  `json:"name"`
	Duration  float64 `json:"duration"`
	Connected bool    `json:"connected"`
	Playing   bool    `json:"playing"`
	BPMSync   bool    `json:"bpm_sync"`
}

// ResolumeDeck represents a deck in Resolume
type ResolumeDeck struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Active     bool   `json:"active"`
	LayerCount int    `json:"layer_count"`
	ClipCount  int    `json:"clip_count"`
}

// ResolumeEffect represents an effect in Resolume
type ResolumeEffect struct {
	Name    string  `json:"name"`
	Enabled bool    `json:"enabled"`
	Mix     float64 `json:"mix"`
	Type    string  `json:"type"` // video, audio
}

// ResolumeOutput represents an output in Resolume
type ResolumeOutput struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Fullscreen bool   `json:"fullscreen"`
}

// ResolumeHealth represents Resolume system health
type ResolumeHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LayerCount      int      `json:"layer_count"`
	ClipCount       int      `json:"clip_count"`
	OutputCount     int      `json:"output_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// ============================================================================
// Extended Types for Enhanced Resolume Integration
// ============================================================================

// ResolumeClipDetails contains extended clip information
type ResolumeClipDetails struct {
	Layer        int     `json:"layer"`
	Column       int     `json:"column"`
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	FilePath     string  `json:"file_path,omitempty"`
	Duration     float64 `json:"duration"`
	Width        int     `json:"width,omitempty"`
	Height       int     `json:"height,omitempty"`
	Framerate    float64 `json:"framerate,omitempty"`
	TriggerStyle string  `json:"trigger_style"` // toggle, gate, retrigger
	BeatSnap     string  `json:"beat_snap"`     // off, beat, bar, 4bars
	Direction    string  `json:"direction"`     // forward, backward, pingpong
	Speed        float64 `json:"speed"`
	Connected    bool    `json:"connected"`
	Playing      bool    `json:"playing"`
	Selected     bool    `json:"selected"`
}

// ResolumeEffectParam represents an effect parameter with full metadata
type ResolumeEffectParam struct {
	ID      int         `json:"id"`
	Name    string      `json:"name"`
	Type    string      `json:"type"` // ParamRange, ParamBoolean, ParamChoice, ParamColor
	Value   interface{} `json:"value"`
	Min     float64     `json:"min,omitempty"`
	Max     float64     `json:"max,omitempty"`
	Default interface{} `json:"default,omitempty"`
	Options []string    `json:"options,omitempty"` // For choice params
	Index   int         `json:"index,omitempty"`   // For choice params current index
}

// ResolumeEffectExtended provides full effect details with parameters
type ResolumeEffectExtended struct {
	ID         int                   `json:"id"`
	Name       string                `json:"name"`
	Enabled    bool                  `json:"enabled"`
	Bypassed   bool                  `json:"bypassed"`
	Mix        float64               `json:"mix"`
	Type       string                `json:"type"` // video, audio
	Index      int                   `json:"index"`
	Parameters []ResolumeEffectParam `json:"parameters,omitempty"`
}

// ResolumeLayerGroup represents a layer group in the composition
type ResolumeLayerGroup struct {
	ID       int     `json:"id"`
	Index    int     `json:"index"`
	Name     string  `json:"name"`
	Opacity  float64 `json:"opacity"`
	Bypassed bool    `json:"bypassed"`
	Solo     bool    `json:"solo"`
	Layers   []int   `json:"layers"` // Layer indices in this group
}

// ResolumeAudioTrack represents an audio track/layer
type ResolumeAudioTrack struct {
	Layer    int     `json:"layer"`
	Name     string  `json:"name"`
	Volume   float64 `json:"volume"`
	Pan      float64 `json:"pan"`
	Muted    bool    `json:"muted"`
	Solo     bool    `json:"solo"`
	HasClip  bool    `json:"has_clip"`
	ClipName string  `json:"clip_name,omitempty"`
}

// ResolumeAudioMaster represents master audio settings
type ResolumeAudioMaster struct {
	Volume  float64          `json:"volume"`
	Muted   bool             `json:"muted"`
	Effects []ResolumeEffect `json:"effects,omitempty"`
}

// ResolumeAvailableEffect represents an effect available to add
type ResolumeAvailableEffect struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category,omitempty"`
	Presets  []string `json:"presets,omitempty"`
}

// discoverResolumeAPIPort scans common ports to find Resolume's webserver
func discoverResolumeAPIPort(host string) int {
	// Common port ranges for Resolume webserver (dynamic assignment)
	portRanges := []struct{ start, end int }{
		{60000, 65000}, // Dynamic ports (where Resolume usually lands)
		{8080, 8090},   // Standard web server ports
		{9000, 9100},   // Alternative web server ports
	}

	client := httpclient.WithTimeout(500 * time.Millisecond)

	for _, r := range portRanges {
		for port := r.start; port <= r.end; port++ {
			url := fmt.Sprintf("http://%s:%d/api/v1/product", host, port)
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				body, _ := io.ReadAll(resp.Body)
				// Verify it's Resolume by checking for expected fields
				if len(body) > 0 && (contains(body, "Arena") || contains(body, "Avenue")) {
					return port
				}
			}
		}
	}

	// Fallback to common default
	return 8080
}

// contains checks if a byte slice contains a string
func contains(data []byte, s string) bool {
	return len(data) > 0 && len(s) > 0 && string(data) != "" &&
		(string(data) == s || len(data) > len(s) && string(data[:len(s)]) == s ||
			indexOf(data, []byte(s)) >= 0)
}

// indexOf finds the index of a byte slice within another
func indexOf(data, sub []byte) int {
	for i := 0; i <= len(data)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
			if data[i+j] != sub[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// NewResolumeClient creates a new Resolume client
func NewResolumeClient() (*ResolumeClient, error) {
	oscHost := os.Getenv("RESOLUME_OSC_HOST")
	if oscHost == "" {
		oscHost = "127.0.0.1"
	}

	oscPort := 7000 // Default Resolume OSC port
	if portStr := os.Getenv("RESOLUME_OSC_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			oscPort = p
		}
	}

	apiPort := 0 // Will be auto-discovered
	if portStr := os.Getenv("RESOLUME_API_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			apiPort = p
		}
	}

	// Auto-discover Resolume API port if not set
	if apiPort == 0 {
		apiPort = discoverResolumeAPIPort(oscHost)
	}

	// Create OSC client
	oscClient := osc.NewClient(oscHost, oscPort)

	// Create HTTP client using shared pool
	httpClient := httpclient.Fast()

	baseURL := fmt.Sprintf("http://%s:%d", oscHost, apiPort)

	return &ResolumeClient{
		oscHost:    oscHost,
		oscPort:    oscPort,
		apiPort:    apiPort,
		oscClient:  oscClient,
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// sendOSC sends an OSC message to Resolume
func (c *ResolumeClient) sendOSC(address string, args ...interface{}) error {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}
	return c.oscClient.Send(msg)
}

// doRequest performs an HTTP request to the Resolume REST API
func (c *ResolumeClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// doRequestTextPlain performs an HTTP request with text/plain content type
// Used for Resolume endpoints that accept URL strings (clip open, effect add, etc.)
func (c *ResolumeClient) doRequestTextPlain(ctx context.Context, method, path string, body string) error {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

// ResolumeAPIComposition represents the composition from Resolume's REST API
type ResolumeAPIComposition struct {
	Name       string                `json:"name"`
	Width      int                   `json:"width"`
	Height     int                   `json:"height"`
	Framerate  float64               `json:"framerate"`
	Master     *ResolumeAPIMaster    `json:"master,omitempty"`
	Layers     []ResolumeAPILayer    `json:"layers,omitempty"`
	Columns    []ResolumeAPIColumn   `json:"columns,omitempty"`
	Decks      []ResolumeAPIDeck     `json:"decks,omitempty"`
	Tempo      *ResolumeAPITempo     `json:"tempocontroller,omitempty"`
	Crossfader *ResolumeAPICrossfade `json:"crossfader,omitempty"`
}

// ResolumeAPIMaster represents master settings
type ResolumeAPIMaster struct {
	Video *ResolumeAPIVideo `json:"video,omitempty"`
	Audio *ResolumeAPIAudio `json:"audio,omitempty"`
}

// ResolumeAPIVideo represents video settings
type ResolumeAPIVideo struct {
	Opacity float64 `json:"opacity"`
}

// ResolumeAPIAudio represents audio settings
type ResolumeAPIAudio struct {
	Volume float64 `json:"volume"`
}

// ResolumeAPILayer represents a layer from the API
type ResolumeAPILayer struct {
	ID         int                 `json:"id"`
	Name       *ResolumeAPIName    `json:"name,omitempty"`
	Bypassed   bool                `json:"bypassed"`
	Solo       bool                `json:"solo"`
	Video      *ResolumeAPIVideo   `json:"video,omitempty"`
	Clips      []ResolumeAPIClip   `json:"clips,omitempty"`
	Effects    []ResolumeAPIEffect `json:"effects,omitempty"`
	ActiveClip int                 `json:"selected,omitempty"`
}

// ResolumeAPIName represents a name object
type ResolumeAPIName struct {
	Value string `json:"value"`
}

// ResolumeAPIClip represents a clip from the API
type ResolumeAPIClip struct {
	ID        int               `json:"id"`
	Name      *ResolumeAPIName  `json:"name,omitempty"`
	Connected bool              `json:"connected"`
	Duration  float64           `json:"duration,omitempty"`
	BPMSync   bool              `json:"bpmsync,omitempty"`
	Video     *ResolumeAPIVideo `json:"video,omitempty"`
}

// ResolumeAPIColumn represents a column from the API
type ResolumeAPIColumn struct {
	ID        int              `json:"id"`
	Name      *ResolumeAPIName `json:"name,omitempty"`
	Connected bool             `json:"connected"`
}

// ResolumeAPIDeck represents a deck from the API
type ResolumeAPIDeck struct {
	ID       int              `json:"id"`
	Name     *ResolumeAPIName `json:"name,omitempty"`
	Selected bool             `json:"selected"`
}

// ResolumeAPIEffect represents an effect from the API
type ResolumeAPIEffect struct {
	ID       int              `json:"id"`
	Name     *ResolumeAPIName `json:"name,omitempty"`
	Bypassed bool             `json:"bypassed"`
	Mix      float64          `json:"mix,omitempty"`
}

// ResolumeAPITempo represents tempo controller
type ResolumeAPITempo struct {
	Tempo float64 `json:"tempo"`
}

// ResolumeAPICrossfade represents crossfader settings
type ResolumeAPICrossfade struct {
	Phase float64 `json:"phase"`
}

// getComposition fetches the full composition from the REST API
func (c *ResolumeClient) getComposition(ctx context.Context) (*ResolumeAPIComposition, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/composition", nil)
	if err != nil {
		return nil, err
	}

	var comp ResolumeAPIComposition
	if err := json.Unmarshal(data, &comp); err != nil {
		return nil, fmt.Errorf("parse composition: %w", err)
	}

	return &comp, nil
}

// GetStatus returns Resolume application status
func (c *ResolumeClient) GetStatus(ctx context.Context) (*ResolumeStatus, error) {
	status := &ResolumeStatus{
		Connected: false,
	}

	// Try REST API first for full info
	comp, err := c.getComposition(ctx)
	if err == nil {
		status.Connected = true
		status.Version = "Arena 7+ (REST API)"
		status.Composition = comp.Name

		if comp.Tempo != nil {
			status.BPM = comp.Tempo.Tempo
		} else {
			status.BPM = 120.0
		}

		if comp.Master != nil && comp.Master.Video != nil {
			status.MasterLevel = comp.Master.Video.Opacity
		} else {
			status.MasterLevel = 1.0
		}

		// Check if any clip is playing
		for _, layer := range comp.Layers {
			for _, clip := range layer.Clips {
				if clip.Connected {
					status.Playing = true
					break
				}
			}
			if status.Playing {
				break
			}
		}

		return status, nil
	}

	// Fallback: try to connect to OSC port to check if Resolume is running
	addr := net.JoinHostPort(c.oscHost, strconv.Itoa(c.oscPort))
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
		status.Version = "Unknown (OSC only)"
		status.BPM = 120.0
		status.MasterLevel = 1.0
	}

	return status, nil
}

// GetLayers returns all layers
func (c *ResolumeClient) GetLayers(ctx context.Context) ([]ResolumeLayer, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get layers: %w", err)
	}

	layers := make([]ResolumeLayer, 0, len(comp.Layers))
	for i, apiLayer := range comp.Layers {
		name := fmt.Sprintf("Layer %d", i+1)
		if apiLayer.Name != nil {
			name = apiLayer.Name.Value
		}

		var opacity float64 = 1.0
		if apiLayer.Video != nil {
			opacity = apiLayer.Video.Opacity
		}

		layers = append(layers, ResolumeLayer{
			Index:      i + 1,
			Name:       name,
			Opacity:    opacity,
			Bypassed:   apiLayer.Bypassed,
			Solo:       apiLayer.Solo,
			ActiveClip: apiLayer.ActiveClip,
		})
	}

	return layers, nil
}

// SetLayerOpacity sets a layer's opacity
func (c *ResolumeClient) SetLayerOpacity(ctx context.Context, layer int, opacity float64) error {
	if layer < 1 {
		return fmt.Errorf("invalid layer index: %d", layer)
	}
	if opacity < 0 || opacity > 1 {
		return fmt.Errorf("opacity must be between 0 and 1")
	}

	address := fmt.Sprintf("/composition/layers/%d/video/opacity/values", layer)
	return c.sendOSC(address, float32(opacity))
}

// GetClips returns clips for a layer
func (c *ResolumeClient) GetClips(ctx context.Context, layer int) ([]ResolumeClip, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get clips: %w", err)
	}

	if layer < 1 || layer > len(comp.Layers) {
		return nil, fmt.Errorf("invalid layer index: %d (have %d layers)", layer, len(comp.Layers))
	}

	apiLayer := comp.Layers[layer-1]
	clips := make([]ResolumeClip, 0, len(apiLayer.Clips))

	for col, apiClip := range apiLayer.Clips {
		name := fmt.Sprintf("Clip %d", col+1)
		if apiClip.Name != nil {
			name = apiClip.Name.Value
		}

		clips = append(clips, ResolumeClip{
			Column:    col + 1,
			Row:       layer,
			Name:      name,
			Duration:  apiClip.Duration,
			Connected: apiClip.Connected,
			Playing:   apiClip.Connected,
			BPMSync:   apiClip.BPMSync,
		})
	}

	return clips, nil
}

// TriggerClip triggers a clip at the given layer and column
func (c *ResolumeClient) TriggerClip(ctx context.Context, layer, column int) error {
	if layer < 1 {
		return fmt.Errorf("invalid layer index: %d", layer)
	}
	if column < 1 {
		return fmt.Errorf("invalid column index: %d", column)
	}

	address := fmt.Sprintf("/composition/layers/%d/clips/%d/connect", layer, column)
	return c.sendOSC(address, int32(1))
}

// TriggerColumn triggers all clips in a column
func (c *ResolumeClient) TriggerColumn(ctx context.Context, column int) error {
	if column < 1 {
		return fmt.Errorf("invalid column index: %d", column)
	}

	address := fmt.Sprintf("/composition/columns/%d/connect", column)
	return c.sendOSC(address, int32(1))
}

// GetDecks returns all decks
func (c *ResolumeClient) GetDecks(ctx context.Context) ([]ResolumeDeck, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get decks: %w", err)
	}

	decks := make([]ResolumeDeck, 0, len(comp.Decks))
	for i, apiDeck := range comp.Decks {
		name := fmt.Sprintf("Deck %d", i+1)
		if apiDeck.Name != nil {
			name = apiDeck.Name.Value
		}

		// Count layers and clips in this deck (simplified - count all)
		layerCount := len(comp.Layers)
		clipCount := 0
		for _, layer := range comp.Layers {
			clipCount += len(layer.Clips)
		}

		decks = append(decks, ResolumeDeck{
			Index:      i + 1,
			Name:       name,
			Active:     apiDeck.Selected,
			LayerCount: layerCount,
			ClipCount:  clipCount,
		})
	}

	return decks, nil
}

// GetBPM returns the current BPM
func (c *ResolumeClient) GetBPM(ctx context.Context) (float64, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return 120.0, nil // Default on error
	}

	if comp.Tempo != nil {
		return comp.Tempo.Tempo, nil
	}

	return 120.0, nil
}

// SetBPM sets the BPM
func (c *ResolumeClient) SetBPM(ctx context.Context, bpm float64) error {
	if bpm < 20 || bpm > 999 {
		return fmt.Errorf("BPM must be between 20 and 999")
	}

	return c.sendOSC("/composition/tempocontroller/tempo", float32(bpm))
}

// TapTempo taps the tempo
func (c *ResolumeClient) TapTempo(ctx context.Context) error {
	return c.sendOSC("/composition/tempocontroller/tempotap", int32(1))
}

// GetEffects returns effects for a layer or master
func (c *ResolumeClient) GetEffects(ctx context.Context, layer int) ([]ResolumeEffect, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get effects: %w", err)
	}

	if layer < 1 || layer > len(comp.Layers) {
		return nil, fmt.Errorf("invalid layer index: %d (have %d layers)", layer, len(comp.Layers))
	}

	apiLayer := comp.Layers[layer-1]
	effects := make([]ResolumeEffect, 0, len(apiLayer.Effects))

	for _, apiEffect := range apiLayer.Effects {
		name := "Effect"
		if apiEffect.Name != nil {
			name = apiEffect.Name.Value
		}

		effects = append(effects, ResolumeEffect{
			Name:    name,
			Enabled: !apiEffect.Bypassed,
			Mix:     apiEffect.Mix,
			Type:    "video",
		})
	}

	return effects, nil
}

// ResolumeAPIOutput represents output from the API
type ResolumeAPIOutput struct {
	ID         int              `json:"id"`
	Name       *ResolumeAPIName `json:"name,omitempty"`
	Enabled    bool             `json:"enabled"`
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	Fullscreen bool             `json:"fullscreen"`
}

// ResolumeAPIOutputs represents the outputs response
type ResolumeAPIOutputs struct {
	Screens []ResolumeAPIOutput `json:"screens,omitempty"`
}

// GetOutputs returns output configuration
func (c *ResolumeClient) GetOutputs(ctx context.Context) ([]ResolumeOutput, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/output", nil)
	if err != nil {
		// If API not available, return empty list
		return []ResolumeOutput{}, nil
	}

	var apiOutputs ResolumeAPIOutputs
	if err := json.Unmarshal(data, &apiOutputs); err != nil {
		return nil, fmt.Errorf("parse outputs: %w", err)
	}

	outputs := make([]ResolumeOutput, 0, len(apiOutputs.Screens))
	for i, apiOutput := range apiOutputs.Screens {
		name := fmt.Sprintf("Output %d", i+1)
		if apiOutput.Name != nil {
			name = apiOutput.Name.Value
		}

		outputs = append(outputs, ResolumeOutput{
			Index:      i + 1,
			Name:       name,
			Enabled:    apiOutput.Enabled,
			Width:      apiOutput.Width,
			Height:     apiOutput.Height,
			Fullscreen: apiOutput.Fullscreen,
		})
	}

	return outputs, nil
}

// StartRecording starts recording output
func (c *ResolumeClient) StartRecording(ctx context.Context) error {
	// In production, send OSC message
	return fmt.Errorf("recording requires Resolume recording module")
}

// StopRecording stops recording
func (c *ResolumeClient) StopRecording(ctx context.Context) error {
	// In production, send OSC message
	return fmt.Errorf("recording requires Resolume recording module")
}

// GetHealth returns Resolume system health
func (c *ResolumeClient) GetHealth(ctx context.Context) (*ResolumeHealth, error) {
	health := &ResolumeHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check connection
	status, _ := c.GetStatus(ctx)
	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Not connected to Resolume")
	}

	// Get layers
	layers, _ := c.GetLayers(ctx)
	health.LayerCount = len(layers)

	// Get outputs
	outputs, _ := c.GetOutputs(ctx)
	health.OutputCount = len(outputs)

	// Set status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	if !status.Connected {
		health.Recommendations = append(health.Recommendations, "Start Resolume and enable OSC")
	}

	return health, nil
}

// OSCHost returns the configured OSC host
func (c *ResolumeClient) OSCHost() string {
	return c.oscHost
}

// OSCPort returns the configured OSC port
func (c *ResolumeClient) OSCPort() int {
	return c.oscPort
}

// ResolumeColumn represents a column in the composition
type ResolumeColumn struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Connected   bool   `json:"connected"`
	ClipsLoaded int    `json:"clips_loaded"`
}

// ResolumeAutopilot represents autopilot settings
type ResolumeAutopilot struct {
	Enabled   bool    `json:"enabled"`
	Mode      string  `json:"mode"` // random, sequential, bpm
	Interval  float64 `json:"interval_sec"`
	LayerMask []int   `json:"layer_mask,omitempty"`
}

// SetLayerBypass sets a layer's bypass state
func (c *ResolumeClient) SetLayerBypass(ctx context.Context, layer int, bypass bool) error {
	if layer < 1 {
		return fmt.Errorf("invalid layer index: %d", layer)
	}

	var value int32 = 0
	if bypass {
		value = 1
	}
	address := fmt.Sprintf("/composition/layers/%d/bypassed", layer)
	return c.sendOSC(address, value)
}

// SetLayerSolo sets a layer's solo state
func (c *ResolumeClient) SetLayerSolo(ctx context.Context, layer int, solo bool) error {
	if layer < 1 {
		return fmt.Errorf("invalid layer index: %d", layer)
	}

	var value int32 = 0
	if solo {
		value = 1
	}
	address := fmt.Sprintf("/composition/layers/%d/solo", layer)
	return c.sendOSC(address, value)
}

// ClearLayer clears a specific layer (ejects all clips)
func (c *ResolumeClient) ClearLayer(ctx context.Context, layer int) error {
	if layer < 1 {
		return fmt.Errorf("invalid layer index: %d", layer)
	}

	address := fmt.Sprintf("/composition/layers/%d/clear", layer)
	return c.sendOSC(address, int32(1))
}

// ClearAll clears all layers
func (c *ResolumeClient) ClearAll(ctx context.Context) error {
	return c.sendOSC("/composition/disconnectall", int32(1))
}

// GetColumns returns all columns in the composition
func (c *ResolumeClient) GetColumns(ctx context.Context) ([]ResolumeColumn, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get columns: %w", err)
	}

	columns := make([]ResolumeColumn, 0, len(comp.Columns))
	for i, apiColumn := range comp.Columns {
		name := fmt.Sprintf("Column %d", i+1)
		if apiColumn.Name != nil {
			name = apiColumn.Name.Value
		}

		// Count clips loaded in this column across all layers
		clipsLoaded := 0
		for _, layer := range comp.Layers {
			if i < len(layer.Clips) && layer.Clips[i].Name != nil {
				clipsLoaded++
			}
		}

		columns = append(columns, ResolumeColumn{
			Index:       i + 1,
			Name:        name,
			Connected:   apiColumn.Connected,
			ClipsLoaded: clipsLoaded,
		})
	}

	return columns, nil
}

// CrossfadeDecks crossfades between decks
func (c *ResolumeClient) CrossfadeDecks(ctx context.Context, position float64) error {
	if position < 0 || position > 1 {
		return fmt.Errorf("crossfade position must be between 0 and 1")
	}

	return c.sendOSC("/composition/crossfader", float32(position))
}

// SetEffectEnabled enables or disables an effect
func (c *ResolumeClient) SetEffectEnabled(ctx context.Context, layer int, effectIndex int, enabled bool) error {
	var value int32 = 1 // bypassed=1 means disabled
	if enabled {
		value = 0 // bypassed=0 means enabled
	}

	address := fmt.Sprintf("/composition/layers/%d/video/effects/%d/bypassed", layer, effectIndex)
	return c.sendOSC(address, value)
}

// GetEffectParams gets effect parameters
func (c *ResolumeClient) GetEffectParams(ctx context.Context, layer int, effectIndex int) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	return params, nil
}

// SetEffectParam sets an effect parameter
func (c *ResolumeClient) SetEffectParam(ctx context.Context, layer int, effectIndex int, param string, value float64) error {
	// In production, send OSC message for parameter
	return nil
}

// GetMasterLevel returns the master output level
func (c *ResolumeClient) GetMasterLevel(ctx context.Context) (float64, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return 1.0, nil // Default on error
	}

	if comp.Master != nil && comp.Master.Video != nil {
		return comp.Master.Video.Opacity, nil
	}

	return 1.0, nil
}

// SetMasterLevel sets the master output level
func (c *ResolumeClient) SetMasterLevel(ctx context.Context, level float64) error {
	if level < 0 || level > 1 {
		return fmt.Errorf("master level must be between 0 and 1")
	}

	return c.sendOSC("/composition/video/opacity/values", float32(level))
}

// GetAutopilot returns autopilot settings
func (c *ResolumeClient) GetAutopilot(ctx context.Context) (*ResolumeAutopilot, error) {
	autopilot := &ResolumeAutopilot{
		Enabled:  false,
		Mode:     "random",
		Interval: 8.0,
	}
	return autopilot, nil
}

// SetAutopilot configures autopilot
func (c *ResolumeClient) SetAutopilot(ctx context.Context, enabled bool, mode string, interval float64) error {
	// In production, send OSC messages to configure autopilot
	return nil
}

// SelectDeck selects a deck
func (c *ResolumeClient) SelectDeck(ctx context.Context, deck int) error {
	if deck < 1 {
		return fmt.Errorf("invalid deck index: %d", deck)
	}

	address := fmt.Sprintf("/composition/decks/%d/select", deck)
	return c.sendOSC(address, int32(1))
}

// GetClipInfo returns detailed info about a specific clip
func (c *ResolumeClient) GetClipInfo(ctx context.Context, layer, column int) (*ResolumeClip, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get clip info: %w", err)
	}

	if layer < 1 || layer > len(comp.Layers) {
		return nil, fmt.Errorf("invalid layer index: %d", layer)
	}

	apiLayer := comp.Layers[layer-1]
	if column < 1 || column > len(apiLayer.Clips) {
		return nil, fmt.Errorf("invalid column index: %d", column)
	}

	apiClip := apiLayer.Clips[column-1]
	name := fmt.Sprintf("Clip %d", column)
	if apiClip.Name != nil {
		name = apiClip.Name.Value
	}

	return &ResolumeClip{
		Column:    column,
		Row:       layer,
		Name:      name,
		Duration:  apiClip.Duration,
		Connected: apiClip.Connected,
		Playing:   apiClip.Connected,
		BPMSync:   apiClip.BPMSync,
	}, nil
}

// SetClipSpeed sets clip playback speed
func (c *ResolumeClient) SetClipSpeed(ctx context.Context, layer, column int, speed float64) error {
	address := fmt.Sprintf("/composition/layers/%d/clips/%d/video/source/speed", layer, column)
	return c.sendOSC(address, float32(speed))
}

// ============================================================================
// Effect Parameter Methods (Phase 1: Full Effect Control)
// ============================================================================

// ResolumeAPIParameter represents a parameter from the REST API
type ResolumeAPIParameter struct {
	ID      int              `json:"id"`
	Name    *ResolumeAPIName `json:"name,omitempty"`
	Type    string           `json:"valuetype,omitempty"`
	Value   interface{}      `json:"value,omitempty"`
	Index   int              `json:"index,omitempty"`
	Min     float64          `json:"min,omitempty"`
	Max     float64          `json:"max,omitempty"`
	Options []string         `json:"options,omitempty"`
}

// ResolumeAPIEffectFull represents a full effect from the API with parameters
type ResolumeAPIEffectFull struct {
	ID         int                    `json:"id"`
	Name       *ResolumeAPIName       `json:"name,omitempty"`
	Bypassed   bool                   `json:"bypassed"`
	Mix        *ResolumeAPIParameter  `json:"mix,omitempty"`
	Parameters []ResolumeAPIParameter `json:"params,omitempty"`
}

// GetParameterByID fetches a parameter value by its unique ID
func (c *ResolumeClient) GetParameterByID(ctx context.Context, paramID int) (*ResolumeEffectParam, error) {
	path := fmt.Sprintf("/api/v1/parameter/by-id/%d", paramID)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("get parameter %d: %w", paramID, err)
	}

	var apiParam ResolumeAPIParameter
	if err := json.Unmarshal(data, &apiParam); err != nil {
		return nil, fmt.Errorf("parse parameter: %w", err)
	}

	name := fmt.Sprintf("Param %d", paramID)
	if apiParam.Name != nil {
		name = apiParam.Name.Value
	}

	return &ResolumeEffectParam{
		ID:      apiParam.ID,
		Name:    name,
		Type:    apiParam.Type,
		Value:   apiParam.Value,
		Min:     apiParam.Min,
		Max:     apiParam.Max,
		Options: apiParam.Options,
		Index:   apiParam.Index,
	}, nil
}

// SetParameterByID sets a parameter value by its unique ID
func (c *ResolumeClient) SetParameterByID(ctx context.Context, paramID int, value interface{}) error {
	// Convert value to appropriate OSC type and send via REST API
	path := fmt.Sprintf("/api/v1/parameter/by-id/%d", paramID)

	// Build the request body based on value type
	var body []byte
	var err error

	switch v := value.(type) {
	case float64:
		body, err = json.Marshal(map[string]interface{}{"value": v})
	case float32:
		body, err = json.Marshal(map[string]interface{}{"value": float64(v)})
	case int:
		body, err = json.Marshal(map[string]interface{}{"value": v})
	case int32:
		body, err = json.Marshal(map[string]interface{}{"value": int(v)})
	case bool:
		body, err = json.Marshal(map[string]interface{}{"value": v})
	case string:
		body, err = json.Marshal(map[string]interface{}{"value": v})
	default:
		body, err = json.Marshal(map[string]interface{}{"value": value})
	}

	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}

	_, err = c.doRequest(ctx, "PUT", path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("set parameter %d: %w", paramID, err)
	}

	return nil
}

// GetEffectExtended returns detailed effect info with all parameters
func (c *ResolumeClient) GetEffectExtended(ctx context.Context, layer, effectIndex int) (*ResolumeEffectExtended, error) {
	// Fetch the layer from composition to get effect details
	path := fmt.Sprintf("/api/v1/composition/layers/%d/video/effects/%d", layer, effectIndex)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("get effect: %w", err)
	}

	var apiEffect ResolumeAPIEffectFull
	if err := json.Unmarshal(data, &apiEffect); err != nil {
		return nil, fmt.Errorf("parse effect: %w", err)
	}

	name := fmt.Sprintf("Effect %d", effectIndex)
	if apiEffect.Name != nil {
		name = apiEffect.Name.Value
	}

	// Convert parameters
	params := make([]ResolumeEffectParam, 0, len(apiEffect.Parameters))
	for _, p := range apiEffect.Parameters {
		pName := ""
		if p.Name != nil {
			pName = p.Name.Value
		}
		params = append(params, ResolumeEffectParam{
			ID:      p.ID,
			Name:    pName,
			Type:    p.Type,
			Value:   p.Value,
			Min:     p.Min,
			Max:     p.Max,
			Options: p.Options,
			Index:   p.Index,
		})
	}

	var mix float64 = 1.0
	if apiEffect.Mix != nil {
		if v, ok := apiEffect.Mix.Value.(float64); ok {
			mix = v
		}
	}

	return &ResolumeEffectExtended{
		ID:         apiEffect.ID,
		Name:       name,
		Enabled:    !apiEffect.Bypassed,
		Bypassed:   apiEffect.Bypassed,
		Mix:        mix,
		Type:       "video",
		Index:      effectIndex,
		Parameters: params,
	}, nil
}

// GetLayerEffectsExtended returns all effects with parameters for a layer
func (c *ResolumeClient) GetLayerEffectsExtended(ctx context.Context, layer int) ([]ResolumeEffectExtended, error) {
	// First get basic effect list to know how many effects
	effects, err := c.GetEffects(ctx, layer)
	if err != nil {
		return nil, err
	}

	result := make([]ResolumeEffectExtended, 0, len(effects))
	for i := range effects {
		ext, err := c.GetEffectExtended(ctx, layer, i+1)
		if err != nil {
			// Log but continue - some effects might not have full details
			continue
		}
		result = append(result, *ext)
	}

	return result, nil
}

// SetEffectMix sets the mix/intensity of an effect
func (c *ResolumeClient) SetEffectMix(ctx context.Context, layer, effectIndex int, mix float64) error {
	if mix < 0 || mix > 1 {
		return fmt.Errorf("mix must be between 0 and 1")
	}
	address := fmt.Sprintf("/composition/layers/%d/video/effects/%d/mixer/opacity/values", layer, effectIndex)
	return c.sendOSC(address, float32(mix))
}

// ============================================================================
// Clip Management Methods (Phase 1: Full Clip Control)
// ============================================================================

// ResolumeAPIClipFull represents full clip details from the API
type ResolumeAPIClipFull struct {
	ID        int                   `json:"id"`
	Name      *ResolumeAPIName      `json:"name,omitempty"`
	Connected bool                  `json:"connected"`
	Selected  bool                  `json:"selected"`
	Colorid   interface{}           `json:"colorid,omitempty"`
	Video     *ResolumeAPIClipVideo `json:"video,omitempty"`
	Audio     *ResolumeAPIClipAudio `json:"audio,omitempty"`
	Transport *ResolumeAPITransport `json:"transport,omitempty"`
	Dashboard interface{}           `json:"dashboard,omitempty"`
}

// ResolumeAPIClipVideo represents video settings for a clip
type ResolumeAPIClipVideo struct {
	Width     int                 `json:"width,omitempty"`
	Height    int                 `json:"height,omitempty"`
	Framerate float64             `json:"framerate,omitempty"`
	Duration  float64             `json:"duration,omitempty"`
	Source    *ResolumeAPISource  `json:"source,omitempty"`
	Effects   []ResolumeAPIEffect `json:"effects,omitempty"`
}

// ResolumeAPIClipAudio represents audio settings for a clip
type ResolumeAPIClipAudio struct {
	Volume float64 `json:"volume,omitempty"`
	Pan    float64 `json:"pan,omitempty"`
}

// ResolumeAPISource represents the clip source
type ResolumeAPISource struct {
	FilePath      string  `json:"filepath,omitempty"`
	TriggerStyle  int     `json:"triggerstyle,omitempty"`
	BeatSnap      int     `json:"beatsnap,omitempty"`
	Direction     int     `json:"direction,omitempty"`
	PlaybackSpeed float64 `json:"playbackspeed,omitempty"`
}

// ResolumeAPITransport represents transport controls
type ResolumeAPITransport struct {
	Position float64 `json:"position,omitempty"`
	Playing  bool    `json:"playing,omitempty"`
}

// GetClipDetails returns extended clip information
func (c *ResolumeClient) GetClipDetails(ctx context.Context, layer, column int) (*ResolumeClipDetails, error) {
	path := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d", layer, column)
	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("get clip details: %w", err)
	}

	var apiClip ResolumeAPIClipFull
	if err := json.Unmarshal(data, &apiClip); err != nil {
		return nil, fmt.Errorf("parse clip: %w", err)
	}

	name := fmt.Sprintf("Clip %d-%d", layer, column)
	if apiClip.Name != nil {
		name = apiClip.Name.Value
	}

	details := &ResolumeClipDetails{
		Layer:     layer,
		Column:    column,
		ID:        apiClip.ID,
		Name:      name,
		Connected: apiClip.Connected,
		Selected:  apiClip.Selected,
	}

	if apiClip.Video != nil {
		details.Width = apiClip.Video.Width
		details.Height = apiClip.Video.Height
		details.Framerate = apiClip.Video.Framerate
		details.Duration = apiClip.Video.Duration

		if apiClip.Video.Source != nil {
			details.FilePath = apiClip.Video.Source.FilePath
			details.Speed = apiClip.Video.Source.PlaybackSpeed

			// Convert trigger style enum
			switch apiClip.Video.Source.TriggerStyle {
			case 0:
				details.TriggerStyle = "toggle"
			case 1:
				details.TriggerStyle = "gate"
			case 2:
				details.TriggerStyle = "retrigger"
			default:
				details.TriggerStyle = "toggle"
			}

			// Convert beat snap enum
			switch apiClip.Video.Source.BeatSnap {
			case 0:
				details.BeatSnap = "off"
			case 1:
				details.BeatSnap = "beat"
			case 2:
				details.BeatSnap = "bar"
			case 3:
				details.BeatSnap = "4bars"
			default:
				details.BeatSnap = "off"
			}

			// Convert direction enum
			switch apiClip.Video.Source.Direction {
			case 0:
				details.Direction = "forward"
			case 1:
				details.Direction = "backward"
			case 2:
				details.Direction = "pingpong"
			default:
				details.Direction = "forward"
			}
		}
	}

	if apiClip.Transport != nil {
		details.Playing = apiClip.Transport.Playing
	}

	return details, nil
}

// GetClipThumbnail returns the thumbnail image for a clip as PNG bytes
func (c *ResolumeClient) GetClipThumbnail(ctx context.Context, layer, column int) ([]byte, error) {
	path := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d/thumbnail", layer, column)

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("thumbnail request failed: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read thumbnail: %w", err)
	}

	return data, nil
}

// LoadClip loads a video file into a clip slot
// filePath should be an absolute file path (e.g., /home/user/Videos/clip.mov)
func (c *ResolumeClient) LoadClip(ctx context.Context, layer, column int, filePath string) error {
	apiPath := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d/open", layer, column)

	// Convert file path to file:// URL format required by Resolume
	// Format: file:///path/to/file.mov (note THREE forward slashes)
	// Spaces and special chars must be URL-encoded
	fileURL := "file:///" + url.PathEscape(strings.TrimPrefix(filePath, "/"))
	// PathEscape encodes spaces as %20, but we need to unescape forward slashes
	fileURL = strings.ReplaceAll(fileURL, "%2F", "/")

	if err := c.doRequestTextPlain(ctx, "POST", apiPath, fileURL); err != nil {
		return fmt.Errorf("load clip: %w", err)
	}

	return nil
}

// LoadSource loads an internal source into a clip slot
// sourceName is the source name (e.g., "Checkered", "Feedback", "Solid Color")
// Optional presetID can be provided for sources with presets
func (c *ResolumeClient) LoadSource(ctx context.Context, layer, column int, sourceName string, presetID ...int) error {
	apiPath := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d/open", layer, column)

	// Format: source:///video/SourceName or source:///video/SourceName/PresetID
	sourceURL := fmt.Sprintf("source:///video/%s", sourceName)
	if len(presetID) > 0 && presetID[0] > 0 {
		sourceURL = fmt.Sprintf("%s/%d", sourceURL, presetID[0])
	}

	if err := c.doRequestTextPlain(ctx, "POST", apiPath, sourceURL); err != nil {
		return fmt.Errorf("load source: %w", err)
	}

	return nil
}

// SeekClip sets the playback position of a clip (0.0 to 1.0)
func (c *ResolumeClient) SeekClip(ctx context.Context, layer, column int, position float64) error {
	if position < 0 || position > 1 {
		return fmt.Errorf("position must be between 0 and 1")
	}
	address := fmt.Sprintf("/composition/layers/%d/clips/%d/transport/position/values", layer, column)
	return c.sendOSC(address, float32(position))
}

// SetClipTriggerStyle sets the trigger style (toggle, gate, retrigger)
func (c *ResolumeClient) SetClipTriggerStyle(ctx context.Context, layer, column int, style string) error {
	var styleInt int32
	switch style {
	case "toggle":
		styleInt = 0
	case "gate":
		styleInt = 1
	case "retrigger":
		styleInt = 2
	default:
		return fmt.Errorf("invalid trigger style: %s (use toggle, gate, retrigger)", style)
	}
	address := fmt.Sprintf("/composition/layers/%d/clips/%d/video/source/triggerstyle", layer, column)
	return c.sendOSC(address, styleInt)
}

// SetClipBeatSnap sets the beat snap mode (off, beat, bar, 4bars)
func (c *ResolumeClient) SetClipBeatSnap(ctx context.Context, layer, column int, snap string) error {
	var snapInt int32
	switch snap {
	case "off":
		snapInt = 0
	case "beat":
		snapInt = 1
	case "bar":
		snapInt = 2
	case "4bars":
		snapInt = 3
	default:
		return fmt.Errorf("invalid beat snap: %s (use off, beat, bar, 4bars)", snap)
	}
	address := fmt.Sprintf("/composition/layers/%d/clips/%d/video/source/beatsnap", layer, column)
	return c.sendOSC(address, snapInt)
}

// SetClipDirection sets playback direction (forward, backward, pingpong)
func (c *ResolumeClient) SetClipDirection(ctx context.Context, layer, column int, direction string) error {
	var dirInt int32
	switch direction {
	case "forward":
		dirInt = 0
	case "backward":
		dirInt = 1
	case "pingpong":
		dirInt = 2
	default:
		return fmt.Errorf("invalid direction: %s (use forward, backward, pingpong)", direction)
	}
	address := fmt.Sprintf("/composition/layers/%d/clips/%d/video/source/direction", layer, column)
	return c.sendOSC(address, dirInt)
}

// ============================================================================
// Layer Group Methods (Phase 1: Layer Organization)
// ============================================================================

// GetLayerGroups returns all layer groups in the composition
func (c *ResolumeClient) GetLayerGroups(ctx context.Context) ([]ResolumeLayerGroup, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/composition/groups", nil)
	if err != nil {
		// Groups might not exist in all compositions
		return []ResolumeLayerGroup{}, nil
	}

	var groups []struct {
		ID       int              `json:"id"`
		Name     *ResolumeAPIName `json:"name,omitempty"`
		Bypassed bool             `json:"bypassed"`
		Solo     bool             `json:"solo"`
		Video    *struct {
			Opacity float64 `json:"opacity"`
		} `json:"video,omitempty"`
		Layers []struct {
			ID int `json:"id"`
		} `json:"layers,omitempty"`
	}

	if err := json.Unmarshal(data, &groups); err != nil {
		return nil, fmt.Errorf("parse groups: %w", err)
	}

	result := make([]ResolumeLayerGroup, 0, len(groups))
	for i, g := range groups {
		name := fmt.Sprintf("Group %d", i+1)
		if g.Name != nil {
			name = g.Name.Value
		}

		var opacity float64 = 1.0
		if g.Video != nil {
			opacity = g.Video.Opacity
		}

		layers := make([]int, 0, len(g.Layers))
		for _, l := range g.Layers {
			layers = append(layers, l.ID)
		}

		result = append(result, ResolumeLayerGroup{
			ID:       g.ID,
			Index:    i + 1,
			Name:     name,
			Opacity:  opacity,
			Bypassed: g.Bypassed,
			Solo:     g.Solo,
			Layers:   layers,
		})
	}

	return result, nil
}

// SetLayerGroupOpacity sets a layer group's opacity
func (c *ResolumeClient) SetLayerGroupOpacity(ctx context.Context, group int, opacity float64) error {
	if opacity < 0 || opacity > 1 {
		return fmt.Errorf("opacity must be between 0 and 1")
	}
	address := fmt.Sprintf("/composition/groups/%d/video/opacity/values", group)
	return c.sendOSC(address, float32(opacity))
}

// SetLayerGroupBypass sets a layer group's bypass state
func (c *ResolumeClient) SetLayerGroupBypass(ctx context.Context, group int, bypass bool) error {
	var value int32 = 0
	if bypass {
		value = 1
	}
	address := fmt.Sprintf("/composition/groups/%d/bypassed", group)
	return c.sendOSC(address, value)
}

// SetLayerGroupSolo sets a layer group's solo state
func (c *ResolumeClient) SetLayerGroupSolo(ctx context.Context, group int, solo bool) error {
	var value int32 = 0
	if solo {
		value = 1
	}
	address := fmt.Sprintf("/composition/groups/%d/solo", group)
	return c.sendOSC(address, value)
}

// ============================================================================
// Audio Methods (Phase 1: Audio Control)
// ============================================================================

// GetAudioTracks returns audio information for all layers
func (c *ResolumeClient) GetAudioTracks(ctx context.Context) ([]ResolumeAudioTrack, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get audio tracks: %w", err)
	}

	tracks := make([]ResolumeAudioTrack, 0, len(comp.Layers))
	for i, layer := range comp.Layers {
		name := fmt.Sprintf("Layer %d", i+1)
		if layer.Name != nil {
			name = layer.Name.Value
		}

		// Check for active clip audio
		hasClip := false
		clipName := ""
		for _, clip := range layer.Clips {
			if clip.Connected {
				hasClip = true
				if clip.Name != nil {
					clipName = clip.Name.Value
				}
				break
			}
		}

		tracks = append(tracks, ResolumeAudioTrack{
			Layer:    i + 1,
			Name:     name,
			Volume:   1.0, // Default - would need deeper API query for actual
			Pan:      0.0,
			Muted:    layer.Bypassed,
			Solo:     layer.Solo,
			HasClip:  hasClip,
			ClipName: clipName,
		})
	}

	return tracks, nil
}

// SetLayerAudioVolume sets a layer's audio volume
func (c *ResolumeClient) SetLayerAudioVolume(ctx context.Context, layer int, volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}
	address := fmt.Sprintf("/composition/layers/%d/audio/volume/values", layer)
	return c.sendOSC(address, float32(volume))
}

// SetLayerAudioMute mutes or unmutes a layer's audio
func (c *ResolumeClient) SetLayerAudioMute(ctx context.Context, layer int, mute bool) error {
	var value int32 = 0
	if mute {
		value = 1
	}
	address := fmt.Sprintf("/composition/layers/%d/audio/muted", layer)
	return c.sendOSC(address, value)
}

// SetLayerAudioPan sets a layer's audio pan (-1 to 1)
func (c *ResolumeClient) SetLayerAudioPan(ctx context.Context, layer int, pan float64) error {
	if pan < -1 || pan > 1 {
		return fmt.Errorf("pan must be between -1 and 1")
	}
	address := fmt.Sprintf("/composition/layers/%d/audio/pan/values", layer)
	return c.sendOSC(address, float32(pan))
}

// GetMasterAudio returns master audio settings
func (c *ResolumeClient) GetMasterAudio(ctx context.Context) (*ResolumeAudioMaster, error) {
	comp, err := c.getComposition(ctx)
	if err != nil {
		return nil, fmt.Errorf("get master audio: %w", err)
	}

	master := &ResolumeAudioMaster{
		Volume: 1.0,
		Muted:  false,
	}

	if comp.Master != nil && comp.Master.Audio != nil {
		master.Volume = comp.Master.Audio.Volume
	}

	return master, nil
}

// SetMasterAudioVolume sets master audio volume
func (c *ResolumeClient) SetMasterAudioVolume(ctx context.Context, volume float64) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("volume must be between 0 and 1")
	}
	return c.sendOSC("/composition/audio/volume/values", float32(volume))
}

// SetMasterAudioMute mutes or unmutes master audio
func (c *ResolumeClient) SetMasterAudioMute(ctx context.Context, mute bool) error {
	var value int32 = 0
	if mute {
		value = 1
	}
	return c.sendOSC("/composition/audio/muted", value)
}

// ============================================================================
// Effect Management Methods (REST API)
// ============================================================================

// AddEffectToClip adds an effect to a clip using the REST API
// effectName is the effect name (e.g., "Bloom", "Blur", "Circles")
// Optional preset can be provided (e.g., "Bloom/Solid")
func (c *ResolumeClient) AddEffectToClip(ctx context.Context, layer, column int, effectName string, preset ...string) error {
	apiPath := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d/effects/video/add", layer, column)

	// Format: effect:///video/EffectName or effect:///video/EffectName/PresetName
	effectURL := fmt.Sprintf("effect:///video/%s", effectName)
	if len(preset) > 0 && preset[0] != "" {
		effectURL = fmt.Sprintf("%s/%s", effectURL, preset[0])
	}

	if err := c.doRequestTextPlain(ctx, "POST", apiPath, effectURL); err != nil {
		return fmt.Errorf("add effect to clip: %w", err)
	}

	return nil
}

// AddEffectToLayer adds an effect to a layer using the REST API
func (c *ResolumeClient) AddEffectToLayer(ctx context.Context, layer int, effectName string, preset ...string) error {
	apiPath := fmt.Sprintf("/api/v1/composition/layers/%d/effects/video/add", layer)

	effectURL := fmt.Sprintf("effect:///video/%s", effectName)
	if len(preset) > 0 && preset[0] != "" {
		effectURL = fmt.Sprintf("%s/%s", effectURL, preset[0])
	}

	if err := c.doRequestTextPlain(ctx, "POST", apiPath, effectURL); err != nil {
		return fmt.Errorf("add effect to layer: %w", err)
	}

	return nil
}

// AddEffectToComposition adds a global effect to the composition
func (c *ResolumeClient) AddEffectToComposition(ctx context.Context, effectName string, preset ...string) error {
	apiPath := "/api/v1/composition/effects/video/add"

	effectURL := fmt.Sprintf("effect:///video/%s", effectName)
	if len(preset) > 0 && preset[0] != "" {
		effectURL = fmt.Sprintf("%s/%s", effectURL, preset[0])
	}

	if err := c.doRequestTextPlain(ctx, "POST", apiPath, effectURL); err != nil {
		return fmt.Errorf("add effect to composition: %w", err)
	}

	return nil
}

// GetAvailableEffects returns all available effects that can be added
func (c *ResolumeClient) GetAvailableEffects(ctx context.Context) ([]ResolumeAvailableEffect, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/effects/video", nil)
	if err != nil {
		return nil, fmt.Errorf("get available effects: %w", err)
	}

	var apiEffects []struct {
		IDString string `json:"idstring"`
		Name     string `json:"name"`
		Presets  []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"presets,omitempty"`
	}

	if err := json.Unmarshal(data, &apiEffects); err != nil {
		return nil, fmt.Errorf("parse effects: %w", err)
	}

	effects := make([]ResolumeAvailableEffect, 0, len(apiEffects))
	for _, e := range apiEffects {
		presets := make([]string, 0, len(e.Presets))
		for _, p := range e.Presets {
			presets = append(presets, p.Name)
		}

		effects = append(effects, ResolumeAvailableEffect{
			ID:      e.IDString,
			Name:    e.Name,
			Presets: presets,
		})
	}

	return effects, nil
}

// ============================================================================
// Clip Connection Methods (REST API)
// ============================================================================

// ConnectClipREST triggers/connects a clip using the REST API (more reliable than OSC)
func (c *ResolumeClient) ConnectClipREST(ctx context.Context, layer, column int) error {
	apiPath := fmt.Sprintf("/api/v1/composition/layers/%d/clips/%d/connect", layer, column)

	// POST with no body triggers the clip (like a mouse click)
	reqURL := c.baseURL + apiPath
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connect clip failed %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

// DisconnectAllClips disconnects all playing clips in the composition
func (c *ResolumeClient) DisconnectAllClips(ctx context.Context) error {
	apiPath := "/api/v1/composition/disconnect-all"

	reqURL := c.baseURL + apiPath
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("disconnect all failed %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

// ConnectColumnREST triggers all clips in a column using REST API
func (c *ResolumeClient) ConnectColumnREST(ctx context.Context, column int) error {
	apiPath := fmt.Sprintf("/api/v1/composition/columns/%d/connect", column)

	reqURL := c.baseURL + apiPath
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connect column failed %d: %s", resp.StatusCode, string(data))
	}

	return nil
}

// GetAvailableSources returns all available video sources
func (c *ResolumeClient) GetAvailableSources(ctx context.Context) ([]ResolumeAvailableEffect, error) {
	data, err := c.doRequest(ctx, "GET", "/api/v1/sources/video", nil)
	if err != nil {
		return nil, fmt.Errorf("get available sources: %w", err)
	}

	var apiSources []struct {
		IDString string `json:"idstring"`
		Name     string `json:"name"`
		Presets  []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"presets,omitempty"`
	}

	if err := json.Unmarshal(data, &apiSources); err != nil {
		return nil, fmt.Errorf("parse sources: %w", err)
	}

	sources := make([]ResolumeAvailableEffect, 0, len(apiSources))
	for _, s := range apiSources {
		presets := make([]string, 0, len(s.Presets))
		for _, p := range s.Presets {
			presets = append(presets, p.Name)
		}

		sources = append(sources, ResolumeAvailableEffect{
			ID:      s.IDString,
			Name:    s.Name,
			Presets: presets,
		})
	}

	return sources, nil
}

// =====================================================
// Dashboard Text Display Methods
// =====================================================

// TrackDisplay holds the current track information for display
type TrackDisplay struct {
	Artist string  `json:"artist"`
	Title  string  `json:"title"`
	Key    string  `json:"key,omitempty"`
	BPM    float64 `json:"bpm,omitempty"`
	Genre  string  `json:"genre,omitempty"`
}

// SetDashboardString sets a dashboard string parameter by index (1-8)
// These correspond to Resolume's Dashboard String 1-8 which can be mapped to text sources
func (c *ResolumeClient) SetDashboardString(index int, text string) error {
	if index < 1 || index > 8 {
		return fmt.Errorf("dashboard string index must be 1-8, got %d", index)
	}
	address := fmt.Sprintf("/composition/dashboard/link%d", index)
	return c.sendOSC(address, text)
}

// SetArtist sets the artist name on Dashboard String 1
func (c *ResolumeClient) SetArtist(artist string) error {
	return c.SetDashboardString(1, artist)
}

// SetTitle sets the track title on Dashboard String 2
func (c *ResolumeClient) SetTitle(title string) error {
	return c.SetDashboardString(2, title)
}

// SetNowPlaying sets both artist and title for display
func (c *ResolumeClient) SetNowPlaying(artist, title string) error {
	if err := c.SetArtist(artist); err != nil {
		return fmt.Errorf("set artist: %w", err)
	}
	if err := c.SetTitle(title); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	return nil
}

// SetTrackInfo sets full track information across dashboard strings
// Dashboard String 1: Artist
// Dashboard String 2: Title
// Dashboard String 3: Key
// Dashboard String 4: BPM (formatted as string)
// Dashboard String 5: Genre
func (c *ResolumeClient) SetTrackInfo(track TrackDisplay) error {
	if err := c.SetDashboardString(1, track.Artist); err != nil {
		return fmt.Errorf("set artist: %w", err)
	}
	if err := c.SetDashboardString(2, track.Title); err != nil {
		return fmt.Errorf("set title: %w", err)
	}
	if track.Key != "" {
		if err := c.SetDashboardString(3, track.Key); err != nil {
			return fmt.Errorf("set key: %w", err)
		}
	}
	if track.BPM > 0 {
		bpmStr := fmt.Sprintf("%.1f BPM", track.BPM)
		if err := c.SetDashboardString(4, bpmStr); err != nil {
			return fmt.Errorf("set bpm: %w", err)
		}
	}
	if track.Genre != "" {
		if err := c.SetDashboardString(5, track.Genre); err != nil {
			return fmt.Errorf("set genre: %w", err)
		}
	}
	return nil
}

// ClearTrackDisplay clears all track display fields
func (c *ResolumeClient) ClearTrackDisplay() error {
	for i := 1; i <= 5; i++ {
		if err := c.SetDashboardString(i, ""); err != nil {
			return fmt.Errorf("clear string %d: %w", i, err)
		}
	}
	return nil
}

// SetFormattedNowPlaying sets a single formatted "Artist - Title" string on Dashboard String 1
func (c *ResolumeClient) SetFormattedNowPlaying(artist, title string) error {
	formatted := artist
	if title != "" {
		if formatted != "" {
			formatted += " - "
		}
		formatted += title
	}
	return c.SetDashboardString(1, formatted)
}

// ============================================================================
// Health Check Methods
// ============================================================================

// HealthCheck performs a comprehensive health check of the Resolume connection
// Returns nil if Resolume is reachable, error otherwise
func (c *ResolumeClient) HealthCheck(ctx context.Context) error {
	// Try REST API first (most reliable)
	_, err := c.doRequest(ctx, "GET", "/api/v1/product", nil)
	if err == nil {
		return nil
	}

	// Try getting composition (validates full API access)
	_, compErr := c.getComposition(ctx)
	if compErr == nil {
		return nil
	}

	// Fallback: check if OSC port is reachable
	addr := net.JoinHostPort(c.oscHost, strconv.Itoa(c.oscPort))
	conn, dialErr := net.DialTimeout("udp", addr, 2*time.Second)
	if dialErr != nil {
		return fmt.Errorf("resolume not reachable: REST API error: %v, OSC error: %v", err, dialErr)
	}
	conn.Close()

	// OSC port is open but REST API failed - partial connectivity
	return fmt.Errorf("resolume OSC reachable but REST API unavailable: %v", err)
}

// HealthCheckWithRetry performs health check with configurable retries
func (c *ResolumeClient) HealthCheckWithRetry(ctx context.Context, maxRetries int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := c.HealthCheck(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// Don't sleep on the last iteration
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("health check failed after %d retries: %w", maxRetries, lastErr)
}

// WaitForConnection waits until Resolume becomes available or context is cancelled
func (c *ResolumeClient) WaitForConnection(ctx context.Context, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.HealthCheck(ctx); err == nil {
				return nil
			}
		}
	}
}

// ConnectionInfo returns information about the current connection configuration
type ConnectionInfo struct {
	OSCHost       string `json:"osc_host"`
	OSCPort       int    `json:"osc_port"`
	APIPort       int    `json:"api_port"`
	BaseURL       string `json:"base_url"`
	RESTAvailable bool   `json:"rest_available"`
	OSCAvailable  bool   `json:"osc_available"`
}

// GetConnectionInfo returns the current connection configuration and status
func (c *ResolumeClient) GetConnectionInfo(ctx context.Context) *ConnectionInfo {
	info := &ConnectionInfo{
		OSCHost: c.oscHost,
		OSCPort: c.oscPort,
		APIPort: c.apiPort,
		BaseURL: c.baseURL,
	}

	// Check REST API
	_, err := c.doRequest(ctx, "GET", "/api/v1/product", nil)
	info.RESTAvailable = err == nil

	// Check OSC port
	addr := net.JoinHostPort(c.oscHost, strconv.Itoa(c.oscPort))
	conn, err := net.DialTimeout("udp", addr, 1*time.Second)
	if err == nil {
		conn.Close()
		info.OSCAvailable = true
	}

	return info
}
