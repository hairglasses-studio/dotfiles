#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

hg_require curl jq ollama
hg_local_llm_export_env

version_url="${OLLAMA_BASE_URL%/}/api/version"
tags_url="${OLLAMA_BASE_URL%/}/api/tags"
generate_url="${OLLAMA_BASE_URL%/}/api/generate"
embed_url="${OLLAMA_BASE_URL%/}/api/embed"

mapfile -t required_models < <(hg_local_llm_smoke_models)
mapfile -t cloud_aliases < <(hg_local_llm_cloud_models)

hg_info "Checking Ollama at $OLLAMA_BASE_URL"
version_json="$(curl -fsS "$version_url")"
hg_ok "Ollama version: $(jq -r '.version // "unknown"' <<<"$version_json")"

tags_json="$(curl -fsS "$tags_url")"
for model in "${required_models[@]}"; do
  if hg_local_llm_model_installed_json "$tags_json" "$model"; then
    hg_ok "Installed baseline model: $model"
  else
    hg_die "Missing baseline model: $model"
  fi
done

generate_json="$(curl -fsS "$generate_url" \
  -H 'Content-Type: application/json' \
  -d "$(jq -cn --arg model "$OLLAMA_CHAT_MODEL" --arg prompt 'Reply with OK.' --arg keep_alive "$OLLAMA_KEEP_ALIVE" '{model: $model, prompt: $prompt, stream: false, keep_alive: $keep_alive}')")"
generate_text="$(jq -r '.response // empty' <<<"$generate_json")"
[[ -n "$generate_text" ]] || hg_die "Generate smoke returned an empty response"
hg_ok "Generate smoke succeeded with $OLLAMA_CHAT_MODEL"

embed_json="$(curl -fsS "$embed_url" \
  -H 'Content-Type: application/json' \
  -d "$(jq -cn --arg model "$OLLAMA_EMBED_MODEL" --arg input 'local llm smoke test' '{model: $model, input: $input}')")"
embed_count="$(jq -r '(.embeddings // []) | length' <<<"$embed_json")"
[[ "$embed_count" -gt 0 ]] || hg_die "Embed smoke returned no embeddings"
hg_ok "Embed smoke succeeded with $OLLAMA_EMBED_MODEL"

hg_ok "Keep-alive profile: $OLLAMA_KEEP_ALIVE"
hg_info "Configured cloud aliases:"
for model in "${cloud_aliases[@]}"; do
  hg_ok "Cloud alias: $model"
done

hg_info "Installed models from ollama list:"
ollama list
