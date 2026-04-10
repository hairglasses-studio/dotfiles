#!/usr/bin/env bash
# hg-local-llm.sh — Shared local Ollama defaults for workspace scripts and launchers.

hg_local_llm_resolve_base_url() {
  local base_url="${OLLAMA_BASE_URL:-}"
  local host="${OLLAMA_HOST:-}"
  local port="${OLLAMA_PORT:-11434}"

  if [[ -n "$base_url" ]]; then
    printf '%s\n' "${base_url%/}"
    return 0
  fi

  if [[ -z "$host" ]]; then
    printf 'http://127.0.0.1:11434\n'
    return 0
  fi

  if [[ "$host" =~ ^https?:// ]]; then
    printf '%s\n' "${host%/}"
    return 0
  fi

  printf 'http://%s:%s\n' "$host" "$port"
}

hg_local_llm_export_env() {
  local host_port host port

  export OLLAMA_BASE_URL="$(hg_local_llm_resolve_base_url)"
  export OLLAMA_CHAT_MODEL="${OLLAMA_CHAT_MODEL:-qwen3:8b}"
  export OLLAMA_FAST_MODEL="${OLLAMA_FAST_MODEL:-qwen2.5-coder:7b}"
  export OLLAMA_CODE_MODEL="${OLLAMA_CODE_MODEL:-devstral-small-2}"
  export OLLAMA_HEAVY_CODE_MODEL="${OLLAMA_HEAVY_CODE_MODEL:-devstral-2}"
  export OLLAMA_HIGH_CONTEXT_CODE_MODEL="${OLLAMA_HIGH_CONTEXT_CODE_MODEL:-qwen3-coder-next}"
  export OLLAMA_CLOUD_CODE_MODEL="${OLLAMA_CLOUD_CODE_MODEL:-glm-5.1:cloud}"
  export OLLAMA_CLOUD_VERIFIED_CODE_MODEL="${OLLAMA_CLOUD_VERIFIED_CODE_MODEL:-glm-5:cloud}"
  export OLLAMA_MULTILINGUAL_CODE_MODEL="${OLLAMA_MULTILINGUAL_CODE_MODEL:-minimax-m2.1:cloud}"
  export OLLAMA_THINKING_CODE_MODEL="${OLLAMA_THINKING_CODE_MODEL:-kimi-k2-thinking:cloud}"
  export OLLAMA_EMBED_MODEL="${OLLAMA_EMBED_MODEL:-nomic-embed-text:v1.5}"
  export OLLAMA_API_KEY="${OLLAMA_API_KEY:-ollama}"
  export OLLAMA_KEEP_ALIVE="${OLLAMA_KEEP_ALIVE:-15m}"

  host_port="${OLLAMA_BASE_URL#*://}"
  host_port="${host_port%%/*}"
  host="${host_port%:*}"
  port="${host_port##*:}"

  if [[ "$host_port" != *:* ]]; then
    host="$host_port"
    port="11434"
  fi

  export OLLAMA_HOST="${OLLAMA_HOST:-$host}"
  export OLLAMA_PORT="${OLLAMA_PORT:-$port}"
}

hg_local_llm_smoke_models() {
  hg_local_llm_export_env

  printf '%s\n' \
    "$OLLAMA_CHAT_MODEL" \
    "$OLLAMA_FAST_MODEL" \
    "$OLLAMA_CODE_MODEL" \
    "$OLLAMA_EMBED_MODEL" \
    | awk 'length($0) > 0 && !seen[$0]++'
}

hg_local_llm_local_models() {
  hg_local_llm_export_env

  printf '%s\n' \
    "$OLLAMA_CHAT_MODEL" \
    "$OLLAMA_FAST_MODEL" \
    "$OLLAMA_CODE_MODEL" \
    "$OLLAMA_HEAVY_CODE_MODEL" \
    "$OLLAMA_HIGH_CONTEXT_CODE_MODEL" \
    "$OLLAMA_EMBED_MODEL" \
    | awk 'length($0) > 0 && !seen[$0]++'
}

hg_local_llm_cloud_models() {
  hg_local_llm_export_env

  printf '%s\n' \
    "$OLLAMA_CLOUD_CODE_MODEL" \
    "$OLLAMA_CLOUD_VERIFIED_CODE_MODEL" \
    "$OLLAMA_MULTILINGUAL_CODE_MODEL" \
    "$OLLAMA_THINKING_CODE_MODEL" \
    | awk 'length($0) > 0 && !seen[$0]++'
}

hg_local_llm_code_alias_pairs() {
  hg_local_llm_export_env

  printf '%s\n' \
    "code-fast=$OLLAMA_FAST_MODEL" \
    "code-compact=$OLLAMA_FAST_MODEL" \
    "code-primary=$OLLAMA_CODE_MODEL" \
    "code-reasoner=$OLLAMA_CODE_MODEL" \
    "code-long=$OLLAMA_HIGH_CONTEXT_CODE_MODEL" \
    "code-heavy=$OLLAMA_HEAVY_CODE_MODEL"
}
