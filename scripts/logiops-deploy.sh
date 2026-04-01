#!/usr/bin/env bash
# logiops-deploy.sh — Deploy logiops config to /etc/logid.cfg
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$DOTFILES/logiops/logid.cfg"
DST="/etc/logid.cfg"

if [[ ! -f "$SRC" ]]; then
    echo "Error: Source config not found at $SRC"
    exit 1
fi

echo "Deploying logiops config..."
sudo cp "$SRC" "$DST"
echo "  Copied $SRC -> $DST"

echo "Restarting logid service..."
sudo systemctl restart logid.service
echo "logiops config deployed. MX Master 4 settings active."
