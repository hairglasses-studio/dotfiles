#!/usr/bin/env bash
# Propagate Hairglasses Neon palette to all consumers.
# chezmoi runs this whenever .chezmoidata.toml palette values change.
# hash: {{ include ".chezmoidata.toml" | sha256sum }}
set -euo pipefail

DOTFILES_DIR="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}"
if [[ -x "$DOTFILES_DIR/scripts/palette-propagate.sh" ]]; then
    echo "[chezmoi] Propagating palette..."
    "$DOTFILES_DIR/scripts/palette-propagate.sh" 2>/dev/null || true
fi
