#!/usr/bin/env bash
# refind-deploy.sh — Full rEFInd deployment: config, theme, hooks, boot order
# Requires: sudo, mounted /boot/efi, efibootmgr
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
REFIND_DIR="/boot/efi/EFI/refind"
BOOT_DIR="/boot/efi/EFI/Boot"
HOOKS_DIR="/etc/pacman.d/hooks"

red()   { printf '\033[0;31m%s\033[0m\n' "$*"; }
green() { printf '\033[0;32m%s\033[0m\n' "$*"; }
cyan()  { printf '\033[0;36m%s\033[0m\n' "$*"; }

die() { red "ERROR: $*"; exit 1; }

# ── Preflight ────────────────────────────────────────
cyan "=== rEFInd Deployment ==="

mountpoint -q /boot/efi || die "ESP not mounted at /boot/efi"
[[ -d "$REFIND_DIR" ]]  || die "rEFInd not found at $REFIND_DIR"
command -v efibootmgr &>/dev/null || die "efibootmgr not installed"

ESP_USE=$(df --output=pcent /boot/efi | tail -1 | tr -d ' %')
if (( ESP_USE > 90 )); then
    red "WARNING: ESP is ${ESP_USE}% full — consider cleaning up first"
fi

# ── Phase 1: Deploy rEFInd config ────────────────────
cyan "Deploying rEFInd config..."
sudo cp "$DOTFILES/refind/refind.conf" "$REFIND_DIR/refind.conf"
sudo cp "$DOTFILES/refind/refind_linux.conf" "$REFIND_DIR/refind_linux.conf"
green "  Config deployed"

# ── Phase 2: Deploy Matrix theme (without .git) ─────
cyan "Deploying Matrix theme..."
sudo mkdir -p "$REFIND_DIR/themes/matrix/icons"
sudo rsync -rlpt --delete --exclude='.git' \
    "$DOTFILES/refind/themes/matrix/" "$REFIND_DIR/themes/matrix/"
green "  Theme deployed"

# ── Phase 3: Deploy pacman hooks ─────────────────────
cyan "Deploying pacman hooks..."
sudo mkdir -p "$HOOKS_DIR"
sudo cp "$DOTFILES/etc/pacman.d/hooks/98-install-grub.hook" "$HOOKS_DIR/"
sudo cp "$DOTFILES/etc/pacman.d/hooks/99-refind-update.hook" "$HOOKS_DIR/"
sudo cp "$DOTFILES/scripts/refind-boot-guard.sh" /usr/local/bin/
sudo cp "$DOTFILES/scripts/refind-kernel-sync.sh" /usr/local/bin/
sudo chmod +x /usr/local/bin/refind-boot-guard.sh /usr/local/bin/refind-kernel-sync.sh
green "  Hooks deployed"

# ── Phase 4: Set fallback bootloader ─────────────────
cyan "Setting fallback bootloader to rEFInd..."
sudo cp "$REFIND_DIR/refind_x64.efi" "$BOOT_DIR/bootx64.efi"
sudo cp "$DOTFILES/refind/refind-fallback.conf" "$BOOT_DIR/refind.conf"
green "  Fallback set"

# ── Phase 5: Set boot order ─────────────────────────
cyan "Setting UEFI boot order..."
REFIND_BOOTNUM=$(efibootmgr | grep -i "rEFInd" | grep -oP 'Boot\K[0-9A-Fa-f]+')
if [[ -n "$REFIND_BOOTNUM" ]]; then
    CURRENT_ORDER=$(efibootmgr | grep "^BootOrder:" | awk '{print $2}')
    NEW_ORDER="$REFIND_BOOTNUM"
    IFS=',' read -ra ENTRIES <<< "$CURRENT_ORDER"
    for entry in "${ENTRIES[@]}"; do
        [[ "$entry" != "$REFIND_BOOTNUM" ]] && NEW_ORDER="$NEW_ORDER,$entry"
    done
    sudo efibootmgr -o "$NEW_ORDER" > /dev/null
    green "  Boot order: $NEW_ORDER"
else
    red "  WARNING: rEFInd entry not found — boot order unchanged"
fi

# ── Phase 6: Validate ───────────────────────────────
cyan "Validating..."
ESP_USE=$(df --output=pcent /boot/efi | tail -1 | tr -d ' %')
ORDER=$(efibootmgr | grep "^BootOrder:" | awk '{print $2}')
FALLBACK_SIZE=$(stat -c%s "$BOOT_DIR/bootx64.efi" 2>/dev/null || echo 0)
REFIND_SIZE=$(stat -c%s "$REFIND_DIR/refind_x64.efi" 2>/dev/null || echo 0)

echo "  ESP usage:  ${ESP_USE}%"
echo "  Boot order: $ORDER"

if [[ "$FALLBACK_SIZE" == "$REFIND_SIZE" ]]; then
    green "  Fallback:   rEFInd (size match)"
else
    red "  Fallback:   MISMATCH (${FALLBACK_SIZE} != ${REFIND_SIZE})"
fi

if grep -q 'loader.*/@/boot' "$REFIND_DIR/refind.conf"; then
    green "  Btrfs paths: /@/ prefix present"
else
    red "  Btrfs paths: MISSING /@/ prefix!"
fi

green "=== rEFInd deployment complete ==="
echo "Reboot to test. Emergency: F12/F8 at POST → select Manjaro (GRUB)"
