#!/usr/bin/env bash
# dev-console.sh — Yakuake-style Claude Code system dev console
#
# Bound to Mod+grave via pyprland scratchpad (dev-console).
#
# Session resume order:
#   1. If ~/.local/state/dev-console/pinned-session-id exists and contains a
#      valid session UUID that has a JSONL file, resume that specific session.
#   2. Otherwise fall back to `claude --continue` (most-recent session in cwd).
#
# Pin a session with: pin-dev-console-session [<session-id>]
# Unpin with:         pin-dev-console-session --unpin
set -euo pipefail

cd "$HOME/hairglasses-studio/dotfiles"

PIN_FILE="$HOME/.local/state/dev-console/pinned-session-id"
PROJECT_DIR="$HOME/.claude/projects/-home-hg-hairglasses-studio-dotfiles"

if [[ -s "$PIN_FILE" ]]; then
    SESSION_ID=$(tr -d '[:space:]' < "$PIN_FILE")
    if [[ -n "$SESSION_ID" ]] && [[ -f "$PROJECT_DIR/$SESSION_ID.jsonl" ]]; then
        exec claude --resume "$SESSION_ID"
    fi
    echo "dev-console: pinned session '$SESSION_ID' missing, falling back to --continue" >&2
fi

exec claude --continue
