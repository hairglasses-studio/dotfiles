// Package gmail provides MCP tools for Gmail integration.
package gmail

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Gmail tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "gmail"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Gmail integration for email management, show bookings, and communications"
}

// Tools returns the Gmail tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_gmail_inbox",
				mcp.WithDescription("Get recent inbox messages"),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 25)")),
			),
			Handler:             handleInbox,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "inbox", "email", "messages"},
			UseCases:            []string{"Check inbox", "View recent emails", "Morning briefing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_unread",
				mcp.WithDescription("Get unread messages"),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 25)")),
			),
			Handler:             handleUnread,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "unread", "email", "new"},
			UseCases:            []string{"Check new emails", "Unread count", "Pending messages"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_search",
				mcp.WithDescription("Search Gmail messages using Gmail search syntax"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Gmail search query (e.g., 'from:booking@venue.com', 'subject:gig', 'after:2024/01/01')")),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 25)")),
			),
			Handler:             handleSearch,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "search", "find", "query"},
			UseCases:            []string{"Find booking emails", "Search by sender", "Find show confirmations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_get",
				mcp.WithDescription("Get a specific email message by ID"),
				mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
				mcp.WithBoolean("include_body", mcp.Description("Include full message body (default: true)")),
			),
			Handler:             handleGetMessage,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "get", "read", "message"},
			UseCases:            []string{"Read email", "View full message", "Get details"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_thread",
				mcp.WithDescription("Get an email thread with all messages"),
				mcp.WithString("thread_id", mcp.Required(), mcp.Description("Gmail thread ID")),
			),
			Handler:             handleGetThread,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "thread", "conversation", "history"},
			UseCases:            []string{"View conversation", "Follow email thread", "Get context"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_labels",
				mcp.WithDescription("List all Gmail labels"),
			),
			Handler:             handleLabels,
			Category:            "gmail",
			Subcategory:         "management",
			Tags:                []string{"gmail", "labels", "folders", "organize"},
			UseCases:            []string{"View labels", "Get label IDs", "Organize emails"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_starred",
				mcp.WithDescription("Get starred messages"),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 25)")),
			),
			Handler:             handleStarred,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "starred", "important", "flagged"},
			UseCases:            []string{"View important emails", "Check starred", "Priority inbox"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_sent",
				mcp.WithDescription("Get sent messages"),
				mcp.WithNumber("limit", mcp.Description("Maximum messages to return (default: 25)")),
			),
			Handler:             handleSent,
			Category:            "gmail",
			Subcategory:         "messages",
			Tags:                []string{"gmail", "sent", "outgoing", "history"},
			UseCases:            []string{"View sent emails", "Check sent messages", "Outbox"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_send",
				mcp.WithDescription("Send an email"),
				mcp.WithArray("to", mcp.Required(), mcp.Description("Recipient email addresses"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithArray("cc", mcp.Description("CC email addresses"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject")),
				mcp.WithString("body", mcp.Required(), mcp.Description("Email body content")),
				mcp.WithBoolean("is_html", mcp.Description("Whether body is HTML (default: false)")),
			),
			Handler:             handleSend,
			Category:            "gmail",
			Subcategory:         "compose",
			Tags:                []string{"gmail", "send", "compose", "email"},
			UseCases:            []string{"Send email", "Reply to booking", "Contact venue"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_draft",
				mcp.WithDescription("Create an email draft"),
				mcp.WithArray("to", mcp.Required(), mcp.Description("Recipient email addresses"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithArray("cc", mcp.Description("CC email addresses"), mcp.Items(map[string]any{"type": "string"})),
				mcp.WithString("subject", mcp.Required(), mcp.Description("Email subject")),
				mcp.WithString("body", mcp.Required(), mcp.Description("Email body content")),
				mcp.WithBoolean("is_html", mcp.Description("Whether body is HTML (default: false)")),
			),
			Handler:             handleDraft,
			Category:            "gmail",
			Subcategory:         "compose",
			Tags:                []string{"gmail", "draft", "compose", "save"},
			UseCases:            []string{"Create draft", "Save for later", "Prepare email"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_mark_read",
				mcp.WithDescription("Mark a message as read"),
				mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
			),
			Handler:             handleMarkRead,
			Category:            "gmail",
			Subcategory:         "actions",
			Tags:                []string{"gmail", "read", "mark", "status"},
			UseCases:            []string{"Mark as read", "Clear notification", "Update status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_star",
				mcp.WithDescription("Star or unstar a message"),
				mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
				mcp.WithBoolean("star", mcp.Description("True to star, false to unstar (default: true)")),
			),
			Handler:             handleStar,
			Category:            "gmail",
			Subcategory:         "actions",
			Tags:                []string{"gmail", "star", "flag", "important"},
			UseCases:            []string{"Star email", "Mark important", "Flag for follow-up"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_archive",
				mcp.WithDescription("Archive a message (remove from inbox)"),
				mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
			),
			Handler:             handleArchive,
			Category:            "gmail",
			Subcategory:         "actions",
			Tags:                []string{"gmail", "archive", "organize", "cleanup"},
			UseCases:            []string{"Archive email", "Clean inbox", "Organize"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
		{
			Tool: mcp.NewTool("aftrs_gmail_trash",
				mcp.WithDescription("Move a message to trash"),
				mcp.WithString("message_id", mcp.Required(), mcp.Description("Gmail message ID")),
			),
			Handler:             handleTrash,
			Category:            "gmail",
			Subcategory:         "actions",
			Tags:                []string{"gmail", "trash", "delete", "remove"},
			UseCases:            []string{"Delete email", "Trash message", "Remove"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "google",
		},
	}
}

var getGmailClient = tools.LazyClient(clients.GetGmailClient)

func handleInbox(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	messages, err := client.GetInbox(ctx, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get inbox: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}), nil
}

func handleUnread(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	messages, err := client.GetUnread(ctx, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get unread: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}), nil
}

func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	messages, err := client.SearchMessages(ctx, query, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to search: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
		"query":    query,
	}), nil
}

func handleGetMessage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	includeBody := tools.GetBoolParam(req, "include_body", true)

	message, err := client.GetMessage(ctx, messageID, includeBody)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get message: %w", err)), nil
	}

	return tools.JSONResult(message), nil
}

func handleGetThread(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	threadID, errResult := tools.RequireStringParam(req, "thread_id")
	if errResult != nil {
		return errResult, nil
	}

	thread, err := client.GetThread(ctx, threadID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get thread: %w", err)), nil
	}

	return tools.JSONResult(thread), nil
}

func handleLabels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	labels, err := client.ListLabels(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to list labels: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"labels": labels,
		"count":  len(labels),
	}), nil
}

func handleStarred(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	messages, err := client.GetStarred(ctx, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get starred: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}), nil
}

func handleSent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 25)

	messages, err := client.GetSent(ctx, int64(limit))
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get sent: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	}), nil
}

func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	to := tools.GetStringArrayParam(req, "to")
	if len(to) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("to is required")), nil
	}

	subject, errResult := tools.RequireStringParam(req, "subject")
	if errResult != nil {
		return errResult, nil
	}

	body, errResult := tools.RequireStringParam(req, "body")
	if errResult != nil {
		return errResult, nil
	}

	emailReq := &clients.SendEmailRequest{
		To:      to,
		CC:      tools.GetStringArrayParam(req, "cc"),
		Subject: subject,
		Body:    body,
		IsHTML:  tools.GetBoolParam(req, "is_html", false),
	}

	sent, err := client.SendEmail(ctx, emailReq)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to send email: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": sent,
		"status":  "sent",
	}), nil
}

func handleDraft(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	to := tools.GetStringArrayParam(req, "to")
	if len(to) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("to is required")), nil
	}

	subject, errResult := tools.RequireStringParam(req, "subject")
	if errResult != nil {
		return errResult, nil
	}

	body, errResult := tools.RequireStringParam(req, "body")
	if errResult != nil {
		return errResult, nil
	}

	emailReq := &clients.SendEmailRequest{
		To:      to,
		CC:      tools.GetStringArrayParam(req, "cc"),
		Subject: subject,
		Body:    body,
		IsHTML:  tools.GetBoolParam(req, "is_html", false),
	}

	draftID, err := client.CreateDraft(ctx, emailReq)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create draft: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"draft_id": draftID,
		"status":   "draft_created",
		"message":  "Draft saved successfully",
	}), nil
}

func handleMarkRead(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	err = client.MarkAsRead(ctx, messageID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to mark as read: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message_id": messageID,
		"status":     "marked_as_read",
	}), nil
}

func handleStar(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	star := tools.GetBoolParam(req, "star", true)

	if star {
		err = client.Star(ctx, messageID)
	} else {
		err = client.Unstar(ctx, messageID)
	}

	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to update star: %w", err)), nil
	}

	status := "starred"
	if !star {
		status = "unstarred"
	}

	return tools.JSONResult(map[string]interface{}{
		"message_id": messageID,
		"status":     status,
	}), nil
}

func handleArchive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	err = client.Archive(ctx, messageID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to archive: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message_id": messageID,
		"status":     "archived",
	}), nil
}

func handleTrash(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getGmailClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Gmail client: %w", err)), nil
	}

	messageID, errResult := tools.RequireStringParam(req, "message_id")
	if errResult != nil {
		return errResult, nil
	}

	err = client.Trash(ctx, messageID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to trash: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message_id": messageID,
		"status":     "trashed",
	}), nil
}
