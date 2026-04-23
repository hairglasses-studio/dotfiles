#!/usr/bin/env bash
set -euo pipefail

# validate-md-links.sh — walk tracked markdown files and verify every
# relative `[text](path.md)` or `[text](path.md#anchor)` link points
# at a real file.
#
# Scope:
#   - README.md, AGENTS.md, CLAUDE.md, GEMINI.md (top-level)
#   - docs/**/*.md
#   - .agents/skills/**/SKILL.md + references/**/*.md
#
# Ignores absolute links (http://, https://, mailto:, #anchor-only),
# since web targets aren't something we can resolve at edit time.
#
# Drift class: a rename or move that leaves a doc pointing at the
# old filename. The broken link renders as clickable text on GitHub
# and quietly 404s when someone follows it.
#
# Exit 0 on clean, 1 if any internal link targets a missing file.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

# Regex picks up standard Markdown links with a relative path ending
# in .md (optionally plus #anchor). Skips absolute URLs.
LINK_RE = re.compile(r"\]\(([^)]+?\.md)(#[^)]*)?\)")

def markdown_files() -> list[Path]:
    roots: list[Path] = []
    for top in ("README.md", "AGENTS.md", "CLAUDE.md", "GEMINI.md", "ROADMAP.md"):
        if Path(top).is_file():
            roots.append(Path(top))
    for base in ("docs", ".agents/skills"):
        base_path = Path(base)
        if not base_path.is_dir():
            continue
        for md in base_path.rglob("*.md"):
            roots.append(md)
    return roots

def is_external(target: str) -> bool:
    return bool(re.match(r"^[a-z]+://", target)) or target.startswith("mailto:")

errors: list[str] = []
files_scanned = 0
links_scanned = 0

for md_path in markdown_files():
    files_scanned += 1
    try:
        text = md_path.read_text(errors="ignore")
    except OSError:
        continue
    for match in LINK_RE.finditer(text):
        target = match.group(1).strip()
        if is_external(target):
            continue
        links_scanned += 1
        resolved = (md_path.parent / target).resolve()
        # Guard against links that escape the repo root
        try:
            resolved.relative_to(Path.cwd().resolve())
        except ValueError:
            errors.append(f"{md_path}: link escapes repo root: {target}")
            continue
        if not resolved.exists():
            errors.append(f"{md_path}: broken link -> {target}")

for err in errors:
    print(f"DRIFT: {err}")

print(f"files={files_scanned} links={links_scanned} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
