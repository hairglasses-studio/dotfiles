#!/usr/bin/env bash
# kitty-font-wheel.sh — scale focused Kitty text size from thumb-wheel input
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"
source "$SCRIPT_DIR/lib/kitty-config.sh"

direction="${1:-}"
config_file="${XDG_CONFIG_HOME:-$HOME/.config}/juhradial/config.json"
state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/juhradial"
state_file="$state_dir/kitty-font-size.state"

active_window_json() {
  compositor_query activewindow 2>/dev/null
}

active_window_class() {
  active_window_json | jq -r '.class // ""' 2>/dev/null | tr '[:upper:]' '[:lower:]'
}

state_key() {
  active_window_json | jq -r '.address // "kitty"' 2>/dev/null
}

read_cfg_number() {
  local jq_expr="$1"
  local fallback="$2"
  if [[ -f "$config_file" ]] && command -v jq >/dev/null 2>&1; then
    jq -r "$jq_expr // $fallback" "$config_file" 2>/dev/null || printf '%s\n' "$fallback"
    return 0
  fi

  printf '%s\n' "$fallback"
}

read_state_size() {
  local key="$1"
  [[ -f "$state_file" ]] || return 1
  awk -v key="$key" '$1 == key { print $2; found=1 } END { exit found ? 0 : 1 }' "$state_file"
}

write_state_size() {
  local key="$1"
  local value="$2"
  local tmp found=false

  mkdir -p "$state_dir"
  tmp="$(mktemp)"

  if [[ -f "$state_file" ]]; then
    while read -r existing_key existing_value; do
      [[ -n "${existing_key:-}" ]] || continue
      if [[ "$existing_key" == "$key" ]]; then
        printf '%s %s\n' "$key" "$value" >>"$tmp"
        found=true
      else
        printf '%s %s\n' "$existing_key" "$existing_value" >>"$tmp"
      fi
    done <"$state_file"
  fi

  if ! $found; then
    printf '%s %s\n' "$key" "$value" >>"$tmp"
  fi

  mv "$tmp" "$state_file"
}

trim_number() {
  awk -v value="$1" 'BEGIN {
    formatted = sprintf("%.2f", value)
    sub(/0+$/, "", formatted)
    sub(/\.$/, "", formatted)
    print formatted
  }'
}

ydotool_socket_path() {
  printf '%s\n' "${YDOTOOL_SOCKET:-${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/.ydotool_socket}"
}

send_zoom_key() {
  command -v ydotool >/dev/null 2>&1 || return 1
  case "$1" in
    increase) YDOTOOL_SOCKET="$(ydotool_socket_path)" ydotool key 29:1 42:1 13:1 13:0 42:0 29:0 ;;
    decrease) YDOTOOL_SOCKET="$(ydotool_socket_path)" ydotool key 29:1 42:1 12:1 12:0 42:0 29:0 ;;
    *) return 1 ;;
  esac
}

remote_set_size() {
  local size="$1"
  kitty @ --to unix:@mykitty set-font-size "$size" >/dev/null 2>&1
}

[[ "$(active_window_class)" == "kitty" ]] || exit 0

default_size="$(kitty_get_font_size)"
[[ -n "${default_size:-}" ]] || default_size="7.8"

step="$(read_cfg_number '.thumbwheel.step' '0.5')"
min_size="$(read_cfg_number '.thumbwheel.min_font_size' '6.0')"
max_size="$(read_cfg_number '.thumbwheel.max_font_size' '12.0')"

key="$(state_key)"
current_size="$(read_state_size "$key" || true)"
[[ -n "${current_size:-}" ]] || current_size="$default_size"

case "$direction" in
  right)
    target_size="$(awk -v cur="$current_size" -v step="$step" 'BEGIN { print cur + step }')"
    fallback_action="increase"
    ;;
  left)
    target_size="$(awk -v cur="$current_size" -v step="$step" 'BEGIN { print cur - step }')"
    fallback_action="decrease"
    ;;
  *)
    printf 'Usage: %s {left|right}\n' "$(basename "$0")" >&2
    exit 2
    ;;
esac

target_size="$(awk -v target="$target_size" -v min="$min_size" -v max="$max_size" 'BEGIN {
  if (target < min) target = min
  if (target > max) target = max
  print target
}')"

if [[ "$(trim_number "$target_size")" == "$(trim_number "$current_size")" ]]; then
  exit 0
fi

target_size="$(trim_number "$target_size")"

if remote_set_size "$target_size"; then
  write_state_size "$key" "$target_size"
  exit 0
fi

send_zoom_key "$fallback_action"
write_state_size "$key" "$target_size"
