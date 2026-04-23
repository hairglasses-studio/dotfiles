#!/usr/bin/env bash
set -euo pipefail

# app-launcher.sh — stable Mod+D app launcher entrypoint.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
exec "$SCRIPT_DIR/menu-control.sh" apps
