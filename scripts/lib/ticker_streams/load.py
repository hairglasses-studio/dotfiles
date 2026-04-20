"""load — /proc/loadavg with sparkline + task count."""
from __future__ import annotations

import os
from html import escape

import ticker_render as tr

META = {"name": "load", "preset": None, "refresh": 5}

_LABEL = "\U000f046a LOAD"
_BARS = "\u2581\u2582\u2583\u2584\u2585\u2586\u2587\u2588"


def _sparkline(values):
    if not values:
        return ""
    vmax = max(values) or 1
    out = ""
    for v in values:
        idx = min(len(_BARS) - 1, int((v / vmax) * (len(_BARS) - 1)))
        out += _BARS[idx]
    return out


def build():
    parts = [tr.badge(_LABEL, "#3dffb5")]
    try:
        with open("/proc/loadavg") as f:
            fields = f.read().split()
    except Exception:
        return tr.empty(_LABEL, "#3dffb5", "load unavailable")
    if len(fields) < 3:
        return tr.empty(_LABEL, "#3dffb5", "no load data")
    try:
        one, five, fifteen = float(fields[0]), float(fields[1]), float(fields[2])
    except ValueError:
        return tr.empty(_LABEL, "#3dffb5", "load parse failed")
    spark = _sparkline([fifteen, five, one])
    ncpu = os.cpu_count() or 1
    color = "#ff5c8a" if one > ncpu else "#f7fbff"
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="{color}">'
        f'  {spark} 1m={one:.2f} 5m={five:.2f} 15m={fifteen:.2f}  '
        f'(cpus={ncpu})  \u00b7</span>'
    )
    running = fields[3] if len(fields) > 3 else "?"
    parts.append(
        f'<span font_desc="Maple Mono NF CN 11">'
        f'  {escape(running)} tasks  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
