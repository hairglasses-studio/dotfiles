#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

REQUIRE_HEAVY=0
DEEP_OFFLOAD_MEMORY_GB="${HG_OLLAMA_DEEP_OFFLOAD_MEMORY_GB:-80}"

usage() {
  cat <<'EOF'
Usage: hg-ollama-full-test.sh [--require-heavy]

Options:
  --require-heavy  Fail if the heavy coding lane cannot be validated.
  -h, --help       Show this help text.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
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

hg_require curl jq ollama systemctl awk free
hg_local_llm_export_env

version_url="${OLLAMA_BASE_URL%/}/api/version"
tags_url="${OLLAMA_BASE_URL%/}/api/tags"
ps_url="${OLLAMA_BASE_URL%/}/api/ps"
generate_url="${OLLAMA_BASE_URL%/}/api/generate"
embed_url="${OLLAMA_BASE_URL%/}/api/embed"
anthropic_messages_url="$(hg_local_llm_v1_base_url)/messages"
responses_url="$(hg_local_llm_v1_base_url)/responses"

require_service_env() {
  local env_text="$1"
  local key="$2"
  local value="$3"

  if [[ "$env_text" != *"$key=$value"* ]]; then
    hg_die "systemd env missing $key=$value"
  fi
  hg_ok "systemd env: $key=$value"
}

model_installed() {
  local tags_json="$1"
  local model="$2"
  jq -e --arg model "$model" '.models[]? | select(.name == $model or .name == ($model + ":latest"))' <<<"$tags_json" >/dev/null
}

model_digest() {
  local tags_json="$1"
  local model="$2"
  jq -r --arg model "$model" '.models[]? | select(.name == $model or .name == ($model + ":latest")) | .digest' <<<"$tags_json" | head -n1
}

available_memory_gb() {
  free -g | awk '/^Mem:/ {print $7}'
}

run_generate_exact() {
  local model="$1"
  local expected="$2"
  local prompt="$3"
  local response_json response_text

  response_json="$(curl -fsS "$generate_url" \
    -H 'Content-Type: application/json' \
    -d "$(jq -cn \
      --arg model "$model" \
      --arg prompt "$prompt" \
      --arg keep_alive "$OLLAMA_KEEP_ALIVE" \
      '{model: $model, prompt: $prompt, stream: false, keep_alive: $keep_alive, options: {temperature: 0}}')")"

  response_text="$(jq -r '.response // empty' <<<"$response_json" | tr -d '\r' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  [[ "$response_text" == *"$expected"* ]] || hg_die "Generate test for $model did not contain expected token $expected"
  hg_ok "Generate test passed for $model"
}

run_anthropic_messages_exact() {
  local model="$1"
  local expected="$2"
  local prompt="$3"
  local response_json response_text

  response_json="$(curl -fsS "$anthropic_messages_url" \
    -H 'Content-Type: application/json' \
    -H "x-api-key: $OLLAMA_API_KEY" \
    -H 'anthropic-version: 2023-06-01' \
    -d "$(jq -cn \
      --arg model "$model" \
      --arg prompt "$prompt" \
      '{model: $model, max_tokens: 64, messages: [{role: "user", content: $prompt}]}')")"

  response_text="$(jq -r '[.content[]? | .text // empty] | join(" ")' <<<"$response_json" | tr -d '\r' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  [[ "$response_text" == *"$expected"* ]] || hg_die "Anthropic compatibility test for $model did not contain expected token $expected"
  hg_ok "Anthropic compatibility test passed for $model"
}

run_responses_exact() {
  local model="$1"
  local expected="$2"
  local prompt="$3"
  local response_json response_text

  response_json="$(curl -fsS "$responses_url" \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $OLLAMA_API_KEY" \
    -d "$(jq -cn \
      --arg model "$model" \
      --arg prompt "$prompt" \
      '{model: $model, input: $prompt}')")"

  response_text="$(jq -r '(.output_text // .output[0].content[0].text // empty)' <<<"$response_json" | tr -d '\r' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  [[ "$response_text" == *"$expected"* ]] || hg_die "Responses compatibility test for $model did not contain expected token $expected"
  hg_ok "Responses compatibility test passed for $model"
}

systemctl is-active --quiet ollama || hg_die "ollama service is not active"
hg_ok "ollama service is active"

service_env="$(systemctl show ollama -p Environment --value)"
require_service_env "$service_env" "OLLAMA_FLASH_ATTENTION" "1"
require_service_env "$service_env" "OLLAMA_KV_CACHE_TYPE" "q8_0"
require_service_env "$service_env" "OLLAMA_MAX_LOADED_MODELS" "1"
require_service_env "$service_env" "OLLAMA_NUM_PARALLEL" "1"
require_service_env "$service_env" "OLLAMA_KEEP_ALIVE" "15m"

version_json="$(curl -fsS "$version_url")"
hg_ok "Ollama version: $(jq -r '.version // "unknown"' <<<"$version_json")"

tags_json="$(curl -fsS "$tags_url")"
mapfile -t required_models < <(hg_local_llm_smoke_models)
for model in "${required_models[@]}"; do
  if model_installed "$tags_json" "$model"; then
    hg_ok "Installed baseline model: $model"
  else
    hg_die "Missing baseline model: $model"
  fi
done

while IFS='=' read -r alias_name source_model; do
  [[ -n "$alias_name" && -n "$source_model" ]] || continue

  if ! model_installed "$tags_json" "$source_model"; then
    hg_info "Skipping alias check for $alias_name; source model $source_model is not installed"
    continue
  fi

  if ! model_installed "$tags_json" "$alias_name"; then
    hg_die "Missing managed alias $alias_name for source model $source_model"
  fi

  alias_digest="$(model_digest "$tags_json" "$alias_name")"
  source_digest="$(model_digest "$tags_json" "$source_model")"
  [[ -n "$alias_digest" && -n "$source_digest" ]] || hg_die "Could not resolve digest for alias $alias_name or source model $source_model"
  [[ "$alias_digest" == "$source_digest" ]] || hg_die "Managed alias $alias_name does not match source model $source_model"
  hg_ok "Managed alias: $alias_name -> $source_model"
done < <(hg_local_llm_code_alias_pairs)

run_generate_exact "$OLLAMA_CHAT_MODEL" "CHAT_OK" "Reply with exactly CHAT_OK."
run_generate_exact "$OLLAMA_FAST_MODEL" "FAST_CODE_OK" "Reply with exactly FAST_CODE_OK."
run_generate_exact "$OLLAMA_CODE_MODEL" "CODE_OK" "Reply with exactly CODE_OK."
run_anthropic_messages_exact "$OLLAMA_CHAT_MODEL" "ANTHROPIC_OK" "Reply with exactly ANTHROPIC_OK."
run_responses_exact "$OLLAMA_FAST_MODEL" "RESPONSES_OK" "Reply with exactly RESPONSES_OK."

if model_installed "$tags_json" "$OLLAMA_HIGH_CONTEXT_CODE_MODEL"; then
  run_generate_exact "$OLLAMA_HIGH_CONTEXT_CODE_MODEL" "HIGH_CONTEXT_CODE_OK" "Reply with exactly HIGH_CONTEXT_CODE_OK."
else
  hg_info "High-context lane: $OLLAMA_HIGH_CONTEXT_CODE_MODEL (not-installed)"
fi

embed_json="$(curl -fsS "$embed_url" \
  -H 'Content-Type: application/json' \
  -d "$(jq -cn --arg model "$OLLAMA_EMBED_MODEL" --arg input 'full ollama functionality test' '{model: $model, input: $input}')")"
embed_count="$(jq -r '(.embeddings // []) | length' <<<"$embed_json")"
[[ "$embed_count" -gt 0 ]] || hg_die "Embed test returned no embeddings"
hg_ok "Embed test passed for $OLLAMA_EMBED_MODEL"

ps_json="$(curl -fsS "$ps_url")"
loaded_count="$(jq -r '(.models // []) | length' <<<"$ps_json")"
[[ "$loaded_count" -gt 0 ]] || hg_die "No running models were retained after generate tests"
hg_ok "Running models retained after requests: $loaded_count"

if ! jq -e \
  --arg chat_model "$OLLAMA_CHAT_MODEL" \
  --arg fast_model "$OLLAMA_FAST_MODEL" \
  --arg code_model "$OLLAMA_CODE_MODEL" \
  --arg embed_model "$OLLAMA_EMBED_MODEL" \
  '.models[]?
    | select(
        .name == $chat_model or .name == ($chat_model + ":latest") or
        .name == $fast_model or .name == ($fast_model + ":latest") or
        .name == $code_model or .name == ($code_model + ":latest") or
        .name == $embed_model or .name == ($embed_model + ":latest")
      )' <<<"$ps_json" >/dev/null; then
  hg_die "Expected a resident baseline model in /api/ps after requests"
fi

if ! jq -e '.models[]? | select((.context_length // 0) > 0 and (.expires_at // "") != "")' <<<"$ps_json" >/dev/null; then
  hg_die "Running models did not report context_length and expires_at"
fi
hg_ok "Running models expose context length and residency metadata"

sleep 2
ps_after_json="$(curl -fsS "$ps_url")"
loaded_after_count="$(jq -r '(.models // []) | length' <<<"$ps_after_json")"
[[ "$loaded_after_count" -gt 0 ]] || hg_die "Resident models were unloaded immediately despite keep-alive"
hg_ok "Resident models remained loaded after keep-alive check"

if model_installed "$tags_json" "$OLLAMA_HEAVY_CODE_MODEL"; then
  mem_gb="$(available_memory_gb)"
  if (( mem_gb >= DEEP_OFFLOAD_MEMORY_GB )); then
    run_generate_exact "$OLLAMA_HEAVY_CODE_MODEL" "HEAVY_CODE_OK" "Reply with exactly HEAVY_CODE_OK."
  elif (( REQUIRE_HEAVY == 1 )); then
    hg_die "Heavy model installed but available memory ${mem_gb}GiB is below ${DEEP_OFFLOAD_MEMORY_GB}GiB"
  else
    hg_warn "Skipping heavy-code generate for $OLLAMA_HEAVY_CODE_MODEL; available memory ${mem_gb}GiB is below ${DEEP_OFFLOAD_MEMORY_GB}GiB"
  fi
elif (( REQUIRE_HEAVY == 1 )); then
  hg_die "Heavy model $OLLAMA_HEAVY_CODE_MODEL is not installed"
else
  hg_warn "Skipping heavy-code generate; $OLLAMA_HEAVY_CODE_MODEL is not installed"
fi

hg_info "Current ollama ps:"
ollama ps
