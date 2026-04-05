// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// SlackClient provides access to the Slack API for sending messages.
type SlackClient struct {
	token   string
	channel string
}

// SlackStatus represents Slack connection status.
type SlackStatus struct {
	Connected bool   `json:"connected"`
	Channel   string `json:"channel"`
	HasToken  bool   `json:"has_token"`
}

var (
	slackClientSingleton *SlackClient
	slackClientOnce      sync.Once
	slackClientErr       error

	// TestOverrideSlackClient, when non-nil, is returned by GetSlackClient.
	TestOverrideSlackClient *SlackClient
)

// GetSlackClient returns the singleton Slack client.
func GetSlackClient() (*SlackClient, error) {
	if TestOverrideSlackClient != nil {
		return TestOverrideSlackClient, nil
	}
	slackClientOnce.Do(func() {
		slackClientSingleton, slackClientErr = NewSlackClient()
	})
	return slackClientSingleton, slackClientErr
}

// NewTestSlackClient creates an in-memory test client.
func NewTestSlackClient() *SlackClient {
	return &SlackClient{
		token:   "xoxb-test-token",
		channel: "#test",
	}
}

// NewSlackClient creates a new Slack client from environment.
func NewSlackClient() (*SlackClient, error) {
	token := os.Getenv("SLACK_TOKEN")
	channel := os.Getenv("SLACK_CHANNEL")
	if channel == "" {
		channel = "#general"
	}

	return &SlackClient{
		token:   token,
		channel: channel,
	}, nil
}

// GetStatus returns Slack connection status.
func (c *SlackClient) GetStatus(ctx context.Context) (*SlackStatus, error) {
	return &SlackStatus{
		Connected: c.token != "",
		Channel:   c.channel,
		HasToken:  c.token != "",
	}, nil
}

// SendMessage sends a message to the configured channel.
func (c *SlackClient) SendMessage(ctx context.Context, channel, message string) error {
	if c.token == "" {
		return fmt.Errorf("SLACK_TOKEN not configured")
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}
	// Stub: would use Slack Web API chat.postMessage
	return nil
}

// Channel returns the configured default channel.
func (c *SlackClient) Channel() string {
	return c.channel
}
