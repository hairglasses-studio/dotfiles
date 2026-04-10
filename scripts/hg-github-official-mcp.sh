#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

hg_require docker gh

dotenv_github_pat="$(hg_env_file_value "$HOME/.env" "GITHUB_PAT")"

if [[ -n "$dotenv_github_pat" ]]; then
  TOKEN="$dotenv_github_pat"
elif [[ -n "${GITHUB_PERSONAL_ACCESS_TOKEN:-}" ]]; then
  TOKEN="$GITHUB_PERSONAL_ACCESS_TOKEN"
elif [[ -n "${GITHUB_PAT:-}" ]]; then
  TOKEN="$GITHUB_PAT"
elif [[ -n "${GITHUB_TOKEN:-}" ]]; then
  TOKEN="$GITHUB_TOKEN"
elif [[ -n "${GH_TOKEN:-}" ]]; then
  TOKEN="$GH_TOKEN"
else
  TOKEN="$(gh auth token)"
fi

[[ -n "$TOKEN" ]] || hg_die "No GitHub token available for github-mcp-server"

export GITHUB_PERSONAL_ACCESS_TOKEN="$TOKEN"
export GITHUB_TOOLSETS="${GITHUB_TOOLSETS:-default,stargazers}"

docker_args=(
  -i
  --rm
  -e GITHUB_PERSONAL_ACCESS_TOKEN
  -e GITHUB_TOOLSETS
)

for name in \
  GITHUB_TOOLS \
  GITHUB_DYNAMIC_TOOLSETS \
  GITHUB_READ_ONLY \
  GITHUB_INSIDERS \
  GITHUB_LOCKDOWN_MODE \
  GITHUB_MCP_SERVER_NAME \
  GITHUB_MCP_SERVER_TITLE
do
  if [[ -n "${!name:-}" ]]; then
    docker_args+=(-e "$name")
  fi
done

exec docker run "${docker_args[@]}" ghcr.io/github/github-mcp-server "$@"
