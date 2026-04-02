#!/usr/bin/env bash
# refind-boot-guard.sh — Runs grub-install then restores rEFInd as first boot entry
# Deployed by dotfiles to survive pacman -Syu
set -euo pipefail

LOG="/var/log/refind-boot-guard.log"
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG"; }

log "GRUB upgrade detected — running grub-install then restoring rEFInd"

# Let GRUB install itself (keeps it functional as fallback)
/usr/bin/install-grub 2>&1 | tee -a "$LOG" || log "WARNING: install-grub failed"

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
