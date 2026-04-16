#!/usr/bin/env python3
"""keybind-ticker.py — Cyberpunk multi-stream ticker for Hyprland.

GTK4 DrawingArea with PangoCairo at 240Hz frame-clock sync.
Features:
  - Multi-stream content: keybinds, system stats, fleet status, weather
  - Visual effects: water caustic, neon glow, gradient, scanlines, glitch
  - Interactive: scroll wheel speed, click-to-copy, hover tooltip
  - Effect presets: ambient, cyberpunk, minimal, clean

Usage:
  keybind-ticker.py                    # regular window
  keybind-ticker.py --layer            # layer-shell bar
  keybind-ticker.py --preset minimal   # start with a preset
"""

import gi
import subprocess
import json
import sys
import os
import math
import random
import cairo as _cairo
import numpy as np
from scipy.ndimage import uniform_filter1d
from html import escape

gi.require_version("Gtk", "4.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gdk, Pango, PangoCairo, GLib

LAYER_MODE = "--layer" in sys.argv

if LAYER_MODE:
    gi.require_version("Gtk4LayerShell", "1.0")
    from gi.repository import Gtk4LayerShell


# ══════════════════════════════════════════════════
# Configuration
# ══════════════════════════════════════════════════

BAR_H = 28
REFRESH_S = 300         # stream rotation interval (seconds)
STATE_DIR = os.path.expanduser("~/.local/state/keybind-ticker")

# Maple font weight cycle
FONTS = [
    "Maple Mono NF CN 11",
    "Maple Mono NF CN Bold 11",
    "Maple Mono NF CN Italic 11",
    "Maple Mono NF CN SemiBold 11",
    "Maple Mono NF CN Light 11",
    "Maple Mono NF CN Medium 11",
    "Maple Mono NF CN ExtraBold 11",
    "Maple Mono NF CN Thin 11",
    "Maple Mono NF CN ExtraLight 11",
    "Maple Mono NF CN Bold Italic 11",
]

# Hairglasses Neon gradient stops
GRADIENT_COLORS = [
    (0.161, 0.941, 1.000),  # #29f0ff cyan
    (1.000, 0.278, 0.820),  # #ff47d1 magenta
    (0.239, 1.000, 0.710),  # #3dffb5 green
    (1.000, 0.894, 0.369),  # #ffe45e yellow
    (1.000, 0.361, 0.541),  # #ff5c8a pink
    (0.290, 0.659, 1.000),  # #4aa8ff blue
    (0.161, 0.941, 1.000),  # #29f0ff cyan (wrap)
]

# ── Effect presets ────────────────────────────────
PRESETS = {
    "ambient": {
        "speed": 55.0, "gradient_speed": 40.0, "gradient_span": 800.0,
        "glow_kernel": 17, "glow_base_alpha": 0.35, "glow_pulse_amp": 0.15,
        "glow_pulse_period": 3.5, "water_skip": 120, "scanline_opacity": 0.08,
        "shadow_offset": 2, "shadow_alpha": 0.30,
        "glitch_prob": 0.004, "glitch_frames": 4, "ca_offset": 3,
        "bg_alpha": 0.82, "outline_width": 0.8,
        "wave_amp": 1.5, "wave_freq": 0.015, "wave_speed": 2.0,
    },
    "cyberpunk": {
        "speed": 65.0, "gradient_speed": 55.0, "gradient_span": 600.0,
        "glow_kernel": 21, "glow_base_alpha": 0.50, "glow_pulse_amp": 0.25,
        "glow_pulse_period": 2.5, "water_skip": 80, "scanline_opacity": 0.14,
        "shadow_offset": 3, "shadow_alpha": 0.35,
        "glitch_prob": 0.008, "glitch_frames": 5, "ca_offset": 4,
        "bg_alpha": 0.78, "outline_width": 1.0,
        "wave_amp": 2.0, "wave_freq": 0.02, "wave_speed": 3.0,
    },
    "minimal": {
        "speed": 45.0, "gradient_speed": 30.0, "gradient_span": 1000.0,
        "glow_kernel": 13, "glow_base_alpha": 0.20, "glow_pulse_amp": 0.05,
        "glow_pulse_period": 5.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 1, "shadow_alpha": 0.20,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.90, "outline_width": 0.5,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
    },
    "clean": {
        "speed": 50.0, "gradient_speed": 35.0, "gradient_span": 900.0,
        "glow_kernel": 0, "glow_base_alpha": 0.0, "glow_pulse_amp": 0.0,
        "glow_pulse_period": 1.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 0, "shadow_alpha": 0.0,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.92, "outline_width": 0.0,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
    },
}


def load_preset(name):
    p = PRESETS.get(name, PRESETS["ambient"])
    return type("P", (), p)()


# ══════════════════════════════════════════════════
# Content Streams
# ══════════════════════════════════════════════════

def fmt_mods(mask):
    out = ""
    if mask & 64: out += "Super+"
    if mask & 1:  out += "Shift+"
    if mask & 4:  out += "Ctrl+"
    if mask & 8:  out += "Alt+"
    return out


def _badge(text, bg_hex, fg_hex="#05070d"):
    return f'<span background="{bg_hex}" foreground="{fg_hex}" font_desc="Maple Mono NF CN Bold 10"> {escape(text)} </span>  '


def build_keybinds_markup():
    try:
        raw = subprocess.run(
            ["hyprctl", "binds", "-j"],
            capture_output=True, text=True, timeout=5,
        ).stdout
        binds = json.loads(raw)
    except Exception:
        return _badge("KEYBINDS", "#29f0ff") + '<span font_desc="Maple Mono NF CN 11">  No keybinds loaded  \u00b7</span>', []

    parts = [_badge(" KEYBINDS", "#29f0ff")]
    segments = []
    fc = len(FONTS)
    i = 0
    for b in binds:
        if b.get("has_description") and not b.get("submap") and not b.get("mouse"):
            mods = fmt_mods(b["modmask"])
            desc = escape(b["description"])
            key_text = f"{mods}{b['key']}"
            key = escape(key_text)
            font = FONTS[i % fc]
            parts.append(f'<span font_desc="{font}">  {desc}  {key}  \u00b7</span>')
            segments.append(key_text)
            i += 1
    single = "".join(parts)
    return single + single, segments


def build_system_markup():
    parts = [_badge(" SYSTEM", "#ffe45e")]
    try:
        temps = subprocess.run(["sensors", "-j"], capture_output=True, text=True, timeout=3).stdout
        tj = json.loads(temps)
        for chip in tj.values():
            for key, val in chip.items():
                if isinstance(val, dict) and "Tctl" in str(key):
                    for k2, v2 in val.items():
                        if "input" in k2:
                            parts.append(f'<span font_desc="Maple Mono NF CN SemiBold 11">   {v2:.0f}\u00b0C  \u00b7</span>')
    except Exception:
        pass
    try:
        gpu = subprocess.run(
            ["nvidia-smi", "--query-gpu=power.draw,temperature.gpu,utilization.gpu",
             "--format=csv,noheader,nounits"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip().split(", ")
        if len(gpu) >= 3:
            parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">  GPU {gpu[0]}W  {gpu[1]}\u00b0C  {gpu[2]}%  \u00b7</span>')
    except Exception:
        pass
    try:
        mem = subprocess.run(["free", "-m"], capture_output=True, text=True, timeout=2).stdout
        for line in mem.splitlines():
            if line.startswith("Mem:"):
                fields = line.split()
                pct = int(fields[2]) * 100 // int(fields[1])
                parts.append(f'<span font_desc="Maple Mono NF CN 11">   {pct}%  \u00b7</span>')
    except Exception:
        pass
    try:
        with open("/proc/uptime") as f:
            up_s = float(f.read().split()[0])
            h, m = int(up_s // 3600), int((up_s % 3600) // 60)
            parts.append(f'<span font_desc="Maple Mono NF CN Light 11">  up {h}h{m}m  \u00b7</span>')
    except Exception:
        pass
    single = "".join(parts)
    return single + single, []


def build_fleet_markup():
    parts = [_badge("󰑮 FLEET", "#ff47d1")]
    try:
        with open("/tmp/rg-status.json") as f:
            data = json.load(f)
        fl = data.get("fleet", {})
        cost = data.get("cost", {})
        loops = data.get("loops", {})
        parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">'
                     f'  {fl.get("running",0)} running  {fl.get("completed",0)} done  '
                     f'{fl.get("failed",0)} failed  {fl.get("pending",0)} pending  \u00b7</span>')
        parts.append(f'<span font_desc="Maple Mono NF CN SemiBold 11">'
                     f'  {loops.get("total_runs",0)} loops  ${cost.get("total_spend_usd",0)}  \u00b7</span>')
        models = data.get("models", [])
        for m in models[:3]:
            parts.append(f'<span font_desc="Maple Mono NF CN Italic 11">'
                         f'  {m["model"]} \u00d7{m["count"]}  \u00b7</span>')
    except Exception:
        parts.append('<span font_desc="Maple Mono NF CN 11">  no fleet data  \u00b7</span>')
    single = "".join(parts)
    return single + single, []


def build_weather_markup():
    parts = [_badge(" WEATHER", "#4aa8ff")]
    try:
        raw = subprocess.run(
            [os.path.expanduser("~/hairglasses-studio/dotfiles/scripts/bar-weather.sh")],
            capture_output=True, text=True, timeout=10,
        ).stdout.strip()
        if raw:
            parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">  {escape(raw)}  \u00b7</span>')
        else:
            parts.append('<span font_desc="Maple Mono NF CN 11">  weather unavailable  \u00b7</span>')
    except Exception:
        parts.append('<span font_desc="Maple Mono NF CN 11">  weather unavailable  \u00b7</span>')
    single = "".join(parts)
    return single + single, []


def build_github_markup():
    TYPE_ICONS = {"PullRequest": "", "Issue": "", "Release": "",
                  "Discussion": "󰍡", "CheckSuite": ""}
    parts = [_badge(" GITHUB", "#3dffb5")]
    try:
        raw = subprocess.run(
            ["gh", "api", "/notifications", "--paginate", "--jq",
             '.[] | {type: .subject.type, title: .subject.title, repo: .repository.name, reason: .reason}'],
            capture_output=True, text=True, timeout=15,
        ).stdout.strip()
        if raw:
            seen = 0
            fc = len(FONTS)
            for line in raw.splitlines():
                if seen >= 20:
                    break
                try:
                    n = json.loads(line)
                except Exception:
                    continue
                icon = TYPE_ICONS.get(n.get("type", ""), "")
                title = escape(n.get("title", "")[:60])
                repo = escape(n.get("repo", ""))
                font = FONTS[seen % fc]
                parts.append(f'<span font_desc="{font}">  {icon} {repo}: {title}  \u00b7</span>')
                seen += 1
            if seen == 0:
                parts.append('<span font_desc="Maple Mono NF CN 11">  no notifications  \u00b7</span>')
        else:
            parts.append('<span font_desc="Maple Mono NF CN 11">  no notifications  \u00b7</span>')
    except Exception:
        parts.append('<span font_desc="Maple Mono NF CN 11">  github unavailable  \u00b7</span>')
    single = "".join(parts)
    return single + single, []


def build_notifications_markup():
    URGENCY_ICONS = {"critical": "󰀦", "normal": "󰂚", "low": "󰂞"}
    history = os.path.expanduser(
        "~/.local/state/dotfiles/desktop-control/notifications/history.jsonl")
    parts = [_badge("󰂚 NOTIFICATIONS", "#ff5c8a")]
    try:
        with open(history) as f:
            lines = f.readlines()
        recent = lines[-30:] if len(lines) > 30 else lines
        recent.reverse()
        fc = len(FONTS)
        for i, line in enumerate(recent):
            try:
                n = json.loads(line)
            except Exception:
                continue
            icon = URGENCY_ICONS.get(n.get("urgency", ""), "󰂚")
            app = escape(n.get("app", "")[:20])
            summary = escape(n.get("summary", "")[:40])
            body = escape(n.get("body", "")[:40])
            font = FONTS[i % fc]
            text = f"{summary}: {body}" if body and body != summary else summary
            parts.append(f'<span font_desc="{font}">  {icon} {app} {text}  \u00b7</span>')
    except FileNotFoundError:
        parts.append('<span font_desc="Maple Mono NF CN 11">  no notification history  \u00b7</span>')
    except Exception:
        parts.append('<span font_desc="Maple Mono NF CN 11">  notifications unavailable  \u00b7</span>')
    single = "".join(parts)
    return single + single, []


def build_music_markup():
    parts = [_badge(" MUSIC", "#ff47d1")]
    try:
        status = subprocess.run(
            ["playerctl", "status"], capture_output=True, text=True, timeout=3
        ).stdout.strip()
        if status in ("Playing", "Paused"):
            icon = "" if status == "Playing" else ""
            meta = subprocess.run(
                ["playerctl", "metadata", "--format",
                 "{{artist}} — {{title}} [{{album}}]"],
                capture_output=True, text=True, timeout=3,
            ).stdout.strip()
            pos = subprocess.run(
                ["playerctl", "position", "--format", "{{duration(position)}}"],
                capture_output=True, text=True, timeout=3,
            ).stdout.strip()
            dur = subprocess.run(
                ["playerctl", "metadata", "--format", "{{duration(mpris:length)}}"],
                capture_output=True, text=True, timeout=3,
            ).stdout.strip()
            parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">'
                         f'  {icon} {escape(meta)}  {pos}/{dur}  \u00b7</span>')
        else:
            parts.append('<span font_desc="Maple Mono NF CN 11">  no media playing  \u00b7</span>')
    except Exception:
        parts.append('<span font_desc="Maple Mono NF CN 11">  no media playing  \u00b7</span>')
    single = "".join(parts)
    return single + single, []


# Stream registry
STREAMS = {
    "keybinds": build_keybinds_markup,
    "system": build_system_markup,
    "fleet": build_fleet_markup,
    "weather": build_weather_markup,
    "github": build_github_markup,
    "notifications": build_notifications_markup,
    "music": build_music_markup,
}

STREAM_ORDER = ["keybinds", "system", "fleet", "weather", "github", "notifications", "music"]


# ══════════════════════════════════════════════════
# Visual Helpers
# ══════════════════════════════════════════════════

def make_gradient(x_start, total_width, phase):
    x0 = x_start - phase
    grad = _cairo.LinearGradient(x0, 0, x0 + total_width, 0)
    grad.set_extend(_cairo.Extend.REPEAT)
    n = len(GRADIENT_COLORS)
    for i, (r, g, b) in enumerate(GRADIENT_COLORS):
        grad.add_color_stop_rgb(i / (n - 1), r, g, b)
    return grad


# ── Water caustic ─────────────────────────────────
_WATER_SCALE = 4
_WATER_MAX_ITER = 6
_WATER_TAU = 6.28318530718


def compute_water_caustic(width, height, time_s):
    sw = width // _WATER_SCALE
    sh = max(height // _WATER_SCALE, 2)
    t = time_s * 0.5 + 23.0
    ux = np.linspace(0, 1, sw, dtype=np.float32)
    uy = np.linspace(0, 1, sh, dtype=np.float32)
    u, v = np.meshgrid(ux, uy)
    px = np.mod(u * _WATER_TAU, _WATER_TAU) - 250.0
    py = np.mod(v * _WATER_TAU, _WATER_TAU) - 250.0
    ix, iy = px.copy(), py.copy()
    c = np.ones_like(px)
    inten = 0.005
    for n in range(_WATER_MAX_ITER):
        tn = t * (1.0 - 3.5 / (n + 1))
        ix, iy = (px + np.cos(tn - ix) + np.sin(tn + iy),
                  py + np.sin(tn - iy) + np.cos(tn + ix))
        denom = np.sqrt((px / (np.sin(ix + tn) / inten)) ** 2 +
                        (py / (np.cos(iy + tn) / inten)) ** 2)
        c += 1.0 / np.maximum(denom, 1e-6)
    c /= _WATER_MAX_ITER
    c = 1.17 - np.power(c, 1.4)
    b = np.clip(np.power(np.abs(c), 15.0) * 1.2, 0, 1)
    if _WATER_SCALE > 1:
        b = np.repeat(np.repeat(b, _WATER_SCALE, axis=0), _WATER_SCALE, axis=1)[:height, :width]
    return b


def water_caustic_surface(width, height, time_s):
    bright = compute_water_caustic(width, height, time_s)
    argb = np.zeros((height, width, 4), dtype=np.uint8)
    argb[:, :, 0] = (bright * 0.75 * 255).astype(np.uint8)  # B
    argb[:, :, 1] = (bright * 0.55 * 255).astype(np.uint8)  # G
    argb[:, :, 2] = (bright * 0.10 * 255).astype(np.uint8)  # R
    argb[:, :, 3] = (bright * 0.25 * 255).astype(np.uint8)  # A
    surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
    stride = surf.get_stride()
    buf = surf.get_data()
    for row in range(height):
        buf[row * stride: row * stride + width * 4] = argb[row].tobytes()
    surf.mark_dirty()
    return surf


# ══════════════════════════════════════════════════
# Main Window
# ══════════════════════════════════════════════════

class TickerWindow(Gtk.ApplicationWindow):
    def __init__(self, preset_name="ambient", **kwargs):
        super().__init__(**kwargs)
        self.set_title("keybind-ticker")
        self.set_default_size(1200, BAR_H)
        self.preset = load_preset(preset_name)

        if LAYER_MODE:
            Gtk4LayerShell.init_for_window(self)
            Gtk4LayerShell.set_layer(self, Gtk4LayerShell.Layer.BOTTOM)
            for edge in (Gtk4LayerShell.Edge.BOTTOM, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT):
                Gtk4LayerShell.set_anchor(self, edge, True)
            Gtk4LayerShell.set_exclusive_zone(self, BAR_H)
            Gtk4LayerShell.set_namespace(self, "keybind-ticker")
            display = Gdk.Display.get_default()
            if display:
                for i in range(display.get_monitors().get_n_items()):
                    mon = display.get_monitors().get_item(i)
                    if mon and "DP-3" in (mon.get_connector() or ""):
                        Gtk4LayerShell.set_monitor(self, mon)
                        break

        self.da = Gtk.DrawingArea()
        self.da.set_content_height(BAR_H)
        self.da.set_vexpand(True)
        self.da.set_hexpand(True)
        self.da.set_draw_func(self._draw)
        self.set_child(self.da)

        # ── Interactive controls ──────────────────
        scroll_ctl = Gtk.EventControllerScroll.new(Gtk.EventControllerScrollFlags.VERTICAL)
        scroll_ctl.connect("scroll", self._on_scroll)
        self.da.add_controller(scroll_ctl)

        click_ctl = Gtk.GestureClick.new()
        click_ctl.connect("pressed", self._on_click)
        self.da.add_controller(click_ctl)

        self.da.set_has_tooltip(True)
        self.da.connect("query-tooltip", self._on_tooltip)

        # ── State ─────────────────────────────────
        self.segments = []
        self.offset = 0.0
        self.gradient_phase = 0.0
        self.time_s = 0.0
        self.last_us = None
        self.layout = None
        self.half_w = 1.0

        self.glitch_remaining = 0
        self.glitch_strips = []
        self.water_surf = None
        self.water_frame = 0
        self.glow_cache = None
        self.glow_frame = 0

        # ── Content stream state ──────────────────
        self.stream_order = list(STREAM_ORDER)
        random.shuffle(self.stream_order)
        self.stream_idx = self._restore_stream_idx()

        self._rebuild_stream()
        GLib.timeout_add_seconds(REFRESH_S, self._advance_stream)
        self.da.add_tick_callback(self._tick)

    def _restore_stream_idx(self):
        try:
            with open(os.path.join(STATE_DIR, "current-stream")) as f:
                name = f.read().strip()
            if name in self.stream_order:
                return self.stream_order.index(name)
        except (OSError, ValueError):
            pass
        return 0

    def _advance_stream(self):
        self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
        self._rebuild_stream()
        return GLib.SOURCE_CONTINUE

    def _rebuild_stream(self):
        stream_name = self.stream_order[self.stream_idx]
        builder = STREAMS.get(stream_name, build_keybinds_markup)
        self.ticker_markup, self.segments = builder()
        self.layout = None
        self.da.queue_draw()
        # Persist current stream
        os.makedirs(STATE_DIR, exist_ok=True)
        try:
            with open(os.path.join(STATE_DIR, "current-stream"), "w") as f:
                f.write(stream_name)
        except OSError:
            pass

    def _on_scroll(self, controller, dx, dy):
        p = self.preset
        p.speed = max(10.0, min(200.0, p.speed - dy * 5.0))
        return True

    def _on_click(self, gesture, n_press, x, y):
        if not self.segments:
            return
        # Estimate which segment was clicked based on x position + scroll offset
        n = len(self.segments)
        if n == 0 or self.half_w <= 0:
            return
        seg_width = self.half_w / n
        abs_x = (x + self.offset) % self.half_w
        idx = int(abs_x / seg_width)
        idx = min(idx, n - 1)
        text = self.segments[idx]
        try:
            subprocess.Popen(["wl-copy", text],
                             stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        except Exception:
            pass

    def _on_tooltip(self, widget, x, y, keyboard_mode, tooltip):
        stream = self.stream_order[self.stream_idx]
        seg_info = ""
        if self.segments and self.half_w > 0:
            n = len(self.segments)
            seg_width = self.half_w / n
            abs_x = (x + self.offset) % self.half_w
            idx = min(int(abs_x / seg_width), n - 1)
            seg_info = f'\n<b>Keybind:</b> {escape(self.segments[idx])}\n<i>Click to copy</i>'
        tooltip.set_markup(
            f'<b>Stream:</b> {stream}\n'
            f'<b>Speed:</b> {self.preset.speed:.0f} px/s\n'
            f'<b>Scroll:</b> wheel to adjust speed'
            f'{seg_info}'
        )
        return True

    def _tick(self, widget, frame_clock):
        now = frame_clock.get_frame_time()
        p = self.preset
        if self.last_us is not None:
            dt = min((now - self.last_us) / 1_000_000.0, 0.05)
            self.offset += p.speed * dt
            self.gradient_phase += p.gradient_speed * dt
            self.time_s += dt
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
            if self.gradient_phase >= p.gradient_span:
                self.gradient_phase -= p.gradient_span
            # Glitch
            if self.glitch_remaining > 0:
                self.glitch_remaining -= 1
            elif p.glitch_prob > 0 and random.random() < p.glitch_prob:
                self.glitch_remaining = p.glitch_frames
                self.glitch_strips = [
                    (random.randint(0, BAR_H - 4), random.randint(3, 8),
                     random.randint(-10, 10))
                    for _ in range(3)
                ]
        self.last_us = now
        self.water_frame += 1
        widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _render_text_surface(self, width, height):
        surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        tc = _cairo.Context(surf)
        x = -self.offset
        y = (height - 14) / 2
        p = self.preset

        # Dark stroke outline for contrast over caustic background
        if getattr(p, 'outline_width', 0.8) > 0:
            ow = getattr(p, 'outline_width', 0.8)
            tc.set_source_rgba(0.02, 0.03, 0.05, 0.6)
            tc.set_line_width(ow)
            tc.set_line_join(_cairo.LINE_JOIN_ROUND)
            tc.move_to(x, y)
            PangoCairo.update_layout(tc, self.layout)
            PangoCairo.layout_path(tc, self.layout)
            tc.move_to(x + self.half_w, y)
            PangoCairo.update_layout(tc, self.layout)
            PangoCairo.layout_path(tc, self.layout)
            tc.stroke()

        # Gradient-filled text on top
        text_grad = make_gradient(x, self.half_w, self.gradient_phase)
        tc.set_source(text_grad)
        tc.move_to(x, y)
        PangoCairo.update_layout(tc, self.layout)
        PangoCairo.show_layout(tc, self.layout)
        tc.move_to(x + self.half_w, y)
        PangoCairo.update_layout(tc, self.layout)
        PangoCairo.show_layout(tc, self.layout)
        surf.flush()
        return surf

    def _apply_wave(self, surf, width, height):
        p = self.preset
        amp = getattr(p, 'wave_amp', 0)
        if amp <= 0:
            return surf
        stride = surf.get_stride()
        src = np.frombuffer(surf.get_data(), dtype=np.uint8).copy().reshape(height, stride)
        out = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        out_stride = out.get_stride()
        dst = np.zeros((height, out_stride), dtype=np.uint8)
        freq = getattr(p, 'wave_freq', 0.015)
        speed = getattr(p, 'wave_speed', 2.0)
        strip_w = 4
        for x0 in range(0, width, strip_w):
            x1 = min(x0 + strip_w, width)
            cx = (x0 + x1) // 2
            dy = int(round(amp * math.sin(freq * cx + speed * self.time_s)))
            bx0, bx1 = x0 * 4, x1 * 4
            if dy == 0:
                dst[:, bx0:bx1] = src[:, bx0:bx1]
            elif dy > 0:
                dst[dy:, bx0:bx1] = src[:height - dy, bx0:bx1]
            else:
                dst[:height + dy, bx0:bx1] = src[-dy:, bx0:bx1]
        buf = out.get_data()
        buf[:] = dst.tobytes()
        out.mark_dirty()
        return out

    def _compute_glow(self, text_surf, width, height):
        p = self.preset
        if p.glow_kernel < 3:
            return None
        stride = text_surf.get_stride()
        argb = np.frombuffer(text_surf.get_data(), dtype=np.uint8).reshape(height, stride // 4, 4)
        alpha = argb[:, :width, 3].astype(np.float32)
        blurred = uniform_filter1d(alpha, size=p.glow_kernel, axis=1, mode='constant')
        blurred = np.clip(blurred, 0, 255).astype(np.uint8)
        glow = _cairo.ImageSurface(_cairo.FORMAT_A8, width, height)
        gs = glow.get_stride()
        gb = glow.get_data()
        for row in range(height):
            gb[row * gs: row * gs + width] = blurred[row].tobytes()
        glow.mark_dirty()
        return glow

    def _draw(self, da, cr, width, height):
        p = self.preset

        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            self.layout.set_markup(self.ticker_markup, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = max(logical.width / 2.0, 1.0)
            self.offset = self.offset % self.half_w

        # ═══ BACKGROUND ═══════════════════════════
        cr.set_source_rgba(0.020, 0.027, 0.051, p.bg_alpha)
        cr.paint()

        # Water caustic
        if p.water_skip > 0:
            if self.water_surf is None or self.water_frame % p.water_skip == 0:
                self.water_surf = water_caustic_surface(width, height, self.time_s)
            cr.set_source_surface(self.water_surf, 0, 0)
            cr.paint()

        # Scanlines
        if p.scanline_opacity > 0:
            cr.set_source_rgba(0, 0, 0, p.scanline_opacity)
            for row in range(0, height, 2):
                cr.rectangle(0, row, width, 1)
            cr.fill()

        # Top border
        border_grad = make_gradient(0, width, self.gradient_phase)
        cr.set_source(border_grad)
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # ═══ FOREGROUND ═══════════════════════════
        text_surf = self._render_text_surface(width, height)
        text_surf = self._apply_wave(text_surf, width, height)

        # Glow (cached — recompute every 8 frames)
        self.glow_frame += 1
        if self.glow_cache is None or self.glow_frame % 8 == 0:
            self.glow_cache = self._compute_glow(text_surf, width, height)
        glow_a8 = self.glow_cache
        if glow_a8 is not None:
            glow_alpha = p.glow_base_alpha + p.glow_pulse_amp * math.sin(
                self.time_s * (2 * math.pi / p.glow_pulse_period))
            cr.save()
            cr.set_source(make_gradient(0, width, self.gradient_phase))
            cr.mask_surface(glow_a8, 0, 0)
            cr.restore()

        # Shadow
        if p.shadow_alpha > 0:
            cr.set_source_rgba(0, 0, 0, p.shadow_alpha)
            cr.mask_surface(text_surf, p.shadow_offset, p.shadow_offset)

        # Sharp text
        cr.set_source_surface(text_surf, 0, 0)
        cr.paint()

        # Glitch + CA
        if self.glitch_remaining > 0 and p.ca_offset > 0:
            cr.save()
            cr.set_source_rgba(1, 0, 0, 0.3)
            cr.mask_surface(text_surf, p.ca_offset, 0)
            cr.restore()
            cr.save()
            cr.set_source_rgba(0, 0.3, 1, 0.3)
            cr.mask_surface(text_surf, -p.ca_offset, 0)
            cr.restore()
            for sy, sh, sx in self.glitch_strips:
                cr.save()
                cr.rectangle(0, sy, width, sh)
                cr.clip()
                cr.set_source_surface(text_surf, sx, 0)
                cr.paint()
                cr.restore()


class TickerApp(Gtk.Application):
    def __init__(self):
        super().__init__(application_id="io.hairglasses.keybind_ticker")

    def do_activate(self):
        preset = "ambient"
        for i, arg in enumerate(sys.argv):
            if arg == "--preset" and i + 1 < len(sys.argv):
                preset = sys.argv[i + 1]
        win = TickerWindow(preset_name=preset, application=self)
        win.present()


if __name__ == "__main__":
    TickerApp().run()
