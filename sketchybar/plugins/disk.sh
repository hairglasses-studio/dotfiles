#!/bin/bash
# Root volume disk usage — escalates to red above 90%
RED=0xffff5c57
GRAY=0xff686868

USAGE=$(df -h / | awk 'NR==2{print $5}')
PCT=${USAGE%%%}

if [ "$PCT" -gt 90 ] 2>/dev/null; then
    sketchybar --set $NAME icon.color=$RED label.color=$RED label="$USAGE"
else
    sketchybar --set $NAME icon.color=$GRAY label.color=$GRAY label="$USAGE"
fi
