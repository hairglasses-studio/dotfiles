#!/usr/bin/env bash
set -euo pipefail

# validate-mcp-contracts.sh — consistency gate for MCP contract files.
#
# Checks:
#   1. mcp/dotfiles-mcp/.well-known/mcp.json — tool_count matches the
#      length of the tools array, every tool has a non-empty name and
#      description, tool names are unique.
#   2. mcp/mirror-parity.json — every mirror's canonical_path points
#      at a real directory in this repo.
#
# Drift this catches:
#   - Adding or removing a tool from the well-known manifest without
#     bumping tool_count (the public claim about server surface falls
#     out of sync with reality).
#   - Renaming an MCP module without updating mirror-parity.json (the
#     parity checker would then try to compare against a path that
#     no longer exists).
#
# Exit 0 on clean, 1 on any drift.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import json
import sys
from pathlib import Path

WELL_KNOWN = Path("mcp/dotfiles-mcp/.well-known/mcp.json")
MIRROR_PARITY = Path("mcp/mirror-parity.json")

errors: list[str] = []

# ── .well-known/mcp.json ─────────────────────────────────────────
if WELL_KNOWN.is_file():
    data = json.loads(WELL_KNOWN.read_text())
    tools = data.get("tools", [])
    declared = data.get("tool_count")
    actual = len(tools)
    if declared is None:
        errors.append(f"{WELL_KNOWN}: missing tool_count field")
    elif declared != actual:
        errors.append(
            f"{WELL_KNOWN}: tool_count={declared} but len(tools)={actual} — "
            f"public surface claim out of sync"
        )
    seen: set[str] = set()
    for i, tool in enumerate(tools):
        name = tool.get("name", "")
        desc = tool.get("description", "")
        if not name:
            errors.append(f"{WELL_KNOWN}: tool[{i}] missing name")
            continue
        if not desc:
            errors.append(f"{WELL_KNOWN}: tool {name!r} missing description")
        if name in seen:
            errors.append(f"{WELL_KNOWN}: duplicate tool name {name!r}")
        seen.add(name)
else:
    print(f"skip: {WELL_KNOWN} missing (no drift to check)", file=sys.stderr)

# ── mirror-parity.json ────────────────────────────────────────────
if MIRROR_PARITY.is_file():
    parity = json.loads(MIRROR_PARITY.read_text())
    for mirror in parity.get("mirrors", []):
        path = mirror.get("canonical_path")
        module = mirror.get("module", "?")
        if not path:
            errors.append(f"{MIRROR_PARITY}: mirror {module!r} missing canonical_path")
            continue
        if not Path(path).is_dir():
            errors.append(
                f"{MIRROR_PARITY}: mirror {module!r} canonical_path={path!r} "
                f"is not a directory"
            )
else:
    print(f"skip: {MIRROR_PARITY} missing (no drift to check)", file=sys.stderr)

for err in errors:
    print(f"DRIFT: {err}")

print(f"errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
