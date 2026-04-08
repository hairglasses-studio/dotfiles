#!/usr/bin/env bash
# mod-config.sh — hg config module
# Dotfiles config management — reload, backup, check, list

source "$HG_DOTFILES/scripts/lib/config.sh"
_CFG_TOML="$HG_DOTFILES/dotfiles.toml"

config_description() {
  echo "Dotfiles config — reload, backup, check, list"
}

config_commands() {
  cat <<'CMDS'
reload	Reload a component (hyprland|hyprshell|hypr-dock|hyprdynamicmonitors|autoname|swaync|eww|tmux)
backup	Backup a config file before modification
check	Validate symlinks and feature flags
list	Show managed components and paths
CMDS
}

_config_cmd_reload() {
  local component="${1:-}"
  [[ -n "$component" ]] || hg_die "Usage: hg config reload <component> (hyprland|hyprshell|hypr-dock|hyprdynamicmonitors|autoname|swaync|eww|tmux)"
  case "$component" in
    hyprland|hypr|hyprshell|hypr-dock|hyprdock|hyprdynamicmonitors|monitors|hyprland-autoname-workspaces|autoname|swaync|eww|ironbar|tmux) ;;
    *) hg_die "Unknown component: $component (hyprland|hyprshell|hypr-dock|hyprdynamicmonitors|autoname|swaync|eww|tmux)" ;;
  esac
  config_reload_service "$component"
  hg_ok "Reloaded $component"
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
    "$HOME/.config/hyprshell:$HG_DOTFILES/hyprshell:hyprshell"
    "$HOME/.config/hypr-dock:$HG_DOTFILES/hypr-dock:hypr-dock"
    "$HOME/.config/hyprdynamicmonitors:$HG_DOTFILES/hyprdynamicmonitors:hyprdynamicmonitors"
    "$HOME/.config/hyprland-autoname-workspaces:$HG_DOTFILES/hyprland-autoname-workspaces:autoname"
    "$HOME/.config/eww:$HG_DOTFILES/eww:eww"
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
    "hyprshell:$HG_DOTFILES/hyprshell:systemctl --user restart dotfiles-hyprshell.service"
    "hypr-dock:$HG_DOTFILES/hypr-dock:systemctl --user restart dotfiles-hypr-dock.service"
    "hyprdynamic:$HG_DOTFILES/hyprdynamicmonitors:systemctl --user restart dotfiles-hyprdynamicmonitors.service"
    "autoname:$HG_DOTFILES/hyprland-autoname-workspaces:systemctl --user restart dotfiles-hyprland-autoname-workspaces.service"
    "eww:$HG_DOTFILES/eww:eww reload"
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
    backup) _config_cmd_backup "$@" ;;
    check)  _config_cmd_check ;;
    list)   _config_cmd_list ;;
    *)      hg_die "Unknown config command: $cmd. Run 'hg config --help'." ;;
  esac
}
