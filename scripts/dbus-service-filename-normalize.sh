#!/usr/bin/env bash
# dbus-service-filename-normalize.sh — quarantine vendor service filenames that
# conflict with canonical D-Bus name-based aliases deployed under /etc.
set -euo pipefail

quiet=false
if [[ "${1:-}" == "--quiet" ]]; then
  quiet=true
fi

misnamed_service_files=(
  /usr/share/dbus-1/services/org.erikreider.swaync.service
  /usr/share/dbus-1/services/org.kde.kscreen.service
  /usr/share/dbus-1/services/org.kde.plasma.Notifications.service
  /usr/share/dbus-1/services/org.xfce.Thunar.FileManager1.service
  /usr/share/dbus-1/services/org.xfce.Tumbler.Cache1.service
  /usr/share/dbus-1/services/org.xfce.Tumbler.Manager1.service
  /usr/share/dbus-1/services/org.xfce.Tumbler.Thumbnailer1.service
)

changed=0
for target in "${misnamed_service_files[@]}"; do
  [[ -f "$target" ]] || continue
  quarantined="${target}.disabled-by-dotfiles"
  mv -f "$target" "$quarantined"
  changed=1
  if ! $quiet; then
    printf 'Quarantined D-Bus service filename: %s -> %s\n' "$target" "$quarantined"
  fi
done

if ! $quiet && [[ "$changed" -eq 0 ]]; then
  printf 'D-Bus service filenames already normalized.\n'
fi
