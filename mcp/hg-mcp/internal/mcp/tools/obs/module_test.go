package obs

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestOBSStatus(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS status: %.300s", content.Text)
}

func TestOBSScenes(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleScenes(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS scenes: %.300s", content.Text)
}

func TestOBSSceneSwitchMissingScene(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSceneSwitch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing scene")
	}
}

func TestOBSSources(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSources(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS sources: %.300s", content.Text)
}

func TestOBSSourcesForScene(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"scene": "Main",
	}

	result, err := handleSources(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS scene sources: %.300s", content.Text)
}

func TestOBSSourceVisibilityMissingParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"visible": true,
	}

	result, err := handleSourceVisibility(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing scene and source")
	}
}

func TestOBSStreamMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestOBSStreamInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "invalid",
	}

	result, err := handleStream(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestOBSRecordMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleRecord(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestOBSVirtualCamMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleVirtualCam(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestOBSReplayMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleReplay(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestOBSAudio(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleAudio(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS audio: %.300s", content.Text)
}

func TestOBSMuteMissingSource(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleMute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing source")
	}
}

func TestOBSVolumeMissingSource(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"volume": -10.0,
	}

	result, err := handleVolume(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing source")
	}
}

func TestOBSStudioMode(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action": "status",
	}

	result, err := handleStudioMode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS studio mode: %.200s", content.Text)
}

func TestOBSSettings(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS settings: %.300s", content.Text)
}

func TestOBSHealth(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS health: %.300s", content.Text)
}

func TestOBSSceneItemTransformMissingParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleSceneItemTransform(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing scene and source")
	}
}

func TestOBSSceneItemTransformGet(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"scene":  "Main",
		"source": "Camera",
	}

	result, err := handleSceneItemTransform(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for get transform")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS scene item transform: %.300s", content.Text)
}

func TestOBSSceneItemTransformSet(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"scene":      "Main",
		"source":     "Camera",
		"position_x": 100.0,
		"scale_x":    2.0,
	}

	result, err := handleSceneItemTransform(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for set transform")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "Updated transform") {
		t.Errorf("expected 'Updated transform', got: %s", content.Text)
	}
}

func TestOBSFilterToggleMissingSource(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleFilterToggle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing source")
	}
}

func TestOBSFilterToggleList(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"source": "Camera",
	}

	result, err := handleFilterToggle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for list filters")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS filters: %.300s", content.Text)
}

func TestOBSFilterToggleEnable(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"source":  "Camera",
		"filter":  "Color Correction",
		"enabled": true,
	}

	result, err := handleFilterToggle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for enable filter")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "enabled") {
		t.Errorf("expected 'enabled', got: %s", content.Text)
	}
}

func TestOBSMediaControlMissingParams(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleMediaControl(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing source")
	}
}

func TestOBSMediaControlPlay(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"source": "Video File",
		"action": "play",
	}

	result, err := handleMediaControl(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for media play")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "play") {
		t.Errorf("expected 'play', got: %s", content.Text)
	}
}

func TestOBSMediaControlInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"source": "Video File",
		"action": "invalid",
	}

	result, err := handleMediaControl(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestOBSTransitionSettingsGet(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleTransitionSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for get transition")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS transition: %.300s", content.Text)
}

func TestOBSTransitionSettingsSet(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"name":     "Cut",
		"duration": 0.0,
	}

	result, err := handleTransitionSettings(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for set transition")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "Cut") {
		t.Errorf("expected 'Cut', got: %s", content.Text)
	}
}

func TestOBSScreenshot(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"format": "png",
	}

	result, err := handleScreenshot(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for screenshot")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("OBS screenshot: %.300s", content.Text)
}

func TestOBSScreenshotInvalidFormat(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"format": "bmp",
	}

	result, err := handleScreenshot(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for invalid format")
	}
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
