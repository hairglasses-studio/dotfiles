#!/bin/bash
# Uptime widget — parses kern.boottime for clean display
BOOT=$(sysctl -n kern.boottime | awk -F'[ ,]' '{print $4}')
NOW=$(date +%s)
SECS=$((NOW - BOOT))

DAYS=$((SECS / 86400))
HOURS=$(( (SECS % 86400) / 3600 ))
MINS=$(( (SECS % 3600) / 60 ))

if [ "$DAYS" -gt 0 ]; then
    LABEL="${DAYS}d ${HOURS}h"
elif [ "$HOURS" -gt 0 ]; then
    LABEL="${HOURS}h ${MINS}m"
else
    LABEL="${MINS}m"
fi

sketchybar --set $NAME label="$LABEL"
