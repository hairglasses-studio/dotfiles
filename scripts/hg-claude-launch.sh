#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/hg-agent-launch.sh"

hg_agent_launch_main \
  "claude" \
  "/home/hg/.local/bin/claude" \
  "$0" \
  --dangerously-skip-permissions \
  -- \
  "$@"
