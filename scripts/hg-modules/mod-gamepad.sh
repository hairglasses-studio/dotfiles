#!/usr/bin/env bash
# mod-gamepad.sh — hg gamepad module
# Xbox controller management via makima (Rust input remapper)
# Makima handles per-app profile switching natively on Hyprland

_GAMEPAD_PROFILE_DIR="$HG_DOTFILES/makima"
_GAMEPAD_DEVICE_NAME="Microsoft Xbox Series S|X Controller"

gamepad_description() {
  echo "Xbox controller — status, profiles, restart (makima)"
}

gamepad_commands() {
  cat <<'CMDS'
status	Controller connection + makima service state
profiles	List Xbox controller profiles (desktop + per-app)
restart	Restart makima to pick up profile changes
test	Quick button test via evtest
CMDS
}

_gamepad_status() {
  printf "\n %s%sgamepad%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Controller detection
  if [[ -e /dev/input/js0 ]]; then
    local name
    name="$(grep -A2 'Xbox\|xbox' /proc/bus/input/devices 2>/dev/null | grep 'N: Name=' | head -1 | sed 's/.*Name="//;s/"//')"
    printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "controller" "$HG_RESET" "$HG_GREEN" "${name:-connected}" "$HG_RESET"
  else
    printf " %s%-14s%s %sdisconnected%s\n" "$HG_DIM" "controller" "$HG_RESET" "$HG_DIM" "$HG_RESET"
  fi

  # Makima service
  if systemctl is-active makima.service &>/dev/null; then
    printf " %s%-14s%s %sactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf " %s%-14s%s %sinactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # Profile detection
  local base_profile="$_GAMEPAD_PROFILE_DIR/$_GAMEPAD_DEVICE_NAME.toml"
  if [[ -f "$base_profile" ]]; then
    printf " %s%-14s%s %sdesktop%s\n" "$HG_DIM" "base profile" "$HG_RESET" "$HG_CYAN" "$HG_RESET"
  else
    printf " %s%-14s%s %snot configured%s\n" "$HG_DIM" "base profile" "$HG_RESET" "$HG_DIM" "$HG_RESET"
  fi

  # Per-app profiles
  local app_count=0
  while IFS= read -r -d '' f; do
    app_count=$((app_count + 1))
  done < <(find "$_GAMEPAD_PROFILE_DIR" -maxdepth 1 -name "$_GAMEPAD_DEVICE_NAME::*.toml" -print0 2>/dev/null)
  if [[ $app_count -gt 0 ]]; then
    printf " %s%-14s%s %s%d app override(s)%s\n" "$HG_DIM" "per-app" "$HG_RESET" "$HG_MAGENTA" "$app_count" "$HG_RESET"
  fi

  # Force feedback
  if [[ -e /dev/input/js0 ]]; then
    local ff
    ff="$(grep -A8 -i 'xbox' /proc/bus/input/devices 2>/dev/null | grep 'B: FF=' | head -1 || true)"
    if [[ -n "$ff" ]]; then
      printf " %s%-14s%s %savailable%s\n" "$HG_DIM" "rumble" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    fi
  fi

  printf "\n"
}

_gamepad_profiles() {
  printf "\n %s%sxbox controller profiles%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Base profile
  local base="$_GAMEPAD_PROFILE_DIR/$_GAMEPAD_DEVICE_NAME.toml"
  if [[ -f "$base" ]]; then
    printf "  %s%-40s%s %sbase%s\n" "$HG_CYAN" "$_GAMEPAD_DEVICE_NAME" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  fi

  # Per-app profiles
  while IFS= read -r -d '' f; do
    local app
    app="$(basename "$f" .toml)"
    app="${app#*::}"
    printf "  %s%-40s%s %s→ %s%s\n" "$HG_DIM" "$_GAMEPAD_DEVICE_NAME" "$HG_RESET" "$HG_MAGENTA" "$app" "$HG_RESET"
  done < <(find "$_GAMEPAD_PROFILE_DIR" -maxdepth 1 -name "$_GAMEPAD_DEVICE_NAME::*.toml" -print0 2>/dev/null)

  printf "\n"
}

_gamepad_restart() {
  hg_info "Restarting makima..."
  sudo systemctl restart makima.service
  sleep 0.5
  if systemctl is-active makima.service &>/dev/null; then
    hg_ok "makima restarted"
  else
    hg_error "makima failed to start — check: journalctl --user -u makima -n 20"
  fi
}

_gamepad_test() {
  if [[ ! -e /dev/input/js0 ]]; then
    hg_die "No controller connected at /dev/input/js0"
  fi
  hg_require evtest
  hg_info "Press buttons on the controller (Ctrl+C to stop)"
  local event_dev
  event_dev="$(grep -B5 'js0' /proc/bus/input/devices 2>/dev/null | grep 'H: Handlers=' | grep -oP 'event\d+' | head -1)"
  if [[ -n "$event_dev" ]]; then
    evtest "/dev/input/$event_dev"
  else
    evtest /dev/input/js0
  fi
}

gamepad_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)   _gamepad_status ;;
    profiles) _gamepad_profiles ;;
    restart)  _gamepad_restart ;;
    test)     _gamepad_test ;;
    *)        hg_die "Unknown gamepad command: $cmd. Run 'hg gamepad --help'." ;;
  esac
}
