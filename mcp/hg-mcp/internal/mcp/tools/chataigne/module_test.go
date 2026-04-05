package chataigne

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideChataigneClient = clients.NewTestChataigneClient()
}

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleStatus(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.ChataigneClient, error) {
		return clients.NewTestChataigneClient(), nil
	}
	defer func() { getClient = origGetClient }()

	result, err := handleStatus(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("nil result")
	}
}

func TestHandleState(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.ChataigneClient, error) {
		return clients.NewTestChataigneClient(), nil
	}
	defer func() { getClient = origGetClient }()

	tests := []struct {
		name    string
		action  string
		state   string
		wantErr bool
	}{
		{"missing action", "", "", true},
		{"set without state name", "set", "", true},
		{"invalid action", "reset", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			if tt.action != "" {
				args["action"] = tt.action
			}
			if tt.state != "" {
				args["state"] = tt.state
			}
			req := mcp.CallToolRequest{}
			req.Params.Arguments = args

			result, err := handleState(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && !result.IsError {
				t.Errorf("expected error result")
			}
		})
	}
}

func TestHandleSequence(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.ChataigneClient, error) {
		return clients.NewTestChataigneClient(), nil
	}
	defer func() { getClient = origGetClient }()

	// Test name with missing action
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{"name": "intro"}
	result, err := handleSequence(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing action")
	}

	// Test invalid action
	req.Params.Arguments = map[string]interface{}{"name": "intro", "action": "pause"}
	result, err = handleSequence(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid action")
	}
}

func TestToolsRegistered(t *testing.T) {
	registry := tools.GetRegistry()
	count := 0
	for _, td := range registry.GetAllToolDefinitions() {
		if td.Subcategory == "chataigne" {
			count++
		}
	}
	if count != 5 {
		t.Errorf("expected 5 chataigne tools in registry, found %d", count)
	}
}
