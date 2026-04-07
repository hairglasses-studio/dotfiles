#!/usr/bin/env bash
# hg-new-repo.sh — Scaffold a new hairglasses-studio repo with all standard files.
# Usage: hg-new-repo.sh <name> [go|node|python]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

DOTFILES="$HG_DOTFILES"
ORG_GITHUB="$HG_STUDIO_ROOT/.github"
STUDIO="$HG_STUDIO_ROOT"

workflow_source() {
  local wf="$1"
  if [[ -f "$ORG_GITHUB/workflow-templates/$wf" ]]; then
    printf '%s\n' "$ORG_GITHUB/workflow-templates/$wf"
  else
    printf '%s\n' "$ORG_GITHUB/.github/workflows/$wf"
  fi
}

NAME="${1:-}"
LANG="${2:-go}"

if [[ -z "$NAME" ]]; then
  hg_die "Usage: hg-new-repo.sh <name> [go|node|python]"
fi

REPO_DIR="$STUDIO/$NAME"

if [[ -d "$REPO_DIR" ]]; then
  hg_die "Directory already exists: $REPO_DIR"
fi

hg_info "Scaffolding $NAME ($LANG) at $REPO_DIR"
mkdir -p "$REPO_DIR"
cd "$REPO_DIR"

# ── Git init ─────────────────────────────────
git init -q

# ── Symlinks to dotfiles ─────────────────────
ln -sf "../../dotfiles/editorconfig" .editorconfig
ln -sf "../../dotfiles/make/golangci.yml" .golangci.yml
hg_ok "Symlinked .editorconfig, .golangci.yml"

# ── .gitignore ───────────────────────────────
"$SCRIPT_DIR/hg-gitignore.sh" "$LANG" > .gitignore
hg_ok "Generated .gitignore ($LANG)"

# ── Governance ───────────────────────────────
# LICENSE is per-repo (different copyright holders possible)
cat > LICENSE << 'LICEOF'
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
# CONTRIBUTING.md per-repo (org .github provides default; override with repo-specific)
sed "s/<repo>/$NAME/g" "$ORG_GITHUB/CONTRIBUTING.md" > CONTRIBUTING.md
# CODEOWNERS, issue/PR templates inherited from org .github repo — no local copies needed
hg_ok "Created LICENSE, CONTRIBUTING.md (issue/PR templates inherited from org)"

# ── CI workflows ─────────────────────────────
mkdir -p .github/workflows
# Copy standard workflows from the org templates
for wf in claude-review.yml claude-security.yml codex-review.yml codex-security.yml codex-structured-audit.yml codex-baseline-guard.yml ai-dispatch.yml dependabot-auto-merge.yml; do
  src="$(workflow_source "$wf")"
  [[ "$wf" == "dependabot-auto-merge.yml" ]] && src="$STUDIO/mcpkit/.github/workflows/$wf"
  [[ -f "$src" ]] && command cp -f "$src" ".github/workflows/$wf"
done

# CI workflow (standalone template — dotfiles is private, can't use reusable workflows)
if [[ "$LANG" == "go" ]]; then
  command cp -f "$DOTFILES/make/ci-go.yml" .github/workflows/ci.yml 2>/dev/null || \
  command cp -f "$ORG_GITHUB/workflow-templates/ci-go.yml" .github/workflows/ci.yml 2>/dev/null || true
fi

# Dependabot config
cat > .github/dependabot.yml << 'DEPEOF'
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule:
      interval: weekly
      day: monday
      time: "09:00"
      timezone: America/Los_Angeles
    groups:
      minor-and-patch:
        update-types: [minor, patch]
    open-pull-requests-limit: 10
    commit-message:
      prefix: "deps(go)"
  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
    open-pull-requests-limit: 5
    commit-message:
      prefix: "deps(actions)"
DEPEOF
hg_ok "Created CI workflows + dependabot.yml"

# ── Language-specific files ──────────────────
case "$LANG" in
  go)
    GO_VERSION=$(cat "$DOTFILES/make/go-version" | tr -d '[:space:]')
    cat > go.mod << GOEOF
module github.com/hairglasses-studio/$NAME

go $GO_VERSION
GOEOF

    cat > Makefile << 'MKEOF'
.PHONY: build test vet

build:
	go build ./...

test:
	go test ./... -count=1

vet:
	go vet ./...

-include $(HOME)/hairglasses-studio/dotfiles/make/pipeline.mk
MKEOF

    cat > main.go << 'MAINEOF'
package main

import "fmt"

func main() {
	fmt.Println("hello from $NAME")
}
MAINEOF
    sed -i "s/\$NAME/$NAME/g" main.go
    hg_ok "Created go.mod, Makefile, main.go"
    ;;

  node)
    cat > package.json << PKGEOF
{
  "name": "$NAME",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "build": "echo 'no build step'",
    "test": "echo 'no tests yet'"
  }
}
PKGEOF
    hg_ok "Created package.json"
    ;;

  python)
    cat > pyproject.toml << PYEOF
[project]
name = "$NAME"
version = "0.1.0"
requires-python = ">= 3.11"
PYEOF
    hg_ok "Created pyproject.toml"
    ;;
esac

# ── AGENTS.md skeleton ───────────────────────
cat > AGENTS.md << AGEOF
# $NAME — Agent Instructions

> Canonical instructions: AGENTS.md

## Build & Test

\`\`\`bash
make build
make test
make pipeline-check
\`\`\`

## Architecture

Describe project structure here.

## Key Patterns

Document project conventions here.

## Explicit Skill Surface

- Canonical reusable workflow skills live under \`.agents/skills/\`.
- Generated compatibility mirrors under \`.claude/skills/\` must come from \`dotfiles/scripts/hg-skill-surface-sync.sh\`.
- \`.codex/agents/*.toml\` is for Codex delegation roles, not the primary workflow-skill surface.
AGEOF
hg_ok "Created AGENTS.md skeleton"

# ── Derived agent docs ───────────────────────
"$SCRIPT_DIR/hg-agent-docs.sh" --source agents "$REPO_DIR"
hg_ok "Generated CLAUDE.md, GEMINI.md, and .github/copilot-instructions.md"

# ── Codex config baseline ────────────────────
mkdir -p .codex
command cp -f "$SCRIPT_DIR/../templates/codex-config.standard.toml" .codex/config.toml
hg_ok "Created .codex/config.toml from shared standard"

# ── Gemini settings baseline ─────────────────
"$SCRIPT_DIR/hg-provider-settings-sync.sh" "$REPO_DIR" --repo-name "$NAME" --allow-dirty >/dev/null
hg_ok "Created .claude/settings.json and .gemini/settings.json from shared standard"

# ── Canonical skill surface ──────────────────
SKILL_NAME="${NAME//-/_}_ops"
mkdir -p ".agents/skills/$SKILL_NAME"
cat > .agents/skills/surface.yaml << EOF
{
  "version": 1,
  "skills": [
    {
      "name": "$SKILL_NAME",
      "claude_include_canonical": true,
      "export_plugin": false
    }
  ]
}
EOF
cat > ".agents/skills/$SKILL_NAME/SKILL.md" << EOF
---
name: $SKILL_NAME
description: Core build, test, release, and maintenance workflow for the $NAME repo.
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - Glob
---

# ${NAME} Ops

Use this skill for the default repo workflow in \`$NAME\`.

## Default loop

1. Read \`AGENTS.md\` and confirm the repo-specific build and test commands.
2. Inspect the relevant code or docs before editing.
3. Make focused changes that preserve existing conventions.
4. Run the narrowest useful verification first, then the broader repo checks before finishing.
5. Summarize outcome, verification, and any remaining risks.

Read files under \`.agents/skills/$SKILL_NAME/references/\` if this repo later grows domain-specific workflows.
EOF
"$SCRIPT_DIR/hg-skill-surface-sync.sh" "$REPO_DIR" >/dev/null
hg_ok "Created canonical .agents/skills surface"

if hg_workspace_repo_exists "$NAME"; then
  "$SCRIPT_DIR/hg-repo-profile-sync.sh" sync --allow-dirty --repos="$NAME" >/dev/null
  hg_ok "Applied manifest profile sync"
else
  hg_warn "$NAME: not yet declared in workspace/manifest.json; using scaffold defaults"
fi

# ── Pre-commit hooks ─────────────────────────
"$SCRIPT_DIR/hg-install-hooks.sh"

# ── Initial commit ───────────────────────────
git add -A
git commit -q -m "initial scaffold via hg-new-repo.sh"

echo ""
hg_ok "Repo scaffolded at $REPO_DIR"
hg_info "Next: cd $REPO_DIR && gh repo create hairglasses-studio/$NAME --private --push"
