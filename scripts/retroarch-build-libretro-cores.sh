#!/usr/bin/env bash
set -euo pipefail

dry_run=0
only=""
install_dir="${LIBRETRO_INSTALL_DIR:-/usr/lib/libretro}"

usage() {
    cat <<'EOF'
Usage: retroarch-build-libretro-cores.sh [--dry-run] [--only race|beetle-wswan|ngp|wonderswan|vb]

Builds the libretro cores that are not packaged in pacman or AUR on Arch.
Requires git, make, a C/C++ toolchain, and sudo for the final install step.

Cores built by this script:
  race           -> race_libretro.so            (system: ngp)
  beetle-wswan   -> mednafen_wswan_libretro.so  (system: wonderswan)
  beetle-vb      -> mednafen_vb_libretro.so     (system: vb, fallback when AUR package is unavailable)

Default: builds race and beetle-wswan. Pass --only to restrict.
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) dry_run=1; shift ;;
        --only)
            [[ $# -lt 2 ]] && { usage; exit 2; }
            case "$2" in
                race|ngp) only="race" ;;
                beetle-wswan|wonderswan) only="beetle-wswan" ;;
                beetle-vb|vb) only="beetle-vb" ;;
                *) printf 'unknown core: %s\n' "$2" >&2; exit 2 ;;
            esac
            shift 2 ;;
        --install-dir)
            [[ $# -lt 2 ]] && { usage; exit 2; }
            install_dir="$2"; shift 2 ;;
        -h|--help) usage; exit 0 ;;
        *) usage; exit 2 ;;
    esac
done

if ! command -v git >/dev/null 2>&1; then
    printf 'git is required for retroarch-build-libretro-cores.sh\n' >&2
    exit 3
fi
if ! command -v make >/dev/null 2>&1; then
    printf 'make is required for retroarch-build-libretro-cores.sh\n' >&2
    exit 3
fi

declare -A CORE_REPO=(
    [race]="https://github.com/libretro/race.git"
    [beetle-wswan]="https://github.com/libretro/beetle-wswan-libretro.git"
    [beetle-vb]="https://github.com/libretro/beetle-vb-libretro.git"
)
declare -A CORE_MAKEFILE=(
    [race]="Makefile.libretro"
    [beetle-wswan]="Makefile"
    [beetle-vb]="Makefile"
)
declare -A CORE_ARTIFACT=(
    [race]="race_libretro.so"
    [beetle-wswan]="mednafen_wswan_libretro.so"
    [beetle-vb]="mednafen_vb_libretro.so"
)

if [[ -n "$only" ]]; then
    targets=("$only")
else
    targets=(race beetle-wswan)
fi

build_root=""
cleanup() {
    if [[ -n "$build_root" && -d "$build_root" ]]; then
        rm -rf "$build_root"
    fi
}
if [[ $dry_run -eq 0 ]]; then
    build_root="$(mktemp -d -t libretro-build.XXXXXX)"
    trap cleanup EXIT
fi

for core in "${targets[@]}"; do
    repo="${CORE_REPO[$core]}"
    makefile="${CORE_MAKEFILE[$core]}"
    artifact="${CORE_ARTIFACT[$core]}"
    target_path="${install_dir}/${artifact}"

    if [[ $dry_run -eq 1 ]]; then
        printf 'DRY-RUN %s\n' "$core"
        printf '  clone   %s\n' "$repo"
        printf '  build   make -f %s\n' "$makefile"
        printf '  install %s -> %s\n' "$artifact" "$target_path"
        continue
    fi

    core_dir="${build_root}/${core}"
    printf '[%s] cloning %s\n' "$core" "$repo"
    git clone --depth=1 "$repo" "$core_dir" >/dev/null

    printf '[%s] building via %s\n' "$core" "$makefile"
    make -C "$core_dir" -f "$makefile" -j"$(nproc)"

    if [[ ! -f "${core_dir}/${artifact}" ]]; then
        printf '[%s] expected artifact not produced: %s\n' "$core" "${core_dir}/${artifact}" >&2
        exit 4
    fi

    printf '[%s] installing to %s\n' "$core" "$target_path"
    sudo install -D -m 0644 "${core_dir}/${artifact}" "$target_path"
done

printf 'done\n'
