#!/usr/bin/env bash
# notification-daemon-launch.sh - D-Bus activation dispatcher for notifications.
#
# The repo owns /etc/dbus-1/services/org.freedesktop.Notifications.service.
# During staged shell migration this launcher keeps swaync as the default owner,
# but routes activation to Quickshell after `hg shell notification-cutover` or
# `hg shell full-cutover` persists QUICKSHELL_NOTIFICATION_OWNER=1.

set -euo pipefail

state_env="${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/shell-stack/env"
systemctl_bin="${SYSTEMCTL_BIN:-systemctl}"
busctl_bin="${BUSCTL_BIN:-busctl}"
gdbus_bin="${GDBUS_BIN:-gdbus}"
swaync_bin="${SWAYNC_BIN:-/usr/bin/swaync}"
wait_attempts="${NOTIFICATION_OWNER_WAIT_ATTEMPTS:-25}"
wait_sleep="${NOTIFICATION_OWNER_WAIT_SLEEP:-0.2}"

if [[ -r "$state_env" ]]; then
  # shellcheck disable=SC1090
  source "$state_env"
fi

notification_owner_seen() {
  if command -v "$busctl_bin" >/dev/null 2>&1; then
    "$busctl_bin" --user --list 2>/dev/null \
      | awk '{print $1}' \
      | grep -qx 'org.freedesktop.Notifications'
    return
  fi

  if command -v "$gdbus_bin" >/dev/null 2>&1; then
    "$gdbus_bin" call \
      --session \
      --dest org.freedesktop.DBus \
      --object-path /org/freedesktop/DBus \
      --method org.freedesktop.DBus.ListNames 2>/dev/null \
      | grep -q "'org.freedesktop.Notifications'"
    return
  fi

  return 1
}

if [[ "${QUICKSHELL_NOTIFICATION_OWNER:-0}" == "1" ]]; then
  "$systemctl_bin" --user start dotfiles-quickshell.service >/dev/null 2>&1 || true

  for ((attempt = 0; attempt < wait_attempts; attempt++)); do
    if notification_owner_seen; then
      exit 0
    fi
    sleep "$wait_sleep"
  done

  if [[ "${DOTFILES_NOTIFICATION_NO_FALLBACK:-0}" == "1" ]]; then
    printf 'notification-daemon-launch: Quickshell did not claim org.freedesktop.Notifications\n' >&2
    exit 1
  fi
fi

exec "$swaync_bin"
