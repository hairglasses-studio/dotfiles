"""container-status — running docker containers with health-graded badge."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "container-status", "preset": None, "refresh": 30}

_LABEL = "\U000f08c1 CONTAINERS"
_BAD_MARKERS = ("Restarting", "Exited", "unhealthy")


def build():
    try:
        out = subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}\t{{.Status}}"],
            capture_output=True, text=True, timeout=3,
        )
    except Exception:
        return tr.empty(_LABEL, "#0db7ed", "docker unavailable")
    if out.returncode != 0:
        return tr.empty(_LABEL, "#0db7ed", "docker daemon down")
    rows = [l for l in out.stdout.splitlines() if l.strip()]
    if not rows:
        return tr.empty(_LABEL, "#66708f", "no containers")
    any_bad = any(any(m in r for m in _BAD_MARKERS) for r in rows)
    color = "#ef4444" if any_bad else "#0db7ed"
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11">'
        f'  {len(rows)} running  \u00b7</span>'
    )
    fc = len(FONTS)
    for i, row in enumerate(rows[:8]):
        try:
            name, status = row.split("\t", 1)
        except ValueError:
            continue
        font = FONTS[i % fc]
        name_c = "#ef4444" if any(m in status for m in _BAD_MARKERS) else "#34d399"
        parts.append(
            f'<span font_desc="{font}" foreground="{name_c}">'
            f'  {escape(name)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
