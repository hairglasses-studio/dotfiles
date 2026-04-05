package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// RegisterEventHandlers adds additional event handlers to the bot
// This is called from bot.go but can be extended here
func RegisterEventHandlers(s *discordgo.Session) {
	s.AddHandler(onReactionAdd)
	s.AddHandler(onGuildMemberAdd)
}

// onReactionAdd handles reaction events
func onReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	// Ignore reactions from the bot
	if r.UserID == s.State.User.ID {
		return
	}

	// Log reactions for debugging
	log.Printf("Reaction added: %s by user %s on message %s", r.Emoji.Name, r.UserID, r.MessageID)

	// Future: Handle specific emoji reactions for interactive controls
	// Example: React with a checkmark to approve something
}

// onGuildMemberAdd handles new member joins
func onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	log.Printf("New member joined: %s#%s", m.User.Username, m.User.Discriminator)

	// Future: Send welcome message or assign roles
	// For now, just log the event
}

// SendWelcomeMessage sends a welcome message to a channel
func SendWelcomeMessage(s *discordgo.Session, channelID string, user *discordgo.User) error {
	embed := &discordgo.MessageEmbed{
		Title:       "Welcome to The Aftrs!",
		Description: "Hey " + user.Username + "! Welcome to The Aftrs Studio Discord server.",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Getting Started",
				Value: "Use `/help` to see available bot commands.",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "aftrs-bot",
		},
	}

	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	return err
}
