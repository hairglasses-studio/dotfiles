#!/usr/bin/env bash
# notify.sh — Desktop notification helper for dotfiles scripts
# Source this file: source "$(dirname "$0")/lib/notify.sh"
#
# Usage:
#   hg_notify "Shader" "Switched to bloom-soft"
#   hg_notify_low "Wallpaper" "cyber-rain"
#   hg_notify_critical "Battery" "MX Master 4 low: 15%"

hg_notify() {
  local app="$1" body="$2"
  command -v notify-send &>/dev/null || return 0
  notify-send -a "$app" "$app" "$body" 2>/dev/null || true
}

hg_notify_low() {
  local app="$1" body="$2"
  command -v notify-send &>/dev/null || return 0
  notify-send -u low -a "$app" "$app" "$body" 2>/dev/null || true
}

hg_notify_critical() {
  local app="$1" body="$2"
  command -v notify-send &>/dev/null || return 0
  notify-send -u critical -a "$app" "$app" "$body" 2>/dev/null || true
}
