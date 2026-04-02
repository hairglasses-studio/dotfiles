#!/usr/bin/env bash
# eww-calendar.sh — Generate calendar grid JSON for eww sidebar.
# Usage: eww-calendar.sh [YYYY MM] (defaults to current month)
# Output: JSON {year, month, month_name, weeks: [[{day, today, other}]]}
set -euo pipefail

YEAR="${1:-$(date +%Y)}"
MONTH="${2:-$(date +%-m)}"
TODAY=$(date +%-d)
TODAY_MONTH=$(date +%-m)
TODAY_YEAR=$(date +%Y)

python3 -c "
import calendar, json, sys

year, month = int('$YEAR'), int('$MONTH')
today = int('$TODAY') if year == int('$TODAY_YEAR') and month == int('$TODAY_MONTH') else -1

cal = calendar.Calendar(firstweekday=0)  # Monday start
weeks = []
for week in cal.monthdayscalendar(year, month):
    days = []
    for d in week:
        days.append({
            'day': d if d != 0 else '',
            'today': d == today and d != 0,
            'other': d == 0
        })
    weeks.append(days)

print(json.dumps({
    'year': year,
    'month': month,
    'month_name': calendar.month_name[month],
    'weeks': weeks
}))
"
