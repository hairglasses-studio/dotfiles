#!/usr/bin/env bash
set -euo pipefail

# shader-build.sh — Validate Kitty DarkWindow shader assets

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SHADER_DIR="$DOTFILES_DIR/kitty/shaders/darkwindow"

source "$DOTFILES_DIR/scripts/lib/hg-core.sh"

mode="${1:-build}"

_count() {
  find "$SHADER_DIR" -maxdepth 1 -type f -name '*.glsl' | wc -l | tr -d ' '
}

case "$mode" in
  --check|check)
    [[ -d "$SHADER_DIR" ]] || hg_die "Shader directory not found: $SHADER_DIR"
    hg_ok "Kitty shader catalog ready ($(_count) DarkWindow shaders)"
    ;;

  build|"")
    [[ -d "$SHADER_DIR" ]] || hg_die "Shader directory not found: $SHADER_DIR"
    hg_ok "No transpilation required; using Kitty DarkWindow shaders directly ($(_count) files)"
    ;;

  clean)
    hg_info "Nothing to clean; DarkWindow shaders are stored canonically in $SHADER_DIR"
    ;;

  *)
    cat <<EOF
Usage: shader-build.sh [command]

Commands:
  build   Confirm that the Kitty DarkWindow shader catalog is ready (default)
  check   Same as build, but intended for validation hooks
  clean   No-op for the direct Kitty shader catalog
EOF
    ;;
esac
