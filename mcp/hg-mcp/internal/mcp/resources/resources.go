// Package resources provides MCP resource handlers for exposing data to AI clients.
package resources

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// RegisterResources registers all MCP resources with the server
func RegisterResources(s *server.MCPServer) {
	// Tool catalog resource
	s.AddResource(
		mcp.NewResource(
			"aftrs://tools/catalog",
			"Tool Catalog",
			mcp.WithResourceDescription("Complete catalog of all AFTRS MCP tools with metadata"),
			mcp.WithMIMEType("application/json"),
		),
		handleToolCatalog,
	)

	// Tool categories resource
	s.AddResource(
		mcp.NewResource(
			"aftrs://tools/categories",
			"Tool Categories",
			mcp.WithResourceDescription("List of tool categories and counts"),
			mcp.WithMIMEType("application/json"),
		),
		handleToolCategories,
	)

	// System health resource
	s.AddResource(
		mcp.NewResource(
			"aftrs://health/systems",
			"System Health",
			mcp.WithResourceDescription("Health status of all connected systems"),
			mcp.WithMIMEType("application/json"),
		),
		handleSystemHealth,
	)

	// Tool schema template resource
	s.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"aftrs://tools/schema/{name}",
			"Tool Schema",
			mcp.WithTemplateDescription("JSON schema for a specific tool"),
			mcp.WithTemplateMIMEType("application/json"),
		),
		handleToolSchema,
	)
}

// handleToolCatalog returns the complete tool catalog
func handleToolCatalog(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	registry := tools.GetRegistry()

	catalog := struct {
		TotalTools   int             `json:"total_tools"`
		TotalModules int             `json:"total_modules"`
		Modules      []ModuleSummary `json:"modules"`
	}{
		TotalTools:   registry.ToolCount(),
		TotalModules: registry.ModuleCount(),
		Modules:      make([]ModuleSummary, 0),
	}

	// Get all modules and their tools
	for _, modName := range registry.ListModules() {
		mod, ok := registry.GetModule(modName)
		if !ok {
			continue
		}

		summary := ModuleSummary{
			Name:        mod.Name(),
			Description: mod.Description(),
			Tools:       make([]ToolSummary, 0),
		}

		for _, tool := range mod.Tools() {
			summary.Tools = append(summary.Tools, ToolSummary{
				Name:        tool.Tool.Name,
				Description: tool.Tool.Description,
				Category:    tool.Category,
				Subcategory: tool.Subcategory,
				Complexity:  string(tool.Complexity),
				IsWrite:     tool.IsWrite,
				Tags:        tool.Tags,
			})
		}
		summary.ToolCount = len(summary.Tools)
		catalog.Modules = append(catalog.Modules, summary)
	}

	data, _ := json.MarshalIndent(catalog, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

// handleToolCategories returns tool categories with counts
func handleToolCategories(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	registry := tools.GetRegistry()
	stats := registry.GetToolStats()

	categories := struct {
		Categories map[string]int `json:"categories"`
		Complexity map[string]int `json:"complexity"`
		Total      int            `json:"total"`
	}{
		Categories: stats.ByCategory,
		Complexity: stats.ByComplexity,
		Total:      stats.TotalTools,
	}

	data, _ := json.MarshalIndent(categories, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

// handleSystemHealth returns health status of all systems
func handleSystemHealth(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// This would ideally call each system's health check
	// For now, return a structure indicating which systems are available
	health := struct {
		Systems []SystemHealthSummary `json:"systems"`
	}{
		Systems: []SystemHealthSummary{
			{Name: "ableton", Type: "daw", Protocol: "osc"},
			{Name: "resolume", Type: "vj", Protocol: "osc+http"},
			{Name: "grandma3", Type: "lighting", Protocol: "osc"},
			{Name: "obs", Type: "streaming", Protocol: "websocket"},
			{Name: "touchdesigner", Type: "generative", Protocol: "http"},
			{Name: "atem", Type: "switcher", Protocol: "udp"},
			{Name: "showkontrol", Type: "timecode", Protocol: "http"},
		},
	}

	data, _ := json.MarshalIndent(health, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

// handleToolSchema returns the schema for a specific tool
func handleToolSchema(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract tool name from URI
	// URI format: aftrs://tools/schema/{name}
	uri := req.Params.URI
	toolName := extractToolName(uri)

	registry := tools.GetRegistry()

	// Find the tool directly
	tool, ok := registry.GetTool(toolName)
	if ok {
		schema := ToolSchema{
			Name:        tool.Tool.Name,
			Description: tool.Tool.Description,
			InputSchema: tool.Tool.InputSchema,
			Category:    tool.Category,
			Subcategory: tool.Subcategory,
			Tags:        tool.Tags,
			UseCases:    tool.UseCases,
			Complexity:  string(tool.Complexity),
			IsWrite:     tool.IsWrite,
		}

		data, _ := json.MarshalIndent(schema, "", "  ")
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     string(data),
			},
		}, nil
	}

	// Tool not found
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     `{"error": "tool not found"}`,
		},
	}, nil
}

// extractToolName extracts the tool name from a schema URI
func extractToolName(uri string) string {
	// URI format: aftrs://tools/schema/{name}
	const prefix = "aftrs://tools/schema/"
	if len(uri) > len(prefix) {
		return uri[len(prefix):]
	}
	return ""
}

// ModuleSummary represents a module in the catalog
type ModuleSummary struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	ToolCount   int           `json:"tool_count"`
	Tools       []ToolSummary `json:"tools"`
}

// ToolSummary represents a tool in the catalog
type ToolSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Subcategory string   `json:"subcategory,omitempty"`
	Complexity  string   `json:"complexity"`
	IsWrite     bool     `json:"is_write"`
	Tags        []string `json:"tags,omitempty"`
}

// ToolSchema represents the full schema for a tool
type ToolSchema struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	InputSchema mcp.ToolInputSchema `json:"input_schema"`
	Category    string              `json:"category"`
	Subcategory string              `json:"subcategory,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	UseCases    []string            `json:"use_cases,omitempty"`
	Complexity  string              `json:"complexity"`
	IsWrite     bool                `json:"is_write"`
}

// SystemHealthSummary represents a system's health info
type SystemHealthSummary struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}
