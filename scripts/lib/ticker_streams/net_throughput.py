"""net-throughput — per-interface RX/TX rate via /proc/net/dev delta."""
from __future__ import annotations

import time
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "net-throughput", "preset": None, "refresh": 5}

_LABEL = "\U000f0317 NET TX/RX"
_SKIP_IFACE_PREFIXES = ("veth", "br-", "docker", "tailscale", "p2p-")

_last: tuple[float, dict[str, tuple[int, int]]] | None = None


def _read_proc_net_dev():
    totals = {}
    try:
        with open("/proc/net/dev") as f:
            for line in f.readlines()[2:]:
                fields = line.split()
                if not fields:
                    continue
                iface = fields[0].rstrip(":")
                totals[iface] = (int(fields[1]), int(fields[9]))
    except OSError:
        return {}
    return totals


def _format_rate(bps):
    for unit in ("B/s", "KB/s", "MB/s", "GB/s"):
        if bps < 1024:
            return f"{bps:.1f} {unit}"
        bps /= 1024
    return f"{bps:.1f} TB/s"


def build():
    global _last
    now = time.time()
    snapshot = _read_proc_net_dev()
    if not snapshot:
        return tr.empty(_LABEL, "#10b981", "/proc/net/dev unavailable")
    prev = _last
    _last = (now, snapshot)
    if prev is None:
        return tr.empty(_LABEL, "#10b981", "warming up\u2026")
    prev_now, prev_snap = prev
    dt = max(now - prev_now, 0.001)
    rows = []
    for iface, (rx, tx) in snapshot.items():
        if iface == "lo" or iface.startswith(_SKIP_IFACE_PREFIXES):
            continue
        prx, ptx = prev_snap.get(iface, (rx, tx))
        drx = max(rx - prx, 0) / dt
        dtx = max(tx - ptx, 0) / dt
        rows.append((iface, drx, dtx))
    if not rows:
        return tr.empty(_LABEL, "#10b981", "no interfaces")
    rows.sort(key=lambda r: r[1] + r[2], reverse=True)
    parts = [tr.badge(_LABEL, "#10b981")]
    fc = len(FONTS)
    for i, (iface, drx, dtx) in enumerate(rows[:4]):
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}">'
            f'  {escape(iface)} \u2193 {escape(_format_rate(drx))}'
            f' \u2191 {escape(_format_rate(dtx))}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
