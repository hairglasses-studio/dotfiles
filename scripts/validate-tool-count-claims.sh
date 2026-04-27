#!/usr/bin/env bash
set -euo pipefail

# validate-tool-count-claims.sh — guard docs that claim a rough
# dotfiles-mcp tool count against the live snapshot.
#
# README.md, ROADMAP.md, mcp/dotfiles-mcp/CLAUDE.md, and publishing copy
# all have human-readable "~N tools" / "N tools" / badge figures. The contract snapshot at
# `mcp/dotfiles-mcp/.well-known/mcp.json` is the machine source of
# truth. If the two drift beyond a tolerance band (±15% default)
# the docs risk promising features that moved, or missing ones
# that landed.
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
DOCS = [
    "README.md",
    "ROADMAP.md",
    "docs/publishing/mcp-directories.md",
    "mcp/dotfiles-mcp/CLAUDE.md",
]

if not SNAPSHOT.exists():
    print(f"snapshot not found: {SNAPSHOT}", file=sys.stderr)
    sys.exit(2)

data = json.loads(SNAPSHOT.read_text())
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
if stale:
    print(
        f"live_tool_count={live} tolerance=±{TOLERANCE:.0%} "
        f"accepted_band=[{low}, {high}] stale_claims={len(stale)}"
    )
    for doc, claimed, _ in stale:
        drift = (claimed - live) / live
        print(f"  {doc} claims ~{claimed} tools (drift {drift:+.1%})")
    sys.exit(1)

print(
    f"live_tool_count={live} tolerance=±{TOLERANCE:.0%} "
    f"claims_scanned={len(claims)} errors=0"
)
PY
