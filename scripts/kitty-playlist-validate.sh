#!/usr/bin/env bash
set -euo pipefail

# kitty-playlist-validate.sh — Validate kitty theme playlist entries against
# the bundled kitty/themes/themes.json catalog.
#
# The runtime picker in `kitty-shader-playlist.sh` already skips invalid
# entries via `kitty +kitten themes --dump-theme` (~10s per theme), which
# masks typos in playlist files. This script is the edit-time check: it
# resolves every non-comment entry against the catalog in milliseconds
# and reports misses with fuzzy-match suggestions.
#
# Usage:
#   kitty-playlist-validate.sh              # validate every *.txt in playlists/
#   kitty-playlist-validate.sh cyberpunk    # validate one playlist
#   kitty-playlist-validate.sh a b c        # validate several
#
# Exit codes:
#   0  all entries resolve cleanly
#   1  at least one entry missing from the catalog
#   2  catalog or playlist file missing
#
# Intended to run in CI and pre-commit contexts; also useful as a quick
# sanity check after editing a playlist by hand.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
DOTFILES="$(cd "$(dirname "$SCRIPT_PATH")/.." && pwd)"
CATALOG="$DOTFILES/kitty/themes/themes.json"
PLAYLIST_DIR="$DOTFILES/kitty/themes/playlists"

if [[ ! -f "$CATALOG" ]]; then
    printf 'catalog not found: %s\n' "$CATALOG" >&2
    exit 2
fi

if [[ ! -d "$PLAYLIST_DIR" ]]; then
    printf 'playlist dir not found: %s\n' "$PLAYLIST_DIR" >&2
    exit 2
fi

declare -a targets=()
if [[ $# -eq 0 ]]; then
    while IFS= read -r -d '' path; do
        name="${path##*/}"
        targets+=("${name%.txt}")
    done < <(find "$PLAYLIST_DIR" -maxdepth 1 -type f -name '*.txt' -print0 | sort -z)
else
    targets=("$@")
fi

if [[ ${#targets[@]} -eq 0 ]]; then
    printf 'no playlists to validate (playlist dir empty?)\n' >&2
    exit 0
fi

python3 - "$CATALOG" "$PLAYLIST_DIR" "${targets[@]}" <<'PY'
import difflib
import json
import sys
from pathlib import Path

catalog_path = Path(sys.argv[1])
playlist_dir = Path(sys.argv[2])
targets = sys.argv[3:]

names = {theme["name"] for theme in json.loads(catalog_path.read_text())}

missing_total = 0
for name in targets:
    path = playlist_dir / f"{name}.txt"
    if not path.is_file():
        print(f"{name}: MISSING FILE ({path})", file=sys.stderr)
        missing_total += 1
        continue
    entries = []
    for raw in path.read_text().splitlines():
        s = raw.strip()
        if not s or s.startswith("#"):
            continue
        entries.append(s)
    missing = [e for e in entries if e not in names]
    if missing:
        print(f"{name}: {len(entries)} entries, {len(missing)} missing")
        for m in missing:
            suggestions = difflib.get_close_matches(m, names, n=3, cutoff=0.6)
            if suggestions:
                print(f"  MISSING: {m!r} — did you mean: {', '.join(repr(s) for s in suggestions)}?")
            else:
                print(f"  MISSING: {m!r}")
        missing_total += len(missing)
    else:
        print(f"{name}: {len(entries)} entries, all resolve")

if missing_total:
    print(f"\n{missing_total} missing entr{'y' if missing_total == 1 else 'ies'} total", file=sys.stderr)
    sys.exit(1)
PY
