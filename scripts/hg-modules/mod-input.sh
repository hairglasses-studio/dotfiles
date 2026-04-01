#!/usr/bin/env bash
# mod-input.sh — hg input module
# Manage MX Master 4 input stack: logiops, solaar, makima

input_description() {
  echo "Input devices — logiops, solaar, makima"
}

input_commands() {
  cat <<'CMDS'
status	Show running state of input services
deploy	Deploy logiops config to /etc/ (needs sudo)
sync-solaar	Copy live solaar config back to dotfiles
restart	Restart input services
devices	List makima-detected evdev devices
CMDS
}

_input_status() {
  printf "\n %s%sinput devices%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # logid
  if systemctl is-active logid.service &>/dev/null; then
    printf "  %s%-12s%s %sactive%s\n" "$HG_DIM" "logid" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sinactive%s\n" "$HG_DIM" "logid" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # makima (system service — needs input group)
  if systemctl is-active makima.service &>/dev/null; then
    printf "  %s%-12s%s %sactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sinactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # solaar config
  if [[ -f "$HOME/.config/solaar/config.yaml" ]]; then
    printf "  %s%-12s%s %spresent%s\n" "$HG_DIM" "solaar cfg" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %smissing%s\n" "$HG_DIM" "solaar cfg" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # battery via solaar
  if command -v solaar &>/dev/null; then
    local battery
    battery="$(solaar show 2>/dev/null | grep -a 'Battery' | head -1 | sed 's/.*: //')"
    if [[ -n "$battery" ]]; then
      printf "  %s%-12s%s %s\n" "$HG_DIM" "battery" "$HG_RESET" "$battery"
    fi
  fi

  printf "\n"
}

_input_deploy() {
  local script="$HG_DOTFILES/scripts/logiops-deploy.sh"
  if [[ ! -x "$script" ]]; then
    hg_die "Deploy script not found: $script"
  fi
  "$script"
}

_input_sync_solaar() {
  local live="$HOME/.config/solaar/config.yaml"
  local tracked="$HG_DOTFILES/solaar/config.yaml"

  if [[ ! -f "$live" ]]; then
    hg_die "Live solaar config not found at $live"
  fi

  cp "$live" "$tracked"
  hg_ok "Synced solaar config to dotfiles"
}

_input_restart() {
  hg_info "Restarting input services..."

  if systemctl is-active logid.service &>/dev/null; then
    sudo systemctl restart logid.service
    hg_ok "logid restarted"
  else
    hg_warn "logid not active, skipping"
  fi

  if systemctl is-enabled makima.service &>/dev/null; then
    sudo systemctl restart makima.service
    hg_ok "makima restarted"
  else
    hg_warn "makima not enabled, skipping"
  fi
}

_input_devices() {
  hg_require makima
  makima --list-devices
}

input_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)       _input_status ;;
    deploy)       _input_deploy ;;
    sync-solaar)  _input_sync_solaar ;;
    restart)      _input_restart ;;
    devices)      _input_devices ;;
    *)            hg_die "Unknown input command: $cmd. Run 'hg input --help'." ;;
  esac
}
