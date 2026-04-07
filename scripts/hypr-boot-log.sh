#!/usr/bin/env bash
# Captures Hyprland boot errors/warnings for post-boot review.
# Runs via exec-once in hyprland.conf. Results at ~/.local/share/hyprland-boot-logs/

set -euo pipefail

LOG_DIR="$HOME/.local/share/hyprland-boot-logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT="$LOG_DIR/boot-$TIMESTAMP.log"
LOG_MATCH='(\[ERR\]|\[WRN\]|(^|[^[:alnum:]_])(ERR|WRN|error|warn|failed|critical)([^[:alnum:]_]|$))'
ERROR_MATCH='(\[ERR\]|(^|[^[:alnum:]_])(ERR|error|failed|critical)([^[:alnum:]_]|$))'
WARN_MATCH='(\[WRN\]|(^|[^[:alnum:]_])(WRN|warn)([^[:alnum:]_]|$))'

mkdir -p "$LOG_DIR"

# Wait for Hyprland IPC to be ready
sleep 5

{
    echo "=== HYPRLAND BOOT LOG — $TIMESTAMP ==="
    if command -v jq >/dev/null 2>&1; then
        HYPR_VERSION=$(hyprctl version -j 2>/dev/null | jq -r '.tag // .version // "unknown"' 2>/dev/null || echo "unknown")
    else
        HYPR_VERSION=$(hyprctl version -j 2>/dev/null | sed -n 's/.*"tag":[[:space:]]*"\([^"]*\)".*/\1/p' | head -1)
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
        grep -iE "$LOG_MATCH" "$HYPR_LOG" | head -100 || echo "No errors/warnings found"
    else
        echo "Log not found: $HYPR_LOG"
    fi
    echo ""

    # 3. Failed systemd user services
    echo "=== FAILED USER SERVICES ==="
    systemctl --user --failed --no-pager 2>&1
    echo ""

    # 4. Portal status
    echo "=== XDG PORTAL STATUS ==="
    systemctl --user status xdg-desktop-portal.service xdg-desktop-portal-gtk.service xdg-desktop-portal-hyprland.service 2>&1 | grep -E "Active:|Main PID:|cannot|failed" || echo "Portals OK"
    echo ""

    # 5. NVIDIA status
    echo "=== NVIDIA ==="
    nvidia-smi --query-gpu=driver_version,gpu_name,temperature.gpu,power.draw --format=csv,noheader 2>/dev/null || echo "nvidia-smi not available"

} > "$REPORT" 2>&1

# Desktop notification with summary (exclude section headers and metadata lines)
ERROR_COUNT=$(grep -viE "^===|^Log not found|^hyprctl|^0 loaded" "$REPORT" 2>/dev/null | grep -ciE "$ERROR_MATCH" 2>/dev/null || echo 0)
WARN_COUNT=$(grep -viE "^===|^Log not found|^hyprctl|^0 loaded" "$REPORT" 2>/dev/null | grep -ciE "$WARN_MATCH" 2>/dev/null || echo 0)
if (( ERROR_COUNT > 0 )); then
  notify-send -u critical -a "Hyprland Boot" \
      "Boot log: ${ERROR_COUNT} errors" \
      "${ERROR_COUNT} errors, ${WARN_COUNT} warnings — ${REPORT}" \
      2>/dev/null || true
else
  notify-send -u low -a "Hyprland Boot" \
      "Boot log captured" \
      "${WARN_COUNT} warnings — ${REPORT}" \
      2>/dev/null || true
fi

# Keep only last 10 boot logs
ls -t "$LOG_DIR"/boot-*.log 2>/dev/null | tail -n +11 | xargs rm -f 2>/dev/null || true
