#!/usr/bin/env bash
set -euo pipefail

if ! command -v pacman >/dev/null 2>&1; then
    printf 'This script currently supports pacman-based systems only.\n' >&2
    exit 2
fi

if ! sudo -n true >/dev/null 2>&1; then
    printf 'Passwordless sudo is required for retroarch-install-widescreen-cores.sh.\n' >&2
    exit 3
fi

packages=(
    libretro-flycast
    libretro-dolphin
    libretro-genesis-plus-gx
    libretro-bsnes-hd
    libretro-mupen64plus-next
    libretro-kronos
)

sudo -n pacman -S --noconfirm --needed "${packages[@]}"

find /usr/lib/libretro -maxdepth 1 -type f \
    | rg '(flycast|dolphin|genesis_plus_gx|bsnes_hd|mupen64plus|kronos)' \
    | sort
