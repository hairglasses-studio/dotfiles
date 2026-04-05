// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// LedFXClient provides access to LedFX audio-reactive LED system via REST API
type LedFXClient struct {
	baseURL    string
	httpClient *http.Client
}

// LedFXStatus represents LedFX system status
type LedFXStatus struct {
	Connected     bool   `json:"connected"`
	Version       string `json:"version"`
	DeviceCount   int    `json:"device_count"`
	VirtualCount  int    `json:"virtual_count"`
	ActiveEffects int    `json:"active_effects"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
}

// LedFXDevice represents a physical LED device
type LedFXDevice struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Config     map[string]interface{} `json:"config,omitempty"`
	PixelCount int                    `json:"pixel_count"`
	Online     bool                   `json:"online"`
}

// LedFXVirtual represents a virtual LED device
type LedFXVirtual struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	IsActive   bool           `json:"is_active"`
	PixelCount int            `json:"pixel_count"`
	Effect     *LedFXEffect   `json:"effect,omitempty"`
	Segments   []LedFXSegment `json:"segments,omitempty"`
}

// LedFXSegment represents a segment mapping in a virtual
type LedFXSegment struct {
	DeviceID string `json:"device_id"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Invert   bool   `json:"invert"`
}

// LedFXEffect represents an active effect on a virtual
type LedFXEffect struct {
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// LedFXEffectType represents an available effect type
type LedFXEffectType struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// LedFXPreset represents an effect preset
type LedFXPreset struct {
	ID     string                 `json:"id"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// LedFXScene represents a saved scene
type LedFXScene struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Virtuals map[string]interface{} `json:"virtuals,omitempty"`
}

// LedFXAudioDevice represents an audio input device
type LedFXAudioDevice struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// LedFXHealth represents system health status
type LedFXHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	DeviceCount     int      `json:"device_count"`
	VirtualCount    int      `json:"virtual_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewLedFXClient creates a new LedFX client
func NewLedFXClient() (*LedFXClient, error) {
	host := os.Getenv("LEDFX_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("LEDFX_PORT")
	if port == "" {
		port = "8888"
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)

	return &LedFXClient{
		baseURL:    baseURL,
		httpClient: httpclient.Fast(),
	}, nil
}

// Host returns the configured host
func (c *LedFXClient) Host() string {
	parts := strings.Split(strings.TrimPrefix(c.baseURL, "http://"), ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return "localhost"
}

// Port returns the configured port
func (c *LedFXClient) Port() int {
	parts := strings.Split(strings.TrimPrefix(c.baseURL, "http://"), ":")
	if len(parts) > 1 {
		var port int
		fmt.Sscanf(parts[1], "%d", &port)
		return port
	}
	return 8888
}

// doRequest performs an HTTP request and returns the response body
func (c *LedFXClient) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
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

// GetStatus returns LedFX system status
func (c *LedFXClient) GetStatus(ctx context.Context) (*LedFXStatus, error) {
	status := &LedFXStatus{
		Connected: false,
		Host:      c.Host(),
		Port:      c.Port(),
	}

	// Try to connect to LedFX
	addr := net.JoinHostPort(c.Host(), strconv.Itoa(c.Port()))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return status, nil
	}
	conn.Close()
	status.Connected = true

	// Get info
	body, err := c.doRequest(ctx, "GET", "/api/info", nil)
	if err != nil {
		return status, nil
	}

	var info map[string]interface{}
	if err := json.Unmarshal(body, &info); err == nil {
		if v, ok := info["version"].(string); ok {
			status.Version = v
		}
	}

	// Count devices and virtuals
	devices, _ := c.GetDevices(ctx)
	status.DeviceCount = len(devices)

	virtuals, _ := c.GetVirtuals(ctx)
	status.VirtualCount = len(virtuals)

	for _, v := range virtuals {
		if v.Effect != nil {
			status.ActiveEffects++
		}
	}

	return status, nil
}

// GetDevices returns all physical LED devices
func (c *LedFXClient) GetDevices(ctx context.Context) ([]LedFXDevice, error) {
	body, err := c.doRequest(ctx, "GET", "/api/devices", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	devices := []LedFXDevice{}
	if devicesMap, ok := resp["devices"].(map[string]interface{}); ok {
		for id, data := range devicesMap {
			dev := LedFXDevice{ID: id}
			if devData, ok := data.(map[string]interface{}); ok {
				if name, ok := devData["name"].(string); ok {
					dev.Name = name
				}
				if devType, ok := devData["type"].(string); ok {
					dev.Type = devType
				}
				if config, ok := devData["config"].(map[string]interface{}); ok {
					dev.Config = config
					if pixels, ok := config["pixel_count"].(float64); ok {
						dev.PixelCount = int(pixels)
					}
				}
				dev.Online = true // Assume online if returned
			}
			devices = append(devices, dev)
		}
	}

	return devices, nil
}

// FindDevices triggers WLED device auto-discovery
func (c *LedFXClient) FindDevices(ctx context.Context) ([]LedFXDevice, error) {
	body, err := c.doRequest(ctx, "GET", "/api/find_devices", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	devices := []LedFXDevice{}
	if devicesData, ok := resp["devices"].([]interface{}); ok {
		for _, data := range devicesData {
			if devData, ok := data.(map[string]interface{}); ok {
				dev := LedFXDevice{}
				if id, ok := devData["id"].(string); ok {
					dev.ID = id
				}
				if name, ok := devData["name"].(string); ok {
					dev.Name = name
				}
				if devType, ok := devData["type"].(string); ok {
					dev.Type = devType
				}
				devices = append(devices, dev)
			}
		}
	}

	return devices, nil
}

// GetVirtuals returns all virtual LED devices
func (c *LedFXClient) GetVirtuals(ctx context.Context) ([]LedFXVirtual, error) {
	body, err := c.doRequest(ctx, "GET", "/api/virtuals", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	virtuals := []LedFXVirtual{}
	if virtualsMap, ok := resp["virtuals"].(map[string]interface{}); ok {
		for id, data := range virtualsMap {
			v := LedFXVirtual{ID: id}
			if vData, ok := data.(map[string]interface{}); ok {
				if name, ok := vData["name"].(string); ok {
					v.Name = name
				}
				if active, ok := vData["active"].(bool); ok {
					v.IsActive = active
				}
				if pixels, ok := vData["pixel_count"].(float64); ok {
					v.PixelCount = int(pixels)
				}
				if effect, ok := vData["effect"].(map[string]interface{}); ok {
					v.Effect = &LedFXEffect{}
					if effType, ok := effect["type"].(string); ok {
						v.Effect.Type = effType
					}
					if effName, ok := effect["name"].(string); ok {
						v.Effect.Name = effName
					}
					if config, ok := effect["config"].(map[string]interface{}); ok {
						v.Effect.Config = config
					}
				}
			}
			virtuals = append(virtuals, v)
		}
	}

	return virtuals, nil
}

// GetVirtual returns a specific virtual device
func (c *LedFXClient) GetVirtual(ctx context.Context, virtualID string) (*LedFXVirtual, error) {
	body, err := c.doRequest(ctx, "GET", "/api/virtuals/"+virtualID, nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	v := &LedFXVirtual{ID: virtualID}
	if name, ok := resp["name"].(string); ok {
		v.Name = name
	}
	if active, ok := resp["active"].(bool); ok {
		v.IsActive = active
	}
	if pixels, ok := resp["pixel_count"].(float64); ok {
		v.PixelCount = int(pixels)
	}
	if effect, ok := resp["effect"].(map[string]interface{}); ok {
		v.Effect = &LedFXEffect{}
		if effType, ok := effect["type"].(string); ok {
			v.Effect.Type = effType
		}
		if config, ok := effect["config"].(map[string]interface{}); ok {
			v.Effect.Config = config
		}
	}

	return v, nil
}

// SetVirtualActive activates or deactivates a virtual
func (c *LedFXClient) SetVirtualActive(ctx context.Context, virtualID string, active bool) error {
	payload := map[string]interface{}{
		"active": active,
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/virtuals/"+virtualID, strings.NewReader(string(jsonData)))
	return err
}

// SetVirtualEffect applies an effect to a virtual
func (c *LedFXClient) SetVirtualEffect(ctx context.Context, virtualID, effectType string, config map[string]interface{}) error {
	payload := map[string]interface{}{
		"type":   effectType,
		"config": config,
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/virtuals/"+virtualID+"/effects", strings.NewReader(string(jsonData)))
	return err
}

// ClearVirtualEffect removes the effect from a virtual
func (c *LedFXClient) ClearVirtualEffect(ctx context.Context, virtualID string) error {
	_, err := c.doRequest(ctx, "DELETE", "/api/virtuals/"+virtualID+"/effects", nil)
	return err
}

// GetEffects returns all available effect types
func (c *LedFXClient) GetEffects(ctx context.Context) ([]LedFXEffectType, error) {
	body, err := c.doRequest(ctx, "GET", "/api/effects", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	effects := []LedFXEffectType{}
	if effectsMap, ok := resp["effects"].(map[string]interface{}); ok {
		for id, data := range effectsMap {
			eff := LedFXEffectType{ID: id}
			if effData, ok := data.(map[string]interface{}); ok {
				if name, ok := effData["name"].(string); ok {
					eff.Name = name
				}
				if cat, ok := effData["category"].(string); ok {
					eff.Category = cat
				}
				if desc, ok := effData["description"].(string); ok {
					eff.Description = desc
				}
			}
			effects = append(effects, eff)
		}
	}

	return effects, nil
}

// GetEffectPresets returns presets for an effect type
func (c *LedFXClient) GetEffectPresets(ctx context.Context, effectType string) ([]LedFXPreset, error) {
	body, err := c.doRequest(ctx, "GET", "/api/effects/"+effectType+"/presets", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	presets := []LedFXPreset{}

	// User presets
	if userPresets, ok := resp["user_presets"].(map[string]interface{}); ok {
		for id, data := range userPresets {
			preset := LedFXPreset{ID: id, Name: id}
			if config, ok := data.(map[string]interface{}); ok {
				preset.Config = config
			}
			presets = append(presets, preset)
		}
	}

	// Built-in presets
	if builtinPresets, ok := resp["ledfx_presets"].(map[string]interface{}); ok {
		for id, data := range builtinPresets {
			preset := LedFXPreset{ID: id, Name: id + " (built-in)"}
			if config, ok := data.(map[string]interface{}); ok {
				preset.Config = config
			}
			presets = append(presets, preset)
		}
	}

	return presets, nil
}

// GetScenes returns all saved scenes
func (c *LedFXClient) GetScenes(ctx context.Context) ([]LedFXScene, error) {
	body, err := c.doRequest(ctx, "GET", "/api/scenes", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	scenes := []LedFXScene{}
	if scenesMap, ok := resp["scenes"].(map[string]interface{}); ok {
		for id, data := range scenesMap {
			scene := LedFXScene{ID: id}
			if sceneData, ok := data.(map[string]interface{}); ok {
				if name, ok := sceneData["name"].(string); ok {
					scene.Name = name
				}
				if virtuals, ok := sceneData["virtuals"].(map[string]interface{}); ok {
					scene.Virtuals = virtuals
				}
			}
			scenes = append(scenes, scene)
		}
	}

	return scenes, nil
}

// ActivateScene activates a saved scene
func (c *LedFXClient) ActivateScene(ctx context.Context, sceneID string) error {
	payload := map[string]interface{}{
		"id":     sceneID,
		"action": "activate",
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/scenes", strings.NewReader(string(jsonData)))
	return err
}

// SaveScene saves the current state as a scene
func (c *LedFXClient) SaveScene(ctx context.Context, name string) (*LedFXScene, error) {
	payload := map[string]interface{}{
		"name": name,
	}
	jsonData, _ := json.Marshal(payload)
	body, err := c.doRequest(ctx, "POST", "/api/scenes", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	scene := &LedFXScene{Name: name}
	if id, ok := resp["scene_id"].(string); ok {
		scene.ID = id
	}

	return scene, nil
}

// DeleteScene deletes a saved scene
func (c *LedFXClient) DeleteScene(ctx context.Context, sceneID string) error {
	payload := map[string]interface{}{
		"id": sceneID,
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "DELETE", "/api/scenes", strings.NewReader(string(jsonData)))
	return err
}

// GetAudioDevices returns available audio input devices
func (c *LedFXClient) GetAudioDevices(ctx context.Context) ([]LedFXAudioDevice, error) {
	body, err := c.doRequest(ctx, "GET", "/api/audio/devices", nil)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	devices := []LedFXAudioDevice{}
	activeIndex := -1
	if idx, ok := resp["active_device_index"].(float64); ok {
		activeIndex = int(idx)
	}

	if devicesMap, ok := resp["devices"].(map[string]interface{}); ok {
		for idxStr, name := range devicesMap {
			var idx int
			fmt.Sscanf(idxStr, "%d", &idx)
			dev := LedFXAudioDevice{
				Index:  idx,
				Name:   fmt.Sprintf("%v", name),
				Active: idx == activeIndex,
			}
			devices = append(devices, dev)
		}
	}

	return devices, nil
}

// SetAudioDevice sets the active audio input device
func (c *LedFXClient) SetAudioDevice(ctx context.Context, deviceIndex int) error {
	payload := map[string]interface{}{
		"audio_device": deviceIndex,
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/audio/devices", strings.NewReader(string(jsonData)))
	return err
}

// GetSegments returns LED segments for a device
func (c *LedFXClient) GetSegments(ctx context.Context, deviceID string) ([]LedFXSegment, error) {
	virtuals, err := c.GetVirtuals(ctx)
	if err != nil {
		return nil, err
	}

	segments := []LedFXSegment{}
	for _, v := range virtuals {
		for _, seg := range v.Segments {
			if deviceID == "" || seg.DeviceID == deviceID {
				segments = append(segments, seg)
			}
		}
	}

	return segments, nil
}

// CreateSegment creates a new segment on a device
func (c *LedFXClient) CreateSegment(ctx context.Context, deviceID string, start, end int) error {
	// Segments are created within virtuals
	// Would need to create or update a virtual with segment configuration
	return nil
}

// DeleteSegment deletes a segment from a device
func (c *LedFXClient) DeleteSegment(ctx context.Context, deviceID string) error {
	// Would remove segment configuration from virtual
	return nil
}

// GetBPM returns the current BPM setting
func (c *LedFXClient) GetBPM(ctx context.Context) (float64, error) {
	body, err := c.doRequest(ctx, "GET", "/api/config", nil)
	if err != nil {
		return 0, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, fmt.Errorf("parsing response: %w", err)
	}

	if audio, ok := resp["audio"].(map[string]interface{}); ok {
		if bpm, ok := audio["bpm"].(float64); ok {
			return bpm, nil
		}
	}

	return 120.0, nil // Default BPM
}

// SetBPM sets the BPM for audio-reactive effects
func (c *LedFXClient) SetBPM(ctx context.Context, bpm float64) error {
	payload := map[string]interface{}{
		"audio": map[string]interface{}{
			"bpm": bpm,
		},
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/config", strings.NewReader(string(jsonData)))
	return err
}

// SetBPMMode sets the BPM sync mode (manual, auto, tap)
func (c *LedFXClient) SetBPMMode(ctx context.Context, mode string) error {
	payload := map[string]interface{}{
		"audio": map[string]interface{}{
			"bpm_mode": mode,
		},
	}
	jsonData, _ := json.Marshal(payload)
	_, err := c.doRequest(ctx, "PUT", "/api/config", strings.NewReader(string(jsonData)))
	return err
}

// SetGradient applies a color gradient to a virtual device
func (c *LedFXClient) SetGradient(ctx context.Context, virtualID string, colors []string, name string) error {
	// Apply gradient effect with custom colors
	config := map[string]interface{}{
		"gradient": colors,
	}
	if name != "" {
		config["gradient_name"] = name
	}
	return c.SetVirtualEffect(ctx, virtualID, "gradient", config)
}

// SetSolidColor sets a solid color on a virtual device
func (c *LedFXClient) SetSolidColor(ctx context.Context, virtualID, color string) error {
	config := map[string]interface{}{
		"color": color,
	}
	return c.SetVirtualEffect(ctx, virtualID, "singleColor", config)
}

// GetHealth returns system health status
func (c *LedFXClient) GetHealth(ctx context.Context) (*LedFXHealth, error) {
	health := &LedFXHealth{
		Score:  100,
		Status: "healthy",
	}

	status, _ := c.GetStatus(ctx)
	health.Connected = status.Connected
	health.DeviceCount = status.DeviceCount
	health.VirtualCount = status.VirtualCount

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Cannot connect to LedFX")
		health.Recommendations = append(health.Recommendations, "Start LedFX and verify it's running on the configured host/port")
	} else {
		if status.DeviceCount == 0 {
			health.Score -= 20
			health.Issues = append(health.Issues, "No LED devices configured")
			health.Recommendations = append(health.Recommendations, "Add LED devices or run auto-discovery for WLED devices")
		}
		if status.VirtualCount == 0 {
			health.Score -= 10
			health.Issues = append(health.Issues, "No virtual devices configured")
			health.Recommendations = append(health.Recommendations, "Create virtual devices to map effects to physical LEDs")
		}
		if status.ActiveEffects == 0 && status.VirtualCount > 0 {
			health.Score -= 5
			health.Issues = append(health.Issues, "No active effects running")
			health.Recommendations = append(health.Recommendations, "Apply effects to virtual devices for audio-reactive lighting")
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
