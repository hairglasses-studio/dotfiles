#!/usr/bin/env bash
# juhradial-verify.sh — verify live MX runtime state and the repo-managed patch layer
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

quiet=false
for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

ok_count=0
warn_count=0
fail_count=0

report() {
  local level="$1"
  shift
  case "$level" in
    OK)   ok_count=$((ok_count + 1)) ;;
    WARN) warn_count=$((warn_count + 1)) ;;
    FAIL) fail_count=$((fail_count + 1)) ;;
  esac
  if ! $quiet; then
    printf '%-5s %s\n' "$level" "$*"
  fi
}

check_sync_state() {
  local label="$1"
  local tracked="$2"
  local live="$3"

  if [[ ! -e "$tracked" ]]; then
    report FAIL "$label seed missing at $tracked"
    return
  fi

  if [[ ! -e "$live" ]]; then
    report FAIL "$label missing at $live"
    return
  fi

  if [[ -d "$tracked" ]]; then
    if diff -qr --exclude='.gitkeep' --exclude='*.bak.*' "$tracked" "$live" >/dev/null 2>&1; then
      report OK "$label synced"
    else
      report FAIL "$label drifted from repo seed"
    fi
    return
  fi

  if cmp -s "$tracked" "$live"; then
    report OK "$label synced"
  else
    report FAIL "$label drifted from repo seed"
  fi
}

tracked_dir="$(juhradial_seed_dir)"
live_dir="$(juhradial_config_dir)"

if juhradial_systemctl is-active juhradialmx-daemon.service >/dev/null 2>&1; then
  report OK "juhradialmx-daemon.service active"
else
  report FAIL "juhradialmx-daemon.service inactive"
fi

if juhradial_systemctl is-active ydotool.service >/dev/null 2>&1; then
  report OK "ydotool.service active"
else
  report FAIL "ydotool.service inactive"
fi

if juhradial_overlay_running; then
  report OK "juhradial overlay running"
else
  report WARN "juhradial overlay not running"
fi

if juhradial_patch_applied; then
  report OK "repo patch layer applied in local juhradial checkout"
else
  report FAIL "repo patch layer missing from local juhradial checkout"
fi

check_sync_state "config" "$tracked_dir/config.json" "$live_dir/config.json"
check_sync_state "profiles" "$tracked_dir/profiles.json" "$live_dir/profiles.json"
check_sync_state "macros" "$(juhradial_seed_macros_dir)" "$(juhradial_macros_dir)"

transport="$(juhradial_transport_state)"
case "$transport" in
  bolt)
    report OK "transport is Bolt"
    ;;
  bluetooth)
    report WARN "transport is Bluetooth"
    ;;
  split-brain)
    report FAIL "transport split-brain detected"
    ;;
  *)
    report FAIL "transport is $transport"
    ;;
esac

if battery="$(juhradial_battery_status 2>/dev/null)"; then
  read -r battery_pct battery_charging <<<"$battery"
  report OK "battery ${battery_pct}% (charging: ${battery_charging})"
else
  report FAIL "battery unavailable from juhradial/Bluetooth path"
fi

if [[ -d "$(juhradial_source_dir)/.git" ]]; then
  if "$SCRIPT_DIR/juhradial-patch-guard.sh" --local-source --quiet; then
    report OK "patch guard passed against local pinned checkout"
  else
    report FAIL "patch guard failed against local pinned checkout"
  fi
else
  report WARN "local juhradial source checkout missing; skipped patch guard"
fi

if (( fail_count > 0 )); then
  $quiet || printf 'verify summary: %d ok, %d warn, %d fail\n' "$ok_count" "$warn_count" "$fail_count"
  exit 1
fi

$quiet || printf 'verify summary: %d ok, %d warn, %d fail\n' "$ok_count" "$warn_count" "$fail_count"
