#!/usr/bin/env bash
# Find scripts under scripts/ that aren't referenced from any consumer.
#
# Consumers scanned: install.sh, manjaro/install.sh, ci workflows, systemd
# units, compositor/terminal/bar configs, Makefile, other scripts, docs,
# skills, bats tests. Reports candidates only — does not delete.
#
# Usage:
#   scripts/audit-orphan-scripts.sh            # list unreferenced scripts
#   scripts/audit-orphan-scripts.sh --strict   # exit 1 if any are found

set -euo pipefail

strict=0
if [[ "${1:-}" == "--strict" ]]; then
    strict=1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Allowlist: scripts that are legitimate user-invoked entrypoints with no
# in-repo consumer (manually run at the shell, or installed by hand into
# ~/.claude/settings.json hook blocks, or attached to a device profile).
# One name per line, no extension. When you add a new category, keep the
# comment note so the reason is obvious to the next reader.
allowlist=(
    # Claude Code hook installables — users add these to ~/.claude/settings.json
    # hook blocks (PreToolUse / Stop / StopFailure / PostToolUse).
    ccg-hook
    claude-file-protect
    claude-marathon-sync
    claude-notify-debounce
    claude-post-compact
    claude-rate-limit-hook
    claude-stop-guard
    hypr-config-snapshot

    # Meta utilities invoked manually at the shell (or by fleet scripts).
    # No in-repo consumer by design — they exist as helpers users run
    # on demand while auditing or bootstrapping a workstation.
    hg-codex-global-mcp-sync
    hg-instruction-check
    hg-mcp-smoke
    hg-promptfoo
    hg-providerkit-bootstrap
    hg-repomix
    hg-skill-check
    hg-workspace-check

    # Hardware-specific — EN16 hardware device (optional macropad/controller),
    # bound via the user's device profile when they own the hardware.
    en16-encoder-key
    en16-session-focus

    # Desktop UI entrypoints with off-repo bindings (mouse gestures,
    # prototype bars, CLI screenshot wrapper). quickshell-menu-data is
    # consumed via QS IPC by quickshell/modules/MenuOverlay.qml, not via
    # shell, so the orphan auditor's grep doesn't see the reference.
    quickshell-menu-data
    quickshell-try
    screenshot-crop
)

is_allowlisted() {
    local name="$1"
    local entry
    for entry in "${allowlist[@]}"; do
        [[ "$entry" == "$name" ]] && return 0
    done
    return 1
}

# Files that reference scripts. Compositor configs, terminal configs, bar
# configs, systemd overlays, launcher TOMLs, and CI workflows all count.
consumer_globs=(
    "install.sh"
    "manjaro/install.sh"
    "Makefile"
    "*.mk"
    ".github/workflows/*.yml"
    ".github/workflows/*.yaml"
    ".gitlab-ci.yml"
    "systemd"
    "home/dot_config/systemd"
    "hyprland"
    "tmux"
    "kitty"
    "ironbar"
    "pypr"
    "waybar"
    "swaync"
    "hyprshell"
    "glshell"
    "foot"
    "wezterm"
    "scripts"
    "tests"
    "docs"
    ".claude"
    ".agents"
    "home"
    "zsh"
    "manjaro"
    "mcp"
)

# Build the consumer file list once.
consumer_files() {
    local g
    for g in "${consumer_globs[@]}"; do
        if [[ -d "$g" ]]; then
            find "$g" -type f \( \
                -name '*.sh' -o -name '*.py' -o -name '*.md' -o -name '*.yml' \
                -o -name '*.yaml' -o -name '*.json' -o -name '*.toml' \
                -o -name '*.bats' -o -name '*.go' -o -name '*.conf' \
                -o -name '*.service' -o -name '*.timer' -o -name '*.desktop' \
                -o -name 'symlink_*' -o -name 'Makefile' -o -name '*.mk' \
            \) -print 2>/dev/null
        elif [[ -f "$g" ]]; then
            printf '%s\n' "$g"
        fi
    done | sort -u
}

mapfile -t CONSUMERS < <(consumer_files)

orphan_count=0
orphan_list=()

for script in scripts/*.sh scripts/*.py; do
    [[ -f "$script" ]] || continue
    basename="$(basename "$script")"
    stem="${basename%.*}"

    # Skip the audit tool itself
    [[ "$basename" == "audit-orphan-scripts.sh" ]] && continue

    if is_allowlisted "$stem"; then
        continue
    fi

    # Treat as referenced if any consumer file mentions the basename or the
    # stem (to catch `~/.local/bin/<stem>` symlink targets).
    if grep -lFx -e "$basename" -e "$stem" -- "${CONSUMERS[@]}" >/dev/null 2>&1; then
        continue
    fi
    if grep -lF -e "$basename" -e "$stem" -- "${CONSUMERS[@]}" 2>/dev/null \
        | grep -vFx "$script" \
        | grep -q .; then
        continue
    fi

    orphan_list+=("$script")
    orphan_count=$((orphan_count + 1))
done

if [[ $orphan_count -eq 0 ]]; then
    printf 'orphans=0 scripts_scanned=%d consumers=%d\n' \
        "$(ls scripts/*.sh scripts/*.py 2>/dev/null | wc -l)" \
        "${#CONSUMERS[@]}"
    exit 0
fi

printf 'orphans=%d\n' "$orphan_count"
for s in "${orphan_list[@]}"; do
    printf '  %s\n' "$s"
done

if [[ $strict -eq 1 ]]; then
    exit 1
fi
exit 0
