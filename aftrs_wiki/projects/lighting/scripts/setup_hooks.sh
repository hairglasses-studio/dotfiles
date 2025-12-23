#!/usr/bin/env bash
set -euo pipefail
git config core.hooksPath .githooks
echo "Git hooks enabled (core.hooksPath=.githooks). Pre-commit will render Mermaid diagrams."
