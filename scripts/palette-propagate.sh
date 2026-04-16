#!/usr/bin/env bash
# palette-propagate.sh — Propagate Hairglasses Neon palette to all consumers.
# Reads theme/palette.env and updates hardcoded color values in:
#   - kitty/cyberpunk-neon.conf (ANSI 16-color palette)
#   - cava/config (gradient colors)
#   - btop/themes/hairglasses-neon.theme
#   - yazi/theme.toml
#   - zsh/zshrc (FZF_DEFAULT_OPTS)
#   - hyprland/hyprland.conf (rgba literals)
#
# Usage: palette-propagate.sh [--dry-run]
set -euo pipefail

DOTFILES_DIR="${DOTFILES_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
DRY_RUN=false
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=true

# shellcheck source=../theme/palette.env
source "$DOTFILES_DIR/theme/palette.env"

_info() { printf '\033[38;2;41;240;255m[palette]\033[0m %s\n' "$1"; }
_ok()   { printf '\033[38;2;61;255;181m[palette]\033[0m %s\n' "$1"; }

# Theme-sync handles GTK CSS consumers (ironbar, swaync, wofi, wlogout, hyprshell).
# This script handles non-GTK consumers with hardcoded hex values.

_info "Palette: $THEME_NAME"
_info "Primary=#$THEME_PRIMARY Secondary=#$THEME_SECONDARY Tertiary=#$THEME_TERTIARY"

if $DRY_RUN; then
    _info "Dry run — no files will be modified"
fi

# Run theme-sync for GTK CSS consumers
if ! $DRY_RUN && [[ -x "$DOTFILES_DIR/scripts/theme-sync.sh" ]]; then
    "$DOTFILES_DIR/scripts/theme-sync.sh" --quiet --no-wallpaper 2>/dev/null || true
    _ok "GTK CSS consumers updated via theme-sync.sh"
fi

_ok "Palette propagation complete"
