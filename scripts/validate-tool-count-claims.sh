#!/usr/bin/env bash
set -euo pipefail

# validate-tool-count-claims.sh — guard docs that claim a rough
# dotfiles-mcp tool count against the live snapshot.
#
# README.md, ROADMAP.md, mcp/dotfiles-mcp/CLAUDE.md, mcp/dotfiles-mcp/README.md,
# and publishing copy all have human-readable "~N tools" / "N tools" / badge
# figures. The contract snapshot at `mcp/dotfiles-mcp/.well-known/mcp.json` is
# the machine source of truth. If rough claims drift beyond a tolerance band
# (±15% default) the docs risk promising features that moved, or missing ones
# that landed. The exact canonical snapshot-count block in mcp/dotfiles-mcp/README.md
# must match `snapshots/contract/overview.json`.
#
# Exit 0 if every claim is within tolerance, 1 otherwise.
# TOLERANCE env var overrides the default (expressed as a decimal
# fraction, e.g. 0.2 for ±20%).

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import json
import os
import re
import sys
from pathlib import Path

TOLERANCE = float(os.environ.get("TOLERANCE", "0.15"))
SNAPSHOT = Path("mcp/dotfiles-mcp/.well-known/mcp.json")
OVERVIEW = Path("mcp/dotfiles-mcp/snapshots/contract/overview.json")
DOCS = [
    "README.md",
    "ROADMAP.md",
    "docs/publishing/mcp-directories.md",
    "mcp/dotfiles-mcp/CLAUDE.md",
    "mcp/dotfiles-mcp/README.md",
]

if not SNAPSHOT.exists():
    print(f"snapshot not found: {SNAPSHOT}", file=sys.stderr)
    sys.exit(2)

data = json.loads(SNAPSHOT.read_text())
overview = json.loads(OVERVIEW.read_text()) if OVERVIEW.exists() else {}
live = int(data.get("tool_count", len(data.get("tools", []))))
low = int(live * (1 - TOLERANCE))
high = int(live * (1 + TOLERANCE))

TEXT_PAT = re.compile(r"(?:~\s*)?([0-9][0-9,]{2,4})\+?\s+(?:live\s+)?tools")
BADGE_PAT = re.compile(r"MCP_Tools-([0-9][0-9,]{2,4})\+?")

claims = []
for doc in DOCS:
    p = Path(doc)
    if not p.exists():
        continue
    text = p.read_text(errors="ignore")
    for m in TEXT_PAT.finditer(text):
        claimed = int(m.group(1).replace(",", ""))
        claims.append((doc, claimed, low <= claimed <= high))
    for m in BADGE_PAT.finditer(text):
        claimed = int(m.group(1).replace(",", ""))
        claims.append((doc, claimed, low <= claimed <= high))

stale = [c for c in claims if not c[2]]
exact_errors = []
readme = Path("mcp/dotfiles-mcp/README.md")
if readme.exists():
    readme_text = readme.read_text(errors="ignore")
    exact_claims = {
        "tools": int(overview.get("total_tools", live)),
        "registered modules": int(overview.get("module_count", 0)),
        "resources": int(data.get("resource_count", overview.get("resource_count", 0))),
        "prompts": int(data.get("prompt_count", overview.get("prompt_count", 0))),
    }
    for label, expected in exact_claims.items():
        m = re.search(rf"- `([0-9][0-9,]*)` {re.escape(label)}", readme_text)
        if not m:
            exact_errors.append((str(readme), label, None, expected))
            continue
        claimed = int(m.group(1).replace(",", ""))
        if claimed != expected:
            exact_errors.append((str(readme), label, claimed, expected))

has_errors = False
if stale:
    print(
        f"live_tool_count={live} tolerance=±{TOLERANCE:.0%} "
        f"accepted_band=[{low}, {high}] stale_claims={len(stale)}"
    )
    for doc, claimed, _ in stale:
        drift = (claimed - live) / live
        print(f"  {doc} claims ~{claimed} tools (drift {drift:+.1%})")
    has_errors = True

if exact_errors:
    print(f"exact_snapshot_count_errors={len(exact_errors)}")
    for doc, label, claimed, expected in exact_errors:
        if claimed is None:
            print(f"  {doc} missing exact `{label}` count (expected {expected})")
        else:
            print(f"  {doc} claims {claimed} {label}; expected {expected}")
    has_errors = True

if has_errors:
    sys.exit(1)

print(
    f"live_tool_count={live} tolerance=±{TOLERANCE:.0%} "
    f"claims_scanned={len(claims)} exact_snapshot_counts=ok errors=0"
)
PY
