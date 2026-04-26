#!/usr/bin/env bash
# mod-config.sh — hg config module
# Dotfiles config management — reload, backup, check, list

source "$HG_DOTFILES/scripts/lib/config.sh"
source "$HG_DOTFILES/scripts/lib/tmux-persistence.sh" 2>/dev/null
_CFG_TOML="$HG_DOTFILES/dotfiles.toml"

config_description() {
  echo "Dotfiles config — reload, backup, check, list"
}

config_commands() {
  cat <<'CMDS'
reload	Reload a component (hyprland|swaync|quickshell|tmux + hyprdynamicmonitors|autoname)
restart	Explicitly restart a service-backed component (guarded; use --force to override failed continuity preflight)
lane	Show whether a component uses safe_reload, service_reload, or explicit_restart
backup	Backup a config file before modification
check	Validate symlinks and feature flags
list	Show managed components and paths
CMDS
}

_CONFIG_RELOAD_COMPONENTS="hyprland|hypr|hyprdynamicmonitors|monitors|hyprland-autoname-workspaces|autoname|swaync|quickshell|tmux"
_CONFIG_RESTART_COMPONENTS="hyprdynamicmonitors|monitors|hyprland-autoname-workspaces|autoname|quickshell"

_config_cmd_reload() {
  local component="${1:-}"
  [[ -n "$component" ]] || hg_die "Usage: hg config reload <component> ($_CONFIG_RELOAD_COMPONENTS)"
  case "$component" in
    hyprland|hypr|hyprdynamicmonitors|monitors|hyprland-autoname-workspaces|autoname|swaync|quickshell|tmux) ;;
    *) hg_die "Unknown component: $component ($_CONFIG_RELOAD_COMPONENTS)" ;;
  esac
  config_reload_service "$component"
  hg_ok "Reloaded $component"
}

_config_cmd_restart() {
  local force=false component=""

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --force) force=true ;;
      -*) hg_die "Unknown option for restart: $1" ;;
      *)
        [[ -z "$component" ]] || hg_die "Usage: hg config restart [--force] <component> ($_CONFIG_RESTART_COMPONENTS)"
        component="$1"
        ;;
    esac
    shift
  done

  [[ -n "$component" ]] || hg_die "Usage: hg config restart [--force] <component> ($_CONFIG_RESTART_COMPONENTS)"
  case "$component" in
    hyprdynamicmonitors|monitors|hyprland-autoname-workspaces|autoname|quickshell) ;;
    *) hg_die "Unknown restartable component: $component ($_CONFIG_RESTART_COMPONENTS)" ;;
  esac

  if ! $force && ! tmux_persistence_is_operational; then
    hg_die "Blocked explicit restart for $component: tmux persistence health has failures. Run 'hg rice persistence' or retry with --force."
  fi

  config_restart_service "$component"
  hg_ok "Restarted $component"
}

_config_cmd_lane() {
  local component="${1:-}"
  local verb="${2:-}"
  local reload_lane="" restart_lane=""

  [[ -n "$component" ]] || hg_die "Usage: hg config lane <component> [reload|restart]"
  if [[ -n "$verb" && "$verb" != "reload" && "$verb" != "restart" ]]; then
    hg_die "Unknown action verb: $verb (reload|restart)"
  fi

  if [[ -z "$verb" || "$verb" == "reload" ]]; then
    reload_lane="$(config_action_lane reload "$component" 2>/dev/null || true)"
  fi
  if [[ -z "$verb" || "$verb" == "restart" ]]; then
    restart_lane="$(config_action_lane restart "$component" 2>/dev/null || true)"
  fi

  [[ -n "$reload_lane$restart_lane" ]] || hg_die "Unknown component: $component"

  if [[ -n "$reload_lane" ]]; then
    printf '%s\treload\t%s\n' "$component" "$reload_lane"
  fi
  if [[ -n "$restart_lane" ]]; then
    printf '%s\trestart\t%s\n' "$component" "$restart_lane"
  fi
}

_config_cmd_backup() {
  local file="${1:-}"
  [[ -n "$file" ]] || hg_die "Usage: hg config backup <file>"
  [[ -f "$file" ]] || hg_die "File not found: $file"
  config_backup "$file"
  hg_ok "Backed up $file to ~/.dotfiles-backup/$(date +%Y%m%d)/"
}

_config_cmd_check() {
  printf "\n %s%sconfig check%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  printf " %sSYMLINKS%s\n" "$HG_BOLD" "$HG_RESET"
  local -a _checks=(
    "$HOME/.config/kitty:$HG_DOTFILES/kitty:kitty"
    "$HOME/.config/hypr:$HG_DOTFILES/hyprland:hyprland"
    "$HOME/.config/hyprdynamicmonitors:$HG_DOTFILES/hyprdynamicmonitors:hyprdynamicmonitors"
    "$HOME/.config/hyprland-autoname-workspaces:$HG_DOTFILES/hyprland-autoname-workspaces:autoname"
    "$HOME/.config/quickshell:$HG_DOTFILES/quickshell:quickshell"
    "$HOME/.config/swaync:$HG_DOTFILES/swaync:swaync"
  )
  for _entry in "${_checks[@]}"; do
    local _link _target _name
    _link="$(echo "$_entry" | cut -d: -f1)"
    _target="$(echo "$_entry" | cut -d: -f2)"
    _name="$(echo "$_entry" | cut -d: -f3)"
    if [[ -L "$_link" ]]; then
      local _actual
      _actual="$(readlink -f "$_link" 2>/dev/null)"
      local _expected
      _expected="$(readlink -f "$_target" 2>/dev/null)"
      if [[ "$_actual" == "$_expected" ]]; then
        printf "  %s%-18s%s %sok%s\n" "$HG_CYAN" "$_name" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
      else
        printf "  %s%-18s%s %swrong target%s\n" "$HG_CYAN" "$_name" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
      fi
    elif [[ -e "$_link" ]]; then
      printf "  %s%-18s%s %sexists (not symlink)%s\n" "$HG_CYAN" "$_name" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
    else
      printf "  %s%-18s%s %smissing%s\n" "$HG_CYAN" "$_name" "$HG_RESET" "$HG_DIM" "$HG_RESET"
    fi
  done

  printf "\n %sTERMINAL POLICY%s\n" "$HG_BOLD" "$HG_RESET"
  if grep -Fq '$term = $HOME/.local/bin/kitty-shell-launch' "$HG_DOTFILES/hyprland/hyprland.conf"; then
    printf "  %s%-18s%s %skitty-shell-launch%s\n" "$HG_CYAN" "default terminal" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-18s%s %sdrifted%s\n" "$HG_CYAN" "default terminal" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  fi

  if grep -Eq '^startup_session[[:space:]]+none$' "$HG_DOTFILES/kitty/kitty.conf"; then
    printf "  %s%-18s%s %sstartup_session none%s\n" "$HG_CYAN" "kitty session" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-18s%s %smissing startup_session none%s\n" "$HG_CYAN" "kitty session" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  fi

  if grep -Fq '$HOME/.local/bin/kitty-dev-launch' "$HG_DOTFILES/hyprland/hyprland.conf" \
    && grep -Fq '$HOME/.local/bin/kitty-visual-launch' "$HG_DOTFILES/hyprland/hyprland.conf" \
    && grep -Fq '$HOME/.local/bin/kitty-visual-launch --class=scratchpad' "$HG_DOTFILES/pypr/config.toml"; then
    printf "  %s%-18s%s %smanaged wrappers%s\n" "$HG_CYAN" "hypr surfaces" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-18s%s %slauncher drift%s\n" "$HG_CYAN" "hypr surfaces" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  fi

  if grep -Eq 'dotfiles-kitty-save-session\.(service|timer)' "$HG_DOTFILES/install.sh" "$HG_DOTFILES/manjaro/install.sh"; then
    printf "  %s%-18s%s %sinstaller enables save-session%s\n" "$HG_CYAN" "kitty saver" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  else
    printf "  %s%-18s%s %sopt-in only%s\n" "$HG_CYAN" "kitty saver" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  fi

  if [[ -f "$_CFG_TOML" ]]; then
    printf "\n %sFEATURE FLAGS%s\n" "$HG_BOLD" "$HG_RESET"
    grep -E '^\w+ *= *(true|false)' "$_CFG_TOML" 2>/dev/null | while IFS='=' read -r key val; do
      key="${key// /}"; val="${val// /}"
      if [[ "$val" == "true" ]]; then
        printf "  %s%-18s%s %senabled%s\n" "$HG_CYAN" "$key" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
      else
        printf "  %s%-18s%s %sdisabled%s\n" "$HG_CYAN" "$key" "$HG_RESET" "$HG_DIM" "$HG_RESET"
      fi
    done
  fi
  printf "\n"
}

_config_cmd_list() {
  printf "\n %s%smanaged components%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local -a _components=(
    "kitty:$HG_DOTFILES/kitty:SIGUSR1 reload"
    "hyprland:$HG_DOTFILES/hyprland:hyprctl reload"
    "hyprdynamic:$HG_DOTFILES/hyprdynamicmonitors:systemctl --user restart dotfiles-hyprdynamicmonitors.service"
    "autoname:$HG_DOTFILES/hyprland-autoname-workspaces:systemctl --user restart dotfiles-hyprland-autoname-workspaces.service"
    "quickshell:$HG_DOTFILES/quickshell:systemctl --user restart dotfiles-quickshell.service"
    "swaync:$HG_DOTFILES/swaync:swaync-client --reload-config"
    "tmux:$HG_DOTFILES/tmux:tmux source-file"
    "zsh:$HG_DOTFILES/zsh:source ~/.zshrc"
    "git:$HG_DOTFILES/git:n/a"
    "starship:$HG_DOTFILES/starship:auto-reload"
  )
  for _entry in "${_components[@]}"; do
    local _name _path _reload
    _name="$(echo "$_entry" | cut -d: -f1)"
    _path="$(echo "$_entry" | cut -d: -f2)"
    _reload="$(echo "$_entry" | cut -d: -f3)"
    if [[ -d "$_path" || -f "$_path" ]]; then
      printf "  %s%-12s%s %s%-40s%s %s%s%s\n" "$HG_CYAN" "$_name" "$HG_RESET" "$HG_DIM" "$_path" "$HG_RESET" "$HG_DIM" "$_reload" "$HG_RESET"
    fi
  done
  printf "\n"
}

config_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    reload) _config_cmd_reload "$@" ;;
    restart) _config_cmd_restart "$@" ;;
    lane)   _config_cmd_lane "$@" ;;
    backup) _config_cmd_backup "$@" ;;
    check)  _config_cmd_check ;;
    list)   _config_cmd_list ;;
    *)      hg_die "Unknown config command: $cmd. Run 'hg config --help'." ;;
  esac
}
