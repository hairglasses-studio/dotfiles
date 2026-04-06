#!/usr/bin/env bash
set -euo pipefail
# en16-encoder-key.sh — Convert EN16 encoder rotation to a directional keystroke.
# Called by mapitall EN16 profile (Grid.toml) with {value} substitution.
#
# EN16 relative binary offset: CW sends value > 0.5, CCW sends value < 0.5
# (mapitall normalizes MIDI CC 0-127 to float 0.0-1.0)
#
# Usage: en16-encoder-key.sh <value> [cw_key] [ccw_key]
#   value:    normalized float from mapitall {value} substitution
#   cw_key:   key name for clockwise rotation (default: Down)
#   ccw_key:  key name for counter-clockwise rotation (default: Up)

VALUE="$1"
CW_KEY="${2:-Down}"
CCW_KEY="${3:-Up}"

if (( $(echo "$VALUE > 0.5" | bc -l) )); then
  wtype -k "$CW_KEY"
else
  wtype -k "$CCW_KEY"
fi
