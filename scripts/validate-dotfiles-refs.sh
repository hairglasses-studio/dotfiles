#!/usr/bin/env bash
set -euo pipefail

# validate-dotfiles-refs.sh — guard that every
# `$HOME/hairglasses-studio/dotfiles/<path>` reference in compositor,
# launcher, and bar configs resolves to a tracked file in this repo.
#
# ok 15 (validate-local-bin-refs.sh) only checks `$HOME/.local/bin/*`
# wrappers — bypass paths that point straight at dotfiles scripts
# slip through. A bind line with a stale path silently no-ops on
# hyprctl dispatch; an exec-once line errors into the hyprland log
# without blocking the compositor. Both are invisible at edit time.
#
# Exit 0 on clean, 1 if any reference points at a missing file.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

# Configs that may embed direct $HOME/hairglasses-studio/dotfiles paths.
# Kept narrow on purpose — expanding this list is an explicit decision.
configs = [
    "hyprland/hyprland.conf",
    "pypr/config.toml",
    "ironbar/config.toml",
]

# Match $HOME/hairglasses-studio/dotfiles/<path> up to the next shell
# separator (space, quote, end-of-arg). Strips trailing punctuation
# that's part of TOML/hypr syntax but not the path itself.
ref_pat = re.compile(
    r'\$HOME/hairglasses-studio/dotfiles/([A-Za-z0-9._/-]+)'
)

refs: dict[str, set[str]] = {}
for cfg in configs:
    path = Path(cfg)
    if not path.exists():
        continue
    found = set(ref_pat.findall(path.read_text(errors="ignore")))
    if found:
        refs[cfg] = found

missing: dict[str, list[str]] = {}
for cfg, paths in refs.items():
    gaps = []
    for rel in sorted(paths):
        target = Path(rel)
        if not target.exists():
            gaps.append(rel)
    if gaps:
        missing[cfg] = gaps

total_refs = sum(len(v) for v in refs.values())
total_missing = sum(len(v) for v in missing.values())

for cfg, gaps in missing.items():
    for rel in gaps:
        print(f"MISSING: {cfg} -> $HOME/hairglasses-studio/dotfiles/{rel}")

print(f"configs={len(refs)} refs={total_refs} errors={total_missing}")
sys.exit(1 if total_missing else 0)
PY
