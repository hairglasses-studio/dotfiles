#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"
WORKSPACE_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
ROOT_MCP_PATH="${HG_ROOT_MCP_PATH:-}"
ANTIGRAVITY_DIR="${HG_ANTIGRAVITY_DIR:-$HOME/.gemini/antigravity}"
ANTIGRAVITY_MCP_PATH="${HG_ANTIGRAVITY_MCP_CONFIG_PATH:-$ANTIGRAVITY_DIR/mcp_config.json}"
ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR="${HG_ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR:-$ANTIGRAVITY_DIR/global_workflows}"
ANTIGRAVITY_METADATA_PATH="${HG_ANTIGRAVITY_METADATA_PATH:-$ANTIGRAVITY_DIR/.hg-antigravity-sync.json}"
ANTIGRAVITY_ECOSYSTEM_METADATA_PATH="${HG_ANTIGRAVITY_ECOSYSTEM_METADATA_PATH:-$ANTIGRAVITY_DIR/.hg-antigravity-ecosystem.json}"
ANTIGRAVITY_SETTINGS_PATH="${HG_ANTIGRAVITY_SETTINGS_PATH:-$HOME/.config/Antigravity/User/settings.json}"
ANTIGRAVITY_ENV_MANIFEST_PATH="${HG_ANTIGRAVITY_ENV_MANIFEST_PATH:-$HOME/.config/antigravity/env-sources.json}"
ANTIGRAVITY_LEGACY_ENV_PATH="${HG_ANTIGRAVITY_LEGACY_ENV_PATH:-$HOME/.config/antigravity/env.sh}"
ANTIGRAVITY_WRAPPER_PATH="${HG_ANTIGRAVITY_WRAPPER_PATH:-$HOME/.local/bin/antigravity-managed}"
ANTIGRAVITY_DESKTOP_PATH="${HG_ANTIGRAVITY_DESKTOP_PATH:-$HOME/.local/share/applications/antigravity.desktop}"
ANTIGRAVITY_WORKSPACE_FILE_PATH="${HG_ANTIGRAVITY_WORKSPACE_FILE_PATH:-$WORKSPACE_ROOT/hairglasses-studio.code-workspace}"
ANTIGRAVITY_HOME_DOC_PATH="${HG_ANTIGRAVITY_HOME_DOC_PATH:-$ANTIGRAVITY_DIR/GEMINI.md}"
WORKSPACE_RULE_PATH="${HG_ANTIGRAVITY_RULE_PATH:-$WORKSPACE_ROOT/.agents/rules/antigravity-workspace.md}"
WORKSPACE_SKILLS_DIR="${HG_ANTIGRAVITY_WORKSPACE_SKILLS_DIR:-$WORKSPACE_ROOT/.agents/skills}"
WORKSPACE_WORKFLOWS_DIR="${HG_ANTIGRAVITY_WORKSPACE_WORKFLOWS_DIR:-$WORKSPACE_ROOT/.agents/workflows}"
PERMISSION_PRESET_SOURCE="${HG_ANTIGRAVITY_PERMISSION_PRESET_SOURCE:-$HG_DOTFILES/config/antigravity/permission-presets.md}"
PERMISSION_PRESET_TARGET="${HG_ANTIGRAVITY_PERMISSION_PRESET_TARGET:-$ANTIGRAVITY_DIR/permission-presets.md}"
ECOSYSTEM_SYNC_SCRIPT="${HG_ANTIGRAVITY_ECOSYSTEM_SYNC_SCRIPT:-$SCRIPT_DIR/hg-antigravity-ecosystem-sync.sh}"
ANTIGRAVITY_HOME_START_MARKER="<!-- BEGIN GENERATED ANTIGRAVITY HOME: hg-antigravity-sync -->"
ANTIGRAVITY_HOME_END_MARKER="<!-- END GENERATED ANTIGRAVITY HOME: hg-antigravity-sync -->"
WORKSPACE_HOME_START_MARKER="<!-- BEGIN GENERATED WORKSPACE GLOBAL: hg-workspace-global-sync -->"

pending=0
managed_workspace_skill_updates=0
managed_workspace_workflow_updates=0
missing_env_vars=0
MCP_TOTAL_SERVER_COUNT=0
MCP_ROOT_SHARED_SERVER_COUNT=0
WORKFLOW_REPO_COUNT=0
WORKFLOW_SKILL_COUNT=0
WORKFLOW_WORKSPACE_COUNT=0
WORKFLOW_GLOBAL_COUNT=0
WORKSPACE_SKILL_COUNT=0
GLOBAL_SKILL_COUNT=0

declare -A server_name_seen=()
declare -A workflow_name_seen=()
declare -A workspace_skill_name_seen=()
declare -A imported_env_values=()
declare -A imported_env_sources=()
declare -A managed_workspace_workflow_set=()
declare -A managed_workspace_skill_set=()

usage() {
  cat <<'EOF'
Usage: hg-antigravity-sync.sh [options]

Generate and sync local Antigravity MCP, skills, workflows, launcher, settings,
and runtime env bridge artifacts for the hairglasses-studio workspace.

Options:
  --root <path>                    Workspace root (default: ~/hairglasses-studio)
  --root-mcp <path>                Shared root .mcp.json (default: <root>/.mcp.json)
  --antigravity-dir <path>         Antigravity managed state dir (default: ~/.gemini/antigravity)
  --antigravity-mcp <path>         Antigravity MCP config path
  --antigravity-settings <path>    Antigravity user settings path
  --antigravity-env <path>         Runtime env manifest path
  --antigravity-home-doc <path>    Antigravity home GEMINI.md path
  --antigravity-wrapper <path>     Managed Antigravity launcher wrapper path
  --antigravity-desktop <path>     User-local Antigravity desktop entry path
  --antigravity-workspace <path>   Managed default Antigravity workspace file path
  --rule-path <path>               Workspace Antigravity rule file path
  --dry-run                        Report changes without writing
  --check                          Exit non-zero if managed state is stale
  -h, --help                       Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --root)
      [[ $# -ge 2 ]] || hg_die "--root requires a path"
      WORKSPACE_ROOT="$2"
      shift 2
      ;;
    --root-mcp)
      [[ $# -ge 2 ]] || hg_die "--root-mcp requires a path"
      ROOT_MCP_PATH="$2"
      shift 2
      ;;
    --antigravity-dir)
      [[ $# -ge 2 ]] || hg_die "--antigravity-dir requires a path"
      ANTIGRAVITY_DIR="$2"
      shift 2
      ;;
    --antigravity-mcp)
      [[ $# -ge 2 ]] || hg_die "--antigravity-mcp requires a path"
      ANTIGRAVITY_MCP_PATH="$2"
      shift 2
      ;;
    --antigravity-settings)
      [[ $# -ge 2 ]] || hg_die "--antigravity-settings requires a path"
      ANTIGRAVITY_SETTINGS_PATH="$2"
      shift 2
      ;;
    --antigravity-env)
      [[ $# -ge 2 ]] || hg_die "--antigravity-env requires a path"
      ANTIGRAVITY_ENV_MANIFEST_PATH="$2"
      shift 2
      ;;
    --antigravity-home-doc)
      [[ $# -ge 2 ]] || hg_die "--antigravity-home-doc requires a path"
      ANTIGRAVITY_HOME_DOC_PATH="$2"
      shift 2
      ;;
    --antigravity-wrapper)
      [[ $# -ge 2 ]] || hg_die "--antigravity-wrapper requires a path"
      ANTIGRAVITY_WRAPPER_PATH="$2"
      shift 2
      ;;
    --antigravity-desktop)
      [[ $# -ge 2 ]] || hg_die "--antigravity-desktop requires a path"
      ANTIGRAVITY_DESKTOP_PATH="$2"
      shift 2
      ;;
    --antigravity-workspace)
      [[ $# -ge 2 ]] || hg_die "--antigravity-workspace requires a path"
      ANTIGRAVITY_WORKSPACE_FILE_PATH="$2"
      shift 2
      ;;
    --rule-path)
      [[ $# -ge 2 ]] || hg_die "--rule-path requires a path"
      WORKSPACE_RULE_PATH="$2"
      shift 2
      ;;
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --check)
      MODE="check"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $1"
      ;;
  esac
done

hg_require jq mktemp diff find sort awk realpath cmp /usr/bin/antigravity

WORKSPACE_ROOT="$(cd "$WORKSPACE_ROOT" && pwd)"
ROOT_MCP_PATH="${ROOT_MCP_PATH:-$WORKSPACE_ROOT/.mcp.json}"
[[ -d "$WORKSPACE_ROOT" ]] || hg_die "Workspace root not found: $WORKSPACE_ROOT"
[[ -f "$ROOT_MCP_PATH" ]] || hg_die "Shared root MCP file not found: $ROOT_MCP_PATH"

MODE_ARGS=()
case "$MODE" in
  dry-run) MODE_ARGS=(--dry-run) ;;
  check) MODE_ARGS=(--check) ;;
esac

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

expand_home_tokens() {
  local value="$1"
  value="${value//\$\{HOME\}/$HOME}"
  value="${value//\$HOME/$HOME}"
  printf '%s\n' "$value"
}

sanitize_kebab_name() {
  local raw="$1"
  printf '%s' "$raw" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g'
}

json_string_array_from_lines() {
  local file="$1"
  if [[ ! -s "$file" ]]; then
    printf '[]\n'
    return 0
  fi
  jq -Rsc 'split("\n") | map(select(length > 0))' <"$file"
}

normalize_env_json() {
  local env_json="$1"
  jq -c --arg home "$HOME" '
    (if type == "object" then . else {} end)
    | with_entries(
        .value |= if type == "string"
          then gsub("\\$\\{HOME\\}"; $home) | gsub("\\$HOME"; $home)
          else .
          end
      )
  ' <<<"$env_json"
}

normalize_cwd() {
  local base_dir="$1"
  local cwd="$2"
  cwd="$(expand_home_tokens "$cwd")"
  if [[ -z "$cwd" ]]; then
    printf '\n'
    return 0
  fi

  if [[ "$cwd" == "." ]]; then
    cwd="$base_dir"
  elif [[ "$cwd" != /* ]]; then
    cwd="$base_dir/$cwd"
  fi

  if [[ ! -d "$cwd" ]]; then
    printf '%s\n' "$cwd"
    return 0
  fi

  realpath "$cwd"
}

extract_frontmatter_scalar() {
  local file="$1"
  local key="$2"
  awk -v key="$key" '
    BEGIN { in_frontmatter = 0 }
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { exit }
    in_frontmatter && $0 ~ ("^" key ":[[:space:]]*") {
      sub("^" key ":[[:space:]]*", "", $0)
      gsub(/^["'\''"]|["'\''"]$/, "", $0)
      print
      exit
    }
  ' "$file"
}

validate_markdown_budget() {
  local path="$1"
  local label="$2"
  local length
  length="$(wc -m <"$path" | tr -d ' ')"
  if [[ "$length" -gt 12000 ]]; then
    hg_die "$label exceeds Antigravity 12,000 character limit: $path ($length chars)"
  fi
}

sync_rendered_file() {
  local target="$1"
  local rendered="$2"
  local label="$3"

  if [[ -f "$target" ]] && cmp -s "$target" "$rendered"; then
    rm -f "$rendered"
    return 0
  fi

  pending=1
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      mv "$rendered" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      rm -f "$rendered"
      hg_warn "Would update $label: $target"
      ;;
    check)
      rm -f "$rendered"
      hg_warn "Out of date $label: $target"
      ;;
  esac
}

replace_marked_region() {
  local file="$1"
  local block_file="$2"
  local start_marker="$3"
  local end_marker="$4"
  local tmp
  tmp="$(mktemp)"

  awk -v start="$start_marker" -v end="$end_marker" -v block="$block_file" '
    BEGIN {
      while ((getline line < block) > 0) {
        generated = generated line "\n"
      }
      close(block)
      replaced = 0
      skipping = 0
    }
    $0 == start {
      printf "%s", generated
      replaced = 1
      skipping = 1
      next
    }
    $0 == end {
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
        hg_die "Generated block markers not found in $file"
        ;;
      43)
        hg_die "Generated block start marker in $file is missing a matching end marker"
        ;;
      *)
        return "$status"
        ;;
    esac
  }

  cat "$tmp"
  rm -f "$tmp"
}

replace_legacy_prefix() {
  local target="$1"
  local legacy_file="$2"
  local block_file="$3"
  local legacy_lines
  legacy_lines="$(wc -l <"$legacy_file" | tr -d ' ')"

  if cmp -s "$target" "$legacy_file"; then
    cat "$block_file"
    return 0
  fi

  if cmp -s <(head -n "$legacy_lines" "$target") "$legacy_file"; then
    cat "$block_file"
    tail -n "+$((legacy_lines + 1))" "$target"
    return 0
  fi

  return 1
}

remove_managed_path() {
  local target="$1"
  local label="$2"
  [[ -e "$target" ]] || return 0

  pending=1
  case "$MODE" in
    write)
      rm -rf "$target"
      hg_ok "Removed stale $label: $target"
      ;;
    dry-run)
      hg_warn "Would remove stale $label: $target"
      ;;
    check)
      hg_warn "Unexpected stale $label: $target"
      ;;
  esac
}

sync_symlink() {
  local target="$1"
  local source="$2"
  local label="$3"

  if [[ -e "$target" && ! -L "$target" ]]; then
    hg_die "$label target exists and is not a symlink: $target"
  fi
  if [[ -L "$target" && "$(readlink "$target")" == "$source" ]]; then
    return 0
  fi

  pending=1
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      ln -sfn "$source" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      hg_warn "Would update $label: $target -> $source"
      ;;
    check)
      hg_warn "Out of date $label: $target"
      ;;
  esac
}

list_workspace_member_repos() {
  find "$WORKSPACE_ROOT" -mindepth 2 -maxdepth 2 -type d -name .git -prune \
    | sed 's#/\.git$##' \
    | sort
}

append_antigravity_server() {
  local output_file="$1"
  local server_name="$2"
  local server_json="$3"
  local base_dir="$4"

  if [[ -n "${server_name_seen[$server_name]:-}" ]]; then
    hg_die "Duplicate Antigravity MCP server name: $server_name"
  fi
  server_name_seen["$server_name"]=1

  local sanitized normalized_cwd normalized_env
  sanitized="$(jq -cS '
    if type != "object" then
      {}
    else
      (
        (if (.command // null) != null then {command: .command} else {} end)
        + (if (.args // null) != null then {args: (.args // [])} else {} end)
        + (if (.env // null) != null then {env: (.env // {})} else {} end)
        + (if (.cwd // null) != null then {cwd: .cwd} else {} end)
        + (if (.headers // null) != null then {headers: .headers} else {} end)
        + (if (.authProviderType // null) != null then {authProviderType: .authProviderType} else {} end)
        + (if (.oauth // null) != null then {oauth: .oauth} else {} end)
        + (if (.disabled // null) != null then {disabled: .disabled} else {} end)
        + (if (.disabledTools // null) != null then {disabledTools: (.disabledTools // [])} else {} end)
        + (
            if (.serverUrl // null) != null then {serverUrl: .serverUrl}
            elif (.httpUrl // null) != null then {serverUrl: .httpUrl}
            elif (.url // null) != null then {serverUrl: .url}
            else {}
            end
          )
      )
    end
  ' <<<"$server_json")"
  normalized_cwd="$(normalize_cwd "$base_dir" "$(jq -r '.cwd // ""' <<<"$sanitized")")"
  normalized_env="$(normalize_env_json "$(jq -c '.env // {}' <<<"$sanitized")")"

  jq -cn \
    --arg name "$server_name" \
    --arg cwd "$normalized_cwd" \
    --argjson server "$sanitized" \
    --argjson env "$normalized_env" '
      {
        name: $name,
        value: (
          ($server | del(.cwd, .env))
          + (if ($cwd | length) > 0 then {cwd: $cwd} else {} end)
          + (if ($env | length) > 0 then {env: $env} else {} end)
        )
      }
    ' >>"$output_file"
}

build_antigravity_mcp_config() {
  local output="$tmpdir/antigravity-mcp.ndjson"
  : >"$output"

  while IFS= read -r entry; do
    [[ -n "$entry" ]] || continue
    append_antigravity_server \
      "$output" \
      "$(jq -r '.name' <<<"$entry")" \
      "$(jq -c '.value' <<<"$entry")" \
      "$WORKSPACE_ROOT"
  done < <(
    jq -c '
      (.mcpServers // {})
      | to_entries[]
      | select(.key | startswith("_") | not)
      | {name: .key, value: .value}
    ' "$ROOT_MCP_PATH"
  )

  while IFS= read -r repo; do
    [[ -f "$repo/.mcp.json" ]] || continue
    local repo_key
    repo_key="$(sanitize_kebab_name "$(basename "$repo")")"
    while IFS= read -r entry; do
      [[ -n "$entry" ]] || continue
      append_antigravity_server \
        "$output" \
        "${repo_key}-$(sanitize_kebab_name "$(jq -r '.name' <<<"$entry")")" \
        "$(jq -c '.value' <<<"$entry")" \
        "$repo"
    done < <(
      jq -c '
        (.mcpServers // {})
        | to_entries[]
        | select(.key | startswith("_") | not)
        | {name: .key, value: .value}
      ' "$repo/.mcp.json"
    )
  done < <(list_workspace_member_repos)

  if [[ -s "$output" ]]; then
    jq -S '{mcpServers: ((map({(.name): .value}) | add) // {})}' < <(jq -sc '.' "$output") >"$tmpdir/mcp_config.json"
  else
    printf '{ "mcpServers": {} }\n' >"$tmpdir/mcp_config.json"
  fi

  MCP_TOTAL_SERVER_COUNT="$(jq -r '.mcpServers | keys | length' "$tmpdir/mcp_config.json")"
  MCP_ROOT_SHARED_SERVER_COUNT="$(jq -r '(.mcpServers // {}) | keys | length' "$ROOT_MCP_PATH")"

  sync_rendered_file "$ANTIGRAVITY_MCP_PATH" "$tmpdir/mcp_config.json" "Synced Antigravity MCP config"
}

render_workspace_rule_file() {
  {
    printf '# Antigravity Workspace Rule\n\n'
    printf 'Use this workspace as a control plane first, not as a flat file dump.\n\n'
    printf -- '- Read `%s/CLAUDE.md` and `%s/AGENTS.md` before making repo-wide assumptions.\n' "$WORKSPACE_ROOT" "$WORKSPACE_ROOT"
    printf -- '- Use `%s/workspace/manifest.json` as the canonical managed-repo inventory.\n' "$WORKSPACE_ROOT"
    printf -- '- Check `%s/docs/indexes/SEARCH-GUIDE.md` before net-new research.\n' "$WORKSPACE_ROOT"
    printf -- '- Prefer `%s/docs/` for org-wide context and `%s/whiteclaw/` as a reusable reference repo.\n' "$WORKSPACE_ROOT" "$WORKSPACE_ROOT"
    printf -- '- Prefer repo-local `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` before broad scans.\n'
    printf -- '- Skip generated and cache-heavy trees unless the task explicitly targets them: `workspace/`, `.ralph/`, `.tools/`, `node_modules/`, `.venv/`, `.cache/`, `dist/`, `build/`, `jellyfin-stack/vendor/`, and `.agentMemory/`.\n'
  } >"$tmpdir/antigravity-workspace-rule.md"
  validate_markdown_budget "$tmpdir/antigravity-workspace-rule.md" "Antigravity workspace rule"
  sync_rendered_file "$WORKSPACE_RULE_PATH" "$tmpdir/antigravity-workspace-rule.md" "Synced workspace Antigravity rule"
}

render_home_gemini_doc() {
  local legacy_doc="$tmpdir/GEMINI-legacy.md"
  local block_doc="$tmpdir/GEMINI-block.md"
  local merged_doc="$tmpdir/GEMINI.md"

  {
    printf '# Antigravity Home Rule\n\n'
    printf 'Apply these defaults when working in `/home/hg/hairglasses-studio`.\n\n'
    printf -- '- Start with `%s` when the current workspace is the shared studio root.\n' "$WORKSPACE_RULE_PATH"
    printf -- '- Prefer workspace-local skills in `%s/` before global skills.\n' "$WORKSPACE_SKILLS_DIR"
    printf -- '- Prefer workspace-local workflows in `%s/` before global workflows.\n' "$WORKSPACE_WORKFLOWS_DIR"
    printf -- '- Honor repo-local `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` as the first repo-specific instruction layer.\n'
    printf -- '- Use the managed launcher so OpenAI, Anthropic, Google, and Gemini provider keys are runtime-resolved from local `.env` files.\n'
  } >"$legacy_doc"

  {
    printf '%s\n' "$ANTIGRAVITY_HOME_START_MARKER"
    cat "$legacy_doc"
    printf '\n%s\n' "$ANTIGRAVITY_HOME_END_MARKER"
  } >"$block_doc"

  if [[ -f "$ANTIGRAVITY_HOME_DOC_PATH" ]]; then
    if grep -Fqx "$ANTIGRAVITY_HOME_START_MARKER" "$ANTIGRAVITY_HOME_DOC_PATH"; then
      replace_marked_region \
        "$ANTIGRAVITY_HOME_DOC_PATH" \
        "$block_doc" \
        "$ANTIGRAVITY_HOME_START_MARKER" \
        "$ANTIGRAVITY_HOME_END_MARKER" >"$merged_doc"
    elif replace_legacy_prefix "$ANTIGRAVITY_HOME_DOC_PATH" "$legacy_doc" "$block_doc" >"$merged_doc"; then
      :
    elif grep -Fqx "$WORKSPACE_HOME_START_MARKER" "$ANTIGRAVITY_HOME_DOC_PATH"; then
      {
        cat "$block_doc"
        printf '\n'
        cat "$ANTIGRAVITY_HOME_DOC_PATH"
      } >"$merged_doc"
    else
      {
        cat "$ANTIGRAVITY_HOME_DOC_PATH"
        if [[ -s "$ANTIGRAVITY_HOME_DOC_PATH" ]]; then
          printf '\n\n'
        fi
        cat "$block_doc"
      } >"$merged_doc"
    fi
  else
    cp "$block_doc" "$merged_doc"
  fi

  validate_markdown_budget "$merged_doc" "Antigravity home rule"
  sync_rendered_file "$ANTIGRAVITY_HOME_DOC_PATH" "$merged_doc" "Synced Antigravity home GEMINI.md"
}

build_workspace_skill_links() {
  local staged="$tmpdir/workspace-skills.tsv"
  : >"$staged"

  while IFS= read -r repo; do
    local repo_name repo_key
    repo_name="$(basename "$repo")"
    repo_key="$(sanitize_kebab_name "$repo_name")"

    while IFS= read -r skill_dir; do
      [[ -f "$skill_dir/SKILL.md" ]] || continue
      local skill_name skill_key link_name
      skill_name="$(basename "$skill_dir")"
      skill_key="$(sanitize_kebab_name "$skill_name")"
      link_name="${repo_key}--${skill_key}"
      if [[ -n "${workspace_skill_name_seen[$link_name]:-}" ]]; then
        hg_die "Duplicate Antigravity workspace skill link: $link_name"
      fi
      workspace_skill_name_seen["$link_name"]=1
      managed_workspace_skill_set["$link_name"]=1
      printf '%s\t%s\n' "$link_name" "$skill_dir" >>"$staged"
    done < <(find "$repo/.agents/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)
  done < <(list_workspace_member_repos)

  while IFS=$'\t' read -r link_name source_dir; do
    [[ -n "$link_name" ]] || continue
    sync_symlink "$WORKSPACE_SKILLS_DIR/$link_name" "$source_dir" "Synced workspace Antigravity skill"
    managed_workspace_skill_updates=$((managed_workspace_skill_updates + 1))
  done <"$staged"

  WORKSPACE_SKILL_COUNT="$(wc -l <"$staged" | tr -d ' ')"
}

build_workspace_workflows() {
  local staging_dir="$tmpdir/workspace-workflows"
  local repo_summary_file="$tmpdir/repo-workflow-summary.tsv"
  local managed_names_file="$tmpdir/managed-workspace-workflows.txt"
  mkdir -p "$staging_dir"
  : >"$repo_summary_file"
  : >"$managed_names_file"

  while IFS= read -r repo; do
    local repo_name repo_key skill_count=0
    local -a repo_skill_lines=()
    repo_name="$(basename "$repo")"
    repo_key="$(sanitize_kebab_name "$repo_name")"

    while IFS= read -r skill_file; do
      [[ -f "$skill_file" ]] || continue
      local skill_dir skill_name skill_desc skill_key workflow_name workflow_path
      skill_dir="$(basename "$(dirname "$skill_file")")"
      skill_name="$(extract_frontmatter_scalar "$skill_file" "name")"
      [[ -n "$skill_name" ]] || skill_name="$skill_dir"
      skill_desc="$(extract_frontmatter_scalar "$skill_file" "description")"
      [[ -n "$skill_desc" ]] || skill_desc="Use this workflow when the task matches the repo-local skill surface."
      skill_key="$(sanitize_kebab_name "$skill_dir")"
      workflow_name="${repo_key}--${skill_key}.md"
      if [[ -n "${workflow_name_seen[$workflow_name]:-}" ]]; then
        hg_die "Duplicate Antigravity workspace workflow name: $workflow_name"
      fi
      workflow_name_seen["$workflow_name"]=1
      managed_workspace_workflow_set["$workflow_name"]=1
      printf '%s\n' "$workflow_name" >>"$managed_names_file"
      workflow_path="$staging_dir/$workflow_name"

      {
        printf '# /%s\n\n' "${workflow_name%.md}"
        printf '%s\n\n' "$skill_desc"
        printf '## Source\n\n'
        printf -- '- Repo root: `%s`\n' "$repo"
        printf -- '- Skill: `%s`\n\n' "$skill_file"
        printf '## Steps\n\n'
        printf '1. Read the repo-local `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` files first when present.\n'
        printf '2. Open the canonical skill source and follow it directly instead of duplicating the skill body here.\n'
        printf '3. Reuse repo-local scripts, references, and MCP surfaces linked from the skill folder.\n'
        printf '4. Keep the task scoped to `%s` unless the skill explicitly routes you into shared docs or control-plane tooling.\n' "$repo_name"
      } >"$workflow_path"
      validate_markdown_budget "$workflow_path" "Workspace skill workflow"

      repo_skill_lines+=("- \`${skill_name}\`: ${skill_desc}")
      skill_count=$((skill_count + 1))
    done < <(find "$repo/.agents/skills" -mindepth 2 -maxdepth 2 -type f -name SKILL.md 2>/dev/null | sort)

    if [[ "$skill_count" -eq 0 ]]; then
      continue
    fi

    local repo_workflow_name repo_workflow_path
    repo_workflow_name="${repo_key}.md"
    if [[ -n "${workflow_name_seen[$repo_workflow_name]:-}" ]]; then
      hg_die "Duplicate Antigravity workspace workflow name: $repo_workflow_name"
    fi
    workflow_name_seen["$repo_workflow_name"]=1
    managed_workspace_workflow_set["$repo_workflow_name"]=1
    printf '%s\n' "$repo_workflow_name" >>"$managed_names_file"
    repo_workflow_path="$staging_dir/$repo_workflow_name"
    {
      printf '# /%s\n\n' "${repo_workflow_name%.md}"
      printf 'Entry point for the `%s` repo-local Antigravity skills.\n\n' "$repo_name"
      printf '## Source\n\n'
      printf -- '- Repo root: `%s`\n' "$repo"
      printf -- '- Skill surface: `%s/.agents/skills/`\n\n' "$repo"
      printf '## Available skills\n\n'
      printf '%s\n' "${repo_skill_lines[@]}"
    } >"$repo_workflow_path"
    validate_markdown_budget "$repo_workflow_path" "Workspace repo workflow"
    printf '%s\t%s\t%s\n' "$repo_name" "$skill_count" "$repo_workflow_name" >>"$repo_summary_file"
  done < <(list_workspace_member_repos)

  local routing_path="$staging_dir/00-workspace-routing.md"
  managed_workspace_workflow_set["00-workspace-routing.md"]=1
  printf '%s\n' "00-workspace-routing.md" >>"$managed_names_file"
  {
    printf '# /workspace-routing\n\n'
    printf 'Route work through the shared `hairglasses-studio` control plane before broad scans.\n\n'
    printf '## Steps\n\n'
    printf '1. Start with `CLAUDE.md`, `AGENTS.md`, and `%s`.\n' "$WORKSPACE_RULE_PATH"
    printf '2. Use `workspace/manifest.json` to decide which repo owns the task.\n'
    printf '3. Check `docs/indexes/SEARCH-GUIDE.md` before doing net-new research.\n'
    printf '4. Prefer a repo-local workflow below when the task clearly maps to an existing skill.\n'
    if [[ -s "$repo_summary_file" ]]; then
      printf '\n## Repo fronts\n\n'
      while IFS=$'\t' read -r repo_name skill_count workflow_name; do
        printf -- '- `%s`: %s skill workflows via `/%s`\n' "$repo_name" "$skill_count" "${workflow_name%.md}"
      done <"$repo_summary_file"
    fi
  } >"$routing_path"
  validate_markdown_budget "$routing_path" "Workspace routing workflow"

  while IFS= read -r workflow_name; do
    [[ -n "$workflow_name" ]] || continue
    sync_rendered_file \
      "$WORKSPACE_WORKFLOWS_DIR/$workflow_name" \
      "$staging_dir/$workflow_name" \
      "Synced workspace Antigravity workflow"
    managed_workspace_workflow_updates=$((managed_workspace_workflow_updates + 1))
  done < <(sort -u "$managed_names_file")

  WORKFLOW_WORKSPACE_COUNT="$(sort -u "$managed_names_file" | wc -l | tr -d ' ')"
  WORKFLOW_SKILL_COUNT="$(awk '/--.*\.md$/ {count++} END {print count + 0}' "$managed_names_file")"
  WORKFLOW_REPO_COUNT="$(awk '!/--/ && $0 !~ /^00-/ {count++} END {print count + 0}' "$managed_names_file")"
}

parse_env_file() {
  local env_file="$1"
  [[ -f "$env_file" ]] || return 0
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ "$line" =~ ^[[:space:]]*# ]] && continue
    [[ "$line" =~ ^[[:space:]]*$ ]] && continue
    if [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=(.*)$ ]]; then
      local key raw value
      key="${BASH_REMATCH[2]}"
      raw="${BASH_REMATCH[3]}"
      raw="${raw#"${raw%%[![:space:]]*}"}"
      raw="${raw%"${raw##*[![:space:]]}"}"
      if [[ "$raw" =~ ^\"(.*)\"$ ]]; then
        value="${BASH_REMATCH[1]}"
      elif [[ "$raw" =~ ^\'(.*)\'$ ]]; then
        value="${BASH_REMATCH[1]}"
      else
        value="${raw%%[[:space:]]#*}"
      fi
      case "$key" in
        OPENAI_API_KEY|ANTHROPIC_API_KEY|GOOGLE_API_KEY|GEMINI_API_KEY)
          if [[ -z "${imported_env_values[$key]:-}" && -n "$value" ]]; then
            imported_env_values["$key"]="$value"
            imported_env_sources["$key"]="$env_file"
          fi
          ;;
      esac
    fi
  done <"$env_file"
}

build_env_bridge_manifest() {
  local source_file="$tmpdir/env-sources.txt"
  : >"$source_file"

  if [[ -f "$WORKSPACE_ROOT/.env" ]]; then
    printf '%s\n' "$WORKSPACE_ROOT/.env" >>"$source_file"
    parse_env_file "$WORKSPACE_ROOT/.env"
  fi

  while IFS= read -r repo; do
    if [[ -f "$repo/.env" ]]; then
      printf '%s\n' "$repo/.env" >>"$source_file"
      parse_env_file "$repo/.env"
    fi
  done < <(list_workspace_member_repos)

  {
    printf '{\n'
    printf '  "generator": "dotfiles/scripts/hg-antigravity-sync.sh",\n'
    printf '  "workspace_root": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_ROOT" '$v')"
    printf '  "mode": "runtime-resolved",\n'
    printf '  "allowed_keys": ["OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_API_KEY", "GEMINI_API_KEY"],\n'
    printf '  "ordered_sources": %s\n' "$(json_string_array_from_lines "$source_file")"
    printf '}\n'
  } >"$tmpdir/env-sources.json"

  local key
  for key in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY GEMINI_API_KEY; do
    [[ -n "${imported_env_values[$key]:-}" ]] || missing_env_vars=$((missing_env_vars + 1))
  done

  sync_rendered_file "$ANTIGRAVITY_ENV_MANIFEST_PATH" "$tmpdir/env-sources.json" "Synced Antigravity env source manifest"
  remove_managed_path "$ANTIGRAVITY_LEGACY_ENV_PATH" "legacy Antigravity env bridge"
}

build_default_workspace_file() {
  local repo rel repo_name

  {
    printf '{\n'
    printf '  "folders": [\n'
    printf '    {\n'
    printf '      "name": %s,\n' "$(jq -Rn --arg v "$(basename "$WORKSPACE_ROOT")" '$v')"
    printf '      "path": "."\n'
    printf '    }'

    while IFS= read -r repo; do
      [[ -n "$repo" ]] || continue
      rel="${repo#$WORKSPACE_ROOT/}"
      repo_name="$(basename "$repo")"
      printf ',\n'
      printf '    {\n'
      printf '      "name": %s,\n' "$(jq -Rn --arg v "$repo_name" '$v')"
      printf '      "path": %s\n' "$(jq -Rn --arg v "./$rel" '$v')"
      printf '    }'
    done < <(list_workspace_member_repos)

    printf '\n  ],\n'
    printf '  "settings": {\n'
    printf '    "terminal.integrated.cwd": %s\n' "$(jq -Rn --arg v "$WORKSPACE_ROOT" '$v')"
    printf '  }\n'
    printf '}\n'
  } >"$tmpdir/hairglasses-studio.code-workspace"

  sync_rendered_file "$ANTIGRAVITY_WORKSPACE_FILE_PATH" "$tmpdir/hairglasses-studio.code-workspace" "Synced Antigravity default workspace"
}

build_launcher_wrapper() {
  {
    printf '#!/usr/bin/env bash\n'
    printf 'set -euo pipefail\n\n'
    printf 'default_workspace=%q\n\n' "$ANTIGRAVITY_WORKSPACE_FILE_PATH"
    printf 'parse_runtime_env_file() {\n'
    printf '  local env_file="$1"\n'
    printf '  [[ -f "$env_file" ]] || return 0\n'
    printf '  while IFS= read -r line || [[ -n "$line" ]]; do\n'
    printf '    [[ "$line" =~ ^[[:space:]]*# ]] && continue\n'
    printf '    [[ "$line" =~ ^[[:space:]]*$ ]] && continue\n'
    printf '    if [[ "$line" =~ ^[[:space:]]*(export[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*=(.*)$ ]]; then\n'
    printf '      local key="${BASH_REMATCH[2]}" raw="${BASH_REMATCH[3]}" value=""\n'
    printf '      raw="${raw#"${raw%%%%[![:space:]]*}"}"\n'
    printf '      raw="${raw%%"${raw##*[![:space:]]}"}"\n'
    printf '      if [[ "$raw" =~ ^\\"(.*)\\"$ ]]; then\n'
    printf '        value="${BASH_REMATCH[1]}"\n'
    printf '      elif [[ "$raw" =~ ^\\'\''(.*)\\'\''$ ]]; then\n'
    printf '        value="${BASH_REMATCH[1]}"\n'
    printf '      else\n'
    printf '        value="${raw%%%%[[:space:]]#*}"\n'
    printf '      fi\n'
    printf '      case "$key" in\n'
    printf '        OPENAI_API_KEY|ANTHROPIC_API_KEY|GOOGLE_API_KEY|GEMINI_API_KEY)\n'
    printf '          if [[ -z "${!key:-}" && -n "$value" ]]; then\n'
    printf '            export "$key=$value"\n'
    printf '          fi\n'
    printf '          ;;\n'
    printf '      esac\n'
    printf '    fi\n'
    printf '  done <"$env_file"\n'
    printf '}\n\n'
    printf 'if [[ -f %q ]]; then\n' "$ANTIGRAVITY_ENV_MANIFEST_PATH"
    printf '  while IFS= read -r env_file; do\n'
    printf '    [[ -n "$env_file" ]] || continue\n'
    printf '    parse_runtime_env_file "$env_file"\n'
    printf "  done < <(jq -r '.ordered_sources[]?' %q)\n" "$ANTIGRAVITY_ENV_MANIFEST_PATH"
    printf 'fi\n\n'
    printf 'if [[ "$#" -eq 0 && -f "$default_workspace" ]]; then\n'
    printf '  set -- "$default_workspace"\n'
    printf 'fi\n\n'
    printf 'exec /usr/bin/antigravity "$@"\n'
  } >"$tmpdir/antigravity-managed"
  chmod 755 "$tmpdir/antigravity-managed"
  sync_rendered_file "$ANTIGRAVITY_WRAPPER_PATH" "$tmpdir/antigravity-managed" "Synced Antigravity launcher wrapper"
}

build_desktop_entry() {
  {
    printf '[Desktop Entry]\n'
    printf 'Name=Antigravity\n'
    printf 'Comment=Experience liftoff\n'
    printf 'GenericName=Text Editor\n'
    printf 'Exec=%s %%F\n' "$ANTIGRAVITY_WRAPPER_PATH"
    printf 'Icon=antigravity\n'
    printf 'Type=Application\n'
    printf 'StartupNotify=false\n'
    printf 'StartupWMClass=Antigravity\n'
    printf 'Categories=TextEditor;Development;IDE;\n'
    printf 'MimeType=application/x-antigravity-workspace;\n'
    printf 'Actions=new-empty-window;\n'
    printf 'Keywords=vscode;\n\n'
    printf '[Desktop Action new-empty-window]\n'
    printf 'Name=New Empty Window\n'
    printf 'Exec=%s --new-window %%F\n' "$ANTIGRAVITY_WRAPPER_PATH"
    printf 'Icon=antigravity\n'
  } >"$tmpdir/antigravity.desktop"
  sync_rendered_file "$ANTIGRAVITY_DESKTOP_PATH" "$tmpdir/antigravity.desktop" "Synced Antigravity desktop entry"
}

build_settings_json() {
  local file_excludes search_excludes watcher_excludes
  file_excludes='{
    "**/.*": true
  }'
  search_excludes='{
    "workspace/**": true,
    "**/.ralph/**": true,
    "**/.tools/**": true,
    "**/node_modules/**": true,
    "**/.venv/**": true,
    "**/.cache/**": true,
    "**/dist/**": true,
    "**/build/**": true,
    "**/.agentMemory/**": true,
    "jellyfin-stack/vendor/**": true
  }'
  watcher_excludes="$search_excludes"

  if [[ -f "$ANTIGRAVITY_SETTINGS_PATH" ]]; then
    cp "$ANTIGRAVITY_SETTINGS_PATH" "$tmpdir/antigravity-settings-base.json"
  else
    printf '{}\n' >"$tmpdir/antigravity-settings-base.json"
  fi

  jq -S \
    --argjson file_excludes "$file_excludes" \
    --argjson search_excludes "$search_excludes" \
    --argjson watcher_excludes "$watcher_excludes" '
      . + {
        "security.workspace.trust.untrustedFiles": "open",
        "antigravity.searchMaxWorkspaceFileCount": 50000,
        "antigravity.persistentLanguageServer": true,
        "antigravity.enableCursorImportCursor": false,
        "tfa.status.showQuota": true,
        "tfa.status.showCache": true,
        "tfa.status.warningThreshold": 30,
        "tfa.status.criticalThreshold": 10,
        "tfa.system.debugMode": false,
        "tfa.system.autoAccept": false,
        "tfa.cache.autoClean": false,
        "files.exclude": ((.["files.exclude"] // {}) + $file_excludes),
        "files.watcherExclude": ((.["files.watcherExclude"] // {}) + $watcher_excludes),
        "search.exclude": ((.["search.exclude"] // {}) + $search_excludes)
      }
    ' "$tmpdir/antigravity-settings-base.json" >"$tmpdir/antigravity-settings.json"

  sync_rendered_file "$ANTIGRAVITY_SETTINGS_PATH" "$tmpdir/antigravity-settings.json" "Synced Antigravity user settings"
}

sync_permission_preset() {
  [[ -f "$PERMISSION_PRESET_SOURCE" ]] || return 0
  cp "$PERMISSION_PRESET_SOURCE" "$tmpdir/permission-presets.md"
  sync_rendered_file "$PERMISSION_PRESET_TARGET" "$tmpdir/permission-presets.md" "Synced Antigravity permission presets"
}

remove_stale_workspace_artifacts() {
  if [[ -f "$ANTIGRAVITY_METADATA_PATH" ]]; then
    while IFS= read -r old_name; do
      [[ -n "$old_name" ]] || continue
      [[ -n "${managed_workspace_skill_set[$old_name]:-}" ]] && continue
      remove_managed_path "$WORKSPACE_SKILLS_DIR/$old_name" "workspace Antigravity skill"
    done < <(
      jq -r '.managed_workspace_skills[]? // empty' "$ANTIGRAVITY_METADATA_PATH" 2>/dev/null || true
    )

    while IFS= read -r old_name; do
      [[ -n "$old_name" ]] || continue
      [[ -n "${managed_workspace_workflow_set[$old_name]:-}" ]] && continue
      remove_managed_path "$WORKSPACE_WORKFLOWS_DIR/$old_name" "workspace Antigravity workflow"
    done < <(
      jq -r '.managed_workspace_workflows[]? // empty' "$ANTIGRAVITY_METADATA_PATH" 2>/dev/null || true
    )

    while IFS= read -r legacy_name; do
      [[ -n "$legacy_name" ]] || continue
      [[ -n "${managed_workspace_workflow_set[$legacy_name]:-}" ]] && continue
      remove_managed_path "$ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR/$legacy_name" "legacy Antigravity global workflow"
    done < <(
      jq -r '.managed_workflows[]? // empty' "$ANTIGRAVITY_METADATA_PATH" 2>/dev/null || true
    )
  fi
}

remove_legacy_global_workspace_workflows() {
  [[ -d "$ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR" ]] || return 0
  [[ -d "$WORKSPACE_WORKFLOWS_DIR" ]] || return 0

  local managed_global_workflow
  declare -A managed_global_workflow_set=()

  if [[ -f "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH" ]]; then
    while IFS= read -r managed_global_workflow; do
      [[ -n "$managed_global_workflow" ]] || continue
      managed_global_workflow_set["$managed_global_workflow"]=1
    done < <(jq -r '.managed_global_workflows[]? // empty' "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH" 2>/dev/null || true)
  fi

  local legacy_name
  while IFS= read -r legacy_name; do
    [[ -n "$legacy_name" ]] || continue
    [[ -n "${managed_global_workflow_set[$legacy_name]:-}" ]] && continue
    [[ -f "$WORKSPACE_WORKFLOWS_DIR/$legacy_name" ]] || continue
    remove_managed_path "$ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR/$legacy_name" "legacy Antigravity global workflow"
  done < <(
    find "$ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR" -mindepth 1 -maxdepth 1 -type f -name '*.md' -printf '%f\n' 2>/dev/null | sort
  )
}

write_metadata() {
  local imported_vars_file="$tmpdir/imported-vars.txt"
  local missing_vars_file="$tmpdir/missing-vars.txt"
  local workspace_skill_file="$tmpdir/workspace-skills.txt"
  local workspace_workflow_file="$tmpdir/workspace-workflows.txt"
  : >"$imported_vars_file"
  : >"$missing_vars_file"
  printf '%s\n' "${!managed_workspace_skill_set[@]}" | sort -u >"$workspace_skill_file"
  printf '%s\n' "${!managed_workspace_workflow_set[@]}" | sort -u >"$workspace_workflow_file"

  local key
  for key in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY GEMINI_API_KEY; do
    if [[ -n "${imported_env_values[$key]:-}" ]]; then
      printf '%s\n' "$key" >>"$imported_vars_file"
    else
      printf '%s\n' "$key" >>"$missing_vars_file"
    fi
  done

  local ecosystem_global_workflows=0
  local ecosystem_global_skills=0
  if [[ -f "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH" ]]; then
    ecosystem_global_workflows="$(jq -r '.global_workflow_count // 0' "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH")"
    ecosystem_global_skills="$(jq -r '.live_global_skill_count // 0' "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH")"
  fi
  WORKFLOW_GLOBAL_COUNT="$ecosystem_global_workflows"
  GLOBAL_SKILL_COUNT="$ecosystem_global_skills"

  {
    printf '{\n'
    printf '  "generator": "dotfiles/scripts/hg-antigravity-sync.sh",\n'
    printf '  "workspace_root": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_ROOT" '$v')"
    printf '  "generated_on": %s,\n' "$(jq -Rn --arg v "$(date +%Y-%m-%d)" '$v')"
    printf '  "root_mcp_path": %s,\n' "$(jq -Rn --arg v "$ROOT_MCP_PATH" '$v')"
    printf '  "mcp_config_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_MCP_PATH" '$v')"
    printf '  "settings_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_SETTINGS_PATH" '$v')"
    printf '  "env_source_manifest_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_ENV_MANIFEST_PATH" '$v')"
    printf '  "wrapper_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_WRAPPER_PATH" '$v')"
    printf '  "desktop_entry_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_DESKTOP_PATH" '$v')"
    printf '  "workspace_file_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_WORKSPACE_FILE_PATH" '$v')"
    printf '  "workspace_rule_path": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_RULE_PATH" '$v')"
    printf '  "global_gemini_md_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_HOME_DOC_PATH" '$v')"
    printf '  "workspace_skills_dir": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_SKILLS_DIR" '$v')"
    printf '  "workspace_workflows_dir": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_WORKFLOWS_DIR" '$v')"
    printf '  "permission_presets_path": %s,\n' "$(jq -Rn --arg v "$PERMISSION_PRESET_TARGET" '$v')"
    printf '  "ecosystem_metadata_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_ECOSYSTEM_METADATA_PATH" '$v')"
    printf '  "env_bridge_mode": "runtime-resolved",\n'
    printf '  "total_mcp_server_count": %s,\n' "$MCP_TOTAL_SERVER_COUNT"
    printf '  "root_shared_server_count": %s,\n' "$MCP_ROOT_SHARED_SERVER_COUNT"
    printf '  "repo_workflow_count": %s,\n' "$WORKFLOW_REPO_COUNT"
    printf '  "skill_workflow_count": %s,\n' "$WORKFLOW_SKILL_COUNT"
    printf '  "workspace_workflow_count": %s,\n' "$WORKFLOW_WORKSPACE_COUNT"
    printf '  "global_workflow_count": %s,\n' "$WORKFLOW_GLOBAL_COUNT"
    printf '  "managed_workflow_count": %s,\n' "$((WORKFLOW_WORKSPACE_COUNT + WORKFLOW_GLOBAL_COUNT))"
    printf '  "workspace_skill_count": %s,\n' "$WORKSPACE_SKILL_COUNT"
    printf '  "global_skill_count": %s,\n' "$GLOBAL_SKILL_COUNT"
    printf '  "managed_workspace_skills": %s,\n' "$(json_string_array_from_lines "$workspace_skill_file")"
    printf '  "managed_workspace_workflows": %s,\n' "$(json_string_array_from_lines "$workspace_workflow_file")"
    printf '  "imported_env_vars": %s,\n' "$(json_string_array_from_lines "$imported_vars_file")"
    printf '  "missing_env_vars": %s,\n' "$(json_string_array_from_lines "$missing_vars_file")"
    printf '  "mcp_name_collision_count": 0,\n'
    printf '  "workflow_name_collision_count": 0,\n'
    printf '  "global_gemini_md_present": %s,\n' "$( [[ -f "$ANTIGRAVITY_HOME_DOC_PATH" ]] && printf 'true' || printf 'false' )"
    printf '  "imported_env_sources": {\n'
    local first=1
    for key in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY GEMINI_API_KEY; do
      [[ -n "${imported_env_sources[$key]:-}" ]] || continue
      [[ "$first" -eq 1 ]] || printf ',\n'
      first=0
      printf '    %s: %s' \
        "$(jq -Rn --arg v "$key" '$v')" \
        "$(jq -Rn --arg v "${imported_env_sources[$key]}" '$v')"
    done
    printf '\n  }\n'
    printf '}\n'
  } >"$tmpdir/antigravity-metadata.json"

  sync_rendered_file "$ANTIGRAVITY_METADATA_PATH" "$tmpdir/antigravity-metadata.json" "Synced Antigravity metadata"
}

build_antigravity_mcp_config
render_workspace_rule_file
render_home_gemini_doc
build_workspace_skill_links
build_workspace_workflows
build_env_bridge_manifest
build_default_workspace_file
build_launcher_wrapper
build_desktop_entry
build_settings_json
sync_permission_preset
remove_stale_workspace_artifacts

if [[ -x "$ECOSYSTEM_SYNC_SCRIPT" ]]; then
  "$ECOSYSTEM_SYNC_SCRIPT" "${MODE_ARGS[@]}"
fi

remove_legacy_global_workspace_workflows

write_metadata

case "$MODE" in
  write)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "Antigravity sync already up to date"
    else
      hg_info "Antigravity sync refreshed"
    fi
    ;;
  dry-run)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "No Antigravity changes needed"
    fi
    ;;
  check)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "Antigravity sync up to date"
    else
      exit 1
    fi
    ;;
esac
