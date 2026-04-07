# hg-mcp

Aftrs Studio Operations Platform via MCP Server

A centralized MCP server for managing creative studio operations with **1,190+ tools across 119 modules**, organized into 10 runtime groups, inspired by [cobb](https://github.com/example-corp/cobb). Includes discovery tools so large installations can be navigated without treating the full catalog as the default working surface.

## Team

| Member | GitHub | Role |
|--------|--------|------|
| Mitch | [@hairglasses](https://github.com/hairglasses) | Senior Infrastructure Engineer |
| Luke | [@lukelasleyfilm](https://github.com/lukelasleyfilm) | Technical Collaborator |

---

## Features

- **Discord Integration** - Team coordination, bot notifications
- **AV/Live Performance** - TouchDesigner, Resolume, FFGL, MIDI, DMX/lighting
- **Streaming** - NDI, OBS, video processing
- **Studio Automation** - UNRAID, network, hardware control
- **Retro Gaming** - PS2, RetroArch, emulators, visualizers
- **Self-Improving** - Knowledge graph, pattern learning
- **Journaled Vault** - Obsidian integration for progress tracking

---

## Quick Start (Copy/Paste Ready)

### Prerequisites

You need Go installed on your computer. Here's how to check and install it:

#### macOS/Linux

```bash
# Check if Go is installed (should show version 1.22 or higher)
go version

# If not installed, install with Homebrew (macOS)
brew install go

# Verify installation
go version
```

#### Windows

```powershell
# Check if Go is installed (should show version 1.22 or higher)
go version

# If not installed, install with winget
winget install --id GoLang.Go -e --source winget

# Verify installation (you may need to restart your terminal)
go version
```

### Step 1: Clone the Repository

```bash
# Navigate to your projects folder
cd ~/hairglasses-studio

# Clone this repo (if not already done)
git clone https://github.com/hairglasses-studio/hg-mcp.git

# Enter the project folder
cd hg-mcp
```

### Step 2: Build the Project

#### macOS/Linux

```bash
# Build both the CLI and MCP server
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp
```

#### Windows

```powershell
# Create placeholder for web dist (if needed)
New-Item -ItemType Directory -Path internal\web\dist -Force
"placeholder" | Out-File -FilePath internal\web\dist\index.html

# Build both the CLI and MCP server
go build -o aftrs.exe ./cmd/aftrs
go build -o hg-mcp.exe ./cmd/hg-mcp
```

### Step 3: Run the MCP Server

#### macOS/Linux

```bash
# Run in stdio mode (for Codex, Claude Code, or any stdio MCP client)
./hg-mcp

# OR run in Streamable HTTP mode (MCP 2025 spec, replaces SSE)
MCP_MODE=streamable PORT=8080 ./hg-mcp

# OR run in SSE mode (deprecated April 2026)
MCP_MODE=sse PORT=8080 ./hg-mcp
```

#### Windows

```powershell
# Run in stdio mode (for Codex, Claude Code, or any stdio MCP client)
.\hg-mcp.exe

# OR run in SSE mode (for remote access)
$env:MCP_MODE="sse"; $env:PORT="8080"; .\hg-mcp.exe
```

---

## Codex And Claude Code Integration

`hg-mcp` exposes a very large tool surface. Do not add it as an org-wide shared default. Start with a local-scoped install and use the discovery tools first.

### Recommended: local scope install

Use Codex or Claude Code's current MCP commands instead of editing legacy client config files by hand.

#### macOS/Linux

```bash
codex mcp add hg-mcp -- ./hg-mcp
```

```bash
claude mcp add --scope local hg-mcp -- /path/to/hg-mcp
```

#### Windows

```powershell
codex mcp add hg-mcp -- C:\path\to\hg-mcp.exe
```

```powershell
claude mcp add --scope local hg-mcp -- C:\path\to\hg-mcp.exe
```

Use `--env KEY=value` flags if you need to pass credentials or paths during install. Keep sensitive integrations local unless your whole team needs the exact same setup.

### Discovery-first workflow

- Start with `aftrs_tool_search`, `aftrs_tool_catalog`, and `aftrs_tool_stats`.
- Prefer local scope for personal or exploratory use. Only use project scope when your team has agreed on a curated subset.
- If you build a shared wrapper around `hg-mcp`, expose a narrow allowlist or role-specific profile instead of the raw full catalog.
- When your MCP client supports tool discovery/search, keep discovery tools eager and defer the rest of the catalog.

### Verify the install

After adding the server, restart your MCP client and confirm the server is connected. Because the catalog is large, begin with discovery tools rather than assuming the entire surface should be in active use for every session.

---

## Discord Setup (Priority Integration)

Discord is our first integration for team coordination. Here's how to set it up:

### Step 1: Create a Discord Bot

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application" and give it a name (e.g., "Aftrs Bot")
3. Go to "Bot" in the left sidebar
4. Click "Add Bot"
5. Copy the **Bot Token** (keep this secret!)

### Step 2: Invite Bot to Your Server

Replace `YOUR_CLIENT_ID` with your application's Client ID (found in "General Information"):

```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=274877991936&scope=bot
```

### Step 3: Get Your Server and Channel IDs

1. Open Discord Settings → Advanced → Enable "Developer Mode"
2. Right-click your server name → "Copy Server ID"
3. Right-click a channel → "Copy Channel ID"

### Step 4: Set Environment Variables

```bash
# Add these to your shell profile (~/.zshrc or ~/.bashrc)
export DISCORD_BOT_TOKEN="your-bot-token-here"
export DISCORD_GUILD_ID="your-server-id-here"
export DISCORD_CHANNEL_ID="your-default-channel-id-here"

# Reload your shell
source ~/.zshrc
```

### Step 5: Test the Connection

```bash
# Once the Discord module is built, test with:
./aftrs tool discord status
```

---

## Project Structure

```
hg-mcp/
├── cmd/
│   ├── aftrs/           # CLI tool (run commands from terminal)
│   └── hg-mcp/       # MCP server (connects to Claude)
├── internal/
│   ├── clients/         # Code that talks to external services
│   └── mcp/tools/       # Each feature is a "module" here
├── docs/                # Documentation
├── configs/             # Configuration files
├── Roadmap.md           # Development plan
└── README.md            # This file
```

---

## Runtime Groups

Tools are organized into 10 high-level **runtime groups** for consolidated monitoring and discovery:

```
dj_music          — DJ software, music platforms, sample packs, stem separation
vj_video          — VJ software, video processing, camera control, FFGL plugins
lighting          — DMX, LED panels, smart lights (Nanoleaf, Hue, WLED, grandMA3)
audio_production  — DAWs, audio routing, MIDI, Max for Live
show_control      — Cue management, automation, BPM sync, switching
infrastructure    — Servers, networking, home automation, backups
messaging         — Discord, email, Notion, calendars, task management
inventory         — Hardware liquidation, eBay integration, listings
streaming         — Live streaming, Twitch, YouTube Live
platform          — Tool discovery, dashboards, security, knowledge graph
```

Use `aftrs_tool_discover` with the `runtime_group` parameter to browse tools by group:
```
aftrs_tool_discover(runtime_group="lighting")   # All lighting tools
aftrs_lighting_status                            # Aggregated lighting health
aftrs_dj_status                                  # Aggregated DJ/music health
aftrs_audio_status                               # Aggregated audio health
```

---

## Common Commands

### Building

```bash
# Build everything
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp

# Run tests
go test ./...

# Format code (do this before committing)
go fmt ./...
```

### Git Workflow

```bash
# Check what files changed
git status

# See the actual changes
git diff

# Stage all changes
git add -A

# Commit with a message
git commit -m "feat: add discord notification support"

# Push to GitHub
git push origin main
```

### Pulling Latest Changes

```bash
# Get the latest code from GitHub
git pull origin main

# Rebuild after pulling
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp
```

---

## Configuration

Create a config file at `configs/aftrs.yaml`:

```yaml
# Discord settings
discord:
  bot_token_env: "DISCORD_BOT_TOKEN"  # Uses environment variable
  guild_id_env: "DISCORD_GUILD_ID"
  default_channel_env: "DISCORD_CHANNEL_ID"

# TouchDesigner settings
touchdesigner:
  host: "localhost"
  port: 8090

# Resolume settings (OSC control)
resolume:
  osc_host: "127.0.0.1"
  osc_port: 7000

# UNRAID server
unraid:
  host: "tower.local"
  # API key stored in 1Password
  api_key_1password: "op://Studio/UNRAID/api_key"

# Obsidian vault location
vault:
  path: "~/aftrs-vault"
```

---

## Current Status

**1,190+ tools across 119 modules** organized into 10 runtime groups:

| Runtime Group | Modules | Tools | Description |
|---------------|---------|-------|-------------|
| DJ/Music | rekordbox, serato, traktor, soundcloud, spotify, beatport, samples, stems, cr8, sync | 130+ | Library management, sync, analysis, stem separation |
| VJ/Video | resolume, touchdesigner, obs, ndicv, video, videoai, vj_clips, ffgl, ptz | 150+ | VJ software, video processing, camera control |
| Lighting | grandma3, wled, ledfx, nanoleaf, hue, sacn, qlcplus, companion | 80+ | DMX, LED panels, smart lights, show control |
| Audio Production | ableton, dante, midi, maxforlive, supercollider, ardour | 70+ | DAWs, audio routing, MIDI, plugins |
| Show Control | showkontrol, chains, snapshots, atem, chataigne, streamdeck | 80+ | Cue management, automation, switching |
| Infrastructure | unraid, homeassistant, tailscale, backup, mqtt | 50+ | Servers, home automation, networking |
| Messaging | discord, gmail, notion, calendar, slack, telegram | 80+ | Communication, productivity |
| Inventory | inventory | 50+ | Hardware liquidation, eBay, listings |
| Streaming | obs, twitch, youtube_live | 30+ | Live streaming, chat, broadcast |
| Platform | discovery, dashboard, security, analytics, graph, vault | 80+ | Tool discovery, observability, knowledge |

See [Roadmap.md](Roadmap.md) for the full development plan.

---

## Documentation

| Document | Description |
|----------|-------------|
| [Roadmap.md](Roadmap.md) | Development roadmap with all planned features |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | How the code is organized |
| [docs/research/](docs/research/) | Research notes |

### hw-resale Documentation

The inventory module (51 tools) manages hardware liquidation. Full documentation suite:

| Document | Description |
|----------|-------------|
| [docs/system.md](docs/system.md) | System architecture — Sheets schema, client design, all 39 tools |
| [docs/current_bugs.md](docs/current_bugs.md) | Known bugs, limitations, data quality issues, feature gaps |
| [docs/testing.md](docs/testing.md) | Test architecture, coverage analysis, how to add tests |
| [docs/research.md](docs/research.md) | Platform fees, pricing benchmarks, selling best practices |
| [docs/objectives.md](docs/objectives.md) | Revenue targets ($16,225), KPIs, success criteria |
| [docs/roadmap.md](docs/roadmap.md) | Phase-by-phase sales execution plan |
| [docs/suggestions.md](docs/suggestions.md) | Architecture improvements and enhancement recommendations |
| [docs/INVENTORY-OPS.md](docs/INVENTORY-OPS.md) | Daily workflow quickstart and tool reference |

---

## Troubleshooting

### "command not found: go"

**macOS:**
```bash
brew install go
```

**Windows:**
```powershell
winget install --id GoLang.Go -e --source winget
```

### "permission denied" when running ./hg-mcp (macOS/Linux)

Make it executable:
```bash
chmod +x ./hg-mcp ./aftrs
```

### Build fails with "no matching files found" for web dist (Windows)

Create the placeholder directory:
```powershell
New-Item -ItemType Directory -Path internal\web\dist -Force
"placeholder" | Out-File -FilePath internal\web\dist\index.html
```

Then rebuild:
```powershell
go build -o hg-mcp.exe ./cmd/hg-mcp
```

### Build fails with "cannot find module"

Ensure dependencies are up to date:
```bash
go mod tidy
```

### Discord bot not responding

**macOS/Linux:**
```bash
echo $DISCORD_BOT_TOKEN
echo $DISCORD_GUILD_ID
```

**Windows (PowerShell):**
```powershell
$env:DISCORD_BOT_TOKEN
$env:DISCORD_GUILD_ID
```

### MCP server not showing up in your MCP client

1. Run `codex mcp list` or `claude mcp list` and confirm `hg-mcp` appears
2. Check that the path to `hg-mcp.exe` (Windows) or `hg-mcp` (macOS/Linux) is correct
3. Restart the MCP client completely to verify connection state
4. Check client logs for MCP server errors

---

## Getting Help

- Check the [Roadmap.md](Roadmap.md) for planned features
- Review [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for code structure
- Ask in our Discord server

---

## License

MIT
