#!/usr/bin/env bash

juhradial_dotfiles_dir() {
  local lib_dir
  lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cd "$lib_dir/../.." && pwd
}

juhradial_config_dir() {
  printf '%s\n' "${XDG_CONFIG_HOME:-$HOME/.config}/juhradial"
}

juhradial_macros_dir() {
  printf '%s\n' "$(juhradial_config_dir)/macros"
}

juhradial_seed_dir() {
  printf '%s\n' "$(juhradial_dotfiles_dir)/juhradial"
}

juhradial_seed_macros_dir() {
  printf '%s\n' "$(juhradial_seed_dir)/macros"
}

juhradial_patch_dir() {
  printf '%s\n' "$(juhradial_seed_dir)/patches"
}

juhradial_repo_url() {
  printf '%s\n' "${JUHRADIAL_REPO_URL:-https://github.com/JuhLabs/juhradial-mx.git}"
}

juhradial_pinned_commit() {
  printf '%s\n' "${JUHRADIAL_PINNED_COMMIT:-db83da0a0117fd63081c557d0f1da4d384b1d255}"
}

juhradial_install_dir() {
  printf '%s\n' "${JUHRADIAL_INSTALL_DIR:-$HOME/.local/share/juhradial-mx}"
}

juhradial_source_dir() {
  printf '%s\n' "${JUHRADIAL_SOURCE_DIR:-$HOME/.local/src/juhradial-mx}"
}

juhradial_runtime_dir() {
  if [[ -n "${XDG_RUNTIME_DIR:-}" ]]; then
    printf '%s\n' "$XDG_RUNTIME_DIR"
    return 0
  fi

  printf '/run/user/%s\n' "$(id -u)"
}

juhradial_user_bus() {
  local runtime
  runtime="$(juhradial_runtime_dir)"
  printf 'unix:path=%s/bus\n' "$runtime"
}

juhradial_systemctl() {
  local runtime
  runtime="$(juhradial_runtime_dir)"
  env \
    XDG_RUNTIME_DIR="$runtime" \
    DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)" \
    systemctl --user "$@"
}

juhradial_gdbus() {
  local runtime
  runtime="$(juhradial_runtime_dir)"
  env \
    XDG_RUNTIME_DIR="$runtime" \
    DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)" \
    gdbus "$@"
}

juhradial_dbus_send() {
  local runtime
  runtime="$(juhradial_runtime_dir)"
  env \
    XDG_RUNTIME_DIR="$runtime" \
    DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)" \
    dbus-send "$@"
}

juhradial_reload_config() {
  juhradial_dbus_send \
    --session \
    --type=method_call \
    --dest=org.kde.juhradialmx \
    /org/kde/juhradialmx/Daemon \
    org.kde.juhradialmx.Daemon.ReloadConfig \
    >/dev/null 2>&1 || true
}

juhradial_reload_macro_triggers() {
  juhradial_dbus_send \
    --session \
    --type=method_call \
    --dest=org.kde.juhradialmx \
    /org/kde/juhradialmx/Daemon \
    org.kde.juhradialmx.Daemon.ReloadMacroTriggers \
    >/dev/null 2>&1 || true
}

juhradial_overlay_script() {
  printf '%s\n' "$(juhradial_install_dir)/overlay/juhradial-overlay.py"
}

juhradial_settings_script() {
  printf '%s\n' "$(juhradial_install_dir)/overlay/settings_dashboard.py"
}

juhradial_desktop_file() {
  printf '%s\n' "${XDG_DATA_HOME:-$HOME/.local/share}/applications/org.juhlabs.JuhRadialMX.desktop"
}

juhradial_config_file() {
  printf '%s\n' "$(juhradial_config_dir)/config.json"
}

juhradial_solaar_policy() {
  printf 'recovery-only\n'
}

juhradial_overlay_running() {
  ps -u "$(id -un)" -o args= 2>/dev/null | grep -F 'juhradial-overlay.py' | grep -Fv 'grep -F' >/dev/null 2>&1
}

juhradial_patch_applied() {
  local src_dir patch_dir patch
  src_dir="$(juhradial_source_dir)"
  patch_dir="$(juhradial_patch_dir)"

  [[ -d "$src_dir/.git" && -d "$patch_dir" ]] || return 1

  while IFS= read -r patch; do
    [[ -n "$patch" ]] || continue
    git -C "$src_dir" apply --reverse --check "$patch" >/dev/null 2>&1 || return 1
  done < <(find "$patch_dir" -maxdepth 1 -type f -name '*.patch' | sort)

  return 0
}

juhradial_config_value() {
  local jq_expr="$1"
  local config_file
  config_file="$(juhradial_config_file)"

  if [[ -f "$config_file" ]] && command -v jq >/dev/null 2>&1; then
    jq -r "$jq_expr" "$config_file" 2>/dev/null
  fi
}

juhradial_export_graphical_env() {
  local line key value

  export XDG_RUNTIME_DIR
  XDG_RUNTIME_DIR="$(juhradial_runtime_dir)"
  export DBUS_SESSION_BUS_ADDRESS
  DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)"

  while IFS= read -r line; do
    [[ "$line" == *=* ]] || continue
    key="${line%%=*}"
    value="${line#*=}"
    case "$key" in
      DISPLAY|WAYLAND_DISPLAY|HYPRLAND_INSTANCE_SIGNATURE|XDG_CURRENT_DESKTOP|XDG_SESSION_TYPE)
        export "$key=$value"
        ;;
    esac
  done < <(juhradial_systemctl show-environment 2>/dev/null || true)
}

juhradial_solaar_timeout() {
  local duration="$1"
  shift
  local config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  local cache_home="${XDG_CACHE_HOME:-$HOME/.cache}"
  local state_home="${XDG_STATE_HOME:-$HOME/.local/state}"

  env \
    HOME="$HOME" \
    XDG_CONFIG_HOME="$config_home" \
    XDG_CACHE_HOME="$cache_home" \
    XDG_STATE_HOME="$state_home" \
    timeout "$duration" \
    solaar "$@"
}

juhradial_bluetooth_mac() {
  if [[ -n "${BT_MX_MASTER:-}" ]]; then
    printf '%s\n' "$BT_MX_MASTER"
    return 0
  fi

  bluetoothctl devices Connected 2>/dev/null | awk '/MX Master 4|MX Master/ {print $2; exit}'
}

juhradial_bluetooth_connected() {
  [[ -n "$(juhradial_bluetooth_mac)" ]]
}

juhradial_bluetooth_battery_status() {
  local mac pct

  mac="$(juhradial_bluetooth_mac)"
  [[ -n "$mac" ]] || return 1

  pct="$(bluetoothctl info "$mac" 2>/dev/null | grep -oP 'Battery Percentage:.*\(\K\d+' | head -n1 || true)"
  [[ -n "$pct" ]] || return 1

  printf '%s false\n' "$pct"
}

juhradial_bolt_receiver_present() {
  local dev props

  if lsusb 2>/dev/null | grep -qiE '046d:c548|046d:c547'; then
    return 0
  fi

  for dev in /dev/hidraw*; do
    [[ -e "$dev" ]] || continue
    props="$(udevadm info -q property -n "$dev" 2>/dev/null || true)"
    grep -q '^ID_VENDOR_ID=046d$' <<<"$props" || continue
    if grep -q '^ID_MODEL_ID=c548$' <<<"$props" || grep -q '^ID_MODEL_FROM_DATABASE=Logi Bolt Receiver$' <<<"$props"; then
      return 0
    fi
  done

  return 1
}

juhradial_gdbus_call() {
  local runtime
  runtime="$(juhradial_runtime_dir)"
  env \
    XDG_RUNTIME_DIR="$runtime" \
    DBUS_SESSION_BUS_ADDRESS="$(juhradial_user_bus)" \
    timeout "${JUHRADIAL_DBUS_TIMEOUT:-5s}" \
    gdbus call \
    --session \
    "$@"
}

juhradial_dbus_battery_status() {
  local out pct charging t1 t2 t3 raw_pct raw_charging
  out="$(
    juhradial_gdbus_call \
      --dest org.kde.juhradialmx \
      --object-path /org/kde/juhradialmx/Daemon \
      --method org.kde.juhradialmx.Daemon.GetBatteryStatus 2>/dev/null
  )" || out=""

  if [[ -n "$out" ]]; then
    out="${out#*(}"
    out="${out%)}"
    out="${out//,/ }"
    read -r t1 t2 t3 <<<"$out"
    if [[ "$t1" =~ ^(byte|uint32|int32)$ ]]; then
      raw_pct="$t2"
      raw_charging="$t3"
    else
      raw_pct="$t1"
      raw_charging="$t2"
    fi

    if [[ "$raw_pct" =~ ^0x[0-9A-Fa-f]+$ ]]; then
      pct="$((16#${raw_pct#0x}))"
    else
      pct="${raw_pct//[^0-9]/}"
    fi
    charging="$(printf '%s' "${raw_charging:-false}" | tr '[:upper:]' '[:lower:]' | tr -cd '[:alpha:]')"
    if [[ -n "$pct" ]] && (( 10#$pct > 0 && 10#$pct <= 100 )); then
      printf '%s %s\n' "$pct" "${charging:-false}"
      return 0
    fi
  fi

  return 1
}

juhradial_transport_state() {
  local bt_connected=false
  local bolt_present=false

  juhradial_bluetooth_connected && bt_connected=true
  juhradial_bolt_receiver_present && bolt_present=true

  if $bt_connected && $bolt_present; then
    printf 'split-brain\n'
  elif $bt_connected; then
    printf 'bluetooth\n'
  elif $bolt_present && juhradial_dbus_battery_status >/dev/null 2>&1; then
    printf 'bolt\n'
  else
    printf 'disconnected\n'
  fi
}

juhradial_battery_status() {
  if juhradial_dbus_battery_status; then
    return 0
  fi

  if juhradial_bluetooth_battery_status; then
    return 0
  fi

  return 1
}
