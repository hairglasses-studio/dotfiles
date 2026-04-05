// Package memory provides session memory MCP tools
package memory

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the memory tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "memory" }

// Description returns the module description
func (m *Module) Description() string {
	return "Session memory for team knowledge and insights"
}

// Tools returns all memory tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_memory_remember",
				mcp.WithDescription("Save team knowledge to memory. Use #remember pattern: 'Resolume needs restart after 4 hours'"),
				mcp.WithString("content", mcp.Required(), mcp.Description("Knowledge to remember")),
				mcp.WithString("category", mcp.Required(), mcp.Description("Category: equipment, workflow, venue, show, network")),
				mcp.WithArray("tags", mcp.Description("Optional tags for filtering")),
			),
			Handler:             handleRemember,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_forget",
				mcp.WithDescription("Remove a memory by ID"),
				mcp.WithString("memory_id", mcp.Required(), mcp.Description("Memory ID to remove")),
			),
			Handler:             handleForget,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_list",
				mcp.WithDescription("List saved memories by category"),
				mcp.WithString("category", mcp.Description("Filter by category (optional)")),
			),
			Handler:             handleListMemories,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_search",
				mcp.WithDescription("Search across memories and session insights"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 10)")),
			),
			Handler:             handleSearch,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_save_insight",
				mcp.WithDescription("Save a troubleshooting session insight with symptoms and resolution"),
				mcp.WithArray("symptoms", mcp.Required(), mcp.Description("List of observed symptoms")),
				mcp.WithString("resolution", mcp.Required(), mcp.Description("What fixed the issue")),
				mcp.WithString("root_cause", mcp.Description("Root cause if identified")),
				mcp.WithArray("equipment", mcp.Description("Equipment involved")),
				mcp.WithString("venue", mcp.Description("Venue if applicable")),
				mcp.WithArray("pitfalls", mcp.Description("What to avoid (negative learning)")),
			),
			Handler:             handleSaveInsight,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_insights",
				mcp.WithDescription("List past session insights, optionally filtered"),
				mcp.WithString("equipment", mcp.Description("Filter by equipment")),
				mcp.WithString("venue", mcp.Description("Filter by venue")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 20)")),
			),
			Handler:             handleListInsights,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_load_context",
				mcp.WithDescription("Load relevant memories and insights for a topic"),
				mcp.WithString("topic", mcp.Required(), mcp.Description("Topic to load context for (e.g., 'NDI issues', 'Resolume crash')")),
				mcp.WithArray("equipment", mcp.Description("Equipment to include in context")),
			),
			Handler:             handleLoadContext,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
		{
			Tool: mcp.NewTool("aftrs_memory_stats",
				mcp.WithDescription("Get memory system statistics"),
			),
			Handler:             handleStats,
			Category:            "memory",
			CircuitBreakerGroup: "memory",
		},
	}
}

func handleRemember(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := tools.GetStringParam(request, "content")
	category := tools.GetStringParam(request, "category")

	tags := tools.GetStringArrayParam(request, "tags")

	client := clients.GetSessionMemoryClient()
	memory, err := client.Remember(content, category, tags)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"status":   "remembered",
		"id":       memory.ID,
		"content":  memory.Content,
		"category": memory.Category,
	}

	return tools.JSONResult(result), nil
}

func handleForget(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	memoryID := tools.GetStringParam(request, "memory_id")

	client := clients.GetSessionMemoryClient()
	if err := client.Forget(memoryID); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Memory %s forgotten", memoryID)), nil
}

func handleListMemories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(request, "category")

	client := clients.GetSessionMemoryClient()
	memories := client.ListMemories(category)

	result := map[string]interface{}{
		"count":    len(memories),
		"memories": memories,
	}

	return tools.JSONResult(result), nil
}

func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := tools.GetStringParam(request, "query")
	limit := tools.GetIntParam(request, "limit", 10)

	client := clients.GetSessionMemoryClient()
	results := client.Search(query, limit)

	return tools.JSONResult(results), nil
}

func handleSaveInsight(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resolution := tools.GetStringParam(request, "resolution")
	rootCause := tools.GetStringParam(request, "root_cause")
	venue := tools.GetStringParam(request, "venue")

	symptoms := tools.GetStringArrayParam(request, "symptoms")
	equipment := tools.GetStringArrayParam(request, "equipment")
	pitfalls := tools.GetStringArrayParam(request, "pitfalls")

	insight := &clients.SessionInsight{
		Symptoms:  symptoms,
		RootCause: rootCause,
		Equipment: equipment,
		Venue:     venue,
		Pitfalls:  pitfalls,
		StepsWorked: []clients.InsightStep{
			{
				Action:      resolution,
				Description: resolution,
				Outcome:     "success",
				Order:       1,
			},
		},
		QualityScore: 3,
		Summary:      fmt.Sprintf("Resolved: %s", resolution),
	}

	client := clients.GetSessionMemoryClient()
	if err := client.SaveInsight(insight); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"status":   "saved",
		"id":       insight.ID,
		"symptoms": symptoms,
		"summary":  insight.Summary,
	}

	return tools.JSONResult(result), nil
}

func handleListInsights(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	equipment := tools.GetStringParam(request, "equipment")
	venue := tools.GetStringParam(request, "venue")
	limit := tools.GetIntParam(request, "limit", 20)

	client := clients.GetSessionMemoryClient()
	insights := client.ListInsights(equipment, venue, limit)

	result := map[string]interface{}{
		"count":    len(insights),
		"insights": insights,
	}

	return tools.JSONResult(result), nil
}

func handleLoadContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	topic := tools.GetStringParam(request, "topic")

	equipment := tools.GetStringArrayParam(request, "equipment")

	client := clients.GetSessionMemoryClient()
	ctx2 := client.LoadContext(topic, equipment)

	return tools.JSONResult(ctx2), nil
}

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetSessionMemoryClient()
	stats := client.GetStats()

	return tools.JSONResult(stats), nil
}
