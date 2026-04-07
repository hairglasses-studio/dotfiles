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

# modules-load
for f in "$DOTFILES/etc/modules-load.d/"*.conf; do
    [[ -f "$f" ]] || continue
    sudo install -d /etc/modules-load.d
    sudo install -m644 "$f" "/etc/modules-load.d/$(basename "$f")"
    echo "  Deployed modules-load.d/$(basename "$f")"
done

# Bluetooth
if [[ -f "$DOTFILES/etc/bluetooth/main.conf" ]]; then
    sudo install -d /etc/bluetooth
    sudo install -m644 "$DOTFILES/etc/bluetooth/main.conf" /etc/bluetooth/main.conf
    echo "  Deployed bluetooth/main.conf"
fi

echo "/etc/ configs deployed."
