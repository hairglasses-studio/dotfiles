#!/usr/bin/env bash
# shellcheck shell=bash
# hg-agent-parity.sh — Shared helpers for Claude/Codex/Gemini parity sync and audit.

if [[ -z "${HG_STUDIO_ROOT:-}" ]]; then
  HG_AGENT_PARITY_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  source "$HG_AGENT_PARITY_LIB_DIR/hg-core.sh"
fi

hg_parity_require_tools() {
  hg_require jq
}

hg_parity_objectives_path() {
  if [[ -n "${HG_PARITY_OBJECTIVES_PATH:-}" ]]; then
    printf '%s\n' "$HG_PARITY_OBJECTIVES_PATH"
    return 0
  fi

  local preferred="$HG_STUDIO_ROOT/docs/projects/agent-parity/parity-objectives.json"
  local legacy="$HG_STUDIO_ROOT/docs/projects/codex-migration/parity-objectives.json"
  if [[ -f "$preferred" ]]; then
    printf '%s\n' "$preferred"
  else
    printf '%s\n' "$legacy"
  fi
}

hg_parity_manifest_path() {
  printf '%s\n' "$HG_STUDIO_ROOT/workspace/manifest.json"
}

hg_parity_repo_scope() {
  local repo_name="$1"
  local manifest
  manifest="$(hg_parity_manifest_path)"
  if [[ ! -f "$manifest" ]]; then
    printf 'active_first_party\n'
    return 0
  fi

  jq -r --arg repo "$repo_name" 'first(.repos[] | select(.name == $repo) | .scope) // "active_first_party"' "$manifest"
}

hg_parity_repo_objective_bool() {
  local repo_name="$1"
  local field="$2"
  local default_value="${3:-false}"
  local objectives
  objectives="$(hg_parity_objectives_path)"
  if [[ ! -f "$objectives" ]]; then
    printf '%s\n' "$default_value"
    return 0
  fi

  local repo_scope
  repo_scope="$(hg_parity_repo_scope "$repo_name")"

  jq -r \
    --arg repo "$repo_name" \
    --arg scope "$repo_scope" \
    --arg field "$field" \
    --argjson default "$default_value" '
      if ((.repo_overrides[$repo] // {}) | has($field)) then
        .repo_overrides[$repo][$field]
      elif ((.scope_defaults[$scope] // {}) | has($field)) then
        .scope_defaults[$scope][$field]
      elif ((.defaults // {}) | has($field)) then
        .defaults[$field]
      else
        $default
      end
    ' "$objectives"
}

hg_parity_object_json() {
  local file="$1"
  if [[ ! -f "$file" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -c 'if type == "object" then . else {} end' "$file"
}

hg_parity_repo_has_mcp_json() {
  local repo_path="$1"
  [[ -f "$repo_path/.mcp.json" ]] || return 1
  jq -e '(.mcpServers // {}) | type == "object" and (length > 0)' "$repo_path/.mcp.json" >/dev/null 2>&1
}

hg_parity_kebab_case_server_name() {
  local name="$1"
  printf '%s' "$name" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g'
}

hg_parity_claude_mcp_servers_json() {
  local repo_path="$1"
  if ! hg_parity_repo_has_mcp_json "$repo_path"; then
    printf '{}\n'
    return 0
  fi

  jq -c '.mcpServers // {} | if type == "object" then . else {} end' "$repo_path/.mcp.json"
}

hg_parity_gemini_mcp_servers_json() {
  local repo_path="$1"
  if ! hg_parity_repo_has_mcp_json "$repo_path"; then
    printf '{}\n'
    return 0
  fi

  jq -c '
    .mcpServers // {}
    | if type != "object" then {} else . end
    | with_entries(
        .key |= (
          ascii_downcase
          | gsub("[^a-z0-9]+"; "-")
          | gsub("^-+"; "")
          | gsub("-+$"; "")
          | gsub("-+"; "-")
        )
      )
  ' "$repo_path/.mcp.json"
}

hg_parity_repo_has_claude_hooks() {
  local repo_path="$1"
  if [[ -d "$repo_path/.claude/hooks" ]] && find "$repo_path/.claude/hooks" -type f | grep -q .; then
    return 0
  fi

  if [[ -f "$repo_path/.claude/settings.json" ]] && \
     jq -e '(.hooks // {}) | type == "object" and (keys | length > 0)' \
       "$repo_path/.claude/settings.json" >/dev/null 2>&1; then
    return 0
  fi

  return 1
}

hg_parity_gemini_extension_name() {
  local repo_name="$1"
  printf '%s-workspace\n' "$(hg_parity_kebab_case_server_name "$repo_name")"
}

hg_parity_gemini_extension_relpath() {
  local repo_name="$1"
  local ext_name
  ext_name="$(hg_parity_gemini_extension_name "$repo_name")"
  printf '.gemini/extensions/%s/gemini-extension.json\n' "$ext_name"
}

hg_parity_gemini_owner_path() {
  local repo_path="$1"
  printf '%s\n' "$repo_path/.gemini/.hg-gemini-settings-sync.json"
}

hg_parity_gemini_sync_metadata_json() {
  local repo_path="$1"
  local metadata
  metadata="$(hg_parity_gemini_owner_path "$repo_path")"
  if [[ ! -f "$metadata" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -c 'if type == "object" then . else {} end' "$metadata"
}

hg_parity_gemini_translated_hook_rule_count() {
  local repo_path="$1"
  local metadata
  metadata="$(hg_parity_gemini_owner_path "$repo_path")"
  if [[ ! -f "$metadata" ]]; then
    printf '0\n'
    return 0
  fi

  jq -r '.translated_hook_rules // 0' "$metadata"
}

hg_parity_gemini_unsupported_hook_rule_count() {
  local repo_path="$1"
  local metadata
  metadata="$(hg_parity_gemini_owner_path "$repo_path")"
  if [[ ! -f "$metadata" ]]; then
    printf '0\n'
    return 0
  fi

  jq -r '.unsupported_claude_hook_rules // 0' "$metadata"
}

hg_parity_repo_requires_gemini_extension() {
  local repo_path="$1"
  local repo_name="$2"
  [[ "$(hg_parity_repo_objective_bool "$repo_name" "gemini_extension_scaffold" "false")" == "true" ]]
}

hg_parity_render_claude_settings() {
  local repo_path="$1"
  local current_json generated_mcp
  current_json="$(hg_parity_object_json "$repo_path/.claude/settings.json")"
  generated_mcp="$(hg_parity_claude_mcp_servers_json "$repo_path")"

  jq -S -n \
    --argjson current "$current_json" \
    --argjson generated "$generated_mcp" '
      ($current | if type == "object" then . else {} end) as $cfg
      | if ($generated | length) > 0 then
          $cfg + {mcpServers: (($cfg.mcpServers // {}) + $generated)}
        else
          $cfg
        end
    '
}

hg_parity_render_gemini_settings() {
  local repo_path="$1"
  local current_json generated_mcp
  current_json="$(hg_parity_object_json "$repo_path/.gemini/settings.json")"
  generated_mcp="$(hg_parity_gemini_mcp_servers_json "$repo_path")"

  jq -S -n \
    --argjson current "$current_json" \
    --argjson generated "$generated_mcp" '
      ($current | if type == "object" then . else {} end) as $cfg
      | ($cfg.context // {}) as $context
      | (($context.fileName // []) + ["AGENTS.md", "GEMINI.md"] | unique) as $file_names
      | ($cfg + {context: ($context + {fileName: $file_names})}) as $base
      | if ($generated | length) > 0 then
          $base + {
            mcp: (($base.mcp // {}) + {
              allowed: (((($base.mcp.allowed // []) + ($generated | keys)) | unique) | sort)
            }),
            mcpServers: (($base.mcpServers // {}) + $generated)
          }
        else
          $base
        end
    '
}

hg_parity_render_gemini_extension() {
  local repo_path="$1"
  local repo_name="$2"
  if ! hg_parity_repo_requires_gemini_extension "$repo_path" "$repo_name"; then
    printf '{}\n'
    return 0
  fi

  local ext_name
  ext_name="$(hg_parity_gemini_extension_name "$repo_name")"

  jq -S -n \
    --arg name "$ext_name" '
      {
        name: $name,
        version: "1.0.0",
        contextFileName: "GEMINI.md"
      }
    '
}

hg_parity_compare_expected_file() {
  local expected="$1"
  local path="$2"
  [[ -f "$path" ]] || return 1
  diff -u <(printf '%s\n' "$expected") "$path" >/dev/null 2>&1
}

hg_parity_provider_mcp_bridge_ok() {
  local repo_path="$1"
  if ! hg_parity_repo_has_mcp_json "$repo_path"; then
    printf '1\n'
    return 0
  fi

  local expected_claude expected_gemini
  expected_claude="$(hg_parity_claude_mcp_servers_json "$repo_path")"
  expected_gemini="$(hg_parity_gemini_mcp_servers_json "$repo_path")"

  if [[ ! -f "$repo_path/.claude/settings.json" || ! -f "$repo_path/.gemini/settings.json" ]]; then
    printf '0\n'
    return 0
  fi

  if jq -e \
      --argjson required "$expected_claude" '
        ($required | keys) as $keys
        | $keys | all(. as $k | (.mcpServers[$k] != null))
      ' "$repo_path/.claude/settings.json" >/dev/null 2>&1 && \
     jq -e \
      --argjson required "$expected_gemini" '
        ($required | keys) as $keys
        | $keys | all(. as $k | (.mcpServers[$k] != null))
      ' "$repo_path/.gemini/settings.json" >/dev/null 2>&1; then
    printf '1\n'
  else
    printf '0\n'
  fi
}

hg_parity_provider_hook_bridge_ok() {
  local repo_path="$1"
  local repo_name="$2"
  local unsupported translated
  unsupported="$(hg_parity_gemini_unsupported_hook_rule_count "$repo_path")"
  translated="$(hg_parity_gemini_translated_hook_rule_count "$repo_path")"

  if ! hg_parity_repo_has_claude_hooks "$repo_path"; then
    printf '1\n'
    return 0
  fi

  if [[ ! -f "$repo_path/.gemini/settings.json" ]]; then
    printf '0\n'
    return 0
  fi

  if [[ "$unsupported" -eq 0 && "$translated" -gt 0 ]]; then
    printf '1\n'
  else
    printf '0\n'
  fi
}

hg_parity_provider_drift_count() {
  local repo_path="$1"
  local repo_name="$2"
  local count=0
  local expected

  expected="$(hg_parity_render_claude_settings "$repo_path")"
  hg_parity_compare_expected_file "$expected" "$repo_path/.claude/settings.json" || count=$((count + 1))

  if [[ -x "$HG_STUDIO_ROOT/dotfiles/scripts/hg-gemini-settings-sync.sh" ]]; then
    if ! "$HG_STUDIO_ROOT/dotfiles/scripts/hg-gemini-settings-sync.sh" "$repo_path" --check >/dev/null 2>&1; then
      count=$((count + 1))
    fi
  else
    expected="$(hg_parity_render_gemini_settings "$repo_path")"
    hg_parity_compare_expected_file "$expected" "$repo_path/.gemini/settings.json" || count=$((count + 1))
  fi

  if hg_parity_repo_requires_gemini_extension "$repo_path" "$repo_name"; then
    expected="$(hg_parity_render_gemini_extension "$repo_path" "$repo_name")"
    hg_parity_compare_expected_file "$expected" "$repo_path/$(hg_parity_gemini_extension_relpath "$repo_name")" || count=$((count + 1))
  fi

  printf '%s\n' "$count"
}
