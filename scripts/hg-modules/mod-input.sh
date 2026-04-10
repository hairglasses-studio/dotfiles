#!/usr/bin/env bash
# mod-input.sh — hg input module
# Manage MX Master 4 input stack: juhradial-mx + ydotool

input_description() {
  echo "Input devices — juhradial-mx, ydotool, makima"
}

input_commands() {
  cat <<'CMDS'
status	Show running state of juhradial and related services
install	Install or update the juhradial stack
settings	Open the juhradial settings dashboard
sync	Copy repo-owned juhradial seed config into ~/.config/juhradial
restart	Restart juhradial, ydotool, and the overlay
wheel-fix	Reapply compatible MX wheel hardware settings
battery	Show MX Master 4 battery from juhradial D-Bus
devices	List Logitech hidraw devices and paired MX devices
CMDS
}

source "$HG_DOTFILES/scripts/lib/juhradial.sh"

_input_status() {
  printf "\n %s%sinput devices%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local tracked_config="$HG_DOTFILES/juhradial/config.json"
  local tracked_profiles="$HG_DOTFILES/juhradial/profiles.json"
  local tracked_macros="$HG_DOTFILES/juhradial/macros"
  local live_dir="${XDG_CONFIG_HOME:-$HOME/.config}/juhradial"
  local live_config="$live_dir/config.json"
  local live_profiles="$live_dir/profiles.json"
  local live_macros="$live_dir/macros"
  local status transport

  if juhradial_systemctl is-active juhradialmx-daemon.service &>/dev/null; then
    printf "  %s%-12s%s %sactive%s\n" "$HG_DIM" "juhradial" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sinactive%s\n" "$HG_DIM" "juhradial" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  if juhradial_overlay_running; then
    printf "  %s%-12s%s %srunning%s\n" "$HG_DIM" "overlay" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sstopped%s\n" "$HG_DIM" "overlay" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  fi

  if juhradial_systemctl is-active ydotool.service &>/dev/null; then
    printf "  %s%-12s%s %sactive%s\n" "$HG_DIM" "ydotool" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sinactive%s\n" "$HG_DIM" "ydotool" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  if systemctl is-active makima.service &>/dev/null; then
    printf "  %s%-12s%s %sactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf "  %s%-12s%s %sinactive%s\n" "$HG_DIM" "makima" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
  fi

  transport="$(juhradial_transport_state)"
  case "$transport" in
    bolt)
      printf "  %s%-12s%s %sbolt%s\n" "$HG_DIM" "transport" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
      ;;
    bluetooth)
      printf "  %s%-12s%s %sbluetooth%s\n" "$HG_DIM" "transport" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
      ;;
    split-brain)
      printf "  %s%-12s%s %ssplit-brain%s\n" "$HG_DIM" "transport" "$HG_RESET" "$HG_RED" "$HG_RESET"
      ;;
    *)
      printf "  %s%-12s%s %s%s%s\n" "$HG_DIM" "transport" "$HG_RESET" "$HG_YELLOW" "$transport" "$HG_RESET"
      ;;
  esac

  if [[ -f "$tracked_config" && -f "$live_config" ]]; then
    if diff -q "$tracked_config" "$live_config" &>/dev/null; then
      printf "  %s%-12s%s %ssynced%s\n" "$HG_DIM" "config" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-12s%s %sdrifted%s\n" "$HG_DIM" "config" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
    fi
  elif [[ -f "$tracked_config" ]]; then
    printf "  %s%-12s%s %smissing%s\n" "$HG_DIM" "config" "$HG_RESET" "$HG_RED" "$HG_RESET"
  else
    printf "  %s%-12s%s %smissing seed%s\n" "$HG_DIM" "config" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  if [[ -f "$tracked_profiles" && -f "$live_profiles" ]]; then
    if diff -q "$tracked_profiles" "$live_profiles" &>/dev/null; then
      printf "  %s%-12s%s %ssynced%s\n" "$HG_DIM" "profiles" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-12s%s %sdrifted%s\n" "$HG_DIM" "profiles" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
    fi
  elif [[ -f "$tracked_profiles" ]]; then
    printf "  %s%-12s%s %smissing%s\n" "$HG_DIM" "profiles" "$HG_RESET" "$HG_RED" "$HG_RESET"
  else
    printf "  %s%-12s%s %smissing seed%s\n" "$HG_DIM" "profiles" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  if [[ -d "$tracked_macros" && -d "$live_macros" ]]; then
    if diff -qr --exclude='.gitkeep' --exclude='*.bak.*' "$tracked_macros" "$live_macros" &>/dev/null; then
      printf "  %s%-12s%s %ssynced%s\n" "$HG_DIM" "macros" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-12s%s %sdrifted%s\n" "$HG_DIM" "macros" "$HG_RESET" "$HG_YELLOW" "$HG_RESET"
    fi
  elif [[ -d "$tracked_macros" ]]; then
    printf "  %s%-12s%s %smissing%s\n" "$HG_DIM" "macros" "$HG_RESET" "$HG_RED" "$HG_RESET"
  else
    printf "  %s%-12s%s %smissing seed%s\n" "$HG_DIM" "macros" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  if status="$(juhradial_battery_status 2>/dev/null)"; then
    local battery charging
    read -r battery charging <<<"$status"
    printf "  %s%-12s%s %s%%%s\n" "$HG_DIM" "battery" "$HG_RESET" "$battery" "$HG_RESET"
    printf "  %s%-12s%s %s\n" "$HG_DIM" "charging" "$HG_RESET" "$charging"
  fi

  printf "\n"
}

_input_install() {
  local script="$HG_DOTFILES/scripts/juhradial-install.sh"
  if [[ ! -x "$script" ]]; then
    hg_die "Install script not found: $script"
  fi
  "$script"
}

_input_settings() {
  local script="$HG_DOTFILES/scripts/juhradial-settings.sh"
  if [[ ! -x "$script" ]]; then
    hg_die "Settings launcher not found: $script"
  fi
  "$script"
}

_input_sync() {
  local script="$HG_DOTFILES/scripts/juhradial-sync.sh"
  local wheel_script="$HG_DOTFILES/scripts/juhradial-wheel-apply.sh"
  if [[ ! -x "$script" ]]; then
    hg_die "Sync script not found: $script"
  fi
  "$script"
  juhradial_systemctl restart juhradialmx-daemon.service >/dev/null
  [[ -x "$wheel_script" ]] && "$wheel_script" --quiet || true
  hg_ok "Synced juhradial seed config"
}

_input_restart() {
  hg_info "Restarting input services..."
  juhradial_systemctl restart ydotool.service >/dev/null
  juhradial_systemctl restart juhradialmx-daemon.service >/dev/null
  "$HG_DOTFILES/scripts/juhradial-mx.sh" --restart --quiet
  [[ -x "$HG_DOTFILES/scripts/juhradial-wheel-apply.sh" ]] && "$HG_DOTFILES/scripts/juhradial-wheel-apply.sh" --quiet || true
  hg_ok "juhradial + ydotool restarted"
}

_input_wheel_fix() {
  local script="$HG_DOTFILES/scripts/juhradial-wheel-apply.sh"
  if [[ ! -x "$script" ]]; then
    hg_die "Wheel apply script not found: $script"
  fi
  "$script"
}

_input_battery() {
  local status battery charging
  if ! status="$(juhradial_battery_status 2>/dev/null)"; then
    hg_die "Battery unavailable — juhradial daemon is not reporting yet"
  fi

  read -r battery charging <<<"$status"
  printf 'MX Master 4: %s%% (charging: %s)\n' "$battery" "$charging"
}

_input_devices() {
  printf "\n %s%sjuhradial devices%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  printf "  %s%-12s%s %s\n" "$HG_DIM" "transport" "$HG_RESET" "$(juhradial_transport_state)"

  local found=false
  local dev props model
  for dev in /dev/hidraw*; do
    [[ -e "$dev" ]] || continue
    props="$(udevadm info -q property -n "$dev" 2>/dev/null || true)"
    echo "$props" | grep -q '^ID_VENDOR_ID=046d$' || continue
    model="$(echo "$props" | awk -F= '/^ID_MODEL=/{print $2; exit}')"
    printf "  %s%-12s%s %s (%s)\n" "$HG_DIM" "$dev" "$HG_RESET" "${model:-Logitech}" "hidraw"
    found=true
  done

  if ! $found; then
    printf "  %s%-12s%s %s\n" "$HG_DIM" "hidraw" "$HG_RESET" "no Logitech hidraw devices detected"
  fi

  local bt
  bt="$(bluetoothctl devices 2>/dev/null | grep -iE 'MX Master|Logi|Logitech' || true)"
  if [[ -n "$bt" ]]; then
    printf "\n  %sbluetooth%s\n" "$HG_DIM" "$HG_RESET"
    while IFS= read -r line; do
      [[ -n "$line" ]] || continue
      printf "    %s\n" "$line"
    done <<<"$bt"
  fi

  local solaar_show
  solaar_show="$(juhradial_solaar_timeout "${JUHRADIAL_SOLAAR_TIMEOUT:-8s}" show all 2>/dev/null | tr -d '\000' || true)"
  if [[ -n "$solaar_show" ]]; then
    printf "\n  %ssolaar%s\n" "$HG_DIM" "$HG_RESET"
    while IFS= read -r line; do
      case "$line" in
        "Bolt Receiver"|\
        "MX Master 4"|\
        "  Device path  :"*|\
        "  Has "*|\
        "  1: MX Master 4"|\
        "  2: MX Master 4"|\
        "  3: MX Master 4"|\
        "     Battery:"*)
          printf "    %s\n" "$line"
          ;;
      esac
    done <<<"$solaar_show"
  fi

  printf "\n"
}

input_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)       _input_status ;;
    install)      _input_install ;;
    settings)     _input_settings ;;
    sync)         _input_sync ;;
    restart)      _input_restart ;;
    wheel-fix)    _input_wheel_fix ;;
    battery)      _input_battery ;;
    devices)      _input_devices ;;
    *)            hg_die "Unknown input command: $cmd. Run 'hg input --help'." ;;
  esac
}
