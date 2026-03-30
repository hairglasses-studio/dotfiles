#!/bin/bash
# On wake, WiFi reconnects after a few seconds — delay before querying.
if [ "$SENDER" = "system_woke" ]; then
    sleep 5
fi

SSID="$(ipconfig getsummary en0 2>/dev/null | awk -F' : ' '/SSID/{print $2}' | head -1)"
if [ -z "$SSID" ]; then
    sketchybar --set $NAME icon=󰖪 label="offline"
else
    sketchybar --set $NAME icon=󰖩 label="$SSID"
fi
