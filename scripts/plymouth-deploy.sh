#!/usr/bin/env bash
# plymouth-deploy.sh — Install Cybernet Plymouth theme + rebuild initramfs
# Requires: sudo, plymouth, mkinitcpio
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
THEME_DIR="/usr/share/plymouth/themes/cybernet"

if ! command -v plymouth-set-default-theme &>/dev/null; then
    echo "Error: plymouth not installed"
    echo "Install: sudo pacman -S plymouth"
    exit 1
fi

echo "Installing Cybernet Plymouth theme..."
sudo mkdir -p "$THEME_DIR"
sudo cp "$DOTFILES/plymouth/cybernet/"* "$THEME_DIR/"

echo "Setting as default theme..."
sudo plymouth-set-default-theme cybernet

echo "Rebuilding initramfs (this may take a moment)..."
sudo mkinitcpio -P

echo "Plymouth theme 'cybernet' installed."
echo "Ensure 'plymouth' is in HOOKS in /etc/mkinitcpio.conf"
echo "Kernel params should include: quiet splash"
