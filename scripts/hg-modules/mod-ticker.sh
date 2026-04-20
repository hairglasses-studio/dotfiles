#!/usr/bin/env bash
# mod-ticker.sh — hg ticker module
# Pin/unpin streams, switch playlists, pause/resume, and inspect the ticker.

_TICKER_STATE="$HOME/.local/state/keybind-ticker"
_TICKER_SERVICE="dotfiles-keybind-ticker.service"
_TICKER_HEADLESS="$HG_DOTFILES/scripts/ticker-headless.py"

ticker_description() {
  echo "Keybind ticker control (pin, playlist, pause, status)"
}

ticker_commands() {
  cat <<'CMDS'
status	Print current stream, playlist, and pause state
pin	Pin a specific stream (pass stream name)
unpin	Release pin, resume rotation
playlist	Switch playlist (pass name: main/coding/focus/lock/recording)
pause	Toggle pause
list-streams	List all available streams
list-playlists	List playlist files
show	Render a single stream plain-text once (pass stream name)
restart	Restart the ticker systemd service
logs	Tail the ticker systemd logs
CMDS
}

_ticker_write_state() {
  local key="$1" value="$2"
  mkdir -p "$_TICKER_STATE"
  if [[ -z "$value" ]]; then
    rm -f "$_TICKER_STATE/$key"
  else
    printf '%s' "$value" > "$_TICKER_STATE/$key"
  fi
}

_ticker_read_state() {
  local key="$1" default="${2:-}"
  if [[ -f "$_TICKER_STATE/$key" ]]; then
    cat "$_TICKER_STATE/$key"
  else
    printf '%s' "$default"
  fi
}

ticker_run() {
  local cmd="${1:-status}"
  shift || true
  case "$cmd" in
    status)
      printf 'service   : %s\n' "$(systemctl --user is-active "$_TICKER_SERVICE")"
      printf 'playlist  : %s\n' "$(_ticker_read_state active-playlist main)"
      printf 'current   : %s\n' "$(_ticker_read_state current-stream '(rotating)')"
      printf 'pinned    : %s\n' "$(_ticker_read_state pinned-stream '(none)')"
      if [[ -f "$_TICKER_STATE/paused" ]]; then
        printf 'paused    : yes\n'
      else
        printf 'paused    : no\n'
      fi
      ;;
    pin)
      local stream="${1:?usage: hg ticker pin <stream>}"
      _ticker_write_state pinned-stream "$stream"
      systemctl --user restart "$_TICKER_SERVICE"
      printf 'pinned to %s\n' "$stream"
      ;;
    unpin)
      _ticker_write_state pinned-stream ""
      systemctl --user restart "$_TICKER_SERVICE"
      printf 'unpinned\n'
      ;;
    playlist)
      local name="${1:?usage: hg ticker playlist <name>}"
      _ticker_write_state active-playlist "$name"
      systemctl --user restart "$_TICKER_SERVICE"
      printf 'playlist switched to %s\n' "$name"
      ;;
    pause)
      if [[ -f "$_TICKER_STATE/paused" ]]; then
        rm -f "$_TICKER_STATE/paused"
        printf 'resumed\n'
      else
        mkdir -p "$_TICKER_STATE"
        : > "$_TICKER_STATE/paused"
        printf 'paused\n'
      fi
      systemctl --user restart "$_TICKER_SERVICE"
      ;;
    list-streams)
      python3 "$_TICKER_HEADLESS" --list
      ;;
    list-playlists)
      for f in "$HG_DOTFILES/ticker/content-playlists"/*.txt; do
        printf '%s\n' "$(basename "$f" .txt)"
      done
      ;;
    show)
      local stream="${1:?usage: hg ticker show <stream>}"
      python3 "$_TICKER_HEADLESS" --stream "$stream"
      printf '\n'
      ;;
    restart)
      systemctl --user restart "$_TICKER_SERVICE"
      systemctl --user is-active "$_TICKER_SERVICE"
      ;;
    logs)
      journalctl --user -u "$_TICKER_SERVICE" --since "2 min ago" --no-pager
      ;;
    *)
      printf 'unknown command: %s\n' "$cmd" >&2
      printf 'try: hg ticker --help\n' >&2
      return 2
      ;;
  esac
}
