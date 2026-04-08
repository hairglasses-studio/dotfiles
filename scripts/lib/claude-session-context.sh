#!/usr/bin/env bash
# SessionStart hook: inject lightweight git + recovery context
# Output goes to additionalContext (advisory, not blocking)
set -euo pipefail

context=""

# Git branch + recent commits (if in a git repo)
if git rev-parse --is-inside-work-tree &>/dev/null; then
  branch=$(git branch --show-current 2>/dev/null || echo "detached")
  recent=$(git log --oneline -3 2>/dev/null || true)
  if [[ -n "$recent" ]]; then
    context+="Git branch: $branch"$'\n'
    context+="Recent commits:"$'\n'"$recent"$'\n'
  fi
fi

# Recovery warnings (if any recent crash events)
events_file="$HOME/.claude/recovery-events.jsonl"
if [[ -f "$events_file" ]]; then
  recent_crashes=$(tail -20 "$events_file" 2>/dev/null | grep -c '"event":"crash"' 2>/dev/null || echo "0")
  if [[ "$recent_crashes" -gt 0 ]]; then
    context+="WARNING: $recent_crashes recent crash events in recovery log"$'\n'
  fi
fi

if [[ -n "$context" ]]; then
  echo "$context"
fi
