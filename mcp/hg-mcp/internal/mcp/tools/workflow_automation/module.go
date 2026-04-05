// Package workflow_automation provides workflow automation tools for hg-mcp.
package workflow_automation

import (
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for workflow automation tools
type Module struct{}

func (m *Module) Name() string {
	return "workflow_automation"
}

func (m *Module) Description() string {
	return "Workflow automation tools: chart aggregation, release monitoring, deduplication"
}

func (m *Module) Tools() []tools.ToolDefinition {
	var allTools []tools.ToolDefinition
	allTools = append(allTools, chartTools()...)
	allTools = append(allTools, releaseMonitorTools()...)
	allTools = append(allTools, deduplicationTools()...)
	return allTools
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
