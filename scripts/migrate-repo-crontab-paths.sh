#!/usr/bin/env bash
set -euo pipefail

# migrate-repo-crontab-paths.sh — Rewrite legacy repo paths in the user's
# crontab after a repo rename.
#
# The chromecast4k repo was renamed to hg-android on 2026-04-16. Cron lines
# that still reference $HG_STUDIO_ROOT/chromecast4k/scripts/... point at a
# directory that no longer exists, so every scheduled job silently fails.
# This helper detects those lines and rewrites them in place.
#
# Usage:
#   migrate-repo-crontab-paths.sh --check   # exit 1 if legacy paths found
#   migrate-repo-crontab-paths.sh --apply   # rewrite and re-install crontab
#
# Overridable via env for tests / dry runs:
#   CRONTAB_BIN   binary to use (default: crontab)
#   HG_STUDIO_ROOT  workspace root (default: $HOME/hairglasses-studio)

usage() {
    cat <<'EOF'
Usage: migrate-repo-crontab-paths.sh --check|--apply

Migrate legacy `chromecast4k/scripts/` cron entries to `hg-android/scripts/`.

Options:
  --check    Exit 1 if any legacy path is still in the crontab.
  --apply    Rewrite legacy paths in place and re-install via crontab.
  -h --help  Show this help text.
EOF
}

mode=""
while [[ $# -gt 0 ]]; do
    case "$1" in
        --check) mode="check"; shift ;;
        --apply) mode="apply"; shift ;;
        -h|--help) usage; exit 0 ;;
        *) usage >&2; exit 2 ;;
    esac
done

[[ -n "$mode" ]] || { usage >&2; exit 2; }

CRONTAB_BIN="${CRONTAB_BIN:-crontab}"
HG_STUDIO_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"

legacy_prefix="${HG_STUDIO_ROOT}/chromecast4k/scripts/"
replacement_prefix="${HG_STUDIO_ROOT}/hg-android/scripts/"

current="$("$CRONTAB_BIN" -l 2>/dev/null || true)"

# Escape the prefix for safe sed substitution (paths contain / and possibly
# other sed metacharacters). Using | as the delimiter avoids collisions with
# path separators.
sed_legacy="$(printf '%s' "$legacy_prefix" | sed -e 's/[\\|&]/\\&/g')"
sed_replacement="$(printf '%s' "$replacement_prefix" | sed -e 's/[\\|&]/\\&/g')"

rewritten="$(printf '%s' "$current" | sed "s|${sed_legacy}|${sed_replacement}|g")"

if [[ "$mode" == "check" ]]; then
    if printf '%s' "$current" | grep -F -q "$legacy_prefix"; then
        printf 'Legacy repo paths detected in %s crontab:\n' "$USER"
        printf '%s\n' "$current" | grep -F "$legacy_prefix" | while IFS= read -r line; do
            printf '  %s\n' "$line"
            # Also show what the rewrite would produce.
            rewritten_line="$(printf '%s' "$line" | sed "s|${sed_legacy}|${sed_replacement}|g")"
            printf '  → %s\n' "$rewritten_line"
        done
        exit 1
    fi
    printf 'No legacy repo paths found in %s crontab.\n' "$USER"
    exit 0
fi

# apply mode
if ! printf '%s' "$current" | grep -F -q "$legacy_prefix"; then
    printf 'No legacy repo paths found in %s crontab; nothing to do.\n' "$USER"
    exit 0
fi

tmp_file="$(mktemp --suffix=.crontab)"
trap 'rm -f "$tmp_file"' EXIT
printf '%s\n' "$rewritten" > "$tmp_file"
"$CRONTAB_BIN" "$tmp_file"
printf 'Updated %s crontab — rewrote %s → %s\n' \
    "$USER" "$legacy_prefix" "$replacement_prefix"
