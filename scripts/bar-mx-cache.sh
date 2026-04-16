#!/usr/bin/env bash
# bar-mx-cache.sh — Cache writer for the Ironbar MX Master battery widget
#
# Reads MX Master 4 battery via bluetoothctl and writes a label to /tmp/bar-mx.txt.
# Run by bar-mx.timer (every 5 minutes); Ironbar reads the cache file with `cat`,
# never blocking the GTK main loop on bluetoothctl IPC.
#
# Cache format: plain text label, e.g. "󰂂 85%"  (empty string if disconnected)
# On failure: leaves any existing cache intact.

set -euo pipefail

CACHE_FILE="/tmp/bar-mx.txt"
TMPFILE="$(mktemp /tmp/bar-mx.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

battery=""
while IFS= read -r line; do
  if [[ "$line" =~ Percentage:\ 0x([0-9a-fA-F]+) ]]; then
    battery=$(( 16#${BASH_REMATCH[1]} ))
  fi
done < <(bluetoothctl info 2>/dev/null | grep -i "battery\|percentage" || true)

if [[ -n "$battery" ]]; then
  printf '󰂂 %s%%\n' "$battery" > "$TMPFILE"
else
  printf '' > "$TMPFILE"
fi
mv "$TMPFILE" "$CACHE_FILE"
