#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"
WORKSPACE_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
ROOT_MCP_PATH="${HG_ROOT_MCP_PATH:-}"
ANTIGRAVITY_DIR="${HG_ANTIGRAVITY_DIR:-$HOME/.gemini/antigravity}"
ANTIGRAVITY_MCP_PATH="${HG_ANTIGRAVITY_MCP_CONFIG_PATH:-$ANTIGRAVITY_DIR/mcp_config.json}"
ANTIGRAVITY_WORKFLOWS_DIR="${HG_ANTIGRAVITY_WORKFLOWS_DIR:-$ANTIGRAVITY_DIR/global_workflows}"
ANTIGRAVITY_METADATA_PATH="${HG_ANTIGRAVITY_METADATA_PATH:-$ANTIGRAVITY_DIR/.hg-antigravity-sync.json}"
ANTIGRAVITY_SETTINGS_PATH="${HG_ANTIGRAVITY_SETTINGS_PATH:-$HOME/.config/Antigravity/User/settings.json}"
ANTIGRAVITY_ENV_PATH="${HG_ANTIGRAVITY_ENV_PATH:-$HOME/.config/antigravity/env.sh}"
ANTIGRAVITY_WRAPPER_PATH="${HG_ANTIGRAVITY_WRAPPER_PATH:-$HOME/.local/bin/antigravity-managed}"
ANTIGRAVITY_DESKTOP_PATH="${HG_ANTIGRAVITY_DESKTOP_PATH:-$HOME/.local/share/applications/antigravity.desktop}"
ROOT_RULE_PATH="${HG_ANTIGRAVITY_RULE_PATH:-$WORKSPACE_ROOT/.agents/rules/antigravity-workspace.md}"

pending=0
managed_workflow_updates=0
managed_workflow_removals=0
missing_env_vars=0
MCP_TOTAL_SERVER_COUNT=0
MCP_ROOT_SHARED_SERVER_COUNT=0
WORKFLOW_REPO_COUNT=0
WORKFLOW_SKILL_COUNT=0
WORKFLOW_MANAGED_COUNT=0

declare -A server_name_seen=()
declare -A workflow_name_seen=()
declare -A imported_env_values=()
declare -A imported_env_sources=()
declare -A managed_workflow_set=()

usage() {
  cat <<'EOF'
Usage: hg-antigravity-sync.sh [options]

Generate and sync local Antigravity MCP, workflow, env, launcher, and settings
artifacts for the hairglasses-studio workspace.

Options:
  --root <path>                    Workspace root (default: ~/hairglasses-studio)
  --root-mcp <path>                Shared root .mcp.json (default: <root>/.mcp.json)
  --antigravity-dir <path>         Antigravity managed state dir (default: ~/.gemini/antigravity)
  --antigravity-mcp <path>         Antigravity MCP config path
  --antigravity-workflows <path>   Antigravity global workflows dir
  --antigravity-metadata <path>    Antigravity metadata path
  --antigravity-settings <path>    Antigravity user settings path
  --antigravity-env <path>         Managed env bridge path
  --antigravity-wrapper <path>     Managed Antigravity launcher wrapper path
  --antigravity-desktop <path>     User-local Antigravity desktop entry path
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
    --antigravity-workflows)
      [[ $# -ge 2 ]] || hg_die "--antigravity-workflows requires a path"
      ANTIGRAVITY_WORKFLOWS_DIR="$2"
      shift 2
      ;;
    --antigravity-metadata)
      [[ $# -ge 2 ]] || hg_die "--antigravity-metadata requires a path"
      ANTIGRAVITY_METADATA_PATH="$2"
      shift 2
      ;;
    --antigravity-settings)
      [[ $# -ge 2 ]] || hg_die "--antigravity-settings requires a path"
      ANTIGRAVITY_SETTINGS_PATH="$2"
      shift 2
      ;;
    --antigravity-env)
      [[ $# -ge 2 ]] || hg_die "--antigravity-env requires a path"
      ANTIGRAVITY_ENV_PATH="$2"
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
    --rule-path)
      [[ $# -ge 2 ]] || hg_die "--rule-path requires a path"
      ROOT_RULE_PATH="$2"
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

hg_require jq mktemp diff find sort awk realpath

WORKSPACE_ROOT="$(cd "$WORKSPACE_ROOT" && pwd)"
ROOT_MCP_PATH="${ROOT_MCP_PATH:-$WORKSPACE_ROOT/.mcp.json}"
[[ -d "$WORKSPACE_ROOT" ]] || hg_die "Workspace root not found: $WORKSPACE_ROOT"
[[ -f "$ROOT_MCP_PATH" ]] || hg_die "Shared root MCP file not found: $ROOT_MCP_PATH"
[[ -f /usr/bin/antigravity ]] || hg_die "Antigravity CLI not found at /usr/bin/antigravity"

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
    hg_warn "Antigravity MCP cwd does not exist, keeping original path: $cwd" >&2
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

remove_managed_file() {
  local target="$1"
  local label="$2"
  [[ -e "$target" ]] || return 0
  pending=1
  managed_workflow_removals=$((managed_workflow_removals + 1))
  case "$MODE" in
    write)
      rm -f "$target"
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

list_workspace_git_repos() {
  if [[ -d "$WORKSPACE_ROOT/.git" ]]; then
    printf '%s\n' "$WORKSPACE_ROOT"
  fi
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
        + (if (.disabled // null) != null then {disabled: .disabled} else {} end)
        + (if (.disabledTools // null) != null then {disabledTools: (.disabledTools // [])} else {} end)
        + (if (.tools // null) != null then {tools: .tools} else {} end)
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
    local name value
    name="$(jq -r '.name' <<<"$entry")"
    value="$(jq -c '.value' <<<"$entry")"
    append_antigravity_server "$output" "$name" "$value" "$WORKSPACE_ROOT"
  done < <(
    jq -c '
      (.mcpServers // {})
      | to_entries[]
      | select(.key | startswith("_") | not)
      | {name: .key, value: .value}
    ' "$ROOT_MCP_PATH"
  )

  while IFS= read -r repo; do
    [[ "$repo" == "$WORKSPACE_ROOT" ]] && continue
    [[ -f "$repo/.mcp.json" ]] || continue
    local repo_name repo_key
    repo_name="$(basename "$repo")"
    repo_key="$(sanitize_kebab_name "$repo_name")"
    while IFS= read -r entry; do
      [[ -n "$entry" ]] || continue
      local name value
      name="$(jq -r '.name' <<<"$entry")"
      value="$(jq -c '.value' <<<"$entry")"
      append_antigravity_server "$output" "${repo_key}-$(sanitize_kebab_name "$name")" "$value" "$repo"
    done < <(
      jq -c '
        (.mcpServers // {})
        | to_entries[]
        | select(.key | startswith("_") | not)
        | {name: .key, value: .value}
      ' "$repo/.mcp.json"
    )
  done < <(list_workspace_git_repos)

  if [[ -s "$output" ]]; then
    jq -S '{mcpServers: ((map({(.name): .value}) | add) // {})}' < <(jq -sc '.' "$output") >"$tmpdir/mcp_config.json"
  else
    printf '{ "mcpServers": {} }\n' >"$tmpdir/mcp_config.json"
  fi

  MCP_TOTAL_SERVER_COUNT="$(jq -r '.mcpServers | keys | length' "$tmpdir/mcp_config.json")"
  MCP_ROOT_SHARED_SERVER_COUNT="$(jq -r '(.mcpServers // {}) | keys | length' "$ROOT_MCP_PATH")"

  sync_rendered_file "$ANTIGRAVITY_MCP_PATH" "$tmpdir/mcp_config.json" "Synced Antigravity MCP config"
}

render_rule_file() {
  cat >"$tmpdir/antigravity-rule.md" <<EOF
# Antigravity Workspace Rule

Use this workspace as a control plane first, not as a flat file dump.

- Read \`$WORKSPACE_ROOT/CLAUDE.md\` and \`$WORKSPACE_ROOT/AGENTS.md\` before making repo-wide assumptions.
- Use \`$WORKSPACE_ROOT/workspace/manifest.json\` as the canonical managed-repo inventory.
- Check \`$WORKSPACE_ROOT/docs/indexes/SEARCH-GUIDE.md\` before net-new research.
- Prefer \`$WORKSPACE_ROOT/docs/\` for org-wide context and \`$WORKSPACE_ROOT/whiteclaw/\` as a reusable reference repo.
- Prefer repo-local \`AGENTS.md\`, \`CLAUDE.md\`, and \`GEMINI.md\` before broad scans.
- Skip generated and cache-heavy trees unless the task explicitly targets them: \`workspace/\`, \`.ralph/\`, \`.tools/\`, \`node_modules/\`, \`.venv/\`, \`.cache/\`, \`dist/\`, \`build/\`, and \`jellyfin-stack/vendor/\`.
EOF
  sync_rendered_file "$ROOT_RULE_PATH" "$tmpdir/antigravity-rule.md" "Synced workspace Antigravity rule"
}

write_workflow_file() {
  local path="$1"
  shift
  printf '%s\n' "$@" >"$path"
}

build_antigravity_workflows() {
  local staging_dir="$tmpdir/workflows"
  local repo_summary_file="$tmpdir/repo-workflow-summary.tsv"
  local skill_catalog_file="$tmpdir/skill-workflow-summary.tsv"
  local managed_names_file="$tmpdir/managed-workflows.txt"
  mkdir -p "$staging_dir"
  : >"$repo_summary_file"
  : >"$skill_catalog_file"
  : >"$managed_names_file"

  while IFS= read -r repo; do
    local repo_name repo_key skill_count=0 repo_skill_lines=()
    repo_name="$(basename "$repo")"
    repo_key="$(sanitize_kebab_name "$repo_name")"

    while IFS= read -r skill_file; do
      [[ -f "$skill_file" ]] || continue
      local skill_dir skill_name skill_key skill_desc workflow_name workflow_path rel_skill_path
      skill_dir="$(basename "$(dirname "$skill_file")")"
      skill_name="$(extract_frontmatter_scalar "$skill_file" "name")"
      [[ -n "$skill_name" ]] || skill_name="$skill_dir"
      skill_key="$(sanitize_kebab_name "$skill_dir")"
      workflow_name="${repo_key}--${skill_key}.md"
      if [[ -n "${workflow_name_seen[$workflow_name]:-}" ]]; then
        hg_die "Duplicate Antigravity workflow name: $workflow_name"
      fi
      workflow_name_seen["$workflow_name"]=1
      managed_workflow_set["$workflow_name"]=1
      printf '%s\n' "$workflow_name" >>"$managed_names_file"

      skill_desc="$(extract_frontmatter_scalar "$skill_file" "description")"
      [[ -n "$skill_desc" ]] || skill_desc="Use this workflow when the task matches the local skill surface."
      rel_skill_path="${skill_file#$WORKSPACE_ROOT/}"
      workflow_path="$staging_dir/$workflow_name"
      write_workflow_file "$workflow_path" \
        "# ${repo_name}: ${skill_name}" \
        "" \
        "- Source skill: \`$WORKSPACE_ROOT/$rel_skill_path\`" \
        "- Repo root: \`$repo\`" \
        "- Activation: $skill_desc" \
        "- Route through the repo-local MCP surface and companion docs instead of rewriting the skill body here." \
        "- Read repo-local \`AGENTS.md\`, \`CLAUDE.md\`, and \`GEMINI.md\` first when they exist."

      repo_skill_lines+=("- \`${skill_name}\`: ${skill_desc}")
      printf '%s\t%s\t%s\n' "$repo_name" "$skill_name" "$workflow_name" >>"$skill_catalog_file"
      skill_count=$((skill_count + 1))
    done < <(
      find "$repo/.agents/skills" -mindepth 2 -maxdepth 2 -type f -name SKILL.md 2>/dev/null | sort
    )

    if [[ "$skill_count" -eq 0 ]]; then
      continue
    fi

    local repo_workflow_name repo_workflow_path
    repo_workflow_name="${repo_key}.md"
    if [[ -n "${workflow_name_seen[$repo_workflow_name]:-}" ]]; then
      hg_die "Duplicate Antigravity workflow name: $repo_workflow_name"
    fi
    workflow_name_seen["$repo_workflow_name"]=1
    managed_workflow_set["$repo_workflow_name"]=1
    printf '%s\n' "$repo_workflow_name" >>"$managed_names_file"
    repo_workflow_path="$staging_dir/$repo_workflow_name"
    {
      printf '# %s workspace workflows\n\n' "$repo_name"
      printf -- '- Repo root: `%s`\n' "$repo"
      printf -- '- Skill source: `%s/.agents/skills/`\n' "$repo"
      printf -- '- Preferred docs: repo-local `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, and `README.md`\n'
      printf '\n## Available workflows\n\n'
      printf '%s\n' "${repo_skill_lines[@]}"
    } >"$repo_workflow_path"
    printf '%s\t%s\t%s\n' "$repo_name" "$skill_count" "$repo_workflow_name" >>"$repo_summary_file"
  done < <(list_workspace_git_repos)

  local routing_path="$staging_dir/00-workspace-routing.md"
  managed_workflow_set["00-workspace-routing.md"]=1
  printf '%s\n' "00-workspace-routing.md" >>"$managed_names_file"
  {
    printf '# Workspace routing\n\n'
    printf -- '- Workspace root: `%s`\n' "$WORKSPACE_ROOT"
    printf -- '- Canonical inventory: `%s/workspace/manifest.json`\n' "$WORKSPACE_ROOT"
    printf -- '- Shared research root: `%s/docs`\n' "$WORKSPACE_ROOT"
    printf -- '- Shared Antigravity rule: `%s`\n' "$ROOT_RULE_PATH"
    printf '\n## Priority routing\n\n'
    printf -- '- Start with `CLAUDE.md`, `AGENTS.md`, and repo-local instructions before broad scans.\n'
    printf -- '- Use `docs/indexes/SEARCH-GUIDE.md` before net-new research.\n'
    printf -- '- Treat `whiteclaw/` as a reusable reference repo when a pattern should be extracted.\n'
    printf -- '- Prefer the generated repo workflows below when the task clearly maps to an existing local skill.\n'
    if [[ -s "$repo_summary_file" ]]; then
      printf '\n## Repo workflow fronts\n\n'
      while IFS=$'\t' read -r repo_name skill_count workflow_name; do
        printf -- '- `%s`: %s skill workflows via `%s`\n' "$repo_name" "$skill_count" "$workflow_name"
      done <"$repo_summary_file"
    fi
  } >"$routing_path"

  while IFS= read -r workflow_name; do
    local staged_path target_path
    [[ -n "$workflow_name" ]] || continue
    staged_path="$staging_dir/$workflow_name"
    target_path="$ANTIGRAVITY_WORKFLOWS_DIR/$workflow_name"
    if [[ -f "$target_path" ]] && cmp -s "$target_path" "$staged_path"; then
      continue
    fi
    pending=1
    managed_workflow_updates=$((managed_workflow_updates + 1))
    case "$MODE" in
      write)
        mkdir -p "$ANTIGRAVITY_WORKFLOWS_DIR"
        cp "$staged_path" "$target_path"
        hg_ok "Synced Antigravity workflow: $target_path"
        ;;
      dry-run)
        hg_warn "Would update Antigravity workflow: $target_path"
        ;;
      check)
        hg_warn "Out of date Antigravity workflow: $target_path"
        ;;
    esac
  done < <(sort -u "$managed_names_file")

  WORKFLOW_MANAGED_COUNT="$(sort -u "$managed_names_file" | wc -l | tr -d ' ')"
  WORKFLOW_SKILL_COUNT="$(awk '/--.*\.md$/ {count++} END {print count + 0}' "$managed_names_file")"
  WORKFLOW_REPO_COUNT="$(awk '!/--/ && $0 !~ /^00-/ {count++} END {print count + 0}' "$managed_names_file")"

  if [[ -f "$ANTIGRAVITY_METADATA_PATH" ]]; then
    while IFS= read -r old_name; do
      [[ -n "$old_name" ]] || continue
      [[ -n "${managed_workflow_set[$old_name]:-}" ]] && continue
      remove_managed_file "$ANTIGRAVITY_WORKFLOWS_DIR/$old_name" "Antigravity workflow"
    done < <(jq -r '.managed_workflows[]? // empty' "$ANTIGRAVITY_METADATA_PATH" 2>/dev/null || true)
  fi
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

build_env_bridge() {
  parse_env_file "$WORKSPACE_ROOT/.env"
  while IFS= read -r repo; do
    [[ "$repo" == "$WORKSPACE_ROOT" ]] && continue
    parse_env_file "$repo/.env"
  done < <(list_workspace_git_repos)

  {
    printf '#!/usr/bin/env bash\n'
    printf '# Generated by dotfiles/scripts/hg-antigravity-sync.sh from %s\n' "$WORKSPACE_ROOT"
    local key
    for key in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY GEMINI_API_KEY; do
      if [[ -n "${imported_env_values[$key]:-}" ]]; then
        printf 'export %s=%q\n' "$key" "${imported_env_values[$key]}"
      else
        missing_env_vars=$((missing_env_vars + 1))
      fi
    done
  } >"$tmpdir/antigravity-env.sh"
  chmod 600 "$tmpdir/antigravity-env.sh"
  sync_rendered_file "$ANTIGRAVITY_ENV_PATH" "$tmpdir/antigravity-env.sh" "Synced Antigravity env bridge"
}

build_launcher_wrapper() {
  cat >"$tmpdir/antigravity-managed" <<EOF
#!/usr/bin/env bash
set -euo pipefail

if [[ -f "$ANTIGRAVITY_ENV_PATH" ]]; then
  # shellcheck disable=SC1090
  source "$ANTIGRAVITY_ENV_PATH"
fi

exec /usr/bin/antigravity "\$@"
EOF
  chmod 755 "$tmpdir/antigravity-managed"
  sync_rendered_file "$ANTIGRAVITY_WRAPPER_PATH" "$tmpdir/antigravity-managed" "Synced Antigravity launcher wrapper"
}

build_desktop_entry() {
  cat >"$tmpdir/antigravity.desktop" <<EOF
[Desktop Entry]
Name=Antigravity
Comment=Experience liftoff
GenericName=Text Editor
Exec=$ANTIGRAVITY_WRAPPER_PATH %F
Icon=antigravity
Type=Application
StartupNotify=false
StartupWMClass=Antigravity
Categories=TextEditor;Development;IDE;
MimeType=application/x-antigravity-workspace;
Actions=new-empty-window;
Keywords=vscode;

[Desktop Action new-empty-window]
Name=New Empty Window
Exec=$ANTIGRAVITY_WRAPPER_PATH --new-window %F
Icon=antigravity
EOF
  sync_rendered_file "$ANTIGRAVITY_DESKTOP_PATH" "$tmpdir/antigravity.desktop" "Synced Antigravity desktop entry"
}

build_settings_json() {
  local search_excludes watcher_excludes
  search_excludes='{
    "workspace/**": true,
    "**/.ralph/**": true,
    "**/.tools/**": true,
    "**/node_modules/**": true,
    "**/.venv/**": true,
    "**/.cache/**": true,
    "**/dist/**": true,
    "**/build/**": true,
    "jellyfin-stack/vendor/**": true
  }'
  watcher_excludes="$search_excludes"

  if [[ -f "$ANTIGRAVITY_SETTINGS_PATH" ]]; then
    cp "$ANTIGRAVITY_SETTINGS_PATH" "$tmpdir/antigravity-settings-base.json"
  else
    printf '{}\n' >"$tmpdir/antigravity-settings-base.json"
  fi

  jq -S \
    --argjson search_excludes "$search_excludes" \
    --argjson watcher_excludes "$watcher_excludes" '
      . + {
        "security.workspace.trust.untrustedFiles": "open",
        "antigravity.searchMaxWorkspaceFileCount": 50000,
        "antigravity.persistentLanguageServer": true,
        "antigravity.enableCursorImportCursor": false,
        "files.watcherExclude": ((.["files.watcherExclude"] // {}) + $watcher_excludes),
        "search.exclude": ((.["search.exclude"] // {}) + $search_excludes)
      }
    ' "$tmpdir/antigravity-settings-base.json" >"$tmpdir/antigravity-settings.json"

  sync_rendered_file "$ANTIGRAVITY_SETTINGS_PATH" "$tmpdir/antigravity-settings.json" "Synced Antigravity user settings"
}

write_metadata() {
  local imported_vars_file="$tmpdir/imported-vars.txt"
  local missing_vars_file="$tmpdir/missing-vars.txt"
  local managed_workflows_file="$tmpdir/managed-workflows-final.txt"
  local imported_sources_file="$tmpdir/imported-sources.json"
  : >"$imported_vars_file"
  : >"$missing_vars_file"

  local key
  for key in OPENAI_API_KEY ANTHROPIC_API_KEY GOOGLE_API_KEY GEMINI_API_KEY; do
    if [[ -n "${imported_env_values[$key]:-}" ]]; then
      printf '%s\n' "$key" >>"$imported_vars_file"
    else
      printf '%s\n' "$key" >>"$missing_vars_file"
    fi
  done
  printf '%s\n' "${!managed_workflow_set[@]}" | sort -u >"$managed_workflows_file"

  {
    printf '{\n'
    printf '  "generator": "dotfiles/scripts/hg-antigravity-sync.sh",\n'
    printf '  "workspace_root": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_ROOT" '$v')"
    printf '  "generated_on": %s,\n' "$(jq -Rn --arg v "$(date +%Y-%m-%d)" '$v')"
    printf '  "root_mcp_path": %s,\n' "$(jq -Rn --arg v "$ROOT_MCP_PATH" '$v')"
    printf '  "mcp_config_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_MCP_PATH" '$v')"
    printf '  "settings_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_SETTINGS_PATH" '$v')"
    printf '  "env_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_ENV_PATH" '$v')"
    printf '  "wrapper_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_WRAPPER_PATH" '$v')"
    printf '  "desktop_entry_path": %s,\n' "$(jq -Rn --arg v "$ANTIGRAVITY_DESKTOP_PATH" '$v')"
    printf '  "rule_path": %s,\n' "$(jq -Rn --arg v "$ROOT_RULE_PATH" '$v')"
    printf '  "total_mcp_server_count": %s,\n' "$MCP_TOTAL_SERVER_COUNT"
    printf '  "root_shared_server_count": %s,\n' "$MCP_ROOT_SHARED_SERVER_COUNT"
    printf '  "repo_workflow_count": %s,\n' "$WORKFLOW_REPO_COUNT"
    printf '  "skill_workflow_count": %s,\n' "$WORKFLOW_SKILL_COUNT"
    printf '  "managed_workflow_count": %s,\n' "$WORKFLOW_MANAGED_COUNT"
    printf '  "managed_workflows": %s,\n' "$(json_string_array_from_lines "$managed_workflows_file")"
    printf '  "imported_env_vars": %s,\n' "$(json_string_array_from_lines "$imported_vars_file")"
    printf '  "missing_env_vars": %s,\n' "$(json_string_array_from_lines "$missing_vars_file")"
    printf '  "mcp_name_collision_count": 0,\n'
    printf '  "workflow_name_collision_count": 0,\n'
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
render_rule_file
build_antigravity_workflows
build_env_bridge
build_launcher_wrapper
build_desktop_entry
build_settings_json
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
