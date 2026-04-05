// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// TailscaleClient provides Tailscale VPN management
type TailscaleClient struct {
	binPath string
}

// TailscaleStatus represents Tailscale status
type TailscaleStatus struct {
	BackendState   string                   `json:"BackendState"`
	Self           TailscalePeer            `json:"Self"`
	Peers          map[string]TailscalePeer `json:"Peer,omitempty"`
	MagicDNSSuffix string                   `json:"MagicDNSSuffix"`
	TailnetName    string                   `json:"CurrentTailnet,omitempty"`
}

// TailscalePeer represents a device in the tailnet
type TailscalePeer struct {
	ID             string   `json:"ID"`
	PublicKey      string   `json:"PublicKey"`
	HostName       string   `json:"HostName"`
	DNSName        string   `json:"DNSName"`
	OS             string   `json:"OS"`
	TailscaleIPs   []string `json:"TailscaleIPs"`
	Online         bool     `json:"Online"`
	LastSeen       string   `json:"LastSeen,omitempty"`
	ExitNode       bool     `json:"ExitNode"`
	ExitNodeOption bool     `json:"ExitNodeOption"`
	Active         bool     `json:"Active"`
	Relay          string   `json:"Relay,omitempty"`
	CurAddr        string   `json:"CurAddr,omitempty"`
}

// TailscaleDevice represents a simplified device view
type TailscaleDevice struct {
	Hostname      string   `json:"hostname"`
	DNSName       string   `json:"dns_name"`
	OS            string   `json:"os"`
	IPs           []string `json:"ips"`
	Online        bool     `json:"online"`
	LastSeen      string   `json:"last_seen,omitempty"`
	IsExitNode    bool     `json:"is_exit_node"`
	IsCurrentNode bool     `json:"is_current_node"`
	Connection    string   `json:"connection,omitempty"` // direct, relay
}

// TailscaleNetworkInfo represents network information
type TailscaleNetworkInfo struct {
	Connected    bool     `json:"connected"`
	TailnetName  string   `json:"tailnet_name"`
	MagicDNS     string   `json:"magic_dns"`
	SelfHostname string   `json:"self_hostname"`
	SelfIPs      []string `json:"self_ips"`
	DeviceCount  int      `json:"device_count"`
	OnlineCount  int      `json:"online_count"`
	ExitNode     string   `json:"exit_node,omitempty"`
}

// NewTailscaleClient creates a new Tailscale client
func NewTailscaleClient() (*TailscaleClient, error) {
	// Check if tailscale is available
	binPath, err := exec.LookPath("tailscale")
	if err != nil {
		return nil, fmt.Errorf("tailscale CLI not found: %w", err)
	}

	return &TailscaleClient{
		binPath: binPath,
	}, nil
}

// GetStatus returns the current Tailscale status
func (c *TailscaleClient) GetStatus(ctx context.Context) (*TailscaleStatus, error) {
	cmd := exec.CommandContext(ctx, c.binPath, "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tailscale status: %w", err)
	}

	var status TailscaleStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nil, fmt.Errorf("failed to parse tailscale status: %w", err)
	}

	return &status, nil
}

// GetNetworkInfo returns simplified network information
func (c *TailscaleClient) GetNetworkInfo(ctx context.Context) (*TailscaleNetworkInfo, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return &TailscaleNetworkInfo{Connected: false}, nil
	}

	info := &TailscaleNetworkInfo{
		Connected:    status.BackendState == "Running",
		TailnetName:  status.TailnetName,
		MagicDNS:     status.MagicDNSSuffix,
		SelfHostname: status.Self.HostName,
		SelfIPs:      status.Self.TailscaleIPs,
		DeviceCount:  len(status.Peers) + 1, // Include self
	}

	// Count online devices and find exit node
	for _, peer := range status.Peers {
		if peer.Online {
			info.OnlineCount++
		}
		if peer.ExitNode && peer.Active {
			info.ExitNode = peer.HostName
		}
	}
	if status.Self.Online {
		info.OnlineCount++
	}

	return info, nil
}

// ListDevices returns all devices in the tailnet
func (c *TailscaleClient) ListDevices(ctx context.Context) ([]TailscaleDevice, error) {
	status, err := c.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	devices := []TailscaleDevice{}

	// Add self first
	devices = append(devices, TailscaleDevice{
		Hostname:      status.Self.HostName,
		DNSName:       status.Self.DNSName,
		OS:            status.Self.OS,
		IPs:           status.Self.TailscaleIPs,
		Online:        status.Self.Online,
		IsExitNode:    status.Self.ExitNode,
		IsCurrentNode: true,
		Connection:    "local",
	})

	// Add peers
	for _, peer := range status.Peers {
		connection := "direct"
		if peer.Relay != "" && peer.CurAddr == "" {
			connection = "relay"
		}

		device := TailscaleDevice{
			Hostname:      peer.HostName,
			DNSName:       peer.DNSName,
			OS:            peer.OS,
			IPs:           peer.TailscaleIPs,
			Online:        peer.Online,
			LastSeen:      peer.LastSeen,
			IsExitNode:    peer.ExitNode,
			IsCurrentNode: false,
			Connection:    connection,
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// Ping pings a device in the tailnet
func (c *TailscaleClient) Ping(ctx context.Context, target string) (*PingResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.binPath, "ping", "--c", "3", target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &PingResult{
			Success: false,
			Target:  target,
			Error:   string(output),
		}, nil
	}

	result := &PingResult{
		Success: true,
		Target:  target,
		Output:  string(output),
	}

	// Parse output for latency
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "pong from") {
			// Extract latency: "pong from hostname (100.x.x.x) via DERP(region) in 50ms"
			if idx := strings.LastIndex(line, " in "); idx >= 0 {
				latencyStr := strings.TrimSpace(line[idx+4:])
				result.Latency = latencyStr
			}
		}
	}

	return result, nil
}

// PingResult represents the result of a ping
type PingResult struct {
	Success bool   `json:"success"`
	Target  string `json:"target"`
	Latency string `json:"latency,omitempty"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SetExitNode sets or clears the exit node
func (c *TailscaleClient) SetExitNode(ctx context.Context, node string) error {
	args := []string{"set"}
	if node == "" || node == "none" {
		args = append(args, "--exit-node=")
	} else {
		args = append(args, fmt.Sprintf("--exit-node=%s", node))
	}

	cmd := exec.CommandContext(ctx, c.binPath, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set exit node: %s", string(output))
	}

	return nil
}

// GetExitNodes returns available exit nodes
func (c *TailscaleClient) GetExitNodes(ctx context.Context) ([]TailscaleDevice, error) {
	devices, err := c.ListDevices(ctx)
	if err != nil {
		return nil, err
	}

	exitNodes := []TailscaleDevice{}
	for _, device := range devices {
		if device.IsExitNode {
			exitNodes = append(exitNodes, device)
		}
	}

	return exitNodes, nil
}

// IsConnected checks if Tailscale is connected
func (c *TailscaleClient) IsConnected(ctx context.Context) bool {
	info, err := c.GetNetworkInfo(ctx)
	if err != nil {
		return false
	}
	return info.Connected
}

// Up brings up the Tailscale connection
func (c *TailscaleClient) Up(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.binPath, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up tailscale: %s", string(output))
	}
	return nil
}

// Down brings down the Tailscale connection
func (c *TailscaleClient) Down(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.binPath, "down")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring down tailscale: %s", string(output))
	}
	return nil
}

// GetVersion returns the Tailscale version
func (c *TailscaleClient) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.binPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
