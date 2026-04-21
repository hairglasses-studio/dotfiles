#!/usr/bin/env bash
set -euo pipefail

# validate-python-imports.sh — attempt to import every Python module
# under scripts/lib/ and scripts/lib/ticker_streams/.
#
# py_compile only catches syntax-level errors. Import attempts catch
# the next layer:
#   - module-scope NameError (typo referencing an undeclared name)
#   - ModuleNotFoundError (typo in `import`)
#   - SyntaxError at any defer-imported submodule
#
# Covers library modules only; top-level scripts/*.py files are
# scripts that take argparse entrypoints and may have side effects at
# module load. py_compile already exercises those at parse time.
#
# Exit 0 if every module imports cleanly, 1 otherwise.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import importlib.util
import sys
from pathlib import Path

# Ticker plugins import sibling helpers (ticker_render, etc.) by bare
# name — they're designed to be loaded with scripts/lib/ on sys.path.
# Put both library roots on sys.path up front so the loader resolves
# "from ticker_render import …" the same way the ticker process does.
sys.path.insert(0, str(Path("scripts/lib").resolve()))
sys.path.insert(0, str(Path("scripts/lib/ticker_streams").resolve()))

TARGETS = [
    Path("scripts/lib"),
    Path("scripts/lib/ticker_streams"),
]

broken = []
imported = 0
for base in TARGETS:
    if not base.is_dir():
        continue
    for p in sorted(base.glob("*.py")):
        if p.name == "__init__.py":
            continue
        spec = importlib.util.spec_from_file_location(p.stem, p)
        mod = importlib.util.module_from_spec(spec)
        try:
            spec.loader.exec_module(mod)
            imported += 1
        except Exception as exc:
            broken.append((str(p), type(exc).__name__, str(exc)[:120]))

if broken:
    print(f"imported={imported} broken={len(broken)}")
    for path, err_type, msg in broken:
        print(f"  {path}: {err_type}: {msg}")
    sys.exit(1)

print(f"imported={imported} errors=0")
PY
