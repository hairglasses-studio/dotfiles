"""cve-alerts — arch-audit CVE advisories with severity-graded badge.

Cache is written by bar-cve-cache.sh. When arch-audit is not installed
the writer produces an `install arch-audit ...` hint as the first line,
which we render with an amber badge instead of a severity colour.
"""
from __future__ import annotations

from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "cve-alerts", "preset": "cyberpunk", "refresh": 3600}

_LABEL = "\U000f0ce4 CVE"


def _read_lines():
    try:
        with open("/tmp/bar-cve.txt") as f:
            return [l.rstrip("\n") for l in f if l.strip()]
    except FileNotFoundError:
        return None
    except OSError:
        return []


def build():
    lines = _read_lines()
    if lines is None:
        return tr.empty(_LABEL, "#ef4444", "cve cache missing")
    if not lines:
        return tr.empty(_LABEL, "#34d399", "no advisories")
    first = lines[0]
    if first.startswith("install arch-audit"):
        parts = [tr.badge(_LABEL, "#f59e0b")]
        parts.append(
            f'<span font_desc="Maple Mono NF CN 11" foreground="#fbbf24">'
            f'  {escape(first)}  \u00b7</span>'
        )
        return tr.dup("".join(parts)), []
    any_critical = any("Critical" in l for l in lines[1:])
    any_high = any("High" in l for l in lines[1:])
    color = "#dc2626" if any_critical else ("#f97316" if any_high else "#f59e0b")
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="#fca5a5">'
        f'  {escape(first)}  \u00b7</span>'
    )
    fc = len(FONTS)
    for i, line in enumerate(lines[1:7]):
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="#fecaca">'
            f'  {escape(line)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
