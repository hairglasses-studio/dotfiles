// Package slack provides Slack messaging tools for hg-mcp.
package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Slack integration.
type Module struct{}

func (m *Module) Name() string {
	return "slack"
}

func (m *Module) Description() string {
	return "Slack messaging and notifications"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_slack_status",
				mcp.WithDescription("Get Slack connection status."),
			),
			Handler:             handleStatus,
			Category:            "messaging",
			Subcategory:         "slack",
			Tags:                []string{"slack", "status", "messaging"},
			UseCases:            []string{"Check Slack token configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "slack",
		},
		{
			Tool: mcp.NewTool("aftrs_slack_health",
				mcp.WithDescription("Check Slack health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "messaging",
			Subcategory:         "slack",
			Tags:                []string{"slack", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Slack issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "slack",
		},
		{
			Tool: mcp.NewTool("aftrs_slack_send",
				mcp.WithDescription("Send a message to a Slack channel."),
				mcp.WithString("message", mcp.Required(), mcp.Description("Message text to send")),
				mcp.WithString("channel", mcp.Description("Channel to send to (default: configured SLACK_CHANNEL)")),
			),
			Handler:             handleSend,
			Category:            "messaging",
			Subcategory:         "slack",
			Tags:                []string{"slack", "send", "message", "notification"},
			UseCases:            []string{"Send notifications", "Post messages to channels"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "slack",
		},
	}
}

var getClient = tools.LazyClient(clients.GetSlackClient)

// handleStatus returns Slack connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Slack client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Slack Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Configured\n\n")
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Create a Slack Bot at https://api.slack.com/apps\n")
		sb.WriteString("2. Add `chat:write` scope\n")
		sb.WriteString("3. Install to workspace and copy Bot Token\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export SLACK_TOKEN=xoxb-your-bot-token\n")
		sb.WriteString("export SLACK_CHANNEL=#general\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Default Channel:** %s\n", status.Channel))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Slack health and recommendations.
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	score := 100
	var issues []string
	var recommendations []string

	if !status.HasToken {
		score -= 50
		issues = append(issues, "SLACK_TOKEN not configured")
		recommendations = append(recommendations,
			"Set SLACK_TOKEN environment variable",
			"Create a Slack Bot at https://api.slack.com/apps",
		)
	}

	healthStatus := "healthy"
	if score < 80 {
		healthStatus = "degraded"
	}
	if score < 50 {
		healthStatus = "critical"
	}

	health := map[string]interface{}{
		"score":           score,
		"status":          healthStatus,
		"connected":       status.Connected,
		"issues":          issues,
		"recommendations": recommendations,
	}
	return tools.JSONResult(health), nil
}

// handleSend sends a message to a Slack channel.
func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	channel := tools.GetStringParam(req, "channel")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if channel == "" {
		channel = client.Channel()
	}

	err = client.SendMessage(ctx, channel, message)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message sent to %s", channel)), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
