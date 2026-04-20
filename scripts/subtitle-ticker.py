#!/usr/bin/env python3
"""subtitle-ticker.py — live audio captions (v2) with mute-state fallback.

v2 behaviour (gated on ``--enable-whisper``): capture the default input
via ``pw-record``, feed 1s 16 kHz mono chunks into a pywhispercpp model
on a worker thread, and render each transcribed segment in the banner
for 5 seconds before fading out.

v1 fallback (always on, or when --enable-whisper is unavailable):
display a "MUTED — audio detached" banner when the default sink is
muted. Used as a visible indicator that sound output is off.

Installing the whisper backend (NVIDIA):
    yay -S python-pywhispercpp-cuda
(or -rocm for AMD / -cpu for software-only). After install, run:
    scripts/subtitle-ticker.py --enable-whisper --model tiny.en
The tiny model is ~75 MB and is auto-downloaded into
``~/.cache/pywhispercpp/`` on first use.
"""

from __future__ import annotations

import argparse
import os
import queue
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
CAPTION_TTL_S = 5.0

try:
    from pywhispercpp.model import Model as _WhisperModel  # type: ignore
    _WHISPER_AVAILABLE = True
except Exception:
    _WhisperModel = None  # type: ignore
    _WHISPER_AVAILABLE = False


def _sink_muted() -> bool:
    try:
        out = subprocess.run(
            ["pactl", "get-sink-mute", "@DEFAULT_SINK@"],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
        return out.endswith("yes")
    except Exception:
        return False


def _whisper_install_hint() -> str:
    return (
        "pywhispercpp not found. Install one of:\n"
        "  yay -S python-pywhispercpp-cuda     # NVIDIA GPU\n"
        "  yay -S python-pywhispercpp-rocm     # AMD GPU\n"
        "  yay -S python-pywhispercpp-cpu      # software-only\n"
    )


class WhisperWorker(threading.Thread):
    """Streams audio from ``pw-record`` into pywhispercpp. Posts transcribed
    text segments onto ``out_queue`` as ``(text, epoch_s)`` tuples."""

    def __init__(self, model_name: str, out_queue: queue.Queue,
                 chunk_seconds: float = 2.0) -> None:
        super().__init__(daemon=True)
        self.model_name = model_name
        self.out_queue = out_queue
        self.chunk_seconds = chunk_seconds
        self._stop = threading.Event()

    def stop(self) -> None:
        self._stop.set()

    def run(self) -> None:
        if not _WHISPER_AVAILABLE:
            return
        try:
            model = _WhisperModel(self.model_name, n_threads=4)
        except Exception as e:
            self.out_queue.put((f"[whisper load failed: {e}]", time.time()))
            return
        sample_rate = 16000
        bytes_per_sample = 2  # s16le
        chunk_bytes = int(sample_rate * self.chunk_seconds) * bytes_per_sample
        cmd = [
            "pw-record", "--format=s16", f"--rate={sample_rate}",
            "--channels=1", "--latency-msec=200", "-",
        ]
        try:
            proc = subprocess.Popen(cmd, stdout=subprocess.PIPE,
                                    stderr=subprocess.DEVNULL)
        except FileNotFoundError:
            self.out_queue.put(("[pw-record missing]", time.time()))
            return
        import numpy as np
        try:
            while not self._stop.is_set():
                assert proc.stdout is not None
                raw = proc.stdout.read(chunk_bytes)
                if not raw or len(raw) < chunk_bytes:
                    break
                audio = np.frombuffer(raw, dtype=np.int16).astype(np.float32) / 32768.0
                try:
                    segs = model.transcribe(audio)
                except Exception:
                    continue
                for seg in segs:
                    text = (seg.text or "").strip()
                    if text:
                        self.out_queue.put((text, time.time()))
        finally:
            try:
                proc.kill()
            except Exception:
                pass


class SubtitleWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str, *, enable_whisper: bool,
                 model_name: str):
        super().__init__(application=app)
        self.enable_whisper = enable_whisper and _WHISPER_AVAILABLE
        self.model_name = model_name
        self.caption: str = ""
        self.caption_ts: float = 0.0
        self.visible_msg = False
        self._segments: queue.Queue = queue.Queue()

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

        if enable_whisper and not _WHISPER_AVAILABLE:
            # Print hint once so the user knows why captions aren't appearing.
            sys.stderr.write(_whisper_install_hint())

        if self.enable_whisper:
            self.worker: WhisperWorker | None = WhisperWorker(
                self.model_name, self._segments
            )
            self.worker.start()
        else:
            self.worker = None

        GLib.timeout_add(200, self._refresh)
        self._refresh()

    def _refresh(self) -> bool:
        # Drain any whisper segments posted by the worker.
        try:
            while True:
                text, ts = self._segments.get_nowait()
                self.caption = text
                self.caption_ts = ts
        except queue.Empty:
            pass

        now = time.time()
        has_caption = self.caption and (now - self.caption_ts) < CAPTION_TTL_S
        is_muted = _sink_muted()

        want_visible = bool(has_caption or is_muted)
        if want_visible and not self.visible_msg:
            self.visible_msg = True
            self.present()
        elif not want_visible and self.visible_msg:
            self.visible_msg = False
            self.set_visible(False)
        self.da.queue_draw()
        return True

    def _draw(self, widget, cr, w, h):
        if not self.visible_msg:
            return
        tr.fill_bg(cr, w, h, alpha=0.85)
        now = time.time()
        if self.caption and (now - self.caption_ts) < CAPTION_TTL_S:
            # Fade caption as it ages.
            fade = max(0.0, 1.0 - (now - self.caption_ts) / CAPTION_TTL_S)
            markup = (
                f'<span foreground="#29f0ff">'
                f'{GLib.markup_escape_text(self.caption)}</span>'
            )
            layout = PangoCairo.create_layout(cr)
            layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 12"))
            layout.set_markup(markup, -1)
            tw, th = layout.get_pixel_size()
            cr.set_source_rgba(1, 1, 1, fade)
            cr.move_to(max(8, (w - tw) / 2), max(0, (h - th) / 2))
            PangoCairo.show_layout(cr, layout)
        elif _sink_muted():
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
    def __init__(self, monitor: str, *, enable_whisper: bool, model_name: str):
        super().__init__(
            application_id="io.hairglasses.subtitle",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.monitor = monitor
        self.enable_whisper = enable_whisper
        self.model_name = model_name

    def do_activate(self):
        SubtitleWindow(self, self.monitor,
                        enable_whisper=self.enable_whisper,
                        model_name=self.model_name)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    ap.add_argument("--enable-whisper", action="store_true",
                     help="Capture default input and transcribe via pywhispercpp.")
    ap.add_argument("--model", default="tiny.en",
                     help="pywhispercpp model name (default: tiny.en, ~75MB).")
    args = ap.parse_args()
    app = SubtitleApp(args.monitor,
                      enable_whisper=args.enable_whisper,
                      model_name=args.model)
    return app.run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
