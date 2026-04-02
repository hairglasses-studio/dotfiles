#!/usr/bin/env bash
# controller-shader-random.sh — Gamepad shader randomize
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../hg" shader random
