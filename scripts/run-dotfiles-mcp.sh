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

SERVER_ROOT="$REPO_ROOT/mcp/dotfiles-mcp"
BIN_PATH="$REPO_ROOT/.codex/bin/dotfiles-mcp"

needs_build=false
if [[ ! -x "$BIN_PATH" ]]; then
  needs_build=true
elif find "$SERVER_ROOT" -type f \( -name '*.go' -o -name 'go.mod' -o -name 'go.sum' \) -newer "$BIN_PATH" -print -quit | grep -q .; then
  needs_build=true
fi

if [[ "$needs_build" == true ]]; then
  mkdir -p "$(dirname "$BIN_PATH")"
  (
    cd "$SERVER_ROOT"
    env GOWORK=off go build -o "$BIN_PATH" .
  )
fi

exec "$BIN_PATH" "$@"
