#!/usr/bin/env bash
# eww-events.sh — Fetch upcoming calendar events for eww sidebar.
# Uses gcal MCP if available, falls back to static display.
# Output: JSON array of {time, title} (max 5 events)
set -euo pipefail

# Try to read from cached gcal events (written by eww-calendar-sync.sh)
# Transform {start, end, title} → {time, title} for eww widget
CACHE="$HOME/.local/state/hg/calendar-events.json"
if [[ -f "$CACHE" ]] && [[ $(find "$CACHE" -mmin -30 2>/dev/null | wc -l) -gt 0 ]]; then
  python3 -c "
import json, sys, datetime
with open('$CACHE') as f:
    events = json.load(f)
now = datetime.datetime.now()
today = now.strftime('%Y-%m-%d')
tomorrow = (now + datetime.timedelta(days=1)).strftime('%Y-%m-%d')
out = []
for e in events:
    start = e.get('start', '')
    if not (start.startswith(today) or start.startswith(tomorrow)):
        continue
    if 'T' in start:
        time = start.split('T')[1][:5]
    else:
        time = 'All day'
    out.append({'time': time, 'title': e['title']})
    if len(out) >= 5:
        break
print(json.dumps(out))
" 2>/dev/null
  exit 0
fi

# Fallback: show today's date info
python3 -c "
import json, datetime
now = datetime.datetime.now()
events = [
    {'time': now.strftime('%H:%M'), 'title': 'Today: ' + now.strftime('%A, %B %d')},
]
print(json.dumps(events))
"
