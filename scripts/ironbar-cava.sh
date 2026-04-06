#!/usr/bin/env bash
# ironbar-cava.sh — Cava output for ironbar script module (watch mode)
# Outputs Unicode block characters, one line per frame

CAVA_CONFIG=$(mktemp /tmp/ironbar-cava-XXXXXX.conf)
trap 'rm -f "$CAVA_CONFIG"' EXIT

cat > "$CAVA_CONFIG" <<'EOF'
[general]
framerate = 20
bars = 8

[output]
method = raw
raw_target = /dev/stdout
data_format = ascii
ascii_max_range = 7
EOF

exec cava -p "$CAVA_CONFIG" 2>/dev/null | sed -u 's/;//g;s/0/▁/g;s/1/▂/g;s/2/▃/g;s/3/▄/g;s/4/▅/g;s/5/▆/g;s/6/▇/g;s/7/█/g'
