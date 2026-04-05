// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// WLEDClient manages WLED LED controller devices
type WLEDClient struct {
	mu          sync.RWMutex
	devices     map[string]*WLEDDevice
	httpClient  *http.Client
	scanTimeout time.Duration
}

// WLEDDevice represents a WLED controller
type WLEDDevice struct {
	IP         string                 `json:"ip"`
	Name       string                 `json:"name"`
	MAC        string                 `json:"mac,omitempty"`
	Version    string                 `json:"version,omitempty"`
	LEDCount   int                    `json:"led_count"`
	On         bool                   `json:"on"`
	Brightness int                    `json:"brightness"` // 0-255
	Effect     int                    `json:"effect"`
	EffectName string                 `json:"effect_name,omitempty"`
	Palette    int                    `json:"palette"`
	Color      [3]int                 `json:"color"` // RGB
	Segments   []WLEDSegment          `json:"segments,omitempty"`
	ArtNet     *WLEDArtNetConfig      `json:"artnet,omitempty"`
	LastSeen   time.Time              `json:"last_seen"`
	Online     bool                   `json:"online"`
	Info       map[string]interface{} `json:"info,omitempty"`
}

// WLEDSegment represents a segment of LEDs
type WLEDSegment struct {
	ID         int    `json:"id"`
	Start      int    `json:"start"`
	Stop       int    `json:"stop"`
	Length     int    `json:"len"`
	On         bool   `json:"on"`
	Brightness int    `json:"bri"`
	Effect     int    `json:"fx"`
	Palette    int    `json:"pal"`
	Color      [3]int `json:"col"`
}

// WLEDArtNetConfig represents Art-Net configuration
type WLEDArtNetConfig struct {
	Enabled   bool   `json:"enabled"`
	Universe  int    `json:"universe"`
	Mode      string `json:"mode"` // disabled, single, multi
	Timeout   int    `json:"timeout"`
	StartAddr int    `json:"start_address"`
}

// WLEDEffect represents an available effect
type WLEDEffect struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WLEDPalette represents an available color palette
type WLEDPalette struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NewWLEDClient creates a new WLED client
func NewWLEDClient() (*WLEDClient, error) {
	return &WLEDClient{
		devices: make(map[string]*WLEDDevice),
		httpClient: httpclient.Fast(),
		scanTimeout: 3 * time.Second,
	}, nil
}

// DiscoverDevices scans the network for WLED devices
func (c *WLEDClient) DiscoverDevices(ctx context.Context, subnet string) ([]*WLEDDevice, error) {
	if subnet == "" {
		subnet = "192.168.1.0/24"
	}

	// Parse subnet
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	var devices []*WLEDDevice
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Get all IPs in subnet
	ips := expandCIDR(ipNet)

	// Limit concurrency
	sem := make(chan struct{}, 50)

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			device, err := c.probeDevice(ctx, ipAddr)
			if err == nil && device != nil {
				mu.Lock()
				devices = append(devices, device)
				c.devices[ipAddr] = device
				mu.Unlock()
			}
		}(ip)
	}

	wg.Wait()

	return devices, nil
}

// probeDevice checks if an IP is a WLED device
func (c *WLEDClient) probeDevice(ctx context.Context, ip string) (*WLEDDevice, error) {
	url := fmt.Sprintf("http://%s/json/info", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("not a WLED device")
	}

	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	// Check if it's WLED
	if _, ok := info["ver"]; !ok {
		return nil, fmt.Errorf("not a WLED device")
	}

	device := &WLEDDevice{
		IP:       ip,
		Online:   true,
		LastSeen: time.Now(),
		Info:     info,
	}

	// Extract basic info
	if ver, ok := info["ver"].(string); ok {
		device.Version = ver
	}
	if name, ok := info["name"].(string); ok {
		device.Name = name
	}
	if mac, ok := info["mac"].(string); ok {
		device.MAC = mac
	}
	if leds, ok := info["leds"].(map[string]interface{}); ok {
		if count, ok := leds["count"].(float64); ok {
			device.LEDCount = int(count)
		}
	}

	// Get current state
	c.refreshDeviceState(ctx, device)

	return device, nil
}

// refreshDeviceState gets current device state
func (c *WLEDClient) refreshDeviceState(ctx context.Context, device *WLEDDevice) error {
	url := fmt.Sprintf("http://%s/json/state", device.IP)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var state map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return err
	}

	if on, ok := state["on"].(bool); ok {
		device.On = on
	}
	if bri, ok := state["bri"].(float64); ok {
		device.Brightness = int(bri)
	}

	// Get segments
	if segs, ok := state["seg"].([]interface{}); ok {
		device.Segments = make([]WLEDSegment, 0)
		for i, seg := range segs {
			if segMap, ok := seg.(map[string]interface{}); ok {
				segment := WLEDSegment{ID: i}
				if start, ok := segMap["start"].(float64); ok {
					segment.Start = int(start)
				}
				if stop, ok := segMap["stop"].(float64); ok {
					segment.Stop = int(stop)
				}
				if on, ok := segMap["on"].(bool); ok {
					segment.On = on
				}
				if bri, ok := segMap["bri"].(float64); ok {
					segment.Brightness = int(bri)
				}
				if fx, ok := segMap["fx"].(float64); ok {
					segment.Effect = int(fx)
				}
				segment.Length = segment.Stop - segment.Start
				device.Segments = append(device.Segments, segment)
			}
		}
	}

	return nil
}

// GetDevice returns a device by IP
func (c *WLEDClient) GetDevice(ctx context.Context, ip string) (*WLEDDevice, error) {
	c.mu.RLock()
	device, exists := c.devices[ip]
	c.mu.RUnlock()

	if !exists {
		// Try to probe it
		var err error
		device, err = c.probeDevice(ctx, ip)
		if err != nil {
			return nil, fmt.Errorf("device not found: %s", ip)
		}
		c.mu.Lock()
		c.devices[ip] = device
		c.mu.Unlock()
	}

	// Refresh state
	c.refreshDeviceState(ctx, device)

	return device, nil
}

// ListDevices returns all known devices
func (c *WLEDClient) ListDevices(ctx context.Context) []*WLEDDevice {
	c.mu.RLock()
	defer c.mu.RUnlock()

	devices := make([]*WLEDDevice, 0, len(c.devices))
	for _, d := range c.devices {
		devices = append(devices, d)
	}
	return devices
}

// SetPower turns device on or off
func (c *WLEDClient) SetPower(ctx context.Context, ip string, on bool) error {
	return c.sendState(ctx, ip, map[string]interface{}{"on": on})
}

// SetBrightness sets device brightness (0-255)
func (c *WLEDClient) SetBrightness(ctx context.Context, ip string, brightness int) error {
	if brightness < 0 {
		brightness = 0
	}
	if brightness > 255 {
		brightness = 255
	}
	return c.sendState(ctx, ip, map[string]interface{}{"bri": brightness})
}

// SetEffect sets the current effect
func (c *WLEDClient) SetEffect(ctx context.Context, ip string, effectID int) error {
	return c.sendState(ctx, ip, map[string]interface{}{
		"seg": []map[string]interface{}{{"fx": effectID}},
	})
}

// SetColor sets the primary color
func (c *WLEDClient) SetColor(ctx context.Context, ip string, r, g, b int) error {
	return c.sendState(ctx, ip, map[string]interface{}{
		"seg": []map[string]interface{}{
			{"col": [][]int{{r, g, b}}},
		},
	})
}

// SetPalette sets the color palette
func (c *WLEDClient) SetPalette(ctx context.Context, ip string, paletteID int) error {
	return c.sendState(ctx, ip, map[string]interface{}{
		"seg": []map[string]interface{}{{"pal": paletteID}},
	})
}

// sendState sends a state update to the device
func (c *WLEDClient) sendState(ctx context.Context, ip string, state map[string]interface{}) error {
	url := fmt.Sprintf("http://%s/json/state", ip)

	body, err := json.Marshal(state)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set state: HTTP %d", resp.StatusCode)
	}

	return nil
}

// GetEffects returns available effects
func (c *WLEDClient) GetEffects(ctx context.Context, ip string) ([]WLEDEffect, error) {
	url := fmt.Sprintf("http://%s/json/effects", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var effectNames []string
	if err := json.NewDecoder(resp.Body).Decode(&effectNames); err != nil {
		return nil, err
	}

	effects := make([]WLEDEffect, len(effectNames))
	for i, name := range effectNames {
		effects[i] = WLEDEffect{ID: i, Name: name}
	}

	return effects, nil
}

// GetPalettes returns available palettes
func (c *WLEDClient) GetPalettes(ctx context.Context, ip string) ([]WLEDPalette, error) {
	url := fmt.Sprintf("http://%s/json/palettes", ip)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var paletteNames []string
	if err := json.NewDecoder(resp.Body).Decode(&paletteNames); err != nil {
		return nil, err
	}

	palettes := make([]WLEDPalette, len(paletteNames))
	for i, name := range paletteNames {
		palettes[i] = WLEDPalette{ID: i, Name: name}
	}

	return palettes, nil
}

// ConfigureArtNet configures Art-Net settings
func (c *WLEDClient) ConfigureArtNet(ctx context.Context, ip string, config *WLEDArtNetConfig) error {
	// Art-Net config requires setting via the cfg endpoint
	cfg := map[string]interface{}{
		"if": map[string]interface{}{
			"live": map[string]interface{}{
				"en":   config.Enabled,
				"mgrp": 1,
			},
		},
	}

	if config.Enabled {
		cfg["if"].(map[string]interface{})["live"].(map[string]interface{})["port"] = 5568
		cfg["if"].(map[string]interface{})["live"].(map[string]interface{})["dmx"] = map[string]interface{}{
			"uni":  config.Universe,
			"addr": config.StartAddr,
			"mode": 0, // Art-Net mode
		}
	}

	url := fmt.Sprintf("http://%s/json/cfg", ip)
	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SavePreset saves current state as a preset
func (c *WLEDClient) SavePreset(ctx context.Context, ip string, presetID int, name string) error {
	state := map[string]interface{}{
		"psave": presetID,
		"n":     name,
	}
	return c.sendState(ctx, ip, state)
}

// LoadPreset loads a saved preset
func (c *WLEDClient) LoadPreset(ctx context.Context, ip string, presetID int) error {
	return c.sendState(ctx, ip, map[string]interface{}{"ps": presetID})
}

// expandCIDR expands a CIDR to list of IPs
func expandCIDR(ipNet *net.IPNet) []string {
	var ips []string
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}
	// Remove network and broadcast addresses
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}
	return ips
}

// incrementIP increments an IP address
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
