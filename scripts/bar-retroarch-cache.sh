#!/usr/bin/env bash
# bar-retroarch-cache.sh — Cache writer for the Ironbar RetroArch widget
#
# Reads the most recent entry from RetroArch's content_history.lpl and
# the latest mounts-audit report, and writes a label to /tmp/bar-retroarch.txt.
# Ironbar reads the cache file with `cat`, never blocking on JSON parse
# or filesystem walks.
#
# Cache format: "<icon> <title> <core-badge> <mount-badge>"  (empty if no history)
#   icon        — 󰊗 (controller glyph)
#   title       — filename stem with region tags stripped
#   core-badge  — short core name in brackets, e.g. [LRPS2]
#   mount-badge — " ⎈N/M" when mounts exist (ok/total); empty otherwise
#
# Paired with bar-retroarch.timer (30 s).

set -euo pipefail

CACHE_FILE="/tmp/bar-retroarch.txt"
HISTORY_PATH="${HOME}/.config/retroarch/playlists/builtin/content_history.lpl"
MOUNTS_REPORT="${XDG_STATE_HOME:-$HOME/.local/state}/retroarch/mounts-audit.json"

TMPFILE="$(mktemp /tmp/bar-retroarch.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

label=""
if [[ -f "$HISTORY_PATH" ]]; then
    label=$(python3 - "$HISTORY_PATH" "$MOUNTS_REPORT" <<'PY'
import json
import sys
from pathlib import Path

history_path = Path(sys.argv[1])
mounts_path = Path(sys.argv[2])

try:
    data = json.loads(history_path.read_text())
except (OSError, json.JSONDecodeError):
    sys.exit(0)

items = data.get("items") or []
if not items:
    sys.exit(0)

item = items[0]
raw_label = (item.get("label") or "").strip()
if not raw_label:
    # Derive from filename stem, stripping archive-inner suffix + (USA)/(EU) tags
    path = (item.get("path") or "").strip()
    if not path:
        sys.exit(0)
    stem = Path(path.split("#", 1)[0]).stem
    # Strip the rightmost parenthesized region/rev tag, e.g. "Game (USA)" → "Game"
    import re
    while True:
        new = re.sub(r"\s*\([^)]*\)$", "", stem)
        if new == stem:
            break
        stem = new
    raw_label = stem

# Core badge — short name in brackets
core_name = (item.get("core_name") or "").strip()
core_short = ""
if core_name:
    # Take the parenthesized last token if present: "Sony - PS2 (LRPS2)" → "LRPS2"
    import re
    m = re.search(r"\(([^)]+)\)\s*$", core_name)
    if m:
        core_short = m.group(1)
    else:
        core_short = core_name.split("-")[-1].strip()

# Mount badge — read mounts-audit.json if present
mount_badge = ""
try:
    mounts = json.loads(mounts_path.read_text())
    s = mounts.get("summary") or {}
    total = s.get("total")
    active = s.get("active")
    if total is not None and total > 0:
        if active == total:
            mount_badge = f" ⎈{active}/{total}"
        else:
            mount_badge = f" ⎈{active}/{total}!"
except (OSError, json.JSONDecodeError, KeyError):
    pass

bits = [f"\U000F0297 {raw_label}"]  # 󰊗 gamepad-variant
if core_short:
    bits.append(f"[{core_short}]")
if mount_badge:
    bits[-1] = bits[-1] + mount_badge
elif len(bits) == 1:
    # no core, so append mount_badge directly to first bit
    bits[0] = bits[0] + mount_badge

print(" ".join(bits).strip())
PY
)
fi

if [[ -n "$label" ]]; then
    printf '%s\n' "$label" > "$TMPFILE"
else
    printf '' > "$TMPFILE"
fi
mv "$TMPFILE" "$CACHE_FILE"
