// Package swarm provides research swarm MCP tools for pattern discovery
package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the swarm tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "swarm" }

// Description returns the module description
func (m *Module) Description() string {
	return "Research swarm for autonomous pattern discovery and improvement suggestions"
}

// Tools returns all swarm tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_swarm_record",
				mcp.WithDescription("Record a tool usage event for pattern analysis"),
				mcp.WithString("tool", mcp.Required(), mcp.Description("Tool name that was used")),
				mcp.WithString("category", mcp.Description("Tool category")),
				mcp.WithBoolean("success", mcp.Description("Whether the tool succeeded (default: true)")),
				mcp.WithNumber("duration_ms", mcp.Description("Execution duration in milliseconds")),
				mcp.WithString("session_id", mcp.Description("Session ID for sequence tracking")),
			),
			Handler:             handleRecord,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_analyze",
				mcp.WithDescription("Run pattern analysis on usage history. Discovers sequential, co-occurring, and error recovery patterns."),
			),
			Handler:             handleAnalyze,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_patterns",
				mcp.WithDescription("List discovered usage patterns"),
				mcp.WithString("type", mcp.Description("Filter by pattern type (sequential, co-occurring, error_recovery, optimization)")),
				mcp.WithNumber("min_frequency", mcp.Description("Minimum occurrence count (default: 1)")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 20)")),
			),
			Handler:             handlePatterns,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_suggestions",
				mcp.WithDescription("List improvement suggestions generated from patterns"),
				mcp.WithString("type", mcp.Description("Filter by suggestion type (consolidation, chain, alias, new_tool)")),
				mcp.WithString("status", mcp.Description("Filter by status (new, reviewed, accepted, rejected, implemented)")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 20)")),
			),
			Handler:             handleSuggestions,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_suggestion_update",
				mcp.WithDescription("Update the status of an improvement suggestion"),
				mcp.WithString("suggestion_id", mcp.Required(), mcp.Description("Suggestion ID")),
				mcp.WithString("status", mcp.Required(), mcp.Description("New status (reviewed, accepted, rejected, implemented)")),
				mcp.WithString("reviewed_by", mcp.Description("Who reviewed the suggestion")),
			),
			Handler:             handleSuggestionUpdate,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_tool_stats",
				mcp.WithDescription("Get usage statistics for tools"),
				mcp.WithNumber("limit", mcp.Description("Max tools to return (default: 10)")),
			),
			Handler:             handleToolStats,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_workers",
				mcp.WithDescription("Get status of swarm analysis workers"),
			),
			Handler:             handleWorkers,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_stats",
				mcp.WithDescription("Get overall swarm statistics"),
			),
			Handler:             handleStats,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
		{
			Tool: mcp.NewTool("aftrs_swarm_config",
				mcp.WithDescription("Get or update swarm configuration"),
				mcp.WithBoolean("enabled", mcp.Description("Enable/disable swarm analysis")),
				mcp.WithNumber("min_pattern_frequency", mcp.Description("Minimum occurrences for pattern detection")),
				mcp.WithBoolean("auto_suggest", mcp.Description("Auto-generate suggestions from patterns")),
			),
			Handler:             handleConfig,
			Category:            "swarm",
			CircuitBreakerGroup: "swarm",
		},
	}
}

func handleRecord(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tool := tools.GetStringParam(request, "tool")
	category := tools.GetStringParam(request, "category")
	success := tools.GetBoolParam(request, "success", true)
	durationMs := tools.GetIntParam(request, "duration_ms", 0)
	sessionID := tools.GetStringParam(request, "session_id")

	client := clients.GetResearchSwarmClient()
	err := client.RecordUsage(tool, category, nil, success, time.Duration(durationMs)*time.Millisecond, sessionID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"status":  "recorded",
		"tool":    tool,
		"success": success,
	}), nil
}

func handleAnalyze(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetResearchSwarmClient()
	err := client.AnalyzePatterns()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	stats := client.GetStats()
	return tools.JSONResult(map[string]interface{}{
		"status":            "analysis_complete",
		"patterns_found":    stats.TotalPatterns,
		"suggestions_found": stats.TotalSuggestions,
		"patterns_by_type":  stats.PatternsByType,
	}), nil
}

func handlePatterns(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	patternType := tools.GetStringParam(request, "type")
	minFrequency := tools.GetIntParam(request, "min_frequency", 1)
	limit := tools.GetIntParam(request, "limit", 20)

	client := clients.GetResearchSwarmClient()
	patterns := client.GetPatterns(patternType, minFrequency)

	if len(patterns) > limit {
		patterns = patterns[:limit]
	}

	return tools.JSONResult(map[string]interface{}{
		"count":    len(patterns),
		"patterns": patterns,
	}), nil
}

func handleSuggestions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	suggestionType := tools.GetStringParam(request, "type")
	status := tools.GetStringParam(request, "status")
	limit := tools.GetIntParam(request, "limit", 20)

	client := clients.GetResearchSwarmClient()
	suggestions := client.GetSuggestions(suggestionType, status)

	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return tools.JSONResult(map[string]interface{}{
		"count":       len(suggestions),
		"suggestions": suggestions,
	}), nil
}

func handleSuggestionUpdate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	suggestionID := tools.GetStringParam(request, "suggestion_id")
	status := tools.GetStringParam(request, "status")
	reviewedBy := tools.GetStringParam(request, "reviewed_by")
	if reviewedBy == "" {
		reviewedBy = "operator"
	}

	client := clients.GetResearchSwarmClient()
	err := client.UpdateSuggestionStatus(suggestionID, status, reviewedBy)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Suggestion %s updated to status: %s", suggestionID, status)), nil
}

func handleToolStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 10)

	client := clients.GetResearchSwarmClient()
	stats := client.GetToolStats(limit)

	return tools.JSONResult(map[string]interface{}{
		"count":      len(stats),
		"tool_stats": stats,
	}), nil
}

func handleWorkers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetResearchSwarmClient()
	workers := client.GetWorkerStatus()

	return tools.JSONResult(map[string]interface{}{
		"count":   len(workers),
		"workers": workers,
	}), nil
}

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetResearchSwarmClient()
	stats := client.GetStats()

	return tools.JSONResult(stats), nil
}

func handleConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetResearchSwarmClient()
	config := client.GetConfig()

	// Check if updating
	args, ok := request.Params.Arguments.(map[string]interface{})
	if ok {
		updated := false
		if v, ok := args["enabled"].(bool); ok {
			config.Enabled = v
			updated = true
		}
		if v, ok := args["min_pattern_frequency"].(float64); ok {
			config.MinPatternFrequency = int(v)
			updated = true
		}
		if v, ok := args["auto_suggest"].(bool); ok {
			config.AutoSuggest = v
			updated = true
		}

		if updated {
			client.UpdateConfig(config)
		}
	}

	return tools.JSONResult(config), nil
}
