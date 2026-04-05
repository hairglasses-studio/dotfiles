package supercollider

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideSuperColliderClient = clients.NewTestSuperColliderClient()
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
	if !strings.Contains(content.Text, "SuperCollider") {
		t.Errorf("expected 'SuperCollider' in output, got: %.200s", content.Text)
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

func TestHandleSynthCreate(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action":   "create",
		"def_name": "default",
		"node_id":  1000.0,
	}

	result, err := handleSynth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success for synth create")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "Created synth") {
		t.Errorf("expected 'Created synth', got: %s", content.Text)
	}
}

func TestHandleSynthFree(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action":  "free",
		"node_id": 1000.0,
	}

	result, err := handleSynth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success for synth free")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "Freed node") {
		t.Errorf("expected 'Freed node', got: %s", content.Text)
	}
}

func TestHandleSynthMissingAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"node_id": 1000.0,
	}

	result, err := handleSynth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing action")
	}
}

func TestHandleSynthInvalidAction(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action":  "invalid",
		"node_id": 1000.0,
	}

	result, err := handleSynth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestHandleSynthCreateMissingDefName(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"action":  "create",
		"node_id": 1000.0,
	}

	result, err := handleSynth(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing def_name")
	}
}

func TestHandleNode(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"node_id": 1000.0,
		"param":   "freq",
		"value":   440.0,
	}

	result, err := handleNode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	if !strings.Contains(content.Text, "freq") {
		t.Errorf("expected 'freq' in output, got: %s", content.Text)
	}
}

func TestHandleNodeMissingParam(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"node_id": 1000.0,
		"value":   440.0,
	}

	result, err := handleNode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing param")
	}
}

func TestHandleNodeInvalidNodeID(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"node_id": 0.0,
		"param":   "freq",
		"value":   440.0,
	}

	result, err := handleNode(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid node_id")
	}
}

func TestHandleEval(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"code": "{SinOsc.ar(440)}.play",
	}

	result, err := handleEval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("Eval result: %s", content.Text)
}

func TestHandleEvalMissingCode(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleEval(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing code")
	}
}
