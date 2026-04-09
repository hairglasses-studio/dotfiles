#!/usr/bin/env bash

juhradial_dotfiles_dir() {
  local lib_dir
  lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  cd "$lib_dir/../.." && pwd
}

juhradial_config_dir() {
  printf '%s\n' "${XDG_CONFIG_HOME:-$HOME/.config}/juhradial"
}

juhradial_seed_dir() {
  printf '%s\n' "$(juhradial_dotfiles_dir)/juhradial"
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

juhradial_overlay_script() {
  printf '%s\n' "$(juhradial_install_dir)/overlay/juhradial-overlay.py"
}

juhradial_settings_script() {
  printf '%s\n' "$(juhradial_install_dir)/overlay/settings_dashboard.py"
}

juhradial_overlay_running() {
  pgrep -f 'juhradial-overlay(\.py)?' >/dev/null 2>&1
}

juhradial_battery_status() {
  local out pct charging t1 t2 t3
  out="$(
    juhradial_gdbus call \
      --session \
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
      pct="$t2"
      charging="$t3"
    else
      pct="$t1"
      charging="$t2"
    fi
    pct="${pct//[^0-9]/}"
    charging="${charging//[^a-z]/}"
    if [[ -n "$pct" ]] && (( 10#$pct > 0 && 10#$pct <= 100 )); then
      printf '%s %s\n' "$pct" "${charging:-false}"
      return 0
    fi
  fi

  pct="$(bluetoothctl info "${BT_MX_MASTER:-D2:8E:C5:DE:9F:CC}" 2>/dev/null | grep -oP 'Battery Percentage:.*\(\K\d+' || true)"
  if [[ -n "$pct" ]]; then
    printf '%s false\n' "$pct"
    return 0
  fi

  return 1
}
