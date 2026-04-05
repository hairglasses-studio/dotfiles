package graph

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGraphInsights(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleGraphInsights(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Graph insights: %.300s", content.Text)
}

func TestGraphSearchMissingQuery(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleGraphSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return error for missing query
	if !result.IsError {
		t.Error("expected error for missing query")
	}
}

func TestGraphSearch(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"query":    "show setup",
		"max_hops": 2,
	}

	result, err := handleGraphSearch(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Graph search results: %.300s", content.Text)
}

func TestContextFromGraph(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"document":    "test-document",
		"max_results": 5,
	}

	result, err := handleContextFromGraph(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Context from graph: %.300s", content.Text)
}

func TestSimilarShows(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"criteria": "resolume lighting",
	}

	result, err := handleSimilarShows(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Similar shows: %.300s", content.Text)
}

func TestResolutionPath(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"issue": "video dropout during stream",
	}

	result, err := handleResolutionPath(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := result.Content[0].(mcp.TextContent)
	t.Logf("Resolution path: %.300s", content.Text)
}

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}
