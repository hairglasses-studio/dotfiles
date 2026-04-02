#!/usr/bin/env bash
# eww-calendar-sync.sh — Sync Google Calendar events for eww sidebar.
# Fetches today + 7 days of events, writes JSON cache.
# Uses gcalcli if available, otherwise writes a placeholder.
# Designed to run as a oneshot via systemd timer (every 5 min).
set -euo pipefail

STATE_DIR="$HOME/.local/state/hg"
CACHE="$STATE_DIR/calendar-events.json"

mkdir -p "$STATE_DIR"

TODAY=$(date +%Y-%m-%d)
END=$(date -d "+7 days" +%Y-%m-%d)

fetch_gcalcli() {
  # gcalcli agenda output: date time - time title
  # Parse into JSON with python
  gcalcli agenda "$TODAY" "$END" --tsv 2>/dev/null | python3 -c "
import sys, json, datetime

events = []
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    parts = line.split('\t')
    if len(parts) < 5:
        continue
    start_date, start_time, end_date, end_time, title = parts[0], parts[1], parts[2], parts[3], '\t'.join(parts[4:])
    all_day = (start_time == '' and end_time == '')
    events.append({
        'title': title,
        'start': f'{start_date}T{start_time}' if start_time else start_date,
        'end': f'{end_date}T{end_time}' if end_time else end_date,
        'location': '',
        'all_day': all_day
    })

print(json.dumps(events, indent=2))
"
}

fetch_placeholder() {
  # Placeholder events so eww integration can be tested
  python3 -c "
import json, datetime

now = datetime.datetime.now()
today = now.strftime('%Y-%m-%d')
tomorrow = (now + datetime.timedelta(days=1)).strftime('%Y-%m-%d')

events = [
    {
        'title': 'Daily standup',
        'start': f'{today}T09:00',
        'end': f'{today}T09:30',
        'location': 'Google Meet',
        'all_day': False
    },
    {
        'title': 'Lunch break',
        'start': f'{today}T12:00',
        'end': f'{today}T13:00',
        'location': '',
        'all_day': False
    },
    {
        'title': 'Sprint review',
        'start': f'{tomorrow}T14:00',
        'end': f'{tomorrow}T15:00',
        'location': 'Conference Room A',
        'all_day': False
    },
    {
        'title': 'Release day',
        'start': tomorrow,
        'end': tomorrow,
        'location': '',
        'all_day': True
    }
]

print(json.dumps(events, indent=2))
"
}

# Try gcalcli first, fall back to placeholder
if command -v gcalcli &>/dev/null; then
  EVENTS=$(fetch_gcalcli)
else
  echo "gcalcli not found, writing placeholder events" >&2
  EVENTS=$(fetch_placeholder)
fi

# Atomic write to avoid partial reads
TMP=$(mktemp "$CACHE.XXXXXX")
echo "$EVENTS" > "$TMP"
mv -f "$TMP" "$CACHE"

echo "Calendar events synced to $CACHE ($(echo "$EVENTS" | python3 -c 'import sys,json; print(len(json.load(sys.stdin)))') events)"
