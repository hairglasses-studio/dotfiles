// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// TelegramClient provides access to the Telegram Bot API for sending messages.
type TelegramClient struct {
	token  string
	chatID string
}

// TelegramStatus represents Telegram bot connection status.
type TelegramStatus struct {
	Connected bool   `json:"connected"`
	ChatID    string `json:"chat_id"`
	HasToken  bool   `json:"has_token"`
}

var (
	telegramClientSingleton *TelegramClient
	telegramClientOnce      sync.Once
	telegramClientErr       error

	// TestOverrideTelegramClient, when non-nil, is returned by GetTelegramClient.
	TestOverrideTelegramClient *TelegramClient
)

// GetTelegramClient returns the singleton Telegram client.
func GetTelegramClient() (*TelegramClient, error) {
	if TestOverrideTelegramClient != nil {
		return TestOverrideTelegramClient, nil
	}
	telegramClientOnce.Do(func() {
		telegramClientSingleton, telegramClientErr = NewTelegramClient()
	})
	return telegramClientSingleton, telegramClientErr
}

// NewTestTelegramClient creates an in-memory test client.
func NewTestTelegramClient() *TelegramClient {
	return &TelegramClient{
		token:  "123456:ABC-DEF-test-token",
		chatID: "12345678",
	}
}

// NewTelegramClient creates a new Telegram client from environment.
func NewTelegramClient() (*TelegramClient, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	return &TelegramClient{
		token:  token,
		chatID: chatID,
	}, nil
}

// GetStatus returns Telegram bot connection status.
func (c *TelegramClient) GetStatus(ctx context.Context) (*TelegramStatus, error) {
	return &TelegramStatus{
		Connected: c.token != "" && c.chatID != "",
		ChatID:    c.chatID,
		HasToken:  c.token != "",
	}, nil
}

// SendMessage sends a message to the configured chat.
func (c *TelegramClient) SendMessage(ctx context.Context, chatID, message string) error {
	if c.token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN not configured")
	}
	if message == "" {
		return fmt.Errorf("message is required")
	}
	// Stub: would use Telegram Bot API sendMessage endpoint
	return nil
}

// ChatID returns the configured default chat ID.
func (c *TelegramClient) ChatID() string {
	return c.chatID
}
