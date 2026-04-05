package slack

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideSlackClient = clients.NewTestSlackClient()
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
	if !strings.Contains(content.Text, "Slack") {
		t.Errorf("expected 'Slack' in output, got: %.200s", content.Text)
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
		"message": "Hello from test",
	}

	result, err := handleSend(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "sent") {
		t.Errorf("expected 'sent', got: %s", content.Text)
	}
}

func TestHandleSendWithChannel(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"message": "Hello",
		"channel": "#alerts",
	}

	result, err := handleSend(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "#alerts") {
		t.Errorf("expected '#alerts', got: %s", content.Text)
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
