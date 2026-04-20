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
import subprocess
import sys

import gi

gi.require_version("Gtk", "4.0")
gi.require_version("Gtk4LayerShell", "1.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gtk4LayerShell, Gdk, GLib, Gio, Pango, PangoCairo  # noqa: E402

BAR_H = 28


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

        Gtk4LayerShell.init_for_window(self)
        Gtk4LayerShell.set_layer(self, Gtk4LayerShell.Layer.OVERLAY)
        for edge in (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT):
            Gtk4LayerShell.set_anchor(self, edge, True)
        Gtk4LayerShell.set_namespace(self, "hg-subtitle")
        Gtk4LayerShell.set_margin(self, Gtk4LayerShell.Edge.TOP, 100)
        display = Gdk.Display.get_default()
        if display:
            for i in range(display.get_monitors().get_n_items()):
                mon = display.get_monitors().get_item(i)
                if mon and monitor_name in (mon.get_connector() or ""):
                    Gtk4LayerShell.set_monitor(self, mon)
                    break

        self.da = Gtk.DrawingArea()
        self.da.set_content_height(BAR_H)
        self.da.set_vexpand(True)
        self.da.set_hexpand(True)
        self.da.set_draw_func(self._draw)
        self.set_child(self.da)
        self.da.connect("realize", lambda w: w.get_frame_clock().begin_updating())

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
        cr.set_source_rgba(0.02, 0.03, 0.05, 0.85)
        cr.rectangle(0, 0, w, h)
        cr.fill()
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
