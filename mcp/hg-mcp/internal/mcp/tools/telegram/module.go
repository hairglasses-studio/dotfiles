// Package telegram provides Telegram Bot messaging tools for hg-mcp.
package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Telegram integration.
type Module struct{}

func (m *Module) Name() string {
	return "telegram"
}

func (m *Module) Description() string {
	return "Telegram Bot messaging and notifications"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_telegram_status",
				mcp.WithDescription("Get Telegram bot connection status."),
			),
			Handler:             handleStatus,
			Category:            "messaging",
			Subcategory:         "telegram",
			Tags:                []string{"telegram", "status", "bot", "messaging"},
			UseCases:            []string{"Check Telegram bot configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "telegram",
		},
		{
			Tool: mcp.NewTool("aftrs_telegram_health",
				mcp.WithDescription("Check Telegram bot health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "messaging",
			Subcategory:         "telegram",
			Tags:                []string{"telegram", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Telegram bot issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "telegram",
		},
		{
			Tool: mcp.NewTool("aftrs_telegram_send",
				mcp.WithDescription("Send a message via Telegram bot."),
				mcp.WithString("message", mcp.Required(), mcp.Description("Message text to send")),
				mcp.WithString("chat_id", mcp.Description("Chat ID to send to (default: configured TELEGRAM_CHAT_ID)")),
			),
			Handler:             handleSend,
			Category:            "messaging",
			Subcategory:         "telegram",
			Tags:                []string{"telegram", "send", "message", "notification"},
			UseCases:            []string{"Send notifications", "Alert on events"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "telegram",
		},
	}
}

var getClient = tools.LazyClient(clients.GetTelegramClient)

// handleStatus returns Telegram bot connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Telegram client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Telegram Bot Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Configured\n\n")
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Create a bot via @BotFather on Telegram\n")
		sb.WriteString("2. Copy the bot token\n")
		sb.WriteString("3. Get your chat ID (send /start to your bot, then check the API)\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export TELEGRAM_BOT_TOKEN=123456:ABC-DEF...\n")
		sb.WriteString("export TELEGRAM_CHAT_ID=your-chat-id\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Chat ID:** %s\n", status.ChatID))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Telegram bot health and recommendations.
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
		score -= 30
		issues = append(issues, "TELEGRAM_BOT_TOKEN not configured")
		recommendations = append(recommendations, "Set TELEGRAM_BOT_TOKEN environment variable")
	}

	if status.ChatID == "" {
		score -= 20
		issues = append(issues, "TELEGRAM_CHAT_ID not configured")
		recommendations = append(recommendations, "Set TELEGRAM_CHAT_ID environment variable")
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

// handleSend sends a message via Telegram bot.
func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	chatID := tools.GetStringParam(req, "chat_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if chatID == "" {
		chatID = client.ChatID()
	}

	if chatID == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("chat_id is required (no default configured)")), nil
	}

	err = client.SendMessage(ctx, chatID, message)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message sent to chat %s", chatID)), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
