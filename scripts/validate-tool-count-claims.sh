#!/usr/bin/env bash
set -euo pipefail

# validate-tool-count-claims.sh — guard docs that claim a rough
# dotfiles-mcp tool count against the live snapshot.
#
# README.md, ROADMAP.md, and mcp/dotfiles-mcp/CLAUDE.md all have
# human-readable "~N tools" figures. The contract snapshot at
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
    "mcp/dotfiles-mcp/CLAUDE.md",
]

if not SNAPSHOT.exists():
    print(f"snapshot not found: {SNAPSHOT}", file=sys.stderr)
    sys.exit(2)

data = json.loads(SNAPSHOT.read_text())
live = int(data.get("tool_count", len(data.get("tools", []))))
low = int(live * (1 - TOLERANCE))
high = int(live * (1 + TOLERANCE))

# Match "~NNN tools" (three digits so we don't grab arbitrary small
# counts). Capture group 1 is the approximate count.
pat = re.compile(r"~([0-9]{3,4}) tools")

claims = []
for doc in DOCS:
    p = Path(doc)
    if not p.exists():
        continue
    for m in pat.finditer(p.read_text(errors="ignore")):
        claimed = int(m.group(1))
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
