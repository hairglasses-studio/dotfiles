#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

resolved_runtime_dir="$(resolve_runtime_dir || true)"
if [[ -n "$resolved_runtime_dir" ]]; then
  export XDG_RUNTIME_DIR="$resolved_runtime_dir"
fi

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
