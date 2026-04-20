"""network — SSID + signal, primary IPv4, tailscale IP (each a clickable segment)."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "network", "preset": None, "refresh": 30}

_LABEL = " NET"
_SKIP_IFACE = ("docker", "br-", "veth", "tailscale")


def _run(args, timeout=3):
    try:
        return subprocess.run(
            args, capture_output=True, text=True, timeout=timeout,
        )
    except Exception:
        return None


def build():
    parts = [tr.badge(_LABEL, "#38bdf8")]
    segments: list[str] = []

    ssid = signal = None
    wifi = _run(["nmcli", "-t", "-f", "IN-USE,SSID,SIGNAL",
                 "device", "wifi", "list", "--rescan", "no"])
    if wifi and wifi.returncode == 0:
        for line in wifi.stdout.strip().splitlines():
            if line.startswith(("*:", "*\\:")):
                fields = line.split(":")
                if len(fields) >= 3:
                    ssid = fields[1] or None
                    signal = fields[2] or None
                    break

    ipv4 = None
    addrs = _run(["ip", "-o", "-4", "addr", "show", "scope", "global"])
    if addrs and addrs.returncode == 0:
        for line in addrs.stdout.strip().splitlines():
            fields = line.split()
            if len(fields) >= 4:
                iface = fields[1]
                if iface.startswith(_SKIP_IFACE):
                    continue
                ipv4 = fields[3].split("/")[0]
                break

    ts_info = None
    ts = _run(["ip", "-o", "-4", "addr", "show", "tailscale0"], timeout=2)
    if ts and ts.returncode == 0 and ts.stdout.strip():
        fields = ts.stdout.strip().split()
        if len(fields) >= 4:
            ts_ip = fields[3].split("/")[0]
            ts_info = f"tailscale:{ts_ip}"

    if not (ssid or ipv4 or ts_info):
        return tr.empty(_LABEL, "#38bdf8", "offline")

    fc = len(FONTS)
    ix = 0
    if ssid:
        sig_color = "#f7fbff"
        try:
            if int(signal or 0) < 40:
                sig_color = "#ff5c8a"
        except ValueError:
            pass
        font = FONTS[ix % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="{sig_color}">'
            f'  {escape(ssid)} {escape(signal or "?")}%  \u00b7</span>'
        )
        segments.append(ssid)
        ix += 1
    if ipv4:
        font = FONTS[ix % fc]
        parts.append(
            f'<span font_desc="{font}">  {escape(ipv4)}  \u00b7</span>'
        )
        segments.append(ipv4)
        ix += 1
    if ts_info:
        font = FONTS[ix % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="#c084fc">'
            f'  {escape(ts_info)}  \u00b7</span>'
        )
        segments.append(ts_info)
        ix += 1
    return tr.dup("".join(parts)), segments
