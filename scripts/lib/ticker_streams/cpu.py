"""cpu — hwmon temperatures + per-core cpufreq + thread count.

Tier-colouring on temp (≥85 red, ≥75 yellow, else green) and avg-GHz
(<3.5 green, <4.5 yellow, else red).
"""
from __future__ import annotations

import glob
import os
from html import escape

import ticker_render as tr

META = {"name": "cpu", "preset": "cyberpunk", "refresh": 10}

_LABEL = "\U000f04bc CPU"
_HWMON_CACHE: dict | None = None


def _find_hwmon():
    """Locate the k10temp / coretemp / zenpower hwmon entry once."""
    global _HWMON_CACHE
    if _HWMON_CACHE is not None:
        return _HWMON_CACHE
    for h in sorted(glob.glob("/sys/class/hwmon/hwmon*")):
        try:
            name = open(os.path.join(h, "name")).read().strip()
        except OSError:
            continue
        if name in ("k10temp", "coretemp", "zenpower"):
            _HWMON_CACHE = {"path": h, "name": name}
            return _HWMON_CACHE
    _HWMON_CACHE = {}
    return _HWMON_CACHE


def build():
    parts = [tr.badge(_LABEL, "#00d4ff")]
    try:
        hwmon = _find_hwmon()
        temps = []
        if hwmon:
            for lbl_file in sorted(glob.glob(os.path.join(hwmon["path"], "temp*_label"))):
                try:
                    label = open(lbl_file).read().strip()
                    inp = lbl_file[:-len("_label")] + "_input"
                    val = int(open(inp).read().strip()) // 1000
                    temps.append((label, val))
                except (OSError, ValueError):
                    continue

        freqs = []
        for fpath in sorted(
            glob.glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_cur_freq")
        )[:64]:
            try:
                freqs.append(int(open(fpath).read().strip()))
            except (OSError, ValueError):
                continue

        ncpu = os.cpu_count() or 1

        if not temps and not freqs:
            return tr.empty(_LABEL, "#00d4ff", "no thermal/freq data")

        if temps:
            primary = next(
                (v for l, v in temps if l in ("Tctl", "Package id 0")),
                temps[0][1],
            )
            t_color = "#ff5c8a" if primary >= 85 else ("#ffe45e" if primary >= 75 else "#76ff03")
            parts.append(
                f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{t_color}">'
                f'  {primary}\u00b0C  \u00b7</span>'
            )
            extra = [f"{l} {v}" for l, v in temps
                     if l != "Tctl" and l != "Package id 0"][:3]
            if extra:
                parts.append(
                    f'<span font_desc="Maple Mono NF CN 15" foreground="#9fb2ff">'
                    f'  {escape(" ".join(extra))}  \u00b7</span>'
                )

        if freqs:
            avg_ghz = (sum(freqs) / len(freqs)) / 1_000_000
            max_ghz = max(freqs) / 1_000_000
            f_color = "#76ff03" if avg_ghz < 3.5 else ("#ffe45e" if avg_ghz < 4.5 else "#ff5c8a")
            parts.append(
                f'<span font_desc="Maple Mono NF CN 15" foreground="{f_color}">'
                f'  avg {avg_ghz:.2f} GHz  max {max_ghz:.2f} GHz  \u00b7</span>'
            )

        parts.append(
            f'<span font_desc="Maple Mono NF CN 15" foreground="#7ad0ff">'
            f'  {ncpu} threads  \u00b7</span>'
        )
    except Exception:
        return tr.empty(_LABEL, "#00d4ff", "cpu unavailable")
    return tr.dup("".join(parts)), []
