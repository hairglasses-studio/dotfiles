package qlcplus

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	// Override with test client to avoid network calls
	clients.TestOverrideQLCPlusClient = clients.NewTestQLCPlusClient()
}

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleStatus(t *testing.T) {
	// Override getClient to use test client
	origGetClient := getClient
	getClient = func() (*clients.QLCPlusClient, error) {
		return clients.NewTestQLCPlusClient(), nil
	}
	defer func() { getClient = origGetClient }()

	req := mcp.CallToolRequest{}
	result, err := handleStatus(context.Background(), req)
	if err != nil {
		t.Fatalf("handleStatus() returned unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("handleStatus() returned nil result")
	}
}

func TestHandleHealth(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.QLCPlusClient, error) {
		return clients.NewTestQLCPlusClient(), nil
	}
	defer func() { getClient = origGetClient }()

	req := mcp.CallToolRequest{}
	result, err := handleHealth(context.Background(), req)
	if err != nil {
		t.Fatalf("handleHealth() returned unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("handleHealth() returned nil result")
	}
}

func TestHandleFunction(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.QLCPlusClient, error) {
		return clients.NewTestQLCPlusClient(), nil
	}
	defer func() { getClient = origGetClient }()

	tests := []struct {
		name       string
		functionID string
		action     string
		wantErr    bool
	}{
		// start/stop return errors because test client is not actually reachable
		{"start function (unreachable)", "1", "start", true},
		{"stop function (unreachable)", "1", "stop", true},
		{"missing function_id", "", "start", true},
		{"missing action", "1", "", true},
		{"invalid action", "1", "pause", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			if tt.functionID != "" {
				args["function_id"] = tt.functionID
			}
			if tt.action != "" {
				args["action"] = tt.action
			}

			req := mcp.CallToolRequest{}
			req.Params.Arguments = args

			result, err := handleFunction(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("nil result")
			}

			isError := result.IsError
			if tt.wantErr && !isError {
				t.Errorf("expected error result but got success")
			}
			if !tt.wantErr && isError {
				t.Errorf("expected success but got error: %v", result.Content)
			}
		})
	}
}

func TestParseChannelValuePairs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{"single pair", "1:255", 1, false},
		{"multiple pairs", "1:255,2:128,3:0", 3, false},
		{"with spaces", "1:255, 2:128", 2, false},
		{"invalid format", "1-255", 0, true},
		{"channel out of range", "0:255", 0, true},
		{"channel too high", "513:255", 0, true},
		{"value out of range", "1:256", 0, true},
		{"negative value", "1:-1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseChannelValuePairs(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result) != tt.wantLen {
				t.Errorf("got %d pairs, want %d", len(result), tt.wantLen)
			}
		})
	}
}

func TestHandleWidgetSetValidation(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.QLCPlusClient, error) {
		return clients.NewTestQLCPlusClient(), nil
	}
	defer func() { getClient = origGetClient }()

	tests := []struct {
		name     string
		widgetID string
		value    string
		wantErr  bool
	}{
		// valid params still returns error because test client is not reachable
		{"valid params (unreachable)", "1", "255", true},
		{"missing widget_id", "", "255", true},
		{"missing value", "1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{}
			if tt.widgetID != "" {
				args["widget_id"] = tt.widgetID
			}
			if tt.value != "" {
				args["value"] = tt.value
			}

			req := mcp.CallToolRequest{}
			req.Params.Arguments = args

			result, err := handleWidgetSet(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			isError := result.IsError
			if tt.wantErr && !isError {
				t.Errorf("expected error result")
			}
			if !tt.wantErr && isError {
				t.Errorf("expected success but got error")
			}
		})
	}
}

// Verify tool count is discoverable in registry
func TestToolsRegistered(t *testing.T) {
	registry := tools.GetRegistry()
	count := 0
	for _, td := range registry.GetAllToolDefinitions() {
		if td.Category == "lighting" && td.Subcategory == "qlcplus" {
			count++
		}
	}
	if count != 7 {
		t.Errorf("expected 7 qlcplus tools in registry, found %d", count)
	}
}
