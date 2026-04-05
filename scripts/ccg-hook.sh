#!/usr/bin/env bash
# ccg-hook.sh — Claude Code Stop hook for session capture
# Refreshes ccg cache and runs incremental session ingest in background.
# Register in ~/.claude/settings.json under hooks.Stop.
set -uo pipefail

CACHE_FILE="$HOME/.claude/cache/ccg-index.jsonl"
DOCS_DIR="$HOME/hairglasses-studio/docs"

# Refresh ccg cache (fast: ~46ms with Go binary).
if command -v dotfiles-mcp >/dev/null 2>&1; then
    dotfiles-mcp --session-index > "$CACHE_FILE" 2>/dev/null &
fi

# Run incremental session ingest in background (non-blocking).
if [[ -d "$DOCS_DIR/cmd/docs-mcp" ]]; then
    (
        cd "$DOCS_DIR" || exit
        CGO_CFLAGS="-DSQLITE_ENABLE_FTS5" CGO_LDFLAGS="-lm" GOWORK=off \
            DOCS_ROOT="$DOCS_DIR" STUDIO_ROOT="$HOME/hairglasses-studio" \
            go run ./cmd/docs-mcp/ session ingest 2>/dev/null
    ) &
fi

# Don't block Claude Code — exit immediately.
exit 0
