#!/usr/bin/env bash
# Captures Hyprland boot errors/warnings for post-boot review.
# Runs via exec-once in hyprland.conf. Results at ~/.local/share/hyprland-boot-logs/

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/lib/runtime-desktop-env.sh
source "$SCRIPT_DIR/lib/runtime-desktop-env.sh"

LOG_DIR="$HOME/.local/share/hyprland-boot-logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT="$LOG_DIR/boot-$TIMESTAMP.log"
BOOT_DELAY_SECS="${HYPR_BOOT_LOG_DELAY_SECS:-45}"
SYSTEM_BOOT_WINDOW_SECS="${HYPR_BOOT_JOURNAL_WINDOW_SECS:-120}"
LOG_MATCH='(\[ERR\]|\[WRN\]|(^|[^[:alnum:]_])(ERR|WRN|error|warn|failed|critical)([^[:alnum:]_]|$))'
ERROR_MATCH='(\[ERR\]|(^|[^[:alnum:]_])(ERR|error|failed|critical)([^[:alnum:]_]|$))'
WARN_MATCH='(\[WRN\]|(^|[^[:alnum:]_])(WRN|warn)([^[:alnum:]_]|$))'
HYPR_LOG_IGNORE_MATCH="Warning: you're using an NVIDIA GPU|ERR from aquamarine .*drm: getCurrentCRTC: No CRTC 0|ERR from aquamarine .*Wayland backend cannot start: wl_display_connect failed \\(is a wayland compositor running\\?\\)|ERR from aquamarine .*Requested backend \\(wayland\\) could not start, enabling fallbacks|ERR from aquamarine .*Implementation wayland failed, erasing\\.|ERR from aquamarine .*drm: Cannot commit when a page-flip is awaiting|ERR from aquamarine .*atomic drm request: failed to commit: Device or resource busy"
SYSTEM_JOURNAL_IGNORE_MATCH="nvidia: loading out-of-tree module taints kernel\\.|snd_hda_intel .*: no codecs found!|Bluetooth: hci0: HCI Enhanced Setup Synchronous Connection command is advertised, but not supported\\.|Bluetooth: hci0: AOSP quality report is not supported|Bluetooth: hci0: Bad flag given \\(0x1\\) vs supported \\(0x0\\)|bluetoothd\\[[0-9]+\\]: Failed to set default system config for hci0|bluetoothd\\[[0-9]+\\]: src/device\\.c:set_wake_allowed_complete\\(\\) Set device flags return status: Invalid Parameters|jellyfin\\.service: Referenced but unset environment variable evaluates to an empty string: JELLYFIN_NOWEBAPP_OPT, JELLYFIN_SERVICE_OPT|device \\(p2p-dev-wlp11s0\\): error setting IPv4 forwarding to '0': Resource temporarily unavailable|Activation request for 'org\\.freedesktop\\.home1' failed: The systemd unit 'dbus-org\\.freedesktop\\.home1\\.service' could not be found\\.|Activation request for 'org\\.freedesktop\\.resolve1' failed: The systemd unit 'dbus-org\\.freedesktop\\.resolve1\\.service' could not be found\\.|gnome-keyring-daemon\\[[0-9]+\\]: unable to create keyring dir: /\\.local/share/keyrings|greetd\\[[0-9]+\\]: gkr-pam: couldn't unlock the login keyring\\.|greetd\\[[0-9]+\\]: gkr-pam: unable to locate daemon control file|Failed to register with host portal .*App info not found for 'juhradial-mx'|dbus-broker-launch\\[[0-9]+\\]: Service file '/usr/share/dbus-1/services/org\\.erikreider\\.swaync\\.service' is not named after the D-Bus name 'org\\.freedesktop\\.Notifications'\\.|dbus-broker-launch\\[[0-9]+\\]: Service file '/usr/share/dbus-1/services/org\\.kde\\.kscreen\\.service' is not named after the D-Bus name 'org\\.kde\\.KScreen'\\.|dbus-broker-launch\\[[0-9]+\\]: Service file '/usr/share/dbus-1/services/org\\.kde\\.plasma\\.Notifications\\.service' is not named after the D-Bus name 'org\\.freedesktop\\.Notifications'\\.|dbus-broker-launch\\[[0-9]+\\]: Ignoring duplicate name 'org\\.freedesktop\\.Notifications' in service file '/usr/share/dbus-1/services/org\\.kde\\.plasma\\.Notifications\\.service'|dbus-broker-launch\\[[0-9]+\\]: Service file '/usr/share/dbus-1/services/org\\.xfce\\.Thunar\\.FileManager1\\.service' is not named after the D-Bus name 'org\\.freedesktop\\.FileManager1'\\.|dbus-broker-launch\\[[0-9]+\\]: Service file '/usr/share/dbus-1/services/org\\.xfce\\.Tumbler\\.(Cache1|Manager1|Thumbnailer1)\\.service' is not named after the D-Bus name 'org\\.freedesktop\\.thumbnails\\.(Cache1|Manager1|Thumbnailer1)'\\."

mkdir -p "$LOG_DIR"

# Wait for Hyprland IPC plus noisy boot services to settle.
sleep "$BOOT_DELAY_SECS"
refresh_desktop_runtime_env
BOOT_STARTED_AT="$(uptime -s 2>/dev/null || true)"
if [[ -n "$BOOT_STARTED_AT" ]]; then
    BOOT_STARTED_EPOCH="$(date -d "$BOOT_STARTED_AT" '+%s' 2>/dev/null || true)"
    if [[ -n "${BOOT_STARTED_EPOCH:-}" ]]; then
        BOOT_WINDOW_END_AT="$(
            date -d "@$((BOOT_STARTED_EPOCH + SYSTEM_BOOT_WINDOW_SECS))" '+%F %T' 2>/dev/null || true
        )"
    else
        BOOT_WINDOW_END_AT=""
    fi
else
    BOOT_WINDOW_END_AT=""
fi

{
    echo "=== HYPRLAND BOOT LOG — $TIMESTAMP ==="
    if command -v jq >/dev/null 2>&1; then
        HYPR_VERSION=$(hyprctl version -j 2>/dev/null | jq -r '.tag // .version // "unknown"' 2>/dev/null || echo "unknown")
    else
        HYPR_VERSION=$(hyprctl version -j 2>/dev/null | sed -n 's/.*"tag":[[:space:]]*"\([^"]*\)".*/\1/p' | head -1 || echo "unknown")
        HYPR_VERSION=${HYPR_VERSION:-unknown}
    fi
    echo "Hyprland: ${HYPR_VERSION}"
    echo ""

    # 1. Config errors (filter out empty arrays like [""])
    echo "=== CONFIG ERRORS ==="
    CONFIG_ERRORS=$(hyprctl -j configerrors 2>&1 || echo "hyprctl not available")
    CONFIG_STRIPPED=$(echo "$CONFIG_ERRORS" | tr -d '[:space:]')
    if [[ "$CONFIG_STRIPPED" == '[""]' ]] || [[ "$CONFIG_STRIPPED" == "[]" ]]; then
        echo "No config errors"
    else
        echo "$CONFIG_ERRORS"
    fi
    echo ""

    # 2. Hyprland instance log — errors/warnings (retry for race condition)
    echo "=== LOG ERRORS/WARNINGS ==="
    HYPR_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}/hypr/${HYPRLAND_INSTANCE_SIGNATURE:-unknown}"
    HYPR_LOG="${HYPR_RUNTIME_DIR}/hyprland.log"
    LEGACY_HYPR_LOG="/tmp/hypr/${HYPRLAND_INSTANCE_SIGNATURE:-unknown}/hyprland.log"
    for _ in 1 2 3; do
        [ -f "$HYPR_LOG" ] && break
        [ -f "$LEGACY_HYPR_LOG" ] && HYPR_LOG="$LEGACY_HYPR_LOG" && break
        sleep 2
    done
    if [ -f "$HYPR_LOG" ]; then
        FILTERED_LOGS="$(
            grep -iE "$LOG_MATCH" "$HYPR_LOG" \
                | grep -viE "$HYPR_LOG_IGNORE_MATCH" \
                | head -100 \
                || true
        )"
        if [[ -n "$FILTERED_LOGS" ]]; then
            printf '%s\n' "$FILTERED_LOGS"
        else
            echo "No actionable errors/warnings found"
        fi
    else
        echo "Log not found: $HYPR_LOG"
    fi
    echo ""

    # 3. Failed systemd user services
    echo "=== FAILED USER SERVICES ==="
    systemctl --user --failed --no-pager 2>&1
    echo ""

    # 4. System journal warnings/errors since boot (filtered to actionable noise)
    echo "=== SYSTEM JOURNAL WARNINGS ==="
    SYSTEM_JOURNAL_LINES="$(
        if [[ -n "$BOOT_STARTED_AT" && -n "$BOOT_WINDOW_END_AT" ]]; then
            journalctl -b --since "$BOOT_STARTED_AT" --until "$BOOT_WINDOW_END_AT" -p warning..alert --no-pager 2>/dev/null
        else
            journalctl -b -p warning..alert --no-pager 2>/dev/null
        fi \
            | grep -v '^-- No entries --$' \
            | grep -viE "$SYSTEM_JOURNAL_IGNORE_MATCH" \
            | head -200 \
            || true
    )"
    if [[ -n "$SYSTEM_JOURNAL_LINES" ]]; then
        printf '%s\n' "$SYSTEM_JOURNAL_LINES"
    else
        echo "No actionable journal warnings found"
    fi
    echo ""

    # 5. Portal status
    echo "=== XDG PORTAL STATUS ==="
    systemctl --user status xdg-desktop-portal.service xdg-desktop-portal-gtk.service xdg-desktop-portal-hyprland.service 2>&1 | grep -E "Active:|Main PID:|cannot|failed" || echo "Portals OK"
    echo ""

    # 6. NVIDIA status
    echo "=== NVIDIA ==="
    nvidia-smi --query-gpu=driver_version,gpu_name,temperature.gpu,power.draw --format=csv,noheader 2>/dev/null || echo "nvidia-smi not available"

} > "$REPORT" 2>&1

# Desktop notification with summary (exclude section headers and metadata lines)
ERROR_COUNT="$(
  grep -viE "^===|^Log not found|^hyprctl|^0 loaded|^No actionable errors/warnings found$|^No actionable journal warnings found$" "$REPORT" 2>/dev/null \
    | grep -ciE "$ERROR_MATCH" 2>/dev/null \
    || true
)"
WARN_COUNT="$(
  grep -viE "^===|^Log not found|^hyprctl|^0 loaded|^No actionable errors/warnings found$|^No actionable journal warnings found$" "$REPORT" 2>/dev/null \
    | grep -ciE "$WARN_MATCH" 2>/dev/null \
    || true
)"
ERROR_COUNT="${ERROR_COUNT:-0}"
WARN_COUNT="${WARN_COUNT:-0}"
if (( ERROR_COUNT > 0 )); then
  notify-send -u critical -a "Hyprland Boot" \
      "Boot log: ${ERROR_COUNT} errors" \
      "${ERROR_COUNT} errors, ${WARN_COUNT} warnings — ${REPORT}" \
      2>/dev/null || true
elif (( WARN_COUNT > 0 )) && [[ "${HYPR_BOOT_LOG_NOTIFY_WARNINGS:-0}" =~ ^(1|true|TRUE|yes|YES|on|ON)$ ]]; then
  notify-send -u low -a "Hyprland Boot" \
      "Boot log warnings" \
      "${WARN_COUNT} warnings — ${REPORT}" \
      2>/dev/null || true
fi

# Keep only last 10 boot logs
ls -t "$LOG_DIR"/boot-*.log 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
exit 0
