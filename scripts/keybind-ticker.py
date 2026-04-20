#!/usr/bin/env python3
"""keybind-ticker.py — Cyberpunk multi-stream ticker for Hyprland (v3).

GTK4 DrawingArea with PangoCairo at 240Hz frame-clock sync.

Features:
  - Multi-stream content (16 streams): keybinds, system, fleet, weather,
    github, notifications, music, updates, mx-battery, disk, load, workspace,
    claude-sessions, network, audio, shader
  - Visual effects: water caustic, neon glow (breathing), gradient text,
    scanlines, text outline (solid or color-cycling pulse), wave distortion,
    edge fade vignette, typewriter reveal, urgency glitch surge, phosphor
    decay trail, synthwave sweep top border, holo shimmer, ghost echo,
    progress indicator
  - Interactive: scroll (speed) / Shift+scroll (stream switch) /
    middle-click (pause) / right-click (menu) / left-click (copy/open URL),
    hover tooltip, adaptive speed on hover
  - 4 effect presets: ambient, cyberpunk, minimal, clean
  - Per-stream preset + per-stream refresh interval
  - Slow streams (github, music) run on background thread
  - Priority interrupts: critical urgency notifications jump the queue
  - Playlist persistence + order-preserving restore

Usage:
  keybind-ticker.py                     # regular window
  keybind-ticker.py --layer             # layer-shell bar (for systemd)
  keybind-ticker.py --preset minimal    # start with a preset
  keybind-ticker.py --monitor DP-2      # target a different output
  keybind-ticker.py --playlist focus    # load a non-default playlist
  keybind-ticker.py --state-dir ~/.local/state/keybind-ticker-dp3  # isolate state for multi-instance
"""

import gi
import subprocess
import json
import sys
import os
import re
import math
import random
import signal
import glob
import threading
import time as _time
import cairo as _cairo
import numpy as np
from collections import deque
from scipy.ndimage import uniform_filter1d
from html import escape

# Shared rendering helpers (badge / empty / dup / layer-shell / drawing-area).
# Import path handling mirrors the other ticker scripts.
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
import ticker_render as tr  # noqa: E402

gi.require_version("Gtk", "4.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gdk, Pango, PangoCairo, GLib, Gio

# ── CLI arg parsing ───────────────────────────────
LAYER_MODE = "--layer" in sys.argv


def _cli_value(flag, default=None):
    for i, arg in enumerate(sys.argv):
        if arg == flag and i + 1 < len(sys.argv):
            return sys.argv[i + 1]
    return default


MONITOR_NAME = _cli_value("--monitor", "DP-2")
START_PRESET = _cli_value("--preset", "ambient")
START_PLAYLIST = _cli_value("--playlist", None)  # None → resolved at runtime via state file or "main"
STATE_DIR_OVERRIDE = _cli_value("--state-dir", None)

if LAYER_MODE:
    gi.require_version("Gtk4LayerShell", "1.0")
    from gi.repository import Gtk4LayerShell


# ══════════════════════════════════════════════════
# Configuration
# ══════════════════════════════════════════════════

BAR_H = 28
DEFAULT_REFRESH_S = 300
ERROR_BACKOFF_S = 30  # Streams that returned _empty() advance after this instead of their full refresh interval, so a broken stream cannot freeze the ticker.
MAX_DWELL_S = 75  # Cap on how long any single stream stays on screen. Sized so content at ~55 px/s can traverse a 3840px ultrawide (~70s) before rotating. Prevents long cache TTLs (weather=1800, arch-news=3600) from freezing rotation.
MIN_DWELL_S = 12  # Floor so real-time streams (pomodoro, recording) stay readable.
STATE_DIR = os.path.expanduser(STATE_DIR_OVERRIDE) if STATE_DIR_OVERRIDE else os.path.expanduser("~/.local/state/keybind-ticker")
PLAYLIST_DIR = os.path.expanduser(
    "~/hairglasses-studio/dotfiles/ticker/content-playlists")
QUOTES_DIR = os.path.expanduser(
    "~/hairglasses-studio/dotfiles/ticker/quotes")
DEFAULT_PLAYLIST = "main"

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

URL_RE = re.compile(r"https?://\S+")

# ── Effect presets ────────────────────────────────
PRESETS = {
    "ambient": {
        "speed": 55.0, "gradient_speed": 40.0, "gradient_span": 800.0,
        "glow_kernel": 17, "glow_base_alpha": 0.35, "glow_pulse_amp": 0.15,
        "glow_pulse_period": 3.5, "water_skip": 120, "scanline_opacity": 0.08,
        "shadow_offset": 2, "shadow_alpha": 0.30,
        "glitch_prob": 0.004, "glitch_frames": 4, "ca_offset": 3,
        "bg_alpha": 0.82, "outline_width": 0.8, "outline_pulse": True,
        "wave_amp": 1.5, "wave_freq": 0.015, "wave_speed": 2.0,
        "edge_fade": 120, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": True, "holo_shimmer": 0.05,
        "hue_sweep": False,
    },
    "cyberpunk": {
        "speed": 65.0, "gradient_speed": 55.0, "gradient_span": 600.0,
        "glow_kernel": 21, "glow_base_alpha": 0.50, "glow_pulse_amp": 0.25,
        "glow_pulse_period": 2.5, "water_skip": 80, "scanline_opacity": 0.14,
        "shadow_offset": 3, "shadow_alpha": 0.35,
        "glitch_prob": 0.008, "glitch_frames": 5, "ca_offset": 4,
        "bg_alpha": 0.78, "outline_width": 1.0, "outline_pulse": True,
        "wave_amp": 2.0, "wave_freq": 0.02, "wave_speed": 3.0,
        "edge_fade": 100, "progress_bar": True, "phosphor_trail": 0.20,
        "ghost_echo": 0.08, "synthwave_border": True, "holo_shimmer": 0.10,
        "hue_sweep": True,
    },
    "minimal": {
        "speed": 45.0, "gradient_speed": 30.0, "gradient_span": 1000.0,
        "glow_kernel": 13, "glow_base_alpha": 0.20, "glow_pulse_amp": 0.05,
        "glow_pulse_period": 5.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 1, "shadow_alpha": 0.20,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.90, "outline_width": 0.5, "outline_pulse": False,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
        "edge_fade": 60, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": False, "holo_shimmer": 0.0,
        "hue_sweep": False,
    },
    "clean": {
        "speed": 50.0, "gradient_speed": 35.0, "gradient_span": 900.0,
        "glow_kernel": 0, "glow_base_alpha": 0.0, "glow_pulse_amp": 0.0,
        "glow_pulse_period": 1.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 0, "shadow_alpha": 0.0,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.92, "outline_width": 0.0, "outline_pulse": False,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
        "edge_fade": 0, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": False, "holo_shimmer": 0.0,
        "hue_sweep": False,
    },
}


def load_preset(name):
    p = PRESETS.get(name, PRESETS["ambient"])
    return type("P", (), dict(p))()


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


_badge = tr.badge
_empty = tr.empty
_dup = tr.dup


def build_keybinds_markup():
    try:
        raw = subprocess.run(
            ["hyprctl", "binds", "-j"],
            capture_output=True, text=True, timeout=5,
        ).stdout
        binds = json.loads(raw)
    except Exception:
        return _empty(" KEYBINDS", "#29f0ff", "no keybinds loaded")

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
    return _dup(single), segments


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
            parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">  GPU {escape(gpu[0])}W  {escape(gpu[1])}\u00b0C  {escape(gpu[2])}%  \u00b7</span>')
    except Exception:
        pass
    try:
        mem = subprocess.run(["free", "-m"], capture_output=True, text=True, timeout=2).stdout
        for line in mem.splitlines():
            if line.startswith("Mem:"):
                fields = line.split()
                total = int(fields[1])
                if total > 0:
                    pct = int(fields[2]) * 100 // total
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
    return _dup("".join(parts)), []


def build_fleet_markup():
    parts = [_badge("\U000f0168 FLEET", "#ff47d1")]
    try:
        with open("/tmp/rg-status.json") as f:
            data = json.load(f)
        fl = data.get("fleet", {})
        cost = data.get("cost", {})
        loops = data.get("loops", {})
        parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">'
                     f'  {int(fl.get("running",0))} running  {int(fl.get("completed",0))} done  '
                     f'{int(fl.get("failed",0))} failed  {int(fl.get("pending",0))} pending  \u00b7</span>')
        parts.append(f'<span font_desc="Maple Mono NF CN SemiBold 11">'
                     f'  {int(loops.get("total_runs",0))} loops  ${escape(str(cost.get("total_spend_usd",0)))}  \u00b7</span>')
        models = data.get("models", [])
        for m in models[:3]:
            model_name = escape(str(m.get("model", "")))
            count = int(m.get("count", 0))
            parts.append(f'<span font_desc="Maple Mono NF CN Italic 11">'
                         f'  {model_name} \u00d7{count}  \u00b7</span>')
    except Exception:
        return _empty("\U000f0168 FLEET", "#ff47d1", "no fleet data")
    return _dup("".join(parts)), []


def build_github_markup():
    TYPE_ICONS = {"PullRequest": "", "Issue": "", "Release": "",
                  "Discussion": "\U000f0361", "CheckSuite": ""}
    parts = [_badge(" GITHUB", "#3dffb5")]
    try:
        raw = subprocess.run(
            ["gh", "api", "/notifications", "--paginate", "--jq",
             '.[] | {type: .subject.type, title: .subject.title, repo: .repository.name, reason: .reason, url: .subject.url}'],
            capture_output=True, text=True, timeout=15,
        ).stdout.strip()
        if not raw:
            return _empty(" GITHUB", "#3dffb5", "no notifications")
        seen = 0
        fc = len(FONTS)
        segments = []
        for line in raw.splitlines():
            if seen >= 20:
                break
            try:
                n = json.loads(line)
            except Exception:
                continue
            icon = TYPE_ICONS.get(n.get("type", ""), "")
            title = escape(str(n.get("title", ""))[:60])
            repo = escape(str(n.get("repo", "")))
            font = FONTS[seen % fc]
            parts.append(f'<span font_desc="{font}">  {icon} {repo}: {title}  \u00b7</span>')
            # Convert API URL to HTML URL for xdg-open
            api_url = str(n.get("url", ""))
            html_url = api_url.replace("api.github.com/repos/", "github.com/")
            segments.append(html_url or "")
            seen += 1
        if seen == 0:
            return _empty(" GITHUB", "#3dffb5", "no notifications")
    except Exception:
        return _empty(" GITHUB", "#3dffb5", "github unavailable")
    return _dup("".join(parts)), segments


def build_notifications_markup():
    URGENCY_ICONS = {"critical": "\U000f0026", "normal": "\U000f009a", "low": "\U000f009e"}
    history = os.path.expanduser(
        "~/.local/state/dotfiles/desktop-control/notifications/history.jsonl")
    parts = [_badge("\U000f009a NOTIFICATIONS", "#ff5c8a")]
    has_critical = False
    try:
        with open(history) as f:
            lines = f.readlines()
        recent = lines[-30:] if len(lines) > 30 else lines
        recent.reverse()
        if not recent:
            return _empty("\U000f009a NOTIFICATIONS", "#ff5c8a", "no notification history")
        fc = len(FONTS)
        for i, line in enumerate(recent):
            try:
                n = json.loads(line)
            except Exception:
                continue
            urgency = str(n.get("urgency", ""))
            if urgency == "critical":
                has_critical = True
            icon = URGENCY_ICONS.get(urgency, "\U000f009a")
            app = escape(str(n.get("app", ""))[:20])
            summary = escape(str(n.get("summary", ""))[:40])
            body = escape(str(n.get("body", ""))[:40])
            font = FONTS[i % fc]
            text = f"{summary}: {body}" if body and body != summary else summary
            parts.append(f'<span font_desc="{font}">  {icon} {app} {text}  \u00b7</span>')
    except FileNotFoundError:
        return _empty("\U000f009a NOTIFICATIONS", "#ff5c8a", "no notification history")
    except Exception:
        return _empty("\U000f009a NOTIFICATIONS", "#ff5c8a", "notifications unavailable")
    # Side channel: flag the urgent_mode for the ticker to pick up
    markup = _dup("".join(parts))
    segments = ["__URGENT__"] if has_critical else []
    return markup, segments


def build_music_markup():
    # "no media playing" is a normal idle state, NOT an error. Return a real
    # segment instead of _empty so the error-backoff path doesn't
    # short-circuit the 10s refresh into a 30s loop, and so `hg ticker
    # health` doesn't mark this stream as faulting during quiet periods.
    parts = [_badge(" MUSIC", "#ff47d1")]
    try:
        status = subprocess.run(
            ["playerctl", "status"], capture_output=True, text=True, timeout=3
        ).stdout.strip()
    except Exception:
        status = ""
    if status not in ("Playing", "Paused"):
        parts.append(
            '<span font_desc="Maple Mono NF CN 11" foreground="#66708f">'
            '   idle  \u00b7</span>'
        )
        return _dup("".join(parts)), []
    try:
        icon = "" if status == "Playing" else ""
        meta = subprocess.run(
            ["playerctl", "metadata", "--format",
             "{{artist}} \u2014 {{title}} [{{album}}]"],
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
                     f'  {icon} {escape(meta)}  {escape(pos)}/{escape(dur)}  \u00b7</span>')
    except Exception:
        return _empty(" MUSIC", "#ff47d1", "playerctl unavailable")
    return _dup("".join(parts)), []


# ── New streams (Phase 3) ──────────────────────────

def build_updates_markup():
    parts = [_badge("\U000f0f8c UPDATES", "#29f0ff")]
    try:
        with open("/tmp/bar-updates.txt") as f:
            raw = f.read().strip()
        if not raw:
            return _empty("\U000f0f8c UPDATES", "#29f0ff", "no updates")
        parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">  {escape(raw)}  \u00b7</span>')
        # Add top package names if checkupdates is available
        try:
            pkgs = subprocess.run(
                ["checkupdates"], capture_output=True, text=True, timeout=3,
            ).stdout.strip().splitlines()
            fc = len(FONTS)
            for i, line in enumerate(pkgs[:15]):
                name = escape(line.split()[0]) if line else ""
                if name:
                    font = FONTS[i % fc]
                    parts.append(f'<span font_desc="{font}">  {name}  \u00b7</span>')
        except Exception:
            pass
    except FileNotFoundError:
        return _empty("\U000f0f8c UPDATES", "#29f0ff", "updates cache missing")
    except Exception:
        return _empty("\U000f0f8c UPDATES", "#29f0ff", "updates unavailable")
    return _dup("".join(parts)), []


def _sparkline(values, bars="\u2581\u2582\u2583\u2584\u2585\u2586\u2587\u2588"):
    if not values:
        return ""
    vmax = max(values) or 1
    out = ""
    for v in values:
        idx = min(len(bars) - 1, int((v / vmax) * (len(bars) - 1)))
        out += bars[idx]
    return out


_CPU_HWMON_CACHE: dict | None = None


def _find_cpu_hwmon():
    """Locate the k10temp/coretemp hwmon entry once and cache the path."""
    global _CPU_HWMON_CACHE
    if _CPU_HWMON_CACHE is not None:
        return _CPU_HWMON_CACHE
    for h in sorted(glob.glob("/sys/class/hwmon/hwmon*")):
        try:
            name = open(os.path.join(h, "name")).read().strip()
        except OSError:
            continue
        if name in ("k10temp", "coretemp", "zenpower"):
            _CPU_HWMON_CACHE = {"path": h, "name": name}
            return _CPU_HWMON_CACHE
    _CPU_HWMON_CACHE = {}
    return _CPU_HWMON_CACHE


def build_cpu_markup():
    parts = [_badge("\U000f04bc CPU", "#00d4ff")]
    try:
        hwmon = _find_cpu_hwmon()
        temps = []
        if hwmon:
            for lbl_file in sorted(glob.glob(os.path.join(hwmon["path"], "temp*_label"))):
                try:
                    label = open(lbl_file).read().strip()
                    inp = lbl_file[:-len("_label")] + "_input"
                    val = int(open(inp).read().strip()) // 1000
                    temps.append((label, val))
                except (OSError, ValueError):
                    continue

        freqs = []
        for fpath in sorted(glob.glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_cur_freq"))[:64]:
            try:
                freqs.append(int(open(fpath).read().strip()))
            except (OSError, ValueError):
                continue

        ncpu = os.cpu_count() or 1

        if not temps and not freqs:
            return _empty("\U000f04bc CPU", "#00d4ff", "no thermal/freq data")

        if temps:
            # Primary temp = Tctl (AMD) or Package id 0 (Intel), else first
            primary = next((v for l, v in temps if l in ("Tctl", "Package id 0")), temps[0][1])
            t_color = "#ff5c8a" if primary >= 85 else ("#ffe45e" if primary >= 75 else "#76ff03")
            parts.append(
                f'<span font_desc="Maple Mono NF CN Bold 11" foreground="{t_color}">'
                f'  {primary}°C  \u00b7</span>'
            )
            extra = [f"{l} {v}" for l, v in temps if l != "Tctl" and l != "Package id 0"][:3]
            if extra:
                parts.append(
                    f'<span font_desc="Maple Mono NF CN 11" foreground="#9fb2ff">'
                    f'  {escape(" ".join(extra))}  \u00b7</span>'
                )

        if freqs:
            avg_ghz = (sum(freqs) / len(freqs)) / 1_000_000
            max_ghz = max(freqs) / 1_000_000
            f_color = "#76ff03" if avg_ghz < 3.5 else ("#ffe45e" if avg_ghz < 4.5 else "#ff5c8a")
            parts.append(
                f'<span font_desc="Maple Mono NF CN 11" foreground="{f_color}">'
                f'  avg {avg_ghz:.2f} GHz  max {max_ghz:.2f} GHz  \u00b7</span>'
            )

        parts.append(
            f'<span font_desc="Maple Mono NF CN 11" foreground="#7ad0ff">'
            f'  {ncpu} threads  \u00b7</span>'
        )
    except Exception:
        return _empty("\U000f04bc CPU", "#00d4ff", "cpu unavailable")
    return _dup("".join(parts)), []


def build_gpu_markup():
    parts = [_badge("\U000f0a2d GPU", "#76ff03")]
    try:
        with open("/tmp/bar-gpu-full.txt") as f:
            raw = f.read().strip()
        if not raw or "|" not in raw:
            return _empty("\U000f0a2d GPU", "#76ff03", "no gpu data")
        fields = raw.split("|", 5)
        if len(fields) < 6:
            return _empty("\U000f0a2d GPU", "#76ff03", "gpu cache short")
        power, util, temp, mem_used, mem_total, name = fields
        try:
            util_n = int(util)
            temp_n = int(temp)
            mem_used_n = int(mem_used)
            mem_total_n = int(mem_total)
        except ValueError:
            return _empty("\U000f0a2d GPU", "#76ff03", "gpu cache parse")
        util_color = "#ff5c8a" if util_n > 85 else ("#ffe45e" if util_n > 50 else "#76ff03")
        temp_color = "#ff5c8a" if temp_n >= 80 else ("#ffe45e" if temp_n >= 70 else "#76ff03")
        mem_pct = (mem_used_n / mem_total_n * 100) if mem_total_n else 0
        mem_color = "#ff5c8a" if mem_pct > 90 else ("#ffe45e" if mem_pct > 70 else "#76ff03")
        spark = _sparkline([util_n / 100.0 for _ in range(6)])  # placeholder — single reading
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 11" foreground="{util_color}">'
            f'  {util_n}% util  \u00b7</span>'
            f'<span font_desc="Maple Mono NF CN 11" foreground="{temp_color}">'
            f'  {temp_n}°C  \u00b7</span>'
            f'<span font_desc="Maple Mono NF CN 11" foreground="#7ad0ff">'
            f'  {power}W  \u00b7</span>'
            f'<span font_desc="Maple Mono NF CN 11" foreground="{mem_color}">'
            f'  {mem_used_n}/{mem_total_n} MiB ({mem_pct:.0f}%)  \u00b7</span>'
            f'<span font_desc="Maple Mono NF CN 11" foreground="#9fb2ff">'
            f'  {escape(name)}  \u00b7</span>'
        )
    except FileNotFoundError:
        return _empty("\U000f0a2d GPU", "#76ff03", "gpu cache missing")
    except Exception:
        return _empty("\U000f0a2d GPU", "#76ff03", "gpu unavailable")
    return _dup("".join(parts)), []


def build_workspace_markup():
    parts = [_badge("\U000f0708 WORKSPACE", "#ff47d1")]
    try:
        aws = json.loads(subprocess.run(
            ["hyprctl", "activeworkspace", "-j"],
            capture_output=True, text=True, timeout=3,
        ).stdout)
        monitor = escape(str(aws.get("monitor", "")))
        name = escape(str(aws.get("name", "")))
        windows = int(aws.get("windows", 0))
        parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">'
                     f'  ws={name} on {monitor}  {windows} windows  \u00b7</span>')
        # active window
        aw = json.loads(subprocess.run(
            ["hyprctl", "activewindow", "-j"],
            capture_output=True, text=True, timeout=3,
        ).stdout)
        cls = escape(str(aw.get("class", ""))[:30])
        title = escape(str(aw.get("title", ""))[:50])
        if cls:
            parts.append(f'<span font_desc="Maple Mono NF CN Italic 11">  {cls}: {title}  \u00b7</span>')
        # all workspaces summary
        all_ws = json.loads(subprocess.run(
            ["hyprctl", "workspaces", "-j"],
            capture_output=True, text=True, timeout=3,
        ).stdout)
        total = len(all_ws)
        busy = sum(1 for w in all_ws if w.get("windows", 0) > 0)
        parts.append(f'<span font_desc="Maple Mono NF CN 11">  {busy}/{total} workspaces active  \u00b7</span>')
    except Exception:
        return _empty("\U000f0708 WORKSPACE", "#ff47d1", "hyprctl unavailable")
    return _dup("".join(parts)), []


# ── New streams (Phase 4) ──────────────────────────

def build_claude_sessions_markup():
    parts = [_badge(" CLAUDE", "#c084fc")]
    segments = []
    try:
        projects_dir = os.path.expanduser("~/.claude/projects")
        if not os.path.isdir(projects_dir):
            return _empty(" CLAUDE", "#c084fc", "no ~/.claude/projects")
        now = _time.time()
        cutoff = now - 86400  # last 24h
        project_stats = []
        for entry in os.scandir(projects_dir):
            if not entry.is_dir():
                continue
            session_mtimes = []
            try:
                for f in os.scandir(entry.path):
                    if f.name.endswith(".jsonl") and f.is_file():
                        st = f.stat()
                        if st.st_mtime >= cutoff:
                            session_mtimes.append(st.st_mtime)
            except OSError:
                continue
            if session_mtimes:
                project_stats.append((entry.name, len(session_mtimes), max(session_mtimes)))
        if not project_stats:
            return _empty(" CLAUDE", "#c084fc", "no recent sessions")
        project_stats.sort(key=lambda t: t[2], reverse=True)
        total_projects = len(project_stats)
        total_sessions = sum(s for _, s, _ in project_stats)
        parts.append(f'<span font_desc="Maple Mono NF CN Bold 11">'
                     f'  {total_projects} projects · {total_sessions} sessions  \u00b7</span>')
        STUDIO_PREFIX = "-home-hg-hairglasses-studio-"
        HOME_PREFIX = "-home-hg-"
        fc = len(FONTS)
        for i, (encoded, count, mtime) in enumerate(project_stats[:8]):
            if encoded.startswith(STUDIO_PREFIX):
                short = encoded[len(STUDIO_PREFIX):]
                display_path = f"~/hairglasses-studio/{short}"
            elif encoded.startswith(HOME_PREFIX):
                short = encoded[len(HOME_PREFIX):]
                display_path = f"~/{short}"
            else:
                short = encoded.lstrip("-")
                display_path = encoded
            age_s = now - mtime
            if age_s < 60:
                age = f"{int(age_s)}s"
            elif age_s < 3600:
                age = f"{int(age_s / 60)}m"
            else:
                age = f"{int(age_s / 3600)}h"
            font = FONTS[i % fc]
            parts.append(f'<span font_desc="{font}">  {escape(short)} ({count}×, {age} ago)  \u00b7</span>')
            segments.append(display_path)
    except Exception:
        return _empty(" CLAUDE", "#c084fc", "session scan failed")
    return _dup("".join(parts)), segments


def build_network_markup():
    parts = [_badge(" NET", "#38bdf8")]
    segments = []
    try:
        ssid = None
        signal = None
        try:
            wifi = subprocess.run(
                ["nmcli", "-t", "-f", "IN-USE,SSID,SIGNAL",
                 "device", "wifi", "list", "--rescan", "no"],
                capture_output=True, text=True, timeout=3,
            ).stdout.strip().splitlines()
            for line in wifi:
                if line.startswith(("*:", "*\\:")):
                    fields = line.split(":")
                    if len(fields) >= 3:
                        ssid = fields[1] or None
                        signal = fields[2] or None
                        break
        except Exception:
            pass
        ipv4 = None
        try:
            addrs = subprocess.run(
                ["ip", "-o", "-4", "addr", "show", "scope", "global"],
                capture_output=True, text=True, timeout=3,
            ).stdout.strip().splitlines()
            for line in addrs:
                fields = line.split()
                if len(fields) >= 4:
                    iface = fields[1]
                    if iface.startswith(("docker", "br-", "veth", "tailscale")):
                        continue
                    ipv4 = fields[3].split("/")[0]
                    break
        except Exception:
            pass
        ts_info = None
        try:
            ts = subprocess.run(
                ["ip", "-o", "-4", "addr", "show", "tailscale0"],
                capture_output=True, text=True, timeout=2,
            )
            if ts.returncode == 0 and ts.stdout.strip():
                fields = ts.stdout.strip().split()
                if len(fields) >= 4:
                    ts_ip = fields[3].split("/")[0]
                    ts_info = f"tailscale:{ts_ip}"
        except Exception:
            pass
        if not (ssid or ipv4 or ts_info):
            return _empty(" NET", "#38bdf8", "offline")
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
            parts.append(f'<span font_desc="{font}" foreground="{sig_color}">'
                         f'  {escape(ssid)} {escape(signal or "?")}%  \u00b7</span>')
            segments.append(ssid)
            ix += 1
        if ipv4:
            font = FONTS[ix % fc]
            parts.append(f'<span font_desc="{font}">  {escape(ipv4)}  \u00b7</span>')
            segments.append(ipv4)
            ix += 1
        if ts_info:
            font = FONTS[ix % fc]
            parts.append(f'<span font_desc="{font}" foreground="#c084fc">'
                         f'  {escape(ts_info)}  \u00b7</span>')
            segments.append(ts_info)
            ix += 1
    except Exception:
        return _empty(" NET", "#38bdf8", "network unavailable")
    return _dup("".join(parts)), segments


# ══════════════════════════════════════════════════
# Stream registry + metadata
# ══════════════════════════════════════════════════

STREAMS = {
    "keybinds":      build_keybinds_markup,
    "system":        build_system_markup,
    "fleet":         build_fleet_markup,
    "github":        build_github_markup,
    "notifications": build_notifications_markup,
    "music":         build_music_markup,
    "updates":       build_updates_markup,
    "cpu":           build_cpu_markup,
    "gpu":           build_gpu_markup,
    "workspace":     build_workspace_markup,
    "claude-sessions": build_claude_sessions_markup,
    "network":         build_network_markup,
}

# Per-stream metadata: effect preset override + refresh interval (seconds)
STREAM_META = {
    "keybinds":      {"preset": None,        "refresh": 300},
    "system":        {"preset": None,        "refresh": 10},
    "fleet":         {"preset": "cyberpunk", "refresh": 30},
    "github":        {"preset": None,        "refresh": 120},
    "notifications": {"preset": None,        "refresh": 60},
    "music":         {"preset": "minimal",   "refresh": 10},
    "updates":       {"preset": None,        "refresh": 1800},
    "cpu":           {"preset": "cyberpunk", "refresh": 10},
    "gpu":           {"preset": "cyberpunk", "refresh": 10},
    "workspace":     {"preset": None,        "refresh": 5},
    "claude-sessions": {"preset": None,        "refresh": 120},
    "network":         {"preset": None,        "refresh": 30},
}

# Streams whose builders can block for >100ms — run on background thread
SLOW_STREAMS = {"github", "music", "updates", "claude-sessions"}

FALLBACK_ORDER = [
    "keybinds", "system", "fleet", "weather", "github",
    "notifications", "music", "updates", "mx-battery",
    "disk", "load", "cpu", "gpu", "top-procs", "uptime", "tmux", "workspace",
    "claude-sessions", "network", "audio", "shader", "ci", "hacker",
    "calendar", "pomodoro", "token-burn", "dirty-repos", "failed-units",
    "arch-news", "smart-disk", "wifi-quality", "container-status",
    "net-throughput", "kernel-errors", "recording",
    "hn-top", "github-prs", "weather-alerts", "cve-alerts",
]


# ══════════════════════════════════════════════════
# Declarative TOML catalogue (ticker/streams.toml)
#
# Streams declared there override same-named inline builders; new names
# get appended to FALLBACK_ORDER. Keeps the data-definition boilerplate
# out of this file for cache-fed streams that follow a fixed pattern.
# ══════════════════════════════════════════════════

def _load_toml_catalogue():
    import ticker_streams as _ts
    toml_path = os.path.expanduser(
        "~/hairglasses-studio/dotfiles/ticker/streams.toml")
    try:
        toml_builders, toml_meta, toml_order = _ts.load_toml_streams(toml_path)
    except Exception as _e:
        sys.stderr.write(f"streams.toml load failed: {_e}\n")
        return
    for name, fn in toml_builders.items():
        STREAMS[name] = fn
    for name, m in toml_meta.items():
        STREAM_META[name] = {"preset": m.get("preset"),
                             "refresh": int(m.get("refresh", 300))}
        if "dwell" in m:
            STREAM_META[name]["dwell"] = m["dwell"]
    for name in toml_order:
        if name not in FALLBACK_ORDER:
            FALLBACK_ORDER.append(name)


_load_toml_catalogue()


# ══════════════════════════════════════════════════
# Bundled plugin catalogue (scripts/lib/ticker_streams/*.py)
#
# Complex streams live as individual Python modules with a standard
# META + build() contract. The loader imports each and registers the
# resulting callable into the same STREAMS / STREAM_META / FALLBACK_ORDER
# / SLOW_STREAMS structures the inline builders use. Plugins override
# same-named inline builders to allow incremental extraction.
# ══════════════════════════════════════════════════

def _load_bundled_plugins():
    import ticker_streams as _ts
    pkg_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)),
                           "lib", "ticker_streams")
    try:
        builders, meta, order, slow = _ts.load_plugin_streams(pkg_dir)
    except Exception as _e:
        sys.stderr.write(f"bundled plugin load failed: {_e}\n")
        return
    for name, fn in builders.items():
        STREAMS[name] = fn
    for name, m in meta.items():
        STREAM_META[name] = {"preset": m.get("preset"),
                             "refresh": int(m.get("refresh", 300))}
        if "dwell" in m:
            STREAM_META[name]["dwell"] = m["dwell"]
    for name in order:
        if name not in FALLBACK_ORDER:
            FALLBACK_ORDER.append(name)
    SLOW_STREAMS.update(slow)


_load_bundled_plugins()


# ══════════════════════════════════════════════════
# Plugin loader (Phase E1)
#
# Drop a file at ~/.config/keybind-ticker/plugins/<name>.py that defines:
#   META = {"preset": "ambient" | None, "refresh": <int seconds>}
#   def build_markup() -> tuple[str, list[str]]
# The stream auto-registers as "plugin:<name>". If the plugin import fails
# we log a warning and skip it — plugins must never be able to crash the
# ticker.
# ══════════════════════════════════════════════════

def _load_plugins():
    import importlib.util as _ilu
    plugin_dir = os.path.expanduser("~/.config/keybind-ticker/plugins")
    if not os.path.isdir(plugin_dir):
        return
    for entry in sorted(os.listdir(plugin_dir)):
        if not entry.endswith(".py") or entry.startswith("_"):
            continue
        path = os.path.join(plugin_dir, entry)
        name = f"plugin:{entry[:-3]}"
        try:
            spec = _ilu.spec_from_file_location(f"_kt_plugin_{entry[:-3]}", path)
            if spec is None or spec.loader is None:
                continue
            mod = _ilu.module_from_spec(spec)
            spec.loader.exec_module(mod)
            if not hasattr(mod, "build_markup"):
                continue
            STREAMS[name] = mod.build_markup
            STREAM_META[name] = {
                "preset": getattr(mod, "META", {}).get("preset"),
                "refresh": int(getattr(mod, "META", {}).get("refresh", 60)),
            }
            if name not in FALLBACK_ORDER:
                FALLBACK_ORDER.append(name)
        except Exception as _e:
            sys.stderr.write(f"plugin load failed: {entry}: {_e}\n")


_load_plugins()


# ══════════════════════════════════════════════════
# Per-stream click action dispatch (Phase A3)
#
# Each entry is a callable (win, segment_idx, segment_text) → bool. Return
# True to claim the click and skip the generic URL/clipboard fallback.
# Actions run on the main loop so they can freely call widget APIs.
# ══════════════════════════════════════════════════

def _spawn_detached(cmd):
    try:
        subprocess.Popen(cmd, stdout=subprocess.DEVNULL,
                         stderr=subprocess.DEVNULL, start_new_session=True)
        return True
    except Exception:
        return False


def _click_ci(win, idx, text):
    # CI segments look like "✓ <repo>" / "✗ <repo>" / "⏳ <repo>". If the
    # clicked token has a repo name, open its latest workflow runs in gh.
    for token in text.split():
        if "/" in token or token.isidentifier():
            repo = token.strip(":·")
            return _spawn_detached(["gh", "run", "list", "--repo",
                                    f"hairglasses-studio/{repo}", "--web"])
    return False


def _click_arch_news(win, idx, text):
    return _spawn_detached(["xdg-open", "https://archlinux.org/news/"])


def _click_calendar(win, idx, text):
    return _spawn_detached(["xdg-open", "https://calendar.google.com/"])


def _click_updates(win, idx, text):
    # Opening the pacman log is more useful than copying segment text.
    return _spawn_detached(["xdg-open", "/var/log/pacman.log"])


def _click_github(win, idx, text):
    # github stream already renders URL tokens — let the generic fallback
    # handle them. But if the segment is a bare notification title, open the
    # notifications dashboard.
    if text.startswith(("http://", "https://")):
        return False  # generic URL open handles it
    return _spawn_detached(["xdg-open", "https://github.com/notifications"])


def _click_hn(win, idx, text):
    # Filled in by Phase C1 (hn-top stream). Tries to extract the HN item id
    # from the segment prefix "#<id>"; falls back to HN front page.
    import re as _re
    m = _re.search(r"#(\d+)", text)
    if m:
        return _spawn_detached(
            ["xdg-open", f"https://news.ycombinator.com/item?id={m.group(1)}"]
        )
    return _spawn_detached(["xdg-open", "https://news.ycombinator.com/"])


STREAM_CLICK_ACTIONS = {
    "ci":        _click_ci,
    "arch-news": _click_arch_news,
    "calendar":  _click_calendar,
    "updates":   _click_updates,
    "github":    _click_github,
    "hn-top":    _click_hn,
}


def _playlist_path(name):
    return os.path.join(PLAYLIST_DIR, f"{name}.txt")


def list_playlists():
    """Enumerate available playlist names from PLAYLIST_DIR."""
    try:
        names = sorted(
            f[:-4] for f in os.listdir(PLAYLIST_DIR)
            if f.endswith(".txt") and not f.startswith(".")
        )
        return names or [DEFAULT_PLAYLIST]
    except OSError:
        return [DEFAULT_PLAYLIST]


def resolve_playlist_name():
    """Pick the active playlist name: CLI > state file > default."""
    if START_PLAYLIST:
        return START_PLAYLIST
    try:
        with open(os.path.join(STATE_DIR, "active-playlist")) as f:
            name = f.read().strip()
        if name:
            return name
    except OSError:
        pass
    return DEFAULT_PLAYLIST


def load_playlist(name=None):
    """Read playlist <name>.txt; fall back to FALLBACK_ORDER if missing/empty."""
    name = name or resolve_playlist_name()
    try:
        with open(_playlist_path(name)) as f:
            lines = [line.strip() for line in f]
        order = [l for l in lines
                 if l and not l.startswith("#") and l in STREAMS]
        if order:
            return order
    except OSError:
        pass
    return list(FALLBACK_ORDER)


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
        self._base_preset_name = preset_name
        self.preset = load_preset(preset_name)

        target_width = 1200  # windowed-mode fallback; layer-shell overrides below
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
                    if mon and MONITOR_NAME in (mon.get_connector() or ""):
                        Gtk4LayerShell.set_monitor(self, mon)
                        geom = mon.get_geometry()
                        if geom:
                            target_width = max(geom.width, 1)
                        break
        self.set_default_size(target_width, BAR_H)

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

        # Click: left button + middle button handled in same controller with button-check
        click_ctl = Gtk.GestureClick.new()
        click_ctl.set_button(0)  # all buttons
        click_ctl.connect("pressed", self._on_click)
        self.da.add_controller(click_ctl)

        motion_ctl = Gtk.EventControllerMotion.new()
        motion_ctl.connect("enter", self._on_enter)
        motion_ctl.connect("leave", self._on_leave)
        motion_ctl.connect("motion", self._on_motion)
        self.da.add_controller(motion_ctl)

        self.da.set_has_tooltip(True)
        self.da.connect("query-tooltip", self._on_tooltip)

        # ── State ─────────────────────────────────
        self.segments = []
        self.segment_bounds = []  # [(x_start, x_end, text), ...] in layout coords
        self.offset = 0.0
        self.gradient_phase = 0.0
        self.time_s = 0.0
        self.last_us = None
        self.layout = None
        self.half_w = 1.0

        self.paused = self._load_pause_state()
        self.pinned_stream = self._load_pin_state()
        self.urgent_mode = False
        self.urgent_timer_handle = None

        self.hover_x = 0
        self.is_hovering = False
        self.base_speed_cache = self.preset.speed

        self.reveal_chars = 99999  # no typewriter by default; set smaller to animate

        self.glitch_remaining = 0
        self.glitch_strips = []
        self.water_surf = None
        self.water_frame = 0
        self.glow_cache = None
        self.glow_frame = 0

        self.stream_start_s = 0.0

        # Phosphor trail ring buffer (for Phase 6)
        self.phosphor_ring = deque(maxlen=4)
        self.phosphor_frame_ctr = 0

        # ── Content stream state ──────────────────
        self.playlist_name = resolve_playlist_name()
        self.stream_order = load_playlist(self.playlist_name)
        self.stream_idx = self._restore_stream_idx()

        # Background thread state for slow streams
        self._bg_lock = threading.Lock()
        self._bg_result = {}  # stream_name -> (markup, segments)
        self._bg_inflight = set()

        # Per-stream error flag → drives short-backoff scheduling.
        self._stream_errored = {}
        # Per-stream health counters (Phase A4). Populated on each absorb;
        # exported to /tmp/ticker-health.json every 30s for `hg ticker health`.
        self._stream_health = {}

        self._rebuild_stream()
        self._schedule_next_advance()
        self.da.add_tick_callback(self._tick)

        # Force the frame clock to tick continuously. Without this, Hyprland
        # stops delivering frame callbacks when the layer-shell surface is
        # "idle" (no interaction), which freezes every time-driven animation
        # (water caustic, glow breathing, phosphor trail, glitch, gradient)
        # until the user hovers or clicks. begin_updating keeps the clock live.
        self.da.connect("realize", self._on_da_realize)
        if self.da.get_realized():
            self._on_da_realize(self.da)

        # Priority-interrupt listener (watches notification history file)
        self._last_notif_mtime = 0.0
        GLib.timeout_add_seconds(3, self._check_priority_interrupt)

        # Phase A4: export per-stream health snapshot every 30s.
        GLib.timeout_add_seconds(30, self._write_health_snapshot)

    # ── Persistence helpers ─────────────────────

    def reload_from_state(self):
        """Re-read state files (active-playlist, pinned-stream, paused) and
        apply the resulting configuration without restarting. Intended to be
        called via GLib.idle_add from a SIGUSR1 handler so external tools
        (hg ticker playlist / pin / pause) can switch the ticker without
        flickering. Runs on the main loop so it's safe to call widget APIs."""
        # Playlist
        try:
            with open(os.path.join(STATE_DIR, "active-playlist")) as f:
                desired_playlist = f.read().strip()
        except OSError:
            desired_playlist = self.playlist_name
        if desired_playlist and desired_playlist != self.playlist_name:
            new_order = load_playlist(desired_playlist)
            if new_order:
                self.playlist_name = desired_playlist
                self.stream_order = new_order
                self.stream_idx = 0
                self.pinned_stream = None
                self._save_pin_state()
                self._rebuild_stream()
                self._schedule_next_advance()
        # Pinned stream
        desired_pin = self._load_pin_state()
        if desired_pin != self.pinned_stream:
            self.pinned_stream = desired_pin
            if desired_pin and desired_pin in self.stream_order:
                self.stream_idx = self.stream_order.index(desired_pin)
                self._rebuild_stream()
                self._schedule_next_advance()
        # Paused
        desired_paused = self._load_pause_state()
        if desired_paused != self.paused:
            self.paused = desired_paused
        return False  # one-shot idle_add

    def _load_pause_state(self):
        return os.path.exists(os.path.join(STATE_DIR, "paused"))

    def _save_pause_state(self):
        os.makedirs(STATE_DIR, exist_ok=True)
        path = os.path.join(STATE_DIR, "paused")
        if self.paused:
            open(path, "w").close()
        else:
            try:
                os.remove(path)
            except FileNotFoundError:
                pass

    def _load_pin_state(self):
        try:
            with open(os.path.join(STATE_DIR, "pinned-stream")) as f:
                name = f.read().strip()
                if name in STREAMS:
                    return name
        except (OSError, ValueError):
            pass
        return None

    def _save_pin_state(self):
        os.makedirs(STATE_DIR, exist_ok=True)
        path = os.path.join(STATE_DIR, "pinned-stream")
        if self.pinned_stream:
            with open(path, "w") as f:
                f.write(self.pinned_stream)
        else:
            try:
                os.remove(path)
            except FileNotFoundError:
                pass

    def _restore_stream_idx(self):
        try:
            with open(os.path.join(STATE_DIR, "current-stream")) as f:
                name = f.read().strip()
            if name in self.stream_order:
                return self.stream_order.index(name)
        except (OSError, ValueError):
            pass
        return 0

    # ── Stream scheduling ─────────────────────

    def _current_interval(self, stream_name):
        if self._stream_errored.get(stream_name):
            return ERROR_BACKOFF_S
        meta = STREAM_META.get(stream_name, {})
        raw = meta.get("dwell", meta.get("refresh", DEFAULT_REFRESH_S))
        return max(MIN_DWELL_S, min(raw, MAX_DWELL_S))

    def _absorb_segments(self, stream_name, segments):
        """Strip control sentinels and update error/urgent state."""
        now = int(_time.time())
        health = self._stream_health.setdefault(
            stream_name,
            {"last_ok": 0, "last_err": 0, "consecutive_err": 0, "total_err": 0, "total_ok": 0},
        )
        if segments == ["__EMPTY__"]:
            self._stream_errored[stream_name] = True
            health["last_err"] = now
            health["consecutive_err"] += 1
            health["total_err"] += 1
            return []
        if segments == ["__URGENT__"]:
            self._stream_errored[stream_name] = False
            health["last_ok"] = now
            health["consecutive_err"] = 0
            health["total_ok"] += 1
            self._trigger_urgent_mode()
            return []
        self._stream_errored[stream_name] = False
        health["last_ok"] = now
        health["consecutive_err"] = 0
        health["total_ok"] += 1
        return segments

    def _write_health_snapshot(self):
        """Export stream health counters to tmpfs for `hg ticker health`."""
        import json
        path = "/tmp/ticker-health.json"
        tmp = path + ".tmp"
        try:
            with open(tmp, "w") as f:
                json.dump({
                    "pid": os.getpid(),
                    "ts": int(_time.time()),
                    "playlist": self.playlist_name,
                    "streams": self._stream_health,
                }, f)
            os.replace(tmp, path)
        except OSError:
            pass
        return True  # keep the timer alive

    def _schedule_next_advance(self):
        current = self.stream_order[self.stream_idx]
        interval = self._current_interval(current)
        GLib.timeout_add_seconds(interval, self._maybe_advance_stream,
                                 current)

    def _maybe_advance_stream(self, from_stream):
        # Only advance if we're still on this stream
        if self.stream_order[self.stream_idx] != from_stream:
            return GLib.SOURCE_REMOVE
        if self.pinned_stream:
            # Just rebuild, don't advance
            self._rebuild_stream()
        else:
            self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
            self._rebuild_stream()
        self._schedule_next_advance()
        return GLib.SOURCE_REMOVE

    def _rebuild_stream(self, force_name=None):
        stream_name = force_name or (self.pinned_stream
                                     if self.pinned_stream
                                     else self.stream_order[self.stream_idx])
        # Persist current stream
        os.makedirs(STATE_DIR, exist_ok=True)
        try:
            with open(os.path.join(STATE_DIR, "current-stream"), "w") as f:
                f.write(stream_name)
        except OSError:
            pass
        # Apply per-stream preset override
        meta = STREAM_META.get(stream_name, {})
        target_preset = meta.get("preset") or self._base_preset_name
        self._apply_preset(target_preset)
        # Stream start timestamp for progress bar
        self.stream_start_s = self.time_s
        # Typewriter reveal animation: start from 0
        self.reveal_chars = 0

        if stream_name in SLOW_STREAMS:
            # Check if we have a cached result from last time
            with self._bg_lock:
                cached = self._bg_result.get(stream_name)
            if cached is not None:
                markup, segments = cached
            else:
                # Show placeholder while fetching
                markup, segments = _empty(
                    f" {stream_name.upper()}", "#66708f", "loading...")
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)
            self._dispatch_bg_fetch(stream_name)
        else:
            try:
                builder = STREAMS.get(stream_name, build_keybinds_markup)
                markup, segments = builder()
            except Exception:
                markup, segments = _empty(
                    stream_name.upper(), "#ff5c8a", "builder error")
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)

        self.layout = None
        self.segment_bounds = []
        self.da.queue_draw()

    def _dispatch_bg_fetch(self, stream_name):
        with self._bg_lock:
            if stream_name in self._bg_inflight:
                return
            self._bg_inflight.add(stream_name)

        def worker():
            builder = STREAMS.get(stream_name, build_keybinds_markup)
            try:
                result = builder()
            except Exception:
                result = _empty(stream_name.upper(), "#ff5c8a", "builder error")
            with self._bg_lock:
                self._bg_result[stream_name] = result
                self._bg_inflight.discard(stream_name)
            GLib.idle_add(self._bg_apply, stream_name, result)

        t = threading.Thread(target=worker, daemon=True)
        t.start()

    def _bg_apply(self, stream_name, result):
        # Apply result only if we're still showing this stream
        current = (self.pinned_stream if self.pinned_stream
                   else self.stream_order[self.stream_idx])
        if current == stream_name:
            markup, segments = result
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)
            self.layout = None
            self.segment_bounds = []
            self.da.queue_draw()
        return False  # idle_add -> run once

    # ── Urgency surge ─────────────────────────

    def _trigger_urgent_mode(self):
        self.urgent_mode = True
        # Cancel existing timer if any
        if self.urgent_timer_handle is not None:
            GLib.source_remove(self.urgent_timer_handle)
        self.urgent_timer_handle = GLib.timeout_add_seconds(
            10, self._clear_urgent_mode)

    def _clear_urgent_mode(self):
        self.urgent_mode = False
        self.urgent_timer_handle = None
        return GLib.SOURCE_REMOVE

    def _check_priority_interrupt(self):
        history = os.path.expanduser(
            "~/.local/state/dotfiles/desktop-control/notifications/history.jsonl")
        try:
            mtime = os.stat(history).st_mtime
        except OSError:
            return GLib.SOURCE_CONTINUE
        if mtime <= self._last_notif_mtime:
            return GLib.SOURCE_CONTINUE
        self._last_notif_mtime = mtime
        # Peek at last entry
        try:
            with open(history) as f:
                lines = f.readlines()
            if not lines:
                return GLib.SOURCE_CONTINUE
            n = json.loads(lines[-1])
            if n.get("urgency") == "critical":
                # Jump to notifications stream
                if "notifications" in self.stream_order:
                    self.stream_idx = self.stream_order.index("notifications")
                    self._rebuild_stream()
        except Exception:
            pass
        return GLib.SOURCE_CONTINUE

    # ── Preset application ────────────────────

    def _apply_preset(self, name):
        # Preserve runtime-adjusted speed
        new_preset = load_preset(name)
        self.preset = new_preset
        self.base_speed_cache = new_preset.speed

    # ── Input handlers ────────────────────────

    def _on_scroll(self, controller, dx, dy):
        # Shift+scroll: cycle streams instead of speed
        event = controller.get_current_event()
        shift = False
        if event is not None:
            state = event.get_modifier_state()
            shift = bool(state & Gdk.ModifierType.SHIFT_MASK)
        if shift:
            if dy > 0:
                self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
            elif dy < 0:
                self.stream_idx = (self.stream_idx - 1) % len(self.stream_order)
            self._rebuild_stream()
        else:
            p = self.preset
            p.speed = max(10.0, min(200.0, p.speed - dy * 5.0))
            self.base_speed_cache = p.speed
        return True

    def _on_enter(self, controller, x, y):
        self.is_hovering = True
        # Adaptive speed: slow to 20% on hover
        self.preset.speed = self.base_speed_cache * 0.2

    def _on_leave(self, controller):
        self.is_hovering = False
        self.preset.speed = self.base_speed_cache

    def _on_motion(self, controller, x, y):
        self.hover_x = x

    def _on_click(self, gesture, n_press, x, y):
        button = gesture.get_current_button()
        # Middle-click: toggle pause
        if button == 2:
            self.paused = not self.paused
            self._save_pause_state()
            gesture.set_state(Gtk.EventSequenceState.CLAIMED)
            return
        # Right-click: open context menu
        if button == 3:
            self._show_context_menu(x, y)
            gesture.set_state(Gtk.EventSequenceState.CLAIMED)
            return
        # Left-click: first try a per-stream custom action, then fall back
        # to URL-open / clipboard-copy.
        if button == 1:
            idx = self._segment_at_x(x)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx]
            if not text:
                return
            stream = self.stream_order[self.stream_idx] if self.stream_order else None
            action = STREAM_CLICK_ACTIONS.get(stream) if stream else None
            if action is not None:
                try:
                    if action(self, idx, text):
                        return
                except Exception:
                    pass
            if URL_RE.match(text) or text.startswith(("http://", "https://")):
                try:
                    subprocess.Popen(["xdg-open", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass
            else:
                try:
                    subprocess.Popen(["wl-copy", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

    def _segment_at_x(self, x):
        """Accurate segment lookup using Pango byte-index mapping."""
        if not self.segment_bounds or self.half_w <= 0:
            # Fall back to uniform estimate if bounds not computed
            if not self.segments:
                return None
            seg_width = self.half_w / max(len(self.segments), 1)
            abs_x = (x + self.offset) % self.half_w
            return min(int(abs_x / seg_width), len(self.segments) - 1)
        abs_x = (x + self.offset) % self.half_w
        for i, (xs, xe, _text) in enumerate(self.segment_bounds):
            if xs <= abs_x < xe:
                return i
        return None

    def _show_context_menu(self, x, y):
        # Remember which segment the right-click landed on so the per-segment
        # menu items (Copy, Open URL, Dismiss) can target it.
        self._menu_segment_idx = self._segment_at_x(x)
        menu = Gio.Menu()

        # Per-segment actions at the top of the menu.
        seg_section = Gio.Menu()
        cur_stream = self.stream_order[self.stream_idx] if self.stream_order else ""
        seg_text = ""
        if (self._menu_segment_idx is not None
                and self._menu_segment_idx < len(self.segments)):
            seg_text = self.segments[self._menu_segment_idx] or ""
        if seg_text:
            preview = seg_text if len(seg_text) <= 32 else seg_text[:30] + "…"
            seg_section.append(f"Copy: {preview}", "ticker.copy_segment")
            if URL_RE.search(seg_text) or seg_text.startswith(("http://", "https://")):
                seg_section.append("Open URL", "ticker.open_segment")
            seg_section.append(f"Dismiss {cur_stream}", "ticker.dismiss_stream")
            menu.append_section("Segment", seg_section)

        stream_section = Gio.Menu()
        for name in self.stream_order:
            stream_section.append(name, f"ticker.stream::{name}")
        menu.append_section("Streams", stream_section)

        playlist_section = Gio.Menu()
        for name in list_playlists():
            label = f"{name} \u2713" if name == self.playlist_name else name
            playlist_section.append(label, f"ticker.playlist::{name}")
        menu.append_section("Playlists", playlist_section)

        preset_section = Gio.Menu()
        for name in PRESETS.keys():
            preset_section.append(name, f"ticker.preset::{name}")
        menu.append_section("Presets", preset_section)

        action_section = Gio.Menu()
        action_section.append("Pause" if not self.paused else "Resume",
                              "ticker.pause")
        if self.pinned_stream:
            action_section.append(f"Unpin ({self.pinned_stream})",
                                  "ticker.unpin")
        else:
            action_section.append("Pin current stream", "ticker.pin")
        menu.append_section("Actions", action_section)

        self._ensure_actions()

        popover = Gtk.PopoverMenu.new_from_model(menu)
        popover.set_parent(self.da)
        popover.set_pointing_to(Gdk.Rectangle(int(x), int(y), 1, 1))
        popover.popup()

    def _ensure_actions(self):
        if getattr(self, "_actions_registered", False):
            return
        group = Gio.SimpleActionGroup()

        def on_stream(action, param):
            name = param.get_string()
            if name in STREAMS and name in self.stream_order:
                self.stream_idx = self.stream_order.index(name)
                self._rebuild_stream()

        def on_preset(action, param):
            name = param.get_string()
            if name in PRESETS:
                self._base_preset_name = name
                self._apply_preset(name)
                self.layout = None
                self.da.queue_draw()

        def on_pause(action, param):
            self.paused = not self.paused
            self._save_pause_state()

        def on_pin(action, param):
            self.pinned_stream = self.stream_order[self.stream_idx]
            self._save_pin_state()

        def on_unpin(action, param):
            self.pinned_stream = None
            self._save_pin_state()

        def on_copy_segment(action, param):
            idx = getattr(self, "_menu_segment_idx", None)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx] or ""
            if text:
                try:
                    subprocess.Popen(["wl-copy", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

        def on_open_segment(action, param):
            idx = getattr(self, "_menu_segment_idx", None)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx] or ""
            m = URL_RE.search(text)
            url = m.group(0) if m else (text if text.startswith(("http://", "https://")) else None)
            if url:
                try:
                    subprocess.Popen(["xdg-open", url],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

        def on_dismiss_stream(action, param):
            # Advance past the current stream once; the playlist rotation
            # will bring it back next cycle. Useful for noisy streams you
            # want to skip for now.
            if self.stream_order:
                self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
                self._rebuild_stream()
                self._schedule_next_advance()

        def on_playlist(action, param):
            new_name = param.get_string()
            if new_name == self.playlist_name:
                return
            new_order = load_playlist(new_name)
            if not new_order:
                return
            self.playlist_name = new_name
            self.stream_order = new_order
            self.stream_idx = 0
            self.pinned_stream = None
            self._save_pin_state()
            os.makedirs(STATE_DIR, exist_ok=True)
            try:
                with open(os.path.join(STATE_DIR, "active-playlist"), "w") as f:
                    f.write(new_name)
            except OSError:
                pass
            self._rebuild_stream()
            self._schedule_next_advance()

        for act_name, cb, ptype in (
            ("stream",          on_stream,         GLib.VariantType.new("s")),
            ("preset",          on_preset,         GLib.VariantType.new("s")),
            ("playlist",        on_playlist,       GLib.VariantType.new("s")),
            ("copy_segment",    on_copy_segment,   None),
            ("open_segment",    on_open_segment,   None),
            ("dismiss_stream",  on_dismiss_stream, None),
            ("pause",    on_pause,    None),
            ("pin",      on_pin,      None),
            ("unpin",    on_unpin,    None),
        ):
            act = Gio.SimpleAction.new(act_name, ptype)
            act.connect("activate", cb)
            group.add_action(act)
        self.insert_action_group("ticker", group)
        self._actions_registered = True

    def _on_tooltip(self, widget, x, y, keyboard_mode, tooltip):
        stream = (self.pinned_stream if self.pinned_stream
                  else self.stream_order[self.stream_idx])
        seg_info = ""
        if self.segments:
            idx = self._segment_at_x(x)
            if idx is not None and idx < len(self.segments):
                seg_text = self.segments[idx]
                label = "URL" if URL_RE.match(seg_text) else "Segment"
                action = "click to open" if URL_RE.match(seg_text) else "click to copy"
                seg_info = f'\n<b>{label}:</b> {escape(seg_text)}\n<i>{action}</i>'
        status = []
        if self.paused:
            status.append("<b>PAUSED</b>")
        if self.pinned_stream:
            status.append(f"<b>PINNED:</b> {escape(self.pinned_stream)}")
        if self.urgent_mode:
            status.append("<b>URGENT</b>")
        status_line = " \u00b7 ".join(status) + "\n" if status else ""
        tooltip.set_markup(
            f'{status_line}'
            f'<b>Playlist:</b> {escape(self.playlist_name)}\n'
            f'<b>Stream:</b> {escape(stream)}\n'
            f'<b>Preset:</b> {escape(self._base_preset_name)}\n'
            f'<b>Speed:</b> {self.preset.speed:.0f} px/s\n'
            f'<b>Controls:</b> scroll=speed, shift+scroll=switch stream,\n'
            f'mid-click=pause, right-click=menu'
            f'{seg_info}'
        )
        return True

    # ── Tick / scroll ─────────────────────────

    def _on_da_realize(self, widget):
        clock = widget.get_frame_clock()
        if clock is not None:
            clock.begin_updating()

    def _tick(self, widget, frame_clock):
        now = frame_clock.get_frame_time()
        p = self.preset
        if self.last_us is not None:
            dt = min((now - self.last_us) / 1_000_000.0, 0.05)
            if not self.paused:
                self.offset += p.speed * dt
                self.gradient_phase += p.gradient_speed * dt
            self.time_s += dt
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
            if self.gradient_phase >= p.gradient_span:
                self.gradient_phase -= p.gradient_span
            # Typewriter reveal: advance
            if self.reveal_chars < 99999:
                self.reveal_chars += 3
                if self.reveal_chars >= 600:
                    self.reveal_chars = 99999  # disable until next rebuild

            # Glitch
            glitch_prob = p.glitch_prob
            glitch_frames = p.glitch_frames
            ca_offset = p.ca_offset
            if self.urgent_mode:
                glitch_prob = max(glitch_prob, 0.05)
                glitch_frames = max(glitch_frames, 6)
                ca_offset = max(ca_offset, 6)
            if self.glitch_remaining > 0:
                self.glitch_remaining -= 1
            elif glitch_prob > 0 and random.random() < glitch_prob:
                self.glitch_remaining = glitch_frames
                self.glitch_strips = [
                    (random.randint(0, BAR_H - 4), random.randint(3, 8),
                     random.randint(-10, 10))
                    for _ in range(3)
                ]
            # Stash for _draw
            self._runtime_ca_offset = ca_offset
        self.last_us = now
        self.water_frame += 1
        # Throttle redraws to ~60Hz. Offset/time/glitch state still advances
        # via dt every tick (animation stays smooth); only pixel rasterization
        # is capped. Saves ~75% CPU on 240Hz displays with vfr=false.
        if now - getattr(self, "_last_draw_us", 0) >= 16_000:
            widget.queue_draw()
            self._last_draw_us = now
        return GLib.SOURCE_CONTINUE

    # ── Rendering helpers ─────────────────────

    def _compute_segment_bounds(self):
        """Populate self.segment_bounds by finding each U+00B7 separator.

        Pango uses UTF-8 byte offsets; Python str uses code points. We convert
        via len(prefix.encode('utf-8')) to bridge the two.
        """
        self.segment_bounds = []
        if not self.segments or self.layout is None:
            return
        text = self.layout.get_text()
        if not text:
            return
        # Markup is duplicated for seamless scroll — only scan first half
        half_len = len(text) // 2
        cursor_char = 0
        bounds = []
        for seg in self.segments:
            sep_char = text.find("\u00b7", cursor_char, half_len)
            if sep_char == -1:
                break
            start_byte = len(text[:cursor_char].encode("utf-8"))
            end_byte = len(text[:sep_char + 1].encode("utf-8"))
            try:
                rect_start = self.layout.index_to_pos(start_byte)
                rect_end = self.layout.index_to_pos(end_byte)
                xs = rect_start.x / Pango.SCALE
                xe = rect_end.x / Pango.SCALE
            except Exception:
                break
            if xe < xs:
                xe = xs + 20
            bounds.append((xs, xe, seg))
            cursor_char = sep_char + 1
        self.segment_bounds = bounds

    def _render_text_surface(self, width, height):
        surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        tc = _cairo.Context(surf)
        x0 = -self.offset
        y = (height - 14) / 2
        p = self.preset

        # _dup already duplicates the markup once for seamless wrap. Draw
        # exactly 2 copies so short content scrolls as a single strip with
        # trailing empty space, rather than tiling 9+ times across ultrawides.
        step = max(self.half_w, 1.0)
        copies = 2
        xs = [x0 + i * step for i in range(copies)]

        # Dark stroke outline OR color-cycling pulse outline
        ow = getattr(p, "outline_width", 0.8)
        if ow > 0:
            if getattr(p, "outline_pulse", False):
                # Outline uses same gradient as text (lit-edge effect)
                stroke_src = make_gradient(x0, self.half_w, self.gradient_phase)
                tc.set_source(stroke_src)
            else:
                tc.set_source_rgba(0.02, 0.03, 0.05, 0.6)
            tc.set_line_width(ow)
            tc.set_line_join(_cairo.LINE_JOIN_ROUND)
            for xi in xs:
                tc.move_to(xi, y)
                PangoCairo.update_layout(tc, self.layout)
                PangoCairo.layout_path(tc, self.layout)
            tc.stroke()

        # Gradient-filled text on top
        text_grad = make_gradient(x0, self.half_w, self.gradient_phase)
        tc.set_source(text_grad)
        for xi in xs:
            tc.move_to(xi, y)
            PangoCairo.update_layout(tc, self.layout)
            PangoCairo.show_layout(tc, self.layout)
        surf.flush()
        return surf

    def _apply_typewriter_clip(self, surf, width, height):
        """If reveal is active, clip to only the first N glyphs worth of pixels.

        Approximation: each glyph ~8px wide at 11pt Maple Mono. Use offset +
        reveal_chars*8 as the right edge.
        """
        if self.reveal_chars >= 99999:
            return surf
        glyph_w = 8
        reveal_px = int(self.reveal_chars * glyph_w - self.offset)
        if reveal_px >= width:
            return surf
        # Black out pixels past the reveal edge
        stride = surf.get_stride()
        data = np.frombuffer(surf.get_data(), dtype=np.uint8).reshape(height, stride)
        if reveal_px > 0:
            data[:, reveal_px * 4:] = 0
        else:
            data[:, :] = 0
        surf.mark_dirty()
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
        strip_w = 8
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
        gb[:] = np.hstack([blurred, np.zeros((height, gs - width), dtype=np.uint8)]).tobytes()
        glow.mark_dirty()
        return glow

    # ── Draw ──────────────────────────────────

    def _draw(self, da, cr, width, height):
        p = self.preset

        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            self.layout.set_markup(self.ticker_markup, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = max(logical.width / 2.0, 1.0)
            self.offset = self.offset % self.half_w
            self._compute_segment_bounds()

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

        # Holo shimmer (Phase 6)
        shimmer_a = getattr(p, "holo_shimmer", 0.0)
        if shimmer_a > 0:
            for i in range(height):
                v = (math.sin(i * 4.2 + self.time_s * 3) *
                     math.sin(i * 0.7 - self.time_s * 1.5))
                v = (v + 1) / 2  # 0..1
                cr.set_source_rgba(0.161, 0.941, 1.000, shimmer_a * v)
                cr.rectangle(0, i, width, 1)
                cr.fill()

        # Top border (synthwave sweep or static gradient)
        if getattr(p, "synthwave_border", True):
            # Wider sweeping band with hot-pink to purple accent
            phase = self.time_s * 150
            grad = _cairo.LinearGradient(0, 0, width, 0)
            grad.set_extend(_cairo.Extend.REPEAT)
            # Offset colors by phase for sweeping feel
            sweep = (phase % 800) / 800.0
            grad.add_color_stop_rgb((0.0 + sweep) % 1.0, 1.000, 0.278, 0.820)  # magenta
            grad.add_color_stop_rgb((0.33 + sweep) % 1.0, 1.000, 0.361, 0.541)  # pink
            grad.add_color_stop_rgb((0.66 + sweep) % 1.0, 0.290, 0.659, 1.000)  # blue
            grad.add_color_stop_rgb((1.0 + sweep) % 1.0, 1.000, 0.278, 0.820)
            cr.set_source(grad)
        else:
            cr.set_source(make_gradient(0, width, self.gradient_phase))
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # ═══ FOREGROUND ═══════════════════════════
        text_surf = self._render_text_surface(width, height)
        text_surf = self._apply_typewriter_clip(text_surf, width, height)
        text_surf = self._apply_wave(text_surf, width, height)

        # Phosphor decay trail — composite old surfaces at diminishing alpha
        phosphor_a = getattr(p, "phosphor_trail", 0.0)
        if phosphor_a > 0:
            # Append current surface to ring buffer every 2 frames
            self.phosphor_frame_ctr += 1
            if self.phosphor_frame_ctr % 2 == 0:
                self.phosphor_ring.append(text_surf)
            for i, old in enumerate(list(self.phosphor_ring)[:-1]):
                age = len(self.phosphor_ring) - i
                alpha = phosphor_a / age
                cr.save()
                cr.set_source_rgba(0.161, 0.941, 1.000, alpha)
                cr.mask_surface(old, 0, 0)
                cr.restore()

        # Ghost echo (VHS double-image)
        echo_a = getattr(p, "ghost_echo", 0.0)
        if echo_a > 0:
            cr.save()
            cr.set_source_rgba(0.161, 0.941, 1.000, echo_a)
            cr.mask_surface(text_surf, -3 * getattr(p, "shadow_offset", 2), 0)
            cr.restore()

        # Glow (cached — recompute every 8 frames)
        self.glow_frame += 1
        if self.glow_cache is None or self.glow_frame % 8 == 0:
            self.glow_cache = self._compute_glow(text_surf, width, height)
        glow_a8 = self.glow_cache
        if glow_a8 is not None:
            glow_alpha = p.glow_base_alpha + p.glow_pulse_amp * math.sin(
                self.time_s * (2 * math.pi / max(p.glow_pulse_period, 0.1)))
            glow_alpha = max(0.0, min(1.0, glow_alpha))
            cr.save()
            cr.push_group()
            cr.set_source(make_gradient(0, width, self.gradient_phase))
            cr.mask_surface(glow_a8, 0, 0)
            pattern = cr.pop_group()
            cr.set_source(pattern)
            cr.paint_with_alpha(glow_alpha)
            cr.restore()

        # Shadow
        if p.shadow_alpha > 0:
            cr.set_source_rgba(0, 0, 0, p.shadow_alpha)
            cr.mask_surface(text_surf, p.shadow_offset, p.shadow_offset)

        # Sharp text
        cr.set_source_surface(text_surf, 0, 0)
        cr.paint()

        # Glitch + CA
        ca = getattr(self, "_runtime_ca_offset", p.ca_offset)
        if self.glitch_remaining > 0 and ca > 0:
            cr.save()
            cr.set_source_rgba(1, 0, 0, 0.3)
            cr.mask_surface(text_surf, ca, 0)
            cr.restore()
            cr.save()
            cr.set_source_rgba(0, 0.3, 1, 0.3)
            cr.mask_surface(text_surf, -ca, 0)
            cr.restore()
            for sy, sh, sx in self.glitch_strips:
                cr.save()
                cr.rectangle(0, sy, width, sh)
                cr.clip()
                cr.set_source_surface(text_surf, sx, 0)
                cr.paint()
                cr.restore()

        # ═══ POST-PASS ══════════════════════════
        # Edge fade vignette (FG mask)
        fade_px = getattr(p, "edge_fade", 0)
        if fade_px > 0 and width > 2 * fade_px:
            # Left fade
            lgrad = _cairo.LinearGradient(0, 0, fade_px, 0)
            lgrad.add_color_stop_rgba(0, 0.020, 0.027, 0.051, 1.0)
            lgrad.add_color_stop_rgba(1, 0.020, 0.027, 0.051, 0.0)
            cr.set_source(lgrad)
            cr.rectangle(0, 0, fade_px, height)
            cr.fill()
            # Right fade
            rgrad = _cairo.LinearGradient(width - fade_px, 0, width, 0)
            rgrad.add_color_stop_rgba(0, 0.020, 0.027, 0.051, 0.0)
            rgrad.add_color_stop_rgba(1, 0.020, 0.027, 0.051, 1.0)
            cr.set_source(rgrad)
            cr.rectangle(width - fade_px, 0, fade_px, height)
            cr.fill()

        # Progress indicator — 2px bar at bottom, filling as dwell elapses
        if getattr(p, "progress_bar", False):
            current = (self.pinned_stream if self.pinned_stream
                       else self.stream_order[self.stream_idx])
            interval = self._current_interval(current)
            elapsed = self.time_s - self.stream_start_s
            pct = min(elapsed / max(interval, 1), 1.0)
            prog_w = int(width * pct)
            # Dim rail spanning full width so the bar is always anchored
            cr.set_source_rgba(0.35, 0.40, 0.60, 0.22)
            cr.rectangle(0, height - 2, width, 2)
            cr.fill()
            if prog_w > 0:
                prog_grad = make_gradient(0, max(prog_w, 1), self.gradient_phase)
                cr.set_source(prog_grad)
                cr.rectangle(0, height - 2, prog_w, 2)
                cr.fill()

        # Stream-change wipe — brief gradient sweep the first 400ms after a
        # stream rotation, signalling new content has arrived.
        wipe_dur = 0.4
        wipe_age = self.time_s - self.stream_start_s
        if wipe_age < wipe_dur:
            progress = wipe_age / wipe_dur
            band_w = max(int(width * 0.18), 80)
            lead = int((width + band_w) * progress) - band_w
            alpha = 0.55 * (1.0 - progress)
            sweep = _cairo.LinearGradient(lead, 0, lead + band_w, 0)
            sweep.add_color_stop_rgba(0.0, 0.161, 0.941, 1.000, 0.0)
            sweep.add_color_stop_rgba(0.5, 1.000, 0.278, 0.820, alpha)
            sweep.add_color_stop_rgba(1.0, 0.161, 0.941, 1.000, 0.0)
            cr.set_source(sweep)
            cr.rectangle(0, 0, width, height)
            cr.fill()

        # Paused overlay
        if self.paused:
            cr.set_source_rgba(1, 0.278, 0.820, 0.08)
            cr.paint()


class TickerApp(Gtk.Application):
    def __init__(self):
        super().__init__(
            application_id="io.hairglasses.keybind_ticker",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )

    def do_activate(self):
        win = TickerWindow(preset_name=START_PRESET, application=self)
        win.present()
        self._window = win

        # SIGUSR1 → reload state files inline (no restart, no flicker).
        # signal.signal handlers fire in the signal thread; marshal the work
        # onto the GTK main loop via GLib.idle_add so widget calls are safe.
        def _on_sigusr1(signum, frame):
            GLib.idle_add(win.reload_from_state)
        try:
            signal.signal(signal.SIGUSR1, _on_sigusr1)
        except (ValueError, OSError):
            # Not in main thread or platform mismatch — ignore silently;
            # `hg ticker` will fall back to `systemctl --user restart`.
            pass


if __name__ == "__main__":
    # Hide our custom CLI flags from Gtk.Application
    TickerApp().run([sys.argv[0]])
