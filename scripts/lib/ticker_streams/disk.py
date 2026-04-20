"""disk — filesystem usage via `df -h` with red alert over 80%."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "disk", "preset": None, "refresh": 60}

_LABEL = "\U000f02ca DISK"
_SKIP_PREFIXES = ("/dev", "/sys", "/proc", "/run", "/tmp", "/boot/efi")


def build():
    parts = [tr.badge(_LABEL, "#4aa8ff")]
    try:
        raw = subprocess.run(
            ["df", "-h", "--output=target,pcent,avail"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip().splitlines()
    except Exception:
        return tr.empty(_LABEL, "#4aa8ff", "disk unavailable")
    fc = len(FONTS)
    shown = 0
    for line in raw[1:]:
        fields = line.split()
        if len(fields) < 3:
            continue
        mount, pct_s, avail = fields[0], fields[1], fields[2]
        if mount.startswith(_SKIP_PREFIXES):
            continue
        if mount in ("/", "/home") or mount.startswith("/mnt") or mount.startswith("/data"):
            try:
                pct_num = int(pct_s.rstrip("%"))
            except ValueError:
                continue
            color = "#ff5c8a" if pct_num > 80 else "#f7fbff"
            font = FONTS[shown % fc]
            parts.append(
                f'<span font_desc="{font}" foreground="{color}">'
                f'  {escape(mount)} {escape(pct_s)} ({escape(avail)} free)  \u00b7</span>'
            )
            shown += 1
    if shown == 0:
        return tr.empty(_LABEL, "#4aa8ff", "no filesystems")
    return tr.dup("".join(parts)), []
