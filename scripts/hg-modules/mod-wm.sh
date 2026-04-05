#!/usr/bin/env bash
# mod-wm.sh — hg wm module
# Compositor abstraction — works on Hyprland, AeroSpace

source "$HG_DOTFILES/scripts/lib/compositor.sh"

wm_description() {
  echo "Window manager — compositor abstraction layer"
}

wm_commands() {
  cat <<'CMDS'
type	Show compositor type (hyprland|aerospace)
reload	Reload compositor config
windows	List all windows
workspaces	List workspaces
active	Show active window info
jump	Switch to workspace N
dispatch	Send raw compositor command
CMDS
}

_wm_type() {
  local t
  t="$(compositor_type)"
  printf "%s%s%s\n" "$HG_CYAN" "$t" "$HG_RESET"
}

_wm_reload() {
  compositor_reload
  hg_ok "Compositor reloaded"
}

_wm_windows() {
  hg_require jq
  local json
  json="$(compositor_query windows 2>/dev/null)"
  [[ -n "$json" ]] || hg_die "Could not query windows"

  printf "\n %s%swindows%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local comp
  comp="$(compositor_type)"
  if [[ "$comp" == "hyprland" ]]; then
    echo "$json" | jq -r '.[] | "  \(.workspace.id)\t\(.class)\t\(.title[:60])"' 2>/dev/null | column -t -s$'\t'
  else
    echo "$json" | jq '.' 2>/dev/null
  fi
  printf "\n"
}

_wm_workspaces() {
  hg_require jq
  local json
  json="$(compositor_query workspaces 2>/dev/null)"
  [[ -n "$json" ]] || hg_die "Could not query workspaces"

  printf "\n %s%sworkspaces%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local comp
  comp="$(compositor_type)"
  if [[ "$comp" == "hyprland" ]]; then
    echo "$json" | jq -r '.[] | "  \(.id)\t\(.windows) windows\t\(.name // "")"' 2>/dev/null | column -t -s$'\t'
  else
    echo "$json" | jq '.' 2>/dev/null
  fi
  printf "\n"
}

_wm_active() {
  hg_require jq
  local json
  json="$(compositor_query activewindow 2>/dev/null)"
  [[ -n "$json" ]] || hg_die "No active window"

  printf "\n %s%sactive window%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local comp
  comp="$(compositor_type)"
  if [[ "$comp" == "hyprland" ]]; then
    printf "  %s%-12s%s %s\n" "$HG_DIM" "class" "$HG_RESET" "$(echo "$json" | jq -r '.class')"
    printf "  %s%-12s%s %s\n" "$HG_DIM" "title" "$HG_RESET" "$(echo "$json" | jq -r '.title')"
    printf "  %s%-12s%s %s\n" "$HG_DIM" "workspace" "$HG_RESET" "$(echo "$json" | jq -r '.workspace.id')"
    printf "  %s%-12s%s %sx%s\n" "$HG_DIM" "size" "$HG_RESET" "$(echo "$json" | jq -r '.size[0]')" "$(echo "$json" | jq -r '.size[1]')"
  else
    echo "$json" | jq '.' 2>/dev/null
  fi
  printf "\n"
}

_wm_jump() {
  local ws="${1:-}"
  [[ -n "$ws" ]] || hg_die "Usage: hg wm jump <workspace>"
  compositor_workspace "$ws"
}

_wm_dispatch() {
  [[ $# -gt 0 ]] || hg_die "Usage: hg wm dispatch <command>"
  compositor_msg "$*"
}

wm_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    type)       _wm_type ;;
    reload)     _wm_reload ;;
    windows)    _wm_windows ;;
    workspaces) _wm_workspaces ;;
    active)     _wm_active ;;
    jump)       _wm_jump "$@" ;;
    dispatch)   _wm_dispatch "$@" ;;
    *)          hg_die "Unknown wm command: $cmd. Run 'hg wm --help'." ;;
  esac
}
