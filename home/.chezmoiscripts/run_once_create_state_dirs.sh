#!/usr/bin/env bash
# Create runtime state directories (runs once per machine)
set -euo pipefail

mkdir -p \
    "$HOME/.local/state/hypr" \
    "$HOME/.local/state/kitty/sessions" \
    "$HOME/.local/state/dotfiles/desktop-control" \
    "$HOME/.local/state/ops"

# Stub monitors.dynamic.conf if missing
dynamic_conf="$HOME/.local/state/hypr/monitors.dynamic.conf"
if [[ ! -f "$dynamic_conf" ]]; then
    echo "# Dynamically managed by hyprdynamicmonitors — do not edit by hand" > "$dynamic_conf"
fi
