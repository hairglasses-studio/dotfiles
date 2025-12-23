#!/usr/bin/env bash

# Basic sanity tests for the install.sh script.
#
# These tests exercise the various flags in a way that does not
# require network access.  If the script exits with a non‑zero
# status this test will fail.

set -euo pipefail

# Move to the repository root.  This file is located in test/.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

echo "Running install.sh --dry-run"
./install.sh --dry-run >/dev/null

echo "Running install.sh --dry-run --no-install"
./install.sh --dry-run --no-install >/dev/null

echo "Running install.sh --help"
./install.sh --help >/dev/null

echo "All install.sh tests passed"
