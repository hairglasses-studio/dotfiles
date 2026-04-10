#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

WAIT_INTERVAL_SECS="${HG_OLLAMA_HEAVY_WAIT_INTERVAL_SECS:-60}"
FORCE_HEAVY_PROBE="${HG_OLLAMA_FORCE_HEAVY_PROBE:-0}"
HEAVY_ARTIFACT_DIR="${HG_OLLAMA_HEAVY_ARTIFACT_DIR:-/tmp}"
FREE_ARTIFACT="$HEAVY_ARTIFACT_DIR/ollama-heavy-free.txt"
PROBE_ARTIFACT="$HEAVY_ARTIFACT_DIR/ollama-heavy-probe.json"
PROBE_ERROR_ARTIFACT="$HEAVY_ARTIFACT_DIR/ollama-heavy-probe.stderr.log"
FULL_TEST_ARTIFACT="$HEAVY_ARTIFACT_DIR/ollama-heavy-full-test.log"
JOURNAL_ARTIFACT="$HEAVY_ARTIFACT_DIR/ollama-heavy-journal.log"
generate_url=""

usage() {
  cat <<'EOF'
Usage: hg-ollama-verify-heavy.sh

Waits for the configured heavy coding model to be installed, syncs managed
aliases, runs a direct heavy generate probe, then executes the strict full test
path with heavy mode required.

Environment:
  HG_OLLAMA_FORCE_HEAVY_PROBE=1   Force the heavy probe even when available
                                  memory is below the heavy preflight.
  HG_OLLAMA_HEAVY_ARTIFACT_DIR    Directory for probe/test artifacts.
EOF
}

model_installed() {
  local model="$1"
  ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "$model" >/dev/null 2>&1 \
    || ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "${model}:latest" >/dev/null 2>&1
}

run_heavy_probe() {
  local response_json response_text probe_started_at rc
  probe_started_at="$(date '+%Y-%m-%d %H:%M:%S')"

  if ! response_json="$(curl -fsS "$generate_url" \
    -H 'Content-Type: application/json' \
    -d "$(jq -cn \
      --arg model "$OLLAMA_HEAVY_CODE_MODEL" \
      --arg prompt 'Reply with exactly HEAVY_CODE_OK.' \
      --arg keep_alive "$OLLAMA_KEEP_ALIVE" \
      '{model: $model, prompt: $prompt, stream: false, keep_alive: $keep_alive, options: {temperature: 0}}')" \
    2>"$PROBE_ERROR_ARTIFACT")"; then
    rc=$?
    journalctl -u ollama --since "$probe_started_at" --no-pager >"$JOURNAL_ARTIFACT" || true
    if rg -q 'oom-kill|OOM killer|Empty reply from server|Failed with result .oom-kill.' "$JOURNAL_ARTIFACT" "$PROBE_ERROR_ARTIFACT" 2>/dev/null; then
      hg_die "Heavy probe triggered or coincided with an Ollama OOM event. See $JOURNAL_ARTIFACT and $PROBE_ERROR_ARTIFACT"
    fi
    hg_die "Heavy probe failed for $OLLAMA_HEAVY_CODE_MODEL (curl exit $rc). See $PROBE_ERROR_ARTIFACT"
  fi

  printf '%s\n' "$response_json" >"$PROBE_ARTIFACT"

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
mkdir -p "$HEAVY_ARTIFACT_DIR"

while ! model_installed "$OLLAMA_HEAVY_CODE_MODEL"; do
  hg_info "Waiting for heavy model $OLLAMA_HEAVY_CODE_MODEL to finish installing"
  sleep "$WAIT_INTERVAL_SECS"
done

hg_ok "Heavy model installed: $OLLAMA_HEAVY_CODE_MODEL"
free -h | tee "$FREE_ARTIFACT"

"$SCRIPT_DIR/hg-ollama-sync-aliases.sh"

current_threshold="${HG_OLLAMA_DEEP_OFFLOAD_MEMORY_GB:-24}"
available_mem_gb="$(free -g | awk '/^Mem:/ {print $7}')"
if (( available_mem_gb < current_threshold )) && [[ "$FORCE_HEAVY_PROBE" != "1" ]]; then
  hg_die "Available memory ${available_mem_gb}GiB is below the ${current_threshold}GiB heavy preflight. Set HG_OLLAMA_FORCE_HEAVY_PROBE=1 to override."
fi

run_heavy_probe

if (( available_mem_gb < current_threshold )); then
  hg_warn "Heavy probe succeeded with ${available_mem_gb}GiB available, below the current ${current_threshold}GiB preflight"
fi

HG_OLLAMA_DEEP_OFFLOAD_MEMORY_GB=0 "$SCRIPT_DIR/hg-ollama-full-test.sh" --require-heavy | tee "$FULL_TEST_ARTIFACT"
