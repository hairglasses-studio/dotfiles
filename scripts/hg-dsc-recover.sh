#!/usr/bin/env bash
# hg-dsc-recover.sh — restore Samsung ultrawide from DSC fallback to native.
#
# Symptom: DP-2 comes up in 3840×1080@120 instead of 5120×1440@240 (and
# any layer-shell surfaces on DP-2 render with top-pixel bleed above
# their bounds — the ticker shows duplicated content at ~12 px above its
# 28-px strip). The fix depends on whether the monitor is still
# *exposing* its native mode in EDID or has lost DSC handshake entirely.
#
# Usage:
#   hg-dsc-recover.sh                  # DP-2, native 5120x1440@239.76 scale=2
#   hg-dsc-recover.sh DP-2 3840x1080@119.97 4596x271 1
#
# Exit codes:
#   0  already in native mode, OR soft recovery succeeded
#   2  EDID has lost native mode — hardware power-cycle required
#   1  unexpected error
set -euo pipefail

MONITOR="${1:-DP-2}"
NATIVE="${2:-5120x1440@239.76}"
POSITION="${3:-4596x271}"
SCALE="${4:-2}"

_die() { printf 'hg-dsc-recover: %s\n' "$*" >&2; exit 1; }

command -v hyprctl >/dev/null 2>&1 || _die "hyprctl not found (Hyprland not running?)"

_dump_monitor() {
  hyprctl monitors -j | python3 -c "
import json, sys
name = '$MONITOR'
for m in json.load(sys.stdin):
    if m['name'] == name:
        print(f\"{m['width']}x{m['height']}@{m['refreshRate']:.2f}Hz scale={m['scale']}\")
        break
else:
    print('NOT_FOUND')
"
}

_has_native_available() {
  local target="${NATIVE%@*}"
  hyprctl monitors -j | python3 -c "
import json, sys
name = '$MONITOR'
target = '$target'
for m in json.load(sys.stdin):
    if m['name'] == name:
        print('yes' if any(target in mode for mode in m.get('availableModes', [])) else 'no')
        break
else:
    print('no')
"
}

CURRENT="$(_dump_monitor)"
if [[ "$CURRENT" == "NOT_FOUND" ]]; then
  _die "monitor $MONITOR not found"
fi

printf 'before: %s\n' "$CURRENT"

NATIVE_PIX="${NATIVE%@*}"
if [[ "$CURRENT" == *"${NATIVE_PIX}"*"scale=${SCALE}"* ]] \
   || [[ "$CURRENT" == *"${NATIVE_PIX}"*"scale=${SCALE}.0"* ]]; then
  printf 'status: %s already in native mode — nothing to do\n' "$MONITOR"
  exit 0
fi

if [[ "$(_has_native_available)" != "yes" ]]; then
  cat >&2 <<EOF

$MONITOR: EDID no longer exposes $NATIVE — DSC negotiation has fully failed.
Soft recovery cannot fix this. Follow the hardware power-cycle procedure
documented in .claude/rules/nvidia-wayland.md §146:

  1. Exit Hyprland  (hyprctl dispatch exit  — SIGTERMs the session)
  2. Hold $MONITOR power button ~8 seconds until fully off
  3. Unplug monitor power for 10+ seconds
  4. Replug, let the monitor fully boot
  5. Log back in through greetd — fresh EDID read restores DSC
EOF
  exit 2
fi

printf 'applying: hyprctl keyword monitor %s,%s,%s,%s\n' \
  "$MONITOR" "$NATIVE" "$POSITION" "$SCALE"
hyprctl keyword monitor "$MONITOR,$NATIVE,$POSITION,$SCALE" >/dev/null

sleep 2
AFTER="$(_dump_monitor)"
printf 'after:  %s\n' "$AFTER"

# Restart layer-shell overlays on this monitor so they re-read geometry
for svc in dotfiles-keybind-ticker.service dotfiles-window-label.service dotfiles-fleet-sparkline.service; do
  if systemctl --user is-active --quiet "$svc" 2>/dev/null; then
    systemctl --user restart "$svc"
    printf 'restarted %s\n' "$svc"
  fi
done

# Sanity: final check
FINAL="$(_dump_monitor)"
if [[ "$FINAL" == *"${NATIVE_PIX}"*"scale=${SCALE}"* ]] \
   || [[ "$FINAL" == *"${NATIVE_PIX}"*"scale=${SCALE}.0"* ]]; then
  printf 'status: soft recovery succeeded\n'
  exit 0
fi

cat >&2 <<EOF
status: soft recovery did not take hold (still at $FINAL)
        escalate to hardware power-cycle per .claude/rules/nvidia-wayland.md §146
EOF
exit 2
