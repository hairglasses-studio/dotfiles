#!/usr/bin/env bash
# claude-notify-debounce.sh — Debounced desktop notifications for Claude Code hooks.
# Prevents notification spam by enforcing a 30s cooldown per category.
#
# Usage: claude-notify-debounce.sh <category> <title> <body>
# Categories: stop, attention, error, info

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/notify.sh"

CATEGORY="${1:-info}"
TITLE="${2:-Claude Code}"
BODY="${3:-}"
COOLDOWN=30

STAMP_DIR="${XDG_RUNTIME_DIR:-/tmp}/claude-notify"
mkdir -p "$STAMP_DIR"

STAMP_FILE="$STAMP_DIR/$CATEGORY"

if [[ -f "$STAMP_FILE" ]]; then
  last=$(stat -c %Y "$STAMP_FILE" 2>/dev/null || echo 0)
  now=$(date +%s)
  if (( now - last < COOLDOWN )); then
    exit 0
  fi
fi

touch "$STAMP_FILE"

case "$CATEGORY" in
  error)    hg_notify_critical "Claude Code" "$BODY" ;;
  attention) hg_notify "Claude Code" "$BODY" ;;
  *)        hg_notify_low "Claude Code" "$BODY" ;;
esac
