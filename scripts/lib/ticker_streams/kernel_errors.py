"""kernel-errors — recent kernel-level journal errors (last 15 min)."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "kernel-errors", "preset": "cyberpunk", "refresh": 60}

_LABEL = "\U000f00e4 KERNEL"
_RED = "#dc2626"


def build():
    try:
        out = subprocess.run(
            ["journalctl", "-p", "err", "-k", "-n", "5",
             "--since", "15 min ago", "--output=short-precise", "--no-pager"],
            capture_output=True, text=True, timeout=3,
        ).stdout
    except Exception:
        return tr.empty(_LABEL, _RED, "journalctl unavailable")
    lines = [l for l in out.splitlines() if l.strip()]
    if not lines:
        return tr.empty(_LABEL, "#34d399", "no recent errors")
    seen: set[str] = set()
    uniq: list[str] = []
    for l in lines:
        msg = l.split("kernel: ", 1)[-1][:90]
        if msg not in seen:
            seen.add(msg)
            uniq.append(msg)
    parts = [tr.badge(_LABEL, _RED)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="#ef4444">'
        f'  {len(uniq)} err  \u00b7</span>'
    )
    fc = len(FONTS)
    for i, msg in enumerate(uniq[:4]):
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="#fca5a5">'
            f'  {escape(msg)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
