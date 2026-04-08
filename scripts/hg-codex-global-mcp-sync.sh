#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

ROOT_ARGS=()
CODEX_ARGS=()
DRY_RUN=false
LEGACY_POLICY_PATH=""

usage() {
  cat <<'EOF'
Usage: hg-codex-global-mcp-sync.sh [--workspace-root <path>] [--config <path>] [--policy <path>] [--dry-run]

Deprecated compatibility wrapper around hg-workspace-global-sync.sh.

This command now delegates to:
  hg-workspace-global-sync.sh --tools-only

Argument mapping:
  --workspace-root -> --root
  --config         -> --codex-config

Note: the legacy --policy flag is no longer supported because workspace-global
sync derives curated Codex overlays from repo-local metadata plus the workspace
manifest. If you pass --policy, this wrapper exits with an error.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace-root)
      [[ $# -ge 2 ]] || {
        echo "--workspace-root requires a path" >&2
        exit 1
      }
      ROOT_ARGS+=(--root "$2")
      shift 2
      ;;
    --config)
      [[ $# -ge 2 ]] || {
        echo "--config requires a path" >&2
        exit 1
      }
      CODEX_ARGS+=(--codex-config "$2")
      shift 2
      ;;
    --policy)
      [[ $# -ge 2 ]] || {
        echo "--policy requires a path" >&2
        exit 1
      }
      LEGACY_POLICY_PATH="$2"
      shift 2
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ -n "$LEGACY_POLICY_PATH" ]]; then
  echo "error: --policy is not supported by the workspace-global sync path; move any remaining policy logic into repo-local .codex/mcp-profile-policy.json files" >&2
  exit 1
fi

echo "warning: hg-codex-global-mcp-sync.sh is deprecated; use hg-workspace-global-sync.sh --tools-only" >&2

ARGS=(
  --tools-only
  "${ROOT_ARGS[@]}"
  "${CODEX_ARGS[@]}"
)

if [[ "$DRY_RUN" == true ]]; then
  ARGS+=(--dry-run)
fi

exec "$SCRIPT_DIR/hg-workspace-global-sync.sh" "${ARGS[@]}"
