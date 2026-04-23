#!/usr/bin/env bash
set -euo pipefail

# validate-rclone-mount-units.sh — guard that every
# systemd/retroarch-rclone-*.service unit stays on-spec with its
# siblings. The 5 units share a common shape (same rclone flags,
# same mount-dir parent, same ExecStop tool) that drifts silently
# when someone adds a new mount without copying all the details —
# or edits one unit's caching settings while leaving the others
# behind.
#
# Rules enforced per unit:
#   - mount path under ~/Games/RetroArch/mounts/
#   - --read-only flag present (gdrive mounts are RO by policy)
#   - --dir-cache-time flag present (avoids unbounded memory use)
#   - --vfs-cache-mode off (writes not needed; caching just pins
#     files in RAM)
#   - ExecStop uses fusermount3 -u
#   - PartOf=retroarch-rclone.target
#   - WantedBy=retroarch-rclone.target in [Install]
#   - uses %h specifier (not hardcoded /home/<user>)
#
# Exit 0 on clean, 1 if any unit drifts.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

UNIT_DIR = Path("systemd")
RE_HARDCODED_HOME = re.compile(r"/home/[A-Za-z0-9_-]+/")

REQUIRED_FLAGS = ["--read-only", "--dir-cache-time", "--vfs-cache-mode off"]

errors: list[str] = []
units = sorted(UNIT_DIR.glob("retroarch-rclone-*.service"))

if not units:
    print("skip: no retroarch-rclone units tracked")
    sys.exit(0)

for unit in units:
    text = unit.read_text()
    name = unit.name

    # mount path under ~/Games/RetroArch/mounts/
    if "%h/Games/RetroArch/mounts/" not in text:
        errors.append(f"{name}: mount target not under %h/Games/RetroArch/mounts/")

    for flag in REQUIRED_FLAGS:
        if flag not in text:
            errors.append(f"{name}: missing rclone flag {flag!r}")

    if "fusermount3 -u" not in text:
        errors.append(f"{name}: ExecStop should use fusermount3 -u")

    if "PartOf=retroarch-rclone.target" not in text:
        errors.append(f"{name}: missing PartOf=retroarch-rclone.target")

    if "WantedBy=retroarch-rclone.target" not in text:
        errors.append(f"{name}: missing WantedBy=retroarch-rclone.target in [Install]")

    # Hardcoded /home/<user>/ should never appear — use %h.
    hits = RE_HARDCODED_HOME.findall(text)
    # Filter out allowed absolute binaries under /usr/bin, /usr/lib
    if hits:
        errors.append(f"{name}: hardcoded /home/... path present (use %h)")

# Verify the target file exists + reflects all 5 services.
target = UNIT_DIR / "retroarch-rclone.target"
if target.is_file():
    ttext = target.read_text()
    for unit in units:
        want = f"Wants={unit.name}"
        after = f"After={unit.name}"
        if want not in ttext:
            errors.append(f"retroarch-rclone.target: missing {want}")
        if after not in ttext:
            errors.append(f"retroarch-rclone.target: missing {after}")
else:
    errors.append("systemd/retroarch-rclone.target missing")

for err in errors:
    print(f"DRIFT: {err}")

print(f"units={len(units)} errors={len(errors)}")
sys.exit(1 if errors else 0)
PY
