#!/usr/bin/env bash
# mx-reset.sh — Reset MX Master 4 to known-good state
# Fixes: scroll wheel, haptics, torque, BT connection
# Usage: mx-reset.sh [--full]
set -euo pipefail

DEVICE_MAC="D2:8E:C5:DE:9F:C8"
DEVICE_NAME="MX Master 4"

info()  { printf '\033[0;36m:: %s\033[0m\n' "$*"; }
ok()    { printf '\033[0;32m   %s\033[0m\n' "$*"; }
warn()  { printf '\033[0;33m   %s\033[0m\n' "$*"; }
fail()  { printf '\033[0;31m   %s\033[0m\n' "$*"; }

full=false
[[ "${1:-}" == "--full" ]] && full=true

# Step 1: Restart logid (reapplies logiops config)
info "Restarting logid..."
sudo systemctl restart logid
sleep 1
if journalctl -u logid --no-pager -n 3 --since "5 seconds ago" | grep -q "Device found: $DEVICE_NAME"; then
    ok "logid found $DEVICE_NAME"
else
    warn "logid did not find device — may need BT reconnect"
fi

# Step 2: Restore Solaar settings
info "Restoring Solaar settings..."
solaar config "$DEVICE_NAME" scroll-ratchet-torque 75 2>/dev/null && ok "torque: 75" || warn "torque failed"
solaar config "$DEVICE_NAME" haptic-level 60 2>/dev/null && ok "haptic-level: 60" || warn "haptic-level failed"

# Step 3: Restart makima (reloads per-app profiles)
info "Restarting makima..."
sudo systemctl restart makima
sleep 1
profiles=$(journalctl -u makima --no-pager --since "5 seconds ago" | grep -c "Parsing config")
ok "$profiles profiles loaded"

# Step 4: Full BT reconnect (only with --full)
if $full; then
    info "Full BT reconnect..."
    bluetoothctl disconnect "$DEVICE_MAC" 2>/dev/null || true
    sleep 2
    bluetoothctl connect "$DEVICE_MAC" 2>/dev/null && ok "BT reconnected" || fail "BT connect failed — power cycle the mouse"
    sleep 2
    # Re-run steps 1-3 after BT reconnect
    sudo systemctl restart logid
    sleep 1
    solaar config "$DEVICE_NAME" scroll-ratchet-torque 75 2>/dev/null
    solaar config "$DEVICE_NAME" haptic-level 60 2>/dev/null
    sudo systemctl restart makima
    sleep 1
fi

# Step 5: Verify
info "Verifying..."
battery=$(solaar show "$DEVICE_NAME" 2>/dev/null | strings | grep -oP 'Battery: \K\d+' || echo "?")
ok "Battery: ${battery}%"
ok "Done — test scroll and haptic toggle now"
