package puredata

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverridePureDataClient = clients.NewTestPureDataClient()
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
	if !strings.Contains(content.Text, "Pure Data") {
		t.Errorf("expected 'Pure Data' in output, got: %.200s", content.Text)
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

func TestHandleSend(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"message": "freq 440",
	}

	result, err := handleSend(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "freq 440") {
		t.Errorf("expected 'freq 440', got: %s", content.Text)
	}
}

func TestHandleSendMissingMessage(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSend(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing message")
	}
}

func TestHandleDSPOn(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"on": true,
	}

	result, err := handleDSP(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "on") {
		t.Errorf("expected 'on', got: %s", content.Text)
	}
}

func TestHandleDSPOff(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"on": false,
	}

	result, err := handleDSP(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "off") {
		t.Errorf("expected 'off', got: %s", content.Text)
	}
}
