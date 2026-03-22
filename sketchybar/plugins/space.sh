#!/bin/bash
# Highlight active AeroSpace workspace

BLUE=0xff57c7ff
GRAY=0xff686868

if [ "$FOCUSED_WORKSPACE" = "$SID" ] || [ "$SENDER" = "space_change" ]; then
    sketchybar --set $NAME \
        icon.color=0xff000000 \
        background.color=$BLUE \
        background.drawing=on
else
    sketchybar --set $NAME \
        icon.color=$GRAY \
        background.color=0x00000000 \
        background.drawing=off
fi
