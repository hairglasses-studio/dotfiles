#!/usr/bin/env python3
"""subtitle-ticker.py — placeholder subtitle banner. v1 shows audio sink
state (mute / unplugged); v2 will display live transcription from a
whisper-live pipe when that lands.

v1 behaviour: polls `pactl get-sink-mute @DEFAULT_SINK@` every second. If
the default sink is muted, show "MUTED — audio detached" banner; otherwise
stay hidden by exiting.
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
import ticker_render as tr  # noqa: E402

import gi

gi.require_version("Gtk", "4.0")
gi.require_version("Gtk4LayerShell", "1.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gtk4LayerShell, Gdk, GLib, Gio, Pango, PangoCairo  # noqa: E402

BAR_H = tr.BAR_H


def _sink_muted() -> bool:
    try:
        out = subprocess.run(
            ["pactl", "get-sink-mute", "@DEFAULT_SINK@"],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
        return out.endswith("yes")
    except Exception:
        return False


class SubtitleWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str):
        super().__init__(application=app)
        self.visible_msg = False

        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT),
            "hg-subtitle",
            monitor_name,
            layer="OVERLAY",
            margins={Gtk4LayerShell.Edge.TOP: 100},
        )
        self.da = tr.make_drawing_area(BAR_H, self._draw)
        self.set_child(self.da)

        GLib.timeout_add(1000, self._refresh)
        self._refresh()

    def _refresh(self) -> bool:
        muted = _sink_muted()
        if muted and not self.visible_msg:
            self.visible_msg = True
            self.present()
        elif not muted and self.visible_msg:
            self.visible_msg = False
            self.set_visible(False)
        self.da.queue_draw()
        return True

    def _draw(self, widget, cr, w, h):
        if not self.visible_msg:
            return
        tr.fill_bg(cr, w, h, alpha=0.85)
        layout = PangoCairo.create_layout(cr)
        layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 12"))
        layout.set_markup(
            '<span foreground="#ff5c8a">\U000f036d MUTED — audio detached</span>',
            -1,
        )
        tw, th = layout.get_pixel_size()
        cr.move_to(max(8, (w - tw) / 2), max(0, (h - th) / 2))
        PangoCairo.show_layout(cr, layout)


class SubtitleApp(Gtk.Application):
    def __init__(self, monitor: str):
        super().__init__(
            application_id="io.hairglasses.subtitle",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.monitor = monitor

    def do_activate(self):
        SubtitleWindow(self, self.monitor)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    args = ap.parse_args()
    app = SubtitleApp(args.monitor)
    return app.run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
