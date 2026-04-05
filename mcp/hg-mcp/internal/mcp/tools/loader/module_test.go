package loader

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func TestModuleInfo(t *testing.T) {
	m := &Module{}
	if m.Name() != "loader" {
		t.Errorf("Name() = %q, want %q", m.Name(), "loader")
	}
	if m.Description() == "" {
		t.Error("Description() should not be empty")
	}

	toolDefs := m.Tools()
	if len(toolDefs) != 1 {
		t.Fatalf("Tools() returned %d tools, want 1", len(toolDefs))
	}
	if toolDefs[0].Tool.Name != "hg_load_domain" {
		t.Errorf("tool name = %q, want %q", toolDefs[0].Tool.Name, "hg_load_domain")
	}
	if toolDefs[0].RuntimeGroup != tools.RuntimeGroupPlatform {
		t.Errorf("RuntimeGroup = %q, want %q", toolDefs[0].RuntimeGroup, tools.RuntimeGroupPlatform)
	}
}

func TestHandleLoadDomain_ListDeferred(t *testing.T) {
	// Calling without a domain should list deferred domains
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleLoadDomain(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	// The result should contain profile information
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		if !strings.Contains(tc.Text, "profile") && !strings.Contains(tc.Text, "Profile") {
			t.Errorf("expected profile info in response, got: %s", tc.Text[:min(100, len(tc.Text))])
		}
	}
}

func TestHandleLoadDomain_UnknownDomain(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"domain": "nonexistent_domain_xyz",
	}

	result, err := handleLoadDomain(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	// Should either load 0 tools or return an error about unknown domain
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		if !strings.Contains(tc.Text, "unknown") && !strings.Contains(tc.Text, "already loaded") {
			t.Logf("response: %s", tc.Text)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
