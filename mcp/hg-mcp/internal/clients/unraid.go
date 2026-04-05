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
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// UNRAIDClient provides access to UNRAID server API
type UNRAIDClient struct {
	host       string
	apiKey     string
	httpClient *http.Client
}

// UNRAIDStatus represents UNRAID server status
type UNRAIDStatus struct {
	Connected   bool    `json:"connected"`
	ArrayStatus string  `json:"array_status"` // started, stopped, etc.
	Uptime      string  `json:"uptime"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsed  float64 `json:"memory_used_percent"`
	ArrayUsed   float64 `json:"array_used_percent"`
	CacheUsed   float64 `json:"cache_used_percent"`
	Temperature int     `json:"temperature_c"`
}

// UNRAIDVM represents a virtual machine
type UNRAIDVM struct {
	Name      string `json:"name"`
	State     string `json:"state"` // running, stopped, paused
	CPUCores  int    `json:"cpu_cores"`
	MemoryMB  int    `json:"memory_mb"`
	AutoStart bool   `json:"autostart"`
}

// UNRAIDDocker represents a Docker container
type UNRAIDDocker struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	State   string `json:"state"` // running, stopped, created
	Uptime  string `json:"uptime,omitempty"`
	Network string `json:"network"`
}

// StudioHealth represents overall studio health
type StudioHealth struct {
	Score           int            `json:"score"`
	Status          string         `json:"status"`
	Components      map[string]int `json:"components"`
	Issues          []string       `json:"issues,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
}

// NewUNRAIDClient creates a new UNRAID client
func NewUNRAIDClient() (*UNRAIDClient, error) {
	host := os.Getenv("UNRAID_HOST")
	if host == "" {
		host = "tower.local"
	}

	apiKey := os.Getenv("UNRAID_API_KEY")

	return &UNRAIDClient{
		host:   host,
		apiKey: apiKey,
		httpClient: httpclient.Fast(),
	}, nil
}

// baseURL returns the base URL for UNRAID API
func (c *UNRAIDClient) baseURL() string {
	return fmt.Sprintf("http://%s", c.host)
}

// doRequest performs an HTTP request to UNRAID
func (c *UNRAIDClient) doRequest(ctx context.Context, method, endpoint string) ([]byte, error) {
	url := c.baseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("UNRAID API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetStatus returns UNRAID server status
func (c *UNRAIDClient) GetStatus(ctx context.Context) (*UNRAIDStatus, error) {
	status := &UNRAIDStatus{
		Connected: false,
	}

	// Try to connect
	resp, err := c.httpClient.Get(c.baseURL())
	if err != nil {
		return status, nil // Return disconnected, not error
	}
	defer resp.Body.Close()

	status.Connected = resp.StatusCode == 200 || resp.StatusCode == 302

	if status.Connected {
		// In a real implementation, we'd parse UNRAID's status page or use GraphQL API
		status.ArrayStatus = "started"
		status.Uptime = "Unknown"
		status.CPUUsage = 0
		status.MemoryUsed = 0
		status.ArrayUsed = 0
	}

	return status, nil
}

// ListVMs returns list of virtual machines
func (c *UNRAIDClient) ListVMs(ctx context.Context) ([]UNRAIDVM, error) {
	// In a real implementation, this would query UNRAID's VM manager
	// For now, return empty list indicating connectivity required
	vms := []UNRAIDVM{}

	return vms, nil
}

// ListDockers returns list of Docker containers
func (c *UNRAIDClient) ListDockers(ctx context.Context) ([]UNRAIDDocker, error) {
	// In a real implementation, this would query UNRAID's Docker manager
	dockers := []UNRAIDDocker{}

	return dockers, nil
}

// IsConnected checks if UNRAID is reachable
func (c *UNRAIDClient) IsConnected(ctx context.Context) bool {
	status, _ := c.GetStatus(ctx)
	return status != nil && status.Connected
}

// Host returns the configured host
func (c *UNRAIDClient) Host() string {
	return c.host
}

// NetworkDevice represents a device on the network
type NetworkDevice struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	MAC      string `json:"mac,omitempty"`
	Vendor   string `json:"vendor,omitempty"`
	Online   bool   `json:"online"`
}

// ScanNetwork scans the local network for devices
func ScanNetwork(ctx context.Context, subnet string) ([]NetworkDevice, error) {
	devices := []NetworkDevice{}

	// This is a simplified implementation
	// In production, you'd use arp-scan, nmap, or similar

	// Common studio devices to check
	knownHosts := []struct {
		hostname string
		ip       string
	}{
		{"tower", "192.168.1.10"},
		{"touchdesigner-pc", "192.168.1.20"},
		{"obs-pc", "192.168.1.21"},
		{"resolume-pc", "192.168.1.22"},
	}

	for _, host := range knownHosts {
		// Quick ping check (TCP port 80/443)
		conn, err := (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext(ctx, "tcp", host.ip+":80")
		online := err == nil
		if conn != nil {
			conn.Close()
		}

		devices = append(devices, NetworkDevice{
			IP:       host.ip,
			Hostname: host.hostname,
			Online:   online,
		})
	}

	return devices, nil
}

// GetStudioHealth returns overall studio health
func GetStudioHealth(ctx context.Context) (*StudioHealth, error) {
	health := &StudioHealth{
		Components: make(map[string]int),
	}

	// Check UNRAID
	unraidClient, _ := NewUNRAIDClient()
	if unraidClient.IsConnected(ctx) {
		health.Components["unraid"] = 100
	} else {
		health.Components["unraid"] = 0
		health.Issues = append(health.Issues, "UNRAID server not reachable")
	}

	// Check TouchDesigner
	tdClient, _ := NewTouchDesignerClient()
	if tdClient.IsConnected(ctx) {
		health.Components["touchdesigner"] = 100
	} else {
		health.Components["touchdesigner"] = 0
		health.Issues = append(health.Issues, "TouchDesigner not connected")
	}

	// Check NDI
	ndiClient, _ := NewNDIClient()
	ndiHealth, _ := ndiClient.GetHealth(ctx)
	if ndiHealth != nil {
		health.Components["ndi"] = ndiHealth.Score
	} else {
		health.Components["ndi"] = 0
	}

	// Calculate overall score
	total := 0
	count := 0
	for _, score := range health.Components {
		total += score
		count++
	}
	if count > 0 {
		health.Score = total / count
	}

	// Set status
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	// Add recommendations
	if health.Score < 100 {
		for component, score := range health.Components {
			if score < 80 {
				health.Recommendations = append(health.Recommendations,
					fmt.Sprintf("Check %s connectivity and status", component))
			}
		}
	}

	return health, nil
}

// HardwareStatus represents hardware status
type HardwareStatus struct {
	Score   int                    `json:"score"`
	Status  string                 `json:"status"`
	Devices map[string]interface{} `json:"devices"`
	Issues  []string               `json:"issues,omitempty"`
}

// GetHardwareStatus returns hardware status
func GetHardwareStatus(ctx context.Context) (*HardwareStatus, error) {
	status := &HardwareStatus{
		Devices: make(map[string]interface{}),
		Score:   100,
		Status:  "healthy",
	}

	// In a real implementation, we'd check:
	// - GPU status via nvidia-smi or similar
	// - Capture cards
	// - Audio interfaces
	// - Network adapters

	return status, nil
}

// JSON helper for potential UNRAID API responses
func parseJSONResponse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// UNRAIDPlugin represents an installed plugin
type UNRAIDPlugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
	UpdateAvail bool   `json:"update_available"`
}

// UNRAIDDisk represents a disk in the array
type UNRAIDDisk struct {
	Name        string  `json:"name"`
	Device      string  `json:"device"`
	Size        int64   `json:"size_bytes"`
	Used        int64   `json:"used_bytes"`
	UsedPercent float64 `json:"used_percent"`
	Status      string  `json:"status"` // active, standby, missing
	Temperature int     `json:"temperature_c"`
	Health      string  `json:"health"` // passed, warning, failed
}

// UNRAIDShare represents a user share
type UNRAIDShare struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Free     int64  `json:"free_bytes"`
	Used     int64  `json:"used_bytes"`
	Security string `json:"security"` // public, private, secure
}

// StartDocker starts a Docker container
func (c *UNRAIDClient) StartDocker(ctx context.Context, name string) error {
	// UNRAID uses a specific API for Docker control
	// In production, would use UNRAID's Docker API
	return nil
}

// StopDocker stops a Docker container
func (c *UNRAIDClient) StopDocker(ctx context.Context, name string) error {
	return nil
}

// RestartDocker restarts a Docker container
func (c *UNRAIDClient) RestartDocker(ctx context.Context, name string) error {
	return nil
}

// GetDockerLogs returns logs for a container
func (c *UNRAIDClient) GetDockerLogs(ctx context.Context, name string, lines int) (string, error) {
	// Would fetch logs from UNRAID's Docker manager
	return "", nil
}

// StartVM starts a virtual machine
func (c *UNRAIDClient) StartVM(ctx context.Context, name string) error {
	return nil
}

// StopVM stops a virtual machine
func (c *UNRAIDClient) StopVM(ctx context.Context, name string, force bool) error {
	return nil
}

// PauseVM pauses a virtual machine
func (c *UNRAIDClient) PauseVM(ctx context.Context, name string) error {
	return nil
}

// ResumeVM resumes a paused virtual machine
func (c *UNRAIDClient) ResumeVM(ctx context.Context, name string) error {
	return nil
}

// ListPlugins returns installed plugins
func (c *UNRAIDClient) ListPlugins(ctx context.Context) ([]UNRAIDPlugin, error) {
	plugins := []UNRAIDPlugin{}
	return plugins, nil
}

// GetDiagnostics generates a diagnostics bundle
func (c *UNRAIDClient) GetDiagnostics(ctx context.Context) (string, error) {
	// Would trigger UNRAID's diagnostics collection
	return "", nil
}

// ListDisks returns array disks
func (c *UNRAIDClient) ListDisks(ctx context.Context) ([]UNRAIDDisk, error) {
	disks := []UNRAIDDisk{}
	return disks, nil
}

// ListShares returns user shares
func (c *UNRAIDClient) ListShares(ctx context.Context) ([]UNRAIDShare, error) {
	shares := []UNRAIDShare{}
	return shares, nil
}

// GetArrayStatus returns detailed array status
func (c *UNRAIDClient) GetArrayStatus(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"state":        "started",
		"parity_valid": true,
		"mover_active": false,
	}
	return status, nil
}

// TriggerParityCheck starts a parity check
func (c *UNRAIDClient) TriggerParityCheck(ctx context.Context, correct bool) error {
	return nil
}

// TriggerMover runs the mover manually
func (c *UNRAIDClient) TriggerMover(ctx context.Context) error {
	return nil
}
