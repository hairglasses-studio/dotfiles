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

# Recovery warnings (if any recent crash events)
events_file="$HOME/.claude/recovery-events.jsonl"
if [[ -f "$events_file" ]]; then
  recent_crashes=$(tail -20 "$events_file" 2>/dev/null | grep -c '"event":"crash"' 2>/dev/null || true)
  if [[ "${recent_crashes:-0}" -gt 0 ]]; then
    context+="Recent crash events: $recent_crashes"$'\n'
  fi
fi

if [[ -n "$context" ]]; then
  printf '%s' "$context"
fi
