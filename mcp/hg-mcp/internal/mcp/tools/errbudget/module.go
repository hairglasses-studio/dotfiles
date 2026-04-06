package errbudget

import (
	"context"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// GlobalTracker is the singleton error budget tracker used by the middleware
// layer and exposed to agents via MCP tools.
var GlobalTracker = NewTracker(DefaultThreshold)

// Module implements tools.ToolModule for error budget management.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "errbudget" }
func (m *Module) Description() string { return "Per-tool consecutive error tracking and degraded mode" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_errbudget_status",
				mcp.WithDescription("Get the error budget status for all tracked tools or a specific tool. Shows consecutive errors, degraded state, and error history."),
				mcp.WithString("tool_name", mcp.Description("Specific tool name to check (omit for all tools)")),
				mcp.WithBoolean("degraded_only", mcp.Description("If true, only show tools in degraded state (default: false)")),
			),
			Handler:    handleStatus,
			Category:   "platform",
			Tags:       []string{"errbudget", "health", "monitoring", "errors", "degraded"},
			UseCases:   []string{"Check tool health", "Find degraded tools", "Monitor error rates"},
			Complexity: tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_errbudget_reset",
				mcp.WithDescription("Reset the error budget for a specific tool or all tools, clearing their degraded state so they can be invoked again."),
				mcp.WithString("tool_name", mcp.Description("Tool name to reset (omit to reset all)")),
			),
			Handler:    handleReset,
			Category:   "platform",
			Tags:       []string{"errbudget", "reset", "recovery", "health"},
			UseCases:   []string{"Recover degraded tool", "Clear error state", "Manual health reset"},
			Complexity: tools.ComplexitySimple,
			IsWrite:    true,
		},
	}
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolName := tools.GetStringParam(req, "tool_name")
	degradedOnly := tools.GetBoolParam(req, "degraded_only", false)

	if toolName != "" {
		status := GlobalTracker.Status(toolName)
		return tools.JSONResult(status), nil
	}

	allStatus := GlobalTracker.AllStatus()

	if degradedOnly {
		filtered := make([]ToolStatus, 0)
		for _, s := range allStatus {
			if s.Degraded {
				filtered = append(filtered, s)
			}
		}
		allStatus = filtered
	}

	// Sort by tool name for stable output
	sort.Slice(allStatus, func(i, j int) bool {
		return allStatus[i].ToolName < allStatus[j].ToolName
	})

	return tools.JSONResult(map[string]interface{}{
		"threshold":     GlobalTracker.threshold,
		"tracked_tools": len(allStatus),
		"degraded":      len(GlobalTracker.DegradedTools()),
		"tools":         allStatus,
	}), nil
}

func handleReset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolName := tools.GetStringParam(req, "tool_name")

	if toolName != "" {
		GlobalTracker.Reset(toolName)
		return tools.JSONResult(map[string]interface{}{
			"action":    "reset",
			"tool_name": toolName,
			"status":    GlobalTracker.Status(toolName),
		}), nil
	}

	degradedBefore := len(GlobalTracker.DegradedTools())
	GlobalTracker.ResetAll()

	return tools.JSONResult(map[string]interface{}{
		"action":          "reset_all",
		"degraded_before": degradedBefore,
		"degraded_after":  0,
		"message":         fmt.Sprintf("Reset error budget for all tools (%d were degraded)", degradedBefore),
	}), nil
}
