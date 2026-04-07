#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

export WAYLAND_DISPLAY="${WAYLAND_DISPLAY:-wayland-1}"

if [[ -z "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
  hypr_dir="/run/user/$(id -u)/hypr"
  if [[ -d "$hypr_dir" ]]; then
    first_sig="$(ls "$hypr_dir" 2>/dev/null | head -1 || true)"
    if [[ -n "$first_sig" ]]; then
      export HYPRLAND_INSTANCE_SIGNATURE="$first_sig"
    fi
  fi
fi

export DOTFILES_DIR="${DOTFILES_DIR:-$REPO_ROOT}"

cd "$REPO_ROOT/mcp/dotfiles-mcp"
exec go run .
