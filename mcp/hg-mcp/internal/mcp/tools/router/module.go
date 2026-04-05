// Package router provides the smart routing tool for hg-mcp.
package router

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewRouterClient)

// Module implements the ToolModule interface for the smart router
type Module struct{}

func (m *Module) Name() string {
	return "router"
}

func (m *Module) Description() string {
	return "Smart routing - natural language to appropriate tool"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ask",
				mcp.WithDescription("Natural language query router. Ask anything and it routes to the appropriate tool. Examples: 'what's the TouchDesigner FPS?', 'are there any NDI sources?', 'start the show', 'what's wrong with the lighting?'"),
				mcp.WithString("query", mcp.Description("Your question or command in natural language"), mcp.Required()),
				mcp.WithBoolean("explain", mcp.Description("If true, explains the routing instead of executing")),
			),
			Handler:             handleAsk,
			Category:            "router",
			Subcategory:         "query",
			Tags:                []string{"router", "ask", "natural language", "query"},
			UseCases:            []string{"Ask questions naturally", "Route to correct tool"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "router",
		},
	}
}

// handleAsk handles the aftrs_ask tool
func handleAsk(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}
	explain := tools.GetBoolParam(req, "explain", false)

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create router client: %w", err)), nil
	}

	result, err := client.Route(ctx, query)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to route query: %w", err)), nil
	}

	var sb strings.Builder

	if explain {
		sb.WriteString("# Query Routing Explanation\n\n")
		sb.WriteString(fmt.Sprintf("**Query:** %s\n\n", query))

		emoji := getConfidenceEmoji(result.Confidence)
		sb.WriteString(fmt.Sprintf("## Best Match %s\n", emoji))
		sb.WriteString(fmt.Sprintf("**Tool:** `%s`\n", result.Tool))
		sb.WriteString(fmt.Sprintf("**Confidence:** %.0f%%\n", result.Confidence*100))
		sb.WriteString(fmt.Sprintf("**Explanation:** %s\n\n", result.Explanation))

		if len(result.Parameters) > 0 {
			sb.WriteString("## Extracted Parameters\n")
			for k, v := range result.Parameters {
				sb.WriteString(fmt.Sprintf("- **%s:** %s\n", k, v))
			}
			sb.WriteString("\n")
		}

		if len(result.Alternatives) > 0 {
			sb.WriteString("## Alternatives\n")
			for _, alt := range result.Alternatives {
				sb.WriteString(fmt.Sprintf("- `%s`\n", alt))
			}
		}

		sb.WriteString("\nTo execute, run the query again without `explain: true`.\n")
	} else {
		sb.WriteString("# Query Routed\n\n")
		sb.WriteString(fmt.Sprintf("**Query:** %s\n", query))
		emoji := getConfidenceEmoji(result.Confidence)
		sb.WriteString(fmt.Sprintf("**Routed to:** `%s` %s (%.0f%% confidence)\n\n", result.Tool, emoji, result.Confidence*100))
		sb.WriteString(fmt.Sprintf("**Why:** %s\n\n", result.Explanation))

		if len(result.Parameters) > 0 {
			sb.WriteString("## Parameters Detected\n")
			for k, v := range result.Parameters {
				sb.WriteString(fmt.Sprintf("- **%s:** %s\n", k, v))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("## Recommended Action\n")
		sb.WriteString(fmt.Sprintf("Call `%s` with the detected parameters.\n", result.Tool))

		if len(result.Alternatives) > 0 {
			sb.WriteString(fmt.Sprintf("\n*Alternatives: %s*\n", strings.Join(result.Alternatives, ", ")))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// Helper function
func getConfidenceEmoji(confidence float64) string {
	if confidence >= 0.7 {
		return "🟢"
	} else if confidence >= 0.4 {
		return "🟡"
	}
	return "🔴"
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
