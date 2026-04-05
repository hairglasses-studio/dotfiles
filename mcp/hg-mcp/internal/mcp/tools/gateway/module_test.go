package gateway

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestDJGatewayStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"software": "serato",
		"action":   "status",
	}

	result, err := handleDJGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("DJ Serato status: %.300s", content.Text)
}

func TestDJGatewaySearch(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"software": "rekordbox",
		"action":   "search",
		"query":    "techno",
		"limit":    5.0,
	}

	result, err := handleDJGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("DJ Rekordbox search: %.300s", content.Text)
}

func TestDJGatewayMissingSoftware(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "status",
	}

	result, err := handleDJGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing software")
	}
}

func TestAVGatewayStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"software": "obs",
		"action":   "status",
	}

	result, err := handleAVGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("AV OBS status: %.300s", content.Text)
}

func TestAVGatewayScenes(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"software": "obs",
		"action":   "scenes",
	}

	result, err := handleAVGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("AV OBS scenes: %.300s", content.Text)
}

func TestAVGatewayResolumeLayers(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"software": "resolume",
		"action":   "layers",
	}

	result, err := handleAVGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("AV Resolume layers: %.300s", content.Text)
}

func TestLightingGatewayStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"system": "wled",
		"action": "status",
	}

	result, err := handleLightingGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Lighting WLED status: %.300s", content.Text)
}

func TestLightingGatewayFixtures(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"system": "dmx",
		"action": "fixtures",
	}

	result, err := handleLightingGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Lighting DMX fixtures: %.300s", content.Text)
}

func TestAudioGatewayStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"system": "ableton",
		"action": "status",
	}

	result, err := handleAudioGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Audio Ableton status: %.300s", content.Text)
}

func TestAudioGatewayTracks(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"system": "ableton",
		"action": "tracks",
	}

	result, err := handleAudioGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Audio Ableton tracks: %.300s", content.Text)
}

func TestStreamingGatewayStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"platform": "twitch",
		"action":   "status",
	}

	result, err := handleStreamingGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Streaming Twitch status: %.300s", content.Text)
}

func TestStreamingGatewayYouTube(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"platform": "youtube",
		"action":   "status",
	}

	result, err := handleStreamingGateway(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Streaming YouTube status: %.300s", content.Text)
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
