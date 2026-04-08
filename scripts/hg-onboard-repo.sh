#!/usr/bin/env bash
# hg-onboard-repo.sh — Onboard any repo with hairglasses-studio standard files.
# Adds missing: .editorconfig, .golangci.yml, pipeline.mk include, CI workflow,
# LICENSE, CONTRIBUTING.md, pre-commit hooks.
# Usage: hg-onboard-repo.sh <repo_path> [--language auto|go|node|python|shell] [--dry-run]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

hg_require jq

STUDIO="$HG_STUDIO_ROOT"
DOTFILES="$HG_DOTFILES"
ORG_GITHUB="$STUDIO/.github"
SURFACEKIT_ROOT="${SURFACEKIT_ROOT:-$STUDIO/surfacekit}"
SURFACEKIT_RUN="$SURFACEKIT_ROOT/scripts/run-surfacekit.sh"

[[ -f "$SURFACEKIT_RUN" ]] || hg_die "surfacekit launcher not found at $SURFACEKIT_RUN"

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
    -h|--help)    hg_die "Usage: hg-onboard-repo.sh <repo_path> [--language auto|go|node|python|shell] [--dry-run]" ;;
    -*)           hg_die "Unknown argument: $arg" ;;
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

fi

# ── Node.js-specific ─────────────────────────
if [[ "$LANG" == "node" ]]; then
  :
fi

# ── Python-specific ──────────────────────────
if [[ "$LANG" == "python" ]]; then
  :
fi

# ── Hosted automation ────────────────────────
add_note "Hosted GitHub workflows and Dependabot config are not onboarded under the local-only automation policy"

NEEDS_SURFACE_MIGRATE=false
if [[ ! -f .agents/skills/surface.yaml ]]; then
  while IFS= read -r skill_dir; do
    [[ -f "$skill_dir/SKILL.md" ]] || continue
    NEEDS_SURFACE_MIGRATE=true
    break
  done < <(find .agents/skills .claude/skills -mindepth 1 -maxdepth 1 -type d 2>/dev/null || true)
fi

if $NEEDS_SURFACE_MIGRATE; then
  if $DRY_RUN; then
    add_file "surfacekit skill-surface migration"
  else
    "$SURFACEKIT_RUN" migrate repo "$PWD" >/dev/null
    add_file "surfacekit skill-surface migration"
  fi
elif [[ -d .claude/skills && ! -f .agents/skills/surface.yaml ]]; then
  hg_warn "$REPO_NAME: found legacy .claude/skills but no canonical surface manifest"
fi

if $DRY_RUN; then
  add_file "surfacekit repo init"
else
  "$SURFACEKIT_RUN" init repo "$PWD" --repo-name "$REPO_NAME" --allow-dirty >/dev/null
  add_file "surfacekit repo init"
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
if [[ -e .git ]] && [[ ! -f .git/hooks/pre-commit ]]; then
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
