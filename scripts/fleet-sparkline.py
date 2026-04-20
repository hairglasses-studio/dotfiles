#!/usr/bin/env python3
"""fleet-sparkline.py — compact top-right system sparklines for the fleet.

Anchors a 200x80 layer-shell widget to the top-right of DP-2 (configurable
via --monitor) and renders four mini-charts in a 2x2 grid:

    ┌─────────┬─────────┐
    │ CPU %   │ MEM %   │
    ├─────────┼─────────┤
    │ NET KB/s│ GPU °C  │
    └─────────┴─────────┘

- CPU usage from /proc/stat delta
- Memory from /proc/meminfo MemAvailable
- Network RX+TX from /proc/net/dev (summed across non-lo interfaces)
- GPU temp from /tmp/bar-gpu.txt (parsed heuristically)

Samples every 1s, keeps 60 samples (~1 minute of history). No scroll, no
interactivity — a passive telemetry thumbnail for the fleet.
"""

from __future__ import annotations

import argparse
import os
import re
import sys
import time
from collections import deque

sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
import ticker_render as tr  # noqa: E402

import gi

gi.require_version("Gtk", "4.0")
gi.require_version("Gtk4LayerShell", "1.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gtk4LayerShell, Gdk, GLib, Gio, Pango, PangoCairo  # noqa: E402

WIDTH = 200
HEIGHT = 80
HISTORY = 60
GPU_TEMP_RE = re.compile(r"(\d+)\s*[°C]")


def _read_cpu_idle() -> tuple[int, int]:
    try:
        with open("/proc/stat") as f:
            for line in f:
                if line.startswith("cpu "):
                    parts = line.split()
                    nums = [int(p) for p in parts[1:8]]
                    idle = nums[3] + nums[4]
                    total = sum(nums)
                    return idle, total
    except OSError:
        pass
    return 0, 0


def _read_mem_pct() -> float:
    try:
        total = avail = 0
        with open("/proc/meminfo") as f:
            for line in f:
                if line.startswith("MemTotal:"):
                    total = int(line.split()[1])
                elif line.startswith("MemAvailable:"):
                    avail = int(line.split()[1])
                if total and avail:
                    break
        if total:
            return 100.0 * (total - avail) / total
    except OSError:
        pass
    return 0.0


def _read_net_bytes() -> int:
    total = 0
    try:
        with open("/proc/net/dev") as f:
            next(f); next(f)  # header
            for line in f:
                iface, rest = line.split(":", 1)
                iface = iface.strip()
                if iface in ("lo",):
                    continue
                cols = rest.split()
                rx = int(cols[0])
                tx = int(cols[8])
                total += rx + tx
    except (OSError, ValueError, IndexError):
        pass
    return total


def _read_gpu_temp() -> float:
    try:
        with open("/tmp/bar-gpu.txt") as f:
            txt = f.read()
    except OSError:
        return 0.0
    m = GPU_TEMP_RE.search(txt)
    if m:
        try:
            return float(m.group(1))
        except ValueError:
            pass
    return 0.0


class FleetSparklineWindow(Gtk.Window):
    def __init__(self, app, monitor_name: str):
        # Gtk.Window instead of Gtk.ApplicationWindow — the latter reserves a
        # header-bar slot that GTK accounts for in its measure cycle, which
        # desyncs drawing-area coordinates against the layer-shell surface
        # when the window isn't anchored on all four edges. The top half of
        # the widget ended up unpainted because GTK was laying out the DA
        # at y=+40 inside a 200x80 surface.
        super().__init__()
        app.add_window(self)
        self.cpu_buf: deque[float] = deque(maxlen=HISTORY)
        self.mem_buf: deque[float] = deque(maxlen=HISTORY)
        self.net_buf: deque[float] = deque(maxlen=HISTORY)
        self.gpu_buf: deque[float] = deque(maxlen=HISTORY)
        self._prev_idle, self._prev_total = _read_cpu_idle()
        self._prev_net = _read_net_bytes()
        self._last_sample = time.monotonic()

        # Layer-shell anchors. TOP-only with set_default_size makes the
        # surface behave like a floating overlay. The margin_top MUST leave
        # a vertical gap between this surface and any other layer-shell
        # surface above it (like hg-window-label at y=4..32) — edge-adjacent
        # layer surfaces trigger a GTK clip that erases the top half of the
        # lower surface. 48px gives ~16px clearance below the window-label.
        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.TOP,),
            "hg-fleet-sparkline",
            monitor_name,
            layer="TOP",
            margins={Gtk4LayerShell.Edge.TOP: 60},
        )
        self.set_decorated(False)
        self.set_default_size(WIDTH, HEIGHT)
        self.set_size_request(WIDTH, HEIGHT)
        self.da = Gtk.DrawingArea()
        self.da.set_content_width(WIDTH)
        self.da.set_content_height(HEIGHT)
        self.da.set_size_request(WIDTH, HEIGHT)
        self.da.set_draw_func(self._draw)
        self.da.connect(
            "realize",
            lambda w: w.get_frame_clock().begin_updating(),
        )
        self.set_child(self.da)
        self.present()

        GLib.timeout_add_seconds(1, self._sample)

    def _sample(self) -> bool:
        # CPU
        idle, total = _read_cpu_idle()
        d_idle = idle - self._prev_idle
        d_total = total - self._prev_total
        self._prev_idle, self._prev_total = idle, total
        cpu_pct = 100.0 * (1.0 - d_idle / d_total) if d_total > 0 else 0.0
        self.cpu_buf.append(max(0.0, min(100.0, cpu_pct)))

        self.mem_buf.append(_read_mem_pct())

        # Network KB/s (delta over the tick interval)
        now = time.monotonic()
        dt = max(1e-3, now - self._last_sample)
        net = _read_net_bytes()
        kbps = max(0.0, (net - self._prev_net) / 1024.0 / dt)
        self._prev_net = net
        self._last_sample = now
        self.net_buf.append(kbps)

        self.gpu_buf.append(_read_gpu_temp())

        self.da.queue_draw()
        return True

    def _draw_panel(self, cr, x, y, w, h, title, value, buf, *, color,
                    max_hint: float | None = None):
        # Save + restore the cairo state around each panel so leftover path
        # state from earlier strokes / clips in the outer draw doesn't bleed
        # into subsequent panels. This is what was clipping the top-row
        # panels: the stroke() on the preceding panel left a path that the
        # top-row fills rendered inside.
        cr.save()
        try:
            cr.new_path()
            # Panel tile — distinct from outer bg so the 2×2 grid reads clearly.
            cr.set_source_rgba(0.10, 0.12, 0.18, 0.95)
            cr.rectangle(x + 1, y + 1, w - 2, h - 2)
            cr.fill()
            # Subtle border in the panel's accent colour.
            cr.set_source_rgba(*tr.hex_to_rgb(color), 0.35)
            cr.set_line_width(1.0)
            cr.rectangle(x + 1, y + 1, w - 2, h - 2)
            cr.stroke()
            # Title + value
            layout = PangoCairo.create_layout(cr)
            layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 9"))
            layout.set_markup(
                f'<span foreground="#66708f">{title}</span> '
                f'<span foreground="{color}">{value}</span>',
                -1,
            )
            cr.move_to(x + 4, y + 2)
            PangoCairo.show_layout(cr, layout)
            # Sparkline
            if len(buf) < 2:
                return
            sx = x + 4
            sy = y + h - 4
            sw = w - 8
            sh = h - 16  # leave room for title line
            buf_max = max(buf) or 1.0
            m = max(max_hint if max_hint is not None else buf_max, buf_max, 1e-6)
            step = sw / (len(buf) - 1)
            cr.set_source_rgba(*tr.hex_to_rgb(color), 0.95)
            cr.set_line_width(1.2)
            cr.new_path()
            for i, v in enumerate(buf):
                px = sx + i * step
                py = sy - (min(v, m) / m) * sh
                if i == 0:
                    cr.move_to(px, py)
                else:
                    cr.line_to(px, py)
            cr.stroke()
        finally:
            cr.restore()

    def _draw(self, widget, cr, w, h):
        # NOTE: On DP-2 while the Samsung is in DSC-fallback mode (scale=1
        # 3840x1080 instead of native scale=2 5120x1440) the Wayland surface
        # commit comes back ~50px tall instead of the requested 80px and the
        # top rows render as empty area. Power-cycling the monitor restores
        # DSC and the surface renders at full 200x80. No code-side fix is
        # possible until the compositor / GTK stack reconciles after the
        # EDID regains its full mode list.
        tr.fill_bg(cr, w, h, alpha=0.85)
        hw = w / 2
        hh = h / 2
        cpu = self.cpu_buf[-1] if self.cpu_buf else 0.0
        mem = self.mem_buf[-1] if self.mem_buf else 0.0
        net = self.net_buf[-1] if self.net_buf else 0.0
        gpu = self.gpu_buf[-1] if self.gpu_buf else 0.0
        # Normalize net to an adaptive max so low traffic still shows movement
        net_max = max(max(self.net_buf) if self.net_buf else 0.0, 64.0)
        cpu = self.cpu_buf[-1] if self.cpu_buf else 0.0
        mem = self.mem_buf[-1] if self.mem_buf else 0.0
        net = self.net_buf[-1] if self.net_buf else 0.0
        gpu = self.gpu_buf[-1] if self.gpu_buf else 0.0
        net_max = max(max(self.net_buf) if self.net_buf else 0.0, 64.0)
        self._draw_panel(cr, 0, 0, hw, hh,
                          "CPU", f"{cpu:4.0f}%",
                          self.cpu_buf, color="#29f0ff", max_hint=100.0)
        self._draw_panel(cr, hw, 0, hw, hh,
                          "MEM", f"{mem:4.0f}%",
                          self.mem_buf, color="#ff47d1", max_hint=100.0)
        self._draw_panel(cr, 0, hh, hw, hh,
                          "NET", self._fmt_rate(net),
                          self.net_buf, color="#3dffb5", max_hint=net_max)
        self._draw_panel(cr, hw, hh, hw, hh,
                          "GPU", f"{gpu:4.0f}°",
                          self.gpu_buf, color="#ffe45e", max_hint=100.0)

    @staticmethod
    def _fmt_rate(kbps: float) -> str:
        if kbps >= 1024:
            return f"{kbps / 1024:4.1f}M"
        return f"{kbps:4.0f}K"


class FleetSparklineApp(Gtk.Application):
    def __init__(self, monitor: str):
        super().__init__(
            application_id="io.hairglasses.fleet_sparkline",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.monitor = monitor

    def do_activate(self):
        FleetSparklineWindow(self, self.monitor)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    args = ap.parse_args()
    return FleetSparklineApp(args.monitor).run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
