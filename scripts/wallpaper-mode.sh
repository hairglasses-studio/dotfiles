#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

STATE_DIR="${XDG_STATE_HOME:-$HOME/.local/state}/wallpaper-mode"
MODE_FILE="$STATE_DIR/mode"
VALUE_FILE="$STATE_DIR/value"

mkdir -p "$STATE_DIR"

if [[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/hypr/wallpaper.env" ]]; then
  # shellcheck disable=SC1090
  source "${XDG_CONFIG_HOME:-$HOME/.config}/hypr/wallpaper.env"
fi

WALLPAPER_DIR="${WALLPAPER_DIR:-$HOME/Pictures/wallpapers}"
WALLPAPER_VIDEO="${WALLPAPER_VIDEO:-$HOME/Pictures/wallpapers/live/default.mp4}"
WALLPAPER_DEPTH="${WALLPAPER_DEPTH:-$HOME/Pictures/wallpapers/depth/default.jpg}"

_set_state() {
  printf '%s\n' "$1" > "$MODE_FILE"
  printf '%s\n' "${2:-}" > "$VALUE_FILE"
}

_state_mode() {
  cat "$MODE_FILE" 2>/dev/null || true
}

_state_value() {
  cat "$VALUE_FILE" 2>/dev/null || true
}

_notify_missing() {
  local thing="$1"
  hg_notify_critical "Wallpaper" "$thing is not installed"
  hg_die "$thing is required"
}

_notify_missing_file() {
  local label="$1" path="$2"
  hg_notify_critical "Wallpaper" "$label not found: $path"
  hg_die "$label not found: $path"
}

_stop_dynamic_wallpapers() {
  pkill -f shaderbg 2>/dev/null || true
  pkill -x mpvpaper 2>/dev/null || true
  if command -v waydeeper >/dev/null 2>&1; then
    waydeeper stop >/dev/null 2>&1 || true
  fi
  pkill -x waydeeper 2>/dev/null || true
  pkill -x hyprlax 2>/dev/null || true
  pkill -x glshell 2>/dev/null || true
}

_ensure_swww() {
  command -v swww-daemon >/dev/null 2>&1 || _notify_missing "swww-daemon"
  if ! pgrep -x swww-daemon >/dev/null 2>&1; then
    swww-daemon >/dev/null 2>&1 &
    disown
    sleep 0.3
  fi
}

_switch_shader() {
  command -v shader-wallpaper.sh >/dev/null 2>&1 || _notify_missing "shader-wallpaper.sh"
  local action="${1:-random}" value="${2:-}"
  _stop_dynamic_wallpapers

  case "$action" in
    next|random)
      shader-wallpaper.sh "$action"
      _set_state shader "$action"
      ;;
    set)
      [[ -n "$value" ]] || hg_die "shader set requires a shader path"
      shader-wallpaper.sh set "$value"
      _set_state shader "$value"
      ;;
    *)
      hg_die "Unknown shader action: $action"
      ;;
  esac
}

_switch_static() {
  command -v wallpaper-cycle.sh >/dev/null 2>&1 || _notify_missing "wallpaper-cycle.sh"
  local action="${1:-random}" value="${2:-}"
  _ensure_swww
  _stop_dynamic_wallpapers

  case "$action" in
    next|random)
      wallpaper-cycle.sh "$action"
      _set_state static "$action"
      ;;
    set)
      [[ -n "$value" ]] || hg_die "static set requires a wallpaper path"
      wallpaper-cycle.sh set "$value"
      _set_state static "$value"
      ;;
    *)
      hg_die "Unknown static action: $action"
      ;;
  esac
}

_switch_video() {
  command -v mpvpaper >/dev/null 2>&1 || _notify_missing "mpvpaper"
  local path="${1:-$WALLPAPER_VIDEO}"
  [[ -f "$path" ]] || _notify_missing_file "Video wallpaper" "$path"

  _stop_dynamic_wallpapers
  mpvpaper -o "no-audio --loop-playlist=inf --hwdec=auto-safe" ALL "$path" >/dev/null 2>&1 &
  disown

  _set_state video "$path"
  hg_notify_low "Wallpaper" "Video mode: $(basename "$path")"
}

_switch_depth() {
  command -v waydeeper >/dev/null 2>&1 || _notify_missing "waydeeper"
  local path="${1:-$WALLPAPER_DEPTH}"
  [[ -f "$path" ]] || _notify_missing_file "Depth wallpaper" "$path"

  _stop_dynamic_wallpapers
  if ! pgrep -x waydeeper >/dev/null 2>&1; then
    waydeeper daemon >/dev/null 2>&1 &
    disown
    sleep 0.5
  fi
  waydeeper set "$path" >/dev/null 2>&1

  _set_state depth "$path"
  hg_notify_low "Wallpaper" "Depth mode: $(basename "$path")"
}

_switch_parallax() {
  command -v hyprlax >/dev/null 2>&1 || _notify_missing "hyprlax"
  local path="${1:-}"
  _ensure_swww
  _stop_dynamic_wallpapers

  if [[ -n "$path" && -f "$path" ]]; then
    swww img "$path" --transition-type fade --transition-duration 1 >/dev/null 2>&1
  fi

  hyprlax >/dev/null 2>&1 &
  disown

  _set_state parallax "${path:-active}"
  hg_notify_low "Wallpaper" "Parallax mode${path:+: $(basename "$path")}"
}

_restore() {
  local mode value
  mode="$(_state_mode)"
  value="$(_state_value)"

  if [[ -z "$mode" ]]; then
    _switch_shader random
    return 0
  fi

  case "$mode" in
    shader)
      if [[ "$value" == "next" || "$value" == "random" ]]; then
        _switch_shader "$value"
      elif [[ -n "$value" ]]; then
        _switch_shader set "$value"
      else
        _switch_shader random
      fi
      ;;
    static)
      if [[ "$value" == "next" || "$value" == "random" ]]; then
        _switch_static "$value"
      elif [[ -n "$value" ]]; then
        _switch_static set "$value"
      else
        _switch_static random
      fi
      ;;
    video)
      _switch_video "$value"
      ;;
    depth)
      _switch_depth "$value"
      ;;
    parallax)
      _switch_parallax "$value"
      ;;
    papertoy)
      # Legacy mode — papertoy was removed (see deduplication in favor of shaderbg).
      # Fall back to shader mode with the same value.
      _switch_shader set "$value"
      ;;
    *)
      _switch_shader random
      ;;
  esac
}

_status() {
  local mode value
  mode="$(_state_mode)"
  value="$(_state_value)"

  if [[ -z "$mode" ]]; then
    printf 'unknown\n'
    return 0
  fi

  case "$mode" in
    video|depth|parallax)
      printf '%s:%s\n' "$mode" "$(basename "${value:-unknown}")"
      ;;
    shader|static)
      printf '%s:%s\n' "$mode" "${value:-unknown}"
      ;;
    *)
      printf '%s\n' "$mode"
      ;;
  esac
}

main() {
  local mode="${1:-restore}"
  shift || true

  case "$mode" in
    shader)   _switch_shader "${1:-random}" "${2:-}" ;;
    static)   _switch_static "${1:-random}" "${2:-}" ;;
    video)    _switch_video "${1:-}" ;;
    depth)    _switch_depth "${1:-}" ;;
    parallax) _switch_parallax "${1:-}" ;;
    papertoy)
      # Legacy alias — fall through to shader mode. papertoy was deduplicated
      # in favor of shaderbg (Phase 3 consolidation).
      hg_warn "papertoy mode is deprecated; using shader mode"
      _switch_shader "${1:-random}"
      ;;
    stop)     _stop_dynamic_wallpapers ;;
    restore)
      if ! _restore; then
        hg_warn "Wallpaper restore failed, falling back to shader random"
        _switch_shader random
      fi
      ;;
    status) _status ;;
    *)
      hg_die "Usage: wallpaper-mode.sh [shader|static|video|depth|parallax|restore|status|stop]"
      ;;
  esac
}

main "$@"
