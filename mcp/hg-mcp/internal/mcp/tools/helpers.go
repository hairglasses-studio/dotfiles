// Package tools provides the core tool registry and interfaces for the hg-mcp server.
//
// Helper functions delegate to mcpkit/handler, maintaining identical signatures
// so that all 119 tool modules require zero changes.
package tools

import (
	"fmt"

	"github.com/hairglasses-studio/mcpkit/handler"
	"github.com/mark3labs/mcp-go/mcp"
)

// TextResult creates a text result for a tool response.
func TextResult(text string) *mcp.CallToolResult { return handler.TextResult(text) }

// ErrorResult creates an error result for a tool response.
func ErrorResult(err error) *mcp.CallToolResult { return handler.ErrorResult(err) }

// JSONResult creates a JSON result for a tool response.
func JSONResult(data interface{}) *mcp.CallToolResult { return handler.JSONResult(data) }

// GetStringParam extracts a string parameter from the request.
func GetStringParam(req mcp.CallToolRequest, name string) string {
	return handler.GetStringParam(req, name)
}

// GetIntParam extracts an integer parameter from the request.
func GetIntParam(req mcp.CallToolRequest, name string, defaultVal int) int {
	return handler.GetIntParam(req, name, defaultVal)
}

// GetBoolParam extracts a boolean parameter from the request.
func GetBoolParam(req mcp.CallToolRequest, name string, defaultVal bool) bool {
	return handler.GetBoolParam(req, name, defaultVal)
}

// GetFloatParam extracts a float64 parameter from the request.
func GetFloatParam(req mcp.CallToolRequest, name string, defaultVal float64) float64 {
	return handler.GetFloatParam(req, name, defaultVal)
}

// HasParam checks if a parameter was explicitly provided in the request.
func HasParam(req mcp.CallToolRequest, name string) bool {
	return handler.HasParam(req, name)
}

// GetStringArrayParam extracts a string array parameter from the request.
func GetStringArrayParam(req mcp.CallToolRequest, name string) []string {
	return handler.GetStringArrayParam(req, name)
}

// RequireStringParam extracts a string parameter and returns an error result if empty.
func RequireStringParam(req mcp.CallToolRequest, name string) (string, *mcp.CallToolResult) {
	val := GetStringParam(req, name)
	if val == "" {
		return "", ErrorResult(fmt.Errorf("%s is required", name))
	}
	return val, nil
}

// RequireIntParam extracts an int parameter that must be > 0.
// Returns the value and nil on success, or (0, *CallToolResult) on missing/zero.
func RequireIntParam(req mcp.CallToolRequest, name string) (int, *mcp.CallToolResult) {
	val := GetIntParam(req, name, 0)
	if val == 0 {
		return 0, ErrorResult(fmt.Errorf("%s is required and must be > 0", name))
	}
	return val, nil
}

// RequireFloatParam extracts a float64 parameter that must be > 0.
// Returns the value and nil on success, or (0, *CallToolResult) on missing/zero.
func RequireFloatParam(req mcp.CallToolRequest, name string) (float64, *mcp.CallToolResult) {
	val := GetFloatParam(req, name, 0)
	if val == 0 {
		return 0, ErrorResult(fmt.Errorf("%s is required and must be > 0", name))
	}
	return val, nil
}

// RequireStringArrayParam extracts a non-empty string array parameter.
// Returns the value and nil on success, or (nil, *CallToolResult) on missing/empty.
func RequireStringArrayParam(req mcp.CallToolRequest, name string) ([]string, *mcp.CallToolResult) {
	val := GetStringArrayParam(req, name)
	if len(val) == 0 {
		return nil, ErrorResult(fmt.Errorf("%s is required and must not be empty", name))
	}
	return val, nil
}

// OptionalStringParam extracts an optional string parameter with a default.
func OptionalStringParam(req mcp.CallToolRequest, name string, defaultVal string) string {
	val := GetStringParam(req, name)
	if val == "" {
		return defaultVal
	}
	return val
}

// Structured error codes for programmatic categorization.
const (
	ErrClientInit  = handler.ErrClientInit
	ErrInvalidParam = handler.ErrInvalidParam
	ErrTimeout     = handler.ErrTimeout
	ErrNotFound    = handler.ErrNotFound
	ErrAPIError    = handler.ErrAPIError
	ErrPermission  = handler.ErrPermission
)

// CodedErrorResult creates an error result with a structured error code prefix.
func CodedErrorResult(code string, err error) *mcp.CallToolResult {
	return handler.CodedErrorResult(code, err)
}

// ActionableErrorResult creates an error result with suggestions for resolution.
func ActionableErrorResult(code string, err error, suggestions ...string) *mcp.CallToolResult {
	return handler.ActionableErrorResult(code, err, suggestions...)
}

// MaxResponseSize is the maximum response size before truncation (128KB).
const MaxResponseSize = 128 * 1024

// truncateResponse truncates a tool result's text content if it exceeds MaxResponseSize.
func truncateResponse(result *mcp.CallToolResult) *mcp.CallToolResult {
	if result == nil {
		return result
	}
	for i, content := range result.Content {
		if tc, ok := content.(mcp.TextContent); ok {
			if len(tc.Text) > MaxResponseSize {
				tc.Text = tc.Text[:MaxResponseSize] + "\n\n[TRUNCATED: response exceeded 128KB limit]"
				result.Content[i] = tc
			}
		}
	}
	return result
}

// ObjectOutputSchema creates an output schema for tools returning JSON objects.
func ObjectOutputSchema(properties map[string]interface{}, required []string) *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type:       "object",
		Properties: properties,
		Required:   required,
	}
}
