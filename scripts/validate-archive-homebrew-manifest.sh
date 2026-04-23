#!/usr/bin/env bash
set -euo pipefail

# validate-archive-homebrew-manifest.sh — sanity-check every entry in
# scripts/lib/retroarch_archive_homebrew_verified.json against the
# SourceItem registry in scripts/retroarch-archive-homebrew-manifest.py.
#
# Drift vectors this guards:
#   - source_identifier that doesn't match any SourceItem
#   - archive_path whose leading directory is not in the
#     matching SourceItem's system_dirs map (orchestrator silently
#     skips these)
#   - tier that isn't one of the known tiers the manifest classifier
#     emits
#
# The audit and orchestrator happily proceed past these errors at
# runtime — entries just never produce playlist rows. Catch at edit
# time instead.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import ast
import json
import sys
from pathlib import Path

VERIFIED = Path("scripts/lib/retroarch_archive_homebrew_verified.json")
MANIFEST = Path("scripts/retroarch-archive-homebrew-manifest.py")

# Extract SOURCES from the manifest via AST — avoids importing the
# script (which would require its own sys.path gymnastics) and avoids
# pulling in a regex parser that drifts with source reformatting.
tree = ast.parse(MANIFEST.read_text())
sources: dict[str, dict[str, str]] = {}
for node in ast.walk(tree):
    # SOURCES may be declared as plain `SOURCES = [...]` (ast.Assign)
    # or annotated `SOURCES: list[SourceItem] = [...]` (ast.AnnAssign).
    if isinstance(node, ast.Assign):
        targets = [t for t in node.targets if isinstance(t, ast.Name) and t.id == "SOURCES"]
        if not targets:
            continue
        value = node.value
    elif isinstance(node, ast.AnnAssign):
        if not (isinstance(node.target, ast.Name) and node.target.id == "SOURCES"):
            continue
        value = node.value
    else:
        continue
    if not isinstance(value, ast.List):
        continue
    for item in value.elts:
        if not isinstance(item, ast.Call):
            continue
        kwargs = {kw.arg: kw.value for kw in item.keywords}
        identifier_node = kwargs.get("identifier")
        system_dirs_node = kwargs.get("system_dirs")
        if not isinstance(identifier_node, ast.Constant):
            continue
        identifier = identifier_node.value
        dirs: dict[str, str] = {}
        if isinstance(system_dirs_node, ast.Dict):
            for k, v in zip(system_dirs_node.keys, system_dirs_node.values):
                if isinstance(k, ast.Constant) and isinstance(v, ast.Constant):
                    dirs[k.value] = v.value
        sources[identifier] = dirs
    break

if not sources:
    print("ERROR: could not extract SOURCES from retroarch-archive-homebrew-manifest.py",
          file=sys.stderr)
    sys.exit(2)

# Known tiers the manifest classifier emits — see _classify_entry
# in retroarch-archive-homebrew-manifest.py.
KNOWN_TIERS = {
    "public_domain",
    "verified_redistributable",
    "homebrew_unverified",
    "utility_unverified",
}

data = json.loads(VERIFIED.read_text())
entries = data.get("entries", [])

errors: list[str] = []
for i, entry in enumerate(entries):
    label = entry.get("archive_path", f"<index {i}>")
    src_id = entry.get("source_identifier")
    archive_path = entry.get("archive_path", "")
    tier = entry.get("tier")

    if src_id not in sources:
        errors.append(f"{label}: unknown source_identifier={src_id!r} "
                      f"(valid: {sorted(sources)})")
        continue

    if tier not in KNOWN_TIERS:
        errors.append(f"{label}: unknown tier={tier!r}")

    if "/" not in archive_path:
        errors.append(f"{label}: archive_path missing system directory prefix")
        continue

    display_name = archive_path.split("/", 1)[0]
    if display_name not in sources[src_id]:
        errors.append(
            f"{label}: archive_path prefix {display_name!r} not in "
            f"source {src_id!r} system_dirs "
            f"(valid: {sorted(sources[src_id])})"
        )

for err in errors:
    print(f"DRIFT: {err}")

print(f"sources={len(sources)} entries={len(entries)} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
