#!/usr/bin/env bash
# Captures Hyprland boot errors/warnings for post-boot review.
# Runs via exec-once in hyprland.conf. Results at ~/.local/share/hyprland-boot-logs/

set -euo pipefail

LOG_DIR="$HOME/.local/share/hyprland-boot-logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
REPORT="$LOG_DIR/boot-$TIMESTAMP.log"

mkdir -p "$LOG_DIR"

# Wait for Hyprland IPC to be ready
sleep 5

{
    echo "=== HYPRLAND BOOT LOG — $TIMESTAMP ==="
    echo "Hyprland: $(hyprctl version -j 2>/dev/null | grep -o '"tag":"[^"]*"' | head -1 || echo 'unknown')"
    echo ""

    # 1. Config errors
    echo "=== CONFIG ERRORS ==="
    hyprctl -j configerrors 2>&1 || echo "hyprctl not available"
    echo ""

    # 2. Hyprland instance log — errors/warnings
    echo "=== LOG ERRORS/WARNINGS ==="
    HYPR_LOG="/tmp/hypr/${HYPRLAND_INSTANCE_SIGNATURE:-unknown}/hyprland.log"
    if [ -f "$HYPR_LOG" ]; then
        grep -iE "\[ERR\]|\[WRN\]|error|warn|fail|critical" "$HYPR_LOG" | head -100 || echo "No errors/warnings found"
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

# Desktop notification with summary
ERROR_COUNT=$(grep -ciE "\[ERR\]|error|fail|critical" "$REPORT" 2>/dev/null || echo 0)
WARN_COUNT=$(grep -ciE "\[WRN\]|warn" "$REPORT" 2>/dev/null || echo 0)
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
