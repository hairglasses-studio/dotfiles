#!/usr/bin/env bash
# claude-post-tool-reload.sh — PostToolUse hook for Claude Code
# Auto-reloads services when Claude edits config files.
# Called by Claude Code after Write/Edit tools via .claude/settings.json hook.
# Reads JSON from stdin, extracts file_path, maps to component, reloads.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/config.sh" 2>/dev/null || exit 0

# Read tool input from stdin (Claude pipes JSON)
input="$(cat)"

# Extract file_path from JSON (lightweight — no jq dependency)
file_path="$(echo "$input" | grep -oP '"file_path"\s*:\s*"[^"]*"' | head -1 | sed 's/.*"file_path"\s*:\s*"//;s/"$//')"
[[ -n "$file_path" ]] || exit 0

# Map file path to component and reload
case "$file_path" in
  */hyprland/* | */hypr/*)  config_reload_service hyprland ;;
  */mako/*)                 config_reload_service mako ;;
  */eww/*)                  config_reload_service eww ;;
  */waybar/*)               config_reload_service waybar ;;
  */sway/*)                 config_reload_service sway ;;
  */tmux/*)                 config_reload_service tmux ;;
  # ghostty, starship, tattoy auto-reload via file watchers — no action needed
esac

exit 0
