#!/usr/bin/env python3
"""hypr-monitor-watch.sh — re-apply DP-3 transform across Hyprland reloads.

Background: `monitor = DP-3, ..., transform, 3` in a sourced config file
is dropped on `hyprctl reload` because Hyprland's reload path triggers a
DRM disconnect on DP-3 during modeset (see aquamarine debug log
"Connector DP-3 disconnected"), and the transform from the parsed rule
never reaches the reconnected monitor. The runtime `hyprctl keyword`
code path doesn't hit that disconnect — applying the transform from a
daemon that watches for `configreloaded` events keeps DP-3 rotated.

Also handles `monitoradded` / `monitoraddedv2` events so the transform
comes back after cable reseat / hotplug.

Named .sh for systemd unit compatibility even though the body is Python.
"""
from __future__ import annotations

import json
import os
import socket
import subprocess
import sys
import time

DP3_DIRECTIVE = "DP-3, preferred, auto, 1, transform, 3"
SETTLE_S = 0.3


def apply_transform() -> None:
    try:
        raw = subprocess.run(
            ["hyprctl", "monitors", "-j"],
            capture_output=True, text=True, timeout=2,
        ).stdout
        mons = json.loads(raw)
    except Exception:
        return
    if not any(m.get("name") == "DP-3" for m in mons):
        return
    subprocess.run(
        ["hyprctl", "keyword", "monitor", DP3_DIRECTIVE],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, timeout=3,
    )


def socket_path() -> str:
    sig = os.environ.get("HYPRLAND_INSTANCE_SIGNATURE")
    rt = os.environ.get("XDG_RUNTIME_DIR") or f"/run/user/{os.getuid()}"
    if not sig:
        # Last-resort: pick the most recent instance dir.
        hypr_dir = os.path.join(rt, "hypr")
        try:
            entries = sorted(os.listdir(hypr_dir),
                             key=lambda e: os.path.getmtime(os.path.join(hypr_dir, e)),
                             reverse=True)
        except OSError:
            entries = []
        if entries:
            sig = entries[0]
    if not sig:
        raise RuntimeError("no Hyprland instance signature")
    return os.path.join(rt, "hypr", sig, ".socket2.sock")


def main() -> int:
    apply_transform()  # seed once on startup
    path = socket_path()
    s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    s.connect(path)
    buf = b""
    while True:
        chunk = s.recv(4096)
        if not chunk:
            break
        buf += chunk
        while b"\n" in buf:
            line, buf = buf.split(b"\n", 1)
            event = line.decode(errors="replace").strip()
            if not event:
                continue
            if (event.startswith("configreloaded>>")
                    or event.startswith("monitoradded>>DP-3")
                    or "DP-3" in event and event.startswith("monitoraddedv2>>")):
                time.sleep(SETTLE_S)
                apply_transform()
    return 0


if __name__ == "__main__":
    sys.exit(main())
