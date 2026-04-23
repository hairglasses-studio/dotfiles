#!/usr/bin/env bash
set -euo pipefail

# validate-skill-aliases.sh — guard that every hyphenated alias in
# .claude/skills/ points at a real canonical skill directory.
#
# The org convention is: canonical skills live at `.agents/skills/<snake_case>`
# and `.claude/skills/<snake_case>`. For each canonical we ship a
# hyphen-named compatibility alias at `.claude/skills/<kebab-case>`
# whose SKILL.md description reads
#
#     "Hyphenated compatibility alias for the <canonical_name> workflow."
#
# Drift class this catches:
#   - An alias describes a workflow `<name>` that no longer exists as
#     a canonical directory — a rename of `codex_rollout` to `codex_runner`
#     without fixing `codex-rollout/SKILL.md` would leave the alias
#     pointing at a ghost, silently redirecting `/codex-rollout` to nothing.
#
# Exit 0 on clean, 1 on any drift.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import sys
from pathlib import Path

CLAUDE_SKILLS = Path(".claude/skills")
AGENTS_SKILLS = Path(".agents/skills")

if not CLAUDE_SKILLS.is_dir():
    print(f"ERROR: {CLAUDE_SKILLS} missing", file=sys.stderr)
    sys.exit(2)

# For every hyphenated alias directory, the canonical snake_case
# name (hyphen→underscore) must exist as a directory in BOTH
# .claude/skills/ and .agents/skills/. Two alias styles coexist
# in this repo — a short "Hyphenated compatibility alias for the X
# workflow" pointer and a full content mirror with a snake-case
# `name:` frontmatter — but both rely on the canonical directory
# still existing. A rename of `codex_rollout` → `codex_runner`
# without renaming `codex-rollout` silently strands the alias.

errors: list[str] = []
aliases_seen = 0
canonicals_seen = 0

for entry in sorted(CLAUDE_SKILLS.iterdir()):
    if not entry.is_dir():
        continue
    if not (entry / "SKILL.md").is_file():
        continue
    name = entry.name
    if "-" not in name:
        canonicals_seen += 1
        continue

    aliases_seen += 1
    canonical = name.replace("-", "_")
    if not (CLAUDE_SKILLS / canonical).is_dir():
        errors.append(
            f"{name}: alias has no canonical — expected .claude/skills/{canonical}/"
        )
    if not (AGENTS_SKILLS / canonical).is_dir():
        errors.append(
            f"{name}: alias has no canonical — expected .agents/skills/{canonical}/"
        )

for err in errors:
    print(f"DRIFT: {err}")

print(f"canonicals={canonicals_seen} aliases={aliases_seen} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
