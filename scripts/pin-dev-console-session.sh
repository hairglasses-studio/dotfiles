#!/usr/bin/env bash
# pin-dev-console-session.sh — Pin a Claude Code session to the dev-console
# dropdown terminal (Mod+grave).
#
# After pinning, Mod+grave will resume the pinned session instead of the
# most-recent one. Useful when you want a specific long-running session to
# always be behind the dropdown terminal.
#
# Usage:
#   pin-dev-console-session                # pin most recent dotfiles session
#   pin-dev-console-session <session-id>   # pin a specific session UUID
#   pin-dev-console-session --unpin        # clear pin, go back to --continue
#   pin-dev-console-session --show         # print current pin (if any)
set -euo pipefail

STATE_DIR="$HOME/.local/state/dev-console"
PIN_FILE="$STATE_DIR/pinned-session-id"
PROJECT_DIR="$HOME/.claude/projects/-home-hg-hairglasses-studio-dotfiles"

mkdir -p "$STATE_DIR"

case "${1:-}" in
    --unpin|-u)
        if [[ -f "$PIN_FILE" ]]; then
            rm -f "$PIN_FILE"
            echo "dev-console: unpinned (will resume most recent session)"
        else
            echo "dev-console: no pin set"
        fi
        exit 0
        ;;
    --show|-s)
        if [[ -s "$PIN_FILE" ]]; then
            echo "dev-console pinned session: $(cat "$PIN_FILE")"
        else
            echo "dev-console: no pin set (will use claude --continue)"
        fi
        exit 0
        ;;
    --help|-h)
        sed -n '2,13p' "$0" | sed 's/^# \{0,1\}//'
        exit 0
        ;;
esac

if [[ -n "${1:-}" ]]; then
    SESSION_ID="$1"
    if [[ ! -f "$PROJECT_DIR/$SESSION_ID.jsonl" ]]; then
        echo "error: no JSONL for session '$SESSION_ID' in $PROJECT_DIR" >&2
        echo "" >&2
        echo "Available sessions (newest first):" >&2
        ls -1t "$PROJECT_DIR"/*.jsonl 2>/dev/null | head -5 | while read -r f; do
            echo "  $(basename "$f" .jsonl)" >&2
        done
        exit 1
    fi
else
    # Default: most recent session in this project
    LATEST=$(ls -1t "$PROJECT_DIR"/*.jsonl 2>/dev/null | head -1 || true)
    if [[ -z "$LATEST" ]]; then
        echo "error: no sessions found in $PROJECT_DIR" >&2
        exit 1
    fi
    SESSION_ID=$(basename "$LATEST" .jsonl)
fi

printf '%s\n' "$SESSION_ID" > "$PIN_FILE"
echo "dev-console: pinned session $SESSION_ID"
echo "Mod+\` will now resume this session."
