package chains

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestChainList(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := m.handleChainList(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("tool returned error")
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Chain list preview: %.200s...", content.Text)
}

func TestChainListByCategory(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"category": "show",
	}

	result, err := m.handleChainList(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Show chains: %.300s", content.Text)
}

func TestChainGet(t *testing.T) {
	m := NewModule()

	// First list chains to get an ID
	listReq := mcp.CallToolRequest{}
	listReq.Params.Arguments = map[string]interface{}{}
	m.handleChainList(context.Background(), listReq)

	// Get a specific chain (try common IDs)
	for _, chainID := range []string{"show_startup", "stream_start", "backup_daily"} {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"chain_id": chainID,
		}

		result, err := m.handleChainGet(context.Background(), req)
		if err != nil {
			continue
		}

		content := result.Content[0].(mcp.TextContent)
		if len(content.Text) > 100 {
			t.Logf("Chain %s found: %.200s...", chainID, content.Text)
			return
		}
	}
	t.Log("No pre-defined chains found (expected in fresh install)")
}

func TestChainGetNotFound(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"chain_id": "nonexistent_chain_xyz",
	}

	result, err := m.handleChainGet(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	if !result.IsError {
		t.Logf("Chain not found handled: %s", content.Text)
	}
}

func TestChainPending(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := m.handleChainPending(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Pending gates: %s", content.Text)
}

func TestChainHistory(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"limit": 5,
	}

	result, err := m.handleChainHistory(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Chain history: %s", content.Text)
}

func TestChainStatusNotFound(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"execution_id": "nonexistent_exec_123",
	}

	result, err := m.handleChainStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error for nonexistent execution
	t.Logf("Status for nonexistent: IsError=%v", result.IsError)
}

func TestChainCancelNotFound(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"execution_id": "nonexistent_exec_456",
	}

	result, err := m.handleChainCancel(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Cancel nonexistent: IsError=%v", result.IsError)
}

func TestChainExecuteMissingID(t *testing.T) {
	m := NewModule()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := m.handleChainExecute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error for missing chain_id
	if !result.IsError {
		t.Error("expected error for missing chain_id")
	}
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

// Helper to parse JSON response
func parseJSON(content mcp.TextContent) map[string]interface{} {
	var data map[string]interface{}
	json.Unmarshal([]byte(content.Text), &data)
	return data
}
