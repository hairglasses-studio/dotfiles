#!/usr/bin/env python3
"""window-label.py — per-window floating label overlay.

Subscribes to Hyprland's socket2 event stream and renders a 28px layer-shell
label above the active window showing ``<class> · <title>``. Anchored to the
top edge so it doesn't clash with the bottom ticker. Layer=TOP so fullscreen
clients still cover it (intended).

Fades out after 3s of no window change (or status update) and fades back in
on the next ``activewindow>>`` / ``activewindowv2>>`` event.

Usage:
  window-label.py --monitor DP-2
"""

from __future__ import annotations

import argparse
import os
import socket
import subprocess
import sys
import threading
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
FADE_IN_S = 0.25
VISIBLE_S = 3.0
FADE_OUT_S = 0.6


def _hypr_socket2_path() -> str:
    runtime = os.environ.get("XDG_RUNTIME_DIR", f"/run/user/{os.getuid()}")
    hypr_dir = os.path.join(runtime, "hypr")
    if not os.path.isdir(hypr_dir):
        raise FileNotFoundError(hypr_dir)
    entries = sorted(os.listdir(hypr_dir))
    for entry in entries:
        cand = os.path.join(hypr_dir, entry, ".socket2.sock")
        if os.path.exists(cand):
            return cand
    raise FileNotFoundError("Hyprland socket2 not found")


def _active_window() -> tuple[str, str]:
    try:
        out = subprocess.run(
            ["hyprctl", "activewindow", "-j"],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
    except Exception:
        return ("", "")
    if not out or out in ("{}", "null"):
        return ("", "")
    import json
    try:
        data = json.loads(out)
    except ValueError:
        return ("", "")
    return (data.get("class") or "", data.get("title") or "")


class WindowLabelWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str):
        super().__init__(application=app)
        self.cls = ""
        self.title = ""
        self.last_change = 0.0
        self.alpha = 0.0

        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT),
            "hg-window-label",
            monitor_name,
            layer="TOP",
            exclusive_zone=0,
            margins={Gtk4LayerShell.Edge.TOP: 4},
        )
        self.da = tr.make_drawing_area(BAR_H, self._draw)
        self.set_child(self.da)
        self.present()

        # Seed with current active window, then listen for changes.
        self._on_change(*_active_window())
        threading.Thread(target=self._socket_loop, daemon=True).start()
        self.da.add_tick_callback(self._tick)

    def _socket_loop(self) -> None:
        try:
            path = _hypr_socket2_path()
        except FileNotFoundError as e:
            print(f"window-label: {e}", file=sys.stderr)
            return
        sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        try:
            sock.connect(path)
        except OSError as e:
            print(f"window-label: connect failed: {e}", file=sys.stderr)
            return
        buf = b""
        while True:
            chunk = sock.recv(4096)
            if not chunk:
                break
            buf += chunk
            while b"\n" in buf:
                line, buf = buf.split(b"\n", 1)
                text = line.decode("utf-8", "replace")
                if not text:
                    continue
                if text.startswith("activewindow>>"):
                    rest = text.split(">>", 1)[1]
                    cls, _, title = rest.partition(",")
                    GLib.idle_add(self._on_change, cls, title)
                elif text.startswith(("workspace>>", "focusedmon>>")):
                    # Window may not have changed but the focus did; resync.
                    GLib.idle_add(self._on_change, *_active_window())

    def _on_change(self, cls: str, title: str) -> bool:
        self.cls = cls
        self.title = title
        self.last_change = time.monotonic()
        self.da.queue_draw()
        return False

    def _tick(self, widget, frame_clock) -> bool:
        now = time.monotonic()
        dt = now - self.last_change
        # Target alpha: ramp to 1.0 during VISIBLE_S, then decay over FADE_OUT_S.
        if dt < FADE_IN_S:
            target = dt / FADE_IN_S
        elif dt < VISIBLE_S:
            target = 1.0
        elif dt < VISIBLE_S + FADE_OUT_S:
            target = max(0.0, 1.0 - (dt - VISIBLE_S) / FADE_OUT_S)
        else:
            target = 0.0
        if abs(target - self.alpha) > 0.005:
            self.alpha = target
            widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _draw(self, widget, cr, w, h):
        if self.alpha <= 0.01 or not (self.cls or self.title):
            return
        tr.fill_bg(cr, w, h, alpha=0.8 * self.alpha)
        cls_txt = self.cls or "?"
        title_txt = self.title or ""
        markup = (
            f'<span foreground="#29f0ff">{GLib.markup_escape_text(cls_txt)}</span>'
            f'<span foreground="#66708f"> · </span>'
            f'<span foreground="#f7fbff">{GLib.markup_escape_text(title_txt[:120])}</span>'
        )
        layout = PangoCairo.create_layout(cr)
        layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 10"))
        layout.set_markup(markup, -1)
        tw, th = layout.get_pixel_size()
        cr.set_source_rgba(1, 1, 1, self.alpha)
        cr.move_to((w - tw) / 2, max(0, (h - th) / 2))
        PangoCairo.show_layout(cr, layout)


class WindowLabelApp(Gtk.Application):
    def __init__(self, monitor: str):
        super().__init__(
            application_id="io.hairglasses.window_label",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.monitor = monitor

    def do_activate(self):
        WindowLabelWindow(self, self.monitor)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    args = ap.parse_args()
    return WindowLabelApp(args.monitor).run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
