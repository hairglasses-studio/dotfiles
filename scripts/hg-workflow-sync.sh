#!/usr/bin/env bash
# hg-workflow-sync.sh — Hosted workflow sync retired under the local-only automation policy.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

DRY_RUN=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    -h|--help)
      cat <<'EOF'
Usage: hg-workflow-sync.sh [--dry-run]

Hosted GitHub workflow sync is retired. This command is informational only and
does not create, update, commit, or push workflow files.
EOF
      exit 0
      ;;
    *)
      hg_warn "Ignoring retired argument: $arg"
      ;;
  esac
done

hg_info "Hosted workflow sync is retired under the local-only automation policy"
$DRY_RUN && hg_warn "Dry-run mode is informational only"
echo ""
hg_ok "No hosted workflows are managed by hg-workflow-sync.sh"
