package opc

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideOPCClient = clients.NewTestOPCClient()
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
	if !strings.Contains(content.Text, "Open Pixel Control") {
		t.Errorf("expected 'Open Pixel Control' in output, got: %.200s", content.Text)
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

func TestHandleSetPixels(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"channel": 0.0,
		"pixels":  "255,0,0,0,255,0,0,0,255",
	}

	result, err := handleSetPixels(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "3 pixels") {
		t.Errorf("expected '3 pixels', got: %s", content.Text)
	}
}

func TestHandleSetPixelsMissing(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSetPixels(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing pixels")
	}
}

func TestHandleSetPixelsInvalidValue(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"pixels": "255,abc,0",
	}

	result, err := handleSetPixels(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid value")
	}
}

func TestHandleSetPixelsOutOfRange(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"pixels": "256,0,0",
	}

	result, err := handleSetPixels(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for out-of-range value")
	}
}
