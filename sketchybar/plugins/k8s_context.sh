#!/bin/bash
CTX="$(kubectl config current-context 2>/dev/null | rev | cut -d/ -f1 | rev)"
if [ -z "$CTX" ]; then
    sketchybar --set $NAME drawing=off
else
    sketchybar --set $NAME drawing=on label="$CTX"
fi
