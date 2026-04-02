#!/usr/bin/env bash
# eww-cava.sh — Stream cava audio visualization data for eww.
# Outputs JSON array of 8 bar values (0-100) per line, continuously.
# Usage: deflisten cava_bars `eww-cava.sh`
set -euo pipefail

CAVA_CONFIG="/tmp/eww-cava.conf"

# Create temporary cava config for raw output
cat > "$CAVA_CONFIG" << 'EOF'
[general]
bars = 8
framerate = 30
sensitivity = 100

[input]
method = pulse

[output]
method = raw
raw_target = /dev/stdout
data_format = ascii
ascii_max_range = 100
EOF

# Run cava and format output as JSON arrays
cava -p "$CAVA_CONFIG" 2>/dev/null | while IFS=';' read -r -a bars; do
  # Convert semicolon-separated values to JSON array
  json="["
  for i in "${!bars[@]}"; do
    val="${bars[$i]}"
    val="${val%%[[:space:]]*}"  # trim whitespace
    [[ -z "$val" ]] && val="0"
    [[ $i -gt 0 ]] && json+=","
    json+="$val"
  done
  json+="]"
  echo "$json"
done
