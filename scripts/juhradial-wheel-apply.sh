#!/usr/bin/env bash
# juhradial-wheel-apply.sh — apply MX Master 4 wheel hardware settings via Solaar
#
# juhradial currently persists wheel preferences in config.json but does not
# reliably push them back into the device on daemon restart from this workflow.
# Use Solaar as a narrow HID++ bridge for wheel-only state until upstream fixes
# the daemon apply path.
set -euo pipefail

quiet=false
for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

log() {
  $quiet || printf '[juhradial-wheel] %s\n' "$*"
}

if ! command -v solaar >/dev/null 2>&1; then
  log "Solaar not installed; skipping wheel compatibility apply"
  exit 0
fi

device="${JUHRADIAL_WHEEL_DEVICE:-MX Master 4}"
config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
cache_home="${XDG_CACHE_HOME:-$HOME/.cache}"
state_home="${XDG_STATE_HOME:-$HOME/.local/state}"

apply_setting() {
  local name="$1"
  local value="$2"
  local attempt

  for attempt in 1 2 3; do
    if env \
      HOME="$HOME" \
      XDG_CONFIG_HOME="$config_home" \
      XDG_CACHE_HOME="$cache_home" \
      XDG_STATE_HOME="$state_home" \
      timeout 10s solaar config "$device" "$name" "$value" >/dev/null 2>&1; then
      log "Applied $name=$value"
      return 0
    fi
    sleep 1
  done

  printf 'Failed to apply %s=%s via Solaar after %d attempts\n' "$name" "$value" "$attempt" >&2
  return 1
}

apply_setting hires-smooth-invert 0
apply_setting smart-shift 50
apply_setting haptic-level 0
