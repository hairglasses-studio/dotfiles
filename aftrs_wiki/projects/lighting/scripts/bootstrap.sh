#!/usr/bin/env bash
    set -euo pipefail
    REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
    cd "$REPO_ROOT"

    echo "==> Enabling Git hooks (.githooks/)"
    ./scripts/setup_hooks.sh

    echo "==> Updating README badges from git remote (if any)"
    if ./scripts/update_badges_from_git_remote.sh; then
      echo "   - Badges patched"
    else
      echo "   - Skipped (no remote or unsupported URL)"
    fi

    echo "==> Rendering Mermaid diagrams (Docker/Podman/mmdc)"
    if ./scripts/render_diagrams.sh; then
      echo "   - Diagrams rendered to diagrams/_rendered"
    else
      echo "   - Rendering failed. Install Docker/Podman or: npm i -g @mermaid-js/mermaid-cli"
      exit 2
    fi

    echo "==> Validating DMX patch (optional)"
    if command -v python3 >/dev/null 2>&1 && python3 - <<'PY' >/dev/null 2>&1
import sys
try:
    import yaml
except Exception:
    sys.exit(1)
PY
    then
      python3 tools/validate_patch.py || true
    else
      echo "   - PyYAML not available; skipping. Install with: pip install pyyaml"
    fi

    echo "Bootstrap complete."
