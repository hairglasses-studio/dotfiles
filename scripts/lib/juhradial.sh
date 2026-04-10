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

juhradial_overlay_running() {
  ps -u "$(id -un)" -o args= 2>/dev/null | grep -F 'juhradial-overlay.py' | grep -Fv 'grep -F' >/dev/null 2>&1
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

juhradial_transport_state() {
  local bt_connected=false
  local show=""
  local receiver_has_mx=false
  local standalone_mx=false
  local config_reachable=false

  if bluetoothctl devices Connected 2>/dev/null | grep -q 'MX Master 4'; then
    bt_connected=true
  fi

  if juhradial_solaar_timeout "${JUHRADIAL_SOLAAR_TIMEOUT:-8s}" config 'MX Master 4' >/dev/null 2>&1; then
    config_reachable=true
  fi

  show="$(juhradial_solaar_timeout "${JUHRADIAL_SOLAAR_TIMEOUT:-8s}" show all 2>/dev/null | tr -d '\000' || true)"
  if grep -qE '^  [0-9]+: MX Master 4$' <<<"$show"; then
    receiver_has_mx=true
  fi
  if grep -q '^MX Master 4$' <<<"$show"; then
    standalone_mx=true
  fi

  if $bt_connected && $receiver_has_mx; then
    printf 'split-brain\n'
  elif $bt_connected && $standalone_mx; then
    printf 'bluetooth\n'
  elif ! $bt_connected && $config_reachable; then
    printf 'bolt\n'
  elif ! $bt_connected && $receiver_has_mx && ! $standalone_mx; then
    printf 'bolt\n'
  elif $standalone_mx; then
    printf 'bluetooth\n'
  else
    printf 'disconnected\n'
  fi
}

juhradial_battery_status() {
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

  pct="$(juhradial_solaar_timeout "${JUHRADIAL_SOLAAR_TIMEOUT:-8s}" show 'MX Master 4' 2>/dev/null | tr -d '\000' | grep -oP 'Battery:\s+\K\d+(?=%)' | head -n1 || true)"
  if [[ -n "$pct" ]]; then
    printf '%s false\n' "$pct"
    return 0
  fi

  pct="$(bluetoothctl info "${BT_MX_MASTER:-D2:8E:C5:DE:9F:CC}" 2>/dev/null | grep -oP 'Battery Percentage:.*\(\K\d+' || true)"
  if [[ -n "$pct" ]]; then
    printf '%s false\n' "$pct"
    return 0
  fi

  return 1
}
