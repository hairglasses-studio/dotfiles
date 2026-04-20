#!/usr/bin/env python3
"""toast-ticker.py — slide-in, fade-out toast notifications via DBus.

Listens on the session bus at io.hairglasses.toast and renders each
ShowToast(message, color) call as a 28px layer-shell strip that slides in
from the right, holds for 3s, then fades out. Same glyph + gradient stack
as keybind-ticker, but no scrolling and no stream cycling.

Usage (client side):
  gdbus call --session \\
    --dest io.hairglasses.toast \\
    --object-path /io/hairglasses/toast \\
    --method io.hairglasses.Toast.ShowToast \\
    "build complete" "#3dffb5"

CLI flags:
  --monitor DP-2           # target output (default: DP-2)
  --duration 3.0           # hold time in seconds (default: 3.0)
"""

from __future__ import annotations

import argparse
import os
import sys
import time

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
SLIDE_DURATION_S = 0.35
FADE_DURATION_S = 0.4
DBUS_NAME = "io.hairglasses.toast"
DBUS_PATH = "/io/hairglasses/toast"
DBUS_IFACE = "io.hairglasses.Toast"

INTROSPECTION = f"""
<node>
  <interface name="{DBUS_IFACE}">
    <method name="ShowToast">
      <arg type="s" name="message" direction="in"/>
      <arg type="s" name="color"   direction="in"/>
    </method>
  </interface>
</node>
"""


class ToastWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str, duration: float):
        super().__init__(application=app)
        self.monitor_name = monitor_name
        self.duration = duration
        self.queue: list[tuple[str, str]] = []
        self.current: tuple[str, str] | None = None
        self.slide_progress = 0.0  # 0 off-screen, 1 fully shown
        self.state = "hidden"  # hidden | slide_in | hold | fade
        self._anim_start = 0.0

        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.BOTTOM, Gtk4LayerShell.Edge.RIGHT),
            "hg-toast",
            monitor_name,
            layer="OVERLAY",
            margins={
                Gtk4LayerShell.Edge.BOTTOM: BAR_H + 8,
                Gtk4LayerShell.Edge.RIGHT: 16,
            },
        )
        self.set_default_size(360, BAR_H)

        self.da = tr.make_drawing_area(BAR_H, self._draw)
        self.da.set_content_width(360)
        self.set_child(self.da)
        self.da.add_tick_callback(self._tick)
        self.present()

    def push(self, message: str, color: str) -> None:
        self.queue.append((message, color))
        if self.state == "hidden":
            self._start_next()

    def _start_next(self) -> None:
        if not self.queue:
            return
        self.current = self.queue.pop(0)
        self.state = "slide_in"
        self._anim_start = time.monotonic()

    def _tick(self, widget, frame_clock) -> bool:
        if self.state == "hidden":
            return GLib.SOURCE_CONTINUE
        now = time.monotonic()
        elapsed = now - self._anim_start
        if self.state == "slide_in":
            self.slide_progress = min(1.0, elapsed / SLIDE_DURATION_S)
            if self.slide_progress >= 1.0:
                self.state = "hold"
                self._anim_start = now
        elif self.state == "hold":
            if elapsed >= self.duration:
                self.state = "fade"
                self._anim_start = now
        elif self.state == "fade":
            self.slide_progress = max(0.0, 1.0 - (elapsed / FADE_DURATION_S))
            if self.slide_progress <= 0.0:
                self.state = "hidden"
                self.current = None
                self._start_next()
        widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _draw(self, widget, cr, w, h):
        if not self.current:
            return
        msg, color = self.current
        # Ease-out for slide-in, linear for fade
        eased = 1 - (1 - self.slide_progress) ** 3 if self.state != "fade" else self.slide_progress
        alpha = eased

        # Background
        r, g, b = tr.hex_to_rgb(color)
        tr.fill_bg(cr, w, h, alpha=0.88 * alpha)
        # Left accent bar in the provided color
        cr.set_source_rgba(r, g, b, alpha)
        cr.rectangle(0, 0, 4, h)
        cr.fill()

        # Text
        layout = PangoCairo.create_layout(cr)
        font = Pango.FontDescription("Maple Mono NF CN Bold 11")
        layout.set_font_description(font)
        layout.set_markup(
            f'<span foreground="#f7fbff">{GLib.markup_escape_text(msg)}</span>',
            -1,
        )
        tw, th = layout.get_pixel_size()
        cr.move_to(14, max(0, (h - th) / 2))
        cr.set_source_rgba(1, 1, 1, alpha)
        PangoCairo.show_layout(cr, layout)

        # Resize window to accommodate text
        target_w = min(max(tw + 40, 220), 720)
        if self.get_default_size()[0] != target_w:
            self.set_default_size(target_w, BAR_H)
            self.da.set_content_width(target_w)


class ToastApp(Gtk.Application):
    def __init__(self, monitor: str, duration: float):
        super().__init__(
            application_id=DBUS_NAME,
            flags=Gio.ApplicationFlags.IS_SERVICE,
        )
        self.set_inactivity_timeout(0)
        self.monitor_name = monitor
        self.duration = duration
        self.window: ToastWindow | None = None

    def do_activate(self):
        if self.window is None:
            self.window = ToastWindow(self, self.monitor_name, self.duration)
        self.hold()

    def do_dbus_register(self, connection, path):  # type: ignore[override]
        node = Gio.DBusNodeInfo.new_for_xml(INTROSPECTION)
        iface = node.lookup_interface(DBUS_IFACE)
        connection.register_object(
            DBUS_PATH,
            iface,
            self._on_method_call,
            None,
            None,
        )
        return True

    def _on_method_call(self, conn, sender, path, iface, method, params, invocation):
        if method == "ShowToast":
            msg, color = params.unpack()
            GLib.idle_add(self._push, msg, color)
            invocation.return_value(None)
        else:
            invocation.return_error_literal(
                Gio.dbus_error_quark(),
                Gio.DBusError.UNKNOWN_METHOD,
                f"Unknown method: {method}",
            )

    def _push(self, msg: str, color: str):
        if self.window is None:
            self.window = ToastWindow(self, self.monitor_name, self.duration)
        self.window.push(msg, color)
        return False


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    ap.add_argument("--duration", type=float, default=3.0)
    args, rest = ap.parse_known_args()
    app = ToastApp(args.monitor, args.duration)
    return app.run([sys.argv[0]] + rest)


if __name__ == "__main__":
    sys.exit(main())
