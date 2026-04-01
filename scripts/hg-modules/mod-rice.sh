#!/usr/bin/env bash
# mod-rice.sh — hg rice module
# Rice status dashboard, service health, palette compliance

source "$HG_DOTFILES/scripts/lib/compositor.sh" 2>/dev/null
source "$HG_DOTFILES/scripts/lib/config.sh" 2>/dev/null

# Snazzy palette allowed hex values (lowercase, no #)
_SNAZZY_ALLOWED="57c7ff|ff6ac1|5af78e|f3f99d|ff5c57|686868|9aedfe|eff0eb|f1f1f0|000000|1a1a1a|1a1b26"

rice_description() {
  echo "Rice health — status, services, palette, reload-all"
}

rice_commands() {
  cat <<'CMDS'
status	Comprehensive rice status dashboard
services	Check which services are running
palette	Scan configs for non-Snazzy colors
reload-all	Reload all compositor services
CMDS
}

_rice_status() {
  printf "\n %s%srice status%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Compositor
  local comp
  comp="$(compositor_type 2>/dev/null || echo "unknown")"
  printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "compositor" "$HG_RESET" "$HG_CYAN" "$comp" "$HG_RESET"

  # Active shader
  local shader_line shader_name
  shader_line="$(grep -m1 '^custom-shader = ' "$HOME/.config/ghostty/config" 2>/dev/null || true)"
  if [[ -n "$shader_line" && "$shader_line" != *"none"* ]]; then
    shader_name="$(basename "${shader_line#custom-shader = }" .glsl)"
    printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "shader" "$HG_RESET" "$HG_MAGENTA" "$shader_name" "$HG_RESET"
  else
    printf " %s%-14s%s %snone%s\n" "$HG_DIM" "shader" "$HG_RESET" "$HG_DIM" "$HG_RESET"
  fi

  # Wallpaper
  if pgrep -f shaderbg &>/dev/null; then
    local wp_shader
    wp_shader="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/shader-wallpaper/current" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %sshader:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_MAGENTA" "$HG_RESET" "$(basename "$wp_shader" .frag)"
  else
    local wp_static
    wp_static="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/swww/current" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %sstatic:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_GREEN" "$HG_RESET" "$(basename "$wp_static")"
  fi

  # Bar
  if pgrep -x eww &>/dev/null; then
    printf " %s%-14s%s %seww%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  elif pgrep -x waybar &>/dev/null; then
    printf " %s%-14s%s %swaybar%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf " %s%-14s%s %snone%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # Services summary
  printf "\n"
  _rice_services
}

_rice_services() {
  printf " %sSERVICES%s\n" "$HG_BOLD" "$HG_RESET"

  local -a services=(
    "Hyprland:Hyprland"
    "sway:sway"
    "eww:eww"
    "mako:mako"
    "swww:swww-daemon"
    "hypridle:hypridle"
    "waybar:waybar"
    "antimicrox:antimicrox"
    "hyprsunset:hyprsunset"
  )

  for entry in "${services[@]}"; do
    local name="${entry%%:*}" proc="${entry#*:}"
    if pgrep -x "$proc" &>/dev/null; then
      printf "  %s%-14s%s %srunning%s\n" "$HG_CYAN" "$name" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-14s%s %sstopped%s\n" "$HG_CYAN" "$name" "$HG_RESET" "$HG_DIM" "$HG_RESET"
    fi
  done
  printf "\n"
}

_rice_palette() {
  printf "\n %s%spalette scan%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  local scan_dirs=(
    "$HG_DOTFILES/hyprland"
    "$HG_DOTFILES/eww"
    "$HG_DOTFILES/mako"
    "$HG_DOTFILES/wofi"
    "$HG_DOTFILES/wlogout"
    "$HG_DOTFILES/waybar"
    "$HG_DOTFILES/foot"
  )

  local violations=0
  for dir in "${scan_dirs[@]}"; do
    [[ -d "$dir" ]] || continue
    # Find hex colors, exclude allowed palette
    while IFS=: read -r file line content; do
      # Extract hex colors from the line
      local colors
      colors="$(echo "$content" | grep -oiE '#[0-9a-fA-F]{6}' | tr '[:upper:]' '[:lower:]' | sed 's/#//')"
      for color in $colors; do
        if ! echo "$color" | grep -qiE "^($_SNAZZY_ALLOWED)$"; then
          printf "  %s%s%s:%s%s%s #%s%s%s\n" "$HG_DIM" "$(basename "$file")" "$HG_RESET" "$HG_DIM" "$line" "$HG_RESET" "$HG_RED" "$color" "$HG_RESET"
          violations=$((violations + 1))
        fi
      done
    done < <(grep -rnE '#[0-9a-fA-F]{6}' "$dir" 2>/dev/null || true)
  done

  if [[ $violations -eq 0 ]]; then
    hg_ok "All colors are Snazzy-compliant"
  else
    hg_warn "$violations non-Snazzy color(s) found"
  fi
  printf "\n"
}

_rice_reload_all() {
  hg_info "Reloading all services..."
  config_reload_service hyprland
  config_reload_service mako
  config_reload_service eww
  config_reload_service waybar
  hg_ok "All services reloaded"
}

rice_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)     _rice_status ;;
    services)   _rice_services ;;
    palette)    _rice_palette ;;
    reload-all) _rice_reload_all ;;
    *)          hg_die "Unknown rice command: $cmd. Run 'hg rice --help'." ;;
  esac
}
