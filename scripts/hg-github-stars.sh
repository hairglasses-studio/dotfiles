#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

cd "$HG_DOTFILES/mcp/dotfiles-mcp"
exec env GOWORK=off go run ./cmd/github-starsctl "$@"
