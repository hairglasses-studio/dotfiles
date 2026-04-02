#!/usr/bin/env bash
# hg-new-repo.sh — Scaffold a new hairglasses-studio repo with all standard files.
# Usage: hg-new-repo.sh <name> [go|node|python]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

DOTFILES="$HOME/hairglasses-studio/dotfiles"
TEMPLATES="$DOTFILES/github-templates"
STUDIO="$HOME/hairglasses-studio"

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

# ── Governance files ─────────────────────────
cp "$TEMPLATES/LICENSE" .
sed "s/{PROJECT_NAME}/$NAME/g" "$TEMPLATES/CONTRIBUTING.md" > CONTRIBUTING.md
cp "$TEMPLATES/CODEOWNERS" .
mkdir -p .github/ISSUE_TEMPLATE
cp "$TEMPLATES/.github/ISSUE_TEMPLATE/bug_report.md" .github/ISSUE_TEMPLATE/
cp "$TEMPLATES/.github/ISSUE_TEMPLATE/feature_request.md" .github/ISSUE_TEMPLATE/
cp "$TEMPLATES/.github/pull_request_template.md" .github/
hg_ok "Copied LICENSE, CONTRIBUTING.md, CODEOWNERS, issue/PR templates"

# ── CI workflows ─────────────────────────────
mkdir -p .github/workflows
# Copy the 3 standard workflows from any existing repo
for wf in claude-review.yml claude-security.yml dependabot-auto-merge.yml; do
  src="$STUDIO/mcpkit/.github/workflows/$wf"
  if [[ -f "$src" ]]; then
    cp "$src" ".github/workflows/$wf"
  fi
done

# Generate ci.yml that calls reusable workflow
if [[ "$LANG" == "go" ]]; then
  cat > .github/workflows/ci.yml << 'CIEOF'
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
jobs:
  ci:
    uses: hairglasses-studio/dotfiles/.github/workflows/ci-go.yml@main
CIEOF
fi
hg_ok "Created CI workflows"

# ── .codex config ────────────────────────────
mkdir -p .codex
printf 'model = "gpt-5.4-xhigh"\n' > .codex/config.toml
hg_ok "Created .codex/config.toml"

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

# ── CLAUDE.md skeleton ───────────────────────
cat > CLAUDE.md << CLEOF
# $NAME

## Build & Test

\`\`\`bash
make build
make test
make pipeline-check
\`\`\`

## Architecture

TODO: Describe the project structure.

## Key Patterns

TODO: Document conventions specific to this project.
CLEOF
hg_ok "Created CLAUDE.md skeleton"

# ── Pre-commit hooks ─────────────────────────
"$SCRIPT_DIR/hg-install-hooks.sh"

# ── Initial commit ───────────────────────────
git add -A
git commit -q -m "initial scaffold via hg-new-repo.sh"

echo ""
hg_ok "Repo scaffolded at $REPO_DIR"
hg_info "Next: cd $REPO_DIR && gh repo create hairglasses-studio/$NAME --private --push"
