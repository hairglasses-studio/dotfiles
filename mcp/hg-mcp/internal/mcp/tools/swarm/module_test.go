package swarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestRecordUsage(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"tool":        "aftrs_obs_status",
		"category":    "obs",
		"success":     true,
		"duration_ms": 150,
		"session_id":  "test-session-1",
	}

	result, err := handleRecord(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Fatalf("tool returned error")
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	if data["status"] != "recorded" {
		t.Errorf("expected status 'recorded', got '%v'", data["status"])
	}
	t.Logf("Recorded usage: %v", data)
}

func TestRecordMultipleUsages(t *testing.T) {
	// Record a sequence of tool usages
	tools := []string{
		"aftrs_obs_status",
		"aftrs_obs_scenes",
		"aftrs_obs_switch_scene",
		"aftrs_resolume_status",
		"aftrs_resolume_layers",
	}

	for _, tool := range tools {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]interface{}{
			"tool":        tool,
			"category":    "streaming",
			"success":     true,
			"duration_ms": 100,
			"session_id":  "test-session-2",
		}
		handleRecord(context.Background(), req)
	}

	t.Logf("Recorded %d tool usages in sequence", len(tools))
}

func TestAnalyzePatterns(t *testing.T) {
	// First record some usage
	TestRecordMultipleUsages(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAnalyze(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Analysis complete: patterns=%v, suggestions=%v",
		data["patterns_found"], data["suggestions_found"])
}

func TestGetPatterns(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"limit": 10,
	}

	result, err := handlePatterns(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Found %v patterns", data["count"])
	if patterns, ok := data["patterns"].([]interface{}); ok {
		for _, p := range patterns {
			pattern := p.(map[string]interface{})
			t.Logf("  - [%s] %v (freq: %v)", pattern["type"], pattern["tools"], pattern["frequency"])
		}
	}
}

func TestGetSuggestions(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"status": "new",
	}

	result, err := handleSuggestions(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Found %v suggestions", data["count"])
	if suggestions, ok := data["suggestions"].([]interface{}); ok {
		for _, s := range suggestions {
			sugg := s.(map[string]interface{})
			t.Logf("  - [%s] %s (priority: %v)", sugg["type"], sugg["title"], sugg["priority"])
		}
	}
}

func TestToolStats(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"limit": 5,
	}

	result, err := handleToolStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Tool stats: %v tools tracked", data["count"])
	if stats, ok := data["tool_stats"].([]interface{}); ok {
		for _, s := range stats {
			stat := s.(map[string]interface{})
			t.Logf("  - %s: count=%v, success_rate=%.1f%%",
				stat["tool"], stat["count"], stat["success_rate"])
		}
	}
}

func TestSwarmStats(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStats(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Swarm stats: usage=%v, patterns=%v, suggestions=%v",
		data["total_usage"], data["total_patterns"], data["total_suggestions"])
}

func TestConfig(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Swarm config: enabled=%v, min_freq=%v, auto_suggest=%v",
		data["enabled"], data["min_pattern_frequency"], data["auto_suggest"])
}

func TestWorkers(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleWorkers(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Logf("Workers: %v", data["count"])
	if workers, ok := data["workers"].([]interface{}); ok {
		for _, w := range workers {
			worker := w.(map[string]interface{})
			t.Logf("  - %s: %s (findings: %v)", worker["type"], worker["status"], worker["findings"])
		}
	}
}
