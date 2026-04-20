#!/usr/bin/env bash
# mod-wallpaper.sh — hg wallpaper module
# Wraps wallpaper-cycle.sh (static via swww) and shader-wallpaper.sh (animated GLSL)

_WP_CYCLE="$HG_DOTFILES/scripts/wallpaper-cycle.sh"
_WP_SHADER="$HG_DOTFILES/scripts/shader-wallpaper.sh"
_WP_STATE_STATIC="${XDG_STATE_HOME:-$HOME/.local/state}/swww/current"
_WP_STATE_SHADER="${XDG_STATE_HOME:-$HOME/.local/state}/shader-wallpaper/current"

wallpaper_description() {
  echo "Static & shader wallpapers — swww + shaderbg"
}

wallpaper_commands() {
  cat <<'CMDS'
current	Show current wallpaper (static or shader)
next	Next wallpaper in rotation
random	Random wallpaper
set	Set wallpaper by path
shader	Animated shader wallpaper [next|random|set|stop|list]
static	Switch from shader to static wallpaper
engine	Steam Wallpaper Engine scene (linux-wallpaperengine)
CMDS
}

_wallpaper_current() {
  if pgrep -f shaderbg &>/dev/null; then
    local shader
    shader="$(cat "$_WP_STATE_SHADER" 2>/dev/null || echo "unknown")"
    printf "%sshader%s %s%s%s\n" "$HG_MAGENTA" "$HG_RESET" "$HG_CYAN" "$(basename "$shader" .frag)" "$HG_RESET"
  else
    local wp
    wp="$(cat "$_WP_STATE_STATIC" 2>/dev/null || echo "unknown")"
    printf "%sstatic%s %s%s%s\n" "$HG_GREEN" "$HG_RESET" "$HG_CYAN" "$(basename "$wp")" "$HG_RESET"
  fi
}

_wallpaper_set() {
  local path="${1:-}"
  [[ -n "$path" ]] || hg_die "Usage: hg wallpaper set <path>"
  [[ -f "$path" ]] || hg_die "File not found: $path"
  exec bash "$_WP_CYCLE" set "$path"
}

wallpaper_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    current) _wallpaper_current ;;
    next)    exec bash "$_WP_CYCLE" next ;;
    random)  exec bash "$_WP_CYCLE" random ;;
    set)     _wallpaper_set "$@" ;;
    shader)  exec bash "$_WP_SHADER" "${1:-next}" "${@:2}" ;;
    static)  exec bash "$_WP_SHADER" static ;;
    engine)  exec bash "$HG_DOTFILES/scripts/wallpaper-mode.sh" engine ;;
    *)       hg_die "Unknown wallpaper command: $cmd. Run 'hg wallpaper --help'." ;;
  esac
}
