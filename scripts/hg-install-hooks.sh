#!/usr/bin/env bash
# hg-install-hooks.sh — Install language-appropriate pre-commit hooks.
# Run from any repo root. Detects language and writes .git/hooks/pre-commit.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

if [[ ! -d .git ]]; then
  hg_die "Not a git repository (no .git/ in $PWD)"
fi

HOOK=".git/hooks/pre-commit"
mkdir -p .git/hooks

if [[ -f go.mod ]]; then
  cat > "$HOOK" << 'HOOKEOF'
#!/usr/bin/env bash
set -euo pipefail
echo "pre-commit: running go vet..."
go vet ./...
echo "pre-commit: running go test -short..."
go test -short -count=1 ./... 2>&1 | tail -5
HOOKEOF
  chmod +x "$HOOK"
  hg_ok "Installed Go pre-commit hook (vet + test-short)"

elif [[ -f package.json ]]; then
  cat > "$HOOK" << 'HOOKEOF'
#!/usr/bin/env bash
set -euo pipefail
if grep -q '"lint"' package.json 2>/dev/null; then
  echo "pre-commit: running npm run lint..."
  npm run lint
fi
HOOKEOF
  chmod +x "$HOOK"
  hg_ok "Installed Node.js pre-commit hook (lint)"

elif [[ -f pyproject.toml ]] || [[ -f requirements.txt ]]; then
  cat > "$HOOK" << 'HOOKEOF'
#!/usr/bin/env bash
set -euo pipefail
if command -v ruff &>/dev/null; then
  echo "pre-commit: running ruff check..."
  ruff check .
fi
HOOKEOF
  chmod +x "$HOOK"
  hg_ok "Installed Python pre-commit hook (ruff)"

else
  hg_warn "No language detected — skipping hook installation"
  exit 0
fi
