#!/usr/bin/env bash
# juhradial-wheel-apply.sh — apply MX Master 4 wheel hardware settings via Solaar
#
# juhradial currently persists wheel preferences in config.json but does not
# reliably push them back into the device on daemon restart from this workflow.
# Use Solaar as a narrow HID++ bridge for wheel-only state until upstream fixes
# the daemon apply path.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

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
config_file="$(juhradial_config_dir)/config.json"
[[ -f "$config_file" ]] || config_file="$(juhradial_seed_dir)/config.json"

natural="$(jq -r '.scroll.natural // false' "$config_file")"
smooth="$(jq -r '.scroll.smooth // true' "$config_file")"
smartshift="$(jq -r '.scroll.smartshift // false' "$config_file")"
smartshift_threshold="$(jq -r '.scroll.smartshift_threshold // 15' "$config_file")"
scroll_mode="$(jq -r '.scroll.mode // "ratchet"' "$config_file")"
haptics_enabled="$(jq -r '.haptics.enabled // false' "$config_file")"

if [[ "$smartshift" == "true" ]]; then
  smartshift_value="$smartshift_threshold"
else
  smartshift_value="1"
fi

if [[ "$scroll_mode" == "free_spin" ]]; then
  ratchet_mode="Freespinning"
else
  ratchet_mode="Ratcheted"
fi

if [[ "$haptics_enabled" == "true" ]]; then
  haptic_level="60"
else
  haptic_level="0"
fi

dbus_set_scroll_state() {
  if [[ "$smartshift" == "true" ]]; then
    juhradial_gdbus_call \
      --dest org.kde.juhradialmx \
      --object-path /org/kde/juhradialmx/Daemon \
      --method org.kde.juhradialmx.Daemon.SetSmartShift \
      "$smartshift" \
      "$smartshift_threshold" >/dev/null 2>&1 || return 1
  fi

  juhradial_gdbus_call \
    --dest org.kde.juhradialmx \
    --object-path /org/kde/juhradialmx/Daemon \
    --method org.kde.juhradialmx.Daemon.SetHiresscrollMode \
    "$smooth" \
    "$natural" \
    false >/dev/null 2>&1
}

apply_setting() {
  local name="$1"
  local value="$2"
  local attempt

  for attempt in 1 2 3; do
    if juhradial_solaar_timeout 10s config "$device" "$name" "$value" >/dev/null 2>&1; then
      log "Applied $name=$value"
      return 0
    fi
    sleep 1
  done

  printf 'Failed to apply %s=%s via Solaar after %d attempts\n' "$name" "$value" "$attempt" >&2
  return 1
}

if dbus_set_scroll_state; then
  log "Applied scroll mode via juhradial D-Bus"
else
  log "juhradial D-Bus unavailable; falling back to Solaar-only apply"
fi

apply_setting hires-smooth-invert "$natural"
apply_setting hires-smooth-resolution "$smooth"
apply_setting hires-scroll-mode false
apply_setting scroll-ratchet-torque 75
apply_setting scroll-ratchet "$ratchet_mode"
if [[ "$smartshift" == "true" ]]; then
  apply_setting smart-shift "$smartshift_value"
fi
apply_setting thumb-scroll-mode true
apply_setting haptic-level "$haptic_level"
