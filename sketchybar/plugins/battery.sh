#!/bin/bash
GREEN=0xff5af78e
YELLOW=0xfff3f99d
RED=0xffff5c57
GRAY=0xff686868

PERCENTAGE="$(pmset -g batt | grep -Eo "\d+%" | cut -d% -f1)"
CHARGING="$(pmset -g batt | grep 'AC Power')"

if [ -z "$PERCENTAGE" ]; then
    sketchybar --set $NAME drawing=off
    exit 0
fi

if [ -n "$CHARGING" ]; then
    ICON="󰂄"
    COLOR=$GREEN
elif [ "$PERCENTAGE" -gt 50 ]; then
    ICON="󰁹"
    COLOR=$GREEN
elif [ "$PERCENTAGE" -gt 20 ]; then
    ICON="󰁾"
    COLOR=$YELLOW
else
    ICON="󰁺"
    COLOR=$RED
fi

sketchybar --set $NAME \
    drawing=on \
    icon="$ICON" \
    icon.color=$COLOR \
    label="${PERCENTAGE}%" \
    label.color=$GRAY
