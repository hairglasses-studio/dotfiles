// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/hairglasses-studio/hg-mcp/pkg/cache"
	"github.com/hairglasses-studio/hg-mcp/pkg/httpclient"
)

// Discord caches — reduce API calls for structural data that rarely changes
var (
	discordChannelsCache   = cache.New[[]DiscordChannel](30 * time.Second) // Channel list rarely changes
	discordRolesCache      = cache.New[[]DiscordRole](30 * time.Second)    // Roles rarely change
	discordServerInfoCache = cache.New[*GuildInfo](60 * time.Second)       // Server info is static
)

// DiscordClient provides access to Discord Bot API
type DiscordClient struct {
	session   *discordgo.Session
	guildID   string
	channelID string
}

// DiscordMessage represents a Discord message
type DiscordMessage struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
	ChannelID string    `json:"channel_id"`
}

// DiscordChannel represents a Discord channel
type DiscordChannel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	ParentID string `json:"parent_id,omitempty"`
}

// NewDiscordClient creates a new Discord client
func NewDiscordClient() (*DiscordClient, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN environment variable not set")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	// Open connection to Discord
	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("error opening Discord connection: %w", err)
	}

	return &DiscordClient{
		session:   session,
		guildID:   os.Getenv("DISCORD_GUILD_ID"),
		channelID: os.Getenv("DISCORD_CHANNEL_ID"),
	}, nil
}

// Close closes the Discord session
func (c *DiscordClient) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}

// GetStatus returns the bot's connection status
func (c *DiscordClient) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	if c.session == nil {
		return nil, fmt.Errorf("discord session not initialized")
	}

	user := c.session.State.User
	if user == nil {
		return nil, fmt.Errorf("could not get bot user")
	}

	status := map[string]interface{}{
		"connected":     true,
		"bot_id":        user.ID,
		"bot_username":  user.Username,
		"guild_id":      c.guildID,
		"channel_id":    c.channelID,
		"latency_ms":    c.session.HeartbeatLatency().Milliseconds(),
		"gateway_ready": c.session.State.Ready,
	}

	// Get guild info if available
	if c.guildID != "" {
		guild, err := c.session.Guild(c.guildID)
		if err == nil {
			status["guild_name"] = guild.Name
			status["member_count"] = guild.MemberCount
		}
	}

	return status, nil
}

// SendMessage sends a message to a channel
func (c *DiscordClient) SendMessage(ctx context.Context, channelID, content string) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}
	if channelID == "" {
		return nil, fmt.Errorf("no channel ID specified")
	}

	msg, err := c.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   msg.Content,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// ListChannels lists channels in the guild (cached 30s)
func (c *DiscordClient) ListChannels(ctx context.Context) ([]DiscordChannel, error) {
	return discordChannelsCache.GetOrFetch(ctx, func(ctx context.Context) ([]DiscordChannel, error) {
		if c.guildID == "" {
			return nil, fmt.Errorf("no guild ID configured")
		}

		channels, err := c.session.GuildChannels(c.guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to list channels: %w", err)
		}

		result := make([]DiscordChannel, 0, len(channels))
		for _, ch := range channels {
			channelType := "unknown"
			switch ch.Type {
			case discordgo.ChannelTypeGuildText:
				channelType = "text"
			case discordgo.ChannelTypeGuildVoice:
				channelType = "voice"
			case discordgo.ChannelTypeGuildCategory:
				channelType = "category"
			case discordgo.ChannelTypeGuildNews:
				channelType = "news"
			case discordgo.ChannelTypeGuildForum:
				channelType = "forum"
			}

			result = append(result, DiscordChannel{
				ID:       ch.ID,
				Name:     ch.Name,
				Type:     channelType,
				ParentID: ch.ParentID,
			})
		}

		return result, nil
	})
}

// GetHistory gets message history from a channel
func (c *DiscordClient) GetHistory(ctx context.Context, channelID string, limit int) ([]DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}
	if channelID == "" {
		return nil, fmt.Errorf("no channel ID specified")
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	messages, err := c.session.ChannelMessages(channelID, limit, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get channel history: %w", err)
	}

	result := make([]DiscordMessage, 0, len(messages))
	for _, msg := range messages {
		result = append(result, DiscordMessage{
			ID:        msg.ID,
			Content:   msg.Content,
			Author:    msg.Author.Username,
			Timestamp: msg.Timestamp,
			ChannelID: msg.ChannelID,
		})
	}

	return result, nil
}

// GetThreadMessages gets messages from a thread
func (c *DiscordClient) GetThreadMessages(ctx context.Context, threadID string, limit int) ([]DiscordMessage, error) {
	if threadID == "" {
		return nil, fmt.Errorf("thread ID is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	messages, err := c.session.ChannelMessages(threadID, limit, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	result := make([]DiscordMessage, 0, len(messages))
	for _, msg := range messages {
		result = append(result, DiscordMessage{
			ID:        msg.ID,
			Content:   msg.Content,
			Author:    msg.Author.Username,
			Timestamp: msg.Timestamp,
			ChannelID: msg.ChannelID,
		})
	}

	return result, nil
}

// SendNotification sends a formatted notification message
func (c *DiscordClient) SendNotification(ctx context.Context, channelID, title, message, level string) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}
	if channelID == "" {
		return nil, fmt.Errorf("no channel ID specified")
	}

	// Choose color based on level
	var color int
	var emoji string
	switch level {
	case "error", "critical":
		color = 0xFF0000 // Red
		emoji = ":x:"
	case "warning":
		color = 0xFFAA00 // Orange
		emoji = ":warning:"
	case "success":
		color = 0x00FF00 // Green
		emoji = ":white_check_mark:"
	default:
		color = 0x0099FF // Blue
		emoji = ":information_source:"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", emoji, title),
		Description: message,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "hg-mcp",
		},
	}

	msg, err := c.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   title + ": " + message,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// DefaultChannelID returns the configured default channel ID
func (c *DiscordClient) DefaultChannelID() string {
	return c.channelID
}

// GuildID returns the configured guild ID
func (c *DiscordClient) GuildID() string {
	return c.guildID
}

// WebhookPayload represents a Discord webhook payload
type WebhookPayload struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []WebhookEmbed `json:"embeds,omitempty"`
}

// WebhookEmbed represents an embed in a webhook
type WebhookEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *WebhookEmbedFooter `json:"footer,omitempty"`
	Fields      []WebhookEmbedField `json:"fields,omitempty"`
}

// WebhookEmbedFooter represents embed footer
type WebhookEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// WebhookEmbedField represents an embed field
type WebhookEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// StudioEventType represents different studio event types
type StudioEventType string

const (
	EventStreamStart   StudioEventType = "stream_start"
	EventStreamStop    StudioEventType = "stream_stop"
	EventRecordStart   StudioEventType = "record_start"
	EventRecordStop    StudioEventType = "record_stop"
	EventTDError       StudioEventType = "td_error"
	EventTDWarning     StudioEventType = "td_warning"
	EventHealthWarning StudioEventType = "health_warning"
	EventSessionStart  StudioEventType = "session_start"
	EventSessionEnd    StudioEventType = "session_end"
)

// SendWebhook sends a message via Discord webhook URL
func (c *DiscordClient) SendWebhook(ctx context.Context, webhookURL string, payload *WebhookPayload) error {
	if webhookURL == "" {
		webhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	}
	if webhookURL == "" {
		return fmt.Errorf("no webhook URL specified")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := httpclient.Standard()
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// NotifyStudioEvent sends a formatted studio event notification
func (c *DiscordClient) NotifyStudioEvent(ctx context.Context, eventType StudioEventType, title, details string, channelID string) error {
	if channelID == "" {
		channelID = c.channelID
	}

	var color int
	var emoji string

	switch eventType {
	case EventStreamStart:
		color = 0xFF0000 // Red (live)
		emoji = ":red_circle:"
	case EventStreamStop:
		color = 0x808080 // Gray
		emoji = ":black_circle:"
	case EventRecordStart:
		color = 0xFF6600 // Orange
		emoji = ":orange_circle:"
	case EventRecordStop:
		color = 0x808080 // Gray
		emoji = ":black_circle:"
	case EventTDError:
		color = 0xFF0000 // Red
		emoji = ":x:"
	case EventTDWarning:
		color = 0xFFAA00 // Orange
		emoji = ":warning:"
	case EventHealthWarning:
		color = 0xFFAA00 // Orange
		emoji = ":yellow_circle:"
	case EventSessionStart:
		color = 0x00FF00 // Green
		emoji = ":green_circle:"
	case EventSessionEnd:
		color = 0x0099FF // Blue
		emoji = ":blue_circle:"
	default:
		color = 0x5865F2 // Discord blurple
		emoji = ":information_source:"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s %s", emoji, title),
		Description: details,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Studio Notifications",
		},
	}

	_, err := c.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// SendEmbed sends a rich embed message
func (c *DiscordClient) SendEmbed(ctx context.Context, channelID string, embed *discordgo.MessageEmbed) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}
	if channelID == "" {
		return nil, fmt.Errorf("no channel ID specified")
	}

	msg, err := c.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return nil, fmt.Errorf("failed to send embed: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   embed.Title,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// EditMessage edits an existing message
func (c *DiscordClient) EditMessage(ctx context.Context, channelID, messageID, newContent string) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}

	msg, err := c.session.ChannelMessageEdit(channelID, messageID, newContent)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   msg.Content,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// DeleteMessage deletes a message
func (c *DiscordClient) DeleteMessage(ctx context.Context, channelID, messageID string) error {
	if channelID == "" {
		channelID = c.channelID
	}

	return c.session.ChannelMessageDelete(channelID, messageID)
}

// AddReaction adds a reaction to a message
func (c *DiscordClient) AddReaction(ctx context.Context, channelID, messageID, emoji string) error {
	if channelID == "" {
		channelID = c.channelID
	}

	return c.session.MessageReactionAdd(channelID, messageID, emoji)
}

// CreateThread creates a new thread from a message
func (c *DiscordClient) CreateThread(ctx context.Context, channelID, messageID, name string, autoArchiveDuration int) (*DiscordChannel, error) {
	if channelID == "" {
		channelID = c.channelID
	}

	// Auto archive duration must be 60, 1440, 4320, or 10080 minutes
	if autoArchiveDuration == 0 {
		autoArchiveDuration = 1440 // 24 hours default
	}

	thread, err := c.session.MessageThreadStart(channelID, messageID, name, autoArchiveDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return &DiscordChannel{
		ID:       thread.ID,
		Name:     thread.Name,
		Type:     "thread",
		ParentID: thread.ParentID,
	}, nil
}

// CreateThreadWithoutMessage creates a new thread in a channel (not from a message)
func (c *DiscordClient) CreateThreadWithoutMessage(ctx context.Context, channelID, name string, autoArchiveDuration int) (*DiscordChannel, error) {
	if channelID == "" {
		channelID = c.channelID
	}

	if autoArchiveDuration == 0 {
		autoArchiveDuration = 1440
	}

	thread, err := c.session.ThreadStart(channelID, name, discordgo.ChannelTypeGuildPublicThread, autoArchiveDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	return &DiscordChannel{
		ID:       thread.ID,
		Name:     thread.Name,
		Type:     "thread",
		ParentID: thread.ParentID,
	}, nil
}

// Session returns the underlying Discord session
func (c *DiscordClient) Session() *discordgo.Session {
	return c.session
}

// SlashCommandHandler is a function that handles a slash command interaction
type SlashCommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// SlashCommandHandlers maps command names to their handlers
type SlashCommandHandlers map[string]SlashCommandHandler

// RegisterSlashCommandHandlers registers handlers for slash command interactions
func (c *DiscordClient) RegisterSlashCommandHandlers(handlers SlashCommandHandlers) {
	c.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}

		commandName := i.ApplicationCommandData().Name
		if handler, ok := handlers[commandName]; ok {
			if err := handler(s, i); err != nil {
				// Send error response
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf(":x: Error: %v", err),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
		}
	})
}

// RespondToInteraction sends a response to an interaction
func (c *DiscordClient) RespondToInteraction(i *discordgo.InteractionCreate, content string, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
}

// RespondToInteractionWithEmbed sends an embed response to an interaction
func (c *DiscordClient) RespondToInteractionWithEmbed(i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  flags,
		},
	})
}

// DeferInteractionResponse defers the response (shows "thinking..." to user)
func (c *DiscordClient) DeferInteractionResponse(i *discordgo.InteractionCreate, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})
}

// FollowupInteractionResponse sends a follow-up message after deferring
func (c *DiscordClient) FollowupInteractionResponse(i *discordgo.InteractionCreate, content string, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := c.session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
		Flags:   flags,
	})
	return err
}

// FollowupInteractionWithEmbed sends an embed follow-up message after deferring
func (c *DiscordClient) FollowupInteractionWithEmbed(i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	_, err := c.session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  flags,
	})
	return err
}

// GetInteractionOption retrieves a specific option from an interaction
func GetInteractionOption(i *discordgo.InteractionCreate, name string) *discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options
	for _, opt := range options {
		if opt.Name == name {
			return opt
		}
	}
	return nil
}

// GetInteractionStringOption retrieves a string option value
func GetInteractionStringOption(i *discordgo.InteractionCreate, name string) string {
	opt := GetInteractionOption(i, name)
	if opt != nil {
		return opt.StringValue()
	}
	return ""
}

// GetInteractionIntOption retrieves an integer option value
func GetInteractionIntOption(i *discordgo.InteractionCreate, name string) int64 {
	opt := GetInteractionOption(i, name)
	if opt != nil {
		return opt.IntValue()
	}
	return 0
}

// GetInteractionBoolOption retrieves a boolean option value
func GetInteractionBoolOption(i *discordgo.InteractionCreate, name string) bool {
	opt := GetInteractionOption(i, name)
	if opt != nil {
		return opt.BoolValue()
	}
	return false
}

// =============================================================================
// ADMIN: Channel Management
// =============================================================================

// ChannelCreateOptions specifies options for creating a channel
type ChannelCreateOptions struct {
	Name      string
	Type      discordgo.ChannelType
	Topic     string
	ParentID  string
	Position  int
	Bitrate   int // Voice channels only (8000-96000)
	UserLimit int // Voice channels only (0-99)
	NSFW      bool
	Slowmode  int // Seconds (0-21600)
}

// CreateChannel creates a new channel in the guild
func (c *DiscordClient) CreateChannel(ctx context.Context, opts ChannelCreateOptions) (*DiscordChannel, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	data := discordgo.GuildChannelCreateData{
		Name:             opts.Name,
		Type:             opts.Type,
		Topic:            opts.Topic,
		ParentID:         opts.ParentID,
		Position:         opts.Position,
		NSFW:             opts.NSFW,
		RateLimitPerUser: opts.Slowmode,
	}

	if opts.Type == discordgo.ChannelTypeGuildVoice {
		data.Bitrate = opts.Bitrate
		data.UserLimit = opts.UserLimit
	}

	channel, err := c.session.GuildChannelCreateComplex(c.guildID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return &DiscordChannel{
		ID:       channel.ID,
		Name:     channel.Name,
		Type:     channelTypeToString(channel.Type),
		ParentID: channel.ParentID,
	}, nil
}

// CreateCategory creates a new category in the guild
func (c *DiscordClient) CreateCategory(ctx context.Context, name string, position int) (*DiscordChannel, error) {
	return c.CreateChannel(ctx, ChannelCreateOptions{
		Name:     name,
		Type:     discordgo.ChannelTypeGuildCategory,
		Position: position,
	})
}

// DeleteChannel deletes a channel
func (c *DiscordClient) DeleteChannel(ctx context.Context, channelID string) error {
	_, err := c.session.ChannelDelete(channelID)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}
	return nil
}

// ChannelEditOptions specifies options for editing a channel
type ChannelEditOptions struct {
	Name      string
	Topic     string
	Position  int
	ParentID  string
	NSFW      bool
	Slowmode  int
	Bitrate   int
	UserLimit int
}

// EditChannel edits an existing channel
func (c *DiscordClient) EditChannel(ctx context.Context, channelID string, opts ChannelEditOptions) (*DiscordChannel, error) {
	edit := &discordgo.ChannelEdit{}

	if opts.Name != "" {
		edit.Name = opts.Name
	}
	if opts.Topic != "" {
		edit.Topic = opts.Topic
	}
	if opts.Position != 0 {
		edit.Position = &opts.Position
	}
	if opts.ParentID != "" {
		edit.ParentID = opts.ParentID
	}
	edit.NSFW = &opts.NSFW
	if opts.Slowmode > 0 {
		edit.RateLimitPerUser = &opts.Slowmode
	}
	if opts.Bitrate > 0 {
		edit.Bitrate = opts.Bitrate
	}
	if opts.UserLimit > 0 {
		edit.UserLimit = opts.UserLimit
	}

	channel, err := c.session.ChannelEdit(channelID, edit)
	if err != nil {
		return nil, fmt.Errorf("failed to edit channel: %w", err)
	}

	return &DiscordChannel{
		ID:       channel.ID,
		Name:     channel.Name,
		Type:     channelTypeToString(channel.Type),
		ParentID: channel.ParentID,
	}, nil
}

// PermissionOverwrite represents a channel permission overwrite
type PermissionOverwrite struct {
	ID    string // Role or user ID
	Type  int    // 0 = role, 1 = member
	Allow int64  // Allowed permissions bitmask
	Deny  int64  // Denied permissions bitmask
}

// SetChannelPermissions sets permission overwrites for a channel
func (c *DiscordClient) SetChannelPermissions(ctx context.Context, channelID string, overwrites []PermissionOverwrite) error {
	for _, ow := range overwrites {
		err := c.session.ChannelPermissionSet(channelID, ow.ID, discordgo.PermissionOverwriteType(ow.Type), ow.Allow, ow.Deny)
		if err != nil {
			return fmt.Errorf("failed to set permission for %s: %w", ow.ID, err)
		}
	}
	return nil
}

// LockChannel prevents messages by denying send permissions to @everyone
func (c *DiscordClient) LockChannel(ctx context.Context, channelID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}
	// @everyone role ID is the same as guild ID
	return c.session.ChannelPermissionSet(channelID, c.guildID, discordgo.PermissionOverwriteTypeRole, 0, discordgo.PermissionSendMessages)
}

// UnlockChannel restores message permissions for @everyone
func (c *DiscordClient) UnlockChannel(ctx context.Context, channelID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}
	return c.session.ChannelPermissionDelete(channelID, c.guildID)
}

// SetSlowmode sets the slowmode delay for a channel (0-21600 seconds)
func (c *DiscordClient) SetSlowmode(ctx context.Context, channelID string, seconds int) error {
	if seconds < 0 || seconds > 21600 {
		return fmt.Errorf("slowmode must be between 0 and 21600 seconds")
	}

	_, err := c.session.ChannelEdit(channelID, &discordgo.ChannelEdit{
		RateLimitPerUser: &seconds,
	})
	return err
}

// =============================================================================
// ADMIN: Role Management
// =============================================================================

// DiscordRole represents a Discord role
type DiscordRole struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Color       int    `json:"color"`
	Position    int    `json:"position"`
	Permissions int64  `json:"permissions"`
	Mentionable bool   `json:"mentionable"`
	Managed     bool   `json:"managed"`
	MemberCount int    `json:"member_count,omitempty"`
}

// RoleCreateOptions specifies options for creating a role
type RoleCreateOptions struct {
	Name        string
	Color       int   // RGB color value
	Permissions int64 // Permission bitmask
	Hoist       bool  // Display separately in member list
	Mentionable bool
}

// CreateRole creates a new role in the guild
func (c *DiscordClient) CreateRole(ctx context.Context, opts RoleCreateOptions) (*DiscordRole, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	params := &discordgo.RoleParams{
		Name:        opts.Name,
		Color:       &opts.Color,
		Hoist:       &opts.Hoist,
		Permissions: &opts.Permissions,
		Mentionable: &opts.Mentionable,
	}

	role, err := c.session.GuildRoleCreate(c.guildID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &DiscordRole{
		ID:          role.ID,
		Name:        role.Name,
		Color:       role.Color,
		Position:    role.Position,
		Permissions: role.Permissions,
		Mentionable: role.Mentionable,
		Managed:     role.Managed,
	}, nil
}

// DeleteRole deletes a role from the guild
func (c *DiscordClient) DeleteRole(ctx context.Context, roleID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildRoleDelete(c.guildID, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// AssignRole assigns a role to a member
func (c *DiscordClient) AssignRole(ctx context.Context, userID, roleID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberRoleAdd(c.guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}
	return nil
}

// RevokeRole removes a role from a member
func (c *DiscordClient) RevokeRole(ctx context.Context, userID, roleID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberRoleRemove(c.guildID, userID, roleID)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	return nil
}

// ListRoles lists all roles in the guild (cached 30s)
func (c *DiscordClient) ListRoles(ctx context.Context) ([]DiscordRole, error) {
	return discordRolesCache.GetOrFetch(ctx, func(ctx context.Context) ([]DiscordRole, error) {
		if c.guildID == "" {
			return nil, fmt.Errorf("no guild ID configured")
		}

		roles, err := c.session.GuildRoles(c.guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to list roles: %w", err)
		}

		result := make([]DiscordRole, 0, len(roles))
		for _, role := range roles {
			result = append(result, DiscordRole{
				ID:          role.ID,
				Name:        role.Name,
				Color:       role.Color,
				Position:    role.Position,
				Permissions: role.Permissions,
				Mentionable: role.Mentionable,
				Managed:     role.Managed,
			})
		}

		return result, nil
	})
}

// EditRole edits an existing role
func (c *DiscordClient) EditRole(ctx context.Context, roleID string, opts RoleCreateOptions) (*DiscordRole, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	params := &discordgo.RoleParams{
		Name:        opts.Name,
		Color:       &opts.Color,
		Hoist:       &opts.Hoist,
		Permissions: &opts.Permissions,
		Mentionable: &opts.Mentionable,
	}

	role, err := c.session.GuildRoleEdit(c.guildID, roleID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to edit role: %w", err)
	}

	return &DiscordRole{
		ID:          role.ID,
		Name:        role.Name,
		Color:       role.Color,
		Position:    role.Position,
		Permissions: role.Permissions,
		Mentionable: role.Mentionable,
		Managed:     role.Managed,
	}, nil
}

// =============================================================================
// ADMIN: Member Management
// =============================================================================

// DiscordMember represents a Discord guild member with details
type DiscordMember struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Nickname    string    `json:"nickname,omitempty"`
	Avatar      string    `json:"avatar,omitempty"`
	Roles       []string  `json:"roles"`
	JoinedAt    time.Time `json:"joined_at"`
	IsPending   bool      `json:"is_pending"`
	IsDeafened  bool      `json:"is_deafened"`
	IsMuted     bool      `json:"is_muted"`
}

// GetMemberInfo gets detailed information about a guild member
func (c *DiscordClient) GetMemberInfo(ctx context.Context, userID string) (*DiscordMember, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	member, err := c.session.GuildMember(c.guildID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get member info: %w", err)
	}

	displayName := member.User.Username
	if member.Nick != "" {
		displayName = member.Nick
	} else if member.User.GlobalName != "" {
		displayName = member.User.GlobalName
	}

	return &DiscordMember{
		UserID:      member.User.ID,
		Username:    member.User.Username,
		DisplayName: displayName,
		Nickname:    member.Nick,
		Avatar:      member.User.Avatar,
		Roles:       member.Roles,
		JoinedAt:    member.JoinedAt,
		IsPending:   member.Pending,
		IsDeafened:  member.Deaf,
		IsMuted:     member.Mute,
	}, nil
}

// KickMember kicks a member from the guild
func (c *DiscordClient) KickMember(ctx context.Context, userID, reason string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberDeleteWithReason(c.guildID, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to kick member: %w", err)
	}
	return nil
}

// BanMember bans a member from the guild
func (c *DiscordClient) BanMember(ctx context.Context, userID, reason string, deleteMessageDays int) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	if deleteMessageDays < 0 || deleteMessageDays > 7 {
		deleteMessageDays = 0
	}

	err := c.session.GuildBanCreateWithReason(c.guildID, userID, reason, deleteMessageDays)
	if err != nil {
		return fmt.Errorf("failed to ban member: %w", err)
	}
	return nil
}

// UnbanMember removes a ban from a user
func (c *DiscordClient) UnbanMember(ctx context.Context, userID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildBanDelete(c.guildID, userID)
	if err != nil {
		return fmt.Errorf("failed to unban member: %w", err)
	}
	return nil
}

// TimeoutMember applies a timeout to a member (mutes them temporarily)
func (c *DiscordClient) TimeoutMember(ctx context.Context, userID string, duration time.Duration, reason string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	// Maximum timeout is 28 days
	if duration > 28*24*time.Hour {
		duration = 28 * 24 * time.Hour
	}

	timeout := time.Now().Add(duration)
	err := c.session.GuildMemberTimeout(c.guildID, userID, &timeout)
	if err != nil {
		return fmt.Errorf("failed to timeout member: %w", err)
	}
	return nil
}

// RemoveTimeout removes a timeout from a member
func (c *DiscordClient) RemoveTimeout(ctx context.Context, userID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberTimeout(c.guildID, userID, nil)
	if err != nil {
		return fmt.Errorf("failed to remove timeout: %w", err)
	}
	return nil
}

// SetNickname sets a member's nickname
func (c *DiscordClient) SetNickname(ctx context.Context, userID, nickname string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberNickname(c.guildID, userID, nickname)
	if err != nil {
		return fmt.Errorf("failed to set nickname: %w", err)
	}
	return nil
}

// ListMembers lists guild members with optional filtering
func (c *DiscordClient) ListMembers(ctx context.Context, limit int, after string) ([]DiscordMember, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	members, err := c.session.GuildMembers(c.guildID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	result := make([]DiscordMember, 0, len(members))
	for _, m := range members {
		displayName := m.User.Username
		if m.Nick != "" {
			displayName = m.Nick
		} else if m.User.GlobalName != "" {
			displayName = m.User.GlobalName
		}

		result = append(result, DiscordMember{
			UserID:      m.User.ID,
			Username:    m.User.Username,
			DisplayName: displayName,
			Nickname:    m.Nick,
			Avatar:      m.User.Avatar,
			Roles:       m.Roles,
			JoinedAt:    m.JoinedAt,
			IsPending:   m.Pending,
			IsDeafened:  m.Deaf,
			IsMuted:     m.Mute,
		})
	}

	return result, nil
}

// ListBans lists banned users
func (c *DiscordClient) ListBans(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	bans, err := c.session.GuildBans(c.guildID, limit, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list bans: %w", err)
	}

	result := make([]map[string]interface{}, 0, len(bans))
	for _, ban := range bans {
		result = append(result, map[string]interface{}{
			"user_id":  ban.User.ID,
			"username": ban.User.Username,
			"reason":   ban.Reason,
		})
	}

	return result, nil
}

// =============================================================================
// ADMIN: Voice Management
// =============================================================================

// VoiceChannelStatus represents the status of a voice channel
type VoiceChannelStatus struct {
	ChannelID   string   `json:"channel_id"`
	ChannelName string   `json:"channel_name"`
	Members     []string `json:"members"`
	MemberCount int      `json:"member_count"`
}

// GetVoiceStatus gets the status of voice channels
func (c *DiscordClient) GetVoiceStatus(ctx context.Context) ([]VoiceChannelStatus, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	guild, err := c.session.State.Guild(c.guildID)
	if err != nil {
		// Try fetching from API
		guild, err = c.session.Guild(c.guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to get guild: %w", err)
		}
	}

	// Map voice states to channels
	channelMembers := make(map[string][]string)
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID != "" {
			channelMembers[vs.ChannelID] = append(channelMembers[vs.ChannelID], vs.UserID)
		}
	}

	// Get channel names
	channels, err := c.session.GuildChannels(c.guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}

	result := make([]VoiceChannelStatus, 0)
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildVoice || ch.Type == discordgo.ChannelTypeGuildStageVoice {
			members := channelMembers[ch.ID]
			result = append(result, VoiceChannelStatus{
				ChannelID:   ch.ID,
				ChannelName: ch.Name,
				Members:     members,
				MemberCount: len(members),
			})
		}
	}

	return result, nil
}

// VoiceMove moves a member to a different voice channel
func (c *DiscordClient) VoiceMove(ctx context.Context, userID, channelID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberMove(c.guildID, userID, &channelID)
	if err != nil {
		return fmt.Errorf("failed to move member: %w", err)
	}
	return nil
}

// VoiceDisconnect disconnects a member from voice
func (c *DiscordClient) VoiceDisconnect(ctx context.Context, userID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	// Move to nil channel disconnects the user
	err := c.session.GuildMemberMove(c.guildID, userID, nil)
	if err != nil {
		return fmt.Errorf("failed to disconnect member: %w", err)
	}
	return nil
}

// VoiceMute server-mutes a member
func (c *DiscordClient) VoiceMute(ctx context.Context, userID string, muted bool) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberMute(c.guildID, userID, muted)
	if err != nil {
		return fmt.Errorf("failed to mute member: %w", err)
	}
	return nil
}

// VoiceDeafen server-deafens a member
func (c *DiscordClient) VoiceDeafen(ctx context.Context, userID string, deafened bool) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildMemberDeafen(c.guildID, userID, deafened)
	if err != nil {
		return fmt.Errorf("failed to deafen member: %w", err)
	}
	return nil
}

// =============================================================================
// ADMIN: Moderation
// =============================================================================

// PurgeMessages bulk deletes messages from a channel
func (c *DiscordClient) PurgeMessages(ctx context.Context, channelID string, count int) (int, error) {
	if count < 2 || count > 100 {
		return 0, fmt.Errorf("count must be between 2 and 100")
	}

	// Get messages to delete
	messages, err := c.session.ChannelMessages(channelID, count, "", "", "")
	if err != nil {
		return 0, fmt.Errorf("failed to get messages: %w", err)
	}

	// Filter out messages older than 14 days (Discord limitation)
	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)
	msgIDs := make([]string, 0, len(messages))
	for _, msg := range messages {
		if msg.Timestamp.After(twoWeeksAgo) {
			msgIDs = append(msgIDs, msg.ID)
		}
	}

	if len(msgIDs) < 2 {
		return 0, fmt.Errorf("not enough messages to bulk delete (minimum 2, found %d)", len(msgIDs))
	}

	err = c.session.ChannelMessagesBulkDelete(channelID, msgIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to delete messages: %w", err)
	}

	return len(msgIDs), nil
}

// PinMessage pins a message to a channel
func (c *DiscordClient) PinMessage(ctx context.Context, channelID, messageID string) error {
	err := c.session.ChannelMessagePin(channelID, messageID)
	if err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}
	return nil
}

// UnpinMessage unpins a message from a channel
func (c *DiscordClient) UnpinMessage(ctx context.Context, channelID, messageID string) error {
	err := c.session.ChannelMessageUnpin(channelID, messageID)
	if err != nil {
		return fmt.Errorf("failed to unpin message: %w", err)
	}
	return nil
}

// GetPinnedMessages gets pinned messages from a channel
func (c *DiscordClient) GetPinnedMessages(ctx context.Context, channelID string) ([]DiscordMessage, error) {
	messages, err := c.session.ChannelMessagesPinned(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pinned messages: %w", err)
	}

	result := make([]DiscordMessage, 0, len(messages))
	for _, msg := range messages {
		result = append(result, DiscordMessage{
			ID:        msg.ID,
			Content:   msg.Content,
			Author:    msg.Author.Username,
			Timestamp: msg.Timestamp,
			ChannelID: msg.ChannelID,
		})
	}

	return result, nil
}

// AuditLogEntry represents an entry in the audit log
type AuditLogEntry struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	TargetID   string                 `json:"target_id"`
	ActionType int                    `json:"action_type"`
	Reason     string                 `json:"reason,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// GetAuditLog retrieves the guild audit log
func (c *DiscordClient) GetAuditLog(ctx context.Context, actionType int, userID string, limit int) ([]AuditLogEntry, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	auditLog, err := c.session.GuildAuditLog(c.guildID, userID, "", actionType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	result := make([]AuditLogEntry, 0, len(auditLog.AuditLogEntries))
	for _, entry := range auditLog.AuditLogEntries {
		changes := make(map[string]interface{})
		for _, change := range entry.Changes {
			key := ""
			if change.Key != nil {
				key = string(*change.Key)
			}
			changes[key] = map[string]interface{}{
				"old": change.OldValue,
				"new": change.NewValue,
			}
		}

		// Parse snowflake ID to timestamp
		createdAt := snowflakeToTime(entry.ID)

		result = append(result, AuditLogEntry{
			ID:         entry.ID,
			UserID:     entry.UserID,
			TargetID:   entry.TargetID,
			ActionType: int(*entry.ActionType),
			Reason:     entry.Reason,
			Changes:    changes,
			CreatedAt:  createdAt,
		})
	}

	return result, nil
}

// =============================================================================
// ADMIN: Scheduled Events
// =============================================================================

// ScheduledEventType represents the type of scheduled event
type ScheduledEventType int

const (
	ScheduledEventStageInstance ScheduledEventType = 1
	ScheduledEventVoice         ScheduledEventType = 2
	ScheduledEventExternal      ScheduledEventType = 3
)

// ScheduledEvent represents a Discord scheduled event
type ScheduledEvent struct {
	ID          string             `json:"id"`
	GuildID     string             `json:"guild_id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	ChannelID   string             `json:"channel_id,omitempty"`
	Location    string             `json:"location,omitempty"`
	StartTime   time.Time          `json:"start_time"`
	EndTime     time.Time          `json:"end_time,omitempty"`
	EventType   ScheduledEventType `json:"event_type"`
	Status      int                `json:"status"`
	UserCount   int                `json:"user_count"`
}

// ScheduledEventCreateOptions specifies options for creating a scheduled event
type ScheduledEventCreateOptions struct {
	Name        string
	Description string
	ChannelID   string // For voice/stage events
	Location    string // For external events
	StartTime   time.Time
	EndTime     time.Time
	EventType   ScheduledEventType
}

// CreateScheduledEvent creates a new scheduled event
func (c *DiscordClient) CreateScheduledEvent(ctx context.Context, opts ScheduledEventCreateOptions) (*ScheduledEvent, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	params := &discordgo.GuildScheduledEventParams{
		Name:               opts.Name,
		Description:        opts.Description,
		ScheduledStartTime: &opts.StartTime,
		EntityType:         discordgo.GuildScheduledEventEntityType(opts.EventType),
		PrivacyLevel:       discordgo.GuildScheduledEventPrivacyLevelGuildOnly,
	}

	if opts.EventType == ScheduledEventExternal {
		params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
			Location: opts.Location,
		}
		params.ScheduledEndTime = &opts.EndTime
	} else {
		params.ChannelID = opts.ChannelID
	}

	event, err := c.session.GuildScheduledEventCreate(c.guildID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduled event: %w", err)
	}

	endTime := time.Time{}
	if event.ScheduledEndTime != nil {
		endTime = *event.ScheduledEndTime
	}

	return &ScheduledEvent{
		ID:          event.ID,
		GuildID:     event.GuildID,
		Name:        event.Name,
		Description: event.Description,
		ChannelID:   event.ChannelID,
		StartTime:   event.ScheduledStartTime,
		EndTime:     endTime,
		EventType:   ScheduledEventType(event.EntityType),
		Status:      int(event.Status),
		UserCount:   event.UserCount,
	}, nil
}

// UpdateScheduledEvent updates an existing scheduled event
func (c *DiscordClient) UpdateScheduledEvent(ctx context.Context, eventID string, opts ScheduledEventCreateOptions) (*ScheduledEvent, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	params := &discordgo.GuildScheduledEventParams{
		Name:               opts.Name,
		Description:        opts.Description,
		ScheduledStartTime: &opts.StartTime,
	}

	if opts.EventType == ScheduledEventExternal {
		params.EntityMetadata = &discordgo.GuildScheduledEventEntityMetadata{
			Location: opts.Location,
		}
		params.ScheduledEndTime = &opts.EndTime
	} else if opts.ChannelID != "" {
		params.ChannelID = opts.ChannelID
	}

	event, err := c.session.GuildScheduledEventEdit(c.guildID, eventID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update scheduled event: %w", err)
	}

	endTime := time.Time{}
	if event.ScheduledEndTime != nil {
		endTime = *event.ScheduledEndTime
	}

	return &ScheduledEvent{
		ID:          event.ID,
		GuildID:     event.GuildID,
		Name:        event.Name,
		Description: event.Description,
		ChannelID:   event.ChannelID,
		StartTime:   event.ScheduledStartTime,
		EndTime:     endTime,
		EventType:   ScheduledEventType(event.EntityType),
		Status:      int(event.Status),
		UserCount:   event.UserCount,
	}, nil
}

// DeleteScheduledEvent deletes a scheduled event
func (c *DiscordClient) DeleteScheduledEvent(ctx context.Context, eventID string) error {
	if c.guildID == "" {
		return fmt.Errorf("no guild ID configured")
	}

	err := c.session.GuildScheduledEventDelete(c.guildID, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete scheduled event: %w", err)
	}
	return nil
}

// ListScheduledEvents lists scheduled events in the guild
func (c *DiscordClient) ListScheduledEvents(ctx context.Context, withUserCount bool) ([]ScheduledEvent, error) {
	if c.guildID == "" {
		return nil, fmt.Errorf("no guild ID configured")
	}

	events, err := c.session.GuildScheduledEvents(c.guildID, withUserCount)
	if err != nil {
		return nil, fmt.Errorf("failed to list scheduled events: %w", err)
	}

	result := make([]ScheduledEvent, 0, len(events))
	for _, event := range events {
		location := event.EntityMetadata.Location

		endTime := time.Time{}
		if event.ScheduledEndTime != nil {
			endTime = *event.ScheduledEndTime
		}

		result = append(result, ScheduledEvent{
			ID:          event.ID,
			GuildID:     event.GuildID,
			Name:        event.Name,
			Description: event.Description,
			ChannelID:   event.ChannelID,
			Location:    location,
			StartTime:   event.ScheduledStartTime,
			EndTime:     endTime,
			EventType:   ScheduledEventType(event.EntityType),
			Status:      int(event.Status),
			UserCount:   event.UserCount,
		})
	}

	return result, nil
}

// =============================================================================
// ADMIN: Interactive Components
// =============================================================================

// Button represents a Discord button component
type Button struct {
	Label    string
	Style    discordgo.ButtonStyle
	CustomID string
	URL      string // For link buttons only
	Disabled bool
	Emoji    string
}

// SelectOption represents a select menu option
type SelectOption struct {
	Label       string
	Value       string
	Description string
	Emoji       string
	Default     bool
}

// SendMessageWithButtons sends a message with action buttons
func (c *DiscordClient) SendMessageWithButtons(ctx context.Context, channelID, content string, buttons []Button) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}

	components := make([]discordgo.MessageComponent, 0)
	buttonRow := make([]discordgo.MessageComponent, 0, len(buttons))

	for _, btn := range buttons {
		b := discordgo.Button{
			Label:    btn.Label,
			Style:    btn.Style,
			CustomID: btn.CustomID,
			Disabled: btn.Disabled,
		}
		if btn.Style == discordgo.LinkButton {
			b.URL = btn.URL
		}
		if btn.Emoji != "" {
			b.Emoji = &discordgo.ComponentEmoji{Name: btn.Emoji}
		}
		buttonRow = append(buttonRow, b)
	}

	if len(buttonRow) > 0 {
		components = append(components, discordgo.ActionsRow{Components: buttonRow})
	}

	msg, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    content,
		Components: components,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message with buttons: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   msg.Content,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// SendMessageWithSelect sends a message with a select menu
func (c *DiscordClient) SendMessageWithSelect(ctx context.Context, channelID, content, placeholder, customID string, options []SelectOption) (*DiscordMessage, error) {
	if channelID == "" {
		channelID = c.channelID
	}

	selectOptions := make([]discordgo.SelectMenuOption, 0, len(options))
	for _, opt := range options {
		o := discordgo.SelectMenuOption{
			Label:       opt.Label,
			Value:       opt.Value,
			Description: opt.Description,
			Default:     opt.Default,
		}
		if opt.Emoji != "" {
			o.Emoji = &discordgo.ComponentEmoji{Name: opt.Emoji}
		}
		selectOptions = append(selectOptions, o)
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    customID,
					Placeholder: placeholder,
					Options:     selectOptions,
				},
			},
		},
	}

	msg, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    content,
		Components: components,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send message with select: %w", err)
	}

	return &DiscordMessage{
		ID:        msg.ID,
		Content:   msg.Content,
		Author:    msg.Author.Username,
		Timestamp: msg.Timestamp,
		ChannelID: msg.ChannelID,
	}, nil
}

// UpdateMessageComponents updates the components of an existing message
func (c *DiscordClient) UpdateMessageComponents(ctx context.Context, channelID, messageID string, components []discordgo.MessageComponent) error {
	_, err := c.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Components: &components,
	})
	if err != nil {
		return fmt.Errorf("failed to update message components: %w", err)
	}
	return nil
}

// RespondToComponentInteraction responds to a button/select interaction
func (c *DiscordClient) RespondToComponentInteraction(i *discordgo.InteractionCreate, content string, ephemeral bool) error {
	flags := discordgo.MessageFlags(0)
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	return c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   flags,
		},
	})
}

// UpdateOriginalInteractionMessage updates the original message from a component interaction
func (c *DiscordClient) UpdateOriginalInteractionMessage(i *discordgo.InteractionCreate, content string, components []discordgo.MessageComponent) error {
	return c.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: components,
		},
	})
}

// =============================================================================
// ADMIN: Guild/Server Info
// =============================================================================

// GuildInfo represents detailed guild information
type GuildInfo struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Icon              string    `json:"icon,omitempty"`
	OwnerID           string    `json:"owner_id"`
	MemberCount       int       `json:"member_count"`
	ChannelCount      int       `json:"channel_count"`
	RoleCount         int       `json:"role_count"`
	EmojiCount        int       `json:"emoji_count"`
	BoostLevel        int       `json:"boost_level"`
	BoostCount        int       `json:"boost_count"`
	VerificationLevel int       `json:"verification_level"`
	Description       string    `json:"description,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// GetServerInfo gets detailed information about the guild (cached 60s)
func (c *DiscordClient) GetServerInfo(ctx context.Context) (*GuildInfo, error) {
	return discordServerInfoCache.GetOrFetch(ctx, func(ctx context.Context) (*GuildInfo, error) {
		if c.guildID == "" {
			return nil, fmt.Errorf("no guild ID configured")
		}

		guild, err := c.session.Guild(c.guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to get guild: %w", err)
		}

		channels, _ := c.session.GuildChannels(c.guildID)

		return &GuildInfo{
			ID:                guild.ID,
			Name:              guild.Name,
			Icon:              guild.Icon,
			OwnerID:           guild.OwnerID,
			MemberCount:       guild.MemberCount,
			ChannelCount:      len(channels),
			RoleCount:         len(guild.Roles),
			EmojiCount:        len(guild.Emojis),
			BoostLevel:        int(guild.PremiumTier),
			BoostCount:        guild.PremiumSubscriptionCount,
			VerificationLevel: int(guild.VerificationLevel),
			Description:       guild.Description,
			CreatedAt:         snowflakeToTime(guild.ID),
		}, nil
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// channelTypeToString converts Discord channel type to string
func channelTypeToString(t discordgo.ChannelType) string {
	switch t {
	case discordgo.ChannelTypeGuildText:
		return "text"
	case discordgo.ChannelTypeGuildVoice:
		return "voice"
	case discordgo.ChannelTypeGuildCategory:
		return "category"
	case discordgo.ChannelTypeGuildNews:
		return "news"
	case discordgo.ChannelTypeGuildForum:
		return "forum"
	case discordgo.ChannelTypeGuildStageVoice:
		return "stage"
	case discordgo.ChannelTypeGuildPublicThread:
		return "public_thread"
	case discordgo.ChannelTypeGuildPrivateThread:
		return "private_thread"
	default:
		return "unknown"
	}
}

// snowflakeToTime converts a Discord snowflake ID to a time.Time
func snowflakeToTime(snowflake string) time.Time {
	const discordEpoch = 1420070400000
	var id uint64
	fmt.Sscanf(snowflake, "%d", &id)
	timestamp := int64(id>>22) + discordEpoch
	return time.UnixMilli(timestamp)
}

// HasRole checks if a member has a specific role
func HasRole(member *discordgo.Member, roleID string) bool {
	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// HasAnyRole checks if a member has any of the specified roles
func HasAnyRole(member *discordgo.Member, roleIDs ...string) bool {
	for _, roleID := range roleIDs {
		if HasRole(member, roleID) {
			return true
		}
	}
	return false
}
