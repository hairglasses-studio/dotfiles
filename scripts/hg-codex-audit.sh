#!/usr/bin/env bash
set -euo pipefail

ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
WRITE_DOCS=0
WRITE_WIKI_DOCS=0
WRITE_JSON=0

wiki_docs_dir_default() {
  printf '%s\n' "$ROOT/docs/projects/codex-migration"
}

for arg in "$@"; do
  case "$arg" in
    --write-docs)
      WRITE_DOCS=1
      ;;
    --write-wiki-docs)
      WRITE_WIKI_DOCS=1
      ;;
    --write-json)
      WRITE_JSON=1
      ;;
    *)
      echo "Unknown argument: $arg" >&2
      exit 1
      ;;
  esac
done

if ! command -v rg >/dev/null 2>&1; then
  echo "ripgrep (rg) is required" >&2
  exit 1
fi

EXCLUDE_GLOBS=(
  "--glob=!whiteclaw/**"
  "--glob=!.git/**"
  "--glob=!node_modules/**"
  "--glob=!.claude/worktrees/**"
  "--glob=!.ralph/worktrees/**"
  "--glob=!.github/docs/**"
  "--glob=!bin/**"
  "--glob=!build/**"
  "--glob=!dist/**"
)

inventory_csv=""
inventory_json_rows=""
json_separator=""
inventory_md=$'| Repo | `claude mcp` | `.claude/settings.json` | `claude_desktop_config.json` | `AGENTS.md` | `.codex/config.toml` | `.codex-plugin` |\n|------|--------------:|------------------------:|-----------------------------:|-----------:|---------------------:|----------------:|\n'

count_matches() {
  local repo="$1"
  local pattern="$2"
  local count
  count=$(rg -n "$pattern" "$repo" "${EXCLUDE_GLOBS[@]}" 2>/dev/null | wc -l | tr -d ' ')
  printf '%s' "$count"
}

count_files() {
  local repo="$1"
  local path_pattern="$2"
  local count
  count=$(find "$repo" -path "$path_pattern" 2>/dev/null | wc -l | tr -d ' ')
  printf '%s' "$count"
}

repos=()
while IFS= read -r repo; do
  repos+=("$repo")
done < <(find "$ROOT" -mindepth 2 -maxdepth 2 -type d -name .git -prune | sed 's#/\.git$##' | sort)

total_claude_mcp=0
total_claude_settings=0
total_claude_desktop=0
total_missing_agents=0
total_missing_codex=0
total_missing_plugins=0
scanned_repos=0

for repo in "${repos[@]}"; do
  name=$(basename "$repo")
  if [[ "$name" == "docs" || "$name" == "whiteclaw" ]]; then
    continue
  fi
  scanned_repos=$((scanned_repos + 1))
  claude_mcp=$(count_matches "$repo" 'claude mcp')
  claude_settings=$(count_matches "$repo" '\.claude/settings\.json')
  claude_desktop=$(count_matches "$repo" 'claude_desktop_config\.json')
  has_agents=$(count_files "$repo" '*/AGENTS.md')
  has_codex=$(count_files "$repo" '*/.codex/config.toml')
  has_plugin=$(count_files "$repo" '*/.codex-plugin/plugin.json')

  total_claude_mcp=$((total_claude_mcp + claude_mcp))
  total_claude_settings=$((total_claude_settings + claude_settings))
  total_claude_desktop=$((total_claude_desktop + claude_desktop))
  if [[ "$has_agents" -eq 0 ]]; then
    total_missing_agents=$((total_missing_agents + 1))
  fi
  if [[ "$has_codex" -eq 0 ]]; then
    total_missing_codex=$((total_missing_codex + 1))
  fi
  if [[ "$has_plugin" -eq 0 ]]; then
    total_missing_plugins=$((total_missing_plugins + 1))
  fi

  inventory_csv+="${name},${claude_mcp},${claude_settings},${claude_desktop},${has_agents},${has_codex},${has_plugin}"$'\n'
  inventory_json_rows+="${json_separator}"$'    {\n'
  inventory_json_rows+="      \"repo\": \"${name}\","$'\n'
  inventory_json_rows+="      \"claude_mcp_mentions\": ${claude_mcp},"$'\n'
  inventory_json_rows+="      \"claude_settings_mentions\": ${claude_settings},"$'\n'
  inventory_json_rows+="      \"claude_desktop_config_mentions\": ${claude_desktop},"$'\n'
  inventory_json_rows+="      \"agents_md_count\": ${has_agents},"$'\n'
  inventory_json_rows+="      \"codex_config_count\": ${has_codex},"$'\n'
  inventory_json_rows+="      \"codex_plugin_count\": ${has_plugin}"$'\n'
  inventory_json_rows+=$'    }\n'
  json_separator=$',\n'
  inventory_md+="| ${name} | ${claude_mcp} | ${claude_settings} | ${claude_desktop} | ${has_agents} | ${has_codex} | ${has_plugin} |"$'\n'
done

cat <<EOF
Codex audit root: $ROOT
Repos scanned: ${scanned_repos}
claude mcp matches: $total_claude_mcp
.claude/settings.json matches: $total_claude_settings
claude_desktop_config.json matches: $total_claude_desktop
repos missing AGENTS.md: $total_missing_agents
repos missing .codex/config.toml: $total_missing_codex
repos missing .codex-plugin/plugin.json: $total_missing_plugins
EOF

inventory_json="{
  \"generated_on\": \"$(date +%Y-%m-%d)\",
  \"root\": \"${ROOT}\",
  \"summary\": {
    \"repos_scanned\": ${scanned_repos},
    \"claude_mcp_matches\": ${total_claude_mcp},
    \"claude_settings_matches\": ${total_claude_settings},
    \"claude_desktop_config_matches\": ${total_claude_desktop},
    \"repos_missing_agents_md\": ${total_missing_agents},
    \"repos_missing_codex_config\": ${total_missing_codex},
    \"repos_missing_codex_plugin\": ${total_missing_plugins}
  },
  \"repos\": [
${inventory_json_rows}
  ]
}"

if [[ "$WRITE_DOCS" -eq 1 ]]; then
  docs_dir="$ROOT/docs/codex-migration"
  mkdir -p "$docs_dir"

  cat >"$docs_dir/repo-inventory.csv" <<EOF
repo,claude_mcp_mentions,claude_settings_mentions,claude_desktop_config_mentions,agents_md_count,codex_config_count,codex_plugin_count
${inventory_csv%$'\n'}
EOF

  cat >"$docs_dir/repo-inventory.md" <<EOF
# Repo Inventory

Generated by \`dotfiles/scripts/hg-codex-audit.sh --write-docs\` on $(date +%Y-%m-%d).

${inventory_md}
EOF
fi

if [[ "$WRITE_JSON" -eq 1 ]]; then
  docs_dir="$ROOT/docs/codex-migration"
  wiki_docs_dir="$(wiki_docs_dir_default)"
  mkdir -p "$docs_dir" "$wiki_docs_dir"

  printf '%s\n' "$inventory_json" >"$docs_dir/repo-inventory.json"
  printf '%s\n' "$inventory_json" >"$wiki_docs_dir/repo-inventory.json"
fi

if [[ "$WRITE_WIKI_DOCS" -eq 1 ]]; then
  wiki_docs_dir="$(wiki_docs_dir_default)"
  mkdir -p "$wiki_docs_dir"

  cat >"$wiki_docs_dir/repo-inventory.csv" <<EOF
repo,claude_mcp_mentions,claude_settings_mentions,claude_desktop_config_mentions,agents_md_count,codex_config_count,codex_plugin_count
${inventory_csv%$'\n'}
EOF

  cat >"$wiki_docs_dir/repo-inventory.md" <<EOF
# Repo Inventory

Generated by \`dotfiles/scripts/hg-codex-audit.sh --write-wiki-docs\` on $(date +%Y-%m-%d).

Summary from the latest audit:

- Repos scanned: ${scanned_repos}
- \`claude mcp\` matches: ${total_claude_mcp}
- \`.claude/settings.json\` matches: ${total_claude_settings}
- \`claude_desktop_config.json\` matches: ${total_claude_desktop}
- Repos missing \`AGENTS.md\`: ${total_missing_agents}
- Repos missing \`.codex/config.toml\`: ${total_missing_codex}
- Repos missing \`.codex-plugin/plugin.json\`: ${total_missing_plugins}

${inventory_md}
EOF
fi

if [[ "$WRITE_JSON" -eq 1 ]]; then
  docs_dir="$ROOT/docs/codex-migration"
  wiki_docs_dir="$(wiki_docs_dir_default)"
  mkdir -p "$docs_dir" "$wiki_docs_dir"

  printf '%s\n' "$inventory_json" >"$docs_dir/repo-inventory.json"
  printf '%s\n' "$inventory_json" >"$wiki_docs_dir/repo-inventory.json"
fi
