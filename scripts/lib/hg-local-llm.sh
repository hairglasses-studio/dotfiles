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

hg_local_llm_v1_base_url() {
  hg_local_llm_export_env
  printf '%s/v1\n' "${OLLAMA_BASE_URL%/}"
}

hg_local_llm_export_env() {
  local host_port host port

  export OLLAMA_BASE_URL="$(hg_local_llm_resolve_base_url)"
  export OLLAMA_FAST_SOURCE_MODEL="${OLLAMA_FAST_SOURCE_MODEL:-qwen2.5-coder:7b}"
  export OLLAMA_CODE_SOURCE_MODEL="${OLLAMA_CODE_SOURCE_MODEL:-devstral-small-2}"
  export OLLAMA_HEAVY_CODE_SOURCE_MODEL="${OLLAMA_HEAVY_CODE_SOURCE_MODEL:-devstral-2}"
  export OLLAMA_HIGH_CONTEXT_CODE_SOURCE_MODEL="${OLLAMA_HIGH_CONTEXT_CODE_SOURCE_MODEL:-qwen3-coder-next}"
  export OLLAMA_CHAT_MODEL="${OLLAMA_CHAT_MODEL:-code-primary}"
  export OLLAMA_FAST_MODEL="${OLLAMA_FAST_MODEL:-code-fast}"
  export OLLAMA_CODE_MODEL="${OLLAMA_CODE_MODEL:-code-primary}"
  export OLLAMA_HEAVY_CODE_MODEL="${OLLAMA_HEAVY_CODE_MODEL:-code-heavy}"
  export OLLAMA_HIGH_CONTEXT_CODE_MODEL="${OLLAMA_HIGH_CONTEXT_CODE_MODEL:-code-long}"
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

hg_local_llm_export_anthropic_compat_env() {
  hg_local_llm_export_env

  export ANTHROPIC_BASE_URL="$(hg_local_llm_v1_base_url)"
  export ANTHROPIC_AUTH_TOKEN="${ANTHROPIC_AUTH_TOKEN:-${ANTHROPIC_API_KEY:-$OLLAMA_API_KEY}}"
}

hg_local_llm_export_openai_compat_env() {
  hg_local_llm_export_env

  export OPENAI_BASE_URL="$(hg_local_llm_v1_base_url)"
  export OPENAI_API_KEY="${OPENAI_API_KEY:-$OLLAMA_API_KEY}"
}

hg_local_llm_env_keys() {
  cat <<'EOF'
OLLAMA_BASE_URL
OLLAMA_HOST
OLLAMA_PORT
OLLAMA_API_KEY
OLLAMA_KEEP_ALIVE
OLLAMA_CHAT_MODEL
OLLAMA_FAST_MODEL
OLLAMA_CODE_MODEL
OLLAMA_HEAVY_CODE_MODEL
OLLAMA_HIGH_CONTEXT_CODE_MODEL
OLLAMA_CLOUD_CODE_MODEL
OLLAMA_CLOUD_VERIFIED_CODE_MODEL
OLLAMA_MULTILINGUAL_CODE_MODEL
OLLAMA_THINKING_CODE_MODEL
OLLAMA_EMBED_MODEL
EOF
}

hg_local_llm_codex_env_instructions() {
  printf '%s\n' 'source "$HOME/hairglasses-studio/dotfiles/scripts/lib/hg-local-llm.sh" && hg_local_llm_export_env'
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
    "code-fast=$OLLAMA_FAST_SOURCE_MODEL" \
    "code-compact=$OLLAMA_FAST_SOURCE_MODEL" \
    "code-primary=$OLLAMA_CODE_SOURCE_MODEL" \
    "code-reasoner=$OLLAMA_CODE_SOURCE_MODEL" \
    "code-long=$OLLAMA_HIGH_CONTEXT_CODE_SOURCE_MODEL" \
    "code-heavy=$OLLAMA_HEAVY_CODE_SOURCE_MODEL"
}

hg_local_llm_alias_source_model() {
  hg_local_llm_export_env

  case "${1:-}" in
    code-fast|code-compact)
      printf '%s\n' "$OLLAMA_FAST_SOURCE_MODEL"
      ;;
    code-primary|code-reasoner)
      printf '%s\n' "$OLLAMA_CODE_SOURCE_MODEL"
      ;;
    code-long)
      printf '%s\n' "$OLLAMA_HIGH_CONTEXT_CODE_SOURCE_MODEL"
      ;;
    code-heavy)
      printf '%s\n' "$OLLAMA_HEAVY_CODE_SOURCE_MODEL"
      ;;
    *)
      return 1
      ;;
  esac
}

hg_local_llm_pullable_model() {
  local model="${1:-}"
  local source_model=""

  if source_model="$(hg_local_llm_alias_source_model "$model" 2>/dev/null)"; then
    printf '%s\n' "$source_model"
    return 0
  fi
  printf '%s\n' "$model"
}

hg_local_llm_recommended_pull_commands() {
  local model pull_target

  while IFS= read -r model; do
    [[ -n "$model" ]] || continue
    pull_target="$(hg_local_llm_pullable_model "$model")"
    [[ -n "$pull_target" ]] || continue
    printf 'ollama pull %s\n' "$pull_target"
  done < <(hg_local_llm_local_models) | awk 'length($0) > 0 && !seen[$0]++'
}

hg_local_llm_model_installed_json() {
  local tags_json="${1:-}"
  local model="${2:-}"

  jq -e --arg model "$model" '
    .models[]?
    | select(
        .name == $model or .name == ($model + ":latest") or
        .model == $model or .model == ($model + ":latest")
      )
  ' <<<"$tags_json" >/dev/null
}

hg_local_llm_model_digest_json() {
  local tags_json="${1:-}"
  local model="${2:-}"

  jq -r --arg model "$model" '
    .models[]?
    | select(
        .name == $model or .name == ($model + ":latest") or
        .model == $model or .model == ($model + ":latest")
      )
    | .digest
  ' <<<"$tags_json" | head -n1
}
