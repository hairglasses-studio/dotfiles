# Building Discord Bots

Create bots that respond to commands, send notifications, and integrate with AI.

## Prerequisites

- Discord account
- Go installed (`brew install go`)
- A Discord server you can test in

---

## Part 1: Discord Setup

### Create Application

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications)
2. Click "New Application"
3. Name it (e.g., "Luke's Bot")
4. Click "Create"

### Create Bot User

1. Go to "Bot" in left sidebar
2. Click "Add Bot"
3. Click "Yes, do it!"

### Get Bot Token

1. Under "Token", click "Reset Token"
2. Copy the token (keep it secret!)

### Enable Intents

1. Scroll to "Privileged Gateway Intents"
2. Enable:
   - Message Content Intent
   - Server Members Intent (if needed)
3. Save changes

### Invite Bot to Server

1. Go to "OAuth2" → "URL Generator"
2. Check scopes:
   - `bot`
   - `applications.commands`
3. Check permissions:
   - Send Messages
   - Read Message History
   - Use Slash Commands
4. Copy the generated URL
5. Open in browser and add to your server

---

## Part 2: Basic Bot Code

### Project Setup

```bash
mkdir my-discord-bot
cd my-discord-bot
go mod init my-discord-bot
go get github.com/bwmarrin/discordgo
```

### Simple Bot

Create `main.go`:

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

func main() {
    // Get token from environment
    token := os.Getenv("DISCORD_BOT_TOKEN")
    if token == "" {
        fmt.Println("DISCORD_BOT_TOKEN not set")
        return
    }

    // Create Discord session
    dg, err := discordgo.New("Bot " + token)
    if err != nil {
        fmt.Println("Error creating session:", err)
        return
    }

    // Register message handler
    dg.AddHandler(messageCreate)

    // Open connection
    err = dg.Open()
    if err != nil {
        fmt.Println("Error opening connection:", err)
        return
    }

    fmt.Println("Bot is running. Press Ctrl+C to exit.")

    // Wait for interrupt
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
    <-sc

    dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore own messages
    if m.Author.ID == s.State.User.ID {
        return
    }

    // Respond to !ping
    if m.Content == "!ping" {
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    }

    // Respond to !hello
    if m.Content == "!hello" {
        s.ChannelMessageSend(m.ChannelID, "Hello, "+m.Author.Username+"!")
    }
}
```

### Run the Bot

```bash
# Set token
export DISCORD_BOT_TOKEN="your-token-here"

# Run
go run main.go
```

Test in Discord: type `!ping` or `!hello`

---

## Part 3: Slash Commands

Modern Discord bots use slash commands:

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
    {
        Name:        "ping",
        Description: "Check if bot is alive",
    },
    {
        Name:        "greet",
        Description: "Get a greeting",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "name",
                Description: "Who to greet",
                Required:    true,
            },
        },
    },
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "ping": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Pong!",
            },
        })
    },
    "greet": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        name := i.ApplicationCommandData().Options[0].StringValue()
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Hello, %s!", name),
            },
        })
    },
}

func main() {
    token := os.Getenv("DISCORD_BOT_TOKEN")
    dg, _ := discordgo.New("Bot " + token)

    // Handle slash commands
    dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    })

    dg.Open()

    // Register commands
    for _, cmd := range commands {
        dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
    }

    fmt.Println("Bot is running with slash commands")

    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
    <-sc

    dg.Close()
}
```

---

## Part 4: Rich Embeds

Send formatted messages:

```go
func sendEmbed(s *discordgo.Session, channelID string) {
    embed := &discordgo.MessageEmbed{
        Title:       "Session Started",
        Description: "A new studio session has begun",
        Color:       0x00ff00,  // Green
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Time",
                Value:  "2024-12-24 10:00",
                Inline: true,
            },
            {
                Name:   "Type",
                Value:  "Live Stream",
                Inline: true,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: "AFTRS Studio",
        },
    }

    s.ChannelMessageSendEmbed(channelID, embed)
}
```

---

## Part 5: AI Integration

Connect to Claude API:

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
)

func askClaude(prompt string) (string, error) {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")

    payload := map[string]interface{}{
        "model": "claude-3-haiku-20240307",
        "max_tokens": 500,
        "messages": []map[string]string{
            {"role": "user", "content": prompt},
        },
    }

    body, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST",
        "https://api.anthropic.com/v1/messages",
        bytes.NewBuffer(body))

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-api-key", apiKey)
    req.Header.Set("anthropic-version", "2023-06-01")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    // Extract response text
    content := result["content"].([]interface{})[0].(map[string]interface{})
    return content["text"].(string), nil
}

// In your command handler:
commandHandlers["ask"] = func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    question := i.ApplicationCommandData().Options[0].StringValue()

    // Defer response (AI might take a moment)
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
    })

    answer, err := askClaude(question)
    if err != nil {
        answer = "Sorry, I couldn't get a response."
    }

    // Send follow-up
    s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
        Content: &answer,
    })
}
```

---

## Part 6: Running as a Service

### macOS (launchd)

Create `~/Library/LaunchAgents/com.luke.discordbot.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.luke.discordbot</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/my-discord-bot</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>DISCORD_BOT_TOKEN</key>
        <string>your-token-here</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/discordbot.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/discordbot.err</string>
</dict>
</plist>
```

Load it:

```bash
launchctl load ~/Library/LaunchAgents/com.luke.discordbot.plist
```

### Linux (systemd)

Create `/etc/systemd/system/discordbot.service`:

```ini
[Unit]
Description=Discord Bot
After=network.target

[Service]
Type=simple
User=luke
Environment=DISCORD_BOT_TOKEN=your-token
ExecStart=/path/to/my-discord-bot
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable discordbot
sudo systemctl start discordbot
```

---

## Quick Reference

| Task | Code |
|------|------|
| Send message | `s.ChannelMessageSend(channelID, "text")` |
| Send embed | `s.ChannelMessageSendEmbed(channelID, embed)` |
| Reply to interaction | `s.InteractionRespond(i.Interaction, response)` |
| Get user | `m.Author.Username` |
| Get channel | `m.ChannelID` |
| Add reaction | `s.MessageReactionAdd(channelID, messageID, "👍")` |

---

## Next Steps

- [TouchDesigner Development](09-touchdesigner-plugins.md) - Visual programming
- [Resolume & FFGL](10-resolume-ffgl.md) - VJ plugins
