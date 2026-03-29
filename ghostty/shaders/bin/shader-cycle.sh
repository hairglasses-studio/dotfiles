#!/usr/bin/env bash
# shader-cycle — cycle through a curated shader list + Tattoy mode
# Bound to a global keybind (via AeroSpace). Each press advances one step.
# Ghostty auto-reloads config via FSEvents; Tattoy watches its config file.
#
# Cycle entries:
#   *.glsl  → set as Ghostty custom-shader, disable Tattoy shader layer
#   tattoy  → clear Ghostty shader, enable Tattoy shader+cursor layers
#
# Usage:
#   shader-cycle           # advance to next
#   shader-cycle --prev    # go to previous
#   shader-cycle --show    # print current entry without advancing

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADERS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
STATE_DIR="$HOME/.local/state/ghostty"
CYCLE_FILE="$STATE_DIR/cycle-index"
GHOSTTY_CONFIG="$HOME/.config/ghostty/config"
TATTOY_CONFIG="$HOME/Library/Application Support/tattoy/tattoy.toml"

# ── Cycle list ──────────────────────────────────
# Edit this array to change the cycle order.
# "tattoy" is a special entry that activates Tattoy's shader layers.
CYCLE=(
  bloom-soft.glsl
  crt-chromatic.glsl
  starfield-colors.glsl
  underwater.glsl
  halftone.glsl
  old-film.glsl
  auroras.glsl
  creation.glsl
  cyberpunk.glsl
  vaporwave.glsl
  tattoy
)

CYCLE_LEN=${#CYCLE[@]}

mkdir -p "$STATE_DIR" 2>/dev/null

# ── Read current index ──────────────────────────
idx=0
if [[ -f "$CYCLE_FILE" ]]; then
  idx="$(< "$CYCLE_FILE")"
  [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
fi

# ── Parse args ──────────────────────────────────
case "${1:-}" in
  --show)
    echo "${CYCLE[$idx]}"
    exit 0
    ;;
  --prev)
    idx=$(( (idx - 1 + CYCLE_LEN) % CYCLE_LEN ))
    ;;
  *)
    idx=$(( (idx + 1) % CYCLE_LEN ))
    ;;
esac

# ── Save new index ──────────────────────────────
printf '%s' "$idx" > "$CYCLE_FILE"

entry="${CYCLE[$idx]}"

# ── Check if shader needs animation ─────────────
needs_animation() {
  grep -qE '(ghostty_time|iTime|u_time)' "$1" 2>/dev/null
}

# ── Apply: Ghostty shader ───────────────────────
apply_ghostty_shader() {
  local shader_path="$SHADERS_DIR/$1"
  if [[ ! -f "$shader_path" ]]; then
    echo "Not found: $shader_path" >&2
    return 1
  fi

  local anim="false"
  needs_animation "$shader_path" && anim="true"

  local tmp
  tmp="$(mktemp "${GHOSTTY_CONFIG}.XXXXXX")"
  sed -e "s|^#* *custom-shader = .*|custom-shader = $shader_path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
      "$GHOSTTY_CONFIG" > "$tmp"
  mv -f "$tmp" "$GHOSTTY_CONFIG"
}

# ── Apply: disable Ghostty shader ───────────────
clear_ghostty_shader() {
  local tmp
  tmp="$(mktemp "${GHOSTTY_CONFIG}.XXXXXX")"
  sed -e "s|^#* *custom-shader = .*|# custom-shader = none|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = false|" \
      "$GHOSTTY_CONFIG" > "$tmp"
  mv -f "$tmp" "$GHOSTTY_CONFIG"
}

# ── Apply: toggle Tattoy shader layer ───────────
set_tattoy_shader() {
  local enabled="$1"  # true or false
  [[ -f "$TATTOY_CONFIG" ]] || return 0

  local tmp
  tmp="$(mktemp "${TATTOY_CONFIG}.XXXXXX")"
  sed -e "/^\[shader\]/,/^\[/ s|^enabled = .*|enabled = $enabled|" \
      -e "/^\[animated_cursor\]/,/^\[/ s|^enabled = .*|enabled = $enabled|" \
      "$TATTOY_CONFIG" > "$tmp"
  mv -f "$tmp" "$TATTOY_CONFIG"
}

# ── Dispatch ────────────────────────────────────
if [[ "$entry" == "tattoo" || "$entry" == "tattoy" ]]; then
  clear_ghostty_shader
  set_tattoy_shader "true"
  echo "→ tattoy (shader + cursor layers)"
else
  apply_ghostty_shader "$entry"
  set_tattoy_shader "false"
  echo "→ $entry"
fi
