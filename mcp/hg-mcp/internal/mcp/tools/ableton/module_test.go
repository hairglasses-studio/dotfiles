package ableton

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestAbletonStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton status: %.300s", content.Text)
}

func TestAbletonTransportMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestAbletonTransportInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "invalid_action",
	}

	result, err := handleAbletonTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestAbletonTempo(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonTempo(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton tempo: %.200s", content.Text)
}

func TestAbletonTracks(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonTracks(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton tracks: %.300s", content.Text)
}

func TestAbletonTrackMissingIndex(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonTrack(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing index")
	}
}

func TestAbletonTrack(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"index": 0.0,
	}

	result, err := handleAbletonTrack(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton track: %.300s", content.Text)
}

func TestAbletonClipsMissingTrack(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonClips(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing track_index")
	}
}

func TestAbletonClips(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"track_index": 0.0,
	}

	result, err := handleAbletonClips(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton clips: %.300s", content.Text)
}

func TestAbletonClipFireMissingTrack(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"slot_index": 0.0,
	}

	result, err := handleAbletonClipFire(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing track_index")
	}
}

func TestAbletonClipFireMissingSlot(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"track_index": 0.0,
	}

	result, err := handleAbletonClipFire(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing slot_index")
	}
}

func TestAbletonSceneFireMissingScene(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonSceneFire(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing scene_index")
	}
}

func TestAbletonDevicesMissingTrack(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing track_index")
	}
}

func TestAbletonDevices(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"track_index": 0.0,
	}

	result, err := handleAbletonDevices(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton devices: %.300s", content.Text)
}

func TestAbletonDeviceMissingParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonDevice(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing track_index")
	}
}

func TestAbletonCuePoints(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonCuePoints(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton cue points: %.300s", content.Text)
}

func TestAbletonHealth(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAbletonHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Ableton health: %.300s", content.Text)
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
