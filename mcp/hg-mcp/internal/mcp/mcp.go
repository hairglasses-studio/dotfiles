// Package mcp provides the MCP server setup and tool registration.
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/prompts"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/resources"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// RegisterTools registers all tools with the MCP server
func RegisterTools(s *server.MCPServer) {
	registry := tools.GetRegistry()
	registry.RegisterWithServer(s)
}

// RegisterResources registers all resources with the MCP server
func RegisterResources(s *server.MCPServer) {
	resources.RegisterResources(s)
}

// RegisterPrompts registers all prompts with the MCP server
func RegisterPrompts(s *server.MCPServer) {
	prompts.RegisterPrompts(s)
}
