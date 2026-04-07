# AFTRS-MCP Quick Reference

Personal quick reference for using the AFTRS MCP server with Claude Code.

## Quick Commands

### Start/Stop Server

```bash
# Start in stdio mode (for Claude Code)
aftrs-start

# Start in SSE mode (for remote access)
aftrs-sse

# Start Discord bot standalone
aftrs-bot

# Use CLI tool directly
aftrs-cli tool discord status

# Rebuild after code changes
aftrs-rebuild

# View logs
aftrs-logs
```

### Environment Variable Checklist

Before using Discord features, ensure these are set in `~/.zshrc`:

- [ ] `DISCORD_BOT_TOKEN` - Your Discord bot token
- [ ] `DISCORD_GUILD_ID` - Your Discord server ID
- [ ] `DISCORD_CHANNEL_ID` - Default notification channel
- [ ] `DISCORD_STREAM_ALERTS_CHANNEL` - Stream alerts channel
- [ ] `DISCORD_TECH_ALERTS_CHANNEL` - Technical alerts channel
- [ ] `DISCORD_SESSION_LOG_CHANNEL` - Session logging channel
- [ ] `AFTRS_VAULT_PATH` - Obsidian vault path (default: ~/aftrs-vault)

Check current values:
```bash
echo $DISCORD_BOT_TOKEN
echo $DISCORD_GUILD_ID
```

Reload after editing `.zshrc`:
```bash
source ~/.zshrc
```

## Common Tool Usage via Claude Code

### TouchDesigner (25 tools)

**Check TD Status:**
```
Ask Claude: "Check TouchDesigner performance status"
Uses: aftrs_td_status
```

**Monitor Network Health:**
```
Ask Claude: "Analyze TouchDesigner network health and show any bottlenecks"
Uses: aftrs_td_network_health
```

**Execute Python Script:**
```
Ask Claude: "Execute this Python code in TouchDesigner: op('/project1').par.opacity = 0.5"
Uses: aftrs_td_execute
```

**Check GPU Memory:**
```
Ask Claude: "Check TouchDesigner GPU memory usage"
Uses: aftrs_td_gpu_memory
```

**Get/Set Parameters:**
```
Ask Claude: "Get the opacity parameter from /project1/null1"
Ask Claude: "Set /project1/null1 opacity to 0.8"
Uses: aftrs_td_parameters
```

**View Errors:**
```
Ask Claude: "Show all TouchDesigner errors and warnings"
Uses: aftrs_td_errors
```

### Resolume (38 tools)

**Check Connection:**
```
Ask Claude: "Check Resolume connection status"
Uses: aftrs_resolume_status
```

**Trigger Clip:**
```
Ask Claude: "Trigger clip 3 on layer 2 in Resolume"
Uses: aftrs_resolume_trigger
```

**Crossfade Decks:**
```
Ask Claude: "Crossfade from deck A to deck B in Resolume"
Uses: aftrs_resolume_crossfade
```

**Control Effects:**
```
Ask Claude: "Toggle effect 1 on layer 1"
Uses: aftrs_resolume_effects
```

**Set BPM:**
```
Ask Claude: "Set Resolume BPM to 128"
Uses: aftrs_resolume_bpm
```

**List Clips:**
```
Ask Claude: "Show all clips in Resolume layer 1"
Uses: aftrs_resolume_clips
```

### Discord (12 tools)

**Send Message:**
```
Ask Claude: "Send a Discord message: Stream starting in 10 minutes!"
Uses: aftrs_discord_send
```

**Send Formatted Notification:**
```
Ask Claude: "Send a Discord notification with severity 'info': Studio systems online"
Uses: aftrs_discord_notify
```

**List Channels:**
```
Ask Claude: "List all Discord channels"
Uses: aftrs_discord_channels
```

**Create Thread:**
```
Ask Claude: "Create a Discord thread named 'Tonight's Show' in the session log channel"
Uses: aftrs_discord_thread
```

**Get Message History:**
```
Ask Claude: "Get the last 10 messages from the tech alerts channel"
Uses: aftrs_discord_history
```

### Studio Operations (8 consolidated tools)

**Quick Health Check:**
```
Ask Claude: "Run a studio health check"
Uses: hairglasses_studio_health
```

**Comprehensive Health Check:**
```
Ask Claude: "Run a full comprehensive studio health check"
Uses: hairglasses_studio_health_full
```

**Morning Checklist:**
```
Ask Claude: "Run the morning pre-show checklist"
Uses: aftrs_morning_check
```

**Pre-Stream Check:**
```
Ask Claude: "Run pre-stream checks before going live"
Uses: aftrs_pre_stream_check
```

**Show Startup:**
```
Ask Claude: "Run show startup sequence"
Uses: aftrs_show_startup
```

**Emergency Panic Mode:**
```
Ask Claude: "Activate panic mode - emergency shutdown"
Uses: aftrs_panic_mode
```

### Vault / Session Logging

**Save Session Note:**
```
Ask Claude: "Save this to the vault: Great show tonight, TD performed well at 60fps"
Uses: aftrs_vault_save
```

**Search Vault:**
```
Ask Claude: "Search vault for notes about 'performance issues'"
Uses: aftrs_vault_search
```

**Create Session Summary:**
```
Ask Claude: "Create a session summary for today"
Uses: aftrs_session_summary (via /session command)
```

### OBS Streaming (15 tools)

**Check OBS Status:**
```
Ask Claude: "Check OBS connection status"
Uses: aftrs_obs_status
```

**Start/Stop Recording:**
```
Ask Claude: "Start OBS recording"
Ask Claude: "Stop OBS recording"
Uses: aftrs_obs_record
```

**Switch Scene:**
```
Ask Claude: "Switch OBS to the 'Main Camera' scene"
Uses: aftrs_obs_scenes
```

**Control Audio:**
```
Ask Claude: "Set OBS microphone volume to 75%"
Ask Claude: "Mute OBS desktop audio"
Uses: aftrs_obs_volume, aftrs_obs_mute
```

## Discord Bot Setup

### Step 1: Create Discord Bot

1. Go to https://discord.com/developers/applications
2. Click "New Application" → Name it "AFTRS Bot"
3. Go to "Bot" tab → Click "Add Bot"
4. Copy the Bot Token (keep secret!)

### Step 2: Enable Intents

In the Bot settings, enable:
- ✅ Presence Intent
- ✅ Server Members Intent
- ✅ Message Content Intent

### Step 3: Invite Bot to Server

Use this URL (replace `YOUR_CLIENT_ID` with your Application ID from "General Information"):

```
https://discord.com/api/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=274877991936&scope=bot
```

### Step 4: Get IDs

1. Enable Developer Mode: Settings → Advanced → Developer Mode
2. Right-click server name → Copy Server ID
3. Right-click channels → Copy Channel ID

### Step 5: Update .zshrc

Edit `~/.zshrc` and replace placeholder values:

```bash
export DISCORD_BOT_TOKEN="paste-your-actual-bot-token-here"
export DISCORD_GUILD_ID="paste-your-server-id"
export DISCORD_CHANNEL_ID="paste-default-channel-id"
export DISCORD_STREAM_ALERTS_CHANNEL="paste-stream-channel-id"
export DISCORD_TECH_ALERTS_CHANNEL="paste-tech-channel-id"
export DISCORD_SESSION_LOG_CHANNEL="paste-session-channel-id"
```

Then reload:
```bash
source ~/.zshrc
```

## TouchDesigner Configuration

### Enable Python Server

1. Open TouchDesigner
2. Go to: Edit → Preferences (or Cmd+,)
3. Navigate to: Network → Python Server
4. Enable: ✅ Enable Server
5. Set Port: 8090
6. Click Apply

The MCP server will now be able to communicate with TD.

## Resolume Configuration

### Enable OSC Output

1. Open Resolume Arena
2. Go to: Preferences (Cmd+,)
3. Navigate to: OSC Output
4. Enable: ✅ OSC Output
5. Set Target Host: 127.0.0.1
6. Set Target Port: 7000
7. Click OK

The MCP server can now send OSC commands to Resolume.

## OBS Configuration

### Enable WebSocket Server

1. Open OBS Studio
2. Go to: Tools → WebSocket Server Settings
3. Enable: ✅ Enable WebSocket Server
4. Set Port: 4455
5. (Optional) Set password for security
6. Click OK

If you set a password, add it to `.zshrc`:
```bash
export OBS_WEBSOCKET_PASSWORD="your-password"
```

## Troubleshooting

### Server Won't Start

**Check Go Installation:**
```bash
go version  # Should show 1.25+
```

**Rebuild Binaries:**
```bash
aftrs-rebuild
```

**Check Permissions:**
```bash
chmod +x ~/hairglasses-studio/hg-mcp/hg-mcp
chmod +x ~/hairglasses-studio/hg-mcp/aftrs
chmod +x ~/hairglasses-studio/hg-mcp/aftrs-bot
```

### Discord Bot Not Connecting

**Verify Token:**
```bash
echo $DISCORD_BOT_TOKEN  # Should show your token
```

**Check Bot Permissions:**
- Bot has proper permissions in Discord
- Bot was invited with correct OAuth scopes
- Guild ID and Channel IDs are correct

**Test Connection:**
```bash
aftrs-cli tool discord status
```

### TouchDesigner Tools Failing

**Is TD Running?**
- TouchDesigner must be open and running

**Check Python Server:**
- Verify enabled in Preferences
- Verify port is 8090
- Check no firewall blocking

**Test Connection:**
Ask Claude: "Check TouchDesigner status"

### Resolume Tools Failing

**Is Resolume Running?**
- Resolume Arena must be open

**Check OSC Settings:**
- OSC Output enabled in preferences
- Target: 127.0.0.1:7000
- Check no firewall blocking

**Test Connection:**
Ask Claude: "Check Resolume connection status"

### Environment Variables Not Persisting

**After editing .zshrc, reload:**
```bash
source ~/.zshrc
```

**Or restart terminal**

**Verify variables are set:**
```bash
env | grep DISCORD
env | grep AFTRS
```

## File Locations Reference

| File | Location | Purpose |
|------|----------|---------|
| MCP Server | `~/hairglasses-studio/hg-mcp/hg-mcp` | Main MCP server binary |
| CLI Tool | `~/hairglasses-studio/hg-mcp/aftrs` | Command-line tool |
| Discord Bot | `~/hairglasses-studio/hg-mcp/aftrs-bot` | Standalone Discord bot |
| Config File | `~/hairglasses-studio/hg-mcp/configs/aftrs.yaml` | Server configuration |
| Claude MCP | `~/.config/claude/mcp_settings.json` | Claude Code integration |
| Environment | `~/.zshrc` | Environment variables |
| Vault | `~/aftrs-vault/` | Obsidian vault for logging |
| Logs | `~/hairglasses-studio/hg-mcp/logs/` | Server logs |

## Available Tool Modules (730+ tools across 78 modules)

### Top Modules by Tool Count

| Module | Tools | Description |
|--------|-------|-------------|
| Resolume | 38 | VJ software, clips, effects, layers, audio |
| Discord Admin | 35 | Server management, moderation, admin |
| TouchDesigner | 25 | Project control, performance, scripting |
| Lighting | 24 | DMX, Art-Net, fixtures, scenes |
| Rekordbox | 22 | DJ library, playlists, track analysis |
| Resolume Plugins | 18 | FFGL plugins, effect management |
| MIDI | 17 | MIDI devices, notes, CC, mappings |
| VideoAI | 16 | Computer vision, NDI analysis |
| LedFX | 16 | LED strip effects, audio reactive |
| PTZ | 15 | PTZ camera control, presets |
| OBS | 15 | Streaming, recording, scenes, audio |
| Unraid | 14 | Server management, VMs, containers |
| Spotify | 14 | Playback, library, playlists |
| Gmail/GDrive | 14 | Google Workspace integration |
| Workflows | 14 | Workflow execution, scheduling |
| Discord | 12 | Messages, notifications, channels |
| Serato/Traktor/Ableton | 10-12 | DJ software control |
| Vault | 11 | Obsidian integration, session logs |
| Discovery | 11 | Tool search and catalog |
| Twitch/YouTube | 8-10 | Live streaming platforms |

### All Modules by Category

**DJ/Music**: serato, rekordbox, traktor, ableton, spotify, beatport, stems, fingerprint, maxforlive
**AV/Video**: resolume, obs, touchdesigner, atem, ndicv, videoai, videorouting, ffgl, vj_clips
**Lighting**: grandma3, wled, ledfx, lighting
**Audio**: dante, midi, whisper
**Streaming**: twitch, youtube_live, streaming
**Studio**: unraid, homeassistant, streamdeck, hwmonitor, tailscale
**AI/ML**: ollama, whisper, videoai, ptztrack
**Knowledge**: vault, graph, learning, memory
**Workflows**: chains, workflows, healing, showcontrol, showkontrol
**Google**: gmail, gdrive, gtasks, calendar
**Communication**: discord, discord_admin, notion, mqtt
**Sync**: bpmsync, triggersync, paramsync, timecodesync

## Usage Tips

1. **Start Simple**: Begin with health checks and status queries
2. **Use Natural Language**: Claude understands context, describe what you want
3. **Combine Tools**: Ask Claude to run multiple checks in sequence
4. **Create Workflows**: Combine common operations into reusable workflows
5. **Log Sessions**: Use vault tools to document your work
6. **Monitor Performance**: Regular health checks catch issues early

## Example Session Workflow

```
# Morning Setup
1. Ask Claude: "Run the morning studio checklist"
2. Ask Claude: "Check TouchDesigner and Resolume status"
3. Ask Claude: "Send Discord notification: Studio systems online"

# During Show
4. Ask Claude: "Monitor TouchDesigner performance"
5. Ask Claude: "Check GPU memory usage"
6. Ask Claude: "Trigger Resolume clip 5 on layer 1"

# End of Show
7. Ask Claude: "Create a session summary for today"
8. Ask Claude: "Save session notes to vault"
9. Ask Claude: "Send Discord notification: Show complete, systems shutting down"
10. Ask Claude: "Run show shutdown sequence"
```

## Getting Help

- **Documentation**: `~/hairglasses-studio/hg-mcp/README.md`
- **Architecture**: `~/hairglasses-studio/hg-mcp/docs/ARCHITECTURE.md`
- **Tool Examples**: `~/hairglasses-studio/hg-mcp/internal/mcp/tools/`
- **Ask Claude**: Claude can help troubleshoot and explain tools

## Next Steps

1. Configure Discord bot (get token, server ID, channel IDs)
2. Test MCP server with Claude Code
3. Configure TouchDesigner Python server
4. Configure Resolume OSC output
5. Explore knowledge graph and learning features
6. Create custom workflows
