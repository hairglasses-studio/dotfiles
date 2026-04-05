// Package clients provides API clients for external services.
// mcp_federation.go implements MCP server federation for connecting to remote MCP servers
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// MCPFederationClient manages connections to remote MCP servers
type MCPFederationClient struct {
	mu          sync.RWMutex
	connections map[string]*FederatedServer
	httpClient  *http.Client
}

// FederatedServer represents a connected remote MCP server
type FederatedServer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Transport   string            `json:"transport"` // "sse", "http", "stdio"
	Status      string            `json:"status"`    // "connected", "disconnected", "error"
	Tools       []FederatedTool   `json:"tools"`
	AuthType    string            `json:"auth_type,omitempty"` // "none", "apikey", "oauth"
	Tags        []string          `json:"tags,omitempty"`
	LastPing    time.Time         `json:"last_ping"`
	ConnectedAt time.Time         `json:"connected_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// FederatedTool represents a tool from a federated server
type FederatedTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Server      string                 `json:"server"` // Server ID
	InputSchema map[string]interface{} `json:"input_schema,omitempty"`
}

// FederatedCallResult represents the result of calling a federated tool
type FederatedCallResult struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Latency int64       `json:"latency_ms"`
}

// FederationStats contains statistics about federated connections
type FederationStats struct {
	TotalServers     int            `json:"total_servers"`
	ConnectedServers int            `json:"connected_servers"`
	TotalTools       int            `json:"total_tools"`
	ServersByStatus  map[string]int `json:"servers_by_status"`
	ToolsByServer    map[string]int `json:"tools_by_server"`
}

// KnownMCPServer represents a known MCP server that can be connected
type KnownMCPServer struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Transport   string   `json:"transport"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

var (
	mcpFederationOnce     sync.Once
	mcpFederationInstance *MCPFederationClient
)

// GetMCPFederationClient returns the singleton MCP federation client
func GetMCPFederationClient() *MCPFederationClient {
	mcpFederationOnce.Do(func() {
		mcpFederationInstance = NewMCPFederationClient()
	})
	return mcpFederationInstance
}

// NewMCPFederationClient creates a new MCP federation client
func NewMCPFederationClient() *MCPFederationClient {
	return &MCPFederationClient{
		connections: make(map[string]*FederatedServer),
		httpClient: httpclient.Standard(),
	}
}

// KnownServers returns a list of known MCP servers that can be connected
func (c *MCPFederationClient) KnownServers() []KnownMCPServer {
	return []KnownMCPServer{
		// Infrastructure
		{
			Name:        "unraid-monolith",
			URL:         "http://192.168.50.10:8823",
			Transport:   "http",
			Description: "UNRAID server MCP for infrastructure management (23 tools)",
			Category:    "infrastructure",
			Tags:        []string{"unraid", "docker", "storage", "vms", "server"},
		},
		{
			Name:        "opnsense-monolith",
			URL:         "http://192.168.50.1:8822",
			Transport:   "http",
			Description: "OPNsense firewall MCP for network management (15+ tools)",
			Category:    "infrastructure",
			Tags:        []string{"opnsense", "firewall", "network", "security"},
		},
		{
			Name:        "home-assistant",
			URL:         "http://homeassistant.local:8123/mcp",
			Transport:   "http",
			Description: "Home Assistant MCP for smart home control",
			Category:    "infrastructure",
			Tags:        []string{"smarthome", "automation", "iot"},
		},
		// Creative
		{
			Name:        "touchdesigner",
			URL:         "http://localhost:9988",
			Transport:   "http",
			Description: "TouchDesigner MCP server for visual programming",
			Category:    "creative",
			Tags:        []string{"visuals", "generative", "realtime"},
		},
		{
			Name:        "resolume-mcp",
			URL:         "http://localhost:7000",
			Transport:   "http",
			Description: "Resolume Arena/Avenue MCP for VJ control",
			Category:    "creative",
			Tags:        []string{"vj", "visuals", "mapping"},
		},
		// Audio
		{
			Name:        "ableton-mcp",
			URL:         "http://localhost:11000",
			Transport:   "http",
			Description: "Ableton Live MCP for music production",
			Category:    "audio",
			Tags:        []string{"ableton", "daw", "music"},
		},
		{
			Name:        "cr8-mcp",
			URL:         "stdio:cr8-mcp",
			Transport:   "stdio",
			Description: "CR8 DJ/music library MCP (300+ tools)",
			Category:    "audio",
			Tags:        []string{"dj", "music", "library", "rekordbox"},
		},
		// Streaming
		{
			Name:        "obs-mcp",
			URL:         "http://localhost:4455/mcp",
			Transport:   "http",
			Description: "OBS Studio MCP for streaming control",
			Category:    "streaming",
			Tags:        []string{"obs", "streaming", "recording"},
		},
	}
}

// Connect connects to a remote MCP server
func (c *MCPFederationClient) Connect(ctx context.Context, name, url, transport string) (*FederatedServer, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already connected
	if existing, ok := c.connections[name]; ok && existing.Status == "connected" {
		return existing, nil
	}

	server := &FederatedServer{
		ID:          name,
		Name:        name,
		URL:         url,
		Transport:   transport,
		Status:      "connecting",
		Tools:       []FederatedTool{},
		ConnectedAt: time.Now(),
		Metadata:    make(map[string]string),
	}

	// Try to discover tools from the server
	tools, err := c.discoverTools(ctx, server)
	if err != nil {
		server.Status = "error"
		server.Error = err.Error()
		c.connections[name] = server
		return server, fmt.Errorf("failed to connect to %s: %w", name, err)
	}

	server.Tools = tools
	server.Status = "connected"
	server.LastPing = time.Now()
	c.connections[name] = server

	return server, nil
}

// discoverTools discovers available tools from a federated server
func (c *MCPFederationClient) discoverTools(ctx context.Context, server *FederatedServer) ([]FederatedTool, error) {
	// MCP tool discovery via tools/list
	listURL := fmt.Sprintf("%s/tools/list", server.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", listURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Try alternate endpoint
		return c.discoverToolsAlt(ctx, server)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.discoverToolsAlt(ctx, server)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tools []struct {
			Name        string                 `json:"name"`
			Description string                 `json:"description"`
			InputSchema map[string]interface{} `json:"inputSchema"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var tools []FederatedTool
	for _, t := range result.Tools {
		tools = append(tools, FederatedTool{
			Name:        t.Name,
			Description: t.Description,
			Server:      server.ID,
			InputSchema: t.InputSchema,
		})
	}

	return tools, nil
}

// discoverToolsAlt tries alternate discovery methods
func (c *MCPFederationClient) discoverToolsAlt(ctx context.Context, server *FederatedServer) ([]FederatedTool, error) {
	// Try /mcp endpoint for MCP-compliant servers
	mcpURL := fmt.Sprintf("%s/mcp", server.URL)

	req, err := http.NewRequestWithContext(ctx, "GET", mcpURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server unreachable: %w", err)
	}
	defer resp.Body.Close()

	// If we can reach the server, return empty tools (server is up but no discovery)
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return []FederatedTool{}, nil
	}

	return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
}

// Disconnect disconnects from a federated server
func (c *MCPFederationClient) Disconnect(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	server, ok := c.connections[name]
	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	server.Status = "disconnected"
	return nil
}

// CallTool calls a tool on a federated server
func (c *MCPFederationClient) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*FederatedCallResult, error) {
	c.mu.RLock()
	server, ok := c.connections[serverName]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("server not connected: %s", serverName)
	}

	if server.Status != "connected" {
		return nil, fmt.Errorf("server not in connected state: %s", server.Status)
	}

	start := time.Now()

	// Build MCP tool call request
	callURL := fmt.Sprintf("%s/tools/call", server.URL)

	payload := map[string]interface{}{
		"name":      toolName,
		"arguments": args,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", callURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// For now, simulate the call since we need proper MCP client
	// In production, this would use the MCP protocol properly
	resp, err := c.httpClient.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &FederatedCallResult{
			Success: false,
			Error:   err.Error(),
			Latency: latency,
		}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return &FederatedCallResult{
			Success: false,
			Error:   fmt.Sprintf("server returned %d: %s", resp.StatusCode, string(body)),
			Latency: latency,
		}, nil
	}

	var result interface{}
	json.Unmarshal(body, &result)

	// Update last ping
	c.mu.Lock()
	server.LastPing = time.Now()
	c.mu.Unlock()

	_ = payloadBytes // Acknowledge payload for future use

	return &FederatedCallResult{
		Success: true,
		Result:  result,
		Latency: latency,
	}, nil
}

// ListConnections returns all federated server connections
func (c *MCPFederationClient) ListConnections() []*FederatedServer {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var servers []*FederatedServer
	for _, server := range c.connections {
		servers = append(servers, server)
	}
	return servers
}

// GetConnection returns a specific connection by name
func (c *MCPFederationClient) GetConnection(name string) (*FederatedServer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	server, ok := c.connections[name]
	if !ok {
		return nil, fmt.Errorf("server not found: %s", name)
	}
	return server, nil
}

// ListAllTools returns all tools from all connected servers
func (c *MCPFederationClient) ListAllTools() []FederatedTool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var tools []FederatedTool
	for _, server := range c.connections {
		if server.Status == "connected" {
			tools = append(tools, server.Tools...)
		}
	}
	return tools
}

// SearchTools searches for tools across all connected servers
func (c *MCPFederationClient) SearchTools(query string) []FederatedTool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []FederatedTool
	for _, server := range c.connections {
		if server.Status != "connected" {
			continue
		}
		for _, tool := range server.Tools {
			if containsIgnoreCase(tool.Name, query) || containsIgnoreCase(tool.Description, query) {
				results = append(results, tool)
			}
		}
	}
	return results
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsLower(toLower(s), toLower(substr))))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ping checks if a server is still reachable
func (c *MCPFederationClient) Ping(ctx context.Context, name string) (bool, time.Duration, error) {
	c.mu.RLock()
	server, ok := c.connections[name]
	c.mu.RUnlock()

	if !ok {
		return false, 0, fmt.Errorf("server not found: %s", name)
	}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		return false, 0, err
	}

	resp, err := c.httpClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		c.mu.Lock()
		server.Status = "error"
		server.Error = err.Error()
		c.mu.Unlock()
		return false, latency, err
	}
	defer resp.Body.Close()

	c.mu.Lock()
	server.LastPing = time.Now()
	if server.Status == "error" {
		server.Status = "connected"
		server.Error = ""
	}
	c.mu.Unlock()

	return true, latency, nil
}

// GetStats returns federation statistics
func (c *MCPFederationClient) GetStats() *FederationStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &FederationStats{
		ServersByStatus: make(map[string]int),
		ToolsByServer:   make(map[string]int),
	}

	for _, server := range c.connections {
		stats.TotalServers++
		stats.ServersByStatus[server.Status]++

		if server.Status == "connected" {
			stats.ConnectedServers++
			stats.TotalTools += len(server.Tools)
			stats.ToolsByServer[server.Name] = len(server.Tools)
		}
	}

	return stats
}

// RefreshAll refreshes tool lists from all connected servers
func (c *MCPFederationClient) RefreshAll(ctx context.Context) error {
	c.mu.Lock()
	servers := make([]*FederatedServer, 0, len(c.connections))
	for _, s := range c.connections {
		servers = append(servers, s)
	}
	c.mu.Unlock()

	for _, server := range servers {
		if server.Status != "connected" {
			continue
		}

		tools, err := c.discoverTools(ctx, server)
		if err != nil {
			c.mu.Lock()
			server.Status = "error"
			server.Error = err.Error()
			c.mu.Unlock()
			continue
		}

		c.mu.Lock()
		server.Tools = tools
		server.LastPing = time.Now()
		c.mu.Unlock()
	}

	return nil
}
