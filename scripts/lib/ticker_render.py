"""ticker_render — shared rendering helpers for the dotfiles ticker family.

Used by scripts/keybind-ticker.py, scripts/toast-ticker.py,
scripts/rsvp-ticker.py, scripts/lyrics-ticker.py, scripts/subtitle-ticker.py.

To import from a sibling script (since the scripts dir is not a package):

    import sys, os
    sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
    import ticker_render as tr

This module assumes GTK4, Pango, PangoCairo, and Gtk4LayerShell are already
available via gi.require_version() in the caller; the helpers accept
concrete widgets / drawing contexts and do not import gi themselves at
module load time (so `import ticker_render` works in headless contexts like
scripts/ticker-headless.py without pulling in GTK).
"""
from __future__ import annotations

from html import escape as _html_escape

# ── Constants ────────────────────────────────────────────────────────────
BAR_H = 28
DEFAULT_BG_ALPHA = 0.88
HAIRGLASSES_NEON = {
    "primary":   "#29f0ff",
    "secondary": "#ff47d1",
    "tertiary":  "#3dffb5",
    "warning":   "#ffe45e",
    "danger":    "#ff5c8a",
    "blue":      "#4aa8ff",
    "muted":     "#66708f",
    "fg":        "#f7fbff",
    "bg":        "#05070d",
    "surface":   "#0f1219",
}


# ── Ticker badge / empty / dup helpers ───────────────────────────────────
# These are the three core Pango-markup primitives that keybind-ticker.py
# originally defined internally. Moved here so the companion apps (toast,
# lyrics, subtitle, rsvp) can draw consistent badges without re-implementing
# them.

def badge(text: str, bg_hex: str, fg_hex: str = HAIRGLASSES_NEON["bg"]) -> str:
    return (
        f'<span background="{bg_hex}" foreground="{fg_hex}" '
        f'font_desc="Maple Mono NF CN Bold 10"> {_html_escape(text)} </span>  '
    )


def empty(badge_label: str, bg_hex: str, msg: str) -> tuple[str, list[str]]:
    """Return (markup, segments). The ``__EMPTY__`` sentinel in segments lets
    the main ticker loop tag the stream as errored for backoff scheduling."""
    markup = (
        badge(badge_label, bg_hex)
        + f'<span font_desc="Maple Mono NF CN 11">  {_html_escape(msg)}  \u00b7</span>'
    )
    return markup, ["__EMPTY__"]


def dup(markup: str) -> str:
    """Duplicate markup so the render loop can tile it across wide displays
    without a visible seam between copies."""
    return markup + markup


# ── Colour conversion ────────────────────────────────────────────────────

def hex_to_rgb(h: str) -> tuple[float, float, float]:
    """Accept #RGB, #RRGGBB, or bare hex. Return (r, g, b) floats in [0, 1]."""
    h = h.lstrip("#")
    if len(h) == 3:
        h = "".join(c * 2 for c in h)
    if len(h) != 6:
        return (0.24, 1.0, 0.71)  # Hairglasses tertiary as fallback
    return (
        int(h[0:2], 16) / 255.0,
        int(h[2:4], 16) / 255.0,
        int(h[4:6], 16) / 255.0,
    )


# ── GTK4 / PangoCairo helpers (imported lazily) ──────────────────────────

def setup_layer_shell(
    window,
    edges: tuple,
    namespace: str,
    monitor_name: str | None = None,
    *,
    layer: str = "BOTTOM",
    exclusive_zone: int | None = None,
    margins: dict | None = None,
):
    """Initialise gtk4-layer-shell for a window. ``edges`` is a tuple of
    Gtk4LayerShell.Edge values; ``layer`` is the string name of a
    Gtk4LayerShell.Layer value ("BOTTOM", "TOP", "OVERLAY", "BACKGROUND")."""
    import gi
    gi.require_version("Gtk4LayerShell", "1.0")
    gi.require_version("Gdk", "4.0")
    from gi.repository import Gtk4LayerShell, Gdk  # noqa: E402

    Gtk4LayerShell.init_for_window(window)
    layer_value = getattr(Gtk4LayerShell.Layer, layer.upper())
    Gtk4LayerShell.set_layer(window, layer_value)
    for edge in edges:
        Gtk4LayerShell.set_anchor(window, edge, True)
    Gtk4LayerShell.set_namespace(window, namespace)
    if exclusive_zone is not None:
        Gtk4LayerShell.set_exclusive_zone(window, exclusive_zone)
    if margins:
        for edge, px in margins.items():
            Gtk4LayerShell.set_margin(window, edge, px)

    mon = find_monitor(Gdk.Display.get_default(), monitor_name) if monitor_name else None
    if mon is not None:
        Gtk4LayerShell.set_monitor(window, mon)
    return mon


def find_monitor(display, name: str):
    """Return the Gdk.Monitor whose connector name contains ``name``, or None."""
    if display is None or not name:
        return None
    monitors = display.get_monitors()
    for i in range(monitors.get_n_items()):
        mon = monitors.get_item(i)
        if mon and name in (mon.get_connector() or ""):
            return mon
    return None


def make_drawing_area(height: int, draw_func, *, tick: bool = True):
    """Create a Gtk.DrawingArea wired up the way every ticker script does it.

    ``draw_func`` is (widget, cairo_context, width, height) → None.
    When ``tick`` is True, we start the frame clock so Hyprland doesn't
    suspend the layer-shell surface while idle.
    """
    import gi
    gi.require_version("Gtk", "4.0")
    from gi.repository import Gtk  # noqa: E402

    da = Gtk.DrawingArea()
    da.set_content_height(height)
    da.set_vexpand(True)
    da.set_hexpand(True)
    da.set_draw_func(draw_func)
    if tick:
        da.connect(
            "realize",
            lambda w: w.get_frame_clock().begin_updating(),
        )
    return da


def fill_bg(cr, w: float, h: float, *, alpha: float = DEFAULT_BG_ALPHA,
            rgb: tuple[float, float, float] | None = None) -> None:
    """Draw the standard dark panel background. Overridable colour lets the
    toast accent tone paint a slightly different backdrop."""
    r, g, b = rgb if rgb is not None else hex_to_rgb(HAIRGLASSES_NEON["bg"])
    cr.set_source_rgba(r, g, b, alpha)
    cr.rectangle(0, 0, w, h)
    cr.fill()


def draw_text(cr, markup: str, font_desc: str, x: float, y: float, *,
              color_rgba: tuple[float, float, float, float] | None = None):
    """Render Pango markup at (x, y). If ``color_rgba`` is given, paint over
    layout with that source; otherwise let the markup itself dictate colour."""
    import gi
    gi.require_version("Pango", "1.0")
    gi.require_version("PangoCairo", "1.0")
    from gi.repository import Pango, PangoCairo  # noqa: E402

    layout = PangoCairo.create_layout(cr)
    layout.set_font_description(Pango.FontDescription(font_desc))
    layout.set_markup(markup, -1)
    if color_rgba is not None:
        cr.set_source_rgba(*color_rgba)
    cr.move_to(x, y)
    PangoCairo.show_layout(cr, layout)
    return layout


def measure_text(cr, markup: str, font_desc: str) -> tuple[int, int]:
    """Return (width_px, height_px) of the rendered markup."""
    import gi
    gi.require_version("Pango", "1.0")
    gi.require_version("PangoCairo", "1.0")
    from gi.repository import Pango, PangoCairo  # noqa: E402

    layout = PangoCairo.create_layout(cr)
    layout.set_font_description(Pango.FontDescription(font_desc))
    layout.set_markup(markup, -1)
    return layout.get_pixel_size()
