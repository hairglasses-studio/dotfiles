#!/usr/bin/env bash
# dev-console.sh — Yakuake-style Claude Code system dev console
#
# Bound to Mod+grave via pyprland scratchpad (dev-console).
#
# Spawn-time visual chain (runs once per session due to pypr lazy=true):
#   1. Apply a random Hypr-DarkWindow shader from the dev-console playlist
#      to this window via `hyprctl dispatch darkwindow:shade`.
#   2. Play a brief TTE matrix banner announcing the session.
#   3. Resume the claude-code session (pinned or --continue).
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

DOTFILES="$HOME/hairglasses-studio/dotfiles"
PIN_FILE="$HOME/.local/state/dev-console/pinned-session-id"
PROJECT_DIR="$HOME/.claude/projects/-home-hg-hairglasses-studio-dotfiles"
SHADER_PLAYLIST="$DOTFILES/kitty/shaders/playlists/dev-console.txt"

apply_shader() {
  [[ -f "$SHADER_PLAYLIST" ]] || return 0
  command -v hyprctl >/dev/null 2>&1 || return 0
  command -v jq >/dev/null 2>&1 || return 0

  local shader addr
  shader="$(awk 'NF && $1 !~ /^#/ { sub(/\.glsl$/,""); print }' "$SHADER_PLAYLIST" | shuf -n 1 || true)"
  [[ -n "$shader" ]] || return 0

  # The dev-console class is unique (pypr scratchpad), so we can resolve our
  # own Hyprland client address by class. We're running inside the window at
  # this point, so the client already exists.
  addr="$(hyprctl clients -j 2>/dev/null \
    | jq -r '.[] | select(.class=="dev-console") | .address' 2>/dev/null \
    | head -1 || true)"
  [[ -n "$addr" && "$addr" != "null" ]] || return 0

  hyprctl dispatch darkwindow:shade "address:$addr $shader" >/dev/null 2>&1 || true
}

boot_banner() {
  command -v tte >/dev/null 2>&1 || return 0
  # Short matrix flash; capped at 120fps so it completes in <500ms.
  # Fallback to silent if tte errors (e.g. unsupported terminal).
  printf '%s\n' "◢ DEV-CONSOLE ◣ claude-code session resume" \
    | tte --frame-rate 120 matrix --final-color "29f0ff" 2>/dev/null || true
}

apply_shader
boot_banner

if [[ -s "$PIN_FILE" ]]; then
    SESSION_ID=$(tr -d '[:space:]' < "$PIN_FILE")
    if [[ -n "$SESSION_ID" ]] && [[ -f "$PROJECT_DIR/$SESSION_ID.jsonl" ]]; then
        exec claude --resume "$SESSION_ID"
    fi
    echo "dev-console: pinned session '$SESSION_ID' missing, falling back to --continue" >&2
fi

exec claude --continue
