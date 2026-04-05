package bot

import (
	"fmt"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

// NotificationChannels holds channel IDs for different notification types
type NotificationChannels struct {
	StreamAlerts string
	TechAlerts   string
	SessionLog   string
	General      string
}

// Notifier provides methods for sending studio notifications
type Notifier struct {
	session  *discordgo.Session
	channels NotificationChannels
}

// NewNotifier creates a new Notifier
func NewNotifier(session *discordgo.Session) *Notifier {
	return &Notifier{
		session: session,
		channels: NotificationChannels{
			StreamAlerts: os.Getenv("DISCORD_STREAM_ALERTS_CHANNEL"),
			TechAlerts:   os.Getenv("DISCORD_TECH_ALERTS_CHANNEL"),
			SessionLog:   os.Getenv("DISCORD_SESSION_LOG_CHANNEL"),
			General:      os.Getenv("DISCORD_CHANNEL_ID"),
		},
	}
}

// NotifyStreamStart sends a stream start notification
func (n *Notifier) NotifyStreamStart(platform, title string) error {
	channelID := n.channels.StreamAlerts
	if channelID == "" {
		channelID = n.channels.General
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":red_circle: Stream Started",
		Description: fmt.Sprintf("**%s** is now live!", title),
		Color:       0xFF0000,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Platform",
				Value:  platform,
				Inline: true,
			},
			{
				Name:   "Started",
				Value:  time.Now().Format("3:04 PM"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Stream Alerts",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifyStreamStop sends a stream stop notification
func (n *Notifier) NotifyStreamStop(duration string) error {
	channelID := n.channels.StreamAlerts
	if channelID == "" {
		channelID = n.channels.General
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":black_circle: Stream Ended",
		Description: fmt.Sprintf("Stream ended after %s", duration),
		Color:       0x808080,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Stream Alerts",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifyTDError sends a TouchDesigner error notification
func (n *Notifier) NotifyTDError(errorMsg string, operatorPath string) error {
	channelID := n.channels.TechAlerts
	if channelID == "" {
		channelID = n.channels.General
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":x: TouchDesigner Error",
		Description: errorMsg,
		Color:       0xFF0000,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Tech Alerts",
		},
	}

	if operatorPath != "" {
		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:  "Operator",
				Value: operatorPath,
			},
		}
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifyHealthWarning sends a system health warning
func (n *Notifier) NotifyHealthWarning(system string, score int, issues []string) error {
	channelID := n.channels.TechAlerts
	if channelID == "" {
		channelID = n.channels.General
	}

	issueText := "None"
	if len(issues) > 0 {
		issueText = ""
		for _, issue := range issues {
			issueText += "• " + issue + "\n"
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf(":yellow_circle: %s Health Warning", system),
		Description: fmt.Sprintf("Health score dropped to **%d/100**", score),
		Color:       0xFFAA00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Issues",
				Value: issueText,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Tech Alerts",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifySessionStart sends a session start notification
func (n *Notifier) NotifySessionStart(sessionName, operator string) error {
	channelID := n.channels.SessionLog
	if channelID == "" {
		channelID = n.channels.General
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":green_circle: Session Started",
		Description: fmt.Sprintf("**%s** started a new session", operator),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Session",
				Value:  sessionName,
				Inline: true,
			},
			{
				Name:   "Started",
				Value:  time.Now().Format("Mon Jan 2, 3:04 PM"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Session Log",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifySessionEnd sends a session end notification with summary
func (n *Notifier) NotifySessionEnd(sessionName, operator, duration string, highlights []string) error {
	channelID := n.channels.SessionLog
	if channelID == "" {
		channelID = n.channels.General
	}

	highlightText := "None recorded"
	if len(highlights) > 0 {
		highlightText = ""
		for _, h := range highlights {
			highlightText += "• " + h + "\n"
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       ":blue_circle: Session Ended",
		Description: fmt.Sprintf("**%s** ended their session", operator),
		Color:       0x0099FF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Session",
				Value:  sessionName,
				Inline: true,
			},
			{
				Name:   "Duration",
				Value:  duration,
				Inline: true,
			},
			{
				Name:  "Highlights",
				Value: highlightText,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot | Session Log",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}

// NotifyCustom sends a custom notification
func (n *Notifier) NotifyCustom(channelID, title, message string, color int) error {
	if channelID == "" {
		channelID = n.channels.General
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: message,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot",
		},
	}

	_, err := n.session.ChannelMessageSendEmbed(channelID, embed)
	return err
}
