#!/usr/bin/env bash
# kitty-config.sh — Shared kitty config query functions
# Source this from scripts that need to read kitty config state.

_KITTY_CONFIG="${KITTY_CONFIG:-$HOME/.config/kitty/kitty.conf}"

kitty_get_font() {
    grep -m1 '^font_family ' "$_KITTY_CONFIG" 2>/dev/null | sed 's/^font_family  *//'
}

kitty_get_font_size() {
    grep -m1 '^font_size ' "$_KITTY_CONFIG" 2>/dev/null | awk '{print $2}'
}

kitty_get_opacity() {
    grep -m1 '^background_opacity ' "$_KITTY_CONFIG" 2>/dev/null | awk '{print $2}'
}

kitty_get_tab_style() {
    grep -m1 '^tab_bar_style ' "$_KITTY_CONFIG" 2>/dev/null | awk '{print $2}'
}

kitty_reload() {
    kill -USR1 "$(pidof kitty)" 2>/dev/null || true
}
