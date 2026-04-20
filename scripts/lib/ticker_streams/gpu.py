"""gpu — `/tmp/bar-gpu-full.txt` nvidia-smi snapshot with threshold colours.

Ring-buffers the last N util readings in module scope so the rendered
sparkline actually varies over time. Previously the spark was built
from six copies of the current reading (a flat bar); Phase 5 of the
marathon fixes that.
"""
from __future__ import annotations

from collections import deque
from html import escape

import ticker_render as tr

META = {"name": "gpu", "preset": "cyberpunk", "refresh": 10}

_LABEL = "\U000f0a2d GPU"
_BARS = "\u2581\u2582\u2583\u2584\u2585\u2586\u2587\u2588"

# History ring — one entry per build() call (i.e. per stream refresh).
# At refresh=10s, a capacity of 12 gives a 2-minute window.
_UTIL_HISTORY: deque[float] = deque(maxlen=12)


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
    parts = [tr.badge(_LABEL, "#76ff03")]
    try:
        with open("/tmp/bar-gpu-full.txt") as f:
            raw = f.read().strip()
    except FileNotFoundError:
        return tr.empty(_LABEL, "#76ff03", "gpu cache missing")
    except Exception:
        return tr.empty(_LABEL, "#76ff03", "gpu unavailable")

    if not raw or "|" not in raw:
        return tr.empty(_LABEL, "#76ff03", "no gpu data")
    fields = raw.split("|", 5)
    if len(fields) < 6:
        return tr.empty(_LABEL, "#76ff03", "gpu cache short")
    power, util, temp, mem_used, mem_total, name = fields
    try:
        util_n = int(util)
        temp_n = int(temp)
        mem_used_n = int(mem_used)
        mem_total_n = int(mem_total)
    except ValueError:
        return tr.empty(_LABEL, "#76ff03", "gpu cache parse")

    util_color = "#ff5c8a" if util_n > 85 else ("#ffe45e" if util_n > 50 else "#76ff03")
    temp_color = "#ff5c8a" if temp_n >= 80 else ("#ffe45e" if temp_n >= 70 else "#76ff03")
    mem_pct = (mem_used_n / mem_total_n * 100) if mem_total_n else 0
    mem_color = "#ff5c8a" if mem_pct > 90 else ("#ffe45e" if mem_pct > 70 else "#76ff03")
    # Append this reading to the ring; sparkline renders over the last N
    # samples so the glyph actually shows GPU history instead of a flat bar.
    _UTIL_HISTORY.append(util_n / 100.0)
    spark = _sparkline(list(_UTIL_HISTORY))
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{util_color}">'
        f'  {spark} {util_n}% util  \u00b7</span>'
        f'<span font_desc="Maple Mono NF CN 15" foreground="{temp_color}">'
        f'  {temp_n}\u00b0C  \u00b7</span>'
        f'<span font_desc="Maple Mono NF CN 15" foreground="#7ad0ff">'
        f'  {power}W  \u00b7</span>'
        f'<span font_desc="Maple Mono NF CN 15" foreground="{mem_color}">'
        f'  {mem_used_n}/{mem_total_n} MiB ({mem_pct:.0f}%)  \u00b7</span>'
        f'<span font_desc="Maple Mono NF CN 15" foreground="#9fb2ff">'
        f'  {escape(name)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
