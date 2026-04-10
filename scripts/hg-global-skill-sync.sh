#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"

for arg in "$@"; do
  case "$arg" in
    --dry-run) MODE="dry-run" ;;
    --check)   MODE="check" ;;
    *)
      hg_die "Unknown argument: $arg"
      ;;
  esac
done

CLAUDE_COMMANDS="${HG_CLAUDE_COMMANDS_DIR:-$HOME/.claude/commands}"
CLAUDE_SKILLS="${HG_CLAUDE_SKILLS_DIR:-$HOME/.claude/skills}"
AGENTS_SKILLS="${HG_AGENTS_SKILLS_DIR:-$HOME/.agents/skills}"
CODEX_SKILLS="${HG_CODEX_SKILLS_DIR:-$HOME/.codex/skills}"

[[ -d "$CLAUDE_COMMANDS" ]] || hg_die "Claude commands directory not found: $CLAUDE_COMMANDS"

# Skills that are Claude Code-specific (use claude_* MCP tools, session recovery, etc.)
# These have no Codex equivalent and should not be synced.
CLAUDE_ONLY_SKILLS=(
  session-status
  recover
  recover-session
  recovery-dashboard
  workspace-snapshot
  session-diff
  session-autopsy
  recovery-verify
  crash-report
)

is_claude_only() {
  local name="$1"
  for skip in "${CLAUDE_ONLY_SKILLS[@]}"; do
    [[ "$name" == "$skip" ]] && return 0
  done
  return 1
}

# Convert kebab-case filename to snake_case skill name (Codex convention)
to_skill_name() {
  local name="$1"
  printf '%s' "$name" | tr '-' '_'
}

# Extract the first non-empty line as the description, trimmed cleanly
extract_description() {
  local file="$1"
  head -1 "$file" | sed 's/\. *$//; s/:$//; s/^ *//'
}

# Transform $ARGUMENTS references in the body for Codex compatibility.
# Codex receives arguments from conversation context, so we replace the
# placeholder with a note. The natural language instructions remain identical.
transform_body() {
  local file="$1"
  sed \
    -e 's/\$ARGUMENTS/[user-provided arguments]/g' \
    -e 's|`/reconnect`|reconnect all MCP servers|g' \
    -e 's|`/commit`|`git commit`|g' \
    -e 's|`/loop \([^`]*\)`|set up a recurring loop (\1)|g' \
    -e 's|`/pipeline \([^`]*\)`|re-run the pipeline (\1)|g' \
    -e 's| /commit| git commit|g' \
    -e 's| /pipeline | the pipeline |g' \
    -e 's|inside `/loop`|in recurring mode|g' \
    -e 's|`/loop`|a recurring loop|g' \
    -e 's| /loop | a recurring loop |g' \
    -e 's|Run `/reconnect`|Reconnect all MCP servers|g' \
    "$file"
}

# Strip mcp__* tool entries from allowed-tools in a SKILL.md file.
# Outputs the file with MCP-specific tools removed. If allowed-tools
# becomes empty after stripping, the entire block is removed.
strip_mcp_tools() {
  local file="$1"
  local in_frontmatter=0
  local frontmatter_count=0
  local in_allowed_tools=0
  local has_remaining_tools=0
  local allowed_tools_line=""
  local deferred_lines=()

  while IFS= read -r line; do
    # Track frontmatter boundaries
    if [[ "$line" == "---" ]]; then
      frontmatter_count=$((frontmatter_count + 1))
      if [[ $frontmatter_count -eq 1 ]]; then
        in_frontmatter=1
        printf '%s\n' "$line"
        continue
      elif [[ $frontmatter_count -eq 2 ]]; then
        # Flush deferred allowed-tools block if it has remaining tools
        if [[ $in_allowed_tools -eq 1 ]]; then
          if [[ $has_remaining_tools -eq 1 ]]; then
            printf '%s\n' "$allowed_tools_line"
            for dl in "${deferred_lines[@]}"; do
              printf '%s\n' "$dl"
            done
          fi
        fi
        in_frontmatter=0
        in_allowed_tools=0
        printf '%s\n' "$line"
        continue
      fi
    fi

    if [[ $in_frontmatter -eq 1 ]]; then
      # Detect start of allowed-tools block
      if [[ "$line" =~ ^allowed-tools: ]]; then
        in_allowed_tools=1
        has_remaining_tools=0
        allowed_tools_line="$line"
        deferred_lines=()
        continue
      fi

      # Inside allowed-tools block
      if [[ $in_allowed_tools -eq 1 ]]; then
        local stripped="${line#"${line%%[![:space:]]*}"}"
        if [[ "$stripped" == "- "* ]]; then
          if [[ "$stripped" == "- mcp__"* ]]; then
            # Skip MCP tool entries
            continue
          else
            has_remaining_tools=1
            deferred_lines+=("$line")
            continue
          fi
        else
          # Non-tool line — flush the block and continue
          if [[ $has_remaining_tools -eq 1 ]]; then
            printf '%s\n' "$allowed_tools_line"
            for dl in "${deferred_lines[@]}"; do
              printf '%s\n' "$dl"
            done
          fi
          in_allowed_tools=0
          printf '%s\n' "$line"
          continue
        fi
      fi
    fi

    printf '%s\n' "$line"
  done < "$file"
}

workspace_managed_export_name() {
  local skill_dir="$1"
  local owner_file="$skill_dir/.hg-workspace-global-sync.json"
  [[ -f "$owner_file" ]] || return 1

  local repo_name source_name
  repo_name="$(jq -r '.repo // empty' "$owner_file" 2>/dev/null)"
  source_name="$(jq -r '.name // empty' "$owner_file" 2>/dev/null)"
  [[ -n "$repo_name" && -n "$source_name" ]] || return 1

  printf '%s-%s' "$repo_name" "$source_name"
}

rewrite_skill_frontmatter_name() {
  local file="$1"
  local exported_name="$2"
  awk -v exported_name="$exported_name" '
    BEGIN { frontmatter = 0; renamed = 0 }
    /^---$/ { frontmatter++; print; next }
    frontmatter == 1 && renamed == 0 && /^name:[[:space:]]*/ {
      print "name: " exported_name
      renamed = 1
      next
    }
    { print }
  ' "$file"
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

pending_changes=0
stale_agents_count=0
stale_codex_count=0
unexpected_agents_count=0
unexpected_codex_count=0
gemini_builtin_name_skips=0
commands_shadowed_by_skills=0
declare -A desired_agents_dirs=()
declare -A desired_codex_dirs=()
declare -A claude_skill_names=()

if [[ -d "$CLAUDE_SKILLS" ]]; then
  while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || continue
    claude_skill_names["$(basename "$skill_dir")"]=1
  done < <(find "$CLAUDE_SKILLS" -mindepth 1 -maxdepth 1 -type d -print 2>/dev/null | sort)
fi

sync_file() {
  local staged="$1"
  local target="$2"
  local label="$3"

  if [[ -f "$target" ]] && cmp -s "$staged" "$target"; then
    return
  fi

  pending_changes=1
  case "$label" in
    *agents*)
      stale_agents_count=$((stale_agents_count + 1))
      ;;
    *codex*)
      stale_codex_count=$((stale_codex_count + 1))
      ;;
  esac
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      cp "$staged" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      hg_warn "Would update $label: $target"
      ;;
    check)
      hg_warn "Out of date $label: $target"
      ;;
  esac
}

# ---------------------------------------------------------------------------
# Pipeline 1: Claude commands --> Agents skills  (existing)
# Pipeline C: Claude commands --> Codex skills   (new)
# ---------------------------------------------------------------------------

commands_portable=0
commands_skipped=0

for source_file in "$CLAUDE_COMMANDS"/*.md; do
  [[ -f "$source_file" ]] || continue

  basename_md="$(basename "$source_file")"
  raw_name="${basename_md%.md}"

  if is_claude_only "$raw_name"; then
    commands_skipped=$((commands_skipped + 1))
    continue
  fi

  # Prefer the full Claude skill when a command shim and skill share a name.
  # Without this, check mode reports false drift because the command pass and
  # skill pass generate different bodies for the same agents/codex target.
  if [[ -n "${claude_skill_names[$raw_name]:-}" ]]; then
    commands_shadowed_by_skills=$((commands_shadowed_by_skills + 1))
    continue
  fi

  commands_portable=$((commands_portable + 1))
  skill_name="$(to_skill_name "$raw_name")"
  desired_codex_dirs["$raw_name"]=1
  description="$(extract_description "$source_file")"

  # Stage for agents (snake_case) unless the generated name collides with a Gemini built-in.
  if hg_gemini_name_is_builtin "$skill_name"; then
    gemini_builtin_name_skips=$((gemini_builtin_name_skips + 1))
  else
    desired_agents_dirs["$skill_name"]=1
    staged_agents="$tmpdir/cmd_agents_${skill_name}.md"
    {
      printf '%s\n' '---'
      printf 'name: %s\n' "$skill_name"
      printf 'description: >-\n'
      printf '  %s\n' "$description"
      printf '%s\n\n' '---'
      printf '<!-- GENERATED BY hg-global-skill-sync.sh FROM ~/.claude/commands/%s; DO NOT EDIT -->\n\n' "$basename_md"
      transform_body "$source_file"
    } >"$staged_agents"
    sync_file "$staged_agents" "$AGENTS_SKILLS/$skill_name/SKILL.md" "Synced command→agents"
  fi

  # Stage for Codex (kebab-case)
  staged_codex="$tmpdir/cmd_codex_${raw_name}.md"
  {
    printf '%s\n' '---'
    printf 'name: %s\n' "$raw_name"
    printf 'description: >-\n'
    printf '  %s\n' "$description"
    printf '%s\n\n' '---'
    printf '<!-- GENERATED BY hg-global-skill-sync.sh FROM ~/.claude/commands/%s; DO NOT EDIT -->\n\n' "$basename_md"
    transform_body "$source_file"
  } >"$staged_codex"
  sync_file "$staged_codex" "$CODEX_SKILLS/$raw_name/SKILL.md" "Synced command→codex"
done

# ---------------------------------------------------------------------------
# Pipeline A: Claude skills --> Codex skills   (new)
# Pipeline B: Claude skills --> Agents skills  (new)
# ---------------------------------------------------------------------------

skills_synced=0

if [[ -d "$CLAUDE_SKILLS" ]]; then
  for skill_dir in "$CLAUDE_SKILLS"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skill_src="$skill_dir/SKILL.md"
    [[ -f "$skill_src" ]] || continue

    name="$(basename "$skill_dir")"
    exported_name="$name"
    if workspace_exported_name="$(workspace_managed_export_name "$skill_dir")"; then
      exported_name="$workspace_exported_name"
    fi
    snake_name="$(to_skill_name "$exported_name")"
    skills_synced=$((skills_synced + 1))
    desired_codex_dirs["$exported_name"]=1

    # Stage stripped version
    staged_stripped="$tmpdir/skill_stripped_${name}.md"
    strip_mcp_tools "$skill_src" > "$staged_stripped"

    staged_named="$tmpdir/skill_named_${name}.md"
    rewrite_skill_frontmatter_name "$staged_stripped" "$exported_name" > "$staged_named"

    # Insert generation marker after frontmatter closing ---
    staged_codex="$tmpdir/skill_codex_${name}.md"
    awk -v marker="<!-- GENERATED BY hg-global-skill-sync.sh FROM ~/.claude/skills/${name}/SKILL.md; DO NOT EDIT -->" '
      BEGIN { fm=0 }
      /^---$/ { fm++; print; if (fm==2) print "\n" marker; next }
      { print }
    ' "$staged_named" > "$staged_codex"

    # Codex gets kebab-case dir name
    sync_file "$staged_codex" "$CODEX_SKILLS/$exported_name/SKILL.md" "Synced skill→codex"

    # Agents get only names that do not collide with Gemini built-in slash commands.
    if hg_gemini_name_is_builtin "$exported_name" || hg_gemini_name_is_builtin "$snake_name"; then
      gemini_builtin_name_skips=$((gemini_builtin_name_skips + 1))
    else
      desired_agents_dirs["$snake_name"]=1
      sync_file "$staged_codex" "$AGENTS_SKILLS/$snake_name/SKILL.md" "Synced skill→agents"
    fi
  done
fi

# ---------------------------------------------------------------------------
# Purge stale skill directories
# ---------------------------------------------------------------------------

GENERATED_MARKER='GENERATED BY hg-global-skill-sync.sh\|Portable skill synced from'

# Purge stale agents skills
if [[ -d "$AGENTS_SKILLS" ]]; then
  while IFS= read -r dir; do
    skill_name="$(basename "$dir")"
    if [[ -f "$dir/SKILL.md" ]] && grep -q "$GENERATED_MARKER" "$dir/SKILL.md" 2>/dev/null; then
      if [[ -z "${desired_agents_dirs[$skill_name]:-}" ]]; then
        pending_changes=1
        unexpected_agents_count=$((unexpected_agents_count + 1))
        case "$MODE" in
          write)
            rm -rf "$dir"
            hg_ok "Removed stale agents skill: $dir"
            ;;
          dry-run)
            hg_warn "Would remove stale agents skill: $dir"
            ;;
          check)
            hg_warn "Unexpected agents skill: $dir"
            ;;
        esac
      fi
    fi
  done < <(find "$AGENTS_SKILLS" -mindepth 1 -maxdepth 1 -type d -print 2>/dev/null | sort)
fi

# Purge stale Codex skills
if [[ -d "$CODEX_SKILLS" ]]; then
  while IFS= read -r dir; do
    skill_name="$(basename "$dir")"
    if [[ -f "$dir/SKILL.md" ]] && grep -q "$GENERATED_MARKER" "$dir/SKILL.md" 2>/dev/null; then
      if [[ -z "${desired_codex_dirs[$skill_name]:-}" ]]; then
        pending_changes=1
        unexpected_codex_count=$((unexpected_codex_count + 1))
        case "$MODE" in
          write)
            rm -rf "$dir"
            hg_ok "Removed stale codex skill: $dir"
            ;;
          dry-run)
            hg_warn "Would remove stale codex skill: $dir"
            ;;
          check)
            hg_warn "Unexpected codex skill: $dir"
            ;;
        esac
      fi
    fi
  done < <(find "$CODEX_SKILLS" -mindepth 1 -maxdepth 1 -type d -print 2>/dev/null | sort)
fi

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------

total_agents=$((commands_portable + skills_synced))
total_codex=$((commands_portable + skills_synced))

if [[ "$pending_changes" -ne 0 ]]; then
  hg_info "Global skill drift breakdown: agents ${stale_agents_count} stale + ${unexpected_agents_count} unexpected, codex ${stale_codex_count} stale + ${unexpected_codex_count} unexpected"
fi

case "$MODE" in
  write)
    if [[ "$pending_changes" -eq 0 ]]; then
      hg_ok "All skills synchronized (${commands_portable} commands + ${skills_synced} skills → agents+codex, ${commands_skipped} Claude-only skipped, ${commands_shadowed_by_skills} shadowed by full skills, ${gemini_builtin_name_skips} Gemini-reserved names skipped for agents)"
    else
      hg_info "Synchronized ${commands_portable} commands + ${skills_synced} skills → agents+codex (${commands_skipped} Claude-only skipped, ${commands_shadowed_by_skills} shadowed by full skills, ${gemini_builtin_name_skips} Gemini-reserved names skipped for agents)"
    fi
    ;;
  dry-run)
    if [[ "$pending_changes" -eq 0 ]]; then
      hg_ok "No changes needed (${commands_portable} commands + ${skills_synced} skills, ${commands_skipped} Claude-only, ${commands_shadowed_by_skills} shadowed by full skills, ${gemini_builtin_name_skips} Gemini-reserved names skipped for agents)"
    fi
    ;;
  check)
    if [[ "$pending_changes" -eq 0 ]]; then
      hg_ok "All skills up to date (${commands_portable} commands + ${skills_synced} skills, ${commands_skipped} Claude-only, ${commands_shadowed_by_skills} shadowed by full skills, ${gemini_builtin_name_skips} Gemini-reserved names skipped for agents)"
    else
      exit 1
    fi
    ;;
esac
