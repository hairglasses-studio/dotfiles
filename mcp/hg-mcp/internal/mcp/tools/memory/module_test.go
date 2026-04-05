package memory

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestRemember(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"content":  "Test memory: OBS crashes after 6 hours of streaming",
		"category": "equipment",
		"tags":     []interface{}{"obs", "streaming", "stability"},
	}

	result, err := handleRemember(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Remember result: %.200s", content.Text)
}

func TestRememberMinimalParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"content":  "Minimal memory test",
		"category": "workflow",
	}

	result, err := handleRemember(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Remember minimal: %.200s", content.Text)
}

func TestListMemories(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleListMemories(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("All memories: %.300s", content.Text)
}

func TestListMemoriesByCategory(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"category": "equipment",
	}

	result, err := handleListMemories(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Equipment memories: %.300s", content.Text)
}

func TestSearch(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query": "obs crash",
		"limit": 5,
	}

	result, err := handleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Search results: %.300s", content.Text)
}

func TestSearchEmptyQuery(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query": "",
	}

	result, err := handleSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Search with empty query: %.200s", content.Text)
}

func TestSaveInsight(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"symptoms":   []interface{}{"video dropout", "ndi source missing"},
		"resolution": "Restarted NDI Tools and reconnected source",
		"root_cause": "NDI discovery service hung",
		"equipment":  []interface{}{"ndi", "obs"},
	}

	result, err := handleSaveInsight(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Save insight result: %.300s", content.Text)
}

func TestListInsights(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"limit": 5,
	}

	result, err := handleListInsights(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Session insights: %.300s", content.Text)
}

func TestLoadContext(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query": "streaming setup",
	}

	result, err := handleLoadContext(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Load context: %.300s", content.Text)
}

func TestStats(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Memory stats: %.300s", content.Text)
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
