#!/bin/bash
# Now Playing — Apple Music / Spotify

MAGENTA=0xffff6ac1
GRAY=0xff686868

# Check Apple Music first
STATE="$(osascript -e 'tell application "Music" to player state' 2>/dev/null)"
if [ "$STATE" = "playing" ]; then
    TRACK="$(osascript -e 'tell application "Music" to name of current track' 2>/dev/null)"
    ARTIST="$(osascript -e 'tell application "Music" to artist of current track' 2>/dev/null)"
    sketchybar --set "$NAME" \
        drawing=on \
        icon=󰎈 \
        icon.color=$MAGENTA \
        label="${ARTIST} — ${TRACK}" \
        label.color=$GRAY \
        label.max_chars=40
    exit 0
fi

# Check Spotify
STATE="$(osascript -e 'tell application "Spotify" to player state' 2>/dev/null)"
if [ "$STATE" = "playing" ]; then
    TRACK="$(osascript -e 'tell application "Spotify" to name of current track' 2>/dev/null)"
    ARTIST="$(osascript -e 'tell application "Spotify" to artist of current track' 2>/dev/null)"
    sketchybar --set "$NAME" \
        drawing=on \
        icon= \
        icon.color=$MAGENTA \
        label="${ARTIST} — ${TRACK}" \
        label.color=$GRAY \
        label.max_chars=40
    exit 0
fi

# Nothing playing
sketchybar --set "$NAME" drawing=off
