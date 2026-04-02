#!/usr/bin/env bash
# mod-gamepad.sh — hg gamepad module
# Xbox controller + AntiMicroX profile management

_GAMEPAD_PROFILE_DIR="$HOME/.config/antimicrox"
_GAMEPAD_SCRIPTS="$HG_DOTFILES/scripts/gamepad"

gamepad_description() {
  echo "Xbox controller — profiles, status, test"
}

gamepad_commands() {
  cat <<'CMDS'
status	Controller connection + AntiMicroX state
profile	Switch AntiMicroX profile by name
profiles	List available profiles
test	Quick button test via evtest
CMDS
}

_gamepad_status() {
  printf "\n %s%sgamepad%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Controller detection
  if [[ -e /dev/input/js0 ]]; then
    local name
    name="$(cat /proc/bus/input/devices 2>/dev/null | grep -A2 'Xbox\|xbox' | grep 'N: Name=' | head -1 | sed 's/.*Name="//' | sed 's/"//')"
    printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "controller" "$HG_RESET" "$HG_GREEN" "${name:-connected}" "$HG_RESET"
  else
    printf " %s%-14s%s %sdisconnected%s\n" "$HG_DIM" "controller" "$HG_RESET" "$HG_DIM" "$HG_RESET"
  fi

  # AntiMicroX process
  local amx_pid
  amx_pid="$(pgrep -x antimicrox 2>/dev/null)"
  if [[ -n "$amx_pid" ]]; then
    local profile_path
    profile_path="$(ps -p "$amx_pid" -o args= 2>/dev/null | grep -oP -- '--profile \K\S+')"
    local profile_name
    profile_name="$(basename "$profile_path" .gamecontroller.amgp 2>/dev/null)"
    printf " %s%-14s%s %srunning%s %s(pid %s)%s\n" "$HG_DIM" "antimicrox" "$HG_RESET" "$HG_GREEN" "$HG_RESET" "$HG_DIM" "$amx_pid" "$HG_RESET"
    printf " %s%-14s%s %s%s%s\n" "$HG_DIM" "profile" "$HG_RESET" "$HG_CYAN" "${profile_name:-unknown}" "$HG_RESET"
  else
    printf " %s%-14s%s %sstopped%s\n" "$HG_DIM" "antimicrox" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # Force feedback
  if [[ -e /dev/input/js0 ]]; then
    local ff
    ff="$(cat /proc/bus/input/devices 2>/dev/null | grep -A8 'Xbox\|xbox' | grep 'B: FF=' | head -1)"
    if [[ -n "$ff" ]]; then
      printf " %s%-14s%s %savailable%s\n" "$HG_DIM" "rumble" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    fi
  fi

  printf "\n"
}

_gamepad_profile() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: hg gamepad profile <name>"

  local profile_path="$_GAMEPAD_PROFILE_DIR/${name}.gamecontroller.amgp"
  [[ -f "$profile_path" ]] || hg_die "Profile not found: $name (expected $profile_path)"

  pkill antimicrox 2>/dev/null
  sleep 0.3
  antimicrox --tray --hidden --profile "$profile_path" &
  disown
  hg_ok "Switched to profile: $name"
}

_gamepad_profiles() {
  printf "\n %s%savailable profiles%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  local active_profile=""
  local amx_pid
  amx_pid="$(pgrep -x antimicrox 2>/dev/null)"
  if [[ -n "$amx_pid" ]]; then
    active_profile="$(ps -p "$amx_pid" -o args= 2>/dev/null | grep -oP -- '--profile \K\S+')"
    active_profile="$(basename "$active_profile" .gamecontroller.amgp 2>/dev/null)"
  fi

  for f in "$_GAMEPAD_PROFILE_DIR"/*.gamecontroller.amgp; do
    [[ -f "$f" ]] || continue
    local name
    name="$(basename "$f" .gamecontroller.amgp)"
    if [[ "$name" == "$active_profile" ]]; then
      printf "  %s%-30s%s %sactive%s\n" "$HG_CYAN" "$name" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-30s%s\n" "$HG_DIM" "$name" "$HG_RESET"
    fi
  done
  printf "\n"
}

_gamepad_test() {
  if [[ ! -e /dev/input/js0 ]]; then
    hg_die "No controller connected at /dev/input/js0"
  fi
  hg_require evtest
  hg_info "Press buttons on the controller (Ctrl+C to stop)"
  local event_dev
  event_dev="$(cat /proc/bus/input/devices 2>/dev/null | grep -B5 'js0' | grep 'H: Handlers=' | grep -oP 'event\d+' | head -1)"
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
    profile)  _gamepad_profile "$@" ;;
    profiles) _gamepad_profiles ;;
    test)     _gamepad_test ;;
    *)        hg_die "Unknown gamepad command: $cmd. Run 'hg gamepad --help'." ;;
  esac
}
