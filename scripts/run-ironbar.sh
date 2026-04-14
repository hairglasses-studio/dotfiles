#!/usr/bin/env bash
set -euo pipefail

runtime_dir_default="/run/user/$(id -u)"
wait_secs="${IRONBAR_WAIT_SECS:-15}"
mode="${1:-run}"

resolve_wayland_display_for_dir() {
  local runtime_dir="$1"
  [[ -d "$runtime_dir" ]] || return 0

  mapfile -t wayland_sockets < <(find "$runtime_dir" -maxdepth 1 -type s -name 'wayland-*' -printf '%f\n' 2>/dev/null | sort)
  [[ "${#wayland_sockets[@]}" -gt 0 ]] || return 0

  for preferred in "${WAYLAND_DISPLAY:-}" wayland-1 wayland-0; do
    [[ -n "$preferred" ]] || continue
    for socket in "${wayland_sockets[@]}"; do
      if [[ "$socket" == "$preferred" ]]; then
        printf '%s\n' "$socket"
        return 0
      fi
    done
  done

  printf '%s\n' "${wayland_sockets[0]}"
}

resolve_hyprland_signature_for_dir() {
  local runtime_dir="$1"
  local hypr_dir="$runtime_dir/hypr"
  [[ -d "$hypr_dir" ]] || return 0

  find "$hypr_dir" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' 2>/dev/null | sort | head -1 || true
}

resolve_runtime_dir() {
  if [[ -n "${XDG_RUNTIME_DIR:-}" && -d "${XDG_RUNTIME_DIR}" ]]; then
    printf '%s\n' "${XDG_RUNTIME_DIR}"
    return 0
  fi

  if [[ -d "$runtime_dir_default" ]]; then
    printf '%s\n' "$runtime_dir_default"
  fi
}

refresh_runtime_env() {
  local resolved_runtime resolved_wayland resolved_hypr

  resolved_runtime="$(resolve_runtime_dir || true)"
  if [[ -n "$resolved_runtime" ]]; then
    export XDG_RUNTIME_DIR="$resolved_runtime"
  fi

  resolved_wayland="$(resolve_wayland_display_for_dir "${XDG_RUNTIME_DIR:-$runtime_dir_default}" || true)"
  if [[ -n "$resolved_wayland" ]]; then
    export WAYLAND_DISPLAY="$resolved_wayland"
  fi

  if [[ -z "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
    resolved_hypr="$(resolve_hyprland_signature_for_dir "${XDG_RUNTIME_DIR:-$runtime_dir_default}" || true)"
    if [[ -n "$resolved_hypr" ]]; then
      export HYPRLAND_INSTANCE_SIGNATURE="$resolved_hypr"
    fi
  fi

  if [[ -z "${DBUS_SESSION_BUS_ADDRESS:-}" && -S "${XDG_RUNTIME_DIR:-}/bus" ]]; then
    export DBUS_SESSION_BUS_ADDRESS="unix:path=${XDG_RUNTIME_DIR}/bus"
  fi
}

wayland_socket_ready() {
  [[ -n "${XDG_RUNTIME_DIR:-}" ]] || return 1
  [[ -n "${WAYLAND_DISPLAY:-}" ]] || return 1
  [[ -S "${XDG_RUNTIME_DIR%/}/${WAYLAND_DISPLAY}" ]]
}

wait_for_wayland() {
  local waited=0
  while (( waited <= wait_secs )); do
    refresh_runtime_env
    if wayland_socket_ready; then
      return 0
    fi
    sleep 1
    ((waited += 1))
  done
  return 1
}

print_env() {
  printf 'XDG_RUNTIME_DIR=%s\n' "${XDG_RUNTIME_DIR:-}"
  printf 'WAYLAND_DISPLAY=%s\n' "${WAYLAND_DISPLAY:-}"
  printf 'HYPRLAND_INSTANCE_SIGNATURE=%s\n' "${HYPRLAND_INSTANCE_SIGNATURE:-}"
  printf 'DBUS_SESSION_BUS_ADDRESS=%s\n' "${DBUS_SESSION_BUS_ADDRESS:-}"
}

case "$mode" in
  --print-env)
    refresh_runtime_env
    print_env
    exit 0
    ;;
  --check)
    if wait_for_wayland; then
      print_env
      exit 0
    fi
    printf 'run-ironbar: no live Wayland socket after %ss\n' "$wait_secs" >&2
    exit 1
    ;;
  run|--run)
    ;;
  *)
    printf 'Usage: %s [--check|--print-env]\n' "${0##*/}" >&2
    exit 2
    ;;
esac

if ! wait_for_wayland; then
  printf 'run-ironbar: no live Wayland socket after %ss\n' "$wait_secs" >&2
  exit 1
fi

exec /usr/bin/env ironbar \
  -c "${IRONBAR_CONFIG_PATH:-$HOME/.config/ironbar/config.toml}" \
  -t "${IRONBAR_STYLE_PATH:-$HOME/.config/ironbar/style.css}"
