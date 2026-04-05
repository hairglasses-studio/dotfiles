# Understanding MCP Servers

MCP (Model Context Protocol) extends Claude with external tools and capabilities.

## What is MCP?

MCP lets Claude:
- Control external software (TouchDesigner, OBS, Discord)
- Access databases and APIs
- Run custom automation
- Interact with hardware

Think of MCP tools as "superpowers" for Claude.

---

## How It Works

```
┌─────────────┐     JSON-RPC      ┌─────────────┐
│ Claude Code │ ◄──────────────► │ MCP Server  │
└─────────────┘                   └─────────────┘
                                        │
                                        ▼
                                  ┌─────────────┐
                                  │   Tools     │
                                  │ (TD, OBS,   │
                                  │  Discord)   │
                                  └─────────────┘
```

1. You ask Claude to do something
2. Claude identifies which tool to use
3. Claude calls the MCP server
4. The server executes the action
5. Results come back to Claude

---

## Discovering Available Tools

### List All Tools

```
"What MCP tools are available?"
"/tools"
```

### Find Specific Tools

```
"What tools do we have for TouchDesigner?"
"Show me video processing tools"
"Find tools related to Discord"
```

---

## Using MCP Tools

You don't need special syntax. Just describe what you want:

```
"Check the FPS in TouchDesigner"
# Claude uses: aftrs_td_fps tool

"Send a message to Discord"
# Claude uses: aftrs_discord_send tool

"List NDI sources"
# Claude uses: aftrs_ndi_sources tool
```

---

## Configuration

MCP servers are configured in `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "aftrs-mcp": {
      "command": "/path/to/aftrs-mcp",
      "args": [],
      "env": {
        "DISCORD_BOT_TOKEN": "your-token"
      }
    }
  }
}
```

### Adding a New MCP Server

```bash
# Edit the settings file
claude config edit

# Or manually edit
nano ~/.claude/settings.json
```

Add your server:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/my-mcp-server"
    }
  }
}
```

Restart Claude Code to load the new server.

---

## Available MCP Servers

### aftrs-mcp (Studio Automation)

| Category | Tools |
|----------|-------|
| TouchDesigner | FPS monitoring, parameter control, network health |
| Video | AI processing, format conversion |
| Streaming | NDI sources, OBS control |
| Discord | Messaging, notifications |
| Vault | Documentation, session logs |

### webb (Knowledge Graph)

| Category | Tools |
|----------|-------|
| Vault | Search, read, write documents |
| Research | Full research workflows |
| Graph | Knowledge connections |

### playwright (Browser)

| Category | Tools |
|----------|-------|
| Browser | Page navigation, screenshots |
| Forms | Fill and submit |
| Scraping | Extract content |

---

## Example Workflows

### Studio Session Start

```
"Start a new studio session:
1. Check TouchDesigner FPS
2. List available NDI sources
3. Send Discord notification that we're live
4. Create session notes in the vault"
```

### Video Processing

```
"Process the video at ~/Videos/raw.mov:
1. Remove the background
2. Upscale to 4K
3. Save to ~/Videos/processed/"
```

### Research Task

```
"Research MCP server best practices:
1. Search the vault for existing notes
2. Look for related documentation
3. Create a summary document"
```

---

## Troubleshooting

### Server Not Starting

```bash
# Check if the binary exists
ls -la /path/to/mcp-server

# Make sure it's executable
chmod +x /path/to/mcp-server

# Test running it directly
/path/to/mcp-server
```

### Tools Not Showing

```
"Check MCP server status"
"/doctor"
```

Verify in settings:
```bash
cat ~/.claude/settings.json | jq '.mcpServers'
```

### Tool Errors

```
"What error did the last tool call produce?"
"Try the tool again with debug output"
```

---

## Environment Variables

MCP servers often need credentials:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/server",
      "env": {
        "API_KEY": "xxx",
        "DATABASE_URL": "postgres://...",
        "DEBUG": "true"
      }
    }
  }
}
```

Common environment variables:

| Variable | Purpose |
|----------|---------|
| `DISCORD_BOT_TOKEN` | Discord bot authentication |
| `OBSIDIAN_VAULT` | Path to vault |
| `TD_HOST` | TouchDesigner hostname |
| `OBS_PASSWORD` | OBS WebSocket password |

---

## Security Notes

- Keep tokens/secrets in environment variables, not in code
- MCP servers run locally on your machine
- Tools have access to what the server allows
- Review server source code before trusting it

---

## Next Steps

- [Building MCP Modules](06-building-mcp-modules.md) - Create your own tools
- [GitHub Workflows](07-github-workflows.md) - Automate with CI/CD
