#!/usr/bin/env bash
set -euo pipefail
repo_url="$(git config --get remote.origin.url || true)"
if [[ -z "$repo_url" ]]; then
  echo "No remote.origin.url found; cannot update badges." >&2
  exit 1
fi
# Normalize to https://github.com/OWNER/REPO.git
case "$repo_url" in
  git@github.com:*) slug="${repo_url#git@github.com:}";;
  https://github.com/*) slug="${repo_url#https://github.com/}";;
  *) echo "Unsupported remote URL: $repo_url" >&2; exit 2;;
esac
slug="${slug%.git}"
owner="${slug%%/*}"
repo="${slug##*/}"
sed -i.bak "s|github.com/OWNER/REPO|github.com/${owner}/${repo}|g" README.md
rm -f README.md.bak
echo "Updated badges to github.com/${owner}/${repo}"
