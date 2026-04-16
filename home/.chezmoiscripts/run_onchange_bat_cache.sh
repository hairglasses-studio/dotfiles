#!/usr/bin/env bash
# Rebuild bat syntax theme cache when bat config changes.
# hash: {{ include "dot_config/symlink_bat" | sha256sum }}
set -euo pipefail

if command -v bat &>/dev/null; then
    echo "[chezmoi] Rebuilding bat cache..."
    bat cache --build 2>/dev/null || true
fi
