#!/usr/bin/env bash
# shellcheck shell=bash
# hg-agent-parity.sh — Shared helpers for Claude/Codex/Gemini parity sync and audit.

HG_AGENT_PARITY_LIB_DIR="${HG_AGENT_PARITY_LIB_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"

if ! command -v hg_require >/dev/null 2>&1; then
  source "$HG_AGENT_PARITY_LIB_DIR/hg-core.sh"
fi

HG_AGENT_PARITY_DOTFILES_ROOT="${HG_AGENT_PARITY_DOTFILES_ROOT:-${HG_DOTFILES:-$(cd "$HG_AGENT_PARITY_LIB_DIR/.." && pwd)}}"

hg_parity_require_tools() {
  hg_require jq diff
}

hg_parity_objectives_path() {
  if [[ -n "${HG_PARITY_OBJECTIVES_PATH:-}" ]]; then
    printf '%s\n' "$HG_PARITY_OBJECTIVES_PATH"
    return 0
  fi

  local canonical historical
  canonical="$HG_STUDIO_ROOT/docs/projects/agent-parity/parity-objectives.json"
  historical="$HG_STUDIO_ROOT/docs/projects/codex-migration/parity-objectives.json"

  if [[ -f "$canonical" ]]; then
    printf '%s\n' "$canonical"
  else
    printf '%s\n' "$historical"
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

hg_parity_active_mcp_servers_json() {
  local repo_path="$1"
  if [[ ! -f "$repo_path/.mcp.json" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -cS '
    (.mcpServers // {})
    | if type != "object" then {} else . end
    | with_entries(select(.key | startswith("_") | not))
    | with_entries(
        if (.value | type) == "object" then
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
        else
          .
        end
      )
  ' "$repo_path/.mcp.json"
}

hg_parity_repo_has_mcp_json() {
  local repo_path="$1"
  [[ -f "$repo_path/.mcp.json" ]] || return 1
  jq -e 'length > 0' <<<"$(hg_parity_active_mcp_servers_json "$repo_path")" >/dev/null 2>&1
}

hg_parity_kebab_case_server_name() {
  local name="$1"
  printf '%s' "$name" \
    | tr '[:upper:]' '[:lower:]' \
    | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-+/-/g'
}

hg_parity_claude_mcp_servers_json() {
  local repo_path="$1"
  hg_parity_active_mcp_servers_json "$repo_path"
}

hg_parity_gemini_mcp_servers_json() {
  local repo_path="$1"
  jq -c '
    .
    | with_entries(
        .key |= (
          ascii_downcase
          | gsub("[^a-z0-9]+"; "-")
          | gsub("^-+"; "")
          | gsub("-+$"; "")
          | gsub("-+"; "-")
        )
      )
  ' <<<"$(hg_parity_active_mcp_servers_json "$repo_path")"
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

hg_parity_gemini_sync_script() {
  printf '%s\n' "$HG_AGENT_PARITY_DOTFILES_ROOT/scripts/hg-gemini-settings-sync.sh"
}

hg_parity_generated_gemini_settings_count() {
  local repo_path="$1"
  if hg_parity_gemini_settings_current "$repo_path"; then
    printf '1\n'
  else
    printf '0\n'
  fi
}

hg_parity_gemini_mcp_server_count() {
  local repo_path="$1"
  jq -r 'keys | length' <<<"$(hg_parity_gemini_mcp_servers_json "$repo_path")"
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

hg_parity_source_hooks_json() {
  local repo_path="$1"
  local settings_path="$repo_path/.claude/settings.json"
  if [[ ! -f "$settings_path" ]]; then
    printf '{}\n'
    return 0
  fi

  jq -c '
    (.hooks // {}) as $hooks
    | {
        SessionStart: ($hooks.SessionStart // []),
        BeforeTool: (($hooks.BeforeTool // []) + ($hooks.PreToolUse // [])),
        AfterTool: (($hooks.AfterTool // []) + ($hooks.PostToolUse // [])),
        Notification: ($hooks.Notification // []),
        BeforeAgent: (($hooks.BeforeAgent // []) + ($hooks.UserPromptSubmit // []))
      }
    | with_entries(select((.value | type == "array") and (.value | length > 0)))
  ' "$settings_path"
}

hg_parity_supported_source_hook_rule_count() {
  local repo_path="$1"
  local settings_path="$repo_path/.claude/settings.json"
  if [[ ! -f "$settings_path" ]]; then
    printf '0\n'
    return 0
  fi

  jq -r '
    [
      (.hooks.SessionStart // []),
      (.hooks.BeforeTool // []),
      (.hooks.PreToolUse // []),
      (.hooks.AfterTool // []),
      (.hooks.PostToolUse // []),
      (.hooks.Notification // []),
      (.hooks.BeforeAgent // []),
      (.hooks.UserPromptSubmit // [])
    ]
    | flatten
    | length
  ' "$settings_path"
}

hg_parity_unsupported_source_hook_rule_count() {
  local repo_path="$1"
  local settings_path="$repo_path/.claude/settings.json"
  if [[ ! -f "$settings_path" ]]; then
    printf '0\n'
    return 0
  fi

  jq -r '
    [
      (.hooks.Stop // []),
      (.hooks.PostToolUseFailure // []),
      (.hooks.PostCompact // []),
      (.hooks.PreCompact // []),
      (.hooks.SubagentStart // []),
      (.hooks.SubagentStop // []),
      (.hooks.SessionEnd // [])
    ]
    | flatten
    | length
  ' "$settings_path"
}

hg_parity_repo_requires_gemini_extension() {
  local repo_path="$1"
  local repo_name="$2"
  [[ "$(hg_parity_repo_objective_bool "$repo_name" "gemini_extension_scaffold" "false")" == "true" ]]
}

hg_parity_render_claude_settings() {
  local repo_path="$1"
  local current_json generated_mcp has_mcp_file
  current_json="$(hg_parity_object_json "$repo_path/.claude/settings.json")"
  generated_mcp="$(hg_parity_claude_mcp_servers_json "$repo_path")"
  if [[ -f "$repo_path/.mcp.json" ]]; then
    has_mcp_file=true
  else
    has_mcp_file=false
  fi

  jq -S -n \
    --argjson current "$current_json" \
    --argjson generated "$generated_mcp" \
    --argjson has_mcp "$has_mcp_file" '
      ($current | if type == "object" then . else {} end) as $cfg
      | if $has_mcp then
          ($cfg + {mcpServers: $generated})
        elif ($generated | length) > 0 then
          $cfg + {mcpServers: (($cfg.mcpServers // {}) + $generated)}
        else
          $cfg
        end
    '
}

hg_parity_render_gemini_settings() {
  local repo_path="$1"
  local template_path generated_mcp normalized_hooks
  template_path="$HG_AGENT_PARITY_DOTFILES_ROOT/templates/gemini-settings.standard.json"
  generated_mcp="$(hg_parity_gemini_mcp_servers_json "$repo_path")"
  normalized_hooks="$(hg_parity_source_hooks_json "$repo_path")"

  jq -S -n \
    --slurpfile template "$template_path" \
    --argjson generated "$generated_mcp" \
    --argjson hooks "$normalized_hooks" '
      ($template[0] | if type == "object" then . else {} end) as $base
      | (
          if (($base.context.fileName // null) | type) == "array" then
            ($base.context.fileName | unique | sort) as $file_names
            | ($base + {context: (($base.context // {}) + {fileName: $file_names})})
          else
            $base
          end
        ) as $normalized
      | ($normalized + {mcpServers: $generated}) as $with_servers
      | (
          if ($generated | length) > 0 then
            $with_servers + {
              mcp: (($with_servers.mcp // {}) + {
                allowed: (($generated | keys) | sort)
              })
            }
          else
            ($with_servers | del(.mcp))
          end
        )
      | if ($hooks | length) > 0 then
          . + {hooks: $hooks}
        else
          del(.hooks)
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

hg_parity_gemini_settings_current() {
  local repo_path="$1"
  local sync_script
  sync_script="$(hg_parity_gemini_sync_script)"

  if [[ -f "$sync_script" ]]; then
    bash "$sync_script" "$repo_path" --check >/dev/null 2>&1
    return $?
  fi

  local expected
  expected="$(hg_parity_render_gemini_settings "$repo_path")"
  hg_parity_compare_expected_file "$expected" "$repo_path/.gemini/settings.json"
}

hg_parity_gemini_settings_sync() {
  local repo_path="$1"
  local sync_script
  sync_script="$(hg_parity_gemini_sync_script)"
  [[ -f "$sync_script" ]] || return 1
  bash "$sync_script" "$repo_path" >/dev/null 2>&1
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
        . as $cfg
        | ($required | keys) as $keys
        | $keys | all(. as $k | ($cfg.mcpServers[$k] != null))
      ' "$repo_path/.claude/settings.json" >/dev/null 2>&1 && \
     jq -e \
      --argjson required "$expected_gemini" '
        . as $cfg
        | ($required | keys) as $keys
        | $keys | all(. as $k | ($cfg.mcpServers[$k] != null))
      ' "$repo_path/.gemini/settings.json" >/dev/null 2>&1; then
    printf '1\n'
  else
    printf '0\n'
  fi
}

hg_parity_provider_hook_bridge_ok() {
  local repo_path="$1"
  local repo_name="$2"
  local unsupported translated source_hooks gemini_hooks
  unsupported="$(hg_parity_unsupported_source_hook_rule_count "$repo_path")"
  translated="$(hg_parity_supported_source_hook_rule_count "$repo_path")"
  source_hooks="$(hg_parity_source_hooks_json "$repo_path")"

  if ! hg_parity_repo_has_claude_hooks "$repo_path"; then
    printf '1\n'
    return 0
  fi

  if [[ ! -f "$repo_path/.gemini/settings.json" ]]; then
    printf '0\n'
    return 0
  fi

  gemini_hooks="$(jq -cS '(.hooks // {}) | if type == "object" then . else {} end' "$repo_path/.gemini/settings.json")"

  if [[ "$unsupported" -ne 0 ]]; then
    printf '0\n'
    return 0
  fi

  if [[ "$translated" -eq 0 ]]; then
    if [[ "$gemini_hooks" == "{}" ]]; then
      printf '1\n'
    else
      printf '0\n'
    fi
    return 0
  fi

  if [[ "$source_hooks" == "$gemini_hooks" ]]; then
    printf '1\n'
  else
    printf '0\n'
  fi
}

hg_parity_provider_drift_count() {
  local repo_path="$1"
  local count=0
  local expected

  expected="$(hg_parity_render_claude_settings "$repo_path")"
  hg_parity_compare_expected_file "$expected" "$repo_path/.claude/settings.json" || count=$((count + 1))

  hg_parity_gemini_settings_current "$repo_path" || count=$((count + 1))

  printf '%s\n' "$count"
}
