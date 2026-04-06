#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: hg-mcp-smoke.sh <repo_path> [server_name...]

Run a short startup smoke test for repo-local MCP entrypoints defined in .mcp.json.
Exit codes:
  0  all requested servers started cleanly
  1  one or more servers failed
EOF
}

if [[ $# -lt 1 ]]; then
  usage >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

if ! command -v timeout >/dev/null 2>&1; then
  echo "timeout is required" >&2
  exit 1
fi

REPO_PATH="$1"
shift
REPO_PATH="$(cd "$REPO_PATH" && pwd)"
MCP_JSON="$REPO_PATH/.mcp.json"
SMOKE_TIMEOUT="${SMOKE_TIMEOUT:-3s}"

if [[ ! -f "$MCP_JSON" ]]; then
  echo "Missing .mcp.json: $MCP_JSON" >&2
  exit 1
fi

declare -a SERVERS=()
if [[ $# -gt 0 ]]; then
  SERVERS=("$@")
else
  while IFS= read -r server; do
    [[ -n "$server" ]] || continue
    SERVERS+=("$server")
  done < <(jq -r '.mcpServers | keys[] | select(startswith("_") | not)' "$MCP_JSON")
fi

if [[ "${#SERVERS[@]}" -eq 0 ]]; then
  echo "No real MCP servers found in $MCP_JSON"
  exit 0
fi

failures=0

for server in "${SERVERS[@]}"; do
  command_name="$(jq -r --arg name "$server" '.mcpServers[$name].command // empty' "$MCP_JSON")"
  url_value="$(jq -r --arg name "$server" '.mcpServers[$name].url // empty' "$MCP_JSON")"
  cwd_value="$(jq -r --arg name "$server" '.mcpServers[$name].cwd // empty' "$MCP_JSON")"

  if [[ -n "$url_value" ]]; then
    echo "SKIP  $server  http transport ($url_value)"
    continue
  fi

  if [[ -z "$command_name" ]]; then
    echo "FAIL  $server  missing command"
    failures=$((failures + 1))
    continue
  fi

  run_dir="$REPO_PATH"
  if [[ -n "$cwd_value" ]]; then
    if [[ "$cwd_value" = /* ]]; then
      run_dir="$cwd_value"
    else
      run_dir="$REPO_PATH/$cwd_value"
    fi
  fi

  declare -a args=()
  while IFS= read -r arg; do
    args+=("$arg")
  done < <(jq -r --arg name "$server" '.mcpServers[$name].args[]?' "$MCP_JSON")

  declare -a env_args=()
  while IFS=$'\t' read -r env_key env_value; do
    [[ -n "$env_key" ]] || continue
    env_args+=("$env_key=$env_value")
  done < <(jq -r --arg name "$server" '.mcpServers[$name].env // {} | to_entries[] | [.key, .value] | @tsv' "$MCP_JSON")

  stdout_file="$(mktemp)"
  stderr_file="$(mktemp)"

  set +e
  (
    cd "$run_dir"
    env "${env_args[@]}" timeout "$SMOKE_TIMEOUT" "$command_name" "${args[@]}"
  ) >"$stdout_file" 2>"$stderr_file"
  status=$?
  set -e

  stderr_preview="$(sed -n '1,8p' "$stderr_file" | tr '\n' ' ' | sed 's/[[:space:]]\+/ /g; s/[[:space:]]$//')"

  if [[ "$status" -eq 0 || "$status" -eq 124 ]]; then
    if [[ -n "$stderr_preview" ]]; then
      echo "PASS  $server  status=$status  stderr: $stderr_preview"
    else
      echo "PASS  $server  status=$status"
    fi
  else
    echo "FAIL  $server  status=$status  stderr: ${stderr_preview:-<none>}"
    failures=$((failures + 1))
  fi

  rm -f "$stdout_file" "$stderr_file"
done

if [[ "$failures" -gt 0 ]]; then
  exit 1
fi
