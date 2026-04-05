// Package clients provides client implementations for external services.
package clients

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// GmailClient provides Gmail operations
type GmailClient struct {
	service *gmail.Service
	mu      sync.RWMutex
}

// Email represents an email message
type Email struct {
	ID        string            `json:"id"`
	ThreadID  string            `json:"thread_id"`
	Subject   string            `json:"subject"`
	From      string            `json:"from"`
	To        []string          `json:"to,omitempty"`
	CC        []string          `json:"cc,omitempty"`
	Date      time.Time         `json:"date"`
	Snippet   string            `json:"snippet"`
	Body      string            `json:"body,omitempty"`
	Labels    []string          `json:"labels,omitempty"`
	IsUnread  bool              `json:"is_unread"`
	HasAttach bool              `json:"has_attachments"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// EmailThread represents a Gmail thread
type EmailThread struct {
	ID       string  `json:"id"`
	Subject  string  `json:"subject"`
	Snippet  string  `json:"snippet"`
	Messages []Email `json:"messages,omitempty"`
	Count    int     `json:"message_count"`
}

// GmailLabel represents a Gmail label
type GmailLabel struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	MessagesTotal  int64  `json:"messages_total"`
	MessagesUnread int64  `json:"messages_unread"`
	ThreadsTotal   int64  `json:"threads_total"`
	ThreadsUnread  int64  `json:"threads_unread"`
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html"`
}

var (
	gmailClient     *GmailClient
	gmailClientOnce sync.Once
	gmailClientErr  error

	// TestOverrideGmailClient, when non-nil, is returned by GetGmailClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideGmailClient *GmailClient
)

// GetGmailClient returns the singleton Gmail client
func GetGmailClient() (*GmailClient, error) {
	if TestOverrideGmailClient != nil {
		return TestOverrideGmailClient, nil
	}
	gmailClientOnce.Do(func() {
		gmailClient, gmailClientErr = NewGmailClient()
	})
	return gmailClient, gmailClientErr
}

// NewTestGmailClient creates an in-memory test client.
func NewTestGmailClient() *GmailClient {
	return &GmailClient{}
}

// NewGmailClient creates a new Gmail client
func NewGmailClient() (*GmailClient, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	var credFile string

	// Priority: GMAIL_APPLICATION_CREDENTIALS > ~/.config/gcloud/gmail_credentials.json > GOOGLE_APPLICATION_CREDENTIALS
	cfg := config.Get()
	if credFile = os.Getenv("GMAIL_APPLICATION_CREDENTIALS"); credFile != "" {
		// Use Gmail-specific credentials
	} else if homeDir := cfg.Home; homeDir != "" {
		defaultGmailCreds := homeDir + "/.config/gcloud/gmail_credentials.json"
		if _, err := os.Stat(defaultGmailCreds); err == nil {
			credFile = defaultGmailCreds
		}
	}
	if credFile == "" {
		credFile = cfg.GoogleApplicationCredentials
	}

	if credFile != "" {
		opts = append(opts, option.WithCredentialsFile(credFile))
	} else if cfg.GoogleAPIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.GoogleAPIKey))
	} else {
		return nil, fmt.Errorf("no Google credentials configured (set GMAIL_APPLICATION_CREDENTIALS, create ~/.config/gcloud/gmail_credentials.json, or set GOOGLE_APPLICATION_CREDENTIALS)")
	}

	service, err := gmail.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return &GmailClient{service: service}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *GmailClient) IsConfigured() bool {
	return c != nil && c.service != nil
}

// ListLabels returns all Gmail labels
func (c *GmailClient) ListLabels(ctx context.Context) ([]GmailLabel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	resp, err := c.service.Users.Labels.List("me").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	result := make([]GmailLabel, 0, len(resp.Labels))
	for _, l := range resp.Labels {
		result = append(result, GmailLabel{
			ID:             l.Id,
			Name:           l.Name,
			Type:           l.Type,
			MessagesTotal:  l.MessagesTotal,
			MessagesUnread: l.MessagesUnread,
			ThreadsTotal:   l.ThreadsTotal,
			ThreadsUnread:  l.ThreadsUnread,
		})
	}

	return result, nil
}

// ListMessages returns messages matching the query
func (c *GmailClient) ListMessages(ctx context.Context, query string, maxResults int64, labelIDs []string) ([]Email, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if maxResults <= 0 {
		maxResults = 25
	}

	call := c.service.Users.Messages.List("me").Context(ctx).MaxResults(maxResults)
	if query != "" {
		call = call.Q(query)
	}
	if len(labelIDs) > 0 {
		call = call.LabelIds(labelIDs...)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	result := make([]Email, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		email, err := c.GetMessage(ctx, m.Id, false)
		if err != nil {
			continue // Skip messages that fail to load
		}
		result = append(result, *email)
	}

	return result, nil
}

// GetMessage gets a specific message by ID
func (c *GmailClient) GetMessage(ctx context.Context, messageID string, includeBody bool) (*Email, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	format := "metadata"
	if includeBody {
		format = "full"
	}

	msg, err := c.service.Users.Messages.Get("me", messageID).Context(ctx).Format(format).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return c.convertMessage(msg, includeBody), nil
}

// GetThread gets a thread with all messages
func (c *GmailClient) GetThread(ctx context.Context, threadID string) (*EmailThread, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	thread, err := c.service.Users.Threads.Get("me", threadID).Context(ctx).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	messages := make([]Email, 0, len(thread.Messages))
	var subject string
	for i, m := range thread.Messages {
		email := c.convertMessage(m, true)
		messages = append(messages, *email)
		if i == 0 {
			subject = email.Subject
		}
	}

	return &EmailThread{
		ID:       thread.Id,
		Subject:  subject,
		Snippet:  thread.Snippet,
		Messages: messages,
		Count:    len(messages),
	}, nil
}

// SearchMessages searches for messages with a query
func (c *GmailClient) SearchMessages(ctx context.Context, query string, maxResults int64) ([]Email, error) {
	return c.ListMessages(ctx, query, maxResults, nil)
}

// GetUnread returns unread messages
func (c *GmailClient) GetUnread(ctx context.Context, maxResults int64) ([]Email, error) {
	return c.ListMessages(ctx, "is:unread", maxResults, nil)
}

// GetInbox returns inbox messages
func (c *GmailClient) GetInbox(ctx context.Context, maxResults int64) ([]Email, error) {
	return c.ListMessages(ctx, "", maxResults, []string{"INBOX"})
}

// GetSent returns sent messages
func (c *GmailClient) GetSent(ctx context.Context, maxResults int64) ([]Email, error) {
	return c.ListMessages(ctx, "", maxResults, []string{"SENT"})
}

// GetStarred returns starred messages
func (c *GmailClient) GetStarred(ctx context.Context, maxResults int64) ([]Email, error) {
	return c.ListMessages(ctx, "", maxResults, []string{"STARRED"})
}

// MarkAsRead marks a message as read
func (c *GmailClient) MarkAsRead(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}
	return nil
}

// MarkAsUnread marks a message as unread
func (c *GmailClient) MarkAsUnread(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"UNREAD"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to mark as unread: %w", err)
	}
	return nil
}

// Star adds a star to a message
func (c *GmailClient) Star(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"STARRED"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to star message: %w", err)
	}
	return nil
}

// Unstar removes a star from a message
func (c *GmailClient) Unstar(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"STARRED"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to unstar message: %w", err)
	}
	return nil
}

// Archive moves a message to archive (removes from inbox)
func (c *GmailClient) Archive(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"INBOX"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to archive message: %w", err)
	}
	return nil
}

// Trash moves a message to trash
func (c *GmailClient) Trash(ctx context.Context, messageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Trash("me", messageID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to trash message: %w", err)
	}
	return nil
}

// SendEmail sends an email
func (c *GmailClient) SendEmail(ctx context.Context, req *SendEmailRequest) (*Email, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Build the raw message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(req.To, ", ")))
	if len(req.CC) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(req.CC, ", ")))
	}
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", req.Subject))
	if req.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(req.Body)

	raw := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))

	msg := &gmail.Message{Raw: raw}
	sent, err := c.service.Users.Messages.Send("me", msg).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	// Get the sent message details
	return c.GetMessage(ctx, sent.Id, false)
}

// CreateDraft creates a draft email
func (c *GmailClient) CreateDraft(ctx context.Context, req *SendEmailRequest) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(req.To, ", ")))
	if len(req.CC) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(req.CC, ", ")))
	}
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", req.Subject))
	if req.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(req.Body)

	raw := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))

	draft := &gmail.Draft{
		Message: &gmail.Message{Raw: raw},
	}

	created, err := c.service.Users.Drafts.Create("me", draft).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create draft: %w", err)
	}

	return created.Id, nil
}

// AddLabel adds a label to a message
func (c *GmailClient) AddLabel(ctx context.Context, messageID, labelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to add label: %w", err)
	}
	return nil
}

// RemoveLabel removes a label from a message
func (c *GmailClient) RemoveLabel(ctx context.Context, messageID, labelID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := c.service.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{labelID},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to remove label: %w", err)
	}
	return nil
}

func (c *GmailClient) convertMessage(msg *gmail.Message, includeBody bool) *Email {
	email := &Email{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Snippet:  msg.Snippet,
		Labels:   msg.LabelIds,
		Headers:  make(map[string]string),
	}

	// Check for unread
	for _, label := range msg.LabelIds {
		if label == "UNREAD" {
			email.IsUnread = true
			break
		}
	}

	// Parse headers
	if msg.Payload != nil {
		for _, h := range msg.Payload.Headers {
			email.Headers[h.Name] = h.Value
			switch strings.ToLower(h.Name) {
			case "subject":
				email.Subject = h.Value
			case "from":
				email.From = h.Value
			case "to":
				email.To = strings.Split(h.Value, ", ")
			case "cc":
				email.CC = strings.Split(h.Value, ", ")
			case "date":
				if t, err := time.Parse(time.RFC1123Z, h.Value); err == nil {
					email.Date = t
				} else if t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", h.Value); err == nil {
					email.Date = t
				}
			}
		}

		// Check for attachments
		if msg.Payload.Parts != nil {
			for _, part := range msg.Payload.Parts {
				if part.Filename != "" {
					email.HasAttach = true
					break
				}
			}
		}

		// Get body if requested
		if includeBody {
			email.Body = c.getMessageBody(msg.Payload)
		}
	}

	// Use internal date if header date not parsed
	if email.Date.IsZero() && msg.InternalDate != 0 {
		email.Date = time.UnixMilli(msg.InternalDate)
	}

	return email
}

func (c *GmailClient) getMessageBody(payload *gmail.MessagePart) string {
	if payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	// Check parts for body
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			if part.Body != nil && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
		// Recurse into nested parts
		if len(part.Parts) > 0 {
			if body := c.getMessageBody(part); body != "" {
				return body
			}
		}
	}

	return ""
}
