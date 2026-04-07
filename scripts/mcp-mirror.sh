#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  mcp-mirror.sh bootstrap --canonical <path> --mirror <path>
  mcp-mirror.sh sync  --canonical <path> --mirror <path>
  mcp-mirror.sh check --canonical <path> --mirror <path>

Sync or compare a canonical MCP source tree with a standalone publish mirror while
preserving standalone-only repository metadata such as workflows and release files.
EOF
}

mode="${1:-}"
shift || true

canonical=""
mirror=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --canonical)
      shift
      canonical="${1:-}"
      ;;
    --mirror)
      shift
      mirror="${1:-}"
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
  shift || true
done

[[ -n "$mode" && -n "$canonical" && -n "$mirror" ]] || {
  usage
  exit 2
}

mirror_excludes=(
  --exclude '.git/'
  --exclude '.github/'
  --exclude '.well-known/'
  --exclude '.codex/'
  --exclude '.gemini/'
  --exclude '.ralph/'
  --exclude 'coverage.out'
  --exclude 'coverage.html'
  --exclude '*.test'
  --exclude '*.log'
  --exclude 'dotfiles-mcp'
  --exclude 'process-mcp'
  --exclude 'systemd-mcp'
  --exclude 'tmux-mcp'
  --exclude '.DS_Store'
  --exclude '.goreleaser.yaml'
  --exclude '.goreleaser.yml'
  --exclude '.editorconfig'
  --exclude '.golangci.yml'
)

case "$mode" in
  bootstrap)
    rsync -a --delete "${mirror_excludes[@]}" "$mirror"/ "$canonical"/
    ;;
  sync)
    rsync -a --delete "${mirror_excludes[@]}" "$canonical"/ "$mirror"/
    ;;
  check)
    tmp_canonical="$(mktemp -d)"
    tmp_mirror="$(mktemp -d)"
    trap 'rm -rf "$tmp_canonical" "$tmp_mirror"' EXIT
    rsync -a "${mirror_excludes[@]}" "$canonical"/ "$tmp_canonical"/
    rsync -a "${mirror_excludes[@]}" "$mirror"/ "$tmp_mirror"/
    diff -ruN "$tmp_canonical" "$tmp_mirror"
    ;;
  *)
    usage
    exit 2
    ;;
esac
