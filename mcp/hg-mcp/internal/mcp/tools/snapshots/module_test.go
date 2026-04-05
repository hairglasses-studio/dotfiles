package snapshots

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestAutoSnapshotConfigureStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "status",
	}

	result, err := handleAutoSnapshotConfigure(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.IsError {
		t.Error("expected success for status action")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("Auto-snapshot config: %.300s", content.Text)
}

func TestAutoSnapshotConfigureEnable(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action":   "enable",
		"triggers": "pre_show,post_show",
		"max_keep": 5.0,
	}

	result, err := handleAutoSnapshotConfigure(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.IsError {
		t.Error("expected success for enable action")
	}
}

func TestAutoSnapshotConfigureDisable(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "disable",
	}

	result, err := handleAutoSnapshotConfigure(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.IsError {
		t.Error("expected success for disable action")
	}
}

func TestAutoSnapshotConfigureInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "invalid",
	}

	result, err := handleAutoSnapshotConfigure(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}
