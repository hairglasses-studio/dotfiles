package swarm

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

func TestPatternDiscovery(t *testing.T) {
	client := clients.GetResearchSwarmClient()

	// Reset for clean test
	client.UpdateConfig(&clients.SwarmConfig{
		Enabled:             true,
		MinPatternFrequency: 2, // Lower threshold for testing
		AutoSuggest:         true,
		MaxHistorySize:      1000,
	})

	// Record repeated sequential patterns
	// Pattern: obs_status -> obs_scenes -> obs_switch_scene (repeated 3 times)
	for i := 0; i < 3; i++ {
		sessionID := "pattern-test-session"
		client.RecordUsage("aftrs_obs_status", "obs", nil, true, 100, sessionID)
		client.RecordUsage("aftrs_obs_scenes", "obs", nil, true, 50, sessionID)
		client.RecordUsage("aftrs_obs_switch_scene", "obs", nil, true, 200, sessionID)
	}

	// Record co-occurring tools (same session, different order)
	for i := 0; i < 3; i++ {
		sessionID := "cooccur-test-session"
		client.RecordUsage("aftrs_resolume_status", "resolume", nil, true, 100, sessionID)
		client.RecordUsage("aftrs_wled_status", "lighting", nil, true, 80, sessionID)
	}

	// Record error recovery pattern
	for i := 0; i < 3; i++ {
		sessionID := "error-test-session"
		client.RecordUsage("aftrs_obs_start_stream", "obs", nil, false, 500, sessionID) // Fails
		client.RecordUsage("aftrs_obs_restart", "obs", nil, true, 1000, sessionID)      // Recovery
	}

	// Run analysis
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}
	result, err := handleAnalyze(context.Background(), req)
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	var data map[string]interface{}
	content := result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	patternsFound := int(data["patterns_found"].(float64))
	suggestionsFound := int(data["suggestions_found"].(float64))

	t.Logf("Analysis results: %d patterns, %d suggestions", patternsFound, suggestionsFound)

	if patternsFound == 0 {
		t.Error("expected to find patterns")
	}

	// Get patterns
	req.Params.Arguments = map[string]interface{}{"limit": 20}
	result, _ = handlePatterns(context.Background(), req)
	content = result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Log("\nDiscovered Patterns:")
	if patterns, ok := data["patterns"].([]interface{}); ok {
		for _, p := range patterns {
			pattern := p.(map[string]interface{})
			t.Logf("  [%s] %v - freq: %v, conf: %.1f%%",
				pattern["type"], pattern["tools"],
				pattern["frequency"], pattern["confidence"])
			t.Logf("    %s", pattern["description"])
		}
	}

	// Get suggestions
	req.Params.Arguments = map[string]interface{}{"limit": 20}
	result, _ = handleSuggestions(context.Background(), req)
	content = result.Content[0].(mcp.TextContent)
	json.Unmarshal([]byte(content.Text), &data)

	t.Log("\nImprovement Suggestions:")
	if suggestions, ok := data["suggestions"].([]interface{}); ok {
		for _, s := range suggestions {
			sugg := s.(map[string]interface{})
			t.Logf("  [%s] %s", sugg["type"], sugg["title"])
			t.Logf("    %s", sugg["description"])
			t.Logf("    Priority: %v, Impact: %s, Effort: %s",
				sugg["priority"], sugg["impact"], sugg["effort"])
		}
	}
}

func TestSequentialPatternChainSuggestion(t *testing.T) {
	client := clients.GetResearchSwarmClient()

	// Configure for testing
	client.UpdateConfig(&clients.SwarmConfig{
		Enabled:             true,
		MinPatternFrequency: 3,
		AutoSuggest:         true,
		MaxHistorySize:      1000,
	})

	// Create a very frequent sequential pattern (should trigger chain suggestion)
	for i := 0; i < 10; i++ {
		sessionID := "chain-test"
		client.RecordUsage("aftrs_show_precheck", "showcontrol", nil, true, 100, sessionID)
		client.RecordUsage("aftrs_show_start", "showcontrol", nil, true, 200, sessionID)
	}

	// Analyze
	client.AnalyzePatterns()

	// Check for chain suggestion
	suggestions := client.GetSuggestions("chain", "new")
	t.Logf("Chain suggestions found: %d", len(suggestions))

	for _, s := range suggestions {
		t.Logf("  - %s (priority: %d)", s.Title, s.Priority)
	}
}
