#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

WAIT_INTERVAL_SECS="${HG_OLLAMA_HEAVY_WAIT_INTERVAL_SECS:-60}"
generate_url=""

usage() {
  cat <<'EOF'
Usage: hg-ollama-verify-heavy.sh

Waits for the configured heavy coding model to be installed, syncs managed
aliases, runs a direct heavy generate probe, then executes the strict full test
path with heavy mode required.
EOF
}

model_installed() {
  local model="$1"
  ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "$model" >/dev/null 2>&1 \
    || ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "${model}:latest" >/dev/null 2>&1
}

run_heavy_probe() {
  local response_json response_text

  response_json="$(curl -fsS "$generate_url" \
    -H 'Content-Type: application/json' \
    -d "$(jq -cn \
      --arg model "$OLLAMA_HEAVY_CODE_MODEL" \
      --arg prompt 'Reply with exactly HEAVY_CODE_OK.' \
      --arg keep_alive "$OLLAMA_KEEP_ALIVE" \
      '{model: $model, prompt: $prompt, stream: false, keep_alive: $keep_alive, options: {temperature: 0}}')")"

  response_text="$(jq -r '.response // empty' <<<"$response_json" | tr -d '\r' | sed 's/^[[:space:]]*//; s/[[:space:]]*$//')"
  [[ "$response_text" == *"HEAVY_CODE_OK"* ]] || hg_die "Heavy probe for $OLLAMA_HEAVY_CODE_MODEL did not return HEAVY_CODE_OK"
  hg_ok "Heavy probe passed for $OLLAMA_HEAVY_CODE_MODEL"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

hg_require curl jq free ollama awk grep sed
hg_local_llm_export_env
generate_url="${OLLAMA_BASE_URL%/}/api/generate"

while ! model_installed "$OLLAMA_HEAVY_CODE_MODEL"; do
  hg_info "Waiting for heavy model $OLLAMA_HEAVY_CODE_MODEL to finish installing"
  sleep "$WAIT_INTERVAL_SECS"
done

hg_ok "Heavy model installed: $OLLAMA_HEAVY_CODE_MODEL"
free -h

"$SCRIPT_DIR/hg-ollama-sync-aliases.sh"
run_heavy_probe

current_threshold="${HG_OLLAMA_DEEP_OFFLOAD_MEMORY_GB:-80}"
available_mem_gb="$(free -g | awk '/^Mem:/ {print $7}')"
if (( available_mem_gb < current_threshold )); then
  hg_warn "Heavy probe succeeded with ${available_mem_gb}GiB available, below the current ${current_threshold}GiB preflight"
fi

HG_OLLAMA_DEEP_OFFLOAD_MEMORY_GB=0 "$SCRIPT_DIR/hg-ollama-full-test.sh" --require-heavy
