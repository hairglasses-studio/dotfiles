#!/usr/bin/env bash
# mod-ticker.sh — hg ticker module
# Pin/unpin streams, switch playlists, pause/resume, and inspect the ticker.
# All commands dispatch through ticker-control.sh into Quickshell IPC.

_TICKER_CONTROL="$HG_DOTFILES/scripts/ticker-control.sh"

ticker_description() {
  echo "Quickshell ticker control (pin, playlist, pause, status)"
}

ticker_commands() {
  cat <<'CMDS'
status	Print current stream, playlist, and pause state
pin	Pin a specific stream (pass stream name)
unpin	Release pin, resume rotation
playlist	Switch playlist (pass name: main/coding/focus/lock/recording)
pause	Toggle pause
shuffle	Shuffle-mode toggle for current playlist (on|off|toggle)
list-streams	List all available streams
list-playlists	List playlist files
restart	Restart Quickshell (dotfiles-quickshell.service)
logs	Tail dotfiles-quickshell.service systemd logs
recover-monitor	Restore DP-2 to native mode after DSC fallback (ticker clipping)
next	Advance one stream
prev	Rewind one stream
pin-toggle	Pin current stream if unpinned, else unpin
reload	Hot-reload ticker state and stream catalogue
banner	Flash a ticker banner — `hg ticker banner <text> [color]`
snooze-urgent	Dismiss an active urgent-mode escalation early
CMDS
}

ticker_run() {
  local cmd="${1:-status}"
  shift || true
  case "$cmd" in
    status)
      "$_TICKER_CONTROL" status
      ;;
    pin)
      local stream="${1:?usage: hg ticker pin <stream>}"
      "$_TICKER_CONTROL" pin "$stream"
      ;;
    unpin)
      "$_TICKER_CONTROL" unpin
      ;;
    playlist)
      local name="${1:?usage: hg ticker playlist <name>}"
      "$_TICKER_CONTROL" playlist "$name"
      ;;
    pause)
      "$_TICKER_CONTROL" pause
      ;;
    shuffle)
      "$_TICKER_CONTROL" shuffle "${1:-toggle}"
      ;;
    list-streams)
      "$_TICKER_CONTROL" list-streams
      ;;
    list-playlists)
      "$_TICKER_CONTROL" list-playlists
      ;;
    restart)
      systemctl --user restart dotfiles-quickshell.service
      systemctl --user is-active dotfiles-quickshell.service
      ;;
    logs)
      journalctl --user -u dotfiles-quickshell.service --since "2 min ago" --no-pager
      ;;
    recover-monitor)
      # Diagnose + restore the ultrawide from DSC fallback to native mode.
      "$HG_DOTFILES/scripts/hg-dsc-recover.sh" "$@"
      ;;
    next|prev|pin-toggle|reload|snooze-urgent)
      "$_TICKER_CONTROL" "$cmd"
      ;;
    banner)
      local text="${1:?usage: hg ticker banner <text> [color]}"
      local color="${2:-#29f0ff}"
      "$_TICKER_CONTROL" banner "$text" "$color"
      ;;
    *)
      printf 'unknown command: %s\n' "$cmd" >&2
      printf 'try: hg ticker --help\n' >&2
      return 2
      ;;
  esac
}
