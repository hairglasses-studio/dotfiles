#!/usr/bin/env bash
set -euo pipefail

WORKSPACE_ROOT="${WORKSPACE_ROOT:-$HOME/hairglasses-studio}"
CONFIG_PATH="${CONFIG_PATH:-$HOME/.codex/config.toml}"
POLICY_PATH="${POLICY_PATH:-}"
DRY_RUN=false

usage() {
  cat <<'EOF'
Usage: hg-codex-global-mcp-sync.sh [--workspace-root <path>] [--config <path>] [--policy <path>] [--dry-run]

Generate a user-global MCP block in ~/.codex/config.toml from every .mcp.json
found under the hairglasses-studio workspace. If --policy is omitted, the
sync will auto-load workspace/mcp-global-policy.json when present.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --workspace-root)
      [[ $# -ge 2 ]] || {
        echo "--workspace-root requires a path" >&2
        exit 1
      }
      WORKSPACE_ROOT="$2"
      shift 2
      ;;
    --config)
      [[ $# -ge 2 ]] || {
        echo "--config requires a path" >&2
        exit 1
      }
      CONFIG_PATH="$2"
      shift 2
      ;;
    --policy)
      [[ $# -ge 2 ]] || {
        echo "--policy requires a path" >&2
        exit 1
      }
      POLICY_PATH="$2"
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

command -v go >/dev/null 2>&1 || {
  echo "go is required" >&2
  exit 1
}

args=(
  go
  run
  "$WORKSPACE_ROOT/codexkit/cmd/codexkit-global-mcp"
  --workspace-root
  "$WORKSPACE_ROOT"
  --config
  "$CONFIG_PATH"
)

if [[ -n "$POLICY_PATH" ]]; then
  args+=(
    --policy
    "$POLICY_PATH"
  )
fi

if [[ "$DRY_RUN" == true ]]; then
  args+=(--dry-run)
fi

GOCACHE="${GOCACHE:-/tmp/codexkit-global-gocache}" \
GOTMPDIR="${GOTMPDIR:-/tmp}" \
  "${args[@]}"
