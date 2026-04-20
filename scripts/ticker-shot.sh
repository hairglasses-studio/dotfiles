#!/usr/bin/env bash
# ticker-shot.sh — capture ONLY the ticker strip, never the full monitor.
#
# Claude Code (and most vision models) reject images whose longer side exceeds
# ~1568 px or ~1.15M total pixels. A full DP-2 screenshot (3840x1080 = 4.1M px)
# always trips that limit; the ticker itself is a 28 px strip at the bottom.
# This script uses `grim -g` to capture exactly that region — the resulting
# PNG is tiny (~30 KB) and always under Claude's ingestion limits.
#
# Usage:
#   ticker-shot.sh                   current content → /tmp/ticker-shot.png
#   ticker-shot.sh --pin <stream>    pin stream, wait past 400ms wipe, shoot, unpin
#   ticker-shot.sh --monitor DP-3    target a different ticker instance
#   ticker-shot.sh --output path     explicit output path
#   ticker-shot.sh --wait 0.5        extra pre-capture delay (seconds)
#   ticker-shot.sh --height 32       strip height in px (default 28)
#   ticker-shot.sh --scale 0.5       downscale output via magick (default 1.0)
#   ticker-shot.sh --print-geom      just print "X,Y WxH" for the strip and exit
#
# Exits non-zero on failure. On success prints the absolute output path.

set -euo pipefail

MONITOR="DP-2"
OUTPUT="/tmp/ticker-shot.png"
PIN=""
WAIT="0"
SCALE="1.0"
HEIGHT=28
PRINT_GEOM=0

_die() { printf 'ticker-shot: %s\n' "$*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --monitor)    MONITOR="${2:-}"; shift 2 ;;
    --pin)        PIN="${2:-}";     shift 2 ;;
    --output|-o)  OUTPUT="${2:-}";  shift 2 ;;
    --wait)       WAIT="${2:-}";    shift 2 ;;
    --scale)      SCALE="${2:-}";   shift 2 ;;
    --height)     HEIGHT="${2:-}";  shift 2 ;;
    --print-geom) PRINT_GEOM=1;     shift ;;
    -h|--help)
      sed -n '2,19p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) _die "unknown flag: $1 (try --help)" ;;
  esac
done

command -v grim    >/dev/null 2>&1 || _die "grim not installed"
command -v hyprctl >/dev/null 2>&1 || _die "hyprctl not found (Hyprland not running?)"

# Resolve monitor geometry via hyprctl
read -r MX MY MW MH < <(hyprctl monitors -j | python3 -c "
import json, sys
name = '$MONITOR'
for m in json.load(sys.stdin):
    if m['name'] == name:
        print(m['x'], m['y'], m['width'], m['height'])
        break
")
[[ -n "${MX:-}" ]] || _die "monitor $MONITOR not found in hyprctl monitors"

STRIP_Y=$((MY + MH - HEIGHT))
GEOM="${MX},${STRIP_Y} ${MW}x${HEIGHT}"

if [[ "$PRINT_GEOM" == "1" ]]; then
  printf '%s\n' "$GEOM"
  exit 0
fi

# Optional: pin a stream for the duration of the capture
STATE_DIR="$HOME/.local/state/keybind-ticker"
PREV_PIN=""
DID_PIN=0
_restore_pin() {
  [[ "$DID_PIN" == "1" ]] || return 0
  if [[ -n "$PREV_PIN" ]]; then
    printf '%s' "$PREV_PIN" > "$STATE_DIR/pinned-stream"
  else
    rm -f "$STATE_DIR/pinned-stream"
  fi
  pkill -USR1 -f 'keybind-ticker.py --layer' 2>/dev/null || true
}
trap _restore_pin EXIT

if [[ -n "$PIN" ]]; then
  mkdir -p "$STATE_DIR"
  PREV_PIN="$(cat "$STATE_DIR/pinned-stream" 2>/dev/null || true)"
  printf '%s' "$PIN" > "$STATE_DIR/pinned-stream"
  DID_PIN=1
  pkill -USR1 -f 'keybind-ticker.py --layer' 2>/dev/null || true
  # Cover signal handler hop + 400ms stream-change wipe + a few frames settle
  sleep 0.7
fi

# Extra user-requested delay (e.g. for slow-threaded streams still fetching)
if [[ "$WAIT" != "0" && "$WAIT" != "0.0" ]]; then
  sleep "$WAIT"
fi

TMP="$(mktemp /tmp/ticker-shot-XXXXXX.png)"
if ! grim -g "$GEOM" "$TMP" 2>/dev/null; then
  rm -f "$TMP"
  _die "grim capture failed for region '$GEOM'"
fi

# Optional downscale (e.g. --scale 0.5 → 1920x14, further reducing file size)
if [[ "$SCALE" != "1.0" && "$SCALE" != "1" ]]; then
  if ! command -v magick >/dev/null 2>&1; then
    rm -f "$TMP"
    _die "magick not installed (required for --scale)"
  fi
  PERCENT="$(python3 -c "print(int(float('$SCALE') * 100))")"
  magick "$TMP" -filter Lanczos -resize "${PERCENT}%" "$TMP" \
    || { rm -f "$TMP"; _die "magick resize failed"; }
fi

mkdir -p "$(dirname "$OUTPUT")"
mv -f "$TMP" "$OUTPUT"
printf '%s\n' "$OUTPUT"
