#!/usr/bin/env bash
# refind-deploy.sh — Deploy rEFInd Matrix theme to EFI partition
# Requires: sudo, mounted /boot/efi
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
REFIND_DIR="/boot/efi/EFI/refind"

if [[ ! -d "$REFIND_DIR" ]]; then
    echo "Error: rEFInd not found at $REFIND_DIR"
    echo "Is the EFI partition mounted? Try: sudo mount /boot/efi"
    exit 1
fi

echo "Deploying rEFInd Matrix theme..."
sudo cp "$DOTFILES/refind/refind.conf" "$REFIND_DIR/refind.conf"
sudo cp "$DOTFILES/refind/refind_linux.conf" "$REFIND_DIR/refind_linux.conf"
sudo mkdir -p "$REFIND_DIR/themes/matrix/icons"
sudo cp -r "$DOTFILES/refind/themes/matrix/"* "$REFIND_DIR/themes/matrix/"

echo "rEFInd Matrix theme deployed to $REFIND_DIR"
echo "Reboot to see changes."
