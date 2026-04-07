#!/usr/bin/env bash
# greetd-deploy.sh — Deploy greetd config and /etc/environment
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying greetd config..."
sudo install -Dm644 "$DOTFILES/greetd/config.toml" /etc/greetd/config.toml
echo "  Copied greetd/config.toml -> /etc/greetd/config.toml"

echo "Deploying /etc/environment..."
sudo install -Dm644 "$DOTFILES/greetd/environment" /etc/environment
echo "  Copied greetd/environment -> /etc/environment"

if [[ -d "$DOTFILES/greetd/pam" ]]; then
    echo "Deploying greetd PAM configs..."
    for pam_file in "$DOTFILES/greetd/pam/"*; do
        [[ -f "$pam_file" ]] || continue
        sudo install -Dm644 "$pam_file" "/etc/pam.d/$(basename "$pam_file")"
        echo "  Copied greetd/pam/$(basename "$pam_file") -> /etc/pam.d/$(basename "$pam_file")"
    done
fi

echo "greetd config deployed. Hyprland is now the default session."
echo "Reboot to verify."
