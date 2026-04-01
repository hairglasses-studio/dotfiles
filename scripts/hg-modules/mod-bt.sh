#!/usr/bin/env bash
# mod-bt.sh — hg bt module
# Bluetooth device management via bluetoothctl

bt_description() {
  echo "Bluetooth — list, connect, disconnect, scan"
}

bt_commands() {
  cat <<'CMDS'
status	Adapter state and connected devices
list	List paired devices with connection state
connect	Connect to a device by MAC or name
disconnect	Disconnect a device (or all)
scan	Scan for nearby devices
CMDS
}

_bt_status() {
  hg_require bluetoothctl
  printf "\n %s%sbluetooth%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  # Adapter state
  local powered
  powered="$(bluetoothctl show 2>/dev/null | grep -oP 'Powered: \K\w+')"
  if [[ "$powered" == "yes" ]]; then
    printf " %s%-14s%s %son%s\n" "$HG_DIM" "adapter" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
  else
    printf " %s%-14s%s %soff%s\n" "$HG_DIM" "adapter" "$HG_RESET" "$HG_RED" "$HG_RESET"
  fi

  # Discoverable
  local discoverable
  discoverable="$(bluetoothctl show 2>/dev/null | grep -oP 'Discoverable: \K\w+')"
  printf " %s%-14s%s %s\n" "$HG_DIM" "discoverable" "$HG_RESET" "${discoverable:-no}"

  # Connected devices
  local connected
  connected="$(bluetoothctl devices Connected 2>/dev/null)"
  if [[ -n "$connected" ]]; then
    printf "\n %sCONNECTED%s\n" "$HG_BOLD" "$HG_RESET"
    echo "$connected" | while read -r _ mac name; do
      printf "  %s%-20s%s %s%s%s\n" "$HG_GREEN" "$name" "$HG_RESET" "$HG_DIM" "$mac" "$HG_RESET"
    done
  else
    printf "\n %sno connected devices%s\n" "$HG_DIM" "$HG_RESET"
  fi
  printf "\n"
}

_bt_list() {
  hg_require bluetoothctl
  printf "\n %s%spaired devices%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  local connected_macs
  connected_macs="$(bluetoothctl devices Connected 2>/dev/null | awk '{print $2}')"

  bluetoothctl devices Paired 2>/dev/null | while read -r _ mac name; do
    if echo "$connected_macs" | grep -q "$mac"; then
      printf "  %s%-20s%s %s%s%s  %sconnected%s\n" "$HG_GREEN" "$name" "$HG_RESET" "$HG_DIM" "$mac" "$HG_RESET" "$HG_GREEN" "$HG_RESET"
    else
      printf "  %s%-20s%s %s%s%s\n" "$HG_DIM" "$name" "$HG_RESET" "$HG_DIM" "$mac" "$HG_RESET"
    fi
  done
  printf "\n"
}

_bt_connect() {
  hg_require bluetoothctl
  local target="${1:-}"
  [[ -n "$target" ]] || hg_die "Usage: hg bt connect <MAC|name>"

  local mac="$target"
  # If not a MAC address, resolve by name
  if ! [[ "$target" =~ ^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$ ]]; then
    mac="$(bluetoothctl devices 2>/dev/null | grep -i "$target" | awk '{print $2}' | head -1)"
    [[ -n "$mac" ]] || hg_die "Device not found: $target"
  fi

  hg_info "Connecting to $mac..."
  bluetoothctl connect "$mac" 2>&1 | tail -1
}

_bt_disconnect() {
  hg_require bluetoothctl
  local target="${1:-}"

  if [[ -z "$target" ]]; then
    # Disconnect all
    bluetoothctl devices Connected 2>/dev/null | while read -r _ mac name; do
      bluetoothctl disconnect "$mac" &>/dev/null
      hg_ok "Disconnected $name"
    done
  else
    local mac="$target"
    if ! [[ "$target" =~ ^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$ ]]; then
      mac="$(bluetoothctl devices 2>/dev/null | grep -i "$target" | awk '{print $2}' | head -1)"
      [[ -n "$mac" ]] || hg_die "Device not found: $target"
    fi
    bluetoothctl disconnect "$mac" &>/dev/null
    hg_ok "Disconnected $mac"
  fi
}

_bt_scan() {
  hg_require bluetoothctl
  hg_info "Scanning for 10 seconds..."
  timeout 10 bluetoothctl scan on 2>/dev/null | grep -E '^\[NEW\]' | while read -r _ _ mac name; do
    printf "  %s%-20s%s %s%s%s\n" "$HG_CYAN" "${name:-unknown}" "$HG_RESET" "$HG_DIM" "$mac" "$HG_RESET"
  done
  bluetoothctl scan off &>/dev/null
  printf "\n"
}

bt_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    status)     _bt_status ;;
    list)       _bt_list ;;
    connect)    _bt_connect "$@" ;;
    disconnect) _bt_disconnect "$@" ;;
    scan)       _bt_scan ;;
    *)          hg_die "Unknown bt command: $cmd. Run 'hg bt --help'." ;;
  esac
}
