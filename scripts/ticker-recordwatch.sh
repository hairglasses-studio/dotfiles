#!/usr/bin/env bash
# ticker-recordwatch.sh — swap ticker to the recording playlist while a screen
# recorder (wf-recorder / wl-screenrec) is running, and maintain the live
# metadata cache that the `recording` stream renders.
#
# State machine:
#   idle → recording : write pre-recording playlist to state, swap to
#                       "recording", write /tmp/bar-recording.txt with
#                       "tool\tpid\tstart_epoch", notify-send START.
#   recording → idle : clear cache, restore pre-recording playlist,
#                       notify-send STOP.
#
# While recording, re-write the cache every tick so the duration stays fresh.
# Poll interval: 2s (TICKER_RECORDWATCH_POLL_S to override).

set -uo pipefail

STATE_DIR="$HOME/.local/state/keybind-ticker"
ACTIVE_FILE="$STATE_DIR/active-playlist"
PRE_FILE="$STATE_DIR/pre-recording-playlist"
CACHE_FILE="/tmp/bar-recording.txt"
REC_PLAYLIST="recording"
DEFAULT_PLAYLIST="main"
SERVICE="dotfiles-keybind-ticker.service"
POLL_S="${TICKER_RECORDWATCH_POLL_S:-2}"

mkdir -p "$STATE_DIR"

current_playlist() {
  if [[ -f "$ACTIVE_FILE" ]]; then
    cat "$ACTIVE_FILE"
  else
    printf '%s' "$DEFAULT_PLAYLIST"
  fi
}

detect_recorder() {
  # Returns "tool\tpid\tstart_epoch" on stdout if a recorder is running.
  local pid
  for tool in wf-recorder wl-screenrec; do
    pid="$(pgrep -x "$tool" | head -1)"
    if [[ -n "$pid" ]]; then
      # lstart as epoch seconds from /proc
      local start
      start="$(stat -c %Y "/proc/$pid" 2>/dev/null)" || start="$(date +%s)"
      printf '%s\t%s\t%s' "$tool" "$pid" "$start"
      return 0
    fi
  done
  return 1
}

# While recording, companion layer-shell surfaces (window-label,
# fleet-sparkline) get captured in the output video if the user records
# a full-screen region. Stop them on entry so the recording is clean
# (just the ticker bar and whatever the user explicitly includes);
# restart them on exit.
COMPANION_SURFACES=(
  dotfiles-window-label.service
  dotfiles-fleet-sparkline.service
  dotfiles-lyrics-ticker.service
)

enter_recording() {
  local info="$1" pre
  pre="$(current_playlist)"
  if [[ "$pre" != "$REC_PLAYLIST" ]]; then
    printf '%s' "$pre" > "$PRE_FILE"
    printf '%s' "$REC_PLAYLIST" > "$ACTIVE_FILE"
    printf '%s' "$info" > "$CACHE_FILE"
    systemctl --user stop "${COMPANION_SURFACES[@]}" 2>/dev/null || true
    systemctl --user restart "$SERVICE" 2>/dev/null || true
    notify-send -u low -t 2500 -i media-record "Recording started" \
      "Ticker swapped to recording playlist." 2>/dev/null || true
  else
    # Already in recording mode — just refresh the cache in place
    printf '%s' "$info" > "$CACHE_FILE"
  fi
}

exit_recording() {
  local restore="$DEFAULT_PLAYLIST"
  if [[ -f "$PRE_FILE" ]]; then
    restore="$(cat "$PRE_FILE")"
    rm -f "$PRE_FILE"
  fi
  [[ -z "$restore" ]] && restore="$DEFAULT_PLAYLIST"
  : > "$CACHE_FILE"
  printf '%s' "$restore" > "$ACTIVE_FILE"
  systemctl --user restart "$SERVICE" 2>/dev/null || true
  systemctl --user start "${COMPANION_SURFACES[@]}" 2>/dev/null || true
  notify-send -u low -t 2500 -i media-playback-stop "Recording stopped" \
    "Ticker restored to '$restore' playlist." 2>/dev/null || true
}

prev_state=""
while true; do
  if info="$(detect_recorder)"; then
    state="recording"
    enter_recording "$info"
  else
    state="idle"
    # Only restore when moving recording→idle; skip the first tick on startup.
    if [[ "$prev_state" == "recording" ]]; then
      exit_recording
    fi
  fi
  prev_state="$state"
  sleep "$POLL_S"
done
