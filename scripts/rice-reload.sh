#!/usr/bin/env bash
set -euo pipefail

# rice-reload.sh — Reload the entire visual stack in one shot
# Keybind: $mod+Ctrl+R

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/compositor.sh"
source "$SCRIPT_DIR/lib/config.sh"
source "$SCRIPT_DIR/lib/notify.sh"
source "$SCRIPT_DIR/lib/shell-stack.sh"

shell_stack_load
reloaded=""
failed=""
skipped=""

if [[ -x "$SCRIPT_DIR/theme-sync.sh" ]]; then
  if "$SCRIPT_DIR/theme-sync.sh" --quiet 2>/dev/null; then
    reloaded="${reloaded} theme"
  else
    failed="${failed} theme"
  fi
fi

# Hyprland config
if compositor_reload 2>/dev/null; then
  reloaded="${reloaded} hyprland"
else
  failed="${failed} hyprland"
fi

# Ironbar menubar
if shell_stack_bar_cutover; then
  skipped="${skipped} ironbar"
else
  if command -v ironbar >/dev/null 2>&1; then
    if config_reload_service ironbar --quiet 2>/dev/null; then
      reloaded="${reloaded} ironbar"
    else
      failed="${failed} ironbar"
    fi
  else
    failed="${failed} ironbar"
  fi
fi

# Kitty (SIGUSR1)
if pidof kitty >/dev/null 2>&1 && pidof kitty | xargs -r kill -USR1 >/dev/null 2>&1; then
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
if shell_stack_notification_cutover; then
  skipped="${skipped} swaync"
else
  if swaync-client -rs 2>/dev/null; then
    reloaded="${reloaded} swaync"
  else
    failed="${failed} swaync"
  fi
fi

# Quickshell pilot/cutover bar
if shell_stack_quickshell_wanted || systemctl --user is-active dotfiles-quickshell.service >/dev/null 2>&1; then
  if systemctl --user restart dotfiles-quickshell.service 2>/dev/null; then
    reloaded="${reloaded} quickshell"
  else
    failed="${failed} quickshell"
  fi
fi

# Keybind ticker — restart to pick up palette changes
if shell_stack_ticker_cutover; then
  skipped="${skipped} ticker"
else
  if systemctl --user is-active dotfiles-keybind-ticker.service >/dev/null 2>&1; then
    if systemctl --user restart dotfiles-keybind-ticker.service 2>/dev/null; then
      reloaded="${reloaded} ticker"
    else
      failed="${failed} ticker"
    fi
  fi
fi

# Hyprland companion services
for component in \
  hyprshell \
  hypr-dock \
  hyprdynamicmonitors \
  autoname
do
  if config_reload_service "$component" --quiet 2>/dev/null; then
    reloaded="${reloaded} ${component}"
  else
    failed="${failed} ${component}"
  fi
done

# Claude Code — clear stale state, reset ironvars
_claude_state="${XDG_STATE_HOME:-$HOME/.local/state}/claude"
if [[ -d "$_claude_state" ]]; then
  rm -f "$_claude_state"/git-* "$_claude_state/burn-rate" 2>/dev/null
  command -v ironbar &>/dev/null && ironbar var set claude_state "" 2>/dev/null && ironbar var set claude_cost "" 2>/dev/null
  reloaded="${reloaded} claude"
fi

# Shader — rebuild transpiled shaders if build script exists
if [[ -x "$SCRIPT_DIR/../kitty/shaders/bin/shader-build.sh" ]]; then
  "$SCRIPT_DIR/../kitty/shaders/bin/shader-build.sh" build 2>/dev/null
  reloaded="${reloaded} shaders"
fi

# Report
reloaded="${reloaded# }"
failed="${failed# }"
skipped="${skipped# }"

if [[ -n "$failed" ]]; then
  if [[ -n "$skipped" ]]; then
    hg_warn "Reloaded: ${reloaded}  |  Failed: ${failed}  |  Skipped: ${skipped}"
    hg_notify_low "Rice Reload" "OK: ${reloaded}\nFail: ${failed}\nSkip: ${skipped}"
  else
    hg_warn "Reloaded: ${reloaded}  |  Failed: ${failed}"
    hg_notify_low "Rice Reload" "OK: ${reloaded}\nFail: ${failed}"
  fi
else
  if [[ -n "$skipped" ]]; then
    hg_ok "Reloaded: ${reloaded}  |  Skipped: ${skipped}"
    hg_notify_low "Rice Reload" "${reloaded}\nSkip: ${skipped}"
  else
    hg_ok "Reloaded: ${reloaded}"
    hg_notify_low "Rice Reload" "${reloaded}"
  fi
fi
