"""system — CPU temp (sensors -j), GPU (nvidia-smi), mem% (free), uptime."""
from __future__ import annotations

import json
import subprocess
from html import escape

import ticker_render as tr

META = {"name": "system", "preset": None, "refresh": 10}

_LABEL = " SYSTEM"


def build():
    parts = [tr.badge(_LABEL, "#ffe45e")]
    try:
        temps = subprocess.run(
            ["sensors", "-j"], capture_output=True, text=True, timeout=3,
        ).stdout
        tj = json.loads(temps)
        for chip in tj.values():
            for key, val in chip.items():
                if isinstance(val, dict) and "Tctl" in str(key):
                    for k2, v2 in val.items():
                        if "input" in k2:
                            parts.append(
                                f'<span font_desc="Maple Mono NF CN SemiBold 15">'
                                f'   {v2:.0f}\u00b0C  \u00b7</span>'
                            )
    except Exception:
        pass
    try:
        gpu = subprocess.run(
            ["nvidia-smi",
             "--query-gpu=power.draw,temperature.gpu,utilization.gpu",
             "--format=csv,noheader,nounits"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip().split(", ")
        if len(gpu) >= 3:
            parts.append(
                f'<span font_desc="Maple Mono NF CN Bold 15">'
                f'  GPU {escape(gpu[0])}W  {escape(gpu[1])}\u00b0C  '
                f'{escape(gpu[2])}%  \u00b7</span>'
            )
    except Exception:
        pass
    try:
        mem = subprocess.run(
            ["free", "-m"], capture_output=True, text=True, timeout=2,
        ).stdout
        for line in mem.splitlines():
            if line.startswith("Mem:"):
                fields = line.split()
                total = int(fields[1])
                if total > 0:
                    pct = int(fields[2]) * 100 // total
                    parts.append(
                        f'<span font_desc="Maple Mono NF CN 15">'
                        f'   {pct}%  \u00b7</span>'
                    )
    except Exception:
        pass
    try:
        with open("/proc/uptime") as f:
            up_s = float(f.read().split()[0])
        h, m = int(up_s // 3600), int((up_s % 3600) // 60)
        parts.append(
            f'<span font_desc="Maple Mono NF CN Light 15">'
            f'  up {h}h{m}m  \u00b7</span>'
        )
    except Exception:
        pass
    return tr.dup("".join(parts)), []
