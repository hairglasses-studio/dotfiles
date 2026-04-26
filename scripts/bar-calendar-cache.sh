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

# Capture gcalcli output in a subshell with stdin closed so a missing
# OAuth token (which would otherwise prompt for client_id/client_secret)
# fails fast instead of hanging the timer. The capture also keeps a
# non-zero gcalcli exit from tripping `set -o pipefail` when piped into
# tail/awk below; on failure we degrade to a one-line status cache so
# the bar shows what's wrong instead of a stale entry.
if ! agenda="$(timeout 15 gcalcli agenda --military --tsv --details=end --nocolor \
    "$start" "$end" </dev/null 2>/dev/null)"; then
  printf 'gcalcli not authenticated\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

# TSV output: start_date\tstart_time\tend_date\tend_time\t...\ttitle\tlocation
# Fields beyond title vary; safer to parse the last text column.
printf '%s\n' "$agenda" \
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
