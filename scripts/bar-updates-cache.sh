#!/usr/bin/env bash
# bar-updates-cache.sh — Cache writer for the Ironbar updates widget
#
# Runs checkupdates and writes the count to /tmp/bar-updates.txt.
# Run by bar-updates.timer (every 60 minutes); Ironbar reads the cache file
# with `cat`, never blocking the GTK main loop on pacman network I/O.
#
# Cache format: plain text label, e.g. " 7"  (empty string if no updates)
# On failure: leaves any existing cache intact.

set -euo pipefail

CACHE_FILE="/tmp/bar-updates.txt"
TMPFILE="$(mktemp /tmp/bar-updates.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

n=$(checkupdates 2>/dev/null | wc -l) || n=0
if (( n > 0 )); then
  printf ' %s\n' "$n" > "$TMPFILE"
else
  printf '' > "$TMPFILE"
fi
mv "$TMPFILE" "$CACHE_FILE"
