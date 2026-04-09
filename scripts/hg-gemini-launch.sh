#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/hg-agent-launch.sh"

hg_agent_launch_main \
  "gemini" \
  "/home/hg/.local/bin/gemini" \
  "$0" \
  --sandbox=false \
  --approval-mode \
  yolo \
  -- \
  "$@"
