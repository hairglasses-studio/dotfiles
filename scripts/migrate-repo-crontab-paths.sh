#!/usr/bin/env bash
# migrate-repo-crontab-paths.sh — migrate stale repo paths inside a user's crontab
set -euo pipefail

CRONTAB_BIN="${CRONTAB_BIN:-crontab}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STUDIO_ROOT="${HG_STUDIO_ROOT:-$(cd "$SCRIPT_DIR/../.." && pwd)}"

MODE="check"
TARGET_USER=""

usage() {
  cat <<'EOF'
Usage: migrate-repo-crontab-paths.sh [--check|--apply] [--user USER]

Migrates stale repo paths in crontab entries, currently:
  chromecast4k/scripts/monitor-cve.sh          -> hg-android/scripts/monitor-cve.sh
  chromecast4k/scripts/kirkwood-health-cron.sh -> hg-android/scripts/kirkwood-health-cron.sh

Exit codes:
  0  no changes needed, or apply succeeded
  1  legacy paths found during --check
  2  usage error
EOF
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --check)
      MODE="check"
      ;;
    --apply)
      MODE="apply"
      ;;
    --user)
      shift
      [[ "$#" -gt 0 ]] || { usage >&2; exit 2; }
      TARGET_USER="$1"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
  shift
done

if [[ -z "$TARGET_USER" ]]; then
  if [[ "$(id -u)" -eq 0 && -n "${SUDO_USER:-}" ]]; then
    TARGET_USER="$SUDO_USER"
  else
    TARGET_USER="$(id -un)"
  fi
fi

old_monitor="$STUDIO_ROOT/chromecast4k/scripts/monitor-cve.sh"
old_health="$STUDIO_ROOT/chromecast4k/scripts/kirkwood-health-cron.sh"
new_monitor="$STUDIO_ROOT/hg-android/scripts/monitor-cve.sh"
new_health="$STUDIO_ROOT/hg-android/scripts/kirkwood-health-cron.sh"

for path in "$new_monitor" "$new_health"; do
  [[ -x "$path" ]] || {
    printf 'Missing replacement script: %s\n' "$path" >&2
    exit 1
  }
done

crontab_cmd() {
  if [[ "$TARGET_USER" == "$(id -un)" && "$(id -u)" -ne 0 ]]; then
    "$CRONTAB_BIN" "$@"
  else
    "$CRONTAB_BIN" -u "$TARGET_USER" "$@"
  fi
}

current_crontab="$(crontab_cmd -l 2>/dev/null || true)"
if [[ -z "$current_crontab" ]]; then
  printf 'No crontab for %s.\n' "$TARGET_USER"
  exit 0
fi

updated_crontab="${current_crontab//$old_monitor/$new_monitor}"
updated_crontab="${updated_crontab//$old_health/$new_health}"

if [[ "$updated_crontab" == "$current_crontab" ]]; then
  printf 'No legacy repo paths found in %s crontab.\n' "$TARGET_USER"
  exit 0
fi

current_file="$(mktemp)"
updated_file="$(mktemp)"
trap 'rm -f "$current_file" "$updated_file"' EXIT
printf '%s\n' "$current_crontab" > "$current_file"
printf '%s\n' "$updated_crontab" > "$updated_file"

if [[ "$MODE" == "check" ]]; then
  printf 'Legacy repo paths detected in %s crontab:\n' "$TARGET_USER"
  diff -u --label "current:$TARGET_USER" --label "updated:$TARGET_USER" "$current_file" "$updated_file" || true
  exit 1
fi

crontab_cmd "$updated_file"
printf 'Updated %s crontab repo paths:\n' "$TARGET_USER"
diff -u --label "old:$TARGET_USER" --label "new:$TARGET_USER" "$current_file" "$updated_file" || true
