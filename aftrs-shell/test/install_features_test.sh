#!/usr/bin/env bash
# Minimal feature tests for installer additions.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$SCRIPT_DIR/.."

echo "Testing aftrs_initial_unpack.sh --dry-run"
"$ROOT/aftrs_initial_unpack.sh" --dry-run >/dev/null

echo "Testing install.sh --dry-run paths"
AFTRS_LOCAL_ROOT="$ROOT/.." \
AFTRS_SHELL_HOME="$(mktemp -d)" \
AFTRS_COMPLETIONS_DIR="$(mktemp -d)" \
"$ROOT/install.sh" --dry-run --skip-clone --no-install >/dev/null

echo "All feature tests passed"


