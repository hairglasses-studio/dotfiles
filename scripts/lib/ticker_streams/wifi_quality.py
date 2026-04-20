"""wifi-quality — iw-backed link state with signal-graded colour."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr

META = {"name": "wifi-quality", "preset": None, "refresh": 30}

_LABEL = "\U000f05a9 WIFI"


def _run(args, timeout=2):
    try:
        return subprocess.run(args, capture_output=True, text=True,
                              timeout=timeout).stdout
    except Exception:
        return None


def _dbm_color(dbm: int) -> str:
    if dbm >= -60:
        return "#34d399"
    if dbm >= -75:
        return "#fbbf24"
    return "#ef4444"


def build():
    dev_out = _run(["iw", "dev"])
    if dev_out is None:
        return tr.empty(_LABEL, "#38bdf8", "iw unavailable")
    iface = None
    for line in dev_out.splitlines():
        s = line.strip()
        if s.startswith("Interface "):
            iface = s.split(None, 1)[1]
            break
    if not iface:
        return tr.empty(_LABEL, "#38bdf8", "no wifi iface")
    link_out = _run(["iw", "dev", iface, "link"])
    if link_out is None:
        return tr.empty(_LABEL, "#38bdf8", "iw link failed")
    if "Not connected" in link_out:
        return tr.empty(_LABEL, "#66708f", f"{iface}: not connected")
    ssid = signal = bitrate = "\u2014"
    for line in link_out.splitlines():
        s = line.strip()
        if s.startswith("SSID:"):
            ssid = s.split(":", 1)[1].strip()
        elif s.startswith("signal:"):
            signal = s.split(":", 1)[1].strip()
        elif s.startswith("tx bitrate:"):
            bitrate = s.split(":", 1)[1].strip().split()[0] + " Mbps"
    try:
        dbm = int(signal.split()[0])
    except Exception:
        dbm = -100
    color = _dbm_color(dbm)
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{color}">'
        f'  {escape(iface)}  \u00b7</span>'
    )
    parts.append(
        f'<span font_desc="Maple Mono NF CN 15">'
        f'  {escape(ssid)}  \u00b7  {escape(signal)}  \u00b7  {escape(bitrate)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
