#!/usr/bin/env bash
# etc-deploy.sh — Deploy tracked /etc/ configs (sysctl, modprobe, udev, etc.)
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
TRIGGER_USB=false

usage() {
    cat <<'EOF'
Usage: scripts/etc-deploy.sh [--trigger-usb]

  --trigger-usb   Force a broad USB udev trigger after rule updates.
                  Skip by default to avoid disruptive device churn.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --trigger-usb) TRIGGER_USB=true ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            printf 'Unknown option: %s\n' "$1" >&2
            usage >&2
            exit 2
            ;;
    esac
    shift
done

declare -a FOLLOW_UP_NOTES=()

note_follow_up() {
    FOLLOW_UP_NOTES+=("$1")
}

deploy_file_if_changed() {
    local src="$1" target="$2" label="$3" mode="${4:-644}"
    sudo install -d "$(dirname "$target")"
    if sudo test -f "$target" && sudo cmp -s "$src" "$target"; then
        return 1
    fi
    sudo install -m"$mode" "$src" "$target"
    echo "  Deployed $label"
    return 0
}

remove_file_if_present() {
    local target="$1" label="$2"
    if sudo test -e "$target"; then
        sudo rm -f "$target"
        echo "  Removed stale $label"
        return 0
    fi
    return 1
}

quarantine_dbus_service_file() {
    local target="$1" reason="$2" quarantined="${1}.disabled-by-dotfiles"
    if sudo test -f "$target"; then
        sudo mv "$target" "$quarantined"
        echo "  Quarantined D-Bus service file $(basename "$target") ($reason)"
        return 0
    fi
    return 1
}

deploy_tree() {
    local src_root="$1" dst_root="$2" label="$3"
    local changed=1
    [[ -d "$src_root" ]] || return 1

    while IFS= read -r -d '' f; do
        local rel="${f#"$src_root/"}"
        local target="$dst_root/$rel"
        if deploy_file_if_changed "$f" "$target" "$label/$rel"; then
            changed=0
        fi
    done < <(find "$src_root" -type f -print0 | sort -z)

    return "$changed"
}

systemd_unit_exists() {
    local unit="$1"
    local load_state
    load_state="$(sudo systemctl show --property=LoadState --value "$unit" 2>/dev/null || true)"
    [[ -n "$load_state" && "$load_state" != "not-found" ]]
}

echo "Deploying /etc/ configs..."

any_changes=false
sysctl_changed=false
modprobe_changed=false
modules_changed=false
bluetooth_changed=false
networkmanager_changed=false
udev_changed=false
systemd_changed=false
resolved_changed=false
dbus_changed=false
openlinkhub_changed=false
dbus_pacman_hook_changed=false

# sysctl
if [[ -f "$DOTFILES/etc/sysctl.d/99-workstation.conf" ]]; then
    if deploy_file_if_changed "$DOTFILES/etc/sysctl.d/99-workstation.conf" /etc/sysctl.d/99-workstation.conf "sysctl.d/99-workstation.conf"; then
        any_changes=true
        sysctl_changed=true
    fi
fi
if $sysctl_changed; then
    sudo sysctl -p /etc/sysctl.d/99-workstation.conf >/dev/null
    echo "  sysctl tuning applied"
fi

# modprobe (NVIDIA)
for f in "$DOTFILES/etc/modprobe.d/"*.conf; do
    [[ -f "$f" ]] || continue
    if deploy_file_if_changed "$f" "/etc/modprobe.d/$(basename "$f")" "modprobe.d/$(basename "$f")"; then
        any_changes=true
        modprobe_changed=true
    fi
done

if [[ ! -f "$DOTFILES/etc/modprobe.d/blacklist-k10temp.conf" ]]; then
    if remove_file_if_present /etc/modprobe.d/blacklist-k10temp.conf "modprobe.d/blacklist-k10temp.conf"; then
        any_changes=true
        modprobe_changed=true
    fi
fi
if $modprobe_changed; then
    note_follow_up "modprobe config changed; reboot or reload affected kernel modules when safe"
fi

# modules-load
for f in "$DOTFILES/etc/modules-load.d/"*.conf; do
    [[ -f "$f" ]] || continue
    if deploy_file_if_changed "$f" "/etc/modules-load.d/$(basename "$f")" "modules-load.d/$(basename "$f")"; then
        any_changes=true
        modules_changed=true
    fi
done
if $modules_changed; then
    note_follow_up "modules-load config changed; reboot or manually modprobe new modules when safe"
fi

# Bluetooth
if [[ -f "$DOTFILES/etc/bluetooth/main.conf" ]]; then
    bluetooth_dir_mode="$(stat -c '%a' /etc/bluetooth 2>/dev/null || true)"
    sudo install -d -m 0555 /etc/bluetooth
    if [[ "$bluetooth_dir_mode" != "555" ]]; then
        any_changes=true
        echo "  Normalized /etc/bluetooth directory permissions"
    fi
    if deploy_file_if_changed "$DOTFILES/etc/bluetooth/main.conf" /etc/bluetooth/main.conf "bluetooth/main.conf"; then
        any_changes=true
        bluetooth_changed=true
    fi
fi
if $bluetooth_changed; then
    note_follow_up "bluetooth config changed; restart bluetooth.service when safe to apply without interrupting active devices"
fi

# NetworkManager
networkmanager_root="$DOTFILES/etc/NetworkManager"
if [[ -d "$networkmanager_root" ]]; then
    if deploy_tree "$networkmanager_root" /etc/NetworkManager "NetworkManager"; then
        any_changes=true
        networkmanager_changed=true
    fi
fi
if $networkmanager_changed; then
    if systemd_unit_exists NetworkManager.service; then
        if sudo nmcli general reload >/dev/null 2>&1; then
            echo "  Reloaded NetworkManager configuration"
        elif sudo systemctl reload NetworkManager.service >/dev/null 2>&1; then
            echo "  Reloaded NetworkManager.service"
        else
            note_follow_up "NetworkManager config changed; reload or restart NetworkManager.service when safe"
        fi
    else
        note_follow_up "NetworkManager config changed, but NetworkManager.service was not found"
    fi
fi

udev_root="$DOTFILES/etc/udev/rules.d"
if [[ -d "$udev_root" ]]; then
    if deploy_tree "$udev_root" /etc/udev/rules.d "udev"; then
        any_changes=true
        udev_changed=true
    fi
fi
if $udev_changed; then
    sudo udevadm control --reload
    echo "  Reloaded udev rules"
    if $TRIGGER_USB; then
        sudo udevadm trigger --subsystem-match=usb --action=change || true
        echo "  Triggered USB device refresh"
    else
        note_follow_up "udev rules changed; rerun scripts/etc-deploy.sh --trigger-usb only if device permissions do not refresh naturally"
    fi
fi

systemd_root="$DOTFILES/etc/systemd/system"
if [[ -d "$systemd_root" ]]; then
    while IFS= read -r -d '' f; do
        rel="${f#"$systemd_root/"}"
        target="/etc/systemd/system/$rel"
        if deploy_file_if_changed "$f" "$target" "systemd/$rel"; then
            any_changes=true
            systemd_changed=true
            case "$rel" in
                openlinkhub.service|openlinkhub.service.d/*)
                    openlinkhub_changed=true
                    ;;
            esac
        fi
    done < <(find "$systemd_root" -type f -print0 | sort -z)
fi
if $systemd_changed; then
    sudo systemctl daemon-reload
    echo "  Reloaded systemd daemon"
fi

resolved_root="$DOTFILES/etc/systemd/resolved.conf.d"
if [[ -d "$resolved_root" ]]; then
    if deploy_tree "$resolved_root" /etc/systemd/resolved.conf.d "systemd/resolved.conf.d"; then
        any_changes=true
        resolved_changed=true
    fi
fi
if $resolved_changed; then
    if systemd_unit_exists systemd-resolved.service; then
        sudo systemctl enable systemd-resolved.service >/dev/null 2>&1 || true
        sudo systemctl restart systemd-resolved.service || true
        echo "  Applied systemd-resolved config update"
    else
        note_follow_up "systemd-resolved config changed, but systemd-resolved.service was not found"
    fi
fi

dbus_service_root="$DOTFILES/etc/dbus-1/services"
dbus_hook_src="$DOTFILES/etc/pacman.d/hooks/95-dbus-service-filename-normalize.hook"
dbus_hook_target="/etc/pacman.d/hooks/95-dbus-service-filename-normalize.hook"
dbus_normalizer_src="$DOTFILES/scripts/dbus-service-filename-normalize.sh"
dbus_normalizer_target="/usr/local/bin/dbus-service-filename-normalize.sh"
if [[ -d "$dbus_service_root" ]]; then
    if deploy_tree "$dbus_service_root" /etc/dbus-1/services "dbus-1/services"; then
        any_changes=true
        dbus_changed=true
    fi
fi
if [[ -f "$dbus_normalizer_src" ]]; then
    if deploy_file_if_changed "$dbus_normalizer_src" "$dbus_normalizer_target" "usr/local/bin/dbus-service-filename-normalize.sh" 755; then
        any_changes=true
        dbus_changed=true
    fi
fi
if [[ -f "$dbus_hook_src" ]]; then
    if deploy_file_if_changed "$dbus_hook_src" "$dbus_hook_target" "pacman.d/hooks/95-dbus-service-filename-normalize.hook"; then
        any_changes=true
        dbus_changed=true
        dbus_pacman_hook_changed=true
    fi
fi

for service_file in \
    /usr/share/dbus-1/services/org.erikreider.swaync.service \
    /usr/share/dbus-1/services/org.kde.kscreen.service \
    /usr/share/dbus-1/services/org.kde.plasma.Notifications.service \
    /usr/share/dbus-1/services/org.xfce.Thunar.FileManager1.service \
    /usr/share/dbus-1/services/org.xfce.Tumbler.Cache1.service \
    /usr/share/dbus-1/services/org.xfce.Tumbler.Manager1.service \
    /usr/share/dbus-1/services/org.xfce.Tumbler.Thumbnailer1.service
do
    if quarantine_dbus_service_file "$service_file" "canonical override in /etc/dbus-1/services"; then
        any_changes=true
        dbus_changed=true
    fi
done

if $dbus_changed; then
    if [[ -x "$dbus_normalizer_target" ]]; then
        sudo "$dbus_normalizer_target" --quiet || true
    fi
    if $dbus_pacman_hook_changed; then
        echo "  Installed pacman hook for D-Bus service filename normalization"
    fi
    note_follow_up "D-Bus service filenames normalized; restart affected desktop apps or relogin to fully eliminate broker filename warnings"
fi

if $bluetooth_changed; then
    if systemd_unit_exists bluetooth.service; then
        bluetooth_connected="$(bluetoothctl devices Connected 2>/dev/null || true)"
        if [[ -n "$bluetooth_connected" ]] && [[ "${DOTFILES_FORCE_BLUETOOTH_RESTART:-0}" != "1" ]]; then
            connected_summary="$(printf '%s' "$bluetooth_connected" | paste -sd ', ' -)"
            note_follow_up "bluetooth config changed; skipped bluetooth.service restart because connected devices are active (${connected_summary}). Set DOTFILES_FORCE_BLUETOOTH_RESTART=1 to override"
        else
            sudo systemctl try-restart bluetooth.service || true
            echo "  Applied bluetooth.service config update"
        fi
    else
        note_follow_up "bluetooth config changed, but bluetooth.service was not found"
    fi
fi

if systemd_unit_exists openlinkhub.service; then
    sudo systemctl enable openlinkhub.service >/dev/null 2>&1 || true
    if $openlinkhub_changed; then
        sudo systemctl try-restart openlinkhub.service || true
        echo "  Applied openlinkhub.service override update"
    fi
fi

if ! $any_changes; then
    echo "  No /etc changes detected"
fi

if [[ "${#FOLLOW_UP_NOTES[@]}" -gt 0 ]]; then
    echo "Follow-up notes:"
    for note in "${FOLLOW_UP_NOTES[@]}"; do
        echo "  - $note"
    done
fi

echo "/etc/ configs deployed."
