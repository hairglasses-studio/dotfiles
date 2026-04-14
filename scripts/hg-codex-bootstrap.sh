#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-agent-parity.sh"

REPO_PATH=""
REPO_NAME=""
DEFAULT_SKILL_NAME=""
DEFAULT_SKILL_DESCRIPTION=""
ALLOW_DIRTY=false

usage() {
  cat <<'EOF'
Usage: hg-codex-bootstrap.sh <repo_path> [options]

Initialize the shared Codex/Gemini baseline for a repo without depending on the
retired surfacekit control plane.

Options:
  --repo-name <name>                 Repo name override
  --default-skill-name <name>        Create a default canonical skill when the repo has no skill surface
  --default-skill-description <text> Description used for the default skill scaffold
  --allow-dirty                      Allow overwriting tracked generated files
  -h, --help                         Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-name)
      [[ $# -ge 2 ]] || hg_die "--repo-name requires a value"
      REPO_NAME="$2"
      shift 2
      ;;
    --default-skill-name)
      [[ $# -ge 2 ]] || hg_die "--default-skill-name requires a value"
      DEFAULT_SKILL_NAME="$2"
      shift 2
      ;;
    --default-skill-description)
      [[ $# -ge 2 ]] || hg_die "--default-skill-description requires a value"
      DEFAULT_SKILL_DESCRIPTION="$2"
      shift 2
      ;;
    --allow-dirty)
      ALLOW_DIRTY=true
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

hg_require jq python3

REPO_PATH="$(cd "$REPO_PATH" && pwd)"
REPO_NAME="${REPO_NAME:-$(basename "$REPO_PATH")}"

copy_skill_if_missing() {
  local source_file="$1"
  local skill_name="$2"
  local target_dir="$REPO_PATH/.agents/skills/$skill_name"
  local target_file="$target_dir/SKILL.md"
  [[ -f "$target_file" ]] && return 0
  mkdir -p "$target_dir"
  cp "$source_file" "$target_file"
}

render_surface_manifest() {
  local skills_json="$1"
  mkdir -p "$REPO_PATH/.agents/skills"
  python3 - "$REPO_PATH/.agents/skills/surface.yaml" "$REPO_NAME" "$skills_json" <<'PY'
import json
import pathlib
import sys

target = pathlib.Path(sys.argv[1])
plugin_root = sys.argv[2]
skills = json.loads(sys.argv[3])

payload = {
    "version": 1,
    "plugin_root": plugin_root,
    "skills": [
        {
            "name": name,
            "claude_include_canonical": True,
            "export_plugin": False,
        }
        for name in skills
    ],
}
target.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
PY
}

ensure_default_skill_surface() {
  [[ -n "$DEFAULT_SKILL_NAME" ]] || return 0
  local skill_dir="$REPO_PATH/.agents/skills/$DEFAULT_SKILL_NAME"
  local skill_file="$skill_dir/SKILL.md"
  mkdir -p "$skill_dir"
  if [[ ! -f "$skill_file" ]]; then
    cat >"$skill_file" <<EOF
---
name: $DEFAULT_SKILL_NAME
description: ${DEFAULT_SKILL_DESCRIPTION:-Core build, test, release, and maintenance workflow for the $REPO_NAME repo.}
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - Glob
---

# $DEFAULT_SKILL_NAME

Use this skill when working in the $REPO_NAME repo.
EOF
  fi
  render_surface_manifest "[\"$DEFAULT_SKILL_NAME\"]"
}

ensure_skill_surface() {
  [[ -f "$REPO_PATH/.agents/skills/surface.yaml" ]] && return 0

  local -a skills=()
  local skill_dir skill_name

  while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || continue
    skill_name="$(basename "$skill_dir")"
    skills+=("$skill_name")
  done < <(find "$REPO_PATH/.agents/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)

  while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || continue
    skill_name="$(basename "$skill_dir")"
    if [[ ! " ${skills[*]} " =~ " ${skill_name} " ]]; then
      copy_skill_if_missing "$skill_dir/SKILL.md" "$skill_name"
      skills+=("$skill_name")
    fi
  done < <(find "$REPO_PATH/.claude/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort)

  if [[ "${#skills[@]}" -eq 0 ]]; then
    ensure_default_skill_surface
    return 0
  fi

  local skills_json
  skills_json="$(printf '%s\n' "${skills[@]}" | python3 -c 'import json,sys; skills=[line.strip() for line in sys.stdin if line.strip()]; print(json.dumps(sorted(dict.fromkeys(skills))))')"
  render_surface_manifest "$skills_json"
}

ensure_codex_config_baseline() {
  local target="$REPO_PATH/.codex/config.toml"
  [[ -f "$target" ]] && return 0

  local template
  template="$(hg_parity_codex_template_path "$REPO_NAME")"
  [[ -f "$template" ]] || hg_die "Missing Codex template: $template"

  mkdir -p "$(dirname "$target")"
  cp "$template" "$target"
}

sync_args=("$REPO_PATH" "--repo-name" "$REPO_NAME")
if $ALLOW_DIRTY; then
  sync_args+=("--allow-dirty")
fi

ensure_skill_surface
ensure_codex_config_baseline

if [[ -f "$REPO_PATH/AGENTS.md" || -f "$REPO_PATH/CLAUDE.md" ]]; then
  "$SCRIPT_DIR/hg-agent-docs.sh" "$REPO_PATH" >/dev/null
fi

"$SCRIPT_DIR/hg-provider-settings-sync.sh" "${sync_args[@]}" >/dev/null

if [[ -d "$REPO_PATH/.codex/agents" ]]; then
  "$SCRIPT_DIR/hg-provider-role-sync.sh" "$REPO_PATH" >/dev/null
fi

if [[ -f "$REPO_PATH/.agents/skills/surface.yaml" ]]; then
  "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$REPO_PATH" >/dev/null
fi

if [[ -f "$REPO_PATH/.mcp.json" && -f "$REPO_PATH/.codex/config.toml" ]]; then
  "$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$REPO_PATH" >/dev/null
fi
