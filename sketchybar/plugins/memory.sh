#!/bin/bash
# RAM usage via memory_pressure (macOS built-in)
FREE=$(memory_pressure 2>/dev/null | awk '/memory free percentage/{gsub(/%/,""); print $NF}')
if [ -n "$FREE" ]; then
    MEM=$((100 - FREE))
else
    MEM="?"
fi
sketchybar --set $NAME label="${MEM}%"
