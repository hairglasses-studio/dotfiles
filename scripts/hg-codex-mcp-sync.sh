#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

START_MARKER="# BEGIN GENERATED MCP SERVERS: codex-mcp-sync"
END_MARKER="# END GENERATED MCP SERVERS: codex-mcp-sync"

DRY_RUN=false
REPO_PATH=""
POLICY_PATH=""
CONFIG_PATH=""
MCP_JSON_PATH=""

usage() {
  cat <<'EOF'
Usage: hg-codex-mcp-sync.sh <repo_path> [--policy <path>] [--config <path>] [--mcp-json <path>] [--dry-run]

Sync repo-local .mcp.json definitions into a generated MCP block inside .codex/config.toml.

Defaults:
  policy:   <repo>/.codex/mcp-profile-policy.json
  config:   <repo>/.codex/config.toml
  mcp-json: <repo>/.mcp.json
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --policy)
      [[ $# -ge 2 ]] || hg_die "--policy requires a path"
      POLICY_PATH="$2"
      shift 2
      ;;
    --config)
      [[ $# -ge 2 ]] || hg_die "--config requires a path"
      CONFIG_PATH="$2"
      shift 2
      ;;
    --mcp-json)
      [[ $# -ge 2 ]] || hg_die "--mcp-json requires a path"
      MCP_JSON_PATH="$2"
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
    -*)
      hg_die "Unknown argument: $1"
      ;;
    *)
      [[ -z "$REPO_PATH" ]] || hg_die "Only one repo path may be provided"
      REPO_PATH="$1"
      shift
      ;;
  esac
done

[[ -n "$REPO_PATH" ]] || {
  usage >&2
  exit 1
}

hg_require jq mktemp awk flock

REPO_PATH="$(cd "$REPO_PATH" && pwd)"
CONFIG_PATH="${CONFIG_PATH:-$REPO_PATH/.codex/config.toml}"
MCP_JSON_PATH="${MCP_JSON_PATH:-$REPO_PATH/.mcp.json}"
POLICY_PATH="${POLICY_PATH:-$REPO_PATH/.codex/mcp-profile-policy.json}"

[[ -f "$MCP_JSON_PATH" ]] || hg_die "Missing MCP source: $MCP_JSON_PATH"
[[ -f "$CONFIG_PATH" ]] || hg_die "Missing Codex config: $CONFIG_PATH"

if [[ ! -f "$POLICY_PATH" ]]; then
  hg_warn "Policy file not found; generating direct mappings from source server names"
fi

jq -e '.mcpServers and (.mcpServers | type == "object")' "$MCP_JSON_PATH" >/dev/null || hg_die "Invalid .mcp.json: expected top-level mcpServers object"
if [[ -f "$POLICY_PATH" ]]; then
  jq -e '.version and .profiles and (.profiles | type == "array")' "$POLICY_PATH" >/dev/null || hg_die "Invalid policy file: $POLICY_PATH"
  jq -e '
    [.profiles[].name]
    | group_by(.)
    | map(select(length > 1))
    | length == 0
  ' "$POLICY_PATH" >/dev/null || hg_die "Invalid policy file: duplicate generated profile names in $POLICY_PATH"
fi

toml_quote() {
  local s="$1"
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  printf '"%s"' "$s"
}

emit_scalar_line() {
  local key="$1"
  local value="$2"
  printf '%s = %s\n' "$key" "$(toml_quote "$value")"
}

emit_array_line() {
  local key="$1"
  shift
  local values=("$@")
  printf '%s = [' "$key"
  local i
  for i in "${!values[@]}"; do
    [[ "$i" -gt 0 ]] && printf ', '
    toml_quote "${values[$i]}"
  done
  printf ']\n'
}

emit_env_block() {
  local server_name="$1"
  local env_json="$2"
  local env_keys
  mapfile -t env_keys < <(jq -r 'keys[]' <<<"$env_json")
  [[ "${#env_keys[@]}" -gt 0 ]] || return 0
  printf '\n[mcp_servers.%s.env]\n' "$server_name"
  local key
  for key in "${env_keys[@]}"; do
    local value
    value="$(jq -r --arg key "$key" '.[$key]' <<<"$env_json")"
    emit_scalar_line "$key" "$value"
  done
}

profiles_json() {
  if [[ -f "$POLICY_PATH" ]]; then
    jq -c '.profiles[]' "$POLICY_PATH"
  else
    jq -c '.mcpServers | to_entries[] | {name: .key, from: .key, mode: "direct"}' "$MCP_JSON_PATH"
  fi
}

validate_shell_snippet() {
  local label="$1"
  local snippet="$2"

  if [[ "$snippet" == *"go run ./cmd/"* ]]; then
    hg_die "$label uses inline 'go run ./cmd/...'; wrap the server in a portable launcher script"
  fi

  if [[ "$snippet" == *"cd "* && "$snippet" == *"&&"* ]]; then
    hg_die "$label uses inline 'cd ... && ...'; resolve the repo/module root inside a portable launcher script"
  fi
}

validate_portable_launch() {
  local label="$1"
  local command="$2"
  local args_json="$3"
  local cwd="$4"

  if [[ "$cwd" == "." || "$cwd" == "./" ]]; then
    hg_die "$label uses cwd=$cwd; managed MCP launchers must not depend on repo-relative working directories"
  fi

  if [[ "$command" == "./"* || "$command" == "../"* ]]; then
    hg_die "$label uses repo-relative command $command; use an absolute launcher path or a shell wrapper"
  fi

  local first_arg=""
  first_arg="$(jq -r '.[0] // ""' <<<"$args_json")"

  if [[ "$command" == "go" || "$command" == */go ]]; then
    local second_arg=""
    second_arg="$(jq -r '.[1] // ""' <<<"$args_json")"
    if [[ "$first_arg" == "run" && "$second_arg" == ./cmd/* ]]; then
      hg_die "$label uses direct 'go run ./cmd/...'; use a portable launcher script instead"
    fi
  fi

  case "$command" in
    bash|sh|zsh|*/bash|*/sh|*/zsh)
      if [[ "$first_arg" == "./"* || "$first_arg" == "../"* ]]; then
        hg_die "$label uses a repo-relative shell script path ($first_arg); use a portable launcher path instead"
      fi
      local shell_arg
      while IFS= read -r shell_arg; do
        validate_shell_snippet "$label" "$shell_arg"
      done < <(jq -r '.[]' <<<"$args_json")
      ;;
  esac
}

validate_source_servers() {
  while IFS= read -r server; do
    [[ -n "$server" ]] || continue

    local name command cwd args_json
    name="$(jq -r '.name' <<<"$server")"
    command="$(jq -r '.command' <<<"$server")"
    cwd="$(jq -r '.cwd // ""' <<<"$server")"
    args_json="$(jq -c '.args // []' <<<"$server")"

    validate_portable_launch "source server $name" "$command" "$args_json" "$cwd"
  done < <(
    jq -c '
      .mcpServers
      | to_entries[]
      | {
          name: .key,
          command: .value.command,
          args: (.value.args // []),
          cwd: (.value.cwd // "")
        }
    ' "$MCP_JSON_PATH"
  )
}

acquire_write_lock() {
  $DRY_RUN && return 0

  local repo_name lock_path
  repo_name="$(basename "$REPO_PATH")"
  lock_path="$HG_STATE_DIR/codex-mcp-sync-${repo_name}.lock"
  mkdir -p "$(dirname "$lock_path")"
  exec 9>"$lock_path"
  if ! flock -n 9; then
    hg_die "Another Codex MCP sync is already running for $repo_name"
  fi
}

render_block() {
  printf '%s\n' "$START_MARKER"
  printf '# Generated by dotfiles/scripts/hg-codex-mcp-sync.sh from %s' "$(basename "$MCP_JSON_PATH")"
  if [[ -f "$POLICY_PATH" ]]; then
    printf ' and %s' "$(basename "$POLICY_PATH")"
  fi
  printf '\n'

  local first=1
  while IFS= read -r profile; do
    [[ -n "$profile" ]] || continue

    local resolved
    resolved="$(
      jq -cn \
        --argjson profile "$profile" \
        --slurpfile source "$MCP_JSON_PATH" '
          ($source[0].mcpServers[$profile.from]) as $server
          | if $server == null then
              error("missing source server: " + $profile.from)
            else
              {
                name: $profile.name,
                comment: ($profile.comment // ""),
                command: ($profile.override.command // $server.command),
                args: ($profile.override.args // $server.args // []),
                cwd: ($profile.override.cwd // $server.cwd // ""),
                env: (($server.env // {}) + ($profile.override.env // {})),
                enabled_tools: ($profile.enabled_tools // [])
              }
            end
        '
    )" || hg_die "Failed to resolve profile from source data"

    local name comment command cwd args_json env_json tools_json
    name="$(jq -r '.name' <<<"$resolved")"
    comment="$(jq -r '.comment' <<<"$resolved")"
    command="$(jq -r '.command' <<<"$resolved")"
    cwd="$(jq -r '.cwd' <<<"$resolved")"
    args_json="$(jq -c '.args' <<<"$resolved")"
    env_json="$(jq -c '.env' <<<"$resolved")"
    tools_json="$(jq -c '.enabled_tools' <<<"$resolved")"

    [[ -n "$command" && "$command" != "null" ]] || hg_die "Profile $name resolved to an empty command"
    validate_portable_launch "profile $name" "$command" "$args_json" "$cwd"

    [[ "$first" -eq 1 ]] || printf '\n'
    first=0

    [[ -n "$comment" ]] && printf '# %s\n' "$comment"
    printf '[mcp_servers.%s]\n' "$name"
    emit_scalar_line "command" "$command"

    local args
    mapfile -t args < <(jq -r '.[]' <<<"$args_json")
    if [[ "${#args[@]}" -gt 0 ]]; then
      emit_array_line "args" "${args[@]}"
    fi

    if [[ -n "$cwd" && "$cwd" != "." && "$cwd" != "./" ]]; then
      emit_scalar_line "cwd" "$cwd"
    fi

    local tools
    mapfile -t tools < <(jq -r '.[]' <<<"$tools_json")
    if [[ "${#tools[@]}" -gt 0 ]]; then
      printf 'enabled_tools = [\n'
      local tool
      for tool in "${tools[@]}"; do
        printf '  %s,\n' "$(toml_quote "$tool")"
      done
      printf ']\n'
    fi

    emit_env_block "$name" "$env_json"
  done < <(profiles_json)

  printf '\n%s\n' "$END_MARKER"
}

replace_marked_region() {
  local file="$1"
  local block_file="$2"
  local tmp
  tmp="$(mktemp)"
  awk -v start="$START_MARKER" -v end="$END_MARKER" -v block="$block_file" '
    BEGIN {
      while ((getline line < block) > 0) {
        generated = generated line "\n"
      }
      close(block)
      replaced = 0
      skipping = 0
    }
    index($0, start) == 1 {
      printf "%s", generated
      skipping = 1
      replaced = 1
      next
    }
    index($0, end) == 1 {
      skipping = 0
      next
    }
    !skipping { print }
    END {
      if (skipping) {
        exit 43
      }
      if (!replaced) {
        exit 42
      }
    }
  ' "$file" >"$tmp" || {
    local status=$?
    rm -f "$tmp"
    case "$status" in
      42)
        hg_die "Generated MCP block markers not found in $file"
        ;;
      43)
        hg_die "Generated MCP block start marker in $file is missing a matching end marker"
        ;;
      *)
        return "$status"
        ;;
    esac
  }
  cat "$tmp"
  rm -f "$tmp"
}

insert_new_region() {
  local file="$1"
  local block_file="$2"
  awk -v block="$block_file" '
    BEGIN {
      while ((getline line < block) > 0) {
        generated = generated line "\n"
      }
      close(block)
      inserted = 0
    }
    /^\[/ && !inserted {
      printf "%s\n", generated
      inserted = 1
    }
    { print }
    END {
      if (!inserted) {
        if (NR > 0) {
          print ""
        }
        printf "%s", generated
      }
    }
  ' "$file"
}

validate_source_servers
acquire_write_lock

if grep -q '^\[mcp_servers\.' "$CONFIG_PATH" && ! grep -q '^# BEGIN GENERATED MCP SERVERS: codex-mcp-sync$' "$CONFIG_PATH"; then
  hg_die "Unmanaged [mcp_servers.*] blocks already exist in $CONFIG_PATH. Add markers first or clean them up before syncing."
fi

generated_block_file="$(mktemp)"
output_file="$(mktemp)"
trap 'rm -f "$generated_block_file" "$output_file"' EXIT

render_block >"$generated_block_file"

if grep -q '^# BEGIN GENERATED MCP SERVERS: codex-mcp-sync$' "$CONFIG_PATH"; then
  replace_marked_region "$CONFIG_PATH" "$generated_block_file" >"$output_file"
else
  insert_new_region "$CONFIG_PATH" "$generated_block_file" >"$output_file"
fi

if $DRY_RUN; then
  diff -u "$CONFIG_PATH" "$output_file" || true
  exit 0
fi

mv "$output_file" "$CONFIG_PATH"
hg_ok "Synced Codex MCP block in $CONFIG_PATH"
