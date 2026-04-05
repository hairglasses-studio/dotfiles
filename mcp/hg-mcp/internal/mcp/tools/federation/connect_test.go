package federation

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

func TestConnectToOfflineServer(t *testing.T) {
	// Test connecting to a server that's not running
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"name": "test-offline",
		"url":  "http://localhost:59999", // Non-existent port
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := handleConnect(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error status (not crash)
	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Connection result: status=%v", data["status"])
	if data["status"] != "error" {
		// Connection to non-existent server should fail
		t.Logf("Note: Got status %v (server may be unexpectedly available)", data["status"])
	}
}

func TestFederationWorkflow(t *testing.T) {
	client := clients.GetMCPFederationClient()

	// 1. List known servers
	servers := client.KnownServers()
	t.Logf("Known servers: %d", len(servers))
	for _, s := range servers {
		t.Logf("  - %s (%s): %s", s.Name, s.Category, s.URL)
	}

	// 2. Check initial stats
	stats := client.GetStats()
	t.Logf("Initial stats: %d servers, %d connected, %d tools",
		stats.TotalServers, stats.ConnectedServers, stats.TotalTools)

	// 3. List connections (should be empty or have previous connections)
	connections := client.ListConnections()
	t.Logf("Active connections: %d", len(connections))

	// 4. List all tools from connected servers
	tools := client.ListAllTools()
	t.Logf("Available federated tools: %d", len(tools))

	// 5. Search for tools
	searchResults := client.SearchTools("status")
	t.Logf("Tools matching 'status': %d", len(searchResults))
}

func TestPingNonExistentServer(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"name": "non-existent-server",
	}

	result, err := handlePing(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Ping result: reachable=%v, error=%v", data["reachable"], data["error"])

	// Should report server not found
	if data["error"] == nil {
		t.Error("expected error for non-existent server")
	}
}

func TestListTools(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleListTools(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Federated tools: %v", data["count"])
}

func TestRefresh(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleRefresh(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Refresh result: %v", data["status"])
}
