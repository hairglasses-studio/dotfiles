#!/usr/bin/env bash
set -euo pipefail

# validate-mcp-contracts.sh — consistency gate for MCP contract files.
#
# Checks:
#   1. mcp/dotfiles-mcp/.well-known/mcp.json — tool_count matches the
#      length of the tools array, every tool has a non-empty name and
#      description, tool names are unique.
#   2. mcp/mirror-parity.json — every mirror's canonical_path points
#      at a real directory in this repo, every consolidated handler points
#      at a real file, README.md lists only active modules as active rows,
#      and docs do not reference removed bundled MCP source paths.
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
import re
import sys
from pathlib import Path

WELL_KNOWN = Path("mcp/dotfiles-mcp/.well-known/mcp.json")
MIRROR_PARITY = Path("mcp/mirror-parity.json")
README = Path("README.md")
CONTRACT_OVERVIEW = Path("mcp/dotfiles-mcp/snapshots/contract/overview.json")
DOTFILES_MCP_README = Path("mcp/dotfiles-mcp/README.md")
DOTFILES_MCP_ROADMAP = Path("mcp/dotfiles-mcp/ROADMAP.md")
DOCS = [
    Path("README.md"),
    Path("docs"),
    Path("CLAUDE.md"),
    Path("GEMINI.md"),
    Path("AGENTS.md"),
]

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

# ── dotfiles-mcp README snapshot counts ──────────────────────────
if CONTRACT_OVERVIEW.is_file() and DOTFILES_MCP_README.is_file():
    overview = json.loads(CONTRACT_OVERVIEW.read_text())
    readme = DOTFILES_MCP_README.read_text(errors="ignore")
    readme_counts = {
        "tools": overview.get("total_tools"),
        "registered modules": overview.get("module_count"),
        "resources": overview.get("resource_count"),
        "prompts": overview.get("prompt_count"),
    }
    for label, expected in readme_counts.items():
        if expected is None:
            errors.append(f"{CONTRACT_OVERVIEW}: missing expected count for {label}")
            continue
        if not re.search(rf"-\s+`{re.escape(str(expected))}`\s+{re.escape(label)}\b", readme):
            errors.append(f"{DOTFILES_MCP_README}: snapshot count for {label!r} is not {expected}")
if CONTRACT_OVERVIEW.is_file() and DOTFILES_MCP_ROADMAP.is_file():
    overview = json.loads(CONTRACT_OVERVIEW.read_text())
    roadmap = DOTFILES_MCP_ROADMAP.read_text(errors="ignore")
    expected = (
        overview.get("total_tools"),
        overview.get("module_count"),
        overview.get("resource_count"),
        overview.get("prompt_count"),
    )
    if all(value is not None for value in expected):
        tools, modules, resources, prompts = expected
        pattern = (
            rf"`{tools}`\s+tools across `{modules}`\s+registered modules, "
            rf"plus `{resources}`\s+resources and `{prompts}`\s+prompts"
        )
        if not re.search(pattern, roadmap):
            errors.append(
                f"{DOTFILES_MCP_ROADMAP}: current-state snapshot counts do not match overview.json"
            )

# ── mirror-parity.json ────────────────────────────────────────────
if MIRROR_PARITY.is_file():
    parity = json.loads(MIRROR_PARITY.read_text())
    active_modules: set[str] = set()
    retired_modules: set[str] = set()
    for mirror in parity.get("mirrors", []):
        path = mirror.get("canonical_path")
        module = mirror.get("module", "?")
        active_modules.add(module)
        if not path:
            errors.append(f"{MIRROR_PARITY}: mirror {module!r} missing canonical_path")
            continue
        if not Path(path).is_dir():
            errors.append(
                f"{MIRROR_PARITY}: mirror {module!r} canonical_path={path!r} "
                f"is not a directory"
            )
    for retired in parity.get("consolidated", []):
        module = retired.get("module", "?")
        retired_modules.add(module)
        handler = retired.get("handler")
        if not handler:
            errors.append(f"{MIRROR_PARITY}: consolidated module {module!r} missing handler")
            continue
        if not Path(handler).is_file():
            errors.append(
                f"{MIRROR_PARITY}: consolidated module {module!r} handler={handler!r} "
                f"is not a file"
            )
    if README.is_file():
        readme = README.read_text(errors="ignore")
        for module in sorted(active_modules):
            if not re.search(rf"^\|\s*`{re.escape(module)}`\s*\|", readme, re.M):
                errors.append(f"{README}: active MCP module {module!r} missing from table")
        for module in sorted(retired_modules):
            if re.search(rf"^\|\s*`{re.escape(module)}`\s*\|", readme, re.M):
                errors.append(
                    f"{README}: retired MCP module {module!r} is still listed as an active table row"
                )
else:
    print(f"skip: {MIRROR_PARITY} missing (no drift to check)", file=sys.stderr)

# ── removed bundled MCP path references ──────────────────────────
forbidden_paths = [
    "mcp/shader-mcp",
    "mcp/hyprland-mcp",
    "mcp/sway-mcp",
    "mcp/systemd-mcp",
    "mcp/tmux-mcp",
    "mcp/process-mcp",
]
for root in DOCS:
    paths: list[Path]
    if root.is_dir():
        paths = [p for p in root.rglob("*.md") if ".git" not in p.parts]
    elif root.is_file():
        paths = [root]
    else:
        continue
    for path in paths:
        text = path.read_text(errors="ignore")
        for forbidden in forbidden_paths:
            if forbidden in text:
                errors.append(f"{path}: references removed bundled MCP path {forbidden!r}")

for err in errors:
    print(f"DRIFT: {err}")

print(f"errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
