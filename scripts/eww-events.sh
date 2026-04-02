#!/usr/bin/env bash
# eww-events.sh — Fetch upcoming calendar events for eww sidebar.
# Uses gcal MCP if available, falls back to static display.
# Output: JSON array of {time, title} (max 5 events)
set -euo pipefail

# Try to read from cached gcal events (written by external sync)
CACHE="$HOME/.local/state/hg/calendar-events.json"
if [[ -f "$CACHE" ]] && [[ $(find "$CACHE" -mmin -30 2>/dev/null | wc -l) -gt 0 ]]; then
  cat "$CACHE"
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
