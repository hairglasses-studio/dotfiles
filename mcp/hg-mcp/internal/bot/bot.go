// Package bot provides the Discord bot service for aftrs.
package bot

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot service
type Bot struct {
	session       *discordgo.Session
	guildID       string
	commands      []*discordgo.ApplicationCommand
	handlers      map[string]CommandHandler
	sessionLogger *SessionLogger
	notifier      *Notifier
}

// CommandHandler is a function that handles a slash command interaction
type CommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate)

// New creates a new Bot instance
func New() (*Bot, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN environment variable not set")
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		return nil, fmt.Errorf("DISCORD_GUILD_ID environment variable not set")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	// Set intents - includes privileged intents for admin operations
	// NOTE: Privileged intents (GuildMembers, MessageContent) must be enabled in Discord Developer Portal
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMembers | // PRIVILEGED - for member tracking
		discordgo.IntentsGuildBans | // for ban/kick events
		discordgo.IntentsGuildVoiceStates | // for voice channel tracking
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMessageReactions |
		discordgo.IntentsGuildScheduledEvents | // for scheduled event tracking
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent // PRIVILEGED - for message parsing

	bot := &Bot{
		session:  session,
		guildID:  guildID,
		handlers: make(map[string]CommandHandler),
	}

	// Create notifier
	bot.notifier = NewNotifier(session)

	// Create session logger
	bot.sessionLogger = NewSessionLogger(session, bot.notifier)
	log.Println("Session logger initialized")

	// Set global handlers for command use
	SetGlobalHandlers(bot.sessionLogger)

	// Register event handlers
	session.AddHandler(bot.ready)
	session.AddHandler(bot.interactionCreate)
	session.AddHandler(bot.messageCreate)

	return bot, nil
}

// Start connects the bot to Discord and registers commands
func (b *Bot) Start() error {
	// Open connection
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("error opening Discord connection: %w", err)
	}

	// Define slash commands
	b.commands = getCommands()

	// Register command handlers
	b.handlers = getCommandHandlers()

	// Register commands with Discord
	log.Printf("Registering %d slash commands...", len(b.commands))
	for _, cmd := range b.commands {
		_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
		if err != nil {
			log.Printf("Error registering command %s: %v", cmd.Name, err)
		} else {
			log.Printf("Registered command: /%s", cmd.Name)
		}
	}

	return nil
}

// Stop gracefully shuts down the bot
func (b *Bot) Stop() {
	// Remove commands (optional - comment out to keep commands registered)
	log.Println("Removing slash commands...")
	registeredCommands, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err == nil {
		for _, cmd := range registeredCommands {
			if err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID); err != nil {
				log.Printf("Error removing command %s: %v", cmd.Name, err)
			}
		}
	}

	// Close session
	if err := b.session.Close(); err != nil {
		log.Printf("Error closing Discord session: %v", err)
	}
}

// ready is called when the bot is ready
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Bot is ready! Logged in as %s#%s", event.User.Username, event.User.Discriminator)

	// Set status
	if err := s.UpdateGameStatus(0, "The Aftrs Studio"); err != nil {
		log.Printf("Error setting status: %v", err)
	}
}

// interactionCreate handles slash command interactions
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandName := i.ApplicationCommandData().Name
	if handler, ok := b.handlers[commandName]; ok {
		handler(s, i)
	} else {
		log.Printf("Unknown command: %s", commandName)
	}
}

// messageCreate handles incoming messages (for mentions and keywords)
func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the bot was mentioned or if it's a DM
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

	response := fmt.Sprintf("Hey %s! I'm the Aftrs Studio bot. Use `/help` to see available commands.", m.Author.Username)
	s.ChannelMessageSend(m.ChannelID, response)
}

// Session returns the underlying Discord session (for external use)
func (b *Bot) Session() *discordgo.Session {
	return b.session
}

// GuildID returns the configured guild ID
func (b *Bot) GuildID() string {
	return b.guildID
}

// Notifier returns the bot's notifier
func (b *Bot) Notifier() *Notifier {
	return b.notifier
}

// SessionLogger returns the bot's session logger
func (b *Bot) SessionLogger() *SessionLogger {
	return b.sessionLogger
}
