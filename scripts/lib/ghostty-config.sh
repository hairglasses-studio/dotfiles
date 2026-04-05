#!/usr/bin/env bash
# ghostty-config.sh — Shared Ghostty config query functions
# Source this file: source "$(dirname "$0")/lib/ghostty-config.sh"

_GHOSTTY_CONFIG="${GHOSTTY_CONFIG:-$HOME/.config/ghostty/config}"

# Get the current custom-shader path (raw value from config)
ghostty_get_shader_path() {
  grep -m1 '^custom-shader = ' "$_GHOSTTY_CONFIG" 2>/dev/null | sed 's/^custom-shader = //'
}

# Get the current shader name (basename without .glsl)
ghostty_get_shader_name() {
  local path
  path="$(ghostty_get_shader_path)"
  [[ -n "$path" && "$path" != "none" ]] && basename "$path" .glsl || echo ""
}

# Get the current custom-shader-animation value (true/false)
ghostty_get_shader_animation() {
  grep -m1 '^custom-shader-animation = ' "$_GHOSTTY_CONFIG" 2>/dev/null | sed 's/^custom-shader-animation = //' || echo "false"
}
