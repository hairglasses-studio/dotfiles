package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

// getCommands returns all slash commands for the bot
func getCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Show available commands and bot information",
		},
		{
			Name:        "sync",
			Description: "Trigger music library sync",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "source",
					Description: "Sync source: all, beatport, soundcloud, rekordbox (default: all)",
					Required:    false,
				},
			},
		},
		{
			Name:        "status",
			Description: "Check overall system status",
		},
		{
			Name:        "health",
			Description: "Run comprehensive health checks",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "system",
					Description: "System to check: all, td, obs, music, aws (default: all)",
					Required:    false,
				},
			},
		},
		{
			Name:        "whoami",
			Description: "Show bot identity and configuration",
		},
		{
			Name:        "tools",
			Description: "List available MCP tools",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Filter by category: discord, sync, studio, health, all (default: all)",
					Required:    false,
				},
			},
		},
		{
			Name:        "td",
			Description: "TouchDesigner controls",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "Get TouchDesigner status",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "health",
					Description: "Check TouchDesigner health",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "restart",
					Description: "Request TouchDesigner restart",
				},
			},
		},
		{
			Name:        "alert",
			Description: "Send a notification alert",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Alert message",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "level",
					Description: "Alert level: info, warning, error, success (default: info)",
					Required:    false,
				},
			},
		},
		{
			Name:        "studio-status",
			Description: "Get comprehensive studio system status (TD, OBS, Lighting, Streaming)",
		},
		{
			Name:        "stream-status",
			Description: "Check current streaming status",
		},
		{
			Name:        "td-health",
			Description: "Check TouchDesigner project health",
		},
		{
			Name:        "lighting",
			Description: "Get lighting system status",
		},
		{
			Name:        "ping",
			Description: "Check bot latency",
		},
		{
			Name:        "ai",
			Description: "Ask the AI assistant a question",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "question",
					Description: "Your question for the AI",
					Required:    true,
				},
			},
		},
		{
			Name:        "session",
			Description: "Studio session management",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "start",
					Description: "Start a new studio session",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "Session name",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "end",
					Description: "End the current session",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "status",
					Description: "Check current session status",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "log",
					Description: "Log an event to the current session",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "event",
							Description: "Event description",
							Required:    true,
						},
					},
				},
			},
		},
		// Admin command with subcommands (requires admin permission)
		{
			Name:                     "admin",
			Description:              "Server administration commands (requires admin)",
			DefaultMemberPermissions: int64Ptr(discordgo.PermissionAdministrator),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        "channel",
					Description: "Channel management",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "create",
							Description: "Create a new channel",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Channel name", Required: true},
								{Type: discordgo.ApplicationCommandOptionString, Name: "type", Description: "Channel type: text, voice"},
								{Type: discordgo.ApplicationCommandOptionString, Name: "category", Description: "Category ID"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "delete",
							Description: "Delete a channel",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionString, Name: "channel", Description: "Channel ID", Required: true},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "lock",
							Description: "Lock/unlock a channel",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionString, Name: "channel", Description: "Channel ID", Required: true},
								{Type: discordgo.ApplicationCommandOptionBoolean, Name: "locked", Description: "Lock state", Required: true},
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        "role",
					Description: "Role management",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "create",
							Description: "Create a new role",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Role name", Required: true},
								{Type: discordgo.ApplicationCommandOptionString, Name: "color", Description: "Color hex (e.g., #FF0000)"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "assign",
							Description: "Assign a role to a user",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Role", Required: true},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "revoke",
							Description: "Revoke a role from a user",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Role", Required: true},
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        "member",
					Description: "Member management",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "info",
							Description: "Get member info",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "kick",
							Description: "Kick a member",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionString, Name: "reason", Description: "Reason"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "ban",
							Description: "Ban a member",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionString, Name: "reason", Description: "Reason"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "timeout",
							Description: "Timeout a member",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionInteger, Name: "minutes", Description: "Duration in minutes", Required: true},
								{Type: discordgo.ApplicationCommandOptionString, Name: "reason", Description: "Reason"},
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        "voice",
					Description: "Voice channel management",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "status",
							Description: "Get voice channel status",
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "move",
							Description: "Move user to another voice channel",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "Voice channel", Required: true},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "disconnect",
							Description: "Disconnect user from voice",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Name:        "mute",
							Description: "Server mute/unmute user",
							Options: []*discordgo.ApplicationCommandOption{
								{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
								{Type: discordgo.ApplicationCommandOptionBoolean, Name: "muted", Description: "Mute state", Required: true},
							},
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "purge",
					Description: "Delete messages from current channel",
					Options: []*discordgo.ApplicationCommandOption{
						{Type: discordgo.ApplicationCommandOptionInteger, Name: "count", Description: "Number of messages (2-100)", Required: true},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "server",
					Description: "Get server info",
				},
			},
		},
	}
}

// int64Ptr returns a pointer to an int64
func int64Ptr(i int64) *int64 {
	return &i
}

// getCommandHandlers returns the handlers for each command
func getCommandHandlers() map[string]CommandHandler {
	return map[string]CommandHandler{
		"help":          handleHelp,
		"sync":          handleSync,
		"status":        handleStatus,
		"health":        handleHealth,
		"whoami":        handleWhoami,
		"tools":         handleTools,
		"td":            handleTD,
		"alert":         handleAlert,
		"studio-status": handleStudioStatus,
		"stream-status": handleStreamStatus,
		"td-health":     handleTDHealth,
		"lighting":      handleLighting,
		"ping":          handlePing,
		"ai":            handleAI,
		"session":       handleSession,
		"admin":         requireAdmin(handleAdmin),
	}
}

// respondEmbed sends an embed response to an interaction
func respondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// respondText sends a text response to an interaction
func respondText(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// handleHelp shows available commands
func handleHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Aftrs Studio Bot",
		Description: "I help manage The Aftrs audiovisual studio. Here are my commands:",
		Color:       0x5865F2, // Discord blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "/studio-status",
				Value:  "Get comprehensive studio system status",
				Inline: true,
			},
			{
				Name:   "/stream-status",
				Value:  "Check streaming status",
				Inline: true,
			},
			{
				Name:   "/td-health",
				Value:  "TouchDesigner project health",
				Inline: true,
			},
			{
				Name:   "/lighting",
				Value:  "Lighting system status",
				Inline: true,
			},
			{
				Name:   "/ai [question]",
				Value:  "Ask the AI assistant",
				Inline: true,
			},
			{
				Name:   "/session start|end|status|log",
				Value:  "Manage studio sessions",
				Inline: true,
			},
			{
				Name:   "/ping",
				Value:  "Check bot latency",
				Inline: true,
			},
			{
				Name:   "/help",
				Value:  "Show this help message",
				Inline: true,
			},
			{
				Name:   "@mention or DM",
				Value:  "Chat with AI assistant",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handleStudioStatus shows comprehensive studio status
func handleStudioStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	// Defer response since this might take a moment
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var fields []*discordgo.MessageEmbedField
	overallScore := 0
	systemsChecked := 0

	// Check TouchDesigner
	tdClient, err := clients.NewTouchDesignerClient()
	if err == nil {
		tdStatus, err := tdClient.GetStatus(ctx)
		if err == nil {
			// Calculate health from status
			score, status := calculateTDHealth(tdStatus)
			statusText := fmt.Sprintf("Score: %d/100\nStatus: %s\nFPS: %.1f", score, status, tdStatus.FPS)
			if tdStatus.ProjectName != "" {
				statusText += fmt.Sprintf("\nProject: %s", tdStatus.ProjectName)
			}
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   getStatusEmoji(status) + " TouchDesigner",
				Value:  statusText,
				Inline: true,
			})
			overallScore += score
			systemsChecked++
		}
	}

	// Check OBS
	obsClient, err := clients.NewOBSClient()
	if err == nil {
		health, err := obsClient.GetHealth(ctx)
		if err == nil {
			statusText := fmt.Sprintf("Score: %d/100\nStatus: %s", health.Score, health.Status)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   getStatusEmoji(health.Status) + " OBS Studio",
				Value:  statusText,
				Inline: true,
			})
			overallScore += health.Score
			systemsChecked++
		}
	}

	// Check Lighting
	lightingClient, err := clients.NewLightingClient()
	if err == nil {
		health, err := lightingClient.GetHealth(ctx)
		if err == nil {
			statusText := fmt.Sprintf("Score: %d/100\nStatus: %s", health.Score, health.Status)
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   getStatusEmoji(health.Status) + " Lighting",
				Value:  statusText,
				Inline: true,
			})
			overallScore += health.Score
			systemsChecked++
		}
	}

	// Calculate overall
	avgScore := 0
	overallStatus := "unknown"
	if systemsChecked > 0 {
		avgScore = overallScore / systemsChecked
		switch {
		case avgScore >= 80:
			overallStatus = "healthy"
		case avgScore >= 50:
			overallStatus = "degraded"
		default:
			overallStatus = "critical"
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       getStatusEmoji(overallStatus) + " Studio Status",
		Description: fmt.Sprintf("Overall Health: **%d/100** (%s)\nSystems Checked: %d", avgScore, overallStatus, systemsChecked),
		Color:       getStatusColor(overallStatus),
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// calculateTDHealth derives health score and status from TDStatus
func calculateTDHealth(status *clients.TDStatus) (int, string) {
	score := 100

	if !status.Connected {
		return 0, "critical"
	}

	// FPS check
	if status.FPS < 30 {
		score -= 30
	} else if status.FPS < 50 {
		score -= 10
	}

	// Error check
	if status.ErrorCount > 0 {
		score -= status.ErrorCount * 10
	}

	// Warning check
	if status.WarningCount > 0 {
		score -= status.WarningCount * 5
	}

	// Clamp score
	if score < 0 {
		score = 0
	}

	// Determine status
	var healthStatus string
	switch {
	case score >= 80:
		healthStatus = "healthy"
	case score >= 50:
		healthStatus = "degraded"
	default:
		healthStatus = "critical"
	}

	return score, healthStatus
}

// handleStreamStatus shows streaming status
func handleStreamStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	obsClient, err := clients.NewOBSClient()
	if err != nil {
		respondText(s, i, "Could not connect to OBS: "+err.Error())
		return
	}

	status, err := obsClient.GetStatus(ctx)
	if err != nil {
		respondText(s, i, "Could not get OBS status: "+err.Error())
		return
	}

	streamStatus := "Offline"
	streamEmoji := ":red_circle:"
	if status.Streaming {
		streamStatus = fmt.Sprintf("**LIVE** for %s", status.StreamTime)
		streamEmoji = ":red_circle:"
	}

	recordStatus := "Not Recording"
	if status.Recording {
		recordStatus = fmt.Sprintf("Recording for %s", status.RecordTime)
	}

	embed := &discordgo.MessageEmbed{
		Title: streamEmoji + " Stream Status",
		Color: getStreamColor(status.Streaming),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Stream",
				Value:  streamStatus,
				Inline: true,
			},
			{
				Name:   "Recording",
				Value:  recordStatus,
				Inline: true,
			},
			{
				Name:   "Current Scene",
				Value:  status.CurrentScene,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handleTDHealth shows TouchDesigner health
func handleTDHealth(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	tdClient, err := clients.NewTouchDesignerClient()
	if err != nil {
		respondText(s, i, "Could not connect to TouchDesigner: "+err.Error())
		return
	}

	tdStatus, err := tdClient.GetStatus(ctx)
	if err != nil {
		respondText(s, i, "Could not get TD status: "+err.Error())
		return
	}

	score, status := calculateTDHealth(tdStatus)

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Health Score",
			Value:  fmt.Sprintf("%d/100", score),
			Inline: true,
		},
		{
			Name:   "Status",
			Value:  status,
			Inline: true,
		},
		{
			Name:   "FPS",
			Value:  fmt.Sprintf("%.1f", tdStatus.FPS),
			Inline: true,
		},
	}

	if tdStatus.ProjectName != "" {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Project",
			Value:  tdStatus.ProjectName,
			Inline: false,
		})
	}

	var issues []string
	if tdStatus.ErrorCount > 0 {
		issues = append(issues, fmt.Sprintf("%d errors", tdStatus.ErrorCount))
	}
	if tdStatus.WarningCount > 0 {
		issues = append(issues, fmt.Sprintf("%d warnings", tdStatus.WarningCount))
	}
	if len(issues) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Issues",
			Value:  strings.Join(issues, "\n"),
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  getStatusEmoji(status) + " TouchDesigner Health",
		Color:  getStatusColor(status),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handleLighting shows lighting system status
func handleLighting(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	lightingClient, err := clients.NewLightingClient()
	if err != nil {
		respondText(s, i, "Could not connect to lighting system: "+err.Error())
		return
	}

	health, err := lightingClient.GetHealth(ctx)
	if err != nil {
		respondText(s, i, "Could not get lighting health: "+err.Error())
		return
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "Health Score",
			Value:  fmt.Sprintf("%d/100", health.Score),
			Inline: true,
		},
		{
			Name:   "Status",
			Value:  health.Status,
			Inline: true,
		},
		{
			Name:   "Universes",
			Value:  fmt.Sprintf("%d active", health.UniverseCount),
			Inline: true,
		},
		{
			Name:   "Fixtures",
			Value:  fmt.Sprintf("%d configured", health.FixtureCount),
			Inline: true,
		},
		{
			Name:   "ArtNet Nodes",
			Value:  fmt.Sprintf("%d online", health.NodesOnline),
			Inline: true,
		},
	}

	if len(health.Issues) > 0 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Issues",
			Value:  strings.Join(health.Issues, "\n"),
			Inline: false,
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:  getStatusEmoji(health.Status) + " Lighting System",
		Color:  getStatusColor(health.Status),
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handlePing shows bot latency
func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	latency := s.HeartbeatLatency().Milliseconds()

	embed := &discordgo.MessageEmbed{
		Title:       ":ping_pong: Pong!",
		Description: fmt.Sprintf("Latency: **%dms**", latency),
		Color:       0x00FF00,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// Helper functions

func getStatusEmoji(status string) string {
	switch status {
	case "healthy":
		return ":green_circle:"
	case "degraded":
		return ":yellow_circle:"
	case "critical":
		return ":red_circle:"
	default:
		return ":white_circle:"
	}
}

func getStatusColor(status string) int {
	switch status {
	case "healthy":
		return 0x00FF00 // Green
	case "degraded":
		return 0xFFA500 // Orange
	case "critical":
		return 0xFF0000 // Red
	default:
		return 0x808080 // Gray
	}
}

func getStreamColor(streaming bool) int {
	if streaming {
		return 0xFF0000 // Red for live
	}
	return 0x808080 // Gray for offline
}

// Package-level instances for use by command handlers
var (
	globalAIHandler     *AIHandler
	globalSessionLogger *SessionLogger
)

// SetGlobalHandlers sets the global handlers for command use
func SetGlobalHandlers(ai *AIHandler, session *SessionLogger) {
	globalAIHandler = ai
	globalSessionLogger = session
}

// handleAI handles the /ai command
func handleAI(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondText(s, i, "Please provide a question.")
		return
	}

	question := options[0].StringValue()

	if globalAIHandler == nil || !globalAIHandler.IsConfigured() {
		respondText(s, i, "AI assistant is not configured. Set `ANTHROPIC_API_KEY` to enable.")
		return
	}

	// Defer response since AI might take a moment
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Get context and call AI
	ctx := context.Background()
	contextInfo := globalAIHandler.gatherContext(ctx, i.ChannelID, s)

	fullPrompt := question
	if contextInfo != "" {
		fullPrompt = fmt.Sprintf("%s\n\n---\nUser question: %s", contextInfo, question)
	}

	messages := []anthropicMessage{{Role: "user", Content: fullPrompt}}

	response, err := globalAIHandler.callClaude(ctx, messages)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr(fmt.Sprintf("Error: %s", err.Error())),
		})
		return
	}

	// Truncate if needed
	if len(response) > 2000 {
		response = response[:1997] + "..."
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &response,
	})
}

// handleSession handles the /session command
func handleSession(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondText(s, i, "Please provide a subcommand: start, end, status, or log")
		return
	}

	subcommand := options[0].Name
	ctx := context.Background()

	switch subcommand {
	case "start":
		handleSessionStart(s, i, ctx, options[0].Options)
	case "end":
		handleSessionEnd(s, i, ctx)
	case "status":
		handleSessionStatus(s, i)
	case "log":
		handleSessionLog(s, i, options[0].Options)
	default:
		respondText(s, i, "Unknown subcommand: "+subcommand)
	}
}

func handleSessionStart(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if globalSessionLogger == nil {
		respondText(s, i, "Session logger not initialized.")
		return
	}

	if len(options) == 0 {
		respondText(s, i, "Please provide a session name.")
		return
	}

	name := options[0].StringValue()
	operator := i.Member.User.Username

	session, err := globalSessionLogger.StartSession(ctx, name, operator)
	if err != nil {
		respondText(s, i, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":green_circle: Session Started",
		Description: fmt.Sprintf("**%s** started session: **%s**", operator, name),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Session ID",
				Value:  session.ID,
				Inline: true,
			},
			{
				Name:   "Started",
				Value:  session.StartTime.Format("3:04 PM"),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Session Log",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

func handleSessionEnd(s *discordgo.Session, i *discordgo.InteractionCreate, ctx context.Context) {
	if globalSessionLogger == nil {
		respondText(s, i, "Session logger not initialized.")
		return
	}

	session, err := globalSessionLogger.EndSession(ctx)
	if err != nil {
		respondText(s, i, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	duration := "Unknown"
	if session.EndTime != nil {
		d := session.EndTime.Sub(session.StartTime)
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if hours > 0 {
			duration = fmt.Sprintf("%dh %dm", hours, minutes)
		} else {
			duration = fmt.Sprintf("%dm", minutes)
		}
	}

	highlightText := "None"
	if len(session.Highlights) > 0 {
		highlightText = strings.Join(session.Highlights, "\n")
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":blue_circle: Session Ended",
		Description: fmt.Sprintf("Session **%s** has ended", session.Name),
		Color:       0x0099FF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  duration,
				Inline: true,
			},
			{
				Name:   "Events",
				Value:  fmt.Sprintf("%d logged", len(session.Events)),
				Inline: true,
			},
			{
				Name:  "Highlights",
				Value: highlightText,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Session Log",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

func handleSessionStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if globalSessionLogger == nil {
		respondText(s, i, "Session logger not initialized.")
		return
	}

	session := globalSessionLogger.GetActiveSession()
	if session == nil {
		embed := &discordgo.MessageEmbed{
			Title:       ":white_circle: No Active Session",
			Description: "Use `/session start [name]` to begin a new session.",
			Color:       0x808080,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "aftrs-bot | Session Log",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		respondEmbed(s, i, embed)
		return
	}

	duration := time.Since(session.StartTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	durationStr := fmt.Sprintf("%dm", minutes)
	if hours > 0 {
		durationStr = fmt.Sprintf("%dh %dm", hours, minutes)
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":green_circle: Active Session",
		Description: fmt.Sprintf("**%s** (by %s)", session.Name, session.Operator),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  durationStr,
				Inline: true,
			},
			{
				Name:   "Events",
				Value:  fmt.Sprintf("%d logged", len(session.Events)),
				Inline: true,
			},
			{
				Name:   "Highlights",
				Value:  fmt.Sprintf("%d recorded", len(session.Highlights)),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Session Log",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

func handleSessionLog(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if globalSessionLogger == nil {
		respondText(s, i, "Session logger not initialized.")
		return
	}

	if len(options) == 0 {
		respondText(s, i, "Please provide an event description.")
		return
	}

	event := options[0].StringValue()

	err := globalSessionLogger.LogEvent("note", event, "")
	if err != nil {
		respondText(s, i, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	respondText(s, i, fmt.Sprintf(":pencil: Logged: %s", event))
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// handleSync handles the /sync command
func handleSync(s *discordgo.Session, i *discordgo.InteractionCreate) {
	source := "all"
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		source = options[0].StringValue()
	}

	// Defer response since sync might take a moment
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// For now, just acknowledge - actual sync would call the sync service
	embed := &discordgo.MessageEmbed{
		Title:       ":arrows_counterclockwise: Music Sync",
		Description: fmt.Sprintf("Sync request queued for **%s**\n\nUse `sync_trigger` MCP tool for immediate sync.", source),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Source",
				Value:  source,
				Inline: true,
			},
			{
				Name:   "Status",
				Value:  "Queued",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Music Sync",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// handleStatus handles the /status command
func handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// This is a simplified version of studio-status
	handleStudioStatus(s, i)
}

// handleHealth handles the /health command
func handleHealth(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	system := "all"
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		system = options[0].StringValue()
	}

	// Defer response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var fields []*discordgo.MessageEmbedField
	overallScore := 0
	systemsChecked := 0

	// Check TouchDesigner
	if system == "all" || system == "td" {
		tdClient, err := clients.NewTouchDesignerClient()
		if err == nil {
			tdStatus, err := tdClient.GetStatus(ctx)
			if err == nil {
				score, status := calculateTDHealth(tdStatus)
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   getStatusEmoji(status) + " TouchDesigner",
					Value:  fmt.Sprintf("Score: %d/100 | %s", score, status),
					Inline: true,
				})
				overallScore += score
				systemsChecked++
			}
		}
	}

	// Check OBS
	if system == "all" || system == "obs" {
		obsClient, err := clients.NewOBSClient()
		if err == nil {
			health, err := obsClient.GetHealth(ctx)
			if err == nil {
				fields = append(fields, &discordgo.MessageEmbedField{
					Name:   getStatusEmoji(health.Status) + " OBS Studio",
					Value:  fmt.Sprintf("Score: %d/100 | %s", health.Score, health.Status),
					Inline: true,
				})
				overallScore += health.Score
				systemsChecked++
			}
		}
	}

	// Calculate overall
	avgScore := 0
	overallStatus := "unknown"
	if systemsChecked > 0 {
		avgScore = overallScore / systemsChecked
		switch {
		case avgScore >= 80:
			overallStatus = "healthy"
		case avgScore >= 50:
			overallStatus = "degraded"
		default:
			overallStatus = "critical"
		}
	} else {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  ":warning: No Systems Reachable",
			Value: "Could not connect to any monitored systems",
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       getStatusEmoji(overallStatus) + " Health Check",
		Description: fmt.Sprintf("Overall: **%d/100** (%s) | Systems: %d", avgScore, overallStatus, systemsChecked),
		Color:       getStatusColor(overallStatus),
		Fields:      fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Health Check",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
}

// handleWhoami handles the /whoami command
func handleWhoami(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := s.State.User
	latency := s.HeartbeatLatency().Milliseconds()

	var guildName string
	guild, err := s.Guild(i.GuildID)
	if err == nil {
		guildName = guild.Name
	}

	embed := &discordgo.MessageEmbed{
		Title: ":robot: Bot Identity",
		Color: 0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Bot Name",
				Value:  user.Username,
				Inline: true,
			},
			{
				Name:   "Bot ID",
				Value:  user.ID,
				Inline: true,
			},
			{
				Name:   "Latency",
				Value:  fmt.Sprintf("%dms", latency),
				Inline: true,
			},
			{
				Name:   "Server",
				Value:  guildName,
				Inline: true,
			},
			{
				Name:   "MCP Tools",
				Value:  "180+",
				Inline: true,
			},
			{
				Name:   "Version",
				Value:  "1.0.0",
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | The Aftrs Studio",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handleTools handles the /tools command
func handleTools(s *discordgo.Session, i *discordgo.InteractionCreate) {
	category := "all"
	options := i.ApplicationCommandData().Options
	if len(options) > 0 {
		category = options[0].StringValue()
	}

	// Define tool categories
	toolCategories := map[string][]string{
		"discord": {"aftrs_discord_status", "aftrs_discord_send", "aftrs_discord_notify", "aftrs_discord_webhook"},
		"sync":    {"sync_trigger", "sync_status", "sync_health", "sync_stats"},
		"studio":  {"td_status", "td_health", "obs_status", "resolume_status", "lighting_status"},
		"health":  {"health_all", "health_infrastructure", "health_services"},
	}

	var sb strings.Builder
	sb.WriteString("**Available MCP Tools**\n\n")

	if category == "all" {
		for cat, tools := range toolCategories {
			sb.WriteString(fmt.Sprintf("**%s:**\n", strings.Title(cat)))
			for _, tool := range tools {
				sb.WriteString(fmt.Sprintf("• `%s`\n", tool))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("_...and 160+ more tools_")
	} else if tools, ok := toolCategories[category]; ok {
		sb.WriteString(fmt.Sprintf("**%s Tools:**\n", strings.Title(category)))
		for _, tool := range tools {
			sb.WriteString(fmt.Sprintf("• `%s`\n", tool))
		}
	} else {
		sb.WriteString(fmt.Sprintf("Unknown category: %s\n\nAvailable: discord, sync, studio, health, all", category))
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":wrench: MCP Tools",
		Description: sb.String(),
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Tool Discovery",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// handleTD handles the /td command
func handleTD(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondText(s, i, "Please provide a subcommand: status, health, or restart")
		return
	}

	subcommand := options[0].Name
	ctx := context.Background()

	switch subcommand {
	case "status":
		tdClient, err := clients.NewTouchDesignerClient()
		if err != nil {
			respondText(s, i, "Could not connect to TouchDesigner: "+err.Error())
			return
		}

		status, err := tdClient.GetStatus(ctx)
		if err != nil {
			respondText(s, i, "Could not get TD status: "+err.Error())
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: ":art: TouchDesigner Status",
			Color: 0x00FF00,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Connected",
					Value:  fmt.Sprintf("%v", status.Connected),
					Inline: true,
				},
				{
					Name:   "FPS",
					Value:  fmt.Sprintf("%.1f", status.FPS),
					Inline: true,
				},
				{
					Name:   "Project",
					Value:  status.ProjectName,
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "aftrs-bot | TouchDesigner",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		respondEmbed(s, i, embed)

	case "health":
		handleTDHealth(s, i)

	case "restart":
		respondText(s, i, ":warning: TD restart requested. Use MCP tool `td_restart` for actual restart.")
	}
}

// handleAlert handles the /alert command
func handleAlert(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondText(s, i, "Please provide a message for the alert.")
		return
	}

	message := options[0].StringValue()
	level := "info"
	if len(options) > 1 {
		level = options[1].StringValue()
	}

	// Choose color and emoji based on level
	var color int
	var emoji string
	switch level {
	case "error":
		color = 0xFF0000
		emoji = ":x:"
	case "warning":
		color = 0xFFAA00
		emoji = ":warning:"
	case "success":
		color = 0x00FF00
		emoji = ":white_check_mark:"
	default:
		color = 0x0099FF
		emoji = ":information_source:"
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Alert", emoji),
		Description: message,
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Level",
				Value:  level,
				Inline: true,
			},
			{
				Name:   "From",
				Value:  i.Member.User.Username,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Alerts",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// =============================================================================
// Admin Commands with Permission Checks
// =============================================================================

// hasAnyRole checks if a member has any of the specified roles
func hasAnyRole(member *discordgo.Member, roleIDs []string) bool {
	if member == nil {
		return false
	}
	for _, memberRole := range member.Roles {
		for _, allowedRole := range roleIDs {
			if memberRole == allowedRole {
				return true
			}
		}
	}
	return false
}

// requireAdmin is a permission check wrapper for admin-only commands
func requireAdmin(handler CommandHandler) CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Server owner always has access
		guild, err := s.Guild(i.GuildID)
		if err == nil && guild.OwnerID == i.Member.User.ID {
			handler(s, i)
			return
		}

		// Check if user has Discord admin permission
		perms := i.Member.Permissions
		if perms&discordgo.PermissionAdministrator == 0 {
			respondText(s, i, ":x: **Permission Denied**\n\nThis command requires the Administrator permission.")
			return
		}

		handler(s, i)
	}
}

// requireModerator is a permission check wrapper for moderation commands
func requireModerator(handler CommandHandler) CommandHandler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// Check admin permission first
		perms := i.Member.Permissions
		if perms&discordgo.PermissionAdministrator != 0 {
			handler(s, i)
			return
		}

		// Check moderator permissions
		if perms&discordgo.PermissionModerateMembers != 0 ||
			perms&discordgo.PermissionBanMembers != 0 ||
			perms&discordgo.PermissionKickMembers != 0 {
			handler(s, i)
			return
		}

		respondText(s, i, ":x: **Permission Denied**\n\nThis command requires moderation privileges.")
	}
}

// handleAdmin handles the /admin command with subcommands
func handleAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		respondText(s, i, "Please provide a subcommand: channel, role, member, voice, purge, server")
		return
	}

	subcommand := options[0].Name
	subOptions := options[0].Options

	switch subcommand {
	case "channel":
		handleAdminChannel(s, i, subOptions)
	case "role":
		handleAdminRole(s, i, subOptions)
	case "member":
		handleAdminMember(s, i, subOptions)
	case "voice":
		handleAdminVoice(s, i, subOptions)
	case "purge":
		handleAdminPurge(s, i, subOptions)
	case "server":
		handleAdminServer(s, i)
	default:
		respondText(s, i, "Unknown subcommand: "+subcommand)
	}
}

func handleAdminChannel(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) == 0 {
		respondText(s, i, "Please provide a channel action: create, delete, lock")
		return
	}

	action := options[0].Name
	actionOptions := options[0].Options
	ctx := context.Background()

	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	switch action {
	case "create":
		name := getOptionString(actionOptions, "name")
		if name == "" {
			respondText(s, i, "Channel name is required")
			return
		}
		channelType := getOptionString(actionOptions, "type")
		category := getOptionString(actionOptions, "category")

		ch, err := client.CreateChannel(ctx, clients.ChannelCreateOptions{
			Name:     name,
			Type:     channelTypeFromString(channelType),
			ParentID: category,
		})
		if err != nil {
			respondText(s, i, "Failed to create channel: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":white_check_mark: Created channel **%s** (ID: `%s`)", ch.Name, ch.ID))

	case "delete":
		channelID := getOptionString(actionOptions, "channel")
		if channelID == "" {
			respondText(s, i, "Channel ID is required")
			return
		}
		if err := client.DeleteChannel(ctx, channelID); err != nil {
			respondText(s, i, "Failed to delete channel: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":white_check_mark: Deleted channel `%s`", channelID))

	case "lock":
		channelID := getOptionString(actionOptions, "channel")
		locked := getOptionBool(actionOptions, "locked")
		if channelID == "" {
			respondText(s, i, "Channel ID is required")
			return
		}
		if locked {
			if err := client.LockChannel(ctx, channelID); err != nil {
				respondText(s, i, "Failed to lock channel: "+err.Error())
				return
			}
			respondText(s, i, fmt.Sprintf(":lock: Locked channel `%s`", channelID))
		} else {
			if err := client.UnlockChannel(ctx, channelID); err != nil {
				respondText(s, i, "Failed to unlock channel: "+err.Error())
				return
			}
			respondText(s, i, fmt.Sprintf(":unlock: Unlocked channel `%s`", channelID))
		}
	}
}

func handleAdminRole(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) == 0 {
		respondText(s, i, "Please provide a role action: create, assign, revoke")
		return
	}

	action := options[0].Name
	actionOptions := options[0].Options
	ctx := context.Background()

	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	switch action {
	case "create":
		name := getOptionString(actionOptions, "name")
		if name == "" {
			respondText(s, i, "Role name is required")
			return
		}
		colorStr := getOptionString(actionOptions, "color")
		color := parseColorHex(colorStr)

		role, err := client.CreateRole(ctx, clients.RoleCreateOptions{
			Name:  name,
			Color: color,
		})
		if err != nil {
			respondText(s, i, "Failed to create role: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":white_check_mark: Created role **%s** (ID: `%s`)", role.Name, role.ID))

	case "assign":
		userID := getOptionString(actionOptions, "user")
		roleID := getOptionString(actionOptions, "role")
		if userID == "" || roleID == "" {
			respondText(s, i, "User and role are required")
			return
		}
		if err := client.AssignRole(ctx, userID, roleID); err != nil {
			respondText(s, i, "Failed to assign role: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":white_check_mark: Assigned role `%s` to user `%s`", roleID, userID))

	case "revoke":
		userID := getOptionString(actionOptions, "user")
		roleID := getOptionString(actionOptions, "role")
		if userID == "" || roleID == "" {
			respondText(s, i, "User and role are required")
			return
		}
		if err := client.RevokeRole(ctx, userID, roleID); err != nil {
			respondText(s, i, "Failed to revoke role: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":white_check_mark: Revoked role `%s` from user `%s`", roleID, userID))
	}
}

func handleAdminMember(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) == 0 {
		respondText(s, i, "Please provide a member action: kick, ban, timeout, info")
		return
	}

	action := options[0].Name
	actionOptions := options[0].Options
	ctx := context.Background()

	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	switch action {
	case "kick":
		userID := getOptionString(actionOptions, "user")
		reason := getOptionString(actionOptions, "reason")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		if err := client.KickMember(ctx, userID, reason); err != nil {
			respondText(s, i, "Failed to kick member: "+err.Error())
			return
		}
		msg := fmt.Sprintf(":boot: Kicked user `%s`", userID)
		if reason != "" {
			msg += fmt.Sprintf(" - Reason: %s", reason)
		}
		respondText(s, i, msg)

	case "ban":
		userID := getOptionString(actionOptions, "user")
		reason := getOptionString(actionOptions, "reason")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		if err := client.BanMember(ctx, userID, reason, 0); err != nil {
			respondText(s, i, "Failed to ban member: "+err.Error())
			return
		}
		msg := fmt.Sprintf(":hammer: Banned user `%s`", userID)
		if reason != "" {
			msg += fmt.Sprintf(" - Reason: %s", reason)
		}
		respondText(s, i, msg)

	case "timeout":
		userID := getOptionString(actionOptions, "user")
		minutes := getOptionInt(actionOptions, "minutes")
		reason := getOptionString(actionOptions, "reason")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		if minutes <= 0 {
			minutes = 5 // Default 5 minutes
		}
		duration := time.Duration(minutes) * time.Minute
		if err := client.TimeoutMember(ctx, userID, duration, reason); err != nil {
			respondText(s, i, "Failed to timeout member: "+err.Error())
			return
		}
		msg := fmt.Sprintf(":mute: Timed out user `%s` for %d minutes", userID, minutes)
		if reason != "" {
			msg += fmt.Sprintf(" - Reason: %s", reason)
		}
		respondText(s, i, msg)

	case "info":
		userID := getOptionString(actionOptions, "user")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		member, err := client.GetMemberInfo(ctx, userID)
		if err != nil {
			respondText(s, i, "Failed to get member info: "+err.Error())
			return
		}
		embed := &discordgo.MessageEmbed{
			Title: ":bust_in_silhouette: Member Info",
			Color: 0x5865F2,
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Username", Value: member.Username, Inline: true},
				{Name: "Display Name", Value: member.DisplayName, Inline: true},
				{Name: "User ID", Value: member.UserID, Inline: true},
				{Name: "Roles", Value: fmt.Sprintf("%d roles", len(member.Roles)), Inline: true},
				{Name: "Joined", Value: member.JoinedAt.Format("Jan 2, 2006"), Inline: true},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "aftrs-bot | Admin",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		respondEmbed(s, i, embed)
	}
}

func handleAdminVoice(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	if len(options) == 0 {
		respondText(s, i, "Please provide a voice action: status, move, disconnect, mute")
		return
	}

	action := options[0].Name
	actionOptions := options[0].Options
	ctx := context.Background()

	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	switch action {
	case "status":
		status, err := client.GetVoiceStatus(ctx)
		if err != nil {
			respondText(s, i, "Failed to get voice status: "+err.Error())
			return
		}
		var sb strings.Builder
		sb.WriteString("**Voice Channel Status**\n\n")
		for _, ch := range status {
			sb.WriteString(fmt.Sprintf("**%s**: %d members\n", ch.ChannelName, ch.MemberCount))
		}
		if len(status) == 0 {
			sb.WriteString("No voice channels with members")
		}
		respondText(s, i, sb.String())

	case "move":
		userID := getOptionString(actionOptions, "user")
		channelID := getOptionString(actionOptions, "channel")
		if userID == "" || channelID == "" {
			respondText(s, i, "User and channel are required")
			return
		}
		if err := client.VoiceMove(ctx, userID, channelID); err != nil {
			respondText(s, i, "Failed to move user: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":arrow_right: Moved user `%s` to channel `%s`", userID, channelID))

	case "disconnect":
		userID := getOptionString(actionOptions, "user")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		if err := client.VoiceDisconnect(ctx, userID); err != nil {
			respondText(s, i, "Failed to disconnect user: "+err.Error())
			return
		}
		respondText(s, i, fmt.Sprintf(":x: Disconnected user `%s` from voice", userID))

	case "mute":
		userID := getOptionString(actionOptions, "user")
		muted := getOptionBool(actionOptions, "muted")
		if userID == "" {
			respondText(s, i, "User ID is required")
			return
		}
		if err := client.VoiceMute(ctx, userID, muted); err != nil {
			respondText(s, i, "Failed to mute user: "+err.Error())
			return
		}
		if muted {
			respondText(s, i, fmt.Sprintf(":mute: Server muted user `%s`", userID))
		} else {
			respondText(s, i, fmt.Sprintf(":loud_sound: Server unmuted user `%s`", userID))
		}
	}
}

func handleAdminPurge(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	count := getOptionInt(options, "count")
	if count < 2 || count > 100 {
		respondText(s, i, "Count must be between 2 and 100")
		return
	}

	ctx := context.Background()
	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	deleted, err := client.PurgeMessages(ctx, i.ChannelID, count)
	if err != nil {
		respondText(s, i, "Failed to purge messages: "+err.Error())
		return
	}

	respondText(s, i, fmt.Sprintf(":wastebasket: Deleted %d messages", deleted))
}

func handleAdminServer(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	client, err := clients.NewDiscordClient()
	if err != nil {
		respondText(s, i, "Could not connect to Discord: "+err.Error())
		return
	}
	defer client.Close()

	info, err := client.GetServerInfo(ctx)
	if err != nil {
		respondText(s, i, "Failed to get server info: "+err.Error())
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":gear: Server Info",
		Description: info.Name,
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Members", Value: fmt.Sprintf("%d", info.MemberCount), Inline: true},
			{Name: "Channels", Value: fmt.Sprintf("%d", info.ChannelCount), Inline: true},
			{Name: "Roles", Value: fmt.Sprintf("%d", info.RoleCount), Inline: true},
			{Name: "Emojis", Value: fmt.Sprintf("%d", info.EmojiCount), Inline: true},
			{Name: "Boost Level", Value: fmt.Sprintf("%d", info.BoostLevel), Inline: true},
			{Name: "Boosts", Value: fmt.Sprintf("%d", info.BoostCount), Inline: true},
			{Name: "Created", Value: info.CreatedAt.Format("Jan 2, 2006"), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Admin",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	respondEmbed(s, i, embed)
}

// Helper functions for option parsing
func getOptionString(options []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, opt := range options {
		if opt.Name == name {
			return opt.StringValue()
		}
	}
	return ""
}

func getOptionInt(options []*discordgo.ApplicationCommandInteractionDataOption, name string) int {
	for _, opt := range options {
		if opt.Name == name {
			return int(opt.IntValue())
		}
	}
	return 0
}

func getOptionBool(options []*discordgo.ApplicationCommandInteractionDataOption, name string) bool {
	for _, opt := range options {
		if opt.Name == name {
			return opt.BoolValue()
		}
	}
	return false
}

func channelTypeFromString(t string) discordgo.ChannelType {
	switch t {
	case "voice":
		return discordgo.ChannelTypeGuildVoice
	case "news":
		return discordgo.ChannelTypeGuildNews
	case "forum":
		return discordgo.ChannelTypeGuildForum
	default:
		return discordgo.ChannelTypeGuildText
	}
}

func parseColorHex(colorStr string) int {
	if colorStr == "" {
		return 0x5865F2
	}
	var color int
	if _, err := fmt.Sscanf(colorStr, "#%x", &color); err == nil {
		return color
	}
	if _, err := fmt.Sscanf(colorStr, "0x%x", &color); err == nil {
		return color
	}
	return 0x5865F2
}
