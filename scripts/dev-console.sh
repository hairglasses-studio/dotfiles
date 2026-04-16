#!/usr/bin/env bash
# dev-console.sh — Yakuake-style Claude Code system dev console
# Resumes the most recent Claude Code session in the dotfiles repo.
# Bound to Mod+grave via pyprland scratchpad.
set -euo pipefail

cd "$HOME/hairglasses-studio/dotfiles"
exec claude --continue
