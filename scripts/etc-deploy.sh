#!/usr/bin/env bash
# etc-deploy.sh — Deploy tracked /etc/ configs (sysctl, modprobe, etc.)
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying /etc/ configs..."

# sysctl
if [[ -f "$DOTFILES/etc/sysctl.d/99-workstation.conf" ]]; then
    sudo cp "$DOTFILES/etc/sysctl.d/99-workstation.conf" /etc/sysctl.d/
    sudo sysctl -p /etc/sysctl.d/99-workstation.conf
    echo "  sysctl tuning applied"
fi

# modprobe (NVIDIA)
for f in "$DOTFILES/etc/modprobe.d/"*.conf; do
    [[ -f "$f" ]] || continue
    sudo cp "$f" /etc/modprobe.d/
    echo "  Deployed $(basename "$f")"
done

echo "/etc/ configs deployed."
