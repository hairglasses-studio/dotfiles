#!/usr/bin/env bash
# hg-onboard-repo.sh — Onboard any repo with hairglasses-studio standard files.
# Adds missing: .editorconfig, .golangci.yml, pipeline.mk include, CI workflow,
# LICENSE, CONTRIBUTING.md, pre-commit hooks.
# Usage: hg-onboard-repo.sh <repo_path> [--language auto|go|node|python|shell] [--dry-run]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

STUDIO="$HG_STUDIO_ROOT"
DOTFILES="$HG_DOTFILES"
ORG_GITHUB="$STUDIO/.github"

workflow_source() {
  local wf="$1"
  if [[ -f "$ORG_GITHUB/workflow-templates/$wf" ]]; then
    printf '%s\n' "$ORG_GITHUB/workflow-templates/$wf"
  else
    printf '%s\n' "$ORG_GITHUB/.github/workflows/$wf"
  fi
}

REPO_PATH=""
LANG="auto"
DRY_RUN=false

for arg in "$@"; do
  case "$arg" in
    --language=*) LANG="${arg#*=}" ;;
    --dry-run)    DRY_RUN=true ;;
    -*)           ;;
    *)            REPO_PATH="$arg" ;;
  esac
done

if [[ -z "$REPO_PATH" ]]; then
  hg_die "Usage: hg-onboard-repo.sh <repo_path> [--language auto|go|node|python|shell] [--dry-run]"
fi

cd "$REPO_PATH"
REPO_NAME="$(basename "$PWD")"

# ── Language detection ───────────────────────
if [[ "$LANG" == "auto" ]]; then
  if [[ -f go.mod ]]; then
    LANG="go"
  elif [[ -f package.json ]]; then
    LANG="node"
  elif [[ -f pyproject.toml ]] || [[ -f requirements.txt ]] || [[ -f setup.py ]]; then
    LANG="python"
  elif ls *.sh *.zsh *.plugin.zsh 2>/dev/null | head -1 >/dev/null 2>&1; then
    LANG="shell"
  else
    LANG="none"
  fi
fi

hg_info "Onboarding $REPO_NAME (lang=$LANG, dry_run=$DRY_RUN)"
ADDED=0

# ── Helper ───────────────────────────────────
add_file() {
  local desc="$1"
  if $DRY_RUN; then
    hg_warn "Would add: $desc"
  else
    hg_ok "Added: $desc"
  fi
  ADDED=$((ADDED + 1))
}

# ── .editorconfig ────────────────────────────
if [[ ! -L .editorconfig ]] && [[ ! -f .editorconfig ]]; then
  $DRY_RUN || ln -sf "../../dotfiles/editorconfig" .editorconfig
  add_file ".editorconfig (symlink)"
fi

# ── LICENSE ──────────────────────────────────
if [[ ! -f LICENSE ]] && [[ ! -f LICENSE.md ]]; then
  if [[ -f "$ORG_GITHUB/LICENSE" ]]; then
    $DRY_RUN || command cp -f "$ORG_GITHUB/LICENSE" LICENSE 2>/dev/null
  else
    $DRY_RUN || cat > LICENSE << 'LICEOF'
MIT License

Copyright (c) 2024-2026 hairglasses-studio

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
LICEOF
  fi
  add_file "LICENSE (MIT)"
fi

# ── CONTRIBUTING.md ──────────────────────────
if [[ ! -f CONTRIBUTING.md ]]; then
  if [[ -f "$ORG_GITHUB/CONTRIBUTING.md" ]]; then
    $DRY_RUN || sed "s/<repo>/$REPO_NAME/g" "$ORG_GITHUB/CONTRIBUTING.md" > CONTRIBUTING.md
  fi
  add_file "CONTRIBUTING.md"
fi

# ── Go-specific ──────────────────────────────
if [[ "$LANG" == "go" ]]; then
  # .golangci.yml symlink
  if [[ ! -L .golangci.yml ]]; then
    $DRY_RUN || ln -sf "../../dotfiles/make/golangci.yml" .golangci.yml
    add_file ".golangci.yml (symlink)"
  fi

  # pipeline.mk include in Makefile
  if [[ -f Makefile ]] && ! grep -q 'pipeline.mk' Makefile 2>/dev/null; then
    $DRY_RUN || printf '\n-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk\n' >> Makefile
    add_file "pipeline.mk include in Makefile"
  elif [[ ! -f Makefile ]]; then
    $DRY_RUN || cat > Makefile << 'MKEOF'
.PHONY: build test vet

build:
	go build ./...

test:
	go test ./... -count=1

vet:
	go vet ./...

-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
MKEOF
    add_file "Makefile (with pipeline.mk)"
  fi

  # CI workflow
  if [[ ! -f .github/workflows/ci.yml ]]; then
    $DRY_RUN || { mkdir -p .github/workflows && command cp -f "$DOTFILES/make/ci-go.yml" .github/workflows/ci.yml; }
    add_file ".github/workflows/ci.yml (Go)"
  fi
fi

# ── Node.js-specific ─────────────────────────
if [[ "$LANG" == "node" ]]; then
  if [[ ! -f .github/workflows/ci.yml ]]; then
    $DRY_RUN || { mkdir -p .github/workflows && command cp -f "$DOTFILES/make/ci-node.yml" .github/workflows/ci.yml; }
    add_file ".github/workflows/ci.yml (Node)"
  fi
fi

# ── Python-specific ──────────────────────────
if [[ "$LANG" == "python" ]]; then
  if [[ ! -f .github/workflows/ci.yml ]]; then
    $DRY_RUN || { mkdir -p .github/workflows && command cp -f "$DOTFILES/make/ci-python.yml" .github/workflows/ci.yml; }
    add_file ".github/workflows/ci.yml (Python)"
  fi
fi

# ── Standard workflows (all repos) ──────────
$DRY_RUN || mkdir -p .github/workflows
for wf in claude-review.yml claude-security.yml codex-review.yml codex-security.yml codex-structured-audit.yml codex-baseline-guard.yml ai-dispatch.yml dependabot-auto-merge.yml; do
  if [[ ! -f ".github/workflows/$wf" ]]; then
    src="$(workflow_source "$wf")"
    [[ "$wf" == "dependabot-auto-merge.yml" ]] && src="$STUDIO/mcpkit/.github/workflows/$wf"
    if [[ -f "$src" ]]; then
      $DRY_RUN || command cp -f "$src" ".github/workflows/$wf"
      add_file ".github/workflows/$wf"
    fi
  fi
done

# ── Codex config ─────────────────────────────
if [[ ! -f .codex/config.toml ]]; then
  if $DRY_RUN; then
    add_file ".codex/config.toml"
  else
    mkdir -p .codex
    command cp -f "$SCRIPT_DIR/../templates/codex-config.standard.toml" .codex/config.toml
    add_file ".codex/config.toml"
  fi
fi

# ── Gemini settings ──────────────────────────
if [[ ! -f .gemini/settings.json ]]; then
  if $DRY_RUN; then
    add_file ".gemini/settings.json"
  else
    "$SCRIPT_DIR/hg-gemini-settings-sync.sh" "$PWD" >/dev/null
    add_file ".gemini/settings.json"
  fi
fi

# ── Gemini project settings ──────────────────
if [[ ! -f .gemini/settings.json ]]; then
  if $DRY_RUN; then
    add_file ".gemini/settings.json"
  else
    mkdir -p .gemini
    command cp -f "$SCRIPT_DIR/../templates/gemini-settings.standard.json" .gemini/settings.json
    add_file ".gemini/settings.json"
  fi
fi

# ── Codex MCP sync ───────────────────────────
if [[ -f .mcp.json && -f .codex/mcp-profile-policy.json ]]; then
  if $DRY_RUN; then
    add_file ".codex/config.toml MCP block sync"
  else
    "$SCRIPT_DIR/hg-codex-mcp-sync.sh" "$PWD" >/dev/null
    add_file ".codex/config.toml MCP block sync"
  fi
elif [[ -f .mcp.json && ! -f .codex/mcp-profile-policy.json ]]; then
  hg_warn "$REPO_NAME: found .mcp.json but no .codex/mcp-profile-policy.json; skipping Codex MCP sync"
fi

# ── Explicit skill surface sync ──────────────
if [[ -f .agents/skills/surface.yaml ]]; then
  if ! jq -e '.version == 1 and (.skills | type == "array")' .agents/skills/surface.yaml >/dev/null 2>&1; then
    hg_warn "$REPO_NAME: found YAML-only .agents/skills/surface.yaml; skipping repo-local Claude mirror sync"
  elif $DRY_RUN; then
    add_file "canonical skill surface sync"
  else
    "$SCRIPT_DIR/hg-skill-surface-sync.sh" "$PWD" >/dev/null
    add_file "canonical skill surface sync"
  fi
elif [[ -d .claude/skills ]]; then
  hg_warn "$REPO_NAME: found legacy .claude/skills but no .agents/skills/surface.yaml"
fi

# ── Derived agent docs ───────────────────────
if { [[ -f CLAUDE.md ]] || [[ -f AGENTS.md ]]; } && { [[ ! -f CLAUDE.md ]] || [[ ! -f AGENTS.md ]] || [[ ! -f GEMINI.md ]] || [[ ! -f .github/copilot-instructions.md ]]; }; then
  if $DRY_RUN; then
    add_file "CLAUDE.md, AGENTS.md, GEMINI.md, and .github/copilot-instructions.md"
  else
    "$SCRIPT_DIR/hg-agent-docs.sh" "$PWD" >/dev/null
    add_file "CLAUDE.md, AGENTS.md, GEMINI.md, and .github/copilot-instructions.md"
  fi
fi

# ── Pre-commit hooks ────────────────────────
if [[ -d .git ]] && [[ ! -f .git/hooks/pre-commit ]]; then
  $DRY_RUN || "$SCRIPT_DIR/hg-install-hooks.sh" 2>/dev/null || true
  add_file "pre-commit hook"
fi

# ── Managed workspace-global refresh ────────
if hg_workspace_repo_exists "$REPO_NAME" && hg_workspace_repo_bool "$REPO_NAME" "baseline_target"; then
  if $DRY_RUN; then
    hg_warn "Would refresh managed workspace-global sync"
  else
    "$SCRIPT_DIR/hg-workspace-global-sync.sh" --root "$STUDIO" >/dev/null
    hg_ok "Refreshed managed workspace-global sync"
  fi
fi

if hg_workspace_repo_exists "$REPO_NAME"; then
  if $DRY_RUN; then
    "$SCRIPT_DIR/hg-repo-profile-sync.sh" verify --repos="$REPO_NAME" >/dev/null || true
    hg_warn "Verified manifest profile sync for $REPO_NAME"
  else
    "$SCRIPT_DIR/hg-repo-profile-sync.sh" sync --allow-dirty --repos="$REPO_NAME" >/dev/null
    hg_ok "Applied manifest profile sync"
  fi
fi

# ── Summary ──────────────────────────────────
echo ""
if [[ $ADDED -eq 0 ]]; then
  hg_ok "$REPO_NAME: already fully onboarded"
else
  if $DRY_RUN; then
    hg_info "$REPO_NAME: $ADDED files would be added"
  else
    hg_ok "$REPO_NAME: $ADDED files added"
  fi
fi
