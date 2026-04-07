#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"
REPO_PATH=""
SETTINGS_PATH=""
TEMPLATE_PATH=""
MCP_JSON_PATH=""
CLAUDE_SETTINGS_PATH=""
OWNER_PATH=""
LEGACY_CONFIG_PATH=""
ALLOW_DIRTY=false

usage() {
  cat <<'EOF'
Usage: hg-gemini-settings-sync.sh <repo_path> [options]

Generate repo-local .gemini/settings.json from the shared template, repo .mcp.json,
and the supported subset of .claude/settings.json hooks.

Options:
  --settings <path>          Output settings path (default: <repo>/.gemini/settings.json)
  --template <path>          Template JSON (default: dotfiles/templates/gemini-settings.standard.json)
  --mcp-json <path>          MCP source JSON (default: <repo>/.mcp.json)
  --claude-settings <path>   Claude settings source (default: <repo>/.claude/settings.json)
  --owner <path>             Generator metadata path (default: <repo>/.gemini/.hg-gemini-settings-sync.json)
  --legacy-config <path>     Legacy Gemini YAML path (default: <repo>/.gemini/config.yaml)
  --allow-dirty              Overwrite dirty generated Gemini settings during scaffolding/onboarding
  --dry-run                  Print a unified diff without writing
  --check                    Exit non-zero if generated output is stale
  -h, --help                 Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --settings)
      [[ $# -ge 2 ]] || hg_die "--settings requires a path"
      SETTINGS_PATH="$2"
      shift 2
      ;;
    --template)
      [[ $# -ge 2 ]] || hg_die "--template requires a path"
      TEMPLATE_PATH="$2"
      shift 2
      ;;
    --mcp-json)
      [[ $# -ge 2 ]] || hg_die "--mcp-json requires a path"
      MCP_JSON_PATH="$2"
      shift 2
      ;;
    --claude-settings)
      [[ $# -ge 2 ]] || hg_die "--claude-settings requires a path"
      CLAUDE_SETTINGS_PATH="$2"
      shift 2
      ;;
    --owner)
      [[ $# -ge 2 ]] || hg_die "--owner requires a path"
      OWNER_PATH="$2"
      shift 2
      ;;
    --legacy-config)
      [[ $# -ge 2 ]] || hg_die "--legacy-config requires a path"
      LEGACY_CONFIG_PATH="$2"
      shift 2
      ;;
    --allow-dirty)
      ALLOW_DIRTY=true
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

hg_require jq mktemp diff

REPO_PATH="$(cd "$REPO_PATH" && pwd)"
SETTINGS_PATH="${SETTINGS_PATH:-$REPO_PATH/.gemini/settings.json}"
TEMPLATE_PATH="${TEMPLATE_PATH:-$SCRIPT_DIR/../templates/gemini-settings.standard.json}"
MCP_JSON_PATH="${MCP_JSON_PATH:-$REPO_PATH/.mcp.json}"
CLAUDE_SETTINGS_PATH="${CLAUDE_SETTINGS_PATH:-$REPO_PATH/.claude/settings.json}"
OWNER_PATH="${OWNER_PATH:-$REPO_PATH/.gemini/.hg-gemini-settings-sync.json}"
LEGACY_CONFIG_PATH="${LEGACY_CONFIG_PATH:-$REPO_PATH/.gemini/config.yaml}"

[[ -f "$TEMPLATE_PATH" ]] || hg_die "Missing Gemini settings template: $TEMPLATE_PATH"
jq -e 'type == "object"' "$TEMPLATE_PATH" >/dev/null || hg_die "Invalid Gemini settings template: $TEMPLATE_PATH"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

rendered_settings="$tmpdir/settings.json"
rendered_owner="$tmpdir/owner.json"
settings_changed=0
legacy_pending=0
dirty_blockers=0

repo_rel_path() {
  local abs="$1"
  if [[ "$abs" == "$REPO_PATH/"* ]]; then
    printf '%s\n' "${abs#"$REPO_PATH/"}"
  elif [[ "$abs" == "$REPO_PATH" ]]; then
    printf '.\n'
  else
    printf '%s\n' "$abs"
  fi
}

path_is_dirty() {
  local path="$1"
  local rel
  rel="$(repo_rel_path "$path")"

  [[ -d "$REPO_PATH/.git" ]] || return 1
  [[ -n "$(git -C "$REPO_PATH" status --porcelain --untracked-files=all -- "$rel" 2>/dev/null)" ]]
}

sync_rendered_file() {
  local target="$1"
  local rendered="$2"
  local label="$3"

  if [[ -f "$target" ]] && cmp -s "$target" "$rendered"; then
    rm -f "$rendered"
    return 0
  fi

  settings_changed=1
  if [[ -e "$target" ]] && ! $ALLOW_DIRTY && path_is_dirty "$target"; then
    dirty_blockers=$((dirty_blockers + 1))
    case "$MODE" in
      write)
        hg_warn "Skipping dirty $label: $target"
        ;;
      dry-run)
        hg_warn "Would update dirty $label: $target"
        ;;
      check)
        hg_warn "Dirty $label blocks sync: $target"
        ;;
    esac
    rm -f "$rendered"
    return 1
  fi

  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      mv "$rendered" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      if [[ -f "$target" ]]; then
        diff -u "$target" "$rendered" || true
      else
        diff -u /dev/null "$rendered" || true
      fi
      rm -f "$rendered"
      ;;
    check)
      rm -f "$rendered"
      ;;
  esac
}

remove_legacy_config() {
  [[ -e "$LEGACY_CONFIG_PATH" ]] || return 0

  legacy_pending=1
  if ! $ALLOW_DIRTY && path_is_dirty "$LEGACY_CONFIG_PATH"; then
    dirty_blockers=$((dirty_blockers + 1))
    case "$MODE" in
      write)
        hg_warn "Skipping dirty legacy Gemini config: $LEGACY_CONFIG_PATH"
        ;;
      dry-run)
        hg_warn "Would remove dirty legacy Gemini config: $LEGACY_CONFIG_PATH"
        ;;
      check)
        hg_warn "Dirty legacy Gemini config blocks sync: $LEGACY_CONFIG_PATH"
        ;;
    esac
    return 1
  fi

  case "$MODE" in
    write)
      rm -f "$LEGACY_CONFIG_PATH"
      hg_ok "Removed legacy Gemini config: $LEGACY_CONFIG_PATH"
      ;;
    dry-run)
      hg_warn "Would remove legacy Gemini config: $LEGACY_CONFIG_PATH"
      ;;
    check)
      ;;
  esac
}

render_mcp_json() {
  if [[ ! -f "$MCP_JSON_PATH" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -e '.mcpServers and (.mcpServers | type == "object")' "$MCP_JSON_PATH" >/dev/null || hg_die "Invalid .mcp.json: $MCP_JSON_PATH"

  jq '
    (.mcpServers // {})
    | with_entries(select(.key | startswith("_") | not))
    | with_entries(
        .value |= with_entries(
          select(
            .key == "command"
            or .key == "args"
            or .key == "env"
            or .key == "cwd"
            or .key == "url"
            or .key == "httpUrl"
            or .key == "headers"
            or .key == "timeout"
          )
        )
      )
  ' "$MCP_JSON_PATH"
}

render_hooks_json() {
  if [[ ! -f "$CLAUDE_SETTINGS_PATH" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -e 'type == "object"' "$CLAUDE_SETTINGS_PATH" >/dev/null || hg_die "Invalid Claude settings JSON: $CLAUDE_SETTINGS_PATH"

  jq '
    (.hooks // {}) as $hooks
    | {
        SessionStart: ($hooks.SessionStart // []),
        BeforeTool: ($hooks.PreToolUse // []),
        AfterTool: ($hooks.PostToolUse // []),
        Notification: ($hooks.Notification // []),
        BeforeAgent: ($hooks.UserPromptSubmit // [])
      }
    | with_entries(select((.value | type == "array") and (.value | length > 0)))
  ' "$CLAUDE_SETTINGS_PATH"
}

count_translated_hook_rules() {
  if [[ ! -f "$CLAUDE_SETTINGS_PATH" ]]; then
    printf '0\n'
    return 0
  fi

  jq '
    [
      (.hooks.SessionStart // []),
      (.hooks.PreToolUse // []),
      (.hooks.PostToolUse // []),
      (.hooks.Notification // []),
      (.hooks.UserPromptSubmit // [])
    ]
    | flatten
    | length
  ' "$CLAUDE_SETTINGS_PATH"
}

count_unsupported_hook_rules() {
  if [[ ! -f "$CLAUDE_SETTINGS_PATH" ]]; then
    printf '0\n'
    return 0
  fi

  jq '
    (.hooks // {})
    | [
        (.Stop // []),
        (.PostToolUseFailure // []),
        (.PostCompact // []),
        (.PreCompact // []),
        (.SubagentStart // []),
        (.SubagentStop // []),
        (.SessionEnd // [])
      ]
    | flatten
    | length
  ' "$CLAUDE_SETTINGS_PATH"
}

mcp_json="$tmpdir/mcp.json"
hooks_json="$tmpdir/hooks.json"
render_mcp_json >"$mcp_json"
render_hooks_json >"$hooks_json"

translated_hook_rules="$(count_translated_hook_rules)"
unsupported_hook_rules="$(count_unsupported_hook_rules)"
gemini_mcp_server_count="$(jq 'keys | length' "$mcp_json")"

jq \
  --slurpfile mcp "$mcp_json" \
  --slurpfile hooks "$hooks_json" \
  '
    .mcpServers = ($mcp[0] // {})
    | if (($hooks[0] // {}) | length) > 0 then
        .hooks = $hooks[0]
      else
        del(.hooks)
      end
  ' "$TEMPLATE_PATH" | jq '.' >"$rendered_settings"

jq -n \
  --arg generator "dotfiles/scripts/hg-gemini-settings-sync.sh" \
  --arg template "$TEMPLATE_PATH" \
  --arg mcp_json "$MCP_JSON_PATH" \
  --arg claude_settings "$CLAUDE_SETTINGS_PATH" \
  --argjson translated_hook_rules "$translated_hook_rules" \
  --argjson unsupported_hook_rules "$unsupported_hook_rules" \
  --argjson gemini_mcp_server_count "$gemini_mcp_server_count" \
  '{
    generator: $generator,
    template: $template,
    source_mcp_json: $mcp_json,
    source_claude_settings: $claude_settings,
    translated_hook_rules: $translated_hook_rules,
    unsupported_claude_hook_rules: $unsupported_hook_rules,
    gemini_mcp_server_count: $gemini_mcp_server_count
  }' | jq '.' >"$rendered_owner"

sync_rendered_file "$SETTINGS_PATH" "$rendered_settings" "Synced Gemini settings"
sync_rendered_file "$OWNER_PATH" "$rendered_owner" "Synced Gemini settings metadata"
remove_legacy_config || true

if [[ "$unsupported_hook_rules" -gt 0 && "$MODE" == "write" ]]; then
  hg_warn "Left $unsupported_hook_rules Claude-only hook rules provider-specific in $REPO_PATH"
fi

case "$MODE" in
  write)
    if [[ "$settings_changed" -eq 0 && "$legacy_pending" -eq 0 ]]; then
      hg_ok "Gemini settings already up to date for $REPO_PATH"
    fi
    ;;
  dry-run)
    if [[ "$settings_changed" -eq 0 && "$legacy_pending" -eq 0 ]]; then
      hg_ok "No Gemini settings changes needed for $REPO_PATH"
    fi
    ;;
  check)
    if [[ "$settings_changed" -eq 0 && "$legacy_pending" -eq 0 && "$dirty_blockers" -eq 0 ]]; then
      hg_ok "Gemini settings up to date for $REPO_PATH"
    else
      exit 1
    fi
    ;;
esac
