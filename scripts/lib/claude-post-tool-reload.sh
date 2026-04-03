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

# Map file path to component and reload (log failures to stderr)
case "$file_path" in
  */hyprland/* | */hypr/*)  config_reload_service hyprland || echo "[hook] hyprland reload failed" >&2 ;;
  */mako/*)                 config_reload_service mako     || echo "[hook] mako reload failed" >&2 ;;
  */eww/*)                  config_reload_service eww      || echo "[hook] eww reload failed" >&2 ;;
  */waybar/*)               config_reload_service waybar   || echo "[hook] waybar reload failed" >&2 ;;
  */sway/*)                 config_reload_service sway     || echo "[hook] sway reload failed" >&2 ;;
  */tmux/*)                 config_reload_service tmux     || echo "[hook] tmux reload failed" >&2 ;;
  # ghostty, starship, tattoy auto-reload via file watchers — no action needed
esac

# --- Snazzy palette enforcement (non-blocking warning) ---
case "$file_path" in
  *.conf | *.toml | *.scss | *.yuck | *.ini)
    if [[ -f "$file_path" ]]; then
      # Snazzy palette + common accent/base colors
      snazzy_palette="000000 1a1a1a 1a1b26 ff5c57 5af78e f3f99d 57c7ff ff6ac1 9aedfe f1f1f0 686868 eff0eb d4d4d4 264f78 ffffff f8f8f2 282a36"
      # Extract all 6-digit hex color codes from the file
      hex_colors="$(grep -oiE '#[0-9a-fA-F]{6}' "$file_path" 2>/dev/null | sed 's/#//' | tr '[:upper:]' '[:lower:]' | sort -u)"
      for color in $hex_colors; do
        match=false
        for ok in $snazzy_palette; do
          if [[ "$color" == "$ok" ]]; then
            match=true
            break
          fi
        done
        if [[ "$match" == false ]]; then
          echo "[palette] non-Snazzy color #$color in $file_path" >&2
        fi
      done
    fi
    ;;
esac

exit 0
