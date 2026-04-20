#!/usr/bin/env bash
set -euo pipefail

target_root="${RETROARCH_SYSTEM_DIR:-${HOME}/.config/retroarch/system}"
target_dir="${target_root}/dolphin-emu/Sys"
repo_url="${DOLPHIN_LIBRETRO_SYS_REPO:-https://github.com/libretro/dolphin.git}"
branch="${DOLPHIN_LIBRETRO_SYS_BRANCH:-master}"

tmpdir="$(mktemp -d)"
cleanup() {
    rm -rf "${tmpdir}"
}
trap cleanup EXIT

git -C "${tmpdir}" init -q
git -C "${tmpdir}" remote add origin "${repo_url}"
git -C "${tmpdir}" config core.sparseCheckout true
printf 'Data/Sys\n' > "${tmpdir}/.git/info/sparse-checkout"
git -C "${tmpdir}" pull --depth=1 origin "${branch}" >/dev/null

mkdir -p "$(dirname "${target_dir}")"
rm -rf "${target_dir}"
cp -a "${tmpdir}/Data/Sys" "${target_dir}"

printf 'Synced %s\n' "${target_dir}"
