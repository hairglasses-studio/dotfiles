#!/usr/bin/env bash
set -euo pipefail

if ! command -v pacman >/dev/null 2>&1; then
    printf 'This script currently supports pacman-based systems only.\n' >&2
    exit 2
fi

if ! sudo -n true >/dev/null 2>&1; then
    printf 'Passwordless sudo is required for retroarch-install-workstation-cores.sh.\n' >&2
    exit 3
fi

wanted_packages=(
    libretro-sameboy
    libretro-mgba
    libretro-mesen
    libretro-beetle-pce-fast
    libretro-desmume
    libretro-ppsspp
    libretro-flycast
    libretro-dolphin
    libretro-genesis-plus-gx
    libretro-bsnes-hd
    libretro-mupen64plus-next
    libretro-kronos
)

available_packages=()
missing_packages=()

for package in "${wanted_packages[@]}"; do
    if pacman -Si "$package" >/dev/null 2>&1; then
        available_packages+=("$package")
    else
        missing_packages+=("$package")
    fi
done

if [[ ${#available_packages[@]} -gt 0 ]]; then
    sudo -n pacman -S --noconfirm --needed "${available_packages[@]}"
fi

printf 'installed_or_present=%d unavailable=%d\n' "${#available_packages[@]}" "${#missing_packages[@]}"
if [[ ${#missing_packages[@]} -gt 0 ]]; then
    printf 'unavailable_packages=%s\n' "${missing_packages[*]}"
fi

find /usr/lib/libretro -maxdepth 1 -type f \
    | rg '(sameboy|mgba|mesen|pce_fast|desmume|ppsspp|flycast|dolphin|genesis_plus_gx|bsnes_hd|mupen64plus|kronos)' \
    | sort
