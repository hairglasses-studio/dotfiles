"""system — CPU temp (sensors -j), GPU cache read, mem% (/proc/meminfo), uptime.

Reads /tmp/bar-gpu-full.txt (written by the bar-gpu.timer at 10s
cadence) instead of spawning nvidia-smi — same data, zero fork per
refresh. /proc/meminfo replaces `free -m` for the same reason. The
remaining sensors -j spawn is kept: the Tctl reading has no kernel-
exposed equivalent and sensors' own poll cost is modest.
"""
from __future__ import annotations

import json
import subprocess
from html import escape

import ticker_render as tr

META = {"name": "system", "preset": None, "refresh": 10}

_LABEL = "\U000f04bc SYSTEM"
_GPU_CACHE = "/tmp/bar-gpu-full.txt"


def _read_gpu_from_cache():
    """Return (power_w, temp_c, util_pct) or None if the cache is missing
    or malformed. The bar-gpu-cache.sh writer emits pipe-separated
    ``power|temp|util|vram_used|vram_total|name``."""
    try:
        with open(_GPU_CACHE) as f:
            fields = f.read().strip().split("|")
        if len(fields) < 3:
            return None
        return fields[0], fields[1], fields[2]
    except (FileNotFoundError, OSError):
        return None


def _read_mem_pct():
    """Return used-memory percentage derived from /proc/meminfo, or None
    if the parse fails. Uses MemAvailable (the kernel's accurate 'how
    much memory is free for a new workload') rather than MemFree, which
    excludes reclaimable cache."""
    total = avail = 0
    try:
        with open("/proc/meminfo") as f:
            for line in f:
                if line.startswith("MemTotal:"):
                    total = int(line.split()[1])
                elif line.startswith("MemAvailable:"):
                    avail = int(line.split()[1])
                if total and avail:
                    break
    except (FileNotFoundError, OSError):
        return None
    if not total:
        return None
    return (total - avail) * 100 // total


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
                                f'   {v2:.0f}°C  ·</span>'
                            )
    except Exception:
        pass
    gpu = _read_gpu_from_cache()
    if gpu is not None:
        power, temp, util = gpu
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 15">'
            f'  GPU {escape(power)}W  {escape(temp)}°C  '
            f'{escape(util)}%  ·</span>'
        )
    pct = _read_mem_pct()
    if pct is not None:
        parts.append(
            f'<span font_desc="Maple Mono NF CN 15">'
            f'   {pct}%  ·</span>'
        )
    try:
        with open("/proc/uptime") as f:
            up_s = float(f.read().split()[0])
        h, m = int(up_s // 3600), int((up_s % 3600) // 60)
        parts.append(
            f'<span font_desc="Maple Mono NF CN Light 15">'
            f'  up {h}h{m}m  ·</span>'
        )
    except Exception:
        pass
    return tr.dup("".join(parts)), []
