#!/usr/bin/env bash
set -euo pipefail

# retroarch-complete.sh — one-command path to a fully-loaded RetroArch
# workstation. Runs every safe step and hands the user the single
# external-GitHub step (source-build for race + beetle-wswan) as an
# explicit command when it's actually needed.
#
# Each step:
#   - skips if the thing it produces is already present (idempotent),
#   - surfaces what it did (or why it skipped) in a one-line status,
#   - exits the script with a clear summary on success or a clear
#     next-action on first failure.
#
# Does NOT:
#   - run `sudo`, ever (user-local install path throughout),
#   - clone external git repos (delegates that to the user so sandbox
#     and trust boundaries stay visible),
#   - modify retroarch_archive_homebrew_verified.json.
#
# Usage:
#   retroarch-complete
#   retroarch-complete --skip-build    # don't print the build nag
#   retroarch-complete --dry-run       # show the plan without running

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

dry_run=0
skip_build=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) dry_run=1; shift ;;
        --skip-build) skip_build=1; shift ;;
        -h|--help)
            sed -n '3,30p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
            exit 0 ;;
        *) printf 'unknown flag: %s\n' "$1" >&2; exit 2 ;;
    esac
done

run() {
    local label="$1"; shift
    if [[ $dry_run -eq 1 ]]; then
        printf '[plan ] %s: %s\n' "$label" "$*" >&2
        return 0
    fi
    printf '[run  ] %s\n' "$label" >&2
    "$@"
}

note() {
    printf '[note ] %s\n' "$*" >&2
}

ok() {
    printf '[ok   ] %s\n' "$*" >&2
}

# ── Step 1: baseline audit ──────────────────────────────────────────
run "workstation-audit" python3 "${SCRIPT_DIR}/retroarch-workstation-audit.py" >/dev/null

AUDIT_JSON="${XDG_STATE_HOME:-$HOME/.local/state}/retroarch/workstation-audit.json"
if [[ $dry_run -eq 0 && ! -f "$AUDIT_JSON" ]]; then
    printf 'error: audit JSON not produced at %s\n' "$AUDIT_JSON" >&2
    exit 3
fi

# ── Step 2: BIOS/helper dirs ────────────────────────────────────────
run "bios-apply" python3 "${SCRIPT_DIR}/retroarch-bios-apply.py" --system all >/dev/null

# ── Step 3: pacman + AUR cores ──────────────────────────────────────
run "install-workstation-cores" bash "${SCRIPT_DIR}/retroarch-install-workstation-cores.sh"

# ── Step 4: source-built cores (the gated step) ─────────────────────
missing_source_cores=()
if [[ $dry_run -eq 0 ]]; then
    while IFS= read -r line; do
        missing_source_cores+=("$line")
    done < <(python3 - "$AUDIT_JSON" <<'PY'
import json
import sys
data = json.load(open(sys.argv[1]))
source_built = {"race_libretro.so", "mednafen_wswan_libretro.so"}
for core in data.get("cores", []):
    if not core.get("core_installed") and core.get("core_filename") in source_built:
        print(core["core_filename"])
PY
)
fi

if [[ $skip_build -eq 0 && ${#missing_source_cores[@]} -gt 0 ]]; then
    note "source-build cores still missing: ${missing_source_cores[*]}"
    note "the remaining step is external (clones from GitHub + builds locally):"
    note ""
    note "    retroarch-build-libretro-cores --install-dir \"\$HOME/.config/retroarch/cores\""
    note ""
    note "after it finishes, rerun retroarch-complete to finalize."
else
    ok "source-built cores: none missing"
fi

# ── Step 5: network command cfg ─────────────────────────────────────
if [[ $dry_run -eq 0 ]]; then
    network_enabled=$(python3 - "$AUDIT_JSON" <<'PY'
import json
import sys
data = json.load(open(sys.argv[1]))
print("yes" if data.get("runtime", {}).get("network_cmd_enable") else "no")
PY
)
    if [[ "$network_enabled" == "no" ]]; then
        run "apply-network-cmd" python3 "${SCRIPT_DIR}/retroarch-apply-network-cmd.py" >/dev/null
        ok "network_cmd_enable flipped on; restart RetroArch to bind UDP 55355"
    else
        ok "network_cmd_enable already on"
    fi
else
    run "apply-network-cmd (conditional)" python3 "${SCRIPT_DIR}/retroarch-apply-network-cmd.py"
fi

# ── Step 6: post-audit ──────────────────────────────────────────────
if [[ $dry_run -eq 0 ]]; then
    run "post-audit" python3 "${SCRIPT_DIR}/retroarch-workstation-audit.py" >/dev/null
    summary=$(python3 - "$AUDIT_JSON" <<'PY'
import json
import sys
data = json.load(open(sys.argv[1]))
s = data.get("summary", {})
r = data.get("runtime", {})
print(f'core_missing={s.get("core_missing", "?")} '
      f'required_assets_missing={s.get("required_assets_missing", "?")} '
      f'runtime_network={"on" if r.get("network_cmd_enable") else "off"}')
PY
)
    ok "summary: $summary"

    core_missing=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["summary"]["core_missing"])' "$AUDIT_JSON")
    if [[ "$core_missing" == "0" ]]; then
        ok "all cores installed. workstation complete."
    else
        note "$core_missing core(s) still missing — see the build hint above."
    fi
fi
