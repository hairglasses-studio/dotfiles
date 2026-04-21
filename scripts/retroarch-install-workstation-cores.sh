#!/usr/bin/env bash
set -euo pipefail

skip_aur=0
while [[ $# -gt 0 ]]; do
    case "$1" in
        --skip-aur) skip_aur=1; shift ;;
        -h|--help)
            cat <<'EOF'
Usage: retroarch-install-workstation-cores.sh [--skip-aur]

Installs libretro cores packaged in the official Arch repos via pacman,
then installs the AUR cores that have an AUR package (currently beetle-vb).

Options:
  --skip-aur   Skip the AUR pass (useful for CI / non-interactive shells
               that do not have yay/paru configured).
EOF
            exit 0 ;;
        *)
            printf 'unknown argument: %s\n' "$1" >&2
            exit 2 ;;
    esac
done

if ! command -v pacman >/dev/null 2>&1; then
    printf 'This script currently supports pacman-based systems only.\n' >&2
    exit 2
fi

if ! sudo -n true >/dev/null 2>&1; then
    printf 'Passwordless sudo is required for retroarch-install-workstation-cores.sh.\n' >&2
    exit 3
fi

pacman_packages=(
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

aur_packages=(
    libretro-beetle-vb-git
)

available_packages=()
missing_packages=()

for package in "${pacman_packages[@]}"; do
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

aur_helper=""
if [[ $skip_aur -eq 0 ]]; then
    if command -v yay >/dev/null 2>&1; then
        aur_helper="yay"
    elif command -v paru >/dev/null 2>&1; then
        aur_helper="paru"
    fi
fi

aur_installed=()
aur_skipped=()
if [[ ${#aur_packages[@]} -gt 0 ]]; then
    if [[ -z "$aur_helper" ]]; then
        aur_skipped=("${aur_packages[@]}")
        if [[ $skip_aur -eq 1 ]]; then
            printf 'aur_pass=skipped (--skip-aur)\n'
        else
            printf 'aur_pass=skipped (no yay/paru on PATH)\n'
        fi
        printf 'aur_pending=%s\n' "${aur_skipped[*]}"
    else
        for package in "${aur_packages[@]}"; do
            if "$aur_helper" -S --noconfirm --needed "$package"; then
                aur_installed+=("$package")
            else
                aur_skipped+=("$package")
            fi
        done
        printf 'aur_installed=%d aur_failed=%d helper=%s\n' \
            "${#aur_installed[@]}" "${#aur_skipped[@]}" "$aur_helper"
        if [[ ${#aur_skipped[@]} -gt 0 ]]; then
            printf 'aur_failed_packages=%s\n' "${aur_skipped[*]}"
        fi
    fi
fi

find /usr/lib/libretro -maxdepth 1 -type f \
    | rg '(sameboy|mgba|mesen|pce_fast|desmume|ppsspp|flycast|dolphin|genesis_plus_gx|bsnes_hd|mupen64plus|kronos|mednafen_vb|beetle_vb|mednafen_wswan|beetle_wswan|race_libretro)' \
    | sort
