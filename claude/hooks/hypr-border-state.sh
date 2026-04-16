#!/bin/bash
# Dynamic Hyprland border color based on Claude state
# Called from Claude Code hooks with a state argument
# Usage: hypr-border-state.sh <working|done|error|subagent>

STATE="${1:-done}"
DEFAULT="rgba(ff6ac1ee) rgba(6ae4ffee) 45deg"

case "$STATE" in
  working)  COLOR="rgba(ffb86cee) rgba(f3f99dee) 45deg" ;;  # amber→yellow
  done)     COLOR="rgba(5af78eee) rgba(6ae4ffee) 45deg" ;;  # green→cyan
  error)    COLOR="rgba(ff5c57ee) rgba(ff6ac1ee) 45deg" ;;  # red→magenta
  subagent) COLOR="rgba(57c7ffee) rgba(9aedfee) 45deg" ;;  # cyan→light-cyan
  *)        COLOR="$DEFAULT" ;;
esac

hyprctl keyword general:col.active_border "$COLOR" &>/dev/null

# Reset to default after 10s (background, non-blocking)
(sleep 10; hyprctl keyword general:col.active_border "$DEFAULT" &>/dev/null) &
disown
