#!/usr/bin/env bash
# swaync-to-toast.sh — forward a swaync notification to toast-ticker.
#
# Registered in swaync/config.json under `scripts` for the "Critical"
# urgency category. swaync calls this with the notification metadata
# in environment variables (SUMMARY, BODY, APP_NAME, URGENCY). We call
# `gdbus` to invoke ShowToast on io.hairglasses.toast with a color
# keyed off urgency.
#
# Usage (manual test): SUMMARY="build failed" URGENCY=Critical \
#   ~/hairglasses-studio/dotfiles/scripts/swaync-to-toast.sh

set -euo pipefail

summary="${SUMMARY:-${1:-}}"
body="${BODY:-${2:-}}"
urgency="${URGENCY:-${3:-Normal}}"
app_name="${APP_NAME:-swaync}"

# No-op if toast-ticker isn't running — avoid flooding the journal with
# gdbus errors.
if ! pgrep -f toast-ticker.py >/dev/null 2>&1; then
  exit 0
fi

# Colour palette (Hairglasses Neon tokens).
case "$urgency" in
  Critical) color="#ff5c8a" ;;   # danger pink
  Low)      color="#66708f" ;;   # muted
  *)        color="#29f0ff" ;;   # primary cyan
esac

# Prefer summary; if empty, fall back to body or app name.
msg="${summary:-${body:-$app_name}}"
# Collapse newlines so the 28px single-line banner renders cleanly.
msg="${msg//$'\n'/ }"

gdbus call --session \
  --dest io.hairglasses.toast \
  --object-path /io/hairglasses/toast \
  --method io.hairglasses.Toast.ShowToast \
  "$msg" "$color" >/dev/null 2>&1 || true

# Critical urgencies also flip the keybind-ticker into urgent mode so
# the scrolling bar itself flashes (amplified glitch + CA for 10 s).
# The ticker DBus (Phase 1) exposes SetUrgent(b); the call is best-
# effort — if the ticker isn't running we silently skip.
if [[ "$urgency" == "Critical" ]]; then
  gdbus call --session \
    -d io.hairglasses.keybind_ticker \
    -o /io/hairglasses/Ticker \
    -m io.hairglasses.Ticker.SetUrgent \
    true >/dev/null 2>&1 || true
fi
