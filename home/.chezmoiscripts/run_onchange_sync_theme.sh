#!/usr/bin/env bash
# Sync the Hairglasses Neon theme to all GTK CSS consumers.
# chezmoi runs this whenever .chezmoidata.toml palette values change.
# hash: {{ include ".chezmoidata.toml" | sha256sum }}
set -euo pipefail

DOTFILES_DIR="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}"
if [[ -x "$DOTFILES_DIR/scripts/theme-sync.sh" ]]; then
    echo "[chezmoi] Syncing theme..."
    "$DOTFILES_DIR/scripts/theme-sync.sh" --quiet --no-wallpaper 2>/dev/null || true
fi
