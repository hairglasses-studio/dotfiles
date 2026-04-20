#!/usr/bin/env bash
# ticker-record.sh — record only the ticker strip as an MP4.
#
# Mirrors scripts/ticker-shot.sh but produces an H.264 video via
# wf-recorder. Captures exactly the layer-shell region so the output
# contains nothing but the scrolling bar — no wallpaper, no other UI.
#
# Usage:
#   ticker-record.sh                  60 s → ~/Videos/recordings/ticker-YYYYMMDD_HHMMSS.mp4
#   ticker-record.sh 30               30 s recording
#   ticker-record.sh --pin calendar   pin stream, record, auto-unpin (+ playlist switch if needed)
#   ticker-record.sh --output foo.mp4 explicit path
#   ticker-record.sh --monitor DP-3   target a different ticker instance
#   ticker-record.sh --codec libx265  swap codec (default: libx264)
#   ticker-record.sh --audio          include default-source audio
#   ticker-record.sh --height 28      strip height in px (default 39)
#   ticker-record.sh --no-switch      fail instead of auto-switching playlist for a pin
#   ticker-record.sh --print-geom     just print the capture region and exit
#
# Exits non-zero on failure. On success prints the absolute output path.

set -euo pipefail

DURATION=60
MONITOR="DP-2"
OUTPUT=""
PIN=""
NO_SWITCH=0
CODEC=""
AUDIO=""
HEIGHT=39
PRINT_GEOM=0

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
PLAYLISTS_DIR="$(dirname "$SCRIPT_DIR")/ticker/content-playlists"
REC_DIR="$HOME/Videos/recordings"

_die() { printf 'ticker-record: %s\n' "$*" >&2; exit 1; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    -d|--duration) DURATION="${2:-}"; shift 2 ;;
    --monitor)     MONITOR="${2:-}"; shift 2 ;;
    --output|-o)   OUTPUT="${2:-}"; shift 2 ;;
    --pin)         PIN="${2:-}"; shift 2 ;;
    --no-switch)   NO_SWITCH=1; shift ;;
    --codec|-c)    CODEC="${2:-}"; shift 2 ;;
    --audio)       AUDIO="--audio"; shift ;;
    --height)      HEIGHT="${2:-}"; shift 2 ;;
    --print-geom)  PRINT_GEOM=1; shift ;;
    -h|--help)
      sed -n '2,22p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    [0-9]*)        DURATION="$1"; shift ;;
    *) _die "unknown flag: $1 (try --help)" ;;
  esac
done

command -v wf-recorder >/dev/null 2>&1 || _die "wf-recorder not installed"
command -v hyprctl     >/dev/null 2>&1 || _die "hyprctl not found (Hyprland not running?)"

# Resolve geometry. hyprctl reports `x`/`y` in LOGICAL coords but
# `width`/`height` in PHYSICAL pixels on HiDPI outputs (scale > 1); divide
# by scale so grim-style geometry strings stay consistent.
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

if [[ -z "$OUTPUT" ]]; then
  mkdir -p "$REC_DIR"
  OUTPUT="$REC_DIR/ticker-$(date +%Y%m%d_%H%M%S).mp4"
fi

# ── Optional: pin + auto-switch playlist ─────────────────────────────
# Mirrors ticker-shot.sh: save prior pin/playlist, apply requested pin,
# auto-switch playlist if the stream isn't in the active one, restore
# both on EXIT so an interrupted recording never strands the ticker
# stuck on a pinned stream.
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

  ACTIVE_PLAYLIST_FILE="$PLAYLISTS_DIR/${PREV_PLAYLIST}.txt"
  if [[ -r "$ACTIVE_PLAYLIST_FILE" ]] && ! grep -qx -- "$PIN" "$ACTIVE_PLAYLIST_FILE"; then
    if [[ "$NO_SWITCH" == "1" ]]; then
      _die "stream '$PIN' not in playlist '$PREV_PLAYLIST' (rerun without --no-switch)"
    fi
    TARGET_PLAYLIST="$(_find_playlist_with_stream "$PIN" || true)"
    [[ -n "$TARGET_PLAYLIST" ]] || _die "stream '$PIN' not found in any playlist under $PLAYLISTS_DIR"
    printf '%s' "$TARGET_PLAYLIST" > "$STATE_DIR/active-playlist"
    DID_PLAYLIST_SWITCH=1
    _send_usr1
    sleep 0.45
  fi

  printf '%s' "$PIN" > "$STATE_DIR/pinned-stream"
  DID_PIN=1
  _send_usr1
  sleep 0.7
fi

# ── Record ──────────────────────────────────────────────────────────
ARGS=(-g "$GEOM" -f "$OUTPUT")
[[ -n "$CODEC" ]] && ARGS+=(-c "$CODEC")
[[ -n "$AUDIO" ]] && ARGS+=("$AUDIO")

mkdir -p "$(dirname "$OUTPUT")"
printf 'ticker-record: %s for %ds → %s\n' "$GEOM" "$DURATION" "$OUTPUT" >&2

# SIGINT lets wf-recorder finalise the MP4 cleanly; SIGTERM would leave
# a truncated, unplayable file.
set +e
timeout -s INT "${DURATION}s" wf-recorder "${ARGS[@]}" 2>/dev/null
rc=$?
set -e
# timeout exits 124 on expiry; wf-recorder exits 0 on SIGINT after finalise.
if [[ "$rc" -ne 0 && "$rc" -ne 124 && "$rc" -ne 130 ]]; then
  _die "wf-recorder exit $rc"
fi

sleep 0.3
if command -v ffprobe >/dev/null 2>&1; then
  ffprobe -v error -show_entries stream=codec_name,width,height,r_frame_rate,duration \
    -of default=noprint_wrappers=1 "$OUTPUT" >&2 2>&1 | head -6 || true
fi

printf '%s\n' "$OUTPUT"
