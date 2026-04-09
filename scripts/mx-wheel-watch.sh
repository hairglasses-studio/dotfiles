#!/usr/bin/env bash
# mx-wheel-watch.sh — Reapply MX Master 4 wheel settings after BT reconnects
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEVICE_MAC="${BT_MX_MASTER:-D2:8E:C5:DE:9F:CC}"
APPLY_SCRIPT="$SCRIPT_DIR/mx-wheel-apply.sh"
COOLDOWN_SECS="${MX_WHEEL_WATCH_COOLDOWN:-8}"

log() {
  printf '[mx-wheel-watch] %s\n' "$*"
}

if [[ ! -x "$APPLY_SCRIPT" ]]; then
  log "apply script missing: $APPLY_SCRIPT"
  exit 1
fi

last_restore=0

log "starting monitor for $DEVICE_MAC"
"$APPLY_SCRIPT" --quiet || true

stdbuf -oL bluetoothctl --monitor | while IFS= read -r line; do
  case "$line" in
    *"Device $DEVICE_MAC Connected: yes"*)
      now="$(date +%s)"
      if (( now - last_restore < COOLDOWN_SECS )); then
        continue
      fi

      last_restore="$now"
      log "reconnect detected; restoring wheel settings"
      "$APPLY_SCRIPT" --quiet || log "wheel restore failed"
      ;;
  esac
done
