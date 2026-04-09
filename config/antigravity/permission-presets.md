# Antigravity Permission Presets

Use these exact entries in the Antigravity permissions UI when you want a narrow local-development preset for `~/hairglasses-studio`.

## Workspace Files

- `read_file(/home/hg/hairglasses-studio)`
- `write_file(/home/hg/hairglasses-studio)`
- `read_file(/home/hg/.gemini)`
- `write_file(/home/hg/.gemini)`
- `read_file(/home/hg/.config/Antigravity)`
- `write_file(/home/hg/.config/Antigravity)`

## Common Commands

- `command(git)`
- `command(go)`
- `command(make)`
- `command(jq)`
- `command(rg)`
- `command(find)`
- `command(node)`
- `command(npm)`
- `command(pnpm)`
- `command(/usr/bin/antigravity)`

## MCP Examples

- `mcp(studio_process/*)`
- `mcp(studio_systemd/*)`
- `mcp(studio_tmux/*)`
- `mcp(GitHub/*)`
- `mcp(Gmail/*)`
- `mcp(Google Calendar/*)`
- `mcp(Google Drive/*)`

Review the active Antigravity MCP server names in `~/.gemini/antigravity/mcp_config.json` before adding narrower server-specific entries.
