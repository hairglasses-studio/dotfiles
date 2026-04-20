"""top-procs — top CPU consumers via `ps --sort=-%cpu`."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "top-procs", "preset": "cyberpunk", "refresh": 15}

_LABEL = "\U000f0233 TOP"
_SELF_NAMES = ("ps", "ps-procs")


def build():
    parts = [tr.badge(_LABEL, "#ff8855")]
    try:
        out = subprocess.run(
            ["ps", "-eo", "pid,%cpu,%mem,comm", "--sort=-%cpu", "--no-headers"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip().splitlines()
    except Exception:
        return tr.empty(_LABEL, "#ff8855", "ps unavailable")
    rows = []
    for line in out:
        fields = line.split(None, 3)
        if len(fields) < 4:
            continue
        pid, cpu, mem, comm = fields
        if comm in _SELF_NAMES:
            continue
        try:
            cpu_f = float(cpu)
            mem_f = float(mem)
        except ValueError:
            continue
        if cpu_f < 1.0:
            continue
        rows.append((pid, cpu_f, mem_f, comm[:22]))
        if len(rows) >= 6:
            break
    if not rows:
        return tr.empty(_LABEL, "#ff8855", "system idle")
    fc = len(FONTS)
    for i, (_pid, cpu_f, mem_f, comm) in enumerate(rows):
        color = "#ff5c8a" if cpu_f > 75 else ("#ffe45e" if cpu_f > 25 else "#f7fbff")
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="{color}">'
            f'  {escape(comm)} '
            f'<span foreground="#9fb2ff">{cpu_f:.0f}% cpu</span> '
            f'<span foreground="#7ad0ff">{mem_f:.1f}% mem</span>'
            f'  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
