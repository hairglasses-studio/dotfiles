package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

const (
	anthropicAPIURL     = "https://api.anthropic.com/v1/messages"
	anthropicAPIVersion = "2023-06-01"
	defaultModel        = "claude-sonnet-4-20250514"
	maxTokens           = 1024
	maxMessageLength    = 2000 // Discord limit
	rateLimitWindow     = time.Minute
	rateLimitMax        = 10 // Max requests per user per minute
)

// AIHandler handles conversational AI interactions
type AIHandler struct {
	mu           sync.RWMutex
	session      *discordgo.Session
	apiKey       string
	rateLimits   map[string]*userRateLimit
	systemPrompt string
}

// userRateLimit tracks rate limiting for a user
type userRateLimit struct {
	count     int
	resetTime time.Time
}

// anthropicRequest represents a request to the Anthropic API
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

// anthropicMessage represents a message in the conversation
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents the API response
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewAIHandler creates a new AI handler
func NewAIHandler(session *discordgo.Session) *AIHandler {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("PERSONAL_CLAUDE_MAX_ANTHROPIC_API_KEY")
	}

	systemPrompt := `You are the AI assistant for The Aftrs, a creative audiovisual studio. You help the team with:
- Studio operations (TouchDesigner, OBS, lighting, streaming)
- Technical troubleshooting
- Creative project support
- General questions

Keep responses concise and helpful. You have access to studio context when provided.
If asked about something you don't know, suggest using the appropriate MCP tools or checking the vault documentation.`

	return &AIHandler{
		session:      session,
		apiKey:       apiKey,
		rateLimits:   make(map[string]*userRateLimit),
		systemPrompt: systemPrompt,
	}
}

// IsConfigured returns true if the AI handler has an API key
func (h *AIHandler) IsConfigured() bool {
	return h.apiKey != ""
}

// HandleMessage processes a message that mentions the bot
func (h *AIHandler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages
	if m.Author.Bot {
		return
	}

	// Check if bot is mentioned or if it's a DM
	isMentioned := false
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID {
			isMentioned = true
			break
		}
	}

	// Check if DM
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}
	isDM := channel.Type == discordgo.ChannelTypeDM

	if !isMentioned && !isDM {
		return
	}

	// Check if AI is configured
	if !h.IsConfigured() {
		s.ChannelMessageSend(m.ChannelID, "AI assistant is not configured. Set `ANTHROPIC_API_KEY` environment variable.")
		return
	}

	// Check rate limit
	if !h.checkRateLimit(m.Author.ID) {
		s.ChannelMessageSend(m.ChannelID, "You've reached the rate limit. Please wait a minute before trying again.")
		return
	}

	// Show typing indicator
	s.ChannelTyping(m.ChannelID)

	// Extract the actual message (remove bot mention)
	content := h.cleanMessage(m.Content, s.State.User.ID)
	if strings.TrimSpace(content) == "" {
		s.ChannelMessageSend(m.ChannelID, "How can I help you?")
		return
	}

	// Gather context
	ctx := context.Background()
	contextInfo := h.gatherContext(ctx, m.ChannelID, s)

	// Build the full prompt with context
	fullPrompt := content
	if contextInfo != "" {
		fullPrompt = fmt.Sprintf("%s\n\n---\nUser question: %s", contextInfo, content)
	}

	// Get recent conversation history from the channel
	messages := h.getConversationHistory(s, m.ChannelID, m.ID)
	messages = append(messages, anthropicMessage{Role: "user", Content: fullPrompt})

	// Call Claude API
	response, err := h.callClaude(ctx, messages)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Sorry, I encountered an error: %s", err.Error()))
		return
	}

	// Send response (handle Discord's 2000 char limit)
	h.sendResponse(s, m.ChannelID, response, m.Reference())
}

// cleanMessage removes the bot mention from the message
func (h *AIHandler) cleanMessage(content string, botID string) string {
	// Remove <@botID> and <@!botID> mentions
	content = strings.ReplaceAll(content, fmt.Sprintf("<@%s>", botID), "")
	content = strings.ReplaceAll(content, fmt.Sprintf("<@!%s>", botID), "")
	return strings.TrimSpace(content)
}

// checkRateLimit checks if a user is within rate limits
func (h *AIHandler) checkRateLimit(userID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	limit, exists := h.rateLimits[userID]

	if !exists || now.After(limit.resetTime) {
		h.rateLimits[userID] = &userRateLimit{
			count:     1,
			resetTime: now.Add(rateLimitWindow),
		}
		return true
	}

	if limit.count >= rateLimitMax {
		return false
	}

	limit.count++
	return true
}

// gatherContext collects relevant studio context
func (h *AIHandler) gatherContext(ctx context.Context, channelID string, s *discordgo.Session) string {
	var parts []string

	// Get studio status
	studioStatus := h.getStudioStatus(ctx)
	if studioStatus != "" {
		parts = append(parts, "**Current Studio Status:**\n"+studioStatus)
	}

	// Get recent messages for context (last 5)
	messages, err := s.ChannelMessages(channelID, 5, "", "", "")
	if err == nil && len(messages) > 0 {
		var recentMsgs []string
		for i := len(messages) - 1; i >= 0; i-- {
			msg := messages[i]
			if !msg.Author.Bot && msg.Content != "" {
				recentMsgs = append(recentMsgs, fmt.Sprintf("%s: %s", msg.Author.Username, truncate(msg.Content, 100)))
			}
		}
		if len(recentMsgs) > 0 {
			parts = append(parts, "**Recent Channel Context:**\n"+strings.Join(recentMsgs, "\n"))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\n\n")
}

// getStudioStatus gets current studio system status
func (h *AIHandler) getStudioStatus(ctx context.Context) string {
	var status []string

	// Try TouchDesigner
	tdClient, err := clients.NewTouchDesignerClient()
	if err == nil {
		tdStatus, err := tdClient.GetStatus(ctx)
		if err == nil && tdStatus.Connected {
			status = append(status, fmt.Sprintf("- TouchDesigner: Connected, %.1f FPS", tdStatus.FPS))
		}
	}

	// Try OBS
	obsClient, err := clients.NewOBSClient()
	if err == nil {
		obsStatus, err := obsClient.GetStatus(ctx)
		if err == nil {
			streamStatus := "Offline"
			if obsStatus.Streaming {
				streamStatus = "LIVE"
			}
			status = append(status, fmt.Sprintf("- OBS: %s, Scene: %s", streamStatus, obsStatus.CurrentScene))
		}
	}

	if len(status) == 0 {
		return ""
	}

	return strings.Join(status, "\n")
}

// getConversationHistory gets recent bot conversation from the channel
func (h *AIHandler) getConversationHistory(s *discordgo.Session, channelID, beforeID string) []anthropicMessage {
	var history []anthropicMessage

	// Get last 10 messages before the current one
	messages, err := s.ChannelMessages(channelID, 10, beforeID, "", "")
	if err != nil {
		return history
	}

	// Process in chronological order (messages come newest first)
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		// Only include messages from the last 30 minutes
		if time.Since(msg.Timestamp) > 30*time.Minute {
			continue
		}

		if msg.Author.ID == s.State.User.ID {
			// Bot message
			history = append(history, anthropicMessage{
				Role:    "assistant",
				Content: msg.Content,
			})
		} else if !msg.Author.Bot {
			// User message that mentions bot
			for _, mention := range msg.Mentions {
				if mention.ID == s.State.User.ID {
					content := h.cleanMessage(msg.Content, s.State.User.ID)
					if content != "" {
						history = append(history, anthropicMessage{
							Role:    "user",
							Content: content,
						})
					}
					break
				}
			}
		}
	}

	// Limit history to last 6 messages (3 exchanges)
	if len(history) > 6 {
		history = history[len(history)-6:]
	}

	return history
}

// callClaude makes a request to the Claude API
func (h *AIHandler) callClaude(ctx context.Context, messages []anthropicMessage) (string, error) {
	reqBody := anthropicRequest{
		Model:     defaultModel,
		MaxTokens: maxTokens,
		System:    h.systemPrompt,
		Messages:  messages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)

	client := httpclient.Slow()
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("API error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}

// sendResponse sends a response, handling Discord's character limit
func (h *AIHandler) sendResponse(s *discordgo.Session, channelID, response string, ref *discordgo.MessageReference) {
	// Split response if too long
	chunks := splitMessage(response, maxMessageLength)

	for i, chunk := range chunks {
		var err error
		if i == 0 && ref != nil {
			// Reply to the original message
			_, err = s.ChannelMessageSendReply(channelID, chunk, ref)
		} else {
			_, err = s.ChannelMessageSend(channelID, chunk)
		}
		if err != nil {
			fmt.Printf("Failed to send message: %v\n", err)
		}
	}
}

// splitMessage splits a message into chunks that fit Discord's limit
func splitMessage(msg string, maxLen int) []string {
	if len(msg) <= maxLen {
		return []string{msg}
	}

	var chunks []string
	remaining := msg

	for len(remaining) > 0 {
		if len(remaining) <= maxLen {
			chunks = append(chunks, remaining)
			break
		}

		// Find a good break point (newline or space)
		breakPoint := maxLen
		for i := maxLen - 1; i > maxLen/2; i-- {
			if remaining[i] == '\n' || remaining[i] == ' ' {
				breakPoint = i
				break
			}
		}

		chunks = append(chunks, remaining[:breakPoint])
		remaining = strings.TrimLeft(remaining[breakPoint:], " \n")
	}

	return chunks
}

// truncate truncates a string to max length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
