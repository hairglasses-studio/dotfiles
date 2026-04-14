#!/usr/bin/env bash
# etc-deploy.sh — Deploy tracked /etc/ configs (sysctl, modprobe, udev, etc.)
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"

deploy_tree() {
    local src_root="$1" dst_root="$2" label="$3"
    [[ -d "$src_root" ]] || return 0

    while IFS= read -r -d '' f; do
        local rel="${f#"$src_root/"}"
        local target="$dst_root/$rel"
        sudo install -d "$(dirname "$target")"
        sudo install -m644 "$f" "$target"
        echo "  Deployed $label/$rel"
    done < <(find "$src_root" -type f -print0 | sort -z)
}

systemd_unit_exists() {
    local unit="$1"
    sudo systemctl list-unit-files "$unit" --no-legend >/dev/null 2>&1
}

echo "Deploying /etc/ configs..."

# sysctl
if [[ -f "$DOTFILES/etc/sysctl.d/99-workstation.conf" ]]; then
    sudo install -d /etc/sysctl.d
    sudo install -m644 "$DOTFILES/etc/sysctl.d/99-workstation.conf" /etc/sysctl.d/99-workstation.conf
    sudo sysctl -p /etc/sysctl.d/99-workstation.conf
    echo "  sysctl tuning applied"
fi

# modprobe (NVIDIA)
for f in "$DOTFILES/etc/modprobe.d/"*.conf; do
    [[ -f "$f" ]] || continue
    sudo install -d /etc/modprobe.d
    sudo install -m644 "$f" "/etc/modprobe.d/$(basename "$f")"
    echo "  Deployed $(basename "$f")"
done

if [[ ! -f "$DOTFILES/etc/modprobe.d/blacklist-k10temp.conf" ]] && [[ -f /etc/modprobe.d/blacklist-k10temp.conf ]]; then
    sudo rm -f /etc/modprobe.d/blacklist-k10temp.conf
    echo "  Removed stale blacklist-k10temp.conf"
fi

# modules-load
for f in "$DOTFILES/etc/modules-load.d/"*.conf; do
    [[ -f "$f" ]] || continue
    sudo install -d /etc/modules-load.d
    sudo install -m644 "$f" "/etc/modules-load.d/$(basename "$f")"
    echo "  Deployed modules-load.d/$(basename "$f")"
done

# Bluetooth
if [[ -f "$DOTFILES/etc/bluetooth/main.conf" ]]; then
    sudo install -d -m0555 /etc/bluetooth
    sudo install -m644 "$DOTFILES/etc/bluetooth/main.conf" /etc/bluetooth/main.conf
    echo "  Deployed bluetooth/main.conf"
fi

udev_root="$DOTFILES/etc/udev/rules.d"
if [[ -d "$udev_root" ]]; then
    deploy_tree "$udev_root" /etc/udev/rules.d "udev"
    sudo udevadm control --reload
    sudo udevadm trigger --subsystem-match=usb --action=change || true
    echo "  Reloaded udev rules"
fi

systemd_root="$DOTFILES/etc/systemd/system"
if [[ -d "$systemd_root" ]]; then
    deploy_tree "$systemd_root" /etc/systemd/system "systemd"
    sudo systemctl daemon-reload
    echo "  Reloaded systemd daemon"
fi

if systemd_unit_exists bluetooth.service; then
    sudo systemctl restart bluetooth.service || true
    echo "  Restarted bluetooth.service"
fi

if systemd_unit_exists openlinkhub.service; then
    sudo systemctl enable openlinkhub.service >/dev/null 2>&1 || true
    sudo systemctl restart openlinkhub.service || true
    echo "  Ensured openlinkhub.service is enabled and restarted"
fi

echo "/etc/ configs deployed."
