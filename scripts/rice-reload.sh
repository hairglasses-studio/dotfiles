#!/usr/bin/env bash
set -euo pipefail

# rice-reload.sh — Reload the entire visual stack in one shot
# Keybind: $mod+Ctrl+R

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

reloaded=""
failed=""

# Hyprland config
if hyprctl reload 2>/dev/null; then
  reloaded="${reloaded} hyprland"
else
  failed="${failed} hyprland"
fi

# Ironbar (config + CSS)
if ironbar reload 2>/dev/null; then
  reloaded="${reloaded} ironbar"
else
  failed="${failed} ironbar"
fi

# Kitty (SIGUSR1)
if kill -USR1 "$(pidof kitty)" 2>/dev/null; then
  reloaded="${reloaded} kitty"
else
  failed="${failed} kitty"
fi

# GTK theme refresh (gsettings roundtrip forces re-read)
if command -v gsettings &>/dev/null; then
  current="$(gsettings get org.gnome.desktop.interface gtk-theme 2>/dev/null || true)"
  if [[ -n "$current" ]]; then
    gsettings set org.gnome.desktop.interface gtk-theme "$current" 2>/dev/null
    reloaded="${reloaded} gtk"
  fi
fi

# Swaync notifications
if swaync-client -rs 2>/dev/null; then
  reloaded="${reloaded} swaync"
else
  failed="${failed} swaync"
fi

# Shader — rebuild transpiled shaders if build script exists
if [[ -x "$SCRIPT_DIR/../kitty/shaders/bin/shader-build.sh" ]]; then
  "$SCRIPT_DIR/../kitty/shaders/bin/shader-build.sh" build 2>/dev/null
  reloaded="${reloaded} shaders"
fi

# Report
reloaded="${reloaded# }"
failed="${failed# }"

if [[ -n "$failed" ]]; then
  hg_warn "Reloaded: ${reloaded}  |  Failed: ${failed}"
  hg_notify_low "Rice Reload" "OK: ${reloaded}\nFail: ${failed}"
else
  hg_ok "Reloaded: ${reloaded}"
  hg_notify_low "Rice Reload" "${reloaded}"
fi
