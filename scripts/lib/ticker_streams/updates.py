"""updates — /tmp/bar-updates.txt summary + top-N package names via `checkupdates`.

The summary is written by `bar-updates-cache.sh` (systemd-timer driven).
The per-package list spawns `checkupdates` which blocks on pacman DB
lookups — so the stream is marked slow and runs on the background thread.
"""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "updates", "preset": None, "refresh": 1800, "slow": True}

_LABEL = "\U000f0f8c UPDATES"


def build():
    parts = [tr.badge(_LABEL, "#29f0ff")]
    try:
        with open("/tmp/bar-updates.txt") as f:
            raw = f.read().strip()
    except FileNotFoundError:
        return tr.empty(_LABEL, "#29f0ff", "updates cache missing")
    except Exception:
        return tr.empty(_LABEL, "#29f0ff", "updates unavailable")

    if not raw:
        return tr.empty(_LABEL, "#29f0ff", "no updates")
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15">'
        f'  {escape(raw)}  \u00b7</span>'
    )

    try:
        pkgs = subprocess.run(
            ["checkupdates"], capture_output=True, text=True, timeout=3,
        ).stdout.strip().splitlines()
        fc = len(FONTS)
        for i, line in enumerate(pkgs[:15]):
            name = escape(line.split()[0]) if line else ""
            if name:
                font = FONTS[i % fc]
                parts.append(
                    f'<span font_desc="{font}">  {name}  \u00b7</span>'
                )
    except Exception:
        pass
    return tr.dup("".join(parts)), []
