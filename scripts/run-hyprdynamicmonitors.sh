#!/usr/bin/env bash
set -euo pipefail

state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/hyprdynamicmonitors"
log_file="$state_dir/hyprdynamicmonitors.log"
mkdir -p "$state_dir"
touch "$log_file"

timestamp() {
  date --iso-8601=seconds
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
stdbuf -oL -eL hyprdynamicmonitors "$@" 2>&1 | while IFS= read -r line; do
  printf '%s %s\n' "$(timestamp)" "$line" >>"$log_file"
  if should_suppress "$line"; then
    continue
  fi
  printf '%s\n' "$line"
done
status=${PIPESTATUS[0]}
set -e

exit "$status"
