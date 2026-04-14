#!/usr/bin/env bash
set -euo pipefail

state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/hyprdynamicmonitors"
log_file="$state_dir/hyprdynamicmonitors.log"
startup_wait_secs="${HYPRDYNAMICMONITORS_STARTUP_WAIT_SECS:-0}"
mkdir -p "$state_dir"
touch "$log_file"

timestamp() {
  date --iso-8601=seconds
}

log_line() {
  printf '%s %s\n' "$(timestamp)" "$1" >>"$log_file"
}

monitors_ready() {
  local monitors_json
  monitors_json="$(hyprctl -j monitors 2>/dev/null || true)"
  [[ "$monitors_json" == *'"name"'* ]]
}

wait_for_monitors() {
  local waited=0

  while ! monitors_ready; do
    if [[ "$startup_wait_secs" =~ ^[0-9]+$ ]] && (( startup_wait_secs > 0 )) && (( waited >= startup_wait_secs )); then
      log_line 'startup wait expired before Hyprland reported any monitors'
      printf '%s\n' 'hyprdynamicmonitors: startup wait expired before any Hyprland monitors appeared' >&2
      return 1
    fi

    if (( waited == 0 )); then
      log_line 'waiting for Hyprland monitors to appear before starting hyprdynamicmonitors'
    elif (( waited % 30 == 0 )); then
      log_line "still waiting for Hyprland monitors after ${waited}s"
    fi

    sleep 1
    ((waited += 1))
  done

  (( waited > 0 )) && log_line "Hyprland monitors detected after ${waited}s"
  return 0
}

should_suppress() {
  local line="$1"
  case "$line" in
    *'level="warning" msg="No power line available, will use a default:'*)
      return 0
      ;;
    *'level="info" msg="Inferred power line"'*)
      return 0
      ;;
    *'level="info" msg="UPower D-Bus power detection initialized"'*)
      return 0
      ;;
    *'level="info" msg="UPower D-Bus lid detection initialized"'*)
      return 0
      ;;
    *'level="info" msg="Power events are disabled, waiting for ctx cancellation"'*)
      return 0
      ;;
    *'level="info" msg="Lid events are disabled, waiting for ctx cancellation"'*)
      return 0
      ;;
    *'level="info" msg="Using profile"'*)
      return 0
      ;;
    *'level="info" msg="Configuration already correctly linked"'*)
      return 0
      ;;
    *'level="info" msg="Not sending notifications since the config has not been changed"'*)
      return 0
      ;;
    *'level="info" msg="Listening for monitor and power events..."'*)
      return 0
      ;;
  esac
  return 1
}

set +e
wait_for_monitors || exit 1
stdbuf -oL -eL hyprdynamicmonitors "$@" 2>&1 | while IFS= read -r line; do
  log_line "$line"
  if should_suppress "$line"; then
    continue
  fi
  printf '%s\n' "$line"
done
status=${PIPESTATUS[0]}
set -e

exit "$status"
