#!/usr/bin/env bash
set -euo pipefail

# validate-agents-mirror-contract.sh — guard that CLAUDE.md,
# GEMINI.md, and .github/copilot-instructions.md stay thin
# compatibility mirrors that defer to AGENTS.md.
#
# The per-repo convention (hairglasses-studio/CLAUDE.md +
# workspace/manifest.json docs) is that AGENTS.md is the canonical
# instruction surface. CLAUDE.md and GEMINI.md exist only so their
# respective CLIs pick up the context pointer. A mirror that grows
# into a parallel instruction set, or silently loses its AGENTS.md
# reference, is exactly the drift this gate prevents.
#
# Rules per mirror:
#   - file must exist
#   - file must reference AGENTS.md at least once (literal string)
#   - file must stay under MAX_LINES (~50 for CLAUDE/GEMINI, ~20
#     for copilot-instructions) — prevents a mirror from quietly
#     becoming a second instruction source
#
# Exit 0 on clean, 1 on any drift.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import sys
from pathlib import Path

# (mirror_path, max_lines_before_warn)
MIRRORS = [
    ("CLAUDE.md", 50),
    ("GEMINI.md", 50),
    (".github/copilot-instructions.md", 20),
]

errors: list[str] = []

if not Path("AGENTS.md").exists():
    errors.append("AGENTS.md missing — canonical instruction file expected at repo root")
else:
    agents_text = Path("AGENTS.md").read_text()
    if len(agents_text.splitlines()) < 10:
        errors.append(f"AGENTS.md only has {len(agents_text.splitlines())} lines — "
                      f"canonical file looks truncated")

for mirror, max_lines in MIRRORS:
    path = Path(mirror)
    if not path.exists():
        errors.append(f"{mirror}: mirror missing (AGENTS.md convention requires all three)")
        continue
    text = path.read_text()
    n_lines = len(text.splitlines())
    if "AGENTS.md" not in text:
        errors.append(f"{mirror}: no 'AGENTS.md' reference — mirror lost its canonical pointer")
    if n_lines > max_lines:
        errors.append(f"{mirror}: {n_lines} lines (max {max_lines}) — "
                      f"mirror drifted from thin-mirror contract")

for err in errors:
    print(f"DRIFT: {err}")

print(f"mirrors={len(MIRRORS)} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
