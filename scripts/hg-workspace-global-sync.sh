#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"
RUN_SKILLS=true
RUN_TOOLS=true
CHECK_GLOBAL_TARGETS=true

WORKSPACE_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
MANIFEST_PATH="${HG_WORKSPACE_MANIFEST:-}"
ROOT_MCP_PATH="${HG_ROOT_MCP_PATH:-}"
CLAUDE_JSON_PATH="${HG_CLAUDE_JSON_PATH:-$HOME/.claude.json}"
CLAUDE_PROJECT_KEY="${HG_CLAUDE_PROJECT_KEY:-}"
CLAUDE_HOME_DOC_PATH="${HG_CLAUDE_HOME_DOC_PATH:-$HOME/.claude/CLAUDE.md}"
CLAUDE_SKILLS_DIR="${HG_CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}"
CLAUDE_COMMANDS_DIR="${HG_CLAUDE_COMMANDS_DIR:-$HOME/.claude/commands}"
AGENTS_SKILLS_DIR="${HG_AGENTS_SKILLS_DIR:-$HOME/.agents/skills}"
CODEX_SKILLS_DIR="${HG_CODEX_SKILLS_DIR:-$HOME/.codex/skills}"
CODEX_CONFIG_PATH="${HG_CODEX_CONFIG_PATH:-$HOME/.codex/config.toml}"
GEMINI_HOME_DOC_PATH="${HG_GEMINI_HOME_DOC_PATH:-$HOME/.gemini/GEMINI.md}"
GEMINI_PROJECTS_PATH="${HG_GEMINI_PROJECTS_PATH:-$HOME/.gemini/projects.json}"
GEMINI_SETTINGS_PATH="${HG_GEMINI_SETTINGS_PATH:-$HOME/.gemini/settings.json}"
ANTIGRAVITY_SYNC_SCRIPT="${HG_ANTIGRAVITY_SYNC_SCRIPT:-$SCRIPT_DIR/hg-antigravity-sync.sh}"

START_MARKER="# BEGIN GENERATED MCP SERVERS: hg-workspace-global-sync"
END_MARKER="# END GENERATED MCP SERVERS: hg-workspace-global-sync"
LEGACY_START_MARKER="# BEGIN GENERATED MCP SERVERS: hg-global-mcp-sync"
LEGACY_END_MARKER="# END GENERATED MCP SERVERS: hg-global-mcp-sync"
HOME_CONTEXT_START_MARKER="<!-- BEGIN GENERATED WORKSPACE GLOBAL: hg-workspace-global-sync -->"
HOME_CONTEXT_END_MARKER="<!-- END GENERATED WORKSPACE GLOBAL: hg-workspace-global-sync -->"
SKILL_OWNER_FILE=".hg-workspace-global-sync.json"
CLAUDE_PREFIX="studio_"

workspace_global_exported_skill_name() {
  local repo_name="$1"
  local skill_name="$2"
  printf '%s-%s' "$repo_name" "$skill_name"
}

usage() {
  cat <<'EOF'
Usage: hg-workspace-global-sync.sh [options]

Refresh managed workspace-global skills and curated MCP overlays from
workspace/manifest.json and repo-local metadata.

Options:
  --root <path>               Workspace root (default: ~/hairglasses-studio)
  --manifest <path>           Manifest path (default: <root>/workspace/manifest.json)
  --root-mcp <path>           Shared root .mcp.json (default: <root>/.mcp.json)
  --claude-json <path>        Claude settings JSON (default: ~/.claude.json)
  --claude-home-doc <path>    Claude home CLAUDE.md (default: ~/.claude/CLAUDE.md)
  --claude-project-key <key>  Claude project key (default: <root>)
  --claude-skills-dir <path>  Global Claude skills dir (default: ~/.claude/skills)
  --claude-commands-dir <path>
                              Global Claude commands dir (default: ~/.claude/commands)
  --agents-skills-dir <path>  Global Agents skills dir (default: ~/.agents/skills)
  --codex-skills-dir <path>   Global Codex skills dir (default: ~/.codex/skills)
  --codex-config <path>       Codex config TOML (default: ~/.codex/config.toml)
  --gemini-home-doc <path>    Gemini home GEMINI.md (default: ~/.gemini/GEMINI.md)
  --gemini-projects <path>    Gemini projects.json (default: ~/.gemini/projects.json)
  --gemini-settings <path>    Gemini settings JSON (default: ~/.gemini/settings.json)
  --skills-only               Sync managed global skills only
  --tools-only                Sync managed Claude/Codex MCP overlays only
  --dry-run                   Report changes without writing
  --check                     Exit non-zero if managed global state is stale
  --source-check              Validate repo-controlled manifests, skills, and
                              MCP policy sources without comparing user-home
                              overlays under ~/.claude, ~/.codex, or ~/.gemini
  -h, --help                  Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --root)
      [[ $# -ge 2 ]] || hg_die "--root requires a path"
      WORKSPACE_ROOT="$2"
      shift 2
      ;;
    --manifest)
      [[ $# -ge 2 ]] || hg_die "--manifest requires a path"
      MANIFEST_PATH="$2"
      shift 2
      ;;
    --root-mcp)
      [[ $# -ge 2 ]] || hg_die "--root-mcp requires a path"
      ROOT_MCP_PATH="$2"
      shift 2
      ;;
    --claude-json)
      [[ $# -ge 2 ]] || hg_die "--claude-json requires a path"
      CLAUDE_JSON_PATH="$2"
      shift 2
      ;;
    --claude-home-doc)
      [[ $# -ge 2 ]] || hg_die "--claude-home-doc requires a path"
      CLAUDE_HOME_DOC_PATH="$2"
      shift 2
      ;;
    --claude-project-key)
      [[ $# -ge 2 ]] || hg_die "--claude-project-key requires a value"
      CLAUDE_PROJECT_KEY="$2"
      shift 2
      ;;
    --claude-skills-dir)
      [[ $# -ge 2 ]] || hg_die "--claude-skills-dir requires a path"
      CLAUDE_SKILLS_DIR="$2"
      shift 2
      ;;
    --claude-commands-dir)
      [[ $# -ge 2 ]] || hg_die "--claude-commands-dir requires a path"
      CLAUDE_COMMANDS_DIR="$2"
      shift 2
      ;;
    --agents-skills-dir)
      [[ $# -ge 2 ]] || hg_die "--agents-skills-dir requires a path"
      AGENTS_SKILLS_DIR="$2"
      shift 2
      ;;
    --codex-skills-dir)
      [[ $# -ge 2 ]] || hg_die "--codex-skills-dir requires a path"
      CODEX_SKILLS_DIR="$2"
      shift 2
      ;;
    --codex-config)
      [[ $# -ge 2 ]] || hg_die "--codex-config requires a path"
      CODEX_CONFIG_PATH="$2"
      shift 2
      ;;
    --gemini-home-doc)
      [[ $# -ge 2 ]] || hg_die "--gemini-home-doc requires a path"
      GEMINI_HOME_DOC_PATH="$2"
      shift 2
      ;;
    --gemini-projects)
      [[ $# -ge 2 ]] || hg_die "--gemini-projects requires a path"
      GEMINI_PROJECTS_PATH="$2"
      shift 2
      ;;
    --gemini-settings)
      [[ $# -ge 2 ]] || hg_die "--gemini-settings requires a path"
      GEMINI_SETTINGS_PATH="$2"
      shift 2
      ;;
    --skills-only)
      RUN_SKILLS=true
      RUN_TOOLS=false
      shift
      ;;
    --tools-only)
      RUN_SKILLS=false
      RUN_TOOLS=true
      shift
      ;;
    --dry-run)
      MODE="dry-run"
      shift
      ;;
    --check)
      MODE="check"
      shift
      ;;
    --source-check)
      MODE="check"
      CHECK_GLOBAL_TARGETS=false
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

hg_require jq mktemp awk sort diff cmp

WORKSPACE_ROOT="$(cd "$WORKSPACE_ROOT" && pwd)"
MANIFEST_PATH="${MANIFEST_PATH:-$WORKSPACE_ROOT/workspace/manifest.json}"
ROOT_MCP_PATH="${ROOT_MCP_PATH:-$WORKSPACE_ROOT/.mcp.json}"
CLAUDE_PROJECT_KEY="${CLAUDE_PROJECT_KEY:-$WORKSPACE_ROOT}"

[[ -d "$WORKSPACE_ROOT" ]] || hg_die "Workspace root not found: $WORKSPACE_ROOT"
[[ -f "$MANIFEST_PATH" ]] || hg_die "Workspace manifest not found: $MANIFEST_PATH"

if $RUN_SKILLS && $CHECK_GLOBAL_TARGETS; then
  [[ -d "$CLAUDE_COMMANDS_DIR" ]] || hg_die "Claude commands directory not found: $CLAUDE_COMMANDS_DIR"
fi

if $RUN_TOOLS; then
  [[ -f "$ROOT_MCP_PATH" ]] || hg_die "Shared root MCP file not found: $ROOT_MCP_PATH"
  jq -e '.mcpServers and (.mcpServers | type == "object")' "$ROOT_MCP_PATH" >/dev/null || hg_die "Invalid root .mcp.json: expected top-level mcpServers object"
fi

MODE_ARGS=()
case "$MODE" in
  dry-run) MODE_ARGS=(--dry-run) ;;
  check) MODE_ARGS=(--check) ;;
esac

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
: >"$tmpdir/gemini-skill-catalog.tsv"

overall_pending=0
skill_pending=0
tool_pending=0
skill_conflicts=0
global_claude_manual_block_count=0
global_claude_stale_count=0
global_claude_unexpected_count=0
stale_workspace_claude_overlay_count=0
stale_workspace_codex_overlay_count=0
stale_workspace_gemini_overlay_count=0
stale_claude_home_doc_count=0
stale_gemini_home_doc_count=0
stale_gemini_projects_count=0
stale_antigravity_overlay_count=0

managed_repo_count=0
repos_with_skill_surfaces=0
global_canonical_count=0
global_alias_count=0
global_alias_conflict_names=0
claude_foundational_count=0
claude_raw_count=0
claude_tool_count=0
codex_profile_count=0
codex_tool_count=0
gemini_tool_count=0

declare -A desired_claude_skill_dirs=()
declare -A managed_canonical_owners=()
declare -A alias_counts=()
declare -A alias_conflicts_reported=()
declare -A claude_key_seen=()
declare -A codex_name_seen=()
declare -A gemini_name_seen=()
declare -A raw_source_seen=()

resolve_repo_path() {
  local raw_path="$1"
  if [[ "$raw_path" == /* ]]; then
    printf '%s\n' "$raw_path"
  else
    printf '%s/%s\n' "$WORKSPACE_ROOT" "$raw_path"
  fi
}

repo_has_dirty_tracked_changes() {
  local repo_path="$1"
  git -C "$repo_path" status --short --untracked-files=no 2>/dev/null | grep -q .
}

sanitize_name() {
  local raw="$1"
  printf '%s' "$raw" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/_/g; s/^_+//; s/_+$//; s/_{2,}/_/g'
}

sanitize_kebab_name() {
  local raw="$1"
  printf '%s' "$raw" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g'
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

manifest_supports_jq() {
  local file="$1"
  jq -e '.version == 1 and (.skills | type == "array")' "$file" >/dev/null 2>&1
}

manifest_skill_names() {
  local file="$1"
  if manifest_supports_jq "$file"; then
    jq -r '.skills[].name' "$file"
  else
    sed -n 's/^[[:space:]]*- name:[[:space:]]*//p' "$file"
  fi
}

manifest_alias_pairs() {
  local file="$1"
  if manifest_supports_jq "$file"; then
    jq -r '.skills[] as $skill | $skill.claude_aliases[]? | [if type == "object" then .name else . end, $skill.name] | @tsv' "$file"
  fi
}

skill_dir_is_owned() {
  local dir="$1"
  [[ -f "$dir/$SKILL_OWNER_FILE" ]]
}

skill_dir_is_replaceable() {
  local dir="$1"
  skill_dir_is_owned "$dir" && return 0
  [[ -f "$dir/SKILL.md" ]] || return 1
  grep -Eq 'GENERATED BY hg-skill-surface-sync\.sh|Generated by dotfiles/scripts/hg-workspace-global-sync\.sh' "$dir/SKILL.md" 2>/dev/null
}

sync_skill_dir() {
  local staged_dir="$1"
  local target_dir="$2"

  if ! $CHECK_GLOBAL_TARGETS; then
    return 0
  fi

  if [[ -d "$target_dir" ]] && diff -qr "$staged_dir" "$target_dir" >/dev/null 2>&1; then
    return 0
  fi

  if [[ -e "$target_dir" ]] && ! skill_dir_is_replaceable "$target_dir"; then
    overall_pending=1
    skill_pending=1
    skill_conflicts=$((skill_conflicts + 1))
    global_claude_manual_block_count=$((global_claude_manual_block_count + 1))
    case "$MODE" in
      write)
        hg_warn "Skipping manual global Claude skill directory: $target_dir"
        ;;
      dry-run)
        hg_warn "Would skip manual global Claude skill directory: $target_dir"
        ;;
      check)
        hg_warn "Manual global Claude skill blocks managed sync: $target_dir"
        ;;
    esac
    return 1
  fi

  overall_pending=1
  skill_pending=1
  global_claude_stale_count=$((global_claude_stale_count + 1))
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target_dir")"
      rm -rf "$target_dir"
      cp -R "$staged_dir" "$target_dir"
      hg_ok "Synced workspace global Claude skill: $target_dir"
      ;;
    dry-run)
      hg_warn "Would update workspace global Claude skill: $target_dir"
      ;;
    check)
      hg_warn "Out of date workspace global Claude skill: $target_dir"
      ;;
  esac
}

purge_stale_owned_skill_dirs() {
  if ! $CHECK_GLOBAL_TARGETS; then
    return 0
  fi

  [[ -d "$CLAUDE_SKILLS_DIR" ]] || return 0

  while IFS= read -r dir; do
    local skill_name
    skill_name="$(basename "$dir")"
    [[ -n "${desired_claude_skill_dirs[$skill_name]:-}" ]] && continue
    skill_dir_is_owned "$dir" || continue

    overall_pending=1
    skill_pending=1
    global_claude_unexpected_count=$((global_claude_unexpected_count + 1))
    case "$MODE" in
      write)
        rm -rf "$dir"
        hg_ok "Removed stale workspace global Claude skill: $dir"
        ;;
      dry-run)
        hg_warn "Would remove stale workspace global Claude skill: $dir"
        ;;
      check)
        hg_warn "Unexpected workspace global Claude skill: $dir"
        ;;
    esac
  done < <(find "$CLAUDE_SKILLS_DIR" -mindepth 1 -maxdepth 1 -type d -print 2>/dev/null | sort)
}

sync_rendered_file() {
  local current_path="$1"
  local rendered_path="$2"
  local label="$3"

  if ! $CHECK_GLOBAL_TARGETS; then
    rm -f "$rendered_path"
    return 0
  fi

  if [[ -f "$current_path" ]] && cmp -s "$current_path" "$rendered_path"; then
    rm -f "$rendered_path"
    return 0
  fi

  overall_pending=1
  tool_pending=1
  case "$label" in
    "Synced workspace Claude MCP overlay")
      stale_workspace_claude_overlay_count=$((stale_workspace_claude_overlay_count + 1))
      ;;
    "Synced workspace Codex MCP block")
      stale_workspace_codex_overlay_count=$((stale_workspace_codex_overlay_count + 1))
      ;;
    "Synced workspace Gemini MCP overlay")
      stale_workspace_gemini_overlay_count=$((stale_workspace_gemini_overlay_count + 1))
      ;;
    "Synced Claude home memory")
      stale_claude_home_doc_count=$((stale_claude_home_doc_count + 1))
      ;;
    "Synced Gemini home memory")
      stale_gemini_home_doc_count=$((stale_gemini_home_doc_count + 1))
      ;;
    "Synced Gemini project registry")
      stale_gemini_projects_count=$((stale_gemini_projects_count + 1))
      ;;
  esac
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$current_path")"
      mv "$rendered_path" "$current_path"
      hg_ok "$label: $current_path"
      ;;
    dry-run)
      rm -f "$rendered_path"
      hg_warn "Would update $label: $current_path"
      ;;
    check)
      rm -f "$rendered_path"
      hg_warn "Out of date $label: $current_path"
      ;;
  esac
}

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

expand_home_tokens() {
  local value="$1"
  value="${value//\$\{HOME\}/$HOME}"
  value="${value//\$HOME/$HOME}"
  printf '%s\n' "$value"
}

normalize_codex_env_json() {
  local env_json="$1"
  jq -c --arg home "$HOME" '
    with_entries(
      .value |= if type == "string"
        then gsub("\\$\\{HOME\\}"; $home) | gsub("\\$HOME"; $home)
        else .
        end
    )
  ' <<<"$env_json"
}

normalize_codex_cwd() {
  local base_dir="$1"
  local cwd="$2"

  cwd="$(expand_home_tokens "$cwd")"
  if [[ -z "$cwd" ]]; then
    printf '\n'
    return 0
  fi

  if [[ "$cwd" == "." ]]; then
    [[ -n "$base_dir" ]] || hg_die "Codex MCP cwd '.' requires a repo root"
    cwd="$base_dir"
  elif [[ "$cwd" != /* ]]; then
    [[ -n "$base_dir" ]] || hg_die "Codex MCP relative cwd '$cwd' requires a repo root"
    cwd="$base_dir/$cwd"
  fi

  [[ -d "$cwd" ]] || hg_die "Codex MCP cwd does not exist: $cwd"

  (
    cd "$cwd" >/dev/null 2>&1 || exit 1
    pwd
  ) || hg_die "Failed to canonicalize Codex MCP cwd: $cwd"
}

profile_exports_global_codex() {
  local profile_json="$1"
  jq -e '(.global_codex // false) == true' <<<"$profile_json" >/dev/null
}

profile_exports_global_claude() {
  local profile_json="$1"
  jq -e '(.global_claude // false) == true' <<<"$profile_json" >/dev/null
}

profile_exports_global_gemini() {
  local profile_json="$1"
  jq -e '(.global_gemini // false) == true' <<<"$profile_json" >/dev/null
}

profile_exports_raw_source() {
  local profile_json="$1"
  jq -e '(.mode == "review" or .mode == "research") or (.global_raw_source // false) == true' <<<"$profile_json" >/dev/null
}

profile_global_name() {
  local profile_json="$1"
  jq -r '.global_name // .name' <<<"$profile_json"
}

emit_env_block() {
  local server_name="$1"
  local env_json="$2"
  local env_keys=()
  mapfile -t env_keys < <(jq -r 'keys[]' <<<"$env_json")
  [[ "${#env_keys[@]}" -gt 0 ]] || return 0

  printf '\n[mcp_servers.%s.env]\n' "$server_name"
  local key
  for key in "${env_keys[@]}"; do
    local value
    value="$(jq -r --arg key "$key" '.[$key]' <<<"$env_json")"
    value="$(expand_home_tokens "$value")"
    emit_scalar_line "$key" "$value"
  done
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

append_claude_entry() {
  local stem="$1"
  local value_json="$2"
  local generated_key="${CLAUDE_PREFIX}$(sanitize_name "$stem")"

  [[ -n "$generated_key" ]] || hg_die "Failed to generate Claude MCP key for $stem"
  if [[ -n "${claude_key_seen[$generated_key]:-}" ]]; then
    hg_die "Duplicate generated Claude MCP key: $generated_key"
  fi
  claude_key_seen["$generated_key"]=1

  jq -cn \
    --arg key "$generated_key" \
    --argjson value "$value_json" \
    '{key: $key, value: $value}' >>"$tmpdir/claude-mcp.ndjson"

  claude_tool_count=$((claude_tool_count + 1))
}

append_codex_entry() {
  local stem="$1"
  local comment="$2"
  local command="$3"
  local args_json="$4"
  local cwd="$5"
  local env_json="$6"
  local enabled_tools_json="$7"
  local generated_name="${CLAUDE_PREFIX}$(sanitize_name "$stem")"

  [[ -n "$generated_name" ]] || hg_die "Failed to generate Codex MCP name for $stem"
  if [[ -n "${codex_name_seen[$generated_name]:-}" ]]; then
    hg_die "Duplicate generated Codex MCP name: $generated_name"
  fi
  codex_name_seen["$generated_name"]=1

  jq -cn \
    --arg name "$generated_name" \
    --arg comment "$comment" \
    --arg command "$command" \
    --arg cwd "$cwd" \
    --argjson args "$args_json" \
    --argjson env "$env_json" \
    --argjson enabled_tools "$enabled_tools_json" \
    '{
      name: $name,
      comment: $comment,
      command: $command,
      args: $args,
      cwd: $cwd,
      env: $env,
      enabled_tools: $enabled_tools
    }' >>"$tmpdir/codex-mcp.ndjson"

  codex_tool_count=$((codex_tool_count + 1))
}

append_gemini_entry() {
  local stem="$1"
  local value_json="$2"
  local base_dir="${3:-}"
  local generated_name="studio-$(sanitize_kebab_name "$stem")"
  local normalized_cwd normalized_env_json

  [[ -n "$generated_name" ]] || hg_die "Failed to generate Gemini MCP name for $stem"
  if [[ -n "${gemini_name_seen[$generated_name]:-}" ]]; then
    hg_die "Duplicate generated Gemini MCP name: $generated_name"
  fi
  gemini_name_seen["$generated_name"]=1

  normalized_cwd="$(normalize_codex_cwd "$base_dir" "$(jq -r '.cwd // ""' <<<"$value_json")")"
  normalized_env_json="$(normalize_codex_env_json "$(jq -c '.env // {}' <<<"$value_json")")"

  jq -cn \
    --arg name "$generated_name" \
    --argjson value "$value_json" \
    --arg cwd "$normalized_cwd" \
    --argjson env "$normalized_env_json" \
    '{
      name: $name,
      value: (
        (if ($value | has("command")) then {command: $value.command} else {} end)
        + (if ($value | has("args")) then {args: ($value.args // [])} else {} end)
        + (if ($cwd | length) > 0 then {cwd: $cwd} else {} end)
        + (if ($env | length) > 0 then {env: $env} else {} end)
        + (if ($value.url // null) != null then {url: $value.url} else {} end)
        + (if ($value.httpUrl // null) != null then {httpUrl: $value.httpUrl} else {} end)
        + (if ($value.headers // null) != null then {headers: $value.headers} else {} end)
        + (if ($value.timeout // null) != null then {timeout: $value.timeout} else {} end)
      )
    }' >>"$tmpdir/gemini-mcp.ndjson"

  gemini_tool_count=$((gemini_tool_count + 1))
}

sync_managed_workspace_skills() {
  local skill_entries_file="$tmpdir/skill-entries.tsv"
  local alias_entries_file="$tmpdir/alias-entries.tsv"
  : >"$skill_entries_file"
  : >"$alias_entries_file"

  while IFS=$'\t' read -r repo_name repo_rel_path; do
    [[ -n "$repo_name" ]] || continue
    managed_repo_count=$((managed_repo_count + 1))

    local repo_path surface_path
    repo_path="$(resolve_repo_path "$repo_rel_path")"
    surface_path="$repo_path/.agents/skills/surface.yaml"

    if [[ ! -d "$repo_path" ]]; then
      hg_warn "Skipping managed repo missing from workspace: $repo_path"
      continue
    fi
    [[ -f "$surface_path" ]] || continue

    local json_surface=0
    local skill_names=()
    if manifest_supports_jq "$surface_path"; then
      json_surface=1
    fi

    mapfile -t skill_names < <(manifest_skill_names "$surface_path")
    if [[ "${#skill_names[@]}" -eq 0 ]]; then
      hg_warn "Skipping invalid managed skill manifest: $surface_path"
      continue
    fi

    repos_with_skill_surfaces=$((repos_with_skill_surfaces + 1))
    if [[ "$json_surface" -eq 1 ]]; then
      if ! "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$repo_path" "${MODE_ARGS[@]}"; then
        if repo_has_dirty_tracked_changes "$repo_path"; then
          hg_warn "Managed repo has dirty tracked changes; skipping source-contract failure for skill sync drift: $repo_name"
        else
          overall_pending=1
          skill_pending=1
        fi
      fi
    else
      hg_warn "Managed skill manifest is YAML-only; repo-local Claude mirror sync still requires JSON: $surface_path"
    fi

    local canonical_name
    for canonical_name in "${skill_names[@]}"; do
      [[ -n "$canonical_name" ]] || continue
      if [[ -n "${managed_canonical_owners[$canonical_name]:-}" ]]; then
        hg_die "Duplicate managed canonical skill name: $canonical_name"
      fi
      managed_canonical_owners["$canonical_name"]="$repo_name"
      printf 'canonical\t%s\t%s\t%s\t%s\n' "$canonical_name" "$repo_name" "$repo_path" "$canonical_name" >>"$skill_entries_file"
      global_canonical_count=$((global_canonical_count + 1))
    done

    while IFS=$'\t' read -r alias_name canonical_name; do
      [[ -n "$alias_name" ]] || continue
      alias_counts["$alias_name"]=$(( ${alias_counts[$alias_name]:-0} + 1 ))
      printf '%s\t%s\t%s\t%s\n' "$alias_name" "$repo_name" "$repo_path" "$canonical_name" >>"$alias_entries_file"
    done < <(manifest_alias_pairs "$surface_path")
  done < <(
    jq -r '
      .repos[]
      | select(.baseline_target == true)
      | [.name, (if ((.path // "") | length) > 0 then .path else .name end)]
      | @tsv
    ' "$MANIFEST_PATH"
  )

  while IFS=$'\t' read -r alias_name repo_name repo_path canonical_name; do
    [[ -n "$alias_name" ]] || continue

    if [[ "${alias_counts[$alias_name]:-0}" -gt 1 ]]; then
      if [[ -z "${alias_conflicts_reported[$alias_name]:-}" ]]; then
        alias_conflicts_reported["$alias_name"]=1
        global_alias_conflict_names=$((global_alias_conflict_names + 1))
        hg_warn "Skipping conflicting global Claude alias: $alias_name (${alias_counts[$alias_name]} managed providers)"
      fi
      continue
    fi

    if [[ -n "${managed_canonical_owners[$alias_name]:-}" ]]; then
      if [[ -z "${alias_conflicts_reported[$alias_name]:-}" ]]; then
        alias_conflicts_reported["$alias_name"]=1
        global_alias_conflict_names=$((global_alias_conflict_names + 1))
        hg_warn "Skipping conflicting global Claude alias: $alias_name (collides with canonical ${managed_canonical_owners[$alias_name]})"
      fi
      continue
    fi

    printf 'alias\t%s\t%s\t%s\t%s\n' "$alias_name" "$repo_name" "$repo_path" "$canonical_name" >>"$skill_entries_file"
    global_alias_count=$((global_alias_count + 1))
  done <"$alias_entries_file"

  while IFS=$'\t' read -r kind global_name repo_name repo_path canonical_name; do
    [[ -n "$global_name" ]] || continue

    local source_skill canonical_refs staged_dir target_dir description
    if [[ "$kind" == "canonical" ]]; then
      source_skill="$repo_path/.agents/skills/$canonical_name/SKILL.md"
    else
      source_skill="$repo_path/.claude/skills/$global_name/SKILL.md"
    fi
    canonical_refs="$repo_path/.agents/skills/$canonical_name/references"
    staged_dir="$tmpdir/workspace-skill-$global_name"
    target_dir="$CLAUDE_SKILLS_DIR/$global_name"

    desired_claude_skill_dirs["$global_name"]=1

    if [[ ! -f "$source_skill" ]]; then
      overall_pending=1
      skill_pending=1
      case "$MODE" in
        write)
          hg_die "Missing managed workspace skill source: $source_skill"
          ;;
        dry-run)
          hg_warn "Would need managed workspace skill source before global sync: $source_skill"
          ;;
        check)
          hg_warn "Missing managed workspace skill source: $source_skill"
          ;;
      esac
      continue
    fi

    description="$(extract_frontmatter_scalar "$source_skill" "description")"
    exported_gemini_name="$(workspace_global_exported_skill_name "$repo_name" "$global_name")"
    if ! hg_gemini_name_is_builtin "$exported_gemini_name"; then
      printf '%s\t%s\t%s\t%s\t%s\n' "$exported_gemini_name" "${description:-Compatibility mirror for $canonical_name.}" "$kind" "$repo_name" "$canonical_name" >>"$tmpdir/gemini-skill-catalog.tsv"
    fi

    rm -rf "$staged_dir"
    mkdir -p "$staged_dir"
    cp "$source_skill" "$staged_dir/SKILL.md"

    if [[ -d "$canonical_refs" ]]; then
      mkdir -p "$staged_dir/references"
      cp -R "$canonical_refs"/. "$staged_dir/references/"
    fi

    jq -n \
      --arg generator "hg-workspace-global-sync.sh" \
      --arg repo "$repo_name" \
      --arg path "$repo_path" \
      --arg canonical "$canonical_name" \
      --arg kind "$kind" \
      --arg name "$global_name" \
      '{
        generator: $generator,
        repo: $repo,
        repo_path: $path,
        canonical: $canonical,
        kind: $kind,
        name: $name
      }' >"$staged_dir/$SKILL_OWNER_FILE"

    sync_skill_dir "$staged_dir" "$target_dir" || true
  done < <(sort -t $'\t' -k2,2 "$skill_entries_file")

  purge_stale_owned_skill_dirs

  if $CHECK_GLOBAL_TARGETS; then
    if ! HG_CLAUDE_COMMANDS_DIR="$CLAUDE_COMMANDS_DIR" \
         HG_CLAUDE_SKILLS_DIR="$CLAUDE_SKILLS_DIR" \
         HG_AGENTS_SKILLS_DIR="$AGENTS_SKILLS_DIR" \
         HG_CODEX_SKILLS_DIR="$CODEX_SKILLS_DIR" \
         "$SCRIPT_DIR/hg-global-skill-sync.sh" "${MODE_ARGS[@]}"; then
      overall_pending=1
      skill_pending=1
    fi
  fi
}

sync_managed_workspace_tools() {
  : >"$tmpdir/claude-mcp.ndjson"
  : >"$tmpdir/codex-mcp.ndjson"
  : >"$tmpdir/gemini-mcp.ndjson"

  local foundational_name foundation_server foundation_server_resolved
  for foundational_name in systemd tmux process; do
    foundation_server="$(jq -c --arg name "$foundational_name" '.mcpServers[$name]' "$ROOT_MCP_PATH")"
    [[ "$foundation_server" != "null" ]] || hg_die "Shared root .mcp.json missing foundational server: $foundational_name"

    foundation_server_resolved="$(
      jq -cn \
        --argjson server "$foundation_server" \
        --arg workspace_root "$WORKSPACE_ROOT" '
          $server
          | if ((.cwd // "") | startswith("${HOME}/hairglasses-studio")) then
              .cwd = ($workspace_root + ((.cwd // "") | sub("^\\$\\{HOME\\}/hairglasses-studio"; "")))
            elif ((.cwd // "") | startswith("$HOME/hairglasses-studio")) then
              .cwd = ($workspace_root + ((.cwd // "") | sub("^\\$HOME/hairglasses-studio"; "")))
            else
              .
            end
        '
    )"

    append_claude_entry "$foundational_name" "$foundation_server_resolved"
    append_gemini_entry "$foundational_name" "$foundation_server_resolved" "$WORKSPACE_ROOT"
    append_codex_entry \
      "$foundational_name" \
      "shared root .mcp.json:$foundational_name" \
      "$(jq -r '.command' <<<"$foundation_server_resolved")" \
      "$(jq -c '.args // []' <<<"$foundation_server_resolved")" \
      "$(normalize_codex_cwd "$WORKSPACE_ROOT" "$(jq -r '.cwd // ""' <<<"$foundation_server_resolved")")" \
      "$(normalize_codex_env_json "$(jq -c '.env // {}' <<<"$foundation_server_resolved")")" \
      '[]'
    claude_foundational_count=$((claude_foundational_count + 1))
  done

  while IFS=$'\t' read -r repo_name repo_rel_path; do
    [[ -n "$repo_name" ]] || continue

    local repo_path mcp_json policy_path
    repo_path="$(resolve_repo_path "$repo_rel_path")"
    mcp_json="$repo_path/.mcp.json"
    policy_path="$repo_path/.codex/mcp-profile-policy.json"

    if [[ ! -d "$repo_path" ]]; then
      hg_warn "Skipping managed repo missing from workspace: $repo_path"
      continue
    fi

    [[ -f "$mcp_json" ]] || continue
    [[ -f "$policy_path" ]] || continue

    jq -e '.mcpServers and (.mcpServers | type == "object")' "$mcp_json" >/dev/null || hg_die "Invalid .mcp.json: $mcp_json"
    jq -e '.version and .profiles and (.profiles | type == "array")' "$policy_path" >/dev/null || hg_die "Invalid MCP profile policy: $policy_path"
    jq -e '
      [.profiles[].name]
      | group_by(.)
      | map(select(length > 1))
      | length == 0
    ' "$policy_path" >/dev/null || hg_die "Duplicate MCP profile names in $policy_path"

    while IFS= read -r profile_json; do
      [[ -n "$profile_json" ]] || continue

      local profile_name mode source_name raw_id raw_server resolved resolved_server resolved_claude_server global_stem
      profile_name="$(jq -r '.name' <<<"$profile_json")"
      mode="$(jq -r '.mode' <<<"$profile_json")"
      source_name="$(jq -r '.from' <<<"$profile_json")"
      raw_id="${repo_name}:${source_name}"

      raw_server="$(jq -c --arg name "$source_name" '.mcpServers[$name]' "$mcp_json")"
      [[ "$raw_server" != "null" ]] || hg_die "Policy source server not found: $repo_name/.mcp.json:$source_name"

      if profile_exports_raw_source "$profile_json" && [[ -z "${raw_source_seen[$raw_id]:-}" ]]; then
        raw_source_seen["$raw_id"]=1
        append_claude_entry "${repo_name}_${source_name}" "$raw_server"
        append_gemini_entry "${repo_name}_${source_name}" "$raw_server" "$repo_path"
        claude_raw_count=$((claude_raw_count + 1))
      fi

      resolved="$(
        jq -cn \
          --argjson profile "$profile_json" \
          --slurpfile source "$mcp_json" '
            ($source[0].mcpServers[$profile.from]) as $server
            | if $server == null then
                error("missing source server: " + $profile.from)
              else
                {
                  command: ($profile.override.command // $server.command),
                  args: ($profile.override.args // $server.args // []),
                  cwd: ($profile.override.cwd // $server.cwd // ""),
                  env: (($server.env // {}) + ($profile.override.env // {})),
                  enabled_tools: ($profile.enabled_tools // [])
                }
              end
          '
      )" || hg_die "Failed to resolve managed MCP profile: $repo_name:$profile_name"
      resolved_server="$(jq -c 'del(.enabled_tools)' <<<"$resolved")"
      resolved_claude_server="$(
        jq -cn \
          --arg command "$(jq -r '.command' <<<"$resolved_server")" \
          --argjson args "$(jq -c '.args // []' <<<"$resolved_server")" \
          --arg cwd "$(normalize_codex_cwd "$repo_path" "$(jq -r '.cwd // ""' <<<"$resolved_server")")" \
          --argjson env "$(normalize_codex_env_json "$(jq -c '.env // {}' <<<"$resolved_server")")" \
          '{
            command: $command,
            args: $args
          }
          + (if ($cwd | length) > 0 then {cwd: $cwd} else {} end)
          + (if ($env | length) > 0 then {env: $env} else {} end)'
      )"
      global_stem="$(profile_global_name "$profile_json")"

      if profile_exports_global_claude "$profile_json"; then
        append_claude_entry "$global_stem" "$resolved_claude_server"
      fi

      if profile_exports_global_gemini "$profile_json"; then
        append_gemini_entry "$global_stem" "$resolved_server" "$repo_path"
      fi

      if profile_exports_global_codex "$profile_json"; then
        append_codex_entry \
          "$global_stem" \
          "$repo_name/.codex/mcp-profile-policy.json:$profile_name ($mode <- $source_name)" \
          "$(jq -r '.command' <<<"$resolved")" \
          "$(jq -c '.args' <<<"$resolved")" \
          "$(normalize_codex_cwd "$repo_path" "$(jq -r '.cwd' <<<"$resolved")")" \
          "$(normalize_codex_env_json "$(jq -c '.env' <<<"$resolved")")" \
          "$(jq -c '.enabled_tools' <<<"$resolved")"

        codex_profile_count=$((codex_profile_count + 1))
      fi
    done < <(jq -c '.profiles[] | select((.mode == "review" or .mode == "research") or (.global_codex // false) == true or (.global_claude // false) == true or (.global_gemini // false) == true)' "$policy_path")
  done < <(
    jq -r '
      .repos[]
      | select(.baseline_target == true)
      | [.name, (if ((.path // "") | length) > 0 then .path else .name end)]
      | @tsv
    ' "$MANIFEST_PATH"
  )

  local claude_generated_json claude_output_json claude_base_json
  claude_generated_json="$tmpdir/claude-generated.json"
  claude_output_json="$tmpdir/claude-output.json"
  claude_base_json="$tmpdir/claude-base.json"

  if [[ -s "$tmpdir/claude-mcp.ndjson" ]]; then
    jq -sc 'sort_by(.key) | (map({(.key): .value}) | add) // {}' "$tmpdir/claude-mcp.ndjson" >"$claude_generated_json"
  else
    printf '{}\n' >"$claude_generated_json"
  fi

  if [[ -f "$CLAUDE_JSON_PATH" ]]; then
    cp "$CLAUDE_JSON_PATH" "$claude_base_json"
  else
    printf '{}\n' >"$claude_base_json"
  fi

  jq \
    --arg project "$CLAUDE_PROJECT_KEY" \
    --slurpfile generated "$claude_generated_json" '
      .projects = (.projects // {})
      | .projects[$project] = (.projects[$project] // {})
      | .projects[$project].mcpServers = (
          ((.projects[$project].mcpServers // {}) | with_entries(select(.key | startswith("studio_") | not)))
          + $generated[0]
        )
    ' "$claude_base_json" >"$claude_output_json"
  sync_rendered_file "$CLAUDE_JSON_PATH" "$claude_output_json" "Synced workspace Claude MCP overlay"

  local gemini_generated_json gemini_output_json gemini_base_json
  gemini_generated_json="$tmpdir/gemini-generated.json"
  gemini_output_json="$tmpdir/gemini-output.json"
  gemini_base_json="$tmpdir/gemini-base.json"

  if [[ -s "$tmpdir/gemini-mcp.ndjson" ]]; then
    jq -sc 'sort_by(.name) | (map({(.name): .value}) | add) // {}' "$tmpdir/gemini-mcp.ndjson" >"$gemini_generated_json"
  else
    printf '{}\n' >"$gemini_generated_json"
  fi

  if [[ -f "$GEMINI_SETTINGS_PATH" ]]; then
    cp "$GEMINI_SETTINGS_PATH" "$gemini_base_json"
  else
    printf '{}\n' >"$gemini_base_json"
  fi

  jq \
    --slurpfile generated "$gemini_generated_json" '
      .mcpServers = (
        ((.mcpServers // {}) | with_entries(select(.key | startswith("studio-") | not)))
        + $generated[0]
      )
    ' "$gemini_base_json" >"$gemini_output_json"
  sync_rendered_file "$GEMINI_SETTINGS_PATH" "$gemini_output_json" "Synced workspace Gemini MCP overlay"

  local codex_block_file codex_output_file codex_base_file has_new has_legacy
  codex_block_file="$tmpdir/codex-generated-block.toml"
  codex_output_file="$tmpdir/codex-output.toml"
  codex_base_file="$tmpdir/codex-base.toml"

  {
    printf '%s\n' "$START_MARKER"
    printf '# Generated by dotfiles/scripts/hg-workspace-global-sync.sh from %s\n' "$WORKSPACE_ROOT"

    if [[ -s "$tmpdir/codex-mcp.ndjson" ]]; then
      while IFS= read -r entry; do
        [[ -n "$entry" ]] || continue

        local entry_name entry_comment entry_command entry_cwd entry_args_json entry_env_json entry_tools_json
        entry_name="$(jq -r '.name' <<<"$entry")"
        entry_comment="$(jq -r '.comment // ""' <<<"$entry")"
        entry_command="$(jq -r '.command' <<<"$entry")"
        entry_cwd="$(expand_home_tokens "$(jq -r '.cwd // ""' <<<"$entry")")"
        entry_args_json="$(jq -c '.args // []' <<<"$entry")"
        entry_env_json="$(jq -c '.env // {}' <<<"$entry")"
        entry_tools_json="$(jq -c '.enabled_tools // []' <<<"$entry")"

        printf '\n'
        [[ -n "$entry_comment" ]] && printf '# %s\n' "$entry_comment"
        printf '[mcp_servers.%s]\n' "$entry_name"
        emit_scalar_line "command" "$entry_command"

        local args=()
        mapfile -t args < <(jq -r '.[]' <<<"$entry_args_json")
        if [[ "${#args[@]}" -gt 0 ]]; then
          emit_array_line "args" "${args[@]}"
        fi

        if [[ -n "$entry_cwd" ]]; then
          emit_scalar_line "cwd" "$entry_cwd"
        fi

        local tools=()
        mapfile -t tools < <(jq -r '.[]' <<<"$entry_tools_json")
        if [[ "${#tools[@]}" -gt 0 ]]; then
          printf 'enabled_tools = [\n'
          local tool
          for tool in "${tools[@]}"; do
            printf '  %s,\n' "$(toml_quote "$tool")"
          done
          printf ']\n'
        fi

        emit_env_block "$entry_name" "$entry_env_json"
      done < <(jq -sc 'sort_by(.name)[]' "$tmpdir/codex-mcp.ndjson")
    fi

    printf '\n%s\n' "$END_MARKER"
  } >"$codex_block_file"

  if [[ -f "$CODEX_CONFIG_PATH" ]]; then
    cp "$CODEX_CONFIG_PATH" "$codex_base_file"
    has_new=0
    has_legacy=0
    grep -Fqx "$START_MARKER" "$codex_base_file" && has_new=1
    grep -Fqx "$LEGACY_START_MARKER" "$codex_base_file" && has_legacy=1

    if [[ "$has_new" -eq 1 && "$has_legacy" -eq 1 ]]; then
      hg_die "Both workspace-global and legacy global MCP blocks exist in $CODEX_CONFIG_PATH; clean one up first"
    fi

    if [[ "$has_new" -eq 1 ]]; then
      replace_marked_region "$codex_base_file" "$codex_block_file" "$START_MARKER" "$END_MARKER" >"$codex_output_file"
    elif [[ "$has_legacy" -eq 1 ]]; then
      replace_marked_region "$codex_base_file" "$codex_block_file" "$LEGACY_START_MARKER" "$LEGACY_END_MARKER" >"$codex_output_file"
    else
      insert_new_region "$codex_base_file" "$codex_block_file" >"$codex_output_file"
    fi
  else
    cp "$codex_block_file" "$codex_output_file"
  fi

  sync_rendered_file "$CODEX_CONFIG_PATH" "$codex_output_file" "Synced workspace Codex MCP block"
}

sync_gemini_home_context() {
  local gemini_block_file gemini_output_file gemini_base_file gemini_projects_base gemini_projects_output
  gemini_block_file="$tmpdir/gemini-managed-block.md"
  gemini_output_file="$tmpdir/gemini-home-output.md"
  gemini_base_file="$tmpdir/gemini-home-base.md"
  gemini_projects_base="$tmpdir/gemini-projects-base.json"
  gemini_projects_output="$tmpdir/gemini-projects-output.json"

  {
    printf '%s\n' "$HOME_CONTEXT_START_MARKER"
    printf '## Managed Workspace Global Context\n\n'
    printf -- '- Managed workspace root: `%s`\n' "$WORKSPACE_ROOT"
    printf -- '- Canonical inventory: `%s`\n' "$MANIFEST_PATH"
    printf -- '- Use repo-local `GEMINI.md`, `AGENTS.md`, and `CLAUDE.md` first for repo-specific instructions.\n'
    printf -- '- Shared research repo: `%s/docs`\n' "$WORKSPACE_ROOT"
    printf -- '- Shared root `.mcp.json` remains intentionally small: `systemd`, `tmux`, `process`.\n'
    printf -- '- Workspace-managed repo skills are exported globally with a `<repo>-...` prefix so they do not collide with repo-local canonical names.\n'
    printf '\n'

    if [[ -s "$tmpdir/gemini-skill-catalog.tsv" ]]; then
      printf '### Managed Global Workflow Catalog\n\n'
      while IFS=$'\t' read -r skill_name description kind repo_name canonical_name; do
        [[ -n "$skill_name" ]] || continue
        printf -- '- `%s`: %s' "$skill_name" "$description"
        if [[ "$kind" == "alias" ]]; then
          printf ' (alias for `%s` from `%s`)' "$canonical_name" "$repo_name"
        fi
        printf '\n'
      done < <(sort -t $'\t' -k1,1 "$tmpdir/gemini-skill-catalog.tsv")
      printf '\n'
    fi

    printf '### Managed Global MCP Overlays\n\n'
    printf -- '- Claude home memory: `%s`.\n' "$CLAUDE_HOME_DOC_PATH"
    printf -- '- Claude workspace overlay: `%s` managed entries under project `%s` in `%s`.\n' "$claude_tool_count" "$CLAUDE_PROJECT_KEY" "$CLAUDE_JSON_PATH"
    printf -- '- Codex workspace overlay: `%s` managed entries in `%s`.\n' "$codex_tool_count" "$CODEX_CONFIG_PATH"
    printf -- '- Gemini CLI home overlay: `%s` managed entries in `%s`.\n' "$gemini_tool_count" "$GEMINI_SETTINGS_PATH"
    printf -- '- Antigravity home overlay: managed MCP/workflow/env/settings state via `%s`.\n' "$ANTIGRAVITY_SYNC_SCRIPT"
    printf -- '- Portable managed skills are discovered from `%s`.\n' "$AGENTS_SKILLS_DIR"
    printf '\n%s\n' "$HOME_CONTEXT_END_MARKER"
  } >"$gemini_block_file"

  if [[ -f "$GEMINI_HOME_DOC_PATH" ]]; then
    cp "$GEMINI_HOME_DOC_PATH" "$gemini_base_file"
    if grep -Fqx "$HOME_CONTEXT_START_MARKER" "$gemini_base_file"; then
      replace_marked_region "$gemini_base_file" "$gemini_block_file" "$HOME_CONTEXT_START_MARKER" "$HOME_CONTEXT_END_MARKER" >"$gemini_output_file"
    else
      {
        cat "$gemini_base_file"
        if [[ -s "$gemini_base_file" ]]; then
          printf '\n\n'
        fi
        cat "$gemini_block_file"
      } >"$gemini_output_file"
    fi
  else
    cp "$gemini_block_file" "$gemini_output_file"
  fi
  sync_rendered_file "$GEMINI_HOME_DOC_PATH" "$gemini_output_file" "Synced Gemini home memory"

  if [[ -f "$GEMINI_PROJECTS_PATH" ]]; then
    cp "$GEMINI_PROJECTS_PATH" "$gemini_projects_base"
  else
    printf '{ "projects": {} }\n' >"$gemini_projects_base"
  fi

  jq \
    --arg workspace "$WORKSPACE_ROOT" \
    --arg workspace_name "$(basename "$WORKSPACE_ROOT")" \
    --arg docs_path "$WORKSPACE_ROOT/docs" \
    '.projects = (.projects // {})
     | .projects[$workspace] = (.projects[$workspace] // $workspace_name)
     | .projects[$docs_path] = (.projects[$docs_path] // "docs")' \
    "$gemini_projects_base" >"$gemini_projects_output"
  sync_rendered_file "$GEMINI_PROJECTS_PATH" "$gemini_projects_output" "Synced Gemini project registry"
}

sync_claude_home_context() {
  local claude_block_file claude_output_file claude_base_file
  claude_block_file="$tmpdir/claude-managed-block.md"
  claude_output_file="$tmpdir/claude-home-output.md"
  claude_base_file="$tmpdir/claude-home-base.md"

  {
    printf '%s\n' "$HOME_CONTEXT_START_MARKER"
    printf '## Managed Workspace Global Context\n\n'
    printf -- '- Managed workspace root: `%s`\n' "$WORKSPACE_ROOT"
    printf -- '- Canonical inventory: `%s`\n' "$MANIFEST_PATH"
    printf -- '- Use repo-local `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` first for repo-specific instructions.\n'
    printf -- '- Shared research repo: `%s/docs`\n' "$WORKSPACE_ROOT"
    printf -- '- Provider launchers route `codex`, `claude`, and `gemini` through root-owned managed worktrees under `/root/.codex/worktrees`.\n'
    printf '\n'
    printf '### Managed Global MCP Overlays\n\n'
    printf -- '- Claude workspace overlay: `%s` managed entries under project `%s` in `%s`.\n' "$claude_tool_count" "$CLAUDE_PROJECT_KEY" "$CLAUDE_JSON_PATH"
    printf -- '- Codex workspace overlay: `%s` managed entries in `%s`.\n' "$codex_tool_count" "$CODEX_CONFIG_PATH"
    printf -- '- Gemini CLI home overlay: `%s` managed entries in `%s`.\n' "$gemini_tool_count" "$GEMINI_SETTINGS_PATH"
    printf -- '- Antigravity home overlay: managed MCP/workflow/env/settings state via `%s`.\n' "$ANTIGRAVITY_SYNC_SCRIPT"
    printf -- '- Portable managed skills are discovered from `%s`.\n' "$AGENTS_SKILLS_DIR"
    printf '\n%s\n' "$HOME_CONTEXT_END_MARKER"
  } >"$claude_block_file"

  if [[ -f "$CLAUDE_HOME_DOC_PATH" ]]; then
    cp "$CLAUDE_HOME_DOC_PATH" "$claude_base_file"
    if grep -Fqx "$HOME_CONTEXT_START_MARKER" "$claude_base_file"; then
      replace_marked_region "$claude_base_file" "$claude_block_file" "$HOME_CONTEXT_START_MARKER" "$HOME_CONTEXT_END_MARKER" >"$claude_output_file"
    else
      {
        cat "$claude_base_file"
        if [[ -s "$claude_base_file" ]]; then
          printf '\n\n'
        fi
        cat "$claude_block_file"
      } >"$claude_output_file"
    fi
  else
    cp "$claude_block_file" "$claude_output_file"
  fi
  sync_rendered_file "$CLAUDE_HOME_DOC_PATH" "$claude_output_file" "Synced Claude home memory"
}

sync_antigravity_home_state() {
  if ! $CHECK_GLOBAL_TARGETS; then
    return 0
  fi

  [[ -f "$ANTIGRAVITY_SYNC_SCRIPT" ]] || hg_die "Antigravity sync script not found: $ANTIGRAVITY_SYNC_SCRIPT"

  if "$ANTIGRAVITY_SYNC_SCRIPT" --root "$WORKSPACE_ROOT" "${MODE_ARGS[@]}"; then
    return 0
  fi

  overall_pending=1
  tool_pending=1
  stale_antigravity_overlay_count=$((stale_antigravity_overlay_count + 1))
}

if $RUN_SKILLS; then
  sync_managed_workspace_skills
else
  managed_repo_count="$(jq -r '[.repos[] | select(.baseline_target == true)] | length' "$MANIFEST_PATH")"
fi

if $RUN_TOOLS; then
  sync_managed_workspace_tools
fi

if $RUN_SKILLS || $RUN_TOOLS; then
  sync_antigravity_home_state
fi

if $RUN_SKILLS; then
  sync_claude_home_context
  sync_gemini_home_context
fi

if $RUN_SKILLS && $CHECK_GLOBAL_TARGETS; then
  hg_info "Managed workspace global skills: ${global_canonical_count} canonical + ${global_alias_count} aliases (${global_alias_conflict_names} conflicts skipped across ${repos_with_skill_surfaces} skill surfaces)"
  hg_info "Managed home context: Claude memory at $CLAUDE_HOME_DOC_PATH; Gemini memory at $GEMINI_HOME_DOC_PATH; Gemini project registry at $GEMINI_PROJECTS_PATH"
elif $RUN_SKILLS; then
  hg_info "Managed workspace global source contract validated for ${global_canonical_count} canonical skills across ${repos_with_skill_surfaces} skill surfaces"
fi

if $RUN_TOOLS && $CHECK_GLOBAL_TARGETS; then
  hg_info "Managed workspace global tools: Claude ${claude_tool_count} (${claude_foundational_count} shared + ${claude_raw_count} raw) / Codex ${codex_tool_count} (${claude_foundational_count} shared + ${codex_profile_count} curated) / Gemini ${gemini_tool_count} (${claude_foundational_count} shared + ${claude_raw_count} raw) / Antigravity via ${ANTIGRAVITY_SYNC_SCRIPT}"
elif $RUN_TOOLS; then
  hg_info "Managed workspace global tool sources validated: ${claude_raw_count} repo-local raw MCP sources + ${codex_profile_count} curated Codex profiles"
fi

if $CHECK_GLOBAL_TARGETS && [[ "$overall_pending" -ne 0 ]]; then
  hg_info "Managed home overlay drift breakdown: Claude skills ${global_claude_stale_count} stale + ${global_claude_unexpected_count} unexpected owned + ${global_claude_manual_block_count} manual blockers; Claude MCP ${stale_workspace_claude_overlay_count}; Claude home ${stale_claude_home_doc_count}; Codex MCP ${stale_workspace_codex_overlay_count}; Gemini MCP ${stale_workspace_gemini_overlay_count}; Gemini home ${stale_gemini_home_doc_count}; Gemini projects ${stale_gemini_projects_count}; Antigravity ${stale_antigravity_overlay_count}"
fi

case "$MODE" in
  write)
    if [[ "$overall_pending" -eq 0 ]]; then
      hg_ok "Workspace global sync already up to date for $managed_repo_count managed repos"
    else
      hg_info "Workspace global sync refreshed for $managed_repo_count managed repos"
    fi
    ;;
  dry-run)
    if [[ "$overall_pending" -eq 0 ]]; then
      hg_ok "No workspace-global changes needed for $managed_repo_count managed repos"
    fi
    ;;
  check)
    if [[ "$overall_pending" -eq 0 ]]; then
      hg_ok "Workspace global sync up to date for $managed_repo_count managed repos"
    else
      exit 1
    fi
    ;;
esac

if [[ "$skill_conflicts" -gt 0 && "$MODE" == "write" ]]; then
  hg_warn "Managed workspace global skill conflicts left untouched: $skill_conflicts"
fi
