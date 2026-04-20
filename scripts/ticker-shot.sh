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
#   ticker-shot.sh --no-switch       fail (don't auto-switch) when pin isn't in playlist
#   ticker-shot.sh --max-width N     cap output width (default 1536 — Claude ingestion limit)
#
# When `--pin <stream>` targets a stream not in the active playlist
# (e.g. `recording` when main.txt is active), the tool auto-switches to
# the first playlist containing it, shoots, then restores the previous
# playlist + pin on EXIT.
#
# Exits non-zero on failure. On success prints the absolute output path.

set -euo pipefail

MONITOR="DP-2"
OUTPUT="/tmp/ticker-shot.png"
PIN=""
WAIT="0"
SCALE="1.0"
HEIGHT=39
PRINT_GEOM=0
NO_SWITCH=0
MAX_WIDTH=1536

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
PLAYLISTS_DIR="$(dirname "$SCRIPT_DIR")/ticker/content-playlists"

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
    --no-switch)  NO_SWITCH=1;      shift ;;
    --max-width)  MAX_WIDTH="${2:-}"; shift 2 ;;
    -h|--help)
      sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) _die "unknown flag: $1 (try --help)" ;;
  esac
done

command -v grim    >/dev/null 2>&1 || _die "grim not installed"
command -v hyprctl >/dev/null 2>&1 || _die "hyprctl not found (Hyprland not running?)"

# Resolve monitor geometry via hyprctl. `x`/`y` are already in LOGICAL
# coordinates, but `width`/`height` are PHYSICAL pixels on HiDPI outputs
# (scale > 1). grim -g takes logical coords — so divide by scale.
read -r MX MY MW MH < <(hyprctl monitors -j | python3 -c "
import json, sys
name = '$MONITOR'
for m in json.load(sys.stdin):
    if m['name'] == name:
        scale = float(m.get('scale', 1)) or 1.0
        lw = int(round(m['width']  / scale))
        lh = int(round(m['height'] / scale))
        print(m['x'], m['y'], lw, lh)
        break
")
[[ -n "${MX:-}" ]] || _die "monitor $MONITOR not found in hyprctl monitors"

STRIP_Y=$((MY + MH - HEIGHT))
GEOM="${MX},${STRIP_Y} ${MW}x${HEIGHT}"

if [[ "$PRINT_GEOM" == "1" ]]; then
  printf '%s\n' "$GEOM"
  exit 0
fi

# Optional: pin a stream — saves prior pin (+ playlist if we have to
# switch because the requested stream isn't in the active one), then
# restores both on EXIT so an interrupted shot never strands the ticker.
STATE_DIR="$HOME/.local/state/keybind-ticker"
PREV_PIN=""
PREV_PLAYLIST=""
DID_PIN=0
DID_PLAYLIST_SWITCH=0

_send_usr1() { pkill -USR1 -f 'keybind-ticker.py --layer' 2>/dev/null || true; }

_restore() {
  if [[ "$DID_PIN" == "1" ]]; then
    if [[ -n "$PREV_PIN" ]]; then
      printf '%s' "$PREV_PIN" > "$STATE_DIR/pinned-stream"
    else
      rm -f "$STATE_DIR/pinned-stream"
    fi
  fi
  if [[ "$DID_PLAYLIST_SWITCH" == "1" ]]; then
    printf '%s' "$PREV_PLAYLIST" > "$STATE_DIR/active-playlist"
  fi
  [[ "$DID_PIN" == "1" || "$DID_PLAYLIST_SWITCH" == "1" ]] && _send_usr1
}
trap _restore EXIT

_find_playlist_with_stream() {
  local stream="$1"
  for f in "$PLAYLISTS_DIR"/*.txt; do
    [[ -r "$f" ]] || continue
    if grep -qx -- "$stream" "$f" 2>/dev/null; then
      basename "$f" .txt
      return 0
    fi
  done
  return 1
}

if [[ -n "$PIN" ]]; then
  mkdir -p "$STATE_DIR"
  PREV_PIN="$(cat "$STATE_DIR/pinned-stream" 2>/dev/null || true)"
  PREV_PLAYLIST="$(cat "$STATE_DIR/active-playlist" 2>/dev/null || echo main)"

  # Auto-switch playlist if the requested pin isn't in the active one.
  # Reload order in keybind-ticker is (playlist → pin), and a playlist
  # change clears the pin file, so we must apply the playlist first and
  # let it settle before writing + signalling the pin.
  ACTIVE_PLAYLIST_FILE="$PLAYLISTS_DIR/${PREV_PLAYLIST}.txt"
  if [[ -r "$ACTIVE_PLAYLIST_FILE" ]] && ! grep -qx -- "$PIN" "$ACTIVE_PLAYLIST_FILE"; then
    if [[ "$NO_SWITCH" == "1" ]]; then
      _die "stream '$PIN' not in playlist '$PREV_PLAYLIST' (rerun without --no-switch or switch playlist manually)"
    fi
    TARGET_PLAYLIST="$(_find_playlist_with_stream "$PIN" || true)"
    [[ -n "$TARGET_PLAYLIST" ]] || _die "stream '$PIN' not found in any playlist under $PLAYLISTS_DIR"
    printf '%s' "$TARGET_PLAYLIST" > "$STATE_DIR/active-playlist"
    DID_PLAYLIST_SWITCH=1
    _send_usr1
    sleep 0.45   # let playlist change settle — clears pin on ticker side
  fi

  printf '%s' "$PIN" > "$STATE_DIR/pinned-stream"
  DID_PIN=1
  _send_usr1
  sleep 0.7   # signal + 400ms wipe + a few frames settle
fi

# Extra user-requested delay (e.g. for slow-threaded streams still fetching)
if [[ "$WAIT" != "0" && "$WAIT" != "0.0" ]]; then
  sleep "$WAIT"
fi

TMP="$(mktemp /tmp/ticker-shot-XXXXXX.png)"
# `-s 1.0` forces logical-sized output; without it grim uses the greatest
# output scale factor and doubles pixels on HiDPI (scale=2) monitors,
# which trips Claude's dimension cap even when the logical region is tiny.
if ! grim -s 1.0 -g "$GEOM" "$TMP" 2>/dev/null; then
  rm -f "$TMP"
  _die "grim capture failed for region '$GEOM'"
fi

# Explicit user-requested rescale (kept for legacy callers)
if [[ "$SCALE" != "1.0" && "$SCALE" != "1" ]]; then
  command -v magick >/dev/null 2>&1 || { rm -f "$TMP"; _die "magick required for --scale"; }
  PERCENT="$(python3 -c "print(int(float('$SCALE') * 100))")"
  magick "$TMP" -filter Lanczos -resize "${PERCENT}%" "$TMP" \
    || { rm -f "$TMP"; _die "magick resize failed"; }
fi

# Automatic width cap so the PNG always ingests into Claude (1568 px limit
# on the longer side — we use 1536 by default for a safety margin).
# No-op when the region is already narrow enough.
OUT_W="$(magick identify -format '%w' "$TMP" 2>/dev/null || echo 0)"
if [[ "$OUT_W" -gt "$MAX_WIDTH" && "$MAX_WIDTH" -gt 0 ]]; then
  magick "$TMP" -filter Lanczos -resize "${MAX_WIDTH}x" "$TMP" \
    || { rm -f "$TMP"; _die "auto-resize to ${MAX_WIDTH}px failed"; }
fi

mkdir -p "$(dirname "$OUTPUT")"
mv -f "$TMP" "$OUTPUT"
printf '%s\n' "$OUTPUT"
