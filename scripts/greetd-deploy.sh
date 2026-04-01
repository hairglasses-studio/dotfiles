#!/usr/bin/env bash
# greetd-deploy.sh — Deploy greetd config and /etc/environment
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"

echo "Deploying greetd config..."
sudo cp "$DOTFILES/greetd/config.toml" /etc/greetd/config.toml
echo "  Copied greetd/config.toml -> /etc/greetd/config.toml"

echo "Deploying /etc/environment..."
sudo cp "$DOTFILES/greetd/environment" /etc/environment
echo "  Copied greetd/environment -> /etc/environment"

echo "greetd config deployed. Hyprland is now the default session."
echo "Reboot to verify."
