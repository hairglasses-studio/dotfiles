#!/usr/bin/env bash
set -euo pipefail

# validate-orphan-allowlist.sh — guard that every entry in
# scripts/audit-orphan-scripts.sh's allowlist still has a matching
# scripts/<name>.* file on disk.
#
# The orphan audit skips flagging these "intentionally unreferenced
# from the repo" scripts. A rename or removal that leaves the
# allowlist entry behind is a ghost: nothing flags it, and the
# allowlist drifts from ground truth. Over time the list stops
# meaning what its comments say.
#
# Exit 0 on clean, 1 if any allowlist entry points at a file that
# no longer exists.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

AUDIT = Path("scripts/audit-orphan-scripts.sh")
SCRIPTS_DIR = Path("scripts")

if not AUDIT.is_file():
    print(f"ERROR: {AUDIT} missing", file=sys.stderr)
    sys.exit(2)

text = AUDIT.read_text()

# Slice the `allowlist=(` … `)` block. We only care about the
# first occurrence; the block is declared once at the top of the
# script.
match = re.search(r"^allowlist=\(\s*\n(.*?)^\)", text, re.DOTALL | re.MULTILINE)
if not match:
    print("ERROR: could not locate allowlist=( … ) block", file=sys.stderr)
    sys.exit(2)

body = match.group(1)
entries: list[str] = []
for raw in body.splitlines():
    line = raw.strip()
    # Drop blank lines and pure-comment lines.
    if not line or line.startswith("#"):
        continue
    # Strip inline comments after the first whitespace.
    token = line.split("#", 1)[0].strip()
    if token:
        entries.append(token)

errors: list[str] = []
for stem in entries:
    candidates = list(SCRIPTS_DIR.glob(f"{stem}.*"))
    # Exclude backup files (e.g. *.bak, *.bak.<date>) so a noisy
    # backup doesn't pretend the script still exists.
    real = [c for c in candidates if not c.name.endswith(".bak") and ".bak." not in c.name]
    if not real:
        errors.append(f"{stem}: allowlist entry with no scripts/{stem}.* source")

for err in errors:
    print(f"DRIFT: {err}")

print(f"allowlist_entries={len(entries)} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
