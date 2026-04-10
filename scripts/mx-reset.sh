#!/usr/bin/env bash
# mx-reset.sh — Reset MX Master 4 to the repo-managed juhradial state
# Usage: mx-reset.sh [--full]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

info()  { printf '\033[0;36m:: %s\033[0m\n' "$*"; }
ok()    { printf '\033[0;32m   %s\033[0m\n' "$*"; }
warn()  { printf '\033[0;33m   %s\033[0m\n' "$*"; }
fail()  { printf '\033[0;31m   %s\033[0m\n' "$*"; }

full=false
[[ "${1:-}" == "--full" ]] && full=true

info "Syncing repo-managed juhradial config..."
"$SCRIPT_DIR/juhradial-sync.sh" --quiet
ok "juhradial seed config copied"

info "Restarting ydotool + juhradial daemon..."
juhradial_systemctl restart ydotool.service >/dev/null
juhradial_systemctl restart juhradialmx-daemon.service >/dev/null
ok "user services restarted"

if [[ -x "$SCRIPT_DIR/juhradial-wheel-apply.sh" ]]; then
    info "Applying wheel hardware compatibility state..."
    "$SCRIPT_DIR/juhradial-wheel-apply.sh" --quiet || true
    ok "wheel state applied"
fi

info "Restarting juhradial overlay..."
"$SCRIPT_DIR/juhradial-mx.sh" --restart --quiet
ok "overlay restarted"

# Step 4: Extra receiver reapply (only with --full)
if $full; then
    info "Reapplying wheel state after full reset..."
    "$SCRIPT_DIR/juhradial-wheel-apply.sh" --quiet || true
    juhradial_systemctl restart juhradialmx-daemon.service >/dev/null || true
    "$SCRIPT_DIR/juhradial-mx.sh" --restart --quiet || true
    ok "full receiver reset complete"
fi

# Step 5: Verify
info "Verifying..."
transport="$(juhradial_transport_state)"
case "$transport" in
    bolt) ok "Transport: Bolt receiver" ;;
    bluetooth) warn "Transport: Bluetooth (expected Bolt receiver)" ;;
    split-brain) fail "Transport: split-brain (Bluetooth + receiver both visible)" ;;
    *) warn "Transport: ${transport}" ;;
esac

if status="$(juhradial_battery_status 2>/dev/null)"; then
    read -r battery charging <<<"$status"
    ok "Battery: ${battery}% (charging: ${charging})"
else
    warn "Battery unavailable — daemon not reporting yet"
fi
ok "Done — test thumb button, wheel direction, and radial menu now"
