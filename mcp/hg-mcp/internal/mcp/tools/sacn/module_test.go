package sacn

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func init() {
	clients.TestOverrideSACNClient = clients.NewTestSACNClient()
}

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestHandleStatus(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.SACNClient, error) {
		return clients.NewTestSACNClient(), nil
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

func TestHandleHealth(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.SACNClient, error) {
		return clients.NewTestSACNClient(), nil
	}
	defer func() { getClient = origGetClient }()

	result, err := handleHealth(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("nil result")
	}
}

func TestHandleSend(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.SACNClient, error) {
		return clients.NewTestSACNClient(), nil
	}
	defer func() { getClient = origGetClient }()

	tests := []struct {
		name     string
		universe float64
		values   string
		wantErr  bool
	}{
		{"valid", 1, "1:255,2:128", false},
		{"universe out of range", 0, "1:255", true},
		{"universe too high", 64000, "1:255", true},
		{"missing values", 1, "", true},
		{"invalid format", 1, "bad", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := map[string]interface{}{
				"universe": tt.universe,
			}
			if tt.values != "" {
				args["values"] = tt.values
			}

			req := mcp.CallToolRequest{}
			req.Params.Arguments = args

			result, err := handleSend(context.Background(), req)
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

func TestHandleDiscover(t *testing.T) {
	origGetClient := getClient
	getClient = func() (*clients.SACNClient, error) {
		return clients.NewTestSACNClient(), nil
	}
	defer func() { getClient = origGetClient }()

	result, err := handleDiscover(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("nil result")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseChannelValuePairs(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
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

func TestToolsRegistered(t *testing.T) {
	registry := tools.GetRegistry()
	count := 0
	for _, td := range registry.GetAllToolDefinitions() {
		if td.Category == "lighting" && td.Subcategory == "sacn" {
			count++
		}
	}
	if count != 4 {
		t.Errorf("expected 4 sacn tools in registry, found %d", count)
	}
}
