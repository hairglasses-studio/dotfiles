#!/bin/bash
# Push Claude session state to Ironbar via IPC + state file
# Called from Claude Code hooks: claude-bar-state.sh <working|done|error|idle>

STATE="${1:-idle}"
SLUG="${PWD##*/}"
STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/claude"
mkdir -p "$STATE_DIR" 2>/dev/null

# State icons
case "$STATE" in
  working)  ICON="" ;;
  done)     ICON="" ;;
  error)    ICON="" ;;
  *)        ICON="" ;;
esac

# Read cost from stdin hook JSON if available
COST=""
if [[ ! -t 0 ]]; then
  INPUT=$(cat 2>/dev/null || true)
  [[ -n "$INPUT" ]] && COST=$(echo "$INPUT" | jq -r '.cost.total_cost_usd // empty' 2>/dev/null)
fi

# Write state file (fallback for non-IPC consumers)
printf '%s\t%s\t%s\t%s\n' "$STATE" "$SLUG" "${COST:-0}" "$(date +%s)" > "$STATE_DIR/bar-state" 2>/dev/null

# Push to Ironbar via ironbar CLI (sets ironvar variables)
if command -v ironbar &>/dev/null; then
  ironbar var set claude_state "${ICON} ${SLUG}" 2>/dev/null
  [[ -n "$COST" && "$COST" != "0" ]] && ironbar var set claude_cost "\$${COST}" 2>/dev/null || ironbar var set claude_cost "" 2>/dev/null
fi
