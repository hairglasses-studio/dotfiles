#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-local-llm.sh"

hg_require ollama awk
hg_local_llm_export_env

sync_alias() {
  local alias_name="$1"
  local source_model="$2"

  if ! ollama show "$source_model" >/dev/null 2>&1; then
    hg_warn "Skipping $alias_name; source model $source_model is not installed"
    return 0
  fi

  if ollama show "$alias_name" >/dev/null 2>&1; then
    ollama rm "$alias_name" >/dev/null
    hg_info "Removed existing alias $alias_name"
  fi

  ollama cp "$source_model" "$alias_name" >/dev/null
  hg_ok "Synced $alias_name -> $source_model"
}

while IFS='=' read -r alias_name source_model; do
  [[ -n "$alias_name" && -n "$source_model" ]] || continue
  sync_alias "$alias_name" "$source_model"
done < <(hg_local_llm_code_alias_pairs)

hg_info "Managed aliases:"
ollama list | awk 'NR == 1 || $1 ~ /^code-(fast|compact|primary|reasoner|long|heavy)(:latest)?$/'
