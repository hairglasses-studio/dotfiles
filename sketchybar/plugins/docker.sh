#!/bin/bash
# Running Docker container count — hides when daemon unavailable
BLUE=0xff57c7ff
GRAY=0xff686868

if ! docker info &>/dev/null; then
    sketchybar --set $NAME drawing=off
    exit 0
fi

COUNT=$(docker ps -q 2>/dev/null | wc -l | tr -d ' ')

if [ "$COUNT" -gt 0 ]; then
    sketchybar --set $NAME drawing=on icon.color=$BLUE label.color=$BLUE label="$COUNT"
else
    sketchybar --set $NAME drawing=on icon.color=$GRAY label.color=$GRAY label="0"
fi
