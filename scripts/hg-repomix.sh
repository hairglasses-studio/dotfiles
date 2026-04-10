#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

usage() {
  cat <<'EOF'
Usage: hg-repomix.sh [repo_path] [repomix args...]

Wrap repomix with a workspace-safe default launcher. If the native `repomix`
binary is unavailable, the wrapper falls back to `npx --yes repomix`.
EOF
}

repo_path="."
if [[ $# -gt 0 ]]; then
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      ;;
    *)
      repo_path="$1"
      shift
      ;;
  esac
fi

repo_path="$(cd "$repo_path" && pwd)"

if command -v repomix >/dev/null 2>&1; then
  cd "$repo_path"
  exec repomix "$@"
fi

if command -v npx >/dev/null 2>&1; then
  cd "$repo_path"
  exec npx --yes repomix "$@"
fi

hg_die "repomix is not installed and npx is unavailable"
