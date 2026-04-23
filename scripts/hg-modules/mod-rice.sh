#!/usr/bin/env bash
# mod-rice.sh — hg rice module
# Rice status dashboard, service health, and theme pipeline verification

source "$HG_DOTFILES/scripts/lib/compositor.sh" 2>/dev/null
source "$HG_DOTFILES/scripts/lib/config.sh" 2>/dev/null
source "$HG_DOTFILES/scripts/lib/kitty-config.sh" 2>/dev/null
source "$HG_DOTFILES/scripts/lib/tmux-persistence.sh" 2>/dev/null

rice_description() {
  echo "Rice health — status, services, palette, reload-all"
}

rice_commands() {
  cat <<'CMDS'
status	Comprehensive rice status dashboard
services	Check which services are running
palette	Scan configs for non-palette colors
persistence	Check tmux continuity bootstrap and plugin health
reload-all	Reload all compositor services with safe hot-reload lanes
restart-ui	Explicitly restart service-backed UI companions (guarded; use --force to override failed continuity preflight)
CMDS
}

_rice_status() {
  printf "\n %s%srice status%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Compositor
  local comp
  comp="$(compositor_type 2>/dev/null || echo "unknown")"
  printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "compositor" "$HG_RESET" "$HG_CYAN" "$comp" "$HG_RESET"

  # Terminal
  local kitty_font kitty_opacity
  kitty_font="$(kitty_get_font 2>/dev/null || echo "unknown")"
  kitty_opacity="$(kitty_get_opacity 2>/dev/null || echo "unknown")"
  printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "terminal" "$HG_RESET" "$HG_MAGENTA" "kitty" "$HG_RESET"
  printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "font" "$HG_RESET" "$HG_CYAN" "$kitty_font" "$HG_RESET"
  printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "opacity" "$HG_RESET" "$HG_CYAN" "$kitty_opacity" "$HG_RESET"

  # Wallpaper
  local wallpaper_state
  wallpaper_state="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/wallpaper-mode/mode" 2>/dev/null || true)"
  if pgrep -x mpvpaper &>/dev/null; then
    local wp_video
    wp_video="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/wallpaper-mode/value" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %svideo:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_YELLOW" "$HG_RESET" "$(basename "$wp_video")"
  elif pgrep -x waydeeper &>/dev/null; then
    local wp_depth
    wp_depth="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/wallpaper-mode/value" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %sdepth:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_MAGENTA" "$HG_RESET" "$(basename "$wp_depth")"
  elif pgrep -f shaderbg &>/dev/null; then
    local wp_shader
    wp_shader="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/shader-wallpaper/current" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %sshader:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_MAGENTA" "$HG_RESET" "$(basename "$wp_shader" .frag)"
  else
    local wp_static
    wp_static="$(cat "${XDG_STATE_HOME:-$HOME/.local/state}/swww/current" 2>/dev/null || echo "unknown")"
    printf " %s%-14s%s %s%s:%s %s\n" "$HG_DIM" "wallpaper" "$HG_RESET" "$HG_GREEN" "${wallpaper_state:-static}" "$HG_RESET" "$(basename "$wp_static")"
  fi

  # Bar
  if pgrep -x ironbar &>/dev/null; then
    if pgrep -x quickshell &>/dev/null; then
      printf " %s%-14s%s %sironbar + quickshell pilot%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf " %s%-14s%s %sironbar%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    fi
  elif pgrep -x quickshell &>/dev/null; then
    printf " %s%-14s%s %squickshell pilot%s\n" "$HG_DIM" "bar" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
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
    "ironbar:ironbar"
    "quickshell:quickshell"
    "hyprshell:hyprshell"
    "hypr-dock:hypr-dock"
    "hyprdynamicmonitors:hyprdynamicmonitors"
    "autoname:hyprland-autoname-workspaces"
    "swaync:swaync"
    "swww:swww-daemon"
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
  printf "\n %s%stheme scan%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  local scan_dirs=(
    "$HG_DOTFILES/hyprland"
    "$HG_DOTFILES/ironbar"
    "$HG_DOTFILES/quickshell"
    "$HG_DOTFILES/hyprshell"
    "$HG_DOTFILES/swaync"
    "$HG_DOTFILES/wofi"
    "$HG_DOTFILES/wlogout"
    "$HG_DOTFILES/kitty"
    "$HG_DOTFILES/fontconfig"
  )

  local issues=0
  local checks=(
    "Hack Nerd Font|legacy UI font"
    "Matcha-dark-sea|legacy GTK theme"
    "JetBrains Mono|legacy font"
  )

  local entry pattern label dir
  for entry in "${checks[@]}"; do
    pattern="${entry%%|*}"
    label="${entry#*|}"
    for dir in "${scan_dirs[@]}"; do
      [[ -e "$dir" ]] || continue
      while IFS=: read -r file line _; do
        [[ -n "$file" ]] || continue
        printf "  %s%s%s:%s%s%s %s%s%s\n" "$HG_DIM" "$(basename "$file")" "$HG_RESET" "$HG_DIM" "$line" "$HG_RESET" "$HG_RED" "$label" "$HG_RESET"
        issues=$((issues + 1))
      done < <(rg -n \
        --glob '*.conf' \
        --glob '*.css' \
        --glob '*.scss' \
        --glob '*.yuck' \
        --glob '*.ini' \
        --glob '*.json' \
        --glob '*.toml' \
        "$pattern" "$dir" 2>/dev/null || true)
    done
  done

  if [[ $issues -eq 0 ]]; then
    hg_ok "No legacy shell-theme references found"
  else
    hg_warn "$issues legacy theme reference(s) found"
  fi

  if [[ -f "$HG_DOTFILES/theme/palette.env" && -x "$HG_DOTFILES/scripts/theme-sync.sh" ]]; then
    hg_ok "Theme sync pipeline is present"
  else
    hg_warn "Theme sync pipeline is incomplete"
  fi

  if [[ -x "$HG_DOTFILES/scripts/hyprpm-bootstrap.sh" ]]; then
    hg_ok "Hyprland plugin bootstrap is present"
  else
    hg_warn "Hyprland plugin bootstrap is missing"
  fi
  printf "\n"
}

_rice_persistence() {
  "$HG_DOTFILES/scripts/tmux-persistence-health.sh"
}

_rice_reload_all() {
  hg_info "Reloading all services in parallel..."
  config_reload_parallel hyprland hyprshell hypr-dock hyprdynamicmonitors autoname swaync ironbar quickshell
  hg_ok "All services reloaded"
}

_rice_restart_ui() {
  local force=false

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --force) force=true ;;
      -*) hg_die "Unknown option for restart-ui: $1" ;;
      *) hg_die "Usage: hg rice restart-ui [--force]" ;;
    esac
    shift
  done

  if ! $force && ! tmux_persistence_is_operational; then
    hg_die "Blocked UI restart: tmux persistence health has failures. Run 'hg rice persistence' or retry with --force."
  fi

  hg_info "Restarting UI companion services in parallel..."
  config_restart_parallel hyprshell hypr-dock hyprdynamicmonitors autoname
  hg_ok "UI companion services restarted"
}

rice_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)     _rice_status ;;
    services)   _rice_services ;;
    palette)    _rice_palette ;;
    persistence) _rice_persistence ;;
    reload-all) _rice_reload_all ;;
    restart-ui) _rice_restart_ui "$@" ;;
    *)          hg_die "Unknown rice command: $cmd. Run 'hg rice --help'." ;;
  esac
}
