package companion

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideCompanionClient = clients.NewTestCompanionClient()
}

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleStatus(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.CompanionClient, error) {
		return clients.NewTestCompanionClient(), nil
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

func TestHandleButtonPress(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.CompanionClient, error) {
		return clients.NewTestCompanionClient(), nil
	}
	defer func() { getClient = origGetClient }()

	tests := []struct {
		name    string
		page    float64
		bank    float64
		wantErr bool
	}{
		// Valid params return error because test client is not reachable
		{"valid (unreachable)", 1, 1, true},
		{"page out of range", 0, 1, true},
		{"page too high", 100, 1, true},
		{"bank out of range", 1, 0, true},
		{"bank too high", 1, 33, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{}
			req.Params.Arguments = map[string]interface{}{
				"page": tt.page,
				"bank": tt.bank,
			}

			result, err := handleButtonPress(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr && !result.IsError {
				t.Errorf("expected error result")
			}
		})
	}
}

func TestHandleVariableValidation(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.CompanionClient, error) {
		return clients.NewTestCompanionClient(), nil
	}
	defer func() { getClient = origGetClient }()

	// Test missing name
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleVariableGet(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing name")
	}

	// Test variable set missing value
	req.Params.Arguments = map[string]interface{}{"name": "test"}
	result, err = handleVariableSet(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for missing value")
	}
}

func TestToolsRegistered(t *testing.T) {
	registry := tools.GetRegistry()
	count := 0
	for _, td := range registry.GetAllToolDefinitions() {
		if td.Category == "automation" && td.Subcategory == "companion" {
			count++
		}
	}
	if count != 7 {
		t.Errorf("expected 7 companion tools in registry, found %d", count)
	}
}
