#!/usr/bin/env bash
# validate-shader-count-claims.sh -- keep public shader counts honest.

set -euo pipefail

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

shader_dir = Path("kitty/shaders/darkwindow")
actual = len(list(shader_dir.glob("*.glsl")))
errors: list[str] = []

checks = {
    "README.md": [
        re.compile(r"GLSL_Shaders-([0-9]+)"),
        re.compile(r"\b([0-9]+) DarkWindow GLSL shaders\b"),
        re.compile(r"\(terminal \+ ([0-9]+) shaders\)"),
    ],
    "ROADMAP.md": [
        re.compile(r"\b([0-9]+) DarkWindow GLSL shaders\b"),
    ],
    "docs/SPRINT-NEXT.md": [
        re.compile(r"\b([0-9]+)\+? DarkWindow GLSL shaders\b"),
    ],
    "docs/STATUS-BAR-RESEARCH.md": [
        re.compile(r"\b([0-9]+)\+? GLSL shaders\b"),
        re.compile(r"\b([0-9]+)-shader GLSL library\b"),
    ],
}

for rel, patterns in checks.items():
    path = Path(rel)
    if not path.exists():
        continue
    text = path.read_text(errors="ignore")
    for pattern in patterns:
        for match in pattern.finditer(text):
            claimed = int(match.group(1))
            if claimed != actual:
                errors.append(f"{rel}: claims {claimed} shaders but filesystem has {actual}")

if errors:
    for error in errors:
        print(f"DRIFT {error}")
    print(f"shader_count={actual} errors={len(errors)}")
    sys.exit(1)

print(f"shader_count={actual} errors=0")
PY
