#!/usr/bin/env bash
# shader-wallpaper.sh — Procgen shader wallpaper engine via shaderbg
# Usage: shader-wallpaper.sh [next|random|set <shader>|stop|list|static]
#
# Runs Shadertoy-compatible GLSL shaders as live animated wallpapers.
# Falls back to swww for static image wallpapers.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"
source "$SCRIPT_DIR/lib/notify.sh"

SHADER_DIR="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}/wallpaper-shaders"
STATE_FILE="${XDG_STATE_HOME:-$HOME/.local/state}/shader-wallpaper/current"
FPS="${SHADER_WALLPAPER_FPS:-30}"

mkdir -p "$(dirname "$STATE_FILE")"

# ── Pre-flight: require shaderbg ──────────────
command -v shaderbg &>/dev/null || { echo "error: shaderbg not found in PATH" >&2; exit 1; }

_get_output() {
  compositor_output
}

_get_shaders() {
  find "$SHADER_DIR" -maxdepth 1 -type f -name '*.frag' 2>/dev/null | sort
}

_stop() {
  pkill -f 'shaderbg' 2>/dev/null
}

_set_shader() {
  local shader="$1"
  [[ -f "$shader" ]] || { echo "Shader not found: $shader"; return 1; }

  local output
  output=$(_get_output)
  [[ -z "$output" ]] && { echo "No output detected"; return 1; }

  # Kill any running shaderbg
  _stop
  sleep 0.2

  # Kill swww-daemon if running (shader takes over wallpaper layer)
  # Don't kill — shaderbg uses wlr-layer-shell, not the wallpaper protocol

  # Launch shaderbg in background
  shaderbg --fps "$FPS" "$output" "$shader" &
  disown

  echo "$shader" > "$STATE_FILE"
  hg_notify_low "Wallpaper" "$(basename "$shader" .frag)"

  if command -v tte &>/dev/null && [[ -t 1 ]]; then
    echo "SHADER // $(basename "$shader" .frag)" | tte beams \
      --beam-delay 2 --beam-gradient-stops 57c7ff ff6ac1 \
      --final-gradient-stops 57c7ff ff6ac1 2>/dev/null
  else
    echo "Shader: $(basename "$shader" .frag)"
  fi
}

case "${1:-next}" in
  next)
    mapfile -t shaders < <(_get_shaders)
    [[ ${#shaders[@]} -eq 0 ]] && { echo "No shaders in $SHADER_DIR"; exit 1; }
    current=$(cat "$STATE_FILE" 2>/dev/null || echo "")
    idx=0
    for i in "${!shaders[@]}"; do
      [[ "${shaders[$i]}" == "$current" ]] && { idx=$(( (i + 1) % ${#shaders[@]} )); break; }
    done
    _set_shader "${shaders[$idx]}"
    ;;
  random)
    mapfile -t shaders < <(_get_shaders)
    [[ ${#shaders[@]} -eq 0 ]] && { echo "No shaders in $SHADER_DIR"; exit 1; }
    _set_shader "${shaders[$((RANDOM % ${#shaders[@]}))]}"
    ;;
  set)
    [[ -n "${2:-}" ]] || { echo "Usage: shader-wallpaper.sh set <path>"; exit 1; }
    _set_shader "$2"
    ;;
  stop)
    _stop
    echo "Shader wallpaper stopped"
    ;;
  static)
    # Stop shader, switch back to swww static wallpaper
    _stop
    shift
    wallpaper-cycle.sh "${1:-random}"
    ;;
  list)
    _get_shaders | while read -r f; do
      basename "$f" .frag
    done
    ;;
  *)
    echo "Usage: shader-wallpaper.sh [next|random|set <shader>|stop|list|static]"
    exit 1
    ;;
esac
