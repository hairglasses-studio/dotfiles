"""mx-battery — Logitech MX peripheral battery level from bar-mx cache.

Colors text pink when battery percentage falls below 20.
"""
from __future__ import annotations

import re
from html import escape

import ticker_render as tr

META = {"name": "mx-battery", "preset": None, "refresh": 300}

_LABEL = "\U000f0379 MX BATTERY"
_COLOR = "#ffe45e"


def build():
    try:
        with open("/tmp/bar-mx.txt") as f:
            raw = f.read().strip()
    except FileNotFoundError:
        return tr.empty(_LABEL, _COLOR, "mx cache missing")
    except OSError:
        return tr.empty(_LABEL, _COLOR, "mx battery unavailable")
    if not raw:
        return tr.empty(_LABEL, _COLOR, "no mx battery data")
    m = re.search(r"(\d+)", raw)
    pct = int(m.group(1)) if m else None
    color = "#ff5c8a" if pct is not None and pct < 20 else "#f7fbff"
    parts = [tr.badge(_LABEL, _COLOR)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{color}">'
        f'  {escape(raw)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
