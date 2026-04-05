// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// OPNsenseClient provides access to OPNsense firewall API
type OPNsenseClient struct {
	host       string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

// OPNsenseStatus represents firewall status
type OPNsenseStatus struct {
	Connected      bool    `json:"connected"`
	Hostname       string  `json:"hostname"`
	Version        string  `json:"version"`
	Uptime         string  `json:"uptime"`
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	PFStatesCount  int     `json:"pf_states_count"`
	PFStatesMax    int     `json:"pf_states_max"`
	InterfaceCount int     `json:"interface_count"`
}

// OPNsenseFirewallRule represents a firewall rule
type OPNsenseFirewallRule struct {
	UUID        string `json:"uuid"`
	Enabled     bool   `json:"enabled"`
	Action      string `json:"action"`    // pass, block, reject
	Direction   string `json:"direction"` // in, out
	Interface   string `json:"interface"`
	Protocol    string `json:"protocol"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Port        string `json:"port"`
	Description string `json:"description"`
}

// OPNsenseNATRule represents a NAT/port forward rule
type OPNsenseNATRule struct {
	UUID         string `json:"uuid"`
	Enabled      bool   `json:"enabled"`
	Interface    string `json:"interface"`
	Protocol     string `json:"protocol"`
	ExternalIP   string `json:"external_ip"`
	ExternalPort string `json:"external_port"`
	InternalIP   string `json:"internal_ip"`
	InternalPort string `json:"internal_port"`
	Description  string `json:"description"`
}

// OPNsenseInterface represents a network interface
type OPNsenseInterface struct {
	Name      string `json:"name"`
	Device    string `json:"device"`
	Enabled   bool   `json:"enabled"`
	IPAddress string `json:"ip_address"`
	Subnet    string `json:"subnet"`
	Gateway   string `json:"gateway,omitempty"`
	Status    string `json:"status"` // up, down
	MediaType string `json:"media_type"`
	BytesIn   int64  `json:"bytes_in"`
	BytesOut  int64  `json:"bytes_out"`
}

// OPNsenseRoute represents a routing table entry
type OPNsenseRoute struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Flags       string `json:"flags"`
	Interface   string `json:"interface"`
}

// OPNsenseService represents a service
type OPNsenseService struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Running     bool   `json:"running"`
	Enabled     bool   `json:"enabled"`
}

// OPNsenseLogEntry represents a firewall log entry
type OPNsenseLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Interface  string    `json:"interface"`
	Action     string    `json:"action"`
	Direction  string    `json:"direction"`
	Protocol   string    `json:"protocol"`
	SourceIP   string    `json:"source_ip"`
	SourcePort string    `json:"source_port"`
	DestIP     string    `json:"dest_ip"`
	DestPort   string    `json:"dest_port"`
}

// OPNsenseHealth represents firewall health status
type OPNsenseHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// OPNsenseProposedRule represents a proposed firewall rule change
type OPNsenseProposedRule struct {
	ID        string               `json:"id"`
	Action    string               `json:"action"` // add, modify, delete
	Rule      OPNsenseFirewallRule `json:"rule"`
	Reason    string               `json:"reason"`
	CreatedAt time.Time            `json:"created_at"`
	Status    string               `json:"status"` // pending, approved, rejected
}

// NewOPNsenseClient creates a new OPNsense client
func NewOPNsenseClient() (*OPNsenseClient, error) {
	host := os.Getenv("OPNSENSE_HOST")
	if host == "" {
		host = "192.168.50.1"
	}

	apiKey := os.Getenv("OPNSENSE_API_KEY")
	apiSecret := os.Getenv("OPNSENSE_API_SECRET")

	return &OPNsenseClient{
		host:       host,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: httpclient.Standard(),
	}, nil
}

// baseURL returns the base URL for OPNsense API
func (c *OPNsenseClient) baseURL() string {
	return fmt.Sprintf("https://%s/api", c.host)
}

// doRequest performs an authenticated request to OPNsense
func (c *OPNsenseClient) doRequest(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	url := c.baseURL() + endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// OPNsense uses HTTP Basic Auth with API key/secret
	if c.apiKey != "" && c.apiSecret != "" {
		req.SetBasicAuth(c.apiKey, c.apiSecret)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// GetStatus returns firewall status
func (c *OPNsenseClient) GetStatus(ctx context.Context) (*OPNsenseStatus, error) {
	data, err := c.doRequest(ctx, "GET", "/core/system/status", nil)
	if err != nil {
		// Return basic status if API fails
		return &OPNsenseStatus{
			Connected: false,
		}, nil
	}

	var result struct {
		System struct {
			Hostname string  `json:"name"`
			Version  string  `json:"versions"`
			Uptime   string  `json:"uptime"`
			CPU      float64 `json:"cpu,string"`
			Memory   float64 `json:"memory,string"`
		} `json:"system"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &OPNsenseStatus{
		Connected:   true,
		Hostname:    result.System.Hostname,
		Version:     result.System.Version,
		Uptime:      result.System.Uptime,
		CPUUsage:    result.System.CPU,
		MemoryUsage: result.System.Memory,
	}, nil
}

// GetHealth returns firewall health assessment
func (c *OPNsenseClient) GetHealth(ctx context.Context) (*OPNsenseHealth, error) {
	status, err := c.GetStatus(ctx)
	if err != nil || !status.Connected {
		return &OPNsenseHealth{
			Score:  0,
			Status: "unavailable",
			Issues: []string{"OPNsense firewall not reachable"},
		}, nil
	}

	score := 100
	var issues []string
	var recommendations []string

	// Check CPU usage
	if status.CPUUsage > 80 {
		score -= 20
		issues = append(issues, fmt.Sprintf("High CPU usage: %.1f%%", status.CPUUsage))
	}

	// Check memory usage
	if status.MemoryUsage > 80 {
		score -= 20
		issues = append(issues, fmt.Sprintf("High memory usage: %.1f%%", status.MemoryUsage))
	}

	// Check state table
	if status.PFStatesMax > 0 {
		stateUsage := float64(status.PFStatesCount) / float64(status.PFStatesMax) * 100
		if stateUsage > 80 {
			score -= 15
			issues = append(issues, fmt.Sprintf("State table %.1f%% full", stateUsage))
			recommendations = append(recommendations, "Consider increasing state table size or reviewing firewall rules")
		}
	}

	healthStatus := "healthy"
	if score < 50 {
		healthStatus = "critical"
	} else if score < 80 {
		healthStatus = "degraded"
	}

	return &OPNsenseHealth{
		Score:           score,
		Status:          healthStatus,
		Issues:          issues,
		Recommendations: recommendations,
	}, nil
}

// GetFirewallRules returns active firewall rules
func (c *OPNsenseClient) GetFirewallRules(ctx context.Context) ([]OPNsenseFirewallRule, error) {
	data, err := c.doRequest(ctx, "GET", "/firewall/filter/searchRule", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []struct {
			UUID        string `json:"uuid"`
			Enabled     string `json:"enabled"`
			Action      string `json:"action"`
			Direction   string `json:"direction"`
			Interface   string `json:"interface"`
			Protocol    string `json:"protocol"`
			Source      string `json:"source_net"`
			Destination string `json:"destination_net"`
			Port        string `json:"destination_port"`
			Description string `json:"description"`
		} `json:"rows"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse rules: %w", err)
	}

	rules := make([]OPNsenseFirewallRule, len(result.Rows))
	for i, r := range result.Rows {
		rules[i] = OPNsenseFirewallRule{
			UUID:        r.UUID,
			Enabled:     r.Enabled == "1",
			Action:      r.Action,
			Direction:   r.Direction,
			Interface:   r.Interface,
			Protocol:    r.Protocol,
			Source:      r.Source,
			Destination: r.Destination,
			Port:        r.Port,
			Description: r.Description,
		}
	}

	return rules, nil
}

// GetFirewallStates returns current connection states
func (c *OPNsenseClient) GetFirewallStates(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	data, err := c.doRequest(ctx, "GET", "/diagnostics/firewall/states", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		States []map[string]interface{} `json:"states"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse states: %w", err)
	}

	if limit > 0 && limit < len(result.States) {
		return result.States[:limit], nil
	}
	return result.States, nil
}

// GetNATRules returns NAT/port forward rules
func (c *OPNsenseClient) GetNATRules(ctx context.Context) ([]OPNsenseNATRule, error) {
	data, err := c.doRequest(ctx, "GET", "/firewall/nat/searchRule", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []struct {
			UUID        string `json:"uuid"`
			Enabled     string `json:"enabled"`
			Interface   string `json:"interface"`
			Protocol    string `json:"protocol"`
			Target      string `json:"target"`
			LocalPort   string `json:"local-port"`
			Description string `json:"description"`
		} `json:"rows"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse NAT rules: %w", err)
	}

	rules := make([]OPNsenseNATRule, len(result.Rows))
	for i, r := range result.Rows {
		rules[i] = OPNsenseNATRule{
			UUID:         r.UUID,
			Enabled:      r.Enabled == "1",
			Interface:    r.Interface,
			Protocol:     r.Protocol,
			InternalIP:   r.Target,
			InternalPort: r.LocalPort,
			Description:  r.Description,
		}
	}

	return rules, nil
}

// GetInterfaces returns network interfaces
func (c *OPNsenseClient) GetInterfaces(ctx context.Context) ([]OPNsenseInterface, error) {
	data, err := c.doRequest(ctx, "GET", "/interfaces/overview/export", nil)
	if err != nil {
		return nil, err
	}

	var result map[string]struct {
		Device   string `json:"device"`
		Enabled  string `json:"enabled"`
		Addr4    string `json:"addr4"`
		Subnet   string `json:"subnet"`
		Gateway  string `json:"gateway"`
		Status   string `json:"status"`
		Media    string `json:"media"`
		BytesIn  int64  `json:"bytes_in,string"`
		BytesOut int64  `json:"bytes_out,string"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}

	var interfaces []OPNsenseInterface
	for name, iface := range result {
		interfaces = append(interfaces, OPNsenseInterface{
			Name:      name,
			Device:    iface.Device,
			Enabled:   iface.Enabled == "1",
			IPAddress: iface.Addr4,
			Subnet:    iface.Subnet,
			Gateway:   iface.Gateway,
			Status:    iface.Status,
			MediaType: iface.Media,
			BytesIn:   iface.BytesIn,
			BytesOut:  iface.BytesOut,
		})
	}

	return interfaces, nil
}

// GetRoutes returns routing table
func (c *OPNsenseClient) GetRoutes(ctx context.Context) ([]OPNsenseRoute, error) {
	data, err := c.doRequest(ctx, "GET", "/routes/gateway/status", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []struct {
			Network   string `json:"network"`
			Gateway   string `json:"gateway"`
			Flags     string `json:"flags"`
			Interface string `json:"interface"`
		} `json:"items"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}

	routes := make([]OPNsenseRoute, len(result.Items))
	for i, r := range result.Items {
		routes[i] = OPNsenseRoute{
			Destination: r.Network,
			Gateway:     r.Gateway,
			Flags:       r.Flags,
			Interface:   r.Interface,
		}
	}

	return routes, nil
}

// GetServices returns service status
func (c *OPNsenseClient) GetServices(ctx context.Context) ([]OPNsenseService, error) {
	data, err := c.doRequest(ctx, "GET", "/core/service/search", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Running     int    `json:"running"`
			Enabled     int    `json:"enabled"`
		} `json:"rows"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse services: %w", err)
	}

	services := make([]OPNsenseService, len(result.Rows))
	for i, s := range result.Rows {
		services[i] = OPNsenseService{
			Name:        s.Name,
			Description: s.Description,
			Running:     s.Running == 1,
			Enabled:     s.Enabled == 1,
		}
	}

	return services, nil
}

// RestartService restarts a service (only whitelisted services)
func (c *OPNsenseClient) RestartService(ctx context.Context, service string) error {
	// Whitelist of services that can be restarted
	allowed := map[string]bool{
		"unbound": true,
		"dnsmasq": true,
		"dhcpd":   true,
		"ntpd":    true,
		"openvpn": true,
		"haproxy": true,
		"squid":   true,
		"monit":   true,
	}

	if !allowed[strings.ToLower(service)] {
		return fmt.Errorf("service '%s' is not in the allowed restart list", service)
	}

	endpoint := fmt.Sprintf("/core/service/restart/%s", service)
	_, err := c.doRequest(ctx, "POST", endpoint, nil)
	return err
}

// GetLogs returns recent firewall logs
func (c *OPNsenseClient) GetLogs(ctx context.Context, limit int) ([]OPNsenseLogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	endpoint := fmt.Sprintf("/diagnostics/log/core/filter?limit=%d", limit)
	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []struct {
			Timestamp string `json:"timestamp"`
			Interface string `json:"interface"`
			Action    string `json:"action"`
			Direction string `json:"dir"`
			Protocol  string `json:"protoname"`
			SrcIP     string `json:"src"`
			SrcPort   string `json:"srcport"`
			DstIP     string `json:"dst"`
			DstPort   string `json:"dstport"`
		} `json:"rows"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	logs := make([]OPNsenseLogEntry, len(result.Rows))
	for i, l := range result.Rows {
		ts, _ := time.Parse("2006-01-02T15:04:05", l.Timestamp)
		logs[i] = OPNsenseLogEntry{
			Timestamp:  ts,
			Interface:  l.Interface,
			Action:     l.Action,
			Direction:  l.Direction,
			Protocol:   l.Protocol,
			SourceIP:   l.SrcIP,
			SourcePort: l.SrcPort,
			DestIP:     l.DstIP,
			DestPort:   l.DstPort,
		}
	}

	return logs, nil
}

// Ping performs a network ping test
func (c *OPNsenseClient) Ping(ctx context.Context, host string, count int) (map[string]interface{}, error) {
	if count <= 0 {
		count = 4
	}

	endpoint := fmt.Sprintf("/diagnostics/interface/ping?host=%s&count=%d", host, count)
	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ping result: %w", err)
	}

	return result, nil
}

// Traceroute performs a traceroute
func (c *OPNsenseClient) Traceroute(ctx context.Context, host string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/diagnostics/interface/traceroute?host=%s", host)
	data, err := c.doRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Rows []map[string]interface{} `json:"rows"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse traceroute: %w", err)
	}

	return result.Rows, nil
}

// BackupConfig creates a configuration backup
func (c *OPNsenseClient) BackupConfig(ctx context.Context) ([]byte, error) {
	data, err := c.doRequest(ctx, "GET", "/core/backup/download/this", nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}
