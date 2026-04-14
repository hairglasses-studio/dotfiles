#!/usr/bin/env bash
set -euo pipefail

# Keep Claude on the user-owned runtime/auth path; the managed root launcher
# flow is for the other agent CLIs and breaks current Claude auth on this host.
exec /home/hg/.local/bin/claude "$@"
