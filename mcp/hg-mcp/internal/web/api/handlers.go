// Package api provides REST API handlers for the web UI.
package api

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// ToolResponse represents a tool in API responses.
type ToolResponse struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Subcategory string          `json:"subcategory"`
	Tags        []string        `json:"tags"`
	UseCases    []string        `json:"useCases"`
	Complexity  string          `json:"complexity"`
	IsWrite     bool            `json:"isWrite"`
	Parameters  []ToolParameter `json:"parameters"`
}

// ToolParameter represents a tool parameter.
type ToolParameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// CategoryResponse represents a category in API responses.
type CategoryResponse struct {
	Name          string            `json:"name"`
	Count         int               `json:"count"`
	Subcategories []SubcategoryInfo `json:"subcategories"`
}

// SubcategoryInfo represents subcategory information.
type SubcategoryInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ExecuteRequest represents a tool execution request.
type ExecuteRequest struct {
	Arguments map[string]interface{} `json:"arguments"`
}

// ExecuteResponse represents a tool execution response.
type ExecuteResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Duration int64       `json:"duration"` // milliseconds
}

// ConvertToolToResponse converts a ToolDefinition to API response format.
func ConvertToolToResponse(td tools.ToolDefinition) ToolResponse {
	params := extractParameters(td.Tool)

	return ToolResponse{
		Name:        td.Tool.Name,
		Description: td.Tool.Description,
		Category:    td.Category,
		Subcategory: td.Subcategory,
		Tags:        td.Tags,
		UseCases:    td.UseCases,
		Complexity:  string(td.Complexity),
		IsWrite:     td.IsWrite,
		Parameters:  params,
	}
}

// extractParameters extracts parameters from an MCP tool schema.
func extractParameters(tool mcp.Tool) []ToolParameter {
	var params []ToolParameter

	if tool.InputSchema.Properties == nil {
		return params
	}

	requiredSet := make(map[string]bool)
	for _, r := range tool.InputSchema.Required {
		requiredSet[r] = true
	}

	for name, prop := range tool.InputSchema.Properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}

		param := ToolParameter{
			Name:     name,
			Required: requiredSet[name],
		}

		if t, ok := propMap["type"].(string); ok {
			param.Type = t
		}
		if desc, ok := propMap["description"].(string); ok {
			param.Description = desc
		}
		if def, ok := propMap["default"]; ok {
			param.Default = def
		}
		if enum, ok := propMap["enum"].([]interface{}); ok {
			for _, e := range enum {
				if s, ok := e.(string); ok {
					param.Enum = append(param.Enum, s)
				}
			}
		}

		params = append(params, param)
	}

	return params
}

// GetCategories returns all categories with counts.
func GetCategories(registry *tools.ToolRegistry) []CategoryResponse {
	catalog := registry.GetToolCatalog()

	var categories []CategoryResponse
	for catName, subcats := range catalog {
		cat := CategoryResponse{
			Name: catName,
		}

		for subName, tools := range subcats {
			cat.Count += len(tools)
			cat.Subcategories = append(cat.Subcategories, SubcategoryInfo{
				Name:  subName,
				Count: len(tools),
			})
		}

		categories = append(categories, cat)
	}

	return categories
}

// ExecuteTool executes a tool and returns the result.
func ExecuteTool(ctx context.Context, registry *tools.ToolRegistry, toolName string, args map[string]interface{}) ExecuteResponse {
	start := time.Now()

	td, ok := registry.GetTool(toolName)
	if !ok {
		return ExecuteResponse{
			Success:  false,
			Error:    "tool not found: " + toolName,
			Duration: time.Since(start).Milliseconds(),
		}
	}

	// Build MCP request - Arguments is an interface{} that should be map[string]interface{}
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args

	// Execute
	result, err := td.Handler(ctx, req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return ExecuteResponse{
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
		}
	}

	// Extract result data
	var data interface{}
	if result != nil && len(result.Content) > 0 {
		for _, content := range result.Content {
			if textContent, ok := content.(mcp.TextContent); ok {
				// Try to parse as JSON
				var jsonData interface{}
				if err := json.Unmarshal([]byte(textContent.Text), &jsonData); err == nil {
					data = jsonData
				} else {
					data = textContent.Text
				}
				break
			}
		}
	}

	return ExecuteResponse{
		Success:  !result.IsError,
		Data:     data,
		Duration: duration,
	}
}

// SearchResult represents a search result.
type SearchResult struct {
	Tool      ToolResponse `json:"tool"`
	Score     int          `json:"score"`
	MatchType string       `json:"matchType"`
}

// SearchTools searches for tools matching a query.
func SearchTools(registry *tools.ToolRegistry, query string) []SearchResult {
	results := registry.SearchTools(query)

	var searchResults []SearchResult
	for _, r := range results {
		searchResults = append(searchResults, SearchResult{
			Tool:      ConvertToolToResponse(r.Tool),
			Score:     r.Score,
			MatchType: r.MatchType,
		})
	}

	return searchResults
}

// WorkflowResponse represents a workflow in API responses.
type WorkflowResponse struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps"`
}

// WorkflowStep represents a step in a workflow.
type WorkflowStep struct {
	Tool   string                 `json:"tool"`
	Args   map[string]interface{} `json:"args,omitempty"`
	OnFail string                 `json:"onFail,omitempty"`
}

// GetWorkflows returns all available workflows.
func GetWorkflows() []WorkflowResponse {
	// Get workflows from the workflows client
	wc := clients.GetWorkflowsClient()
	workflows := wc.List()

	var result []WorkflowResponse
	for _, wf := range workflows {
		var steps []WorkflowStep
		for _, s := range wf.Steps {
			// Convert Parameters map[string]string to map[string]interface{}
			args := make(map[string]interface{})
			for k, v := range s.Parameters {
				args[k] = v
			}
			steps = append(steps, WorkflowStep{
				Tool:   s.Tool,
				Args:   args,
				OnFail: s.OnFailure,
			})
		}
		result = append(result, WorkflowResponse{
			Name:        wf.Name,
			Description: wf.Description,
			Steps:       steps,
		})
	}

	return result
}

// RunWorkflow executes a workflow by name.
func RunWorkflow(ctx context.Context, registry *tools.ToolRegistry, name string, params map[string]interface{}) ExecuteResponse {
	start := time.Now()

	wc := clients.GetWorkflowsClient()
	result, err := wc.Run(ctx, registry, name, params)

	duration := time.Since(start).Milliseconds()

	if err != nil {
		return ExecuteResponse{
			Success:  false,
			Error:    err.Error(),
			Duration: duration,
		}
	}

	return ExecuteResponse{
		Success:  true,
		Data:     result,
		Duration: duration,
	}
}

// GetWorkflow returns a single workflow by name.
func GetWorkflow(name string) (*WorkflowResponse, error) {
	wc := clients.GetWorkflowsClient()
	wf, err := wc.GetWorkflow(nil, name)
	if err != nil {
		return nil, err
	}

	var steps []WorkflowStep
	for _, s := range wf.Steps {
		args := make(map[string]interface{})
		for k, v := range s.Parameters {
			args[k] = v
		}
		steps = append(steps, WorkflowStep{
			Tool:   s.Tool,
			Args:   args,
			OnFail: s.OnFailure,
		})
	}

	return &WorkflowResponse{
		Name:        wf.Name,
		Description: wf.Description,
		Steps:       steps,
	}, nil
}

// CreateWorkflow creates a new workflow.
func CreateWorkflow(ctx context.Context, req WorkflowResponse) error {
	wc := clients.GetWorkflowsClient()

	wf := &clients.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Category:    "custom",
		Timeout:     300,
		Steps:       make([]clients.WorkflowStep, len(req.Steps)),
	}

	for i, step := range req.Steps {
		params := make(map[string]string)
		for k, v := range step.Args {
			if s, ok := v.(string); ok {
				params[k] = s
			} else {
				// Convert non-string values to JSON
				b, _ := json.Marshal(v)
				params[k] = string(b)
			}
		}
		wf.Steps[i] = clients.WorkflowStep{
			Tool:       step.Tool,
			Parameters: params,
			OnFailure:  step.OnFail,
		}
	}

	return wc.CreateWorkflow(ctx, wf)
}

// UpdateWorkflow updates an existing workflow.
func UpdateWorkflow(ctx context.Context, name string, req WorkflowResponse) error {
	wc := clients.GetWorkflowsClient()

	wf := &clients.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Category:    "custom",
		Timeout:     300,
		Steps:       make([]clients.WorkflowStep, len(req.Steps)),
	}

	for i, step := range req.Steps {
		params := make(map[string]string)
		for k, v := range step.Args {
			if s, ok := v.(string); ok {
				params[k] = s
			} else {
				b, _ := json.Marshal(v)
				params[k] = string(b)
			}
		}
		wf.Steps[i] = clients.WorkflowStep{
			Tool:       step.Tool,
			Parameters: params,
			OnFailure:  step.OnFail,
		}
	}

	return wc.UpdateWorkflow(ctx, name, wf)
}

// DeleteWorkflow deletes a workflow.
func DeleteWorkflow(ctx context.Context, name string) error {
	wc := clients.GetWorkflowsClient()
	return wc.DeleteWorkflow(ctx, name)
}

// PreferencesResponse represents user preferences.
type PreferencesResponse struct {
	Favorites []string          `json:"favorites"`
	Aliases   map[string]string `json:"aliases"`
}

// GetPreferences returns user preferences.
func GetPreferences() PreferencesResponse {
	dc := clients.GetDiscoveryClient()
	return PreferencesResponse{
		Favorites: dc.GetFavorites(),
		Aliases:   dc.GetAliases(),
	}
}

// SetFavorites updates the favorites list.
func SetFavorites(favorites []string) {
	dc := clients.GetDiscoveryClient()
	dc.SetFavorites(favorites)
}

// SetAliases updates the aliases map.
func SetAliases(aliases map[string]string) {
	dc := clients.GetDiscoveryClient()
	dc.SetAliases(aliases)
}

// Helper to extract tool name from URL path
func ExtractToolName(path string) string {
	// /api/v1/tools/{name} -> {name}
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}
