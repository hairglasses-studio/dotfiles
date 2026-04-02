#!/usr/bin/env bash
# refind-boot-guard.sh — Runs grub-install then restores rEFInd as first boot entry
# Deployed by dotfiles to survive pacman -Syu
set -euo pipefail

LOG="/var/log/refind-boot-guard.log"
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG"; }

log "GRUB upgrade detected — running grub-install then restoring rEFInd"

# Let GRUB install itself (keeps it functional as fallback)
/usr/bin/install-grub 2>&1 | tee -a "$LOG" || log "WARNING: install-grub failed"

# Verify grubx64.efi is actually GRUB, not rEFInd (previous installs replaced it)
GRUB_EFI="/boot/efi/EFI/Manjaro/grubx64.efi"
REFIND_EFI="/boot/efi/EFI/refind/refind_x64.efi"
if [[ -f "$GRUB_EFI" && -f "$REFIND_EFI" ]]; then
    GRUB_SIZE=$(stat -c%s "$GRUB_EFI")
    REFIND_SIZE=$(stat -c%s "$REFIND_EFI")
    if [[ "$GRUB_SIZE" == "$REFIND_SIZE" ]]; then
        log "CRITICAL: grubx64.efi is rEFInd (size match: $GRUB_SIZE). Checking for backup..."
        if [[ -f "${GRUB_EFI}.bak" ]]; then
            cp "${GRUB_EFI}.bak" "$GRUB_EFI"
            log "Restored GRUB from backup"
        else
            log "WARNING: No grubx64.efi.bak found — GRUB may be broken as fallback"
        fi
    fi
fi

# Restore rEFInd as first boot entry
REFIND_BOOTNUM=$(efibootmgr | grep -i "rEFInd" | grep -oP 'Boot\K[0-9A-Fa-f]+')
if [[ -n "$REFIND_BOOTNUM" ]]; then
    CURRENT_ORDER=$(efibootmgr | grep "^BootOrder:" | awk '{print $2}')
    NEW_ORDER="$REFIND_BOOTNUM"
    IFS=',' read -ra ENTRIES <<< "$CURRENT_ORDER"
    for entry in "${ENTRIES[@]}"; do
        [[ "$entry" != "$REFIND_BOOTNUM" ]] && NEW_ORDER="$NEW_ORDER,$entry"
    done
    efibootmgr -o "$NEW_ORDER" >> "$LOG" 2>&1
    log "Boot order restored: $NEW_ORDER"
else
    log "WARNING: rEFInd boot entry not found in efibootmgr — boot order NOT restored"
fi
