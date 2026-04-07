#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
workspace_root="$(cd "$repo_root/.." && pwd)"
mode="${1:-sync}"

if [[ $# -gt 0 ]]; then
  shift
fi

repos=("$@")
if [[ "${#repos[@]}" -eq 0 ]]; then
  repos=(dotfiles-mcp process-mcp systemd-mcp tmux-mcp)
fi

case "$mode" in
  bootstrap|sync|check)
    ;;
  *)
    echo "usage: $(basename "$0") [bootstrap|sync|check] [repo ...]" >&2
    exit 2
    ;;
esac

for name in "${repos[@]}"; do
  bash "$repo_root/scripts/mcp-mirror.sh" "$mode" \
    --canonical "$repo_root/mcp/$name" \
    --mirror "$workspace_root/$name"
done
