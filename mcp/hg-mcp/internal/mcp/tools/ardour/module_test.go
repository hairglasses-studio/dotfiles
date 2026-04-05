package ardour

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideArdourClient = clients.NewTestArdourClient()
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
	if !strings.Contains(content.Text, "Ardour") {
		t.Errorf("expected 'Ardour' in output, got: %.200s", content.Text)
	}
	t.Logf("Status: %.300s", content.Text)
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
	content := result.Content[0].(mcp.TextContent)
	t.Logf("Health: %.300s", content.Text)
}

func TestHandleTransportPlay(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "play",
	}

	result, err := handleTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "play") {
		t.Errorf("expected 'play', got: %s", content.Text)
	}
}

func TestHandleTransportStop(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "stop",
	}

	result, err := handleTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
}

func TestHandleTransportLocate(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "locate",
		"frame":  48000.0,
	}

	result, err := handleTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "48000") {
		t.Errorf("expected frame 48000, got: %s", content.Text)
	}
}

func TestHandleTransportMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestHandleTransportInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "invalid",
	}

	result, err := handleTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestHandleTrackSetFader(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"strip_id": 0.0,
		"fader":    0.75,
	}

	result, err := handleTrack(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "fader") {
		t.Errorf("expected 'fader', got: %s", content.Text)
	}
}

func TestHandleTrackMissingStripID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleTrack(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing strip_id")
	}
}

func TestHandleMeter(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"strip_id": 0.0,
	}

	result, err := handleMeter(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("Meter: %s", content.Text)
}

func TestHandleMeterMissingStripID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleMeter(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing strip_id")
	}
}
