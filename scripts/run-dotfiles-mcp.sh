#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

runtime_dir_default="/run/user/$(id -u)"
runtime_scan_root="${DOTFILES_RUNTIME_SCAN_ROOT:-/run/user}"

resolve_wayland_display_for_dir() {
  local runtime_dir="$1"
  [[ -d "$runtime_dir" ]] || return 0

  mapfile -t wayland_sockets < <(find "$runtime_dir" -maxdepth 1 -type s -name 'wayland-*' -printf '%f\n' 2>/dev/null | sort)
  [[ "${#wayland_sockets[@]}" -gt 0 ]] || return 0

  for preferred in wayland-1 wayland-0; do
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

runtime_dir_has_live_desktop() {
  local runtime_dir="$1"
  [[ -n "$(resolve_wayland_display_for_dir "$runtime_dir")" || -n "$(resolve_hyprland_signature_for_dir "$runtime_dir")" ]]
}

resolve_runtime_dir() {
  if [[ -n "${XDG_RUNTIME_DIR:-}" && -d "${XDG_RUNTIME_DIR}" ]] && runtime_dir_has_live_desktop "${XDG_RUNTIME_DIR}"; then
    printf '%s\n' "${XDG_RUNTIME_DIR}"
    return 0
  fi

  local best=""
  local candidate uid
  while IFS= read -r candidate; do
    [[ -d "$candidate" ]] || continue
    if ! runtime_dir_has_live_desktop "$candidate"; then
      continue
    fi
    uid="${candidate##*/}"
    if [[ "$uid" != "0" ]]; then
      printf '%s\n' "$candidate"
      return 0
    fi
    [[ -z "$best" ]] && best="$candidate"
  done < <(find "$runtime_scan_root" -mindepth 1 -maxdepth 1 -type d -printf '%p\n' 2>/dev/null | sort)

  if [[ -n "$best" ]]; then
    printf '%s\n' "$best"
    return 0
  fi

  if [[ -d "$runtime_dir_default" ]]; then
    printf '%s\n' "$runtime_dir_default"
  fi
}

resolved_runtime_dir="$(resolve_runtime_dir || true)"
if [[ -n "$resolved_runtime_dir" ]]; then
  export XDG_RUNTIME_DIR="$resolved_runtime_dir"
fi

resolve_wayland_display() {
  resolve_wayland_display_for_dir "${XDG_RUNTIME_DIR:-$runtime_dir_default}"
}

resolve_hyprland_signature() {
  resolve_hyprland_signature_for_dir "${XDG_RUNTIME_DIR:-$runtime_dir_default}"
}

if [[ -z "${WAYLAND_DISPLAY:-}" ]]; then
  resolved_wayland_display="$(resolve_wayland_display || true)"
  if [[ -n "$resolved_wayland_display" ]]; then
    export WAYLAND_DISPLAY="$resolved_wayland_display"
  fi
fi

if [[ -z "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
  resolved_hyprland_signature="$(resolve_hyprland_signature || true)"
  if [[ -n "$resolved_hyprland_signature" ]]; then
    export HYPRLAND_INSTANCE_SIGNATURE="$resolved_hyprland_signature"
  fi
fi

if [[ -z "${DBUS_SESSION_BUS_ADDRESS:-}" && -S "${XDG_RUNTIME_DIR:-}/bus" ]]; then
  export DBUS_SESSION_BUS_ADDRESS="unix:path=${XDG_RUNTIME_DIR}/bus"
fi

export DOTFILES_DIR="${DOTFILES_DIR:-$REPO_ROOT}"

cd "$REPO_ROOT/mcp/dotfiles-mcp"
exec env GOWORK=off go run .
