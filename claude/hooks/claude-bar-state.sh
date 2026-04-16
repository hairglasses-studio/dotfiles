#!/bin/bash
# Write Claude session state to /tmp/claude-status.json for Ironbar polling
# Called from Claude Code hooks with state + optional JSON on stdin

STATE="${1:-idle}"
SESSION="${CLAUDE_SESSION_ID:-}"
CWD="${PWD:-}"
SLUG="${CWD##*/}"
NOW=$(date +%s)

# Read cost/context from stdin hook JSON if available
COST="" CTX=""
if [[ -t 0 ]]; then
  : # no stdin
else
  INPUT=$(cat 2>/dev/null || true)
  if [[ -n "$INPUT" ]]; then
    COST=$(echo "$INPUT" | jq -r '.cost.total_cost_usd // empty' 2>/dev/null)
    CTX=$(echo "$INPUT" | jq -r '.context_window.used_percentage // empty' 2>/dev/null)
  fi
fi

jq -n \
  --arg state "$STATE" \
  --arg session "$SESSION" \
  --arg slug "$SLUG" \
  --arg cost "${COST:-0}" \
  --arg ctx "${CTX:-0}" \
  --argjson ts "$NOW" \
  '{state: $state, session: $session, slug: $slug, cost: ($cost|tonumber), context_pct: ($ctx|tonumber), timestamp: $ts}' \
  > /tmp/claude-status.json 2>/dev/null
