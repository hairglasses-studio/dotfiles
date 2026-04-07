#!/bin/bash
# Screen recorder toggle — start/stop via PID file
RED=0xffff5c57
GRAY=0xff686868
PIDFILE="/tmp/.sketchybar-recorder.pid"

if [ "$SENDER" = "mouse.clicked" ]; then
    if [ -f "$PIDFILE" ]; then
        kill "$(cat "$PIDFILE")" 2>/dev/null
        rm -f "$PIDFILE"
    else
        screencapture -v "$HOME/Desktop/recording-$(date +%s).mov" &
        echo $! > "$PIDFILE"
    fi
fi

if [ -f "$PIDFILE" ] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    sketchybar --set $NAME icon=󰑊 icon.color=$RED label="REC" label.drawing=on label.color=$RED
else
    rm -f "$PIDFILE"
    sketchybar --set $NAME icon=󰑋 icon.color=$GRAY label.drawing=off
fi
