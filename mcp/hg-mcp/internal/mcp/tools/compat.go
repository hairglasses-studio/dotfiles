// compat.go — MCP SDK compatibility / migration adapter layer.
//
// Types are aliased from mcpkit/registry (which itself aliases from mcp-go).
// Tool modules that import types from the "tools" package require zero changes.
package tools

import (
	"github.com/hairglasses-studio/mcpkit/registry"
)

// ---------------------------------------------------------------------------
// Type aliases for MCP SDK types (via mcpkit/registry).
// ---------------------------------------------------------------------------

type (
	// Tool is the MCP tool definition (name, description, input schema, annotations).
	Tool = registry.Tool

	// CallToolRequest is the incoming request when a tool is invoked.
	CallToolRequest = registry.CallToolRequest

	// CallToolResult is the response returned by a tool handler.
	CallToolResult = registry.CallToolResult

	// ToolInputSchema defines the JSON Schema for a tool's input parameters.
	ToolInputSchema = registry.ToolInputSchema

	// ToolOutputSchema defines the JSON Schema for a tool's structured output.
	ToolOutputSchema = registry.ToolOutputSchema

	// ToolAnnotation holds MCP 2025 hints (read-only, destructive, idempotent, etc.).
	ToolAnnotation = registry.ToolAnnotation

	// TextContent is a text content block within a tool result.
	TextContent = registry.TextContent

	// Content is the interface for content blocks in tool results.
	Content = registry.Content
)

// ---------------------------------------------------------------------------
// Constructor function re-exports.
// ---------------------------------------------------------------------------

var (
	// NewToolResultText creates a CallToolResult containing a single text block.
	NewToolResultText = registry.MakeTextResult

	// NewToolResultError creates a CallToolResult marked as an error.
	NewToolResultError = registry.MakeErrorResult
)
