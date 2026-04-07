#!/bin/bash
# Weather via wttr.in — hides when offline or unknown location
WEATHER=$(curl -s --max-time 5 "wttr.in?format=%c%t" 2>/dev/null | tr -d '+')

if [ -z "$WEATHER" ] || echo "$WEATHER" | grep -q "Unknown"; then
    sketchybar --set $NAME drawing=off
else
    sketchybar --set $NAME drawing=on label="$WEATHER"
fi
