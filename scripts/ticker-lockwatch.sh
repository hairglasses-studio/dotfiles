#!/usr/bin/env bash
# ticker-lockwatch.sh — swap the ticker playlist when hyprlock is active.
#
# Polls pgrep -x hyprlock every 2s. On lock-state transitions, writes the
# target playlist name to the state file and restarts the ticker service so
# the change takes effect immediately. Locked-mode uses a minimal playlist
# that avoids leaking private context (no keybinds, git state, claude
# sessions, or calendar titles) to anyone looking at the monitor.
#
# State machine:
#   unlocked → locked   : save current playlist to .pre-lock, write "lock"
#   locked   → unlocked : read .pre-lock (or "main"), restore it
#
# Run as a systemd user service. Idempotent: safe to run multiple times.

set -uo pipefail

STATE_DIR="$HOME/.local/state/keybind-ticker"
ACTIVE_FILE="$STATE_DIR/active-playlist"
PRELOCK_FILE="$STATE_DIR/pre-lock-playlist"
LOCK_PLAYLIST="lock"
DEFAULT_PLAYLIST="main"
SERVICE="dotfiles-keybind-ticker.service"
POLL_S="${TICKER_LOCKWATCH_POLL_S:-2}"

mkdir -p "$STATE_DIR"

is_locked() {
  pgrep -x hyprlock >/dev/null 2>&1
}

current_playlist() {
  if [[ -f "$ACTIVE_FILE" ]]; then
    cat "$ACTIVE_FILE"
  else
    printf '%s' "$DEFAULT_PLAYLIST"
  fi
}

enter_lock() {
  local pre
  pre="$(current_playlist)"
  [[ "$pre" == "$LOCK_PLAYLIST" ]] && return
  printf '%s' "$pre" > "$PRELOCK_FILE"
  printf '%s' "$LOCK_PLAYLIST" > "$ACTIVE_FILE"
  systemctl --user restart "$SERVICE" 2>/dev/null || true
}

exit_lock() {
  local restore="$DEFAULT_PLAYLIST"
  if [[ -f "$PRELOCK_FILE" ]]; then
    restore="$(cat "$PRELOCK_FILE")"
    rm -f "$PRELOCK_FILE"
  fi
  [[ -z "$restore" ]] && restore="$DEFAULT_PLAYLIST"
  printf '%s' "$restore" > "$ACTIVE_FILE"
  systemctl --user restart "$SERVICE" 2>/dev/null || true
}

prev_state=""
while true; do
  if is_locked; then
    state="locked"
  else
    state="unlocked"
  fi

  if [[ "$state" != "$prev_state" ]]; then
    if [[ "$state" == "locked" ]]; then
      enter_lock
    elif [[ -n "$prev_state" ]]; then
      # Only restore on locked→unlocked; skip the very first tick to avoid
      # clobbering the active playlist on service startup.
      exit_lock
    fi
    prev_state="$state"
  fi

  sleep "$POLL_S"
done
