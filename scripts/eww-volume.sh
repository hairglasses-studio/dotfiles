#!/usr/bin/env bash
# eww-volume.sh — Event-driven volume daemon for eww bar
# Updates eww variables instantly on PulseAudio/PipeWire sink changes
set -euo pipefail

update() {
  local default_sink
  default_sink=$(wpctl status 2>/dev/null | grep -A 1 "Audio/Sink" | grep '\*' | awk '{print $3}' | head -1)
  [[ -z "$default_sink" ]] && default_sink="@DEFAULT_AUDIO_SINK@"

  local vol muted
  vol=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ 2>/dev/null | awk '{printf "%.0f", $2*100}') || vol=0
  muted=$(wpctl get-volume @DEFAULT_AUDIO_SINK@ 2>/dev/null | grep -q MUTED && echo "true" || echo "false")

  eww update bar_vol="$vol" 2>/dev/null
}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_volume_listen() {
  update
  pactl subscribe 2>/dev/null | grep --line-buffered "sink" | while read -r _; do
    update
  done
}

# Resilient: pactl subscribe dies on PipeWire restarts
resilient_listen _volume_listen
