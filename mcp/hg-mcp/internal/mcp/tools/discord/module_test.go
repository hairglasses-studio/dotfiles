package discord

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestDiscordStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Discord status: %.200s", content.Text)
}

func TestDiscordSendMissingMessage(t *testing.T) {
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

func TestDiscordChannels(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"type": "text",
	}

	result, err := handleChannels(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Discord channels: %.200s", content.Text)
}

func TestDiscordHistory(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"limit": 10.0,
	}

	result, err := handleHistory(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Discord history: %.200s", content.Text)
}

func TestDiscordThreadMissingID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleThread(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing thread_id")
	}
}

func TestDiscordNotifyMissingTitle(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"message": "test message",
	}

	result, err := handleNotify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing title")
	}
}

func TestDiscordNotifyMissingMessage(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"title": "Test Title",
	}

	result, err := handleNotify(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing message")
	}
}

func TestDiscordWebhookMissingContent(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing content/title/description")
	}
}

func TestDiscordReactMissingMessageID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"emoji": "👍",
	}

	result, err := handleReact(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing message_id")
	}
}

func TestDiscordEditMissingMessageID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"content": "new content",
	}

	result, err := handleEdit(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing message_id")
	}
}

func TestDiscordDeleteMissingMessageID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleDelete(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing message_id")
	}
}

func TestDiscordThreadCreateMissingName(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleThreadCreate(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing name")
	}
}

func TestDiscordStudioEventMissingType(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"title": "Test Event",
	}

	result, err := handleStudioEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing event_type")
	}
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
