// Package discord provides Discord bot tools for hg-mcp.
package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Discord integration
type Module struct{}

func (m *Module) Name() string {
	return "discord"
}

func (m *Module) Description() string {
	return "Discord bot integration for team communication"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_discord_status",
				mcp.WithDescription("Check Discord bot connection status and latency."),
			),
			Handler:             handleStatus,
			Category:            "discord",
			Subcategory:         "status",
			Tags:                []string{"discord", "status", "connection"},
			UseCases:            []string{"Check if bot is connected", "Monitor bot health"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_send",
				mcp.WithDescription("Send a message to a Discord channel."),
				mcp.WithString("message",
					mcp.Required(),
					mcp.Description("The message content to send"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID to send to (uses default if not specified)"),
				),
			),
			Handler:             handleSend,
			Category:            "discord",
			Subcategory:         "messages",
			Tags:                []string{"discord", "send", "message"},
			UseCases:            []string{"Send notifications", "Post updates"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_channels",
				mcp.WithDescription("List available Discord channels in the configured server."),
				mcp.WithString("type",
					mcp.Description("Filter by channel type: 'text', 'voice', 'category', or 'all' (default: all)"),
				),
			),
			Handler:             handleChannels,
			Category:            "discord",
			Subcategory:         "channels",
			Tags:                []string{"discord", "channels", "list"},
			UseCases:            []string{"Find channel IDs", "Browse server structure"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_history",
				mcp.WithDescription("Get message history from a Discord channel."),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID to get history from (uses default if not specified)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Number of messages to retrieve (default: 25, max: 100)"),
				),
			),
			Handler:             handleHistory,
			Category:            "discord",
			Subcategory:         "messages",
			Tags:                []string{"discord", "history", "messages"},
			UseCases:            []string{"Read recent messages", "Check channel activity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_thread",
				mcp.WithDescription("Get messages from a Discord thread."),
				mcp.WithString("thread_id",
					mcp.Required(),
					mcp.Description("The thread ID to get messages from"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Number of messages to retrieve (default: 25, max: 100)"),
				),
			),
			Handler:             handleThread,
			Category:            "discord",
			Subcategory:         "messages",
			Tags:                []string{"discord", "thread", "messages"},
			UseCases:            []string{"Read thread discussions", "Follow conversations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_notify",
				mcp.WithDescription("Send a formatted notification with title, message, and severity level."),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Notification title"),
				),
				mcp.WithString("message",
					mcp.Required(),
					mcp.Description("Notification body text"),
				),
				mcp.WithString("level",
					mcp.Description("Severity level: 'info', 'success', 'warning', 'error' (default: info)"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID to send to (uses default if not specified)"),
				),
			),
			Handler:             handleNotify,
			Category:            "discord",
			Subcategory:         "notifications",
			Tags:                []string{"discord", "notify", "alert", "embed"},
			UseCases:            []string{"Send formatted alerts", "Studio notifications"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_webhook",
				mcp.WithDescription("Send a message via Discord webhook URL."),
				mcp.WithString("content",
					mcp.Description("Plain text content (optional if embed is provided)"),
				),
				mcp.WithString("title",
					mcp.Description("Embed title"),
				),
				mcp.WithString("description",
					mcp.Description("Embed description/body"),
				),
				mcp.WithString("color",
					mcp.Description("Embed color: 'red', 'green', 'blue', 'orange', or hex (default: blue)"),
				),
				mcp.WithString("webhook_url",
					mcp.Description("Webhook URL (uses DISCORD_WEBHOOK_URL env if not specified)"),
				),
			),
			Handler:             handleWebhook,
			Category:            "discord",
			Subcategory:         "webhooks",
			Tags:                []string{"discord", "webhook", "notification"},
			UseCases:            []string{"Send webhook notifications", "External integrations"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_react",
				mcp.WithDescription("Add a reaction emoji to a message."),
				mcp.WithString("message_id",
					mcp.Required(),
					mcp.Description("ID of the message to react to"),
				),
				mcp.WithString("emoji",
					mcp.Required(),
					mcp.Description("Emoji to react with (e.g., '👍', ':thumbsup:', or custom emoji ID)"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID containing the message"),
				),
			),
			Handler:             handleReact,
			Category:            "discord",
			Subcategory:         "reactions",
			Tags:                []string{"discord", "react", "emoji"},
			UseCases:            []string{"React to messages", "Acknowledge notifications"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_edit",
				mcp.WithDescription("Edit an existing message."),
				mcp.WithString("message_id",
					mcp.Required(),
					mcp.Description("ID of the message to edit"),
				),
				mcp.WithString("content",
					mcp.Required(),
					mcp.Description("New message content"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID containing the message"),
				),
			),
			Handler:             handleEdit,
			Category:            "discord",
			Subcategory:         "messages",
			Tags:                []string{"discord", "edit", "message"},
			UseCases:            []string{"Update messages", "Fix typos"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_delete",
				mcp.WithDescription("Delete a message."),
				mcp.WithString("message_id",
					mcp.Required(),
					mcp.Description("ID of the message to delete"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID containing the message"),
				),
			),
			Handler:             handleDelete,
			Category:            "discord",
			Subcategory:         "messages",
			Tags:                []string{"discord", "delete", "message"},
			UseCases:            []string{"Remove messages", "Clean up channels"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_thread_create",
				mcp.WithDescription("Create a new thread in a channel."),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Thread name"),
				),
				mcp.WithString("message_id",
					mcp.Description("Message ID to create thread from (optional - creates standalone thread if not provided)"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID to create thread in"),
				),
			),
			Handler:             handleThreadCreate,
			Category:            "discord",
			Subcategory:         "threads",
			Tags:                []string{"discord", "thread", "create"},
			UseCases:            []string{"Start discussions", "Organize conversations"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
		{
			Tool: mcp.NewTool("aftrs_discord_studio_event",
				mcp.WithDescription("Send a studio event notification (stream, recording, TD error, session)."),
				mcp.WithString("event_type",
					mcp.Required(),
					mcp.Description("Event type: stream_start, stream_stop, record_start, record_stop, td_error, td_warning, health_warning, session_start, session_end"),
				),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Event title"),
				),
				mcp.WithString("details",
					mcp.Description("Event details/description"),
				),
				mcp.WithString("channel_id",
					mcp.Description("Channel ID (uses default if not specified)"),
				),
			),
			Handler:             handleStudioEvent,
			Category:            "discord",
			Subcategory:         "studio",
			Tags:                []string{"discord", "studio", "event", "notification"},
			UseCases:            []string{"Stream alerts", "TD error notifications", "Session logging"},
			Complexity:          tools.ComplexityModerate,
			IsWrite:             true,
			CircuitBreakerGroup: "discord",
		},
	}
}

var getClient = tools.LazyClient(clients.NewDiscordClient)

// handleStatus handles the aftrs_discord_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.TextResult(fmt.Sprintf("Discord not connected: %v\n\nEnsure DISCORD_BOT_TOKEN is set.", err)), nil
	}
	defer client.Close()

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Discord Bot Status\n\n")
	sb.WriteString(fmt.Sprintf("**Connected:** %v\n", status["connected"]))
	sb.WriteString(fmt.Sprintf("**Bot:** %s (ID: %s)\n", status["bot_username"], status["bot_id"]))
	sb.WriteString(fmt.Sprintf("**Latency:** %v ms\n", status["latency_ms"]))

	if guildName, ok := status["guild_name"]; ok {
		sb.WriteString(fmt.Sprintf("\n**Server:** %s\n", guildName))
		sb.WriteString(fmt.Sprintf("**Members:** %v\n", status["member_count"]))
	}

	if status["channel_id"] != "" {
		sb.WriteString(fmt.Sprintf("**Default Channel:** %s\n", status["channel_id"]))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSend handles the aftrs_discord_send tool
func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	msg, err := client.SendMessage(ctx, channelID, message)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message sent successfully!\n\n**Message ID:** %s\n**Channel:** %s\n**Content:** %s",
		msg.ID, msg.ChannelID, msg.Content)), nil
}

// handleChannels handles the aftrs_discord_channels tool
func handleChannels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	typeFilter := tools.OptionalStringParam(req, "type", "all")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	channels, err := client.ListChannels(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Discord Channels\n\n")

	// Group by type
	grouped := make(map[string][]clients.DiscordChannel)
	for _, ch := range channels {
		if typeFilter == "all" || ch.Type == typeFilter {
			grouped[ch.Type] = append(grouped[ch.Type], ch)
		}
	}

	for chType, chs := range grouped {
		sb.WriteString(fmt.Sprintf("## %s channels\n\n", strings.Title(chType)))
		for _, ch := range chs {
			sb.WriteString(fmt.Sprintf("- **%s** (ID: `%s`)\n", ch.Name, ch.ID))
		}
		sb.WriteString("\n")
	}

	if len(grouped) == 0 {
		sb.WriteString("No channels found matching the filter.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleHistory handles the aftrs_discord_history tool
func handleHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	channelID := tools.GetStringParam(req, "channel_id")
	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	messages, err := client.GetHistory(ctx, channelID, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Channel History (%d messages)\n\n", len(messages)))

	for _, msg := range messages {
		timestamp := msg.Timestamp.Format("Jan 2 15:04")
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("**[%s] %s:** %s\n\n", timestamp, msg.Author, content))
	}

	if len(messages) == 0 {
		sb.WriteString("No messages found in this channel.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleThread handles the aftrs_discord_thread tool
func handleThread(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	threadID, errResult := tools.RequireStringParam(req, "thread_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	messages, err := client.GetThreadMessages(ctx, threadID, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Thread Messages (%d messages)\n\n", len(messages)))

	for _, msg := range messages {
		timestamp := msg.Timestamp.Format("Jan 2 15:04")
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("**[%s] %s:** %s\n\n", timestamp, msg.Author, content))
	}

	if len(messages) == 0 {
		sb.WriteString("No messages found in this thread.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleNotify handles the aftrs_discord_notify tool
func handleNotify(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	level := tools.OptionalStringParam(req, "level", "info")

	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	msg, err := client.SendNotification(ctx, channelID, title, message, level)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Notification sent!\n\n**Title:** %s\n**Level:** %s\n**Channel:** %s\n**Message ID:** %s",
		title, level, msg.ChannelID, msg.ID)), nil
}

// handleWebhook handles the aftrs_discord_webhook tool
func handleWebhook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content := tools.GetStringParam(req, "content")
	title := tools.GetStringParam(req, "title")
	description := tools.GetStringParam(req, "description")
	colorStr := tools.GetStringParam(req, "color")
	webhookURL := tools.GetStringParam(req, "webhook_url")

	if content == "" && title == "" && description == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("at least one of content, title, or description is required")), nil
	}

	// Parse color
	color := 0x0099FF // Default blue
	switch colorStr {
	case "red":
		color = 0xFF0000
	case "green":
		color = 0x00FF00
	case "blue":
		color = 0x0099FF
	case "orange":
		color = 0xFFAA00
	case "":
		// Keep default
	default:
		// Try to parse as hex
		if _, err := fmt.Sscanf(colorStr, "0x%x", &color); err != nil {
			if _, err := fmt.Sscanf(colorStr, "#%x", &color); err != nil {
				// Keep default if parsing fails
			}
		}
	}

	payload := &clients.WebhookPayload{
		Content:  content,
		Username: "aftrs-bot",
	}

	if title != "" || description != "" {
		payload.Embeds = []clients.WebhookEmbed{
			{
				Title:       title,
				Description: description,
				Color:       color,
				Footer: &clients.WebhookEmbedFooter{
					Text: "hg-mcp",
				},
			},
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.SendWebhook(ctx, webhookURL, payload); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult("Webhook sent successfully!"), nil
}

// handleReact handles the aftrs_discord_react tool
func handleReact(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	emoji, errResult := tools.RequireStringParam(req, "emoji")
	if errResult != nil {
		return errResult, nil
	}

	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.AddReaction(ctx, channelID, messageID, emoji); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Added reaction %s to message %s", emoji, messageID)), nil
}

// handleEdit handles the aftrs_discord_edit tool
func handleEdit(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	content, errResult := tools.RequireStringParam(req, "content")
	if errResult != nil {
		return errResult, nil
	}

	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	msg, err := client.EditMessage(ctx, channelID, messageID, content)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message edited successfully!\n\n**Message ID:** %s\n**New content:** %s", msg.ID, msg.Content)), nil
}

// handleDelete handles the aftrs_discord_delete tool
func handleDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.DeleteMessage(ctx, channelID, messageID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Message %s deleted successfully", messageID)), nil
}

// handleThreadCreate handles the aftrs_discord_thread_create tool
func handleThreadCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	messageID := tools.GetStringParam(req, "message_id")
	channelID := tools.GetStringParam(req, "channel_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	var thread *clients.DiscordChannel
	if messageID != "" {
		thread, err = client.CreateThread(ctx, channelID, messageID, name, 1440)
	} else {
		thread, err = client.CreateThreadWithoutMessage(ctx, channelID, name, 1440)
	}

	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Thread created!\n\n**Name:** %s\n**Thread ID:** %s\n**Parent:** %s", thread.Name, thread.ID, thread.ParentID)), nil
}

// handleStudioEvent handles the aftrs_discord_studio_event tool
func handleStudioEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	eventType, errResult := tools.RequireStringParam(req, "event_type")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	details := tools.GetStringParam(req, "details")
	channelID := tools.GetStringParam(req, "channel_id")

	// Map string to StudioEventType
	var studioEvent clients.StudioEventType
	switch eventType {
	case "stream_start":
		studioEvent = clients.EventStreamStart
	case "stream_stop":
		studioEvent = clients.EventStreamStop
	case "record_start":
		studioEvent = clients.EventRecordStart
	case "record_stop":
		studioEvent = clients.EventRecordStop
	case "td_error":
		studioEvent = clients.EventTDError
	case "td_warning":
		studioEvent = clients.EventTDWarning
	case "health_warning":
		studioEvent = clients.EventHealthWarning
	case "session_start":
		studioEvent = clients.EventSessionStart
	case "session_end":
		studioEvent = clients.EventSessionEnd
	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid event_type: %s", eventType)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	defer client.Close()

	if err := client.NotifyStudioEvent(ctx, studioEvent, title, details, channelID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Studio event notification sent!\n\n**Type:** %s\n**Title:** %s", eventType, title)), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
