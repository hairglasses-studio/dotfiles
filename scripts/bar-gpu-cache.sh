#!/usr/bin/env bash
# bar-gpu-cache.sh — Cache writer for the Ironbar GPU power widget
#
# Reads nvidia-smi power draw and writes a label to /tmp/bar-gpu.txt.
# Ironbar reads the cache file with `cat`, never blocking the GTK main loop
# on an nvidia-smi fork + driver IPC.
#
# Cache format: plain text label, e.g. "125W"  (empty string on failure).

set -euo pipefail

CACHE_FILE="/tmp/bar-gpu.txt"
TMPFILE="$(mktemp /tmp/bar-gpu.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

power=$(nvidia-smi --query-gpu=power.draw --format=csv,noheader,nounits 2>/dev/null | awk 'NR==1 {printf "%.0f", $1}')

if [[ -n "$power" ]]; then
  printf '%sW\n' "$power" > "$TMPFILE"
else
  printf '' > "$TMPFILE"
fi
mv "$TMPFILE" "$CACHE_FILE"
