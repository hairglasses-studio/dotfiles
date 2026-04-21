#!/usr/bin/env bash
set -euo pipefail

# validate-config-syntax.sh — repo-wide TOML + JSON syntax gate.
#
# ci-lint.yml validates JSON under clipse/ and TOML via a pip-installed
# legacy library; this gate runs against every tracked *.toml / *.json
# using python's built-in tomllib + json modules (Python 3.11+), so it's
# fast and requires no extra install.
#
# Skips chezmoi `symlink_*` source files — those contain just a target
# path, not valid TOML/JSON.
#
# Usage:
#   validate-config-syntax.sh          # validate all tracked configs
#   validate-config-syntax.sh --count  # print counts only
#
# Exit 0 on clean, 1 if any syntax errors, 2 if git is unavailable.

quiet=0
if [[ "${1:-}" == "--count" ]]; then
    quiet=1
fi

command -v git >/dev/null 2>&1 || {
    printf 'git is required for validate-config-syntax.sh\n' >&2
    exit 2
}

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - "$quiet" <<'PY'
import json
import os
import subprocess
import sys
import tomllib

quiet = sys.argv[1] == "1"
tracked = subprocess.check_output(["git", "ls-files"]).decode().splitlines()


def skip(path: str) -> bool:
    # Chezmoi symlink_* source files contain a target path, not config.
    return os.path.basename(path).startswith("symlink_")


toml_files = [f for f in tracked if f.endswith(".toml") and not skip(f)]
json_files = [f for f in tracked if f.endswith(".json") and not skip(f)]

errors = 0
for f in toml_files:
    try:
        with open(f, "rb") as fh:
            tomllib.load(fh)
    except Exception as exc:
        print(f"TOML FAIL {f}: {exc}")
        errors += 1
for f in json_files:
    try:
        with open(f) as fh:
            json.load(fh)
    except Exception as exc:
        print(f"JSON FAIL {f}: {exc}")
        errors += 1

if not quiet or errors:
    print(f"toml={len(toml_files)} json={len(json_files)} errors={errors}")

sys.exit(0 if errors == 0 else 1)
PY
