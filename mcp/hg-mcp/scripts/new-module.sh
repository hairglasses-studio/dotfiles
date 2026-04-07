#!/bin/bash
# new-module.sh — Scaffold a new hg-mcp tool module
#
# Usage: ./scripts/new-module.sh <package_name> <module_name> <category> [description]
# Example: ./scripts/new-module.sh bandcamp "Bandcamp" music "Bandcamp music store integration"

set -euo pipefail

PKG="${1:?Usage: $0 <package> <module_name> <category> [description]}"
MODULE="${2:?Usage: $0 <package> <module_name> <category> [description]}"
CATEGORY="${3:?Usage: $0 <package> <module_name> <category> [description]}"
DESC="${4:-$MODULE tools}"

DIR="internal/mcp/tools/${PKG}"
PREFIX="aftrs_${PKG}"

if [ -d "$DIR" ]; then
  echo "Error: $DIR already exists"
  exit 1
fi

mkdir -p "$DIR"

# Generate module.go
cat > "${DIR}/module.go" << GOEOF
// Package ${PKG} provides MCP tools for ${MODULE}.
package ${PKG}

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the ${MODULE} tools module.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "${PKG}" }
func (m *Module) Description() string { return "${DESC}" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.Tool{
				Name:        "${PREFIX}_status",
				Description: "Get ${MODULE} status",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: map[string]interface{}{},
				},
			},
			Handler:    handleStatus,
			Category:   "${CATEGORY}",
			Complexity: tools.ComplexitySimple,
		},
	}
}

func handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return tools.JSONResult(map[string]interface{}{
		"module": "${PKG}",
		"status": "ok",
	}), nil
}
GOEOF

# Generate module_test.go
cat > "${DIR}/module_test.go" << GOEOF
package ${PKG}

import (
	"testing"
)

func TestModuleRegistration(t *testing.T) {
	m := &Module{}
	if m.Name() == "" {
		t.Fatal("module name must not be empty")
	}
	if m.Description() == "" {
		t.Fatal("module description must not be empty")
	}
	toolDefs := m.Tools()
	if len(toolDefs) == 0 {
		t.Fatal("module must define at least one tool")
	}
	for _, td := range toolDefs {
		if td.Tool.Name == "" {
			t.Error("tool name must not be empty")
		}
		if td.Handler == nil {
			t.Errorf("tool %s has nil handler", td.Tool.Name)
		}
	}
}
GOEOF

echo "Created ${DIR}/module.go and ${DIR}/module_test.go"
echo "Don't forget to add the import to cmd/hg-mcp/main.go:"
echo "  _ \"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/${PKG}\""
