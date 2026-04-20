#!/usr/bin/env bash
# bar-calendar-cache.sh — Next-events cache for the ticker calendar stream
#
# Runs `gcalcli agenda` for the next 24h and writes 1 event per line to
# /tmp/bar-calendar.txt. Paired with bar-calendar.timer (10-minute interval).
#
# Cache format: one line per event, e.g. "14:30  Standup (30m)".
# Empty file = no events in window.

set -euo pipefail

CACHE_FILE="/tmp/bar-calendar.txt"
TMPFILE="$(mktemp /tmp/bar-calendar.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v gcalcli >/dev/null; then
  printf 'gcalcli missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

start="$(date +'%Y-%m-%dT%H:%M')"
end="$(date -d '+24 hours' +'%Y-%m-%dT%H:%M')"

# TSV output: start_date\tstart_time\tend_date\tend_time\t...\ttitle\tlocation
# Fields beyond title vary; safer to parse the last text column.
gcalcli agenda --military --tsv --details=end --nocolor "$start" "$end" 2>/dev/null \
  | tail -n +2 \
  | awk -F '\t' 'NF >= 5 {
      time = $2
      title = $5
      if ($6 != "") title = title " @ " $6
      gsub(/[\r\n]+/, " ", title)
      printf "%s  %s\n", time, title
    }' \
  | head -8 > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
