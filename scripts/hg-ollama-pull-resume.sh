#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODEL_NAME="${1:-}"
PULL_SLICE="${HG_OLLAMA_PULL_SLICE:-5m}"
RETRY_SLEEP_SECS="${HG_OLLAMA_PULL_RETRY_SLEEP_SECS:-5}"
MAX_ATTEMPTS="${HG_OLLAMA_PULL_MAX_ATTEMPTS:-0}"

usage() {
  cat <<'EOF'
Usage: hg-ollama-pull-resume.sh <model>

Environment:
  HG_OLLAMA_PULL_SLICE            Time slice for each pull attempt. Default: 5m
  HG_OLLAMA_PULL_RETRY_SLEEP_SECS Seconds to wait before reconnecting. Default: 5
  HG_OLLAMA_PULL_MAX_ATTEMPTS     Optional hard cap. Default: 0 (unlimited)
EOF
}

model_installed() {
  local model="$1"
  ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "$model" >/dev/null 2>&1 \
    || ollama list | awk 'NR > 1 {print $1}' | grep -Fx -- "${model}:latest" >/dev/null 2>&1
}

[[ -n "$MODEL_NAME" ]] || {
  usage >&2
  exit 1
}

hg_require ollama timeout awk grep sed

attempt=0
while ! model_installed "$MODEL_NAME"; do
  attempt=$((attempt + 1))
  if (( MAX_ATTEMPTS > 0 && attempt > MAX_ATTEMPTS )); then
    hg_die "Exceeded max attempts ($MAX_ATTEMPTS) while pulling $MODEL_NAME"
  fi

  hg_info "Pull attempt $attempt for $MODEL_NAME (slice $PULL_SLICE)"
  if timeout --foreground "$PULL_SLICE" ollama pull "$MODEL_NAME"; then
    if model_installed "$MODEL_NAME"; then
      break
    fi
    hg_warn "Pull command exited but $MODEL_NAME is not fully installed yet; reconnecting"
  else
    rc=$?
    case "$rc" in
      124)
        hg_warn "Pull slice reached $PULL_SLICE for $MODEL_NAME; reconnecting"
        ;;
      130|143)
        hg_warn "Pull slice interrupted for $MODEL_NAME; reconnecting"
        ;;
      *)
        hg_warn "ollama pull exited with status $rc for $MODEL_NAME; reconnecting"
        ;;
    esac
  fi

  sleep "$RETRY_SLEEP_SECS"
done

hg_ok "Installed $MODEL_NAME"
ollama show "$MODEL_NAME" | sed -n '1,40p'
