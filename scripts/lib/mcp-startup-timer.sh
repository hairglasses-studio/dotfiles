#!/usr/bin/env bash
# SessionStart hook: record session start timestamp for trending.
# Writes append-only JSONL to ~/.claude/startup-times.jsonl.
# Run async — timing data doesn't need to go into additionalContext.
set -euo pipefail

out="$HOME/.claude/startup-times.jsonl"
ts=$(date -u +%Y-%m-%dT%H:%M:%SZ)
epoch=$(date +%s)
cwd="${CWD:-$PWD}"
session="${CLAUDE_SESSION_ID:-unknown}"

printf '{"ts":"%s","epoch":%d,"session":"%s","cwd":"%s"}\n' \
  "$ts" "$epoch" "$session" "$cwd" >> "$out"
