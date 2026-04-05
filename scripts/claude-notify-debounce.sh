#!/usr/bin/env bash
# claude-notify-debounce.sh — Rate-limit Claude Code desktop notifications
# Usage: claude-notify-debounce.sh <category> <title> [body]
# At most 1 notification per COOLDOWN seconds per category.
set -euo pipefail

CATEGORY="${1:-general}"
TITLE="${2:-Claude Code}"
BODY="${3:-}"
DEBOUNCE_DIR="/tmp/claude-notify-debounce"
COOLDOWN=30

mkdir -p "$DEBOUNCE_DIR"
LOCKFILE="$DEBOUNCE_DIR/$CATEGORY"

if [[ -f "$LOCKFILE" ]]; then
    last=$(stat -c %Y "$LOCKFILE" 2>/dev/null || echo 0)
    now=$(date +%s)
    if (( now - last < COOLDOWN )); then
        exit 0
    fi
fi

touch "$LOCKFILE"
exec notify-send -a "Claude Code" -i "/usr/share/icons/Papirus-Dark/48x48/apps/terminal.svg" -u low "$TITLE" "$BODY" 2>/dev/null || true
