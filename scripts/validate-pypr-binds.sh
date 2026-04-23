#!/usr/bin/env bash
set -euo pipefail

# validate-pypr-binds.sh — guard that every `pypr toggle <name>`
# invocation in hyprland.conf resolves to a declared scratchpad
# in pypr/config.toml.
#
# hyprland binds like `exec, pypr toggle dev-console` depend on the
# scratchpad existing in pyprland's own config. A rename or removal
# leaves the keybind dispatching to pypr which politely does
# nothing — no dialog, no log, no feedback at the compositor level.
#
# Handles:
#   - `pypr toggle <name>` (scratchpad toggle)
#   - `pypr <builtin>` where builtin is a known pypr command
#     (expose, zoom, lost_windows) — skipped from verification
#
# Exit 0 on clean, 1 if any bind points at an undeclared scratchpad.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

HYPR_CONF = Path("hyprland/hyprland.conf")
PYPR_CONF = Path("pypr/config.toml")

# pypr ships these as built-in plugins — binds like `pypr expose`
# do not need a scratchpads.<name> block. Drawn from pypr docs.
PYPR_BUILTINS = {
    "expose",
    "zoom",
    "lost_windows",
    "fetch_client_menu",
    "relative_workspaces",
    "monitors",
    "shift_monitors",
    "workspaces_follow_focus",
    "magnify",
    "layout_center",
    "shortcuts_menu",
}

if not HYPR_CONF.is_file() or not PYPR_CONF.is_file():
    # Repo without these files should skip, not fail.
    print("skip: hyprland.conf or pypr/config.toml missing")
    sys.exit(0)

declared = set(
    m.group(1)
    for m in re.finditer(r"^\[scratchpads\.([a-z0-9_-]+)\]", PYPR_CONF.read_text(), re.MULTILINE)
)

# Matches `pypr toggle <name>` or `pypr <builtin>` (with optional
# trailing argument like `pypr expose`).
invoke_re = re.compile(r"\bpypr\s+(toggle\s+([a-z0-9_-]+)|([a-z0-9_-]+))")

errors: list[str] = []
invocations = 0
for match in invoke_re.finditer(HYPR_CONF.read_text()):
    invocations += 1
    if match.group(2):
        target = match.group(2)
        if target not in declared:
            errors.append(
                f"hyprland.conf: `pypr toggle {target}` has no "
                f"[scratchpads.{target}] in pypr/config.toml"
            )
    else:
        builtin = match.group(3)
        # "toggle" by itself means malformed, flag it
        if builtin == "toggle":
            errors.append("hyprland.conf: `pypr toggle` with no target name")
        elif builtin in PYPR_BUILTINS or builtin in declared:
            # Either known built-in or also matches a scratchpad name.
            pass
        else:
            errors.append(
                f"hyprland.conf: `pypr {builtin}` is not a known built-in "
                f"and not declared as a scratchpad"
            )

for err in errors:
    print(f"DRIFT: {err}")

print(f"invocations={invocations} scratchpads={len(declared)} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
