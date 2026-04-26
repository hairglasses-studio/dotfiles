#!/usr/bin/env bash
set -euo pipefail

# validate-local-bin-refs.sh — guard that every `$HOME/.local/bin/<name>`
# reference in compositor, launcher, and bar configs resolves to a
# destination install.sh --print-link-specs declares.
#
# Catches the case where install.sh removes (or renames) a symlink but
# a hyprland/pypr config still invokes the old name — the keybind silently
# no-ops (or launches a literal "command not found" shell error on
# hyprctl dispatch).
#
# Exit 0 on clean, 1 if any reference points outside the install.sh
# destination set.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import subprocess
import sys
from pathlib import Path

configs = [
    "hyprland/hyprland.conf",
    "pypr/config.toml",
]
ref_pat = re.compile(r"\$HOME/\.local/bin/([a-zA-Z0-9._-]+)")

refs: dict[str, set[str]] = {}
for cfg in configs:
    path = Path(cfg)
    if not path.exists():
        continue
    found = set(ref_pat.findall(path.read_text(errors="ignore")))
    if found:
        refs[cfg] = found

# install.sh --print-link-specs produces `SRC|DEST` lines; destinations
# live under $HOME/.local/bin when install.sh exposes a PATH wrapper.
result = subprocess.run(
    ["bash", "install.sh", "--print-link-specs"],
    capture_output=True,
    text=True,
    check=False,
)
dests: set[str] = set()
for line in result.stdout.splitlines():
    if "|" not in line:
        continue
    _, dest = line.split("|", 1)
    if "/.local/bin/" in dest:
        dests.add(Path(dest).name)

missing: dict[str, set[str]] = {}
for cfg, names in refs.items():
    gap = names - dests
    if gap:
        missing[cfg] = gap

total_refs = sum(len(v) for v in refs.values())
if missing:
    print(f"configs_scanned={len(refs)} total_refs={total_refs} "
          f"install_destinations={len(dests)} broken={sum(len(v) for v in missing.values())}")
    for cfg, names in missing.items():
        print(f"  {cfg}:")
        for name in sorted(names):
            print(f"    MISSING: $HOME/.local/bin/{name}")
    sys.exit(1)

print(
    f"configs_scanned={len(refs)} total_refs={total_refs} "
    f"install_destinations={len(dests)} errors=0"
)
PY
