#!/usr/bin/env bash
# refind-kernel-sync.sh — Validate rEFInd loader paths after kernel install/remove
# Deployed by dotfiles to catch stale kernel references
set -euo pipefail

REFIND_CONF="/boot/efi/EFI/refind/refind.conf"
LOG="/var/log/refind-boot-guard.log"
log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG"; }

log "Kernel change detected — validating rEFInd config"

while IFS= read -r line; do
    # Strip the /@/ prefix to get the actual filesystem path
    kernel="${line#*/@}"
    kernel="/${kernel}"
    if [[ ! -f "$kernel" ]]; then
        log "WARNING: rEFInd references $kernel but file not found!"
        echo "WARNING: rEFInd boot entry references missing kernel: $kernel" >&2
    else
        log "OK: $kernel exists"
    fi
done < <(grep '^\s*loader\s' "$REFIND_CONF" | sed 's/^\s*loader\s*//')
