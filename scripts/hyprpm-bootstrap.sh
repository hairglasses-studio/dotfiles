#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

QUIET=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --quiet) QUIET=true ;;
    *) hg_die "Usage: hyprpm-bootstrap.sh [--quiet]" ;;
  esac
  shift
done

command -v hyprpm >/dev/null 2>&1 || {
  $QUIET || hg_warn "Skipping hyprpm bootstrap: hyprpm not installed"
  exit 0
}

if [[ "${EUID:-$(id -u)}" -eq 0 ]]; then
  $QUIET || hg_warn "Skipping hyprpm bootstrap: run as the desktop user, not root"
  exit 0
fi

declare -a repos=(
  "hyprland-plugins|https://github.com/hyprwm/hyprland-plugins|borders-plus-plus hyprbars hyprexpo hyprfocus hyprwinwrap"
  "split-monitor-workspaces|https://github.com/Duckonaut/split-monitor-workspaces|split-monitor-workspaces"
  "dynamic-cursors|https://github.com/VirtCode/hypr-dynamic-cursors|dynamic-cursors"
)

list_output="$(hyprpm list 2>/dev/null || true)"

for entry in "${repos[@]}"; do
  name="${entry%%|*}"
  rest="${entry#*|}"
  url="${rest%%|*}"
  plugins="${rest##*|}"

  if ! printf '%s\n' "$list_output" | grep -Fq "Repository ${name} "; then
    $QUIET || hg_info "Adding hyprpm repository: ${name}"
    hyprpm add "$url" >/dev/null 2>&1 || true
    list_output="$(hyprpm list 2>/dev/null || true)"
  fi

  for plugin in $plugins; do
    if ! printf '%s\n' "$list_output" | grep -Fq "Plugin ${plugin}"; then
      continue
    fi
    $QUIET || hg_info "Ensuring hyprpm plugin is enabled: ${plugin}"
    hyprpm enable "$plugin" >/dev/null 2>&1 || true
  done
done

if [[ -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]]; then
  hyprpm reload -n >/dev/null 2>&1 || true
fi

if ! $QUIET; then
  hg_ok "Hyprland plugins bootstrapped"
fi
