#!/usr/bin/env bash
# controller-ticker-next.sh — Gamepad advances the ticker to the next stream.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../hg" ticker next
