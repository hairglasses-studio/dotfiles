#!/usr/bin/env bash
# controller-ticker-shuffle.sh — Gamepad toggles ticker shuffle mode.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../hg" ticker shuffle
