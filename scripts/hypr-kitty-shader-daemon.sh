#!/usr/bin/env bash
# hypr-kitty-shader-daemon.sh — per-kitty-window random shader applier
#
# Listens on Hyprland's .socket2.sock for `openwindow>>` events. When a new
# kitty window appears, picks a random shader from
# kitty/shaders/playlists/ambient.txt and dispatches it to the new window by
# address via `hyprctl dispatch darkwindow:shade address:<addr> <shader>`.
#
# Per-window theme application is handled separately by kitty's own watcher
# at ~/.config/kitty/watchers/random_theme.py → kitty-shader-playlist.sh
# theme-for-window, which uses kitty's remote control to match by window id.
#
# This daemon does NOT cycle or dispatch on demand — that's the job of
# kitty-shader-playlist.sh (next/prev/random commands). It only fires on the
# initial openwindow event so each new kitty OS window gets its own random
# shader, independent of other kitty windows on the same desktop.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="$(cd "$SCRIPT_DIR/.." && pwd)"
PLAYLIST="$DOTFILES/kitty/shaders/playlists/ambient.txt"

log() { printf '[hypr-kitty-shader-daemon] %s\n' "$*" >&2; }

require() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || { log "missing: $cmd"; exit 127; }
}

require hyprctl
require socat
require shuf

[[ -f "$PLAYLIST" ]] || { log "missing playlist: $PLAYLIST"; exit 1; }

# Build an absolute-addressing hyprland socket2 path from the live session
# signature. Fail fast if the compositor isn't running yet — systemd restarts
# us via Restart=on-failure until the graphical-session.target ExecCondition
# passes.
sig="${HYPRLAND_INSTANCE_SIGNATURE:-}"
if [[ -z "$sig" ]]; then
  # Derive signature from the first hypr/<sig>/ directory under $XDG_RUNTIME_DIR
  rundir="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
  sig="$(find "$rundir/hypr" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' 2>/dev/null | head -1)"
fi
[[ -n "$sig" ]] || { log "no HYPRLAND_INSTANCE_SIGNATURE"; exit 1; }

socket="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/hypr/$sig/.socket2.sock"
[[ -S "$socket" ]] || { log "socket missing: $socket"; exit 1; }

log "listening on $socket"
log "playlist: $PLAYLIST"

pick_shader() {
  awk 'NF && $1 !~ /^#/ { sub(/\.glsl$/, ""); print }' "$PLAYLIST" | shuf -n 1
}

apply_shader() {
  local addr="$1"
  local shader
  shader="$(pick_shader)" || return 1
  [[ -n "$shader" ]] || return 1
  # Fire-and-forget — dispatcher returns "ok" synchronously, errors go to
  # Hyprland's notification system which we can't see from here anyway.
  hyprctl dispatch darkwindow:shade "address:0x$addr $shader" >/dev/null 2>&1 || {
    log "dispatch failed for addr=$addr shader=$shader"
    return 1
  }
  log "addr=0x$addr shader=$shader"
}

# socat UNIX-CONNECT gives us the event stream as UTF-8 lines.
# Event format: EVENT>>DATA  where for openwindow, DATA is:
#   ADDRESS,WORKSPACE_NAME,WINDOW_CLASS,WINDOW_TITLE
# ADDRESS is the hex portion of the pointer (no 0x prefix in the event body).
socat -u "UNIX-CONNECT:$socket" - 2>/dev/null | while IFS= read -r line; do
  case "$line" in
    openwindow\>\>*)
      data="${line#openwindow>>}"
      addr="${data%%,*}"
      rest="${data#*,}"
      ws="${rest%%,*}"
      rest2="${rest#*,}"
      class="${rest2%%,*}"
      title="${rest2#*,}"
      if [[ "$class" == "kitty" ]]; then
        apply_shader "$addr" || true
      fi
      ;;
  esac
done
