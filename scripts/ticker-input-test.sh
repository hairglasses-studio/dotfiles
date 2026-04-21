#!/usr/bin/env bash
# ticker-input-test.sh — smoke-test the keyboard / DBus control surface.
#
# Drives the Phase 1 DBus API through realistic paths and asserts the
# live property readouts match expectations. Safe to run against the
# active ticker — each test restores the previous state on exit.
#
# Out of scope for this harness: Hyprland keybinds (requires a real
# key-event synthesis through wlr-virtual-keyboard; use the dotfiles
# MCP `input_key_press` for that) and makima gamepad bindings
# (requires a physical device).
#
# Exit 0 on success, non-zero on first assertion failure.

set -euo pipefail

BUS="io.hairglasses.keybind_ticker"
OBJ="/io/hairglasses/Ticker"
IFACE="io.hairglasses.Ticker"

_gdbus() {
  local method="$1"; shift
  gdbus call --session -d "$BUS" -o "$OBJ" -m "$IFACE.$method" "$@" 2>&1
}

_prop() {
  gdbus call --session -d "$BUS" -o "$OBJ" \
    -m org.freedesktop.DBus.Properties.Get "$IFACE" "$1" 2>&1 \
    | sed -E "s/.*<'?([^'>]*).*/\1/"
}

_assert() {
  local want="$1" got="$2" label="$3"
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL  %s: expected %q, got %q\n' "$label" "$want" "$got" >&2
    exit 1
  fi
  printf 'ok    %s: %q\n' "$label" "$got"
}

# Sanity: ticker must be up.
if ! gdbus call --session -d "$BUS" -o "$OBJ" \
    -m org.freedesktop.DBus.Peer.Ping >/dev/null 2>&1; then
  printf 'ticker DBus not reachable; is the service running?\n' >&2
  exit 2
fi

prev_pin="$(_prop Pinned)"
prev_shuffle="$(_prop Shuffle)"
prev_playlist="$(_prop Playlist)"

cleanup() {
  # Best-effort restore — don't care if any of these fail.
  if [[ "$prev_pin" == "" ]]; then
    _gdbus Unpin >/dev/null || true
  else
    _gdbus Pin "$prev_pin" >/dev/null || true
  fi
  if [[ "$prev_shuffle" == "true" ]]; then
    _gdbus Shuffle on >/dev/null || true
  else
    _gdbus Shuffle off >/dev/null || true
  fi
  _gdbus SetPlaylist "$prev_playlist" >/dev/null || true
}
trap cleanup EXIT

echo '━━ Pin / Unpin round-trip ━━'
_gdbus Pin load >/dev/null
sleep 0.4
_assert "load" "$(_prop Pinned)" "Pin('load') reflected in Pinned"
_assert "load" "$(_prop CurrentStream)" "Pin('load') advanced CurrentStream"
_gdbus Unpin >/dev/null
sleep 0.2
_assert "" "$(_prop Pinned)" "Unpin cleared Pinned"

echo '━━ PinToggle ━━'
_gdbus PinToggle >/dev/null
sleep 0.3
pinned_after="$(_prop Pinned)"
if [[ -z "$pinned_after" ]]; then
  printf 'FAIL  PinToggle should have pinned a stream\n' >&2
  exit 1
fi
printf 'ok    PinToggle pinned: %q\n' "$pinned_after"
_gdbus PinToggle >/dev/null
sleep 0.2
_assert "" "$(_prop Pinned)" "PinToggle unpinned"

echo '━━ Shuffle on/off ━━'
_gdbus Shuffle on >/dev/null
sleep 0.3
_assert "true" "$(_prop Shuffle)" "Shuffle(on) reflected"
_gdbus Shuffle off >/dev/null
sleep 0.3
_assert "false" "$(_prop Shuffle)" "Shuffle(off) reflected"

echo '━━ NextStream advances ━━'
_gdbus Shuffle off >/dev/null  # deterministic order
before="$(_prop CurrentStream)"
_gdbus NextStream >/dev/null
sleep 0.3
after="$(_prop CurrentStream)"
if [[ "$before" == "$after" ]]; then
  printf 'FAIL  NextStream did not advance (still on %q)\n' "$before" >&2
  exit 1
fi
printf 'ok    NextStream advanced: %q → %q\n' "$before" "$after"

echo '━━ SetUrgent / SnoozeUrgent ━━'
_gdbus SetUrgent true >/dev/null
sleep 0.2
_assert "true" "$(_prop Urgent)" "SetUrgent(true) reflected"
_gdbus SnoozeUrgent >/dev/null
sleep 0.2
_assert "false" "$(_prop Urgent)" "SnoozeUrgent cleared"

echo
echo 'ALL PASS'
