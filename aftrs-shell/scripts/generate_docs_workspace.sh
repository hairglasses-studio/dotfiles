#!/usr/bin/env bash

# Generate a VS Code/Cursor multi-root workspace that includes all git repos under ~/Docs
# Output: ~/Docs/Docs.code-workspace

set -euo pipefail

HOST="$(hostname -s || echo unknown)"
USE_GUM=false
if command -v gum >/dev/null 2>&1 && [ "$HOST" != "wizardstower" ]; then USE_GUM=true; fi
log()   { if $USE_GUM; then gum log --level info "$@"; else echo "[INFO] $*"; fi; }
warn()  { if $USE_GUM; then gum log --level warn "$@"; else echo "[WARN] $*" >&2; fi; }

DOCS_ROOT="${AFTRS_SHELL_HOME:-$HOME/Docs}"
OUT_FILE="$DOCS_ROOT/Docs.code-workspace"

if [ ! -d "$DOCS_ROOT" ]; then
  warn "Docs root not found: $DOCS_ROOT"
  exit 1
fi

log "Scanning git repositories under $DOCS_ROOT"

# Collect repo directories (parent of .git), relative to DOCS_ROOT
mapfile -t repos < <(
  find "$DOCS_ROOT" -type d -name .git -prune 2>/dev/null \
    | sed 's#/.git$##' \
    | sort -u
)

log "Found ${#repos[@]} repositories"

tmp_json="$(mktemp)"

{
  echo '{'
  echo '  "folders": ['

  first=true
  for abs in "${repos[@]}"; do
    # Make path relative for portability when opening the workspace from ~/Docs
    rel="${abs#"$DOCS_ROOT/"}"
    # Skip if outside DOCS_ROOT (paranoia)
    case "$rel" in /*|"") continue ;; esac

    if $first; then first=false; else echo ','; fi
    printf '    { "path": "%s" }' "$rel"
  done

  echo
  echo '  ],'
  echo '  "settings": {'
  echo '    "terminal.integrated.cwd": "${env:HOME}/Docs"'
  echo '  }'
  echo '}'
} > "$tmp_json"

mv "$tmp_json" "$OUT_FILE"
log "Wrote workspace: $OUT_FILE"


