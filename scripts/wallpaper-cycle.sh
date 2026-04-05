#!/usr/bin/env bash
# wallpaper-cycle.sh — Animated wallpaper rotation via swww
# Usage: wallpaper-cycle.sh [next|random|set <path>]
set -euo pipefail

WALLPAPER_DIR="${WALLPAPER_DIR:-$HOME/Pictures/wallpapers}"
STATE_FILE="${XDG_STATE_HOME:-$HOME/.local/state}/swww/current"

# Transition effects — cyberpunk themed
TRANSITIONS=(
  "--transition-type wipe --transition-angle 30 --transition-step 90"
  "--transition-type wave --transition-angle 120 --transition-step 90"
  "--transition-type grow --transition-pos 0.5,0.9 --transition-step 90"
  "--transition-type outer --transition-pos 0.5,0.5 --transition-step 90"
)

mkdir -p "$(dirname "$STATE_FILE")"

_set_wallpaper() {
  local img="$1"
  [[ -f "$img" ]] || { echo "File not found: $img"; return 1; }

  # Pick a random transition
  local tr="${TRANSITIONS[$((RANDOM % ${#TRANSITIONS[@]}))]}"

  # shellcheck disable=SC2086
  swww img "$img" \
    $tr \
    --transition-duration 2 \
    --transition-fps 60 2>/dev/null

  echo "$img" > "$STATE_FILE"

  if command -v tte &>/dev/null && [[ -t 1 ]]; then
    echo "WALLPAPER // $(basename "$img")" | tte slide \
      --movement-speed 2.0 --grouping row \
      --final-gradient-stops 686868 57c7ff 2>/dev/null
  fi
}

_get_wallpapers() {
  find "$WALLPAPER_DIR" -maxdepth 1 -type f \
    \( -name '*.jpg' -o -name '*.jpeg' -o -name '*.png' -o -name '*.gif' -o -name '*.webp' \) \
    2>/dev/null | sort
}

case "${1:-next}" in
  next)
    mapfile -t walls < <(_get_wallpapers)
    [[ ${#walls[@]} -eq 0 ]] && { echo "No wallpapers in $WALLPAPER_DIR"; exit 1; }
    current=$(cat "$STATE_FILE" 2>/dev/null || echo "")
    idx=0
    for i in "${!walls[@]}"; do
      [[ "${walls[$i]}" == "$current" ]] && { idx=$(( (i + 1) % ${#walls[@]} )); break; }
    done
    _set_wallpaper "${walls[$idx]}"
    ;;
  random)
    mapfile -t walls < <(_get_wallpapers)
    [[ ${#walls[@]} -eq 0 ]] && { echo "No wallpapers in $WALLPAPER_DIR"; exit 1; }
    _set_wallpaper "${walls[$((RANDOM % ${#walls[@]}))]}"
    ;;
  set)
    [[ -n "${2:-}" ]] || { echo "Usage: wallpaper-cycle.sh set <path>"; exit 1; }
    _set_wallpaper "$2"
    ;;
  *)
    echo "Usage: wallpaper-cycle.sh [next|random|set <path>]"
    exit 1
    ;;
esac
