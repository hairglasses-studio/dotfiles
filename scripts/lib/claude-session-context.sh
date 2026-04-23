#!/usr/bin/env bash
# SessionStart hook: inject lightweight git + recovery context
# Output goes to additionalContext (advisory, not blocking)
set -euo pipefail

context=""
cwd="${CWD:-$PWD}"

# Git branch + recent commits (if in a git repo)
if git -C "$cwd" rev-parse --is-inside-work-tree &>/dev/null; then
  branch=$(git -C "$cwd" branch --show-current 2>/dev/null || echo "detached")
  recent=$(git -C "$cwd" log --oneline -3 2>/dev/null || true)
  if [[ -n "$recent" ]]; then
    context+="Git branch: $branch"$'\n'
    context+="Recent commits:"$'\n'"$recent"$'\n'
  fi
fi

# Recovery warnings.
#
# Two record schemas land in this file:
#   1. Legacy crash events      — {"event":"crash",...}
#   2. Watchdog-emitted events  — {"type":"mcp_dead"|"event_bus_dead"|
#                                  "event_bus_stale"|"ticker_stale",...}
#
# Surface both so the model learns at session start whether the MCP
# server or the event bus died between sessions. Without this the
# recovery channel is effectively off for anything the watchdog sees —
# the whole chassis exists to report those failures and they must land.
events_file="$HOME/.claude/recovery-events.jsonl"
if [[ -f "$events_file" ]]; then
  recent_crashes=$(tail -40 "$events_file" 2>/dev/null | grep -c '"event":"crash"' 2>/dev/null || true)
  if [[ "${recent_crashes:-0}" -gt 0 ]]; then
    context+="Recent crash events: $recent_crashes"$'\n'
  fi
  # Watchdog-emitted events: count each type separately so the model
  # can dispatch /heal or /canary with a specific focus.
  for t in mcp_dead event_bus_dead event_bus_stale ticker_stale; do
    c=$(tail -40 "$events_file" 2>/dev/null | grep -c "\"type\":\"$t\"" 2>/dev/null || true)
    if [[ "${c:-0}" -gt 0 ]]; then
      context+="Recent watchdog event ($t): $c"$'\n'
    fi
  done
fi

if [[ -n "$context" ]]; then
  printf '%s' "$context"
fi
