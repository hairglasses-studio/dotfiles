package federation

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestListKnownServers(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleListKnownServers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("tool returned error: %v", result.Content)
	}

	// Parse result
	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	if err := json.Unmarshal([]byte(content.Text), &data); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	count := int(data["count"].(float64))
	if count < 5 {
		t.Errorf("expected at least 5 known servers, got %d", count)
	}

	servers := data["servers"].([]interface{})
	t.Logf("Found %d known servers:", count)
	for _, s := range servers {
		server := s.(map[string]interface{})
		t.Logf("  - %s (%s): %s", server["name"], server["category"], server["url"])
	}
}

func TestListKnownServersByCategory(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"category": "creative",
	}

	result, err := handleListKnownServers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	servers := data["servers"].([]interface{})
	for _, s := range servers {
		server := s.(map[string]interface{})
		if server["category"] != "creative" {
			t.Errorf("expected category 'creative', got '%s'", server["category"])
		}
	}
	t.Logf("Found %d creative servers", len(servers))
}

func TestListConnections(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleListConnections(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Active connections: %v", data["count"])
}

func TestStats(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Federation stats: total_servers=%v, connected=%v, total_tools=%v",
		data["total_servers"], data["connected_servers"], data["total_tools"])
}
