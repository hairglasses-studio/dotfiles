#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

JSON_MODE=0
REQUIRE_HEAVY=0

usage() {
  cat <<'EOF'
Usage: hg-ollama-readiness.sh [--json] [--require-heavy]

Options:
  --json           Emit a machine-readable readiness report.
  --require-heavy  Require the heavy coding lane to be installed and ready.
  -h, --help       Show this help text.
EOF
}

json_array_from_args() {
  if [[ "$#" -eq 0 ]]; then
    printf '[]\n'
    return 0
  fi
  jq -cn '$ARGS.positional' --args "$@"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json)
      JSON_MODE=1
      shift
      ;;
    --require-heavy)
      REQUIRE_HEAVY=1
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

hg_require curl jq ollama
hg_local_llm_export_env

version_url="${OLLAMA_BASE_URL%/}/api/version"
tags_url="${OLLAMA_BASE_URL%/}/api/tags"
checked_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

ready=1
reachable=0
version="unknown"
error_detail=""
tags_json='{"models":[]}'

if version_json="$(curl -fsS "$version_url" 2>&1)"; then
  reachable=1
  version="$(jq -r '.version // "unknown"' <<<"$version_json" 2>/dev/null || printf 'unknown')"
else
  ready=0
  error_detail="$version_json"
fi

if (( reachable == 1 )); then
  if ! tags_json="$(curl -fsS "$tags_url" 2>&1)"; then
    ready=0
    error_detail="$tags_json"
    tags_json='{"models":[]}'
  fi
fi

mapfile -t required_models < <(hg_local_llm_smoke_models)
if (( REQUIRE_HEAVY == 1 )); then
  required_models+=("$OLLAMA_HEAVY_CODE_MODEL")
fi
mapfile -t required_models < <(printf '%s\n' "${required_models[@]}" | awk 'length($0) > 0 && !seen[$0]++')

declare -a ready_models=()
declare -a missing_models=()
declare -a alias_checks=()

if (( reachable == 1 )) && [[ -n "$tags_json" ]]; then
  for model in "${required_models[@]}"; do
    if hg_local_llm_model_installed_json "$tags_json" "$model"; then
      ready_models+=("$model")
    else
      missing_models+=("$model")
      ready=0
    fi
  done

  while IFS='=' read -r alias_name source_model; do
    local_status="skipped"
    detail=""
    if [[ -z "$alias_name" || -z "$source_model" ]]; then
      continue
    fi

    if hg_local_llm_model_installed_json "$tags_json" "$source_model"; then
      if ! hg_local_llm_model_installed_json "$tags_json" "$alias_name"; then
        local_status="missing_alias"
        detail="managed alias is missing while backing model is installed"
        ready=0
      else
        alias_digest="$(hg_local_llm_model_digest_json "$tags_json" "$alias_name")"
        source_digest="$(hg_local_llm_model_digest_json "$tags_json" "$source_model")"
        if [[ -n "$alias_digest" && -n "$source_digest" && "$alias_digest" == "$source_digest" ]]; then
          local_status="ok"
        else
          local_status="digest_mismatch"
          detail="managed alias digest does not match backing model"
          ready=0
        fi
      fi
    elif hg_local_llm_model_installed_json "$tags_json" "$alias_name"; then
      local_status="alias_only"
      detail="alias is installed but backing model is absent"
    fi

    alias_checks+=("$(jq -cn \
      --arg alias "$alias_name" \
      --arg source "$source_model" \
      --arg status "$local_status" \
      --arg detail "$detail" \
      '{alias: $alias, source: $source, status: $status, detail: $detail}')")
  done < <(hg_local_llm_code_alias_pairs)
fi

pull_commands_json="$(while IFS= read -r cmd; do printf '%s\n' "$cmd"; done < <(hg_local_llm_recommended_pull_commands) | jq -Rsc 'split("\n") | map(select(length > 0))')"
required_models_json="$(json_array_from_args "${required_models[@]}")"
ready_models_json="$(json_array_from_args "${ready_models[@]}")"
missing_models_json="$(json_array_from_args "${missing_models[@]}")"
if [[ "${#alias_checks[@]}" -eq 0 ]]; then
  alias_checks_json='[]'
else
  alias_checks_json="$(printf '%s\n' "${alias_checks[@]}" | jq -s '.')"
fi

if (( JSON_MODE == 1 )); then
  jq -n \
    --arg checked_at "$checked_at" \
    --arg base_url "$OLLAMA_BASE_URL" \
    --arg version "$version" \
    --arg error "$error_detail" \
    --argjson reachable "$reachable" \
    --argjson ready "$ready" \
    --argjson require_heavy "$REQUIRE_HEAVY" \
    --argjson required_models "$required_models_json" \
    --argjson ready_models "$ready_models_json" \
    --argjson missing_models "$missing_models_json" \
    --argjson alias_checks "$alias_checks_json" \
    --argjson pull_commands "$pull_commands_json" \
    '{
      checked_at: $checked_at,
      base_url: $base_url,
      version: $version,
      reachable: ($reachable == 1),
      ready: ($ready == 1),
      require_heavy: ($require_heavy == 1),
      required_models: $required_models,
      ready_models: $ready_models,
      missing_models: $missing_models,
      alias_checks: $alias_checks,
      pull_commands: $pull_commands,
      error: (if $error == "" then null else $error end)
    }'
  exit $(( ready == 1 ? 0 : 1 ))
fi

hg_info "Checking Ollama readiness at $OLLAMA_BASE_URL"
if (( reachable == 1 )); then
  hg_ok "Ollama version: $version"
else
  hg_warn "Ollama endpoint is not reachable: $error_detail"
fi

for model in "${ready_models[@]}"; do
  hg_ok "Ready model: $model"
done
for model in "${missing_models[@]}"; do
  hg_warn "Missing required model: $model"
done

if [[ "${#alias_checks[@]}" -gt 0 ]]; then
  while IFS= read -r check; do
    status="$(jq -r '.status' <<<"$check")"
    alias_name="$(jq -r '.alias' <<<"$check")"
    source_model="$(jq -r '.source' <<<"$check")"
    detail="$(jq -r '.detail // ""' <<<"$check")"
    case "$status" in
      ok)
        hg_ok "Managed alias: $alias_name -> $source_model"
        ;;
      missing_alias|digest_mismatch)
        hg_warn "Managed alias problem: $alias_name -> $source_model (${detail})"
        ;;
      alias_only)
        hg_warn "Alias-only install: $alias_name -> $source_model (${detail})"
        ;;
    esac
  done < <(printf '%s\n' "${alias_checks[@]}")
fi

if (( ready == 1 )); then
  hg_ok "Ollama readiness passed"
else
  hg_warn "Ollama readiness failed"
  if [[ -n "$error_detail" ]]; then
    hg_warn "$error_detail"
  fi
  hg_info "Recommended recovery commands:"
  while IFS= read -r cmd; do
    hg_ok "$cmd"
  done < <(hg_local_llm_recommended_pull_commands)
fi

exit $(( ready == 1 ? 0 : 1 ))
