#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

runtime_dir_default="/run/user/$(id -u)"
if [[ -z "${XDG_RUNTIME_DIR:-}" && -d "$runtime_dir_default" ]]; then
  export XDG_RUNTIME_DIR="$runtime_dir_default"
fi

resolve_wayland_display() {
  local runtime_dir="${XDG_RUNTIME_DIR:-$runtime_dir_default}"
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

resolve_hyprland_signature() {
  local runtime_dir="${XDG_RUNTIME_DIR:-$runtime_dir_default}"
  local hypr_dir="$runtime_dir/hypr"
  [[ -d "$hypr_dir" ]] || return 0

  find "$hypr_dir" -mindepth 1 -maxdepth 1 -type d -printf '%f\n' 2>/dev/null | sort | head -1 || true
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

export DOTFILES_DIR="${DOTFILES_DIR:-$REPO_ROOT}"

cd "$REPO_ROOT/mcp/dotfiles-mcp"
exec env GOWORK=off go run .
