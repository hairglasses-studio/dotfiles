#!/usr/bin/env bash
# capture-window-gif.sh — Record a Hyprland window or region as a GIF
#
# Usage:
#   capture-window-gif.sh <title-pattern|"ticker"> [seconds]
#
#   title-pattern   Case-insensitive window title substring to match
#   ticker          Shortcut: captures the bottom 56px of DP-3 (keybind-ticker)
#   seconds         Recording duration (default: 3)
#
# Output: path to the generated GIF on stdout (all other output goes to stderr)
#
# Dependencies: wf-recorder, ffmpeg, jq, hyprctl
set -euo pipefail

# ---------------------------------------------------------------------------
# Args
# ---------------------------------------------------------------------------
PATTERN="${1:-}"
DURATION="${2:-3}"

if [[ -z "$PATTERN" ]]; then
    echo "Usage: capture-window-gif.sh <title-pattern|\"ticker\"> [seconds]" >&2
    exit 1
fi

if ! [[ "$DURATION" =~ ^[0-9]+$ ]]; then
    echo "Error: duration must be a positive integer, got: $DURATION" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Dependency checks
# ---------------------------------------------------------------------------
for cmd in wf-recorder ffmpeg jq hyprctl; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "Error: $cmd is required but not installed" >&2
        exit 1
    fi
done

# ---------------------------------------------------------------------------
# Resolve geometry
# ---------------------------------------------------------------------------
# wf-recorder -g expects PHYSICAL pixel coords: "X,Y WxH"
# hyprctl clients -j returns LOGICAL coords (pre-scale).
# DP-3 scale is 2, so logical * 2 = physical.

SCALE=2  # DP-3 scale factor

if [[ "$PATTERN" == "ticker" ]]; then
    # Keybind-ticker: pinned at logical 0,692 with size 2560x28 on DP-3.
    # Physical: 0,1384 5120x56
    GEOMETRY="0,1384 5120x56"
    LABEL="ticker"
else
    # Find window by title pattern
    WINDOW_JSON=$(hyprctl clients -j 2>/dev/null | jq -r --arg pat "$PATTERN" \
        '[.[] | select(.title | ascii_downcase | contains($pat | ascii_downcase))] | first // empty')

    if [[ -z "$WINDOW_JSON" ]]; then
        echo "Error: no window found matching title pattern: $PATTERN" >&2
        echo "Available windows:" >&2
        hyprctl clients -j 2>/dev/null | jq -r '.[].title' >&2
        exit 1
    fi

    # Extract logical coordinates and multiply by scale
    LX=$(echo "$WINDOW_JSON" | jq -r '.at[0]')
    LY=$(echo "$WINDOW_JSON" | jq -r '.at[1]')
    LW=$(echo "$WINDOW_JSON" | jq -r '.size[0]')
    LH=$(echo "$WINDOW_JSON" | jq -r '.size[1]')

    # Convert to physical pixels
    PX=$(( LX * SCALE ))
    PY=$(( LY * SCALE ))
    PW=$(( LW * SCALE ))
    PH=$(( LH * SCALE ))

    GEOMETRY="${PX},${PY} ${PW}x${PH}"
    LABEL=$(echo "$WINDOW_JSON" | jq -r '.title' | tr ' /' '_' | tr -cd '[:alnum:]_-' | head -c 32)
    [[ -z "$LABEL" ]] && LABEL="window"
fi

# ---------------------------------------------------------------------------
# Output paths
# ---------------------------------------------------------------------------
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TMP_MP4=$(mktemp "/tmp/capture-${LABEL}-${TIMESTAMP}-XXXXXX.mp4")
OUT_GIF="/tmp/capture-${LABEL}-${TIMESTAMP}.gif"

# Clean up MP4 on exit (keep GIF)
trap 'rm -f "$TMP_MP4"' EXIT

# ---------------------------------------------------------------------------
# Record
# ---------------------------------------------------------------------------
echo "Recording: geometry=${GEOMETRY}, duration=${DURATION}s" >&2

wf-recorder -g "$GEOMETRY" -f "$TMP_MP4" &
WF_PID=$!

sleep "$DURATION"
kill "$WF_PID" 2>/dev/null || true
wait "$WF_PID" 2>/dev/null || true

if [[ ! -s "$TMP_MP4" ]]; then
    echo "Error: recording produced an empty file — wf-recorder may have failed" >&2
    exit 1
fi

# ---------------------------------------------------------------------------
# Convert to GIF
# ---------------------------------------------------------------------------
echo "Converting to GIF..." >&2

ffmpeg -y -i "$TMP_MP4" \
    -vf "fps=30,scale=iw:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
    "$OUT_GIF" 2>/dev/null

if [[ ! -s "$OUT_GIF" ]]; then
    echo "Error: ffmpeg GIF conversion failed" >&2
    exit 1
fi

GIF_SIZE=$(du -sh "$OUT_GIF" 2>/dev/null | cut -f1)
echo "Done: ${GIF_SIZE} — ${OUT_GIF}" >&2

# Output just the path on stdout for scripting
echo "$OUT_GIF"
