#!/usr/bin/env bash
# swaync-to-toast.sh — forward a swaync notification to the active ticker.
#
# Registered in swaync/config.json under `scripts` for the "Critical"
# urgency category. swaync calls this with the notification metadata
# in environment variables (SUMMARY, BODY, APP_NAME, URGENCY). Quickshell is
# the primary banner owner; legacy toast-ticker DBus is only a rollback path.
#
# Usage (manual test): SUMMARY="build failed" URGENCY=Critical \
#   ~/hairglasses-studio/dotfiles/scripts/swaync-to-toast.sh

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
summary="${SUMMARY:-${1:-}}"
body="${BODY:-${2:-}}"
urgency="${URGENCY:-${3:-Normal}}"
app_name="${APP_NAME:-swaync}"
ticker_control="$script_dir/ticker-control.sh"

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

if ! "$ticker_control" banner "$msg" "$color" >/dev/null 2>&1; then
  # Rollback path for sessions still running the GTK toast service.
  if pgrep -f toast-ticker.py >/dev/null 2>&1; then
    gdbus call --session \
      --dest io.hairglasses.toast \
      --object-path /io/hairglasses/toast \
      --method io.hairglasses.Toast.ShowToast \
      "$msg" "$color" >/dev/null 2>&1 || true
  fi
fi

# Critical urgencies also flip the ticker into urgent mode so the scrolling bar
# itself flashes. The control wrapper prefers Quickshell and falls back to the
# legacy DBus ticker when needed.
if [[ "$urgency" == "Critical" ]]; then
  "$ticker_control" urgent true >/dev/null 2>&1 || true
fi
