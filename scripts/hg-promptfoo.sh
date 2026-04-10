#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

usage() {
  cat <<'EOF'
Usage: hg-promptfoo.sh [repo_path] [promptfoo args...]

Run promptfoo from a repo root. The wrapper prefers a native `promptfoo`
binary and falls back to `npx --yes promptfoo`.
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

if command -v promptfoo >/dev/null 2>&1; then
  cd "$repo_path"
  exec promptfoo "$@"
fi

if command -v npx >/dev/null 2>&1; then
  cd "$repo_path"
  exec npx --yes promptfoo "$@"
fi

hg_die "promptfoo is not installed and npx is unavailable"
