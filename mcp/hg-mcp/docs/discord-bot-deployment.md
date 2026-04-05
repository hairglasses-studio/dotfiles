# Discord Bot Deployment Guide

## Overview

The Aftrs Discord bot (`aftrs-bot`) is a standalone service that provides:
- Slash commands for studio status monitoring
- AI-powered chat via Claude API
- Session logging with vault integration
- Studio notifications (streams, errors, health warnings)

## Prerequisites

1. Discord Bot Token from [Discord Developer Portal](https://discord.com/developers/applications)
2. Anthropic API Key (optional, for AI features)
3. Go 1.21+ for building

## Quick Start

```bash
# Build the bot
go build -o aftrs-bot ./cmd/aftrs-bot

# Set required environment variables
export DISCORD_BOT_TOKEN="your-bot-token"
export DISCORD_GUILD_ID="your-server-id"
export DISCORD_CHANNEL_ID="default-channel-id"

# Optional: Enable AI features
export ANTHROPIC_API_KEY="your-anthropic-key"

# Run the bot
./aftrs-bot
```

## Environment Variables

### Required

| Variable | Description |
|----------|-------------|
| `DISCORD_BOT_TOKEN` | Bot token from Discord Developer Portal |
| `DISCORD_GUILD_ID` | Your Discord server ID |
| `DISCORD_CHANNEL_ID` | Default channel for messages |

### Optional - Notification Channels

| Variable | Description |
|----------|-------------|
| `DISCORD_STREAM_ALERTS_CHANNEL` | Channel for stream notifications |
| `DISCORD_TECH_ALERTS_CHANNEL` | Channel for technical alerts |
| `DISCORD_SESSION_LOG_CHANNEL` | Channel for session logs |
| `DISCORD_WEBHOOK_URL` | Webhook URL for external notifications |

### Optional - Features

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | Enables AI chat features |
| `AFTRS_VAULT_PATH` | Path to vault (default: ~/aftrs-vault) |

## Discord Bot Setup

### 1. Create Discord Application

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Name it "Aftrs Bot" (or your preference)
4. Go to "Bot" section
5. Click "Add Bot"
6. Copy the token (this is your `DISCORD_BOT_TOKEN`)

### 2. Configure Bot Permissions

In the Bot section, enable these Privileged Gateway Intents:
- **Message Content Intent** - Required for @mentions and DMs
- **Server Members Intent** - Optional, for welcome messages

### 3. Generate Invite URL

1. Go to OAuth2 > URL Generator
2. Select scopes: `bot`, `applications.commands`
3. Select permissions:
   - Send Messages
   - Send Messages in Threads
   - Embed Links
   - Add Reactions
   - Read Message History
   - Use Slash Commands
4. Copy and open the generated URL to invite the bot

### 4. Get Server and Channel IDs

1. Enable Developer Mode in Discord (Settings > Advanced)
2. Right-click your server name > Copy Server ID (this is `DISCORD_GUILD_ID`)
3. Right-click a channel > Copy Channel ID (this is `DISCORD_CHANNEL_ID`)

## Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/studio-status` | Get comprehensive studio system status |
| `/stream-status` | Check current streaming status |
| `/td-health` | Check TouchDesigner project health |
| `/lighting` | Get lighting system status |
| `/ai [question]` | Ask the AI assistant a question |
| `/session start [name]` | Start a new studio session |
| `/session end` | End the current session |
| `/session status` | Check current session status |
| `/session log [event]` | Log an event to the current session |
| `/ping` | Check bot latency |

## AI Chat Features

When `ANTHROPIC_API_KEY` is set:
- @mention the bot to chat: `@Aftrs Bot what's the studio status?`
- DM the bot directly for private conversations
- AI includes studio context (TouchDesigner, OBS status)
- Rate limited to 10 requests per user per minute

## Session Logging

Sessions are automatically saved to the vault:
- Location: `~/aftrs-vault/sessions/YYYY-MM-DD/`
- Format: Markdown with timeline and highlights
- Discord notifications posted to `DISCORD_SESSION_LOG_CHANNEL`

## Running as a Service

### systemd (Linux)

Create `/etc/systemd/system/aftrs-bot.service`:

```ini
[Unit]
Description=Aftrs Discord Bot
After=network.target

[Service]
Type=simple
User=your-user
WorkingDirectory=/path/to/hg-mcp
ExecStart=/path/to/hg-mcp/aftrs-bot
Restart=always
RestartSec=10
EnvironmentFile=/path/to/hg-mcp/.env

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable aftrs-bot
sudo systemctl start aftrs-bot
```

### launchd (macOS)

Create `~/Library/LaunchAgents/com.aftrs.bot.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.aftrs.bot</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/aftrs-bot</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>DISCORD_BOT_TOKEN</key>
        <string>your-token</string>
        <key>DISCORD_GUILD_ID</key>
        <string>your-server-id</string>
        <key>DISCORD_CHANNEL_ID</key>
        <string>your-channel-id</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Then:
```bash
launchctl load ~/Library/LaunchAgents/com.aftrs.bot.plist
```

## Troubleshooting

### Bot not responding to commands
- Ensure `DISCORD_GUILD_ID` is correct
- Check that the bot has been invited with `applications.commands` scope
- Wait a few minutes for slash commands to register globally

### AI not working
- Verify `ANTHROPIC_API_KEY` is set correctly
- Check bot logs for API errors
- Ensure Message Content Intent is enabled

### Session logging not saving
- Check that `~/aftrs-vault/sessions/` directory exists or can be created
- Verify write permissions

## Architecture

```
cmd/aftrs-bot/main.go     # Entry point, signal handling
internal/bot/
├── bot.go                # Core bot, event handlers
├── commands.go           # Slash command handlers
├── events.go             # Message/reaction handlers
├── ai.go                 # Claude API integration
├── logger.go             # Session logging
└── notifications.go      # Studio notifications
```

## Related MCP Tools

The MCP server includes 12 Discord tools:
- `aftrs_discord_send` - Send a message
- `aftrs_discord_notify` - Send formatted notification
- `aftrs_discord_embed` - Send rich embed
- `aftrs_discord_webhook` - Send via webhook
- `aftrs_discord_react` - Add reaction
- `aftrs_discord_edit` - Edit message
- `aftrs_discord_delete` - Delete message
- `aftrs_discord_thread_create` - Create thread
- `aftrs_discord_studio_event` - Studio event notification
- And more...

## Version History

- **v0.1.0** - Initial release with all 5 phases complete
  - Standalone bot with 8 slash commands
  - AI chat with Claude integration
  - Session logging with vault integration
  - Studio notifications system
