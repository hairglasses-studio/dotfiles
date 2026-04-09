#!/usr/bin/env bash
# mx-wheel-apply.sh — Restore MX Master 4 scroll wheel state via Solaar
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEVICE_MAC="${BT_MX_MASTER:-D2:8E:C5:DE:9F:CC}"
DEVICE_NAME="${MX_MASTER_DEVICE_NAME:-MX Master 4}"
MAX_ATTEMPTS="${MX_WHEEL_MAX_ATTEMPTS:-12}"
SLEEP_SECS="${MX_WHEEL_RETRY_SLEEP:-1}"
QUIET=false

source "$SCRIPT_DIR/lib/notify.sh"

info() {
  $QUIET || printf '\033[0;36m:: %s\033[0m\n' "$*"
}

ok() {
  $QUIET || printf '\033[0;32m   %s\033[0m\n' "$*"
}

warn() {
  $QUIET || printf '\033[0;33m   %s\033[0m\n' "$*"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quiet)
      QUIET=true
      shift
      ;;
    *)
      printf 'usage: %s [--quiet]\n' "$(basename "$0")" >&2
      exit 2
      ;;
  esac
done

is_connected() {
  bluetoothctl info "$DEVICE_MAC" 2>/dev/null | grep -q 'Connected: yes'
}

apply_setting() {
  local setting="$1"
  local value="$2"
  solaar config "$DEVICE_NAME" "$setting" "$value" >/dev/null 2>&1
}

restore_wheel() {
  apply_setting "hires-smooth-invert" "on" \
    && apply_setting "scroll-ratchet" "Ratcheted" \
    && apply_setting "smart-shift" "1" \
    && apply_setting "scroll-ratchet-torque" "75" \
    && apply_setting "haptic-level" "0" \
    && apply_setting "thumb-scroll-mode" "on"
}

if ! is_connected; then
  info "$DEVICE_NAME is not connected; skipping wheel restore"
  exit 0
fi

info "Restoring $DEVICE_NAME wheel settings..."
for attempt in $(seq 1 "$MAX_ATTEMPTS"); do
  if restore_wheel; then
    ok "wheel settings restored"
    exit 0
  fi

  if (( attempt < MAX_ATTEMPTS )); then
    sleep "$SLEEP_SECS"
  fi
done

warn "failed to restore $DEVICE_NAME wheel settings after $MAX_ATTEMPTS attempts"
hg_notify_critical "MX Master 4" "Wheel restore failed after reconnect"
exit 1
