#!/usr/bin/env bash
# controller-ticker-pin-toggle.sh — Gamepad pins/unpins the current ticker stream.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../hg" ticker pin-toggle
