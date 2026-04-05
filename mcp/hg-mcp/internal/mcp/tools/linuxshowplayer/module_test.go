package linuxshowplayer

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideLinuxShowPlayerClient = clients.NewTestLinuxShowPlayerClient()
}

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "Linux Show Player") {
		t.Errorf("expected 'Linux Show Player' in output, got: %.200s", content.Text)
	}
}

func TestHandleHealth(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestHandleCueGo(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "go",
	}

	result, err := handleCue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "go") {
		t.Errorf("expected 'go', got: %s", content.Text)
	}
}

func TestHandleCueStartSpecific(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "start",
		"cue":    5.0,
	}

	result, err := handleCue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "5") {
		t.Errorf("expected cue 5, got: %s", content.Text)
	}
}

func TestHandleCueMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleCue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestHandleCueInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "invalid",
	}

	result, err := handleCue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestHandleCueStartMissingNumber(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "start",
	}

	result, err := handleCue(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing cue number")
	}
}
