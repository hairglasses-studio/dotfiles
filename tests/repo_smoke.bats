#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_ROOT="${TEST_HOME}/hairglasses-studio"
    export HOME="${TEST_HOME}"
    export HG_STUDIO_ROOT="${TEST_ROOT}"

    mkdir -p "${HOME}" "${HG_STUDIO_ROOT}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "install.sh prints link specs for core runtime and launcher shims" {
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success
    assert_output --partial "${DOTFILES_DIR}/kitty|${HOME}/.config/kitty"
    assert_output --partial "${DOTFILES_DIR}/ironbar|${HOME}/.config/ironbar"
    assert_output --partial "${DOTFILES_DIR}/scripts/kitty-shell-launch.sh|${HOME}/.local/bin/kitty-shell-launch"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-launcher.sh|${HOME}/.local/bin/app-launcher"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-switcher.sh|${HOME}/.local/bin/app-switcher"
}

@test "installers keep kitty save-session helpers opt-in" {
    run bash -lc "! grep -Eq 'dotfiles-kitty-save-session\\.(service|timer)' '${DOTFILES_DIR}/install.sh' '${DOTFILES_DIR}/manjaro/install.sh'"
    assert_success
}

@test "repo config pins the shell-first kitty launch policy" {
    run bash -lc "grep -F 'default_terminal = \"\$HOME/.local/bin/kitty-shell-launch\"' '${DOTFILES_DIR}/hyprshell/config.toml' && grep -Eq '^startup_session[[:space:]]+none$' '${DOTFILES_DIR}/kitty/kitty.conf'"
    assert_success
}

@test "launcher consumers stay pinned to the managed kitty wrappers" {
    # hyprland/pyprland.toml was removed as a stale duplicate of pypr/config.toml
    # in the April 2026 cleanup; the makima Xbox controller TOML was dropped with
    # the gamepad-remapper retirement. The live pinned consumers are just the
    # three below.
    run bash -lc "grep -F '\$HOME/.local/bin/kitty-dev-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch --class=scratchpad' '${DOTFILES_DIR}/pypr/config.toml' && grep -F 'kitty-visual-launch' '${DOTFILES_DIR}/ironbar/config.toml'"
    assert_success
}

@test "install.sh rejects unknown flags with exit code 2" {
    run bash "${DOTFILES_DIR}/install.sh" --definitely-not-a-real-flag
    assert_failure
    [[ "$status" -eq 2 ]]
    assert_output --partial "Unknown option"
}

@test "hg help renders the module surface" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg" --help
    assert_success
    assert_output --partial "MODULES"
    assert_output --partial "doctor"
    assert_output --partial "config"
}

@test "hg gamepad help exposes the controller workflow" {
    # Replaces the old `hg input` test. The `input` module was retired when
    # makima/juhradial were removed; the surviving controller surface is
    # `hg gamepad` (Xbox + makima service control).
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg" gamepad --help
    assert_success
    assert_output --partial "status"
    assert_output --partial "profiles"
    assert_output --partial "restart"
}

@test "hg workflow sync help stays informational" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg-workflow-sync.sh" --help
    assert_success
    assert_output --partial "Hosted GitHub workflow sync is retired"
    assert_output --partial "informational only"
}

@test "hg workflow sync dry-run exits cleanly without mutating workflows" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg-workflow-sync.sh" --dry-run
    assert_success
    assert_output --partial "Hosted workflow sync is retired"
    assert_output --partial "Dry-run mode is informational only"
    assert_output --partial "No hosted workflows are managed"
}

@test "hg mcp mirror parity list exposes mirrored bundled modules" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --list
    assert_success
    assert_output --partial "dotfiles-mcp"
    assert_output --partial "manual_projection"
    assert_output --partial "mapitall"
    assert_output --partial "mapping"
}

@test "hg mcp mirror parity check passes for the tracked mirror set" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --check
    assert_success
    assert_output --partial "PASS  dotfiles-mcp"
    assert_output --partial "PASS  mapitall"
    assert_output --partial "PASS  mapping"
    # tmux-mcp, systemd-mcp, process-mcp were consolidated into dotfiles-mcp
    # on 2026-04-16; they now live in the `consolidated` array of
    # mcp/mirror-parity.json and are no longer tracked by the parity checker.
    refute_output --partial "PASS  tmux-mcp"
    refute_output --partial "PASS  systemd-mcp"
    refute_output --partial "PASS  process-mcp"
}

@test "kitty theme playlists all resolve against the bundled catalog" {
    run bash "${SCRIPTS_DIR}/kitty-playlist-validate.sh"
    assert_success
    assert_output --partial "all resolve"
    refute_output --partial "MISSING"
}

@test "systemd units under systemd/ pass systemd-analyze verify" {
    # Parse-time gate for .service / .timer unit files. Skips if
    # systemd-analyze is missing on the runner (some minimal CI images).
    # Filters the one intentional informational warning about the
    # opt-in kitty-save-session service whose ExecStart is gated by
    # ConditionPathExists — that file isn't installed by default and
    # the condition makes the missing script a no-op, not a bug.
    command -v systemd-analyze >/dev/null 2>&1 || skip "systemd-analyze not installed"
    run bash -c "systemd-analyze verify '${DOTFILES_DIR}'/systemd/*.service '${DOTFILES_DIR}'/systemd/*.timer 2>&1 \
        | grep -vE 'kitty-save-session\\.service: Command .* is not executable' \
        | grep -vE '^[[:space:]]*$' \
        || true"
    assert_success
    # Assert the filtered output has no residual 'Unknown key' or error lines.
    refute_output --partial "Unknown key"
    refute_output --partial "Failed to"
}

@test "matugen templates only reference palette tokens every palette carries" {
    # Guard against the drift where a new matugen template references
    # ${THEME_FOO} but FOO isn't in every palette env — envsubst would
    # render the placeholder empty, silently breaking palette swaps.
    run bash "${SCRIPTS_DIR}/validate-palette-tokens.sh"
    assert_success
    assert_output --partial "errors=0"
}

@test "compositor and bar configs only reference install.sh .local/bin wrappers" {
    # Guard against the case where install.sh drops a symlink (rename,
    # cleanup) but a hyprland/pypr/ironbar config still invokes the old
    # name — the keybind/widget would silently no-op on dispatch.
    run bash "${SCRIPTS_DIR}/validate-local-bin-refs.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "MISSING"
}

@test "/tmp/bar-*.txt cache consumers all have producers" {
    # Every /tmp/bar-<name>.txt path read by ironbar widgets, ticker
    # streams.toml entries, or ticker Python plugins must have a
    # writer under scripts/bar-*-cache.sh or systemd/bar-*.service.
    # A new widget pointing at an unwritten cache renders blank; a
    # new cache file no consumer reads is dead work each refresh.
    run bash "${SCRIPTS_DIR}/validate-bar-cache-refs.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "ORPHAN"
}

@test "health-watchdog.sh is executable and parses cleanly" {
    # The watchdog is called every 30s by dotfiles-health-watchdog.timer
    # (Tier 1.3). It must parse, be executable, and exit 0 even when the
    # MCP binary and /tmp/bar-* caches look healthy — we don't want the
    # watchdog itself generating noise in the journal.
    assert [ -x "${SCRIPTS_DIR}/health-watchdog.sh" ]
    run bash -n "${SCRIPTS_DIR}/health-watchdog.sh"
    assert_success
}

@test "scripts/lib Python modules all import cleanly" {
    # py_compile (ok 19) catches syntax errors. Import attempts catch
    # the next layer: module-scope NameError, ModuleNotFoundError at
    # the \`from foo import\` site, side-effect failures at module
    # load. scripts/lib/ + scripts/lib/ticker_streams/ is ~38 modules.
    run bash "${SCRIPTS_DIR}/validate-python-imports.sh"
    assert_success
    assert_output --partial "errors=0"
}

@test "README/ROADMAP tool counts stay within 15% of the live snapshot" {
    # Human-readable \"~N tools\" figures in README.md, ROADMAP.md,
    # and mcp/dotfiles-mcp/CLAUDE.md are easy to leave stale when
    # tools land or retire. Tolerance band of ±15% keeps rough claims
    # honest without demanding exact-number churn on every tool add.
    run bash "${SCRIPTS_DIR}/validate-tool-count-claims.sh"
    assert_success
    assert_output --partial "errors=0"
}

@test "tracked TOML, JSON, and YAML configs parse cleanly" {
    # Repo-wide syntax gate for config files. Uses Python's built-in
    # tomllib + json modules (Python 3.11+) and PyYAML for YAML. Skips
    # chezmoi symlink_* source files — those carry a target path, not
    # config. Runs in ~50ms across ~160 files.
    run bash "${SCRIPTS_DIR}/validate-config-syntax.sh"
    assert_success
    assert_output --partial "errors=0"
}

@test "scripts Python sources all py_compile cleanly" {
    # Cheap repo-wide syntax gate for Python scripts (~0.1s on 63 files).
    # Catches SyntaxError / IndentationError / unterminated f-strings before
    # they ship — the retroarch bats already does py_compile for its own
    # helper set, this generalizes the check to every *.py under scripts/.
    run bash -c "find '${DOTFILES_DIR}/scripts' -name '*.py' -type f -exec python3 -m py_compile {} +"
    assert_success
}

@test "scripts shell sources all pass bash -n" {
    # Repo-wide parse-time gate for bash scripts under scripts/ and
    # scripts/lib/. Catches syntax errors shellcheck warnings don't —
    # e.g. a stray \`; do\` on its own line, an unmatched \`fi\`, or a
    # heredoc with a mismatched terminator. Skips the one zsh script
    # (scripts/ccg.sh) and the one python-shebang helper
    # (scripts/hypr-monitor-watch.sh). Runs in well under a second.
    run bash -c '
        fails=0
        for f in "'"${DOTFILES_DIR}"'"/scripts/*.sh "'"${DOTFILES_DIR}"'"/scripts/lib/*.sh; do
            [[ -f "$f" ]] || continue
            head -c 40 "$f" | grep -q -E "env (bash|sh)" || continue
            bash -n "$f" 2>&1 || { echo "FAIL: $f"; fails=$((fails + 1)); }
        done
        echo "shell_fails=$fails"
        [[ $fails -eq 0 ]]
    '
    assert_success
    assert_output --partial "shell_fails=0"
}

@test "scripts/ has no unreferenced orphans outside the allowlist" {
    # A new script that neither lands in the allowlist nor gets wired into a
    # consumer (install.sh, systemd, a config, another script, etc.) is a
    # common class of cleanup debt. Keep the audit green; if a genuinely
    # new hook-installable or manual-invoke utility lands, extend the
    # allowlist inside audit-orphan-scripts.sh with a category comment.
    run bash "${SCRIPTS_DIR}/audit-orphan-scripts.sh" --strict
    assert_success
    assert_output --partial "orphans=0"
}

@test "skill surface matches the .agents/skills/ directory listing" {
    # .agents/skills/surface.yaml is the curated registry that drives
    # chezmoi skill publication. A skill dir on disk without an entry
    # in surface.yaml silently fails to publish; an entry without a
    # directory points at nothing. This gate also checks each canonical
    # SKILL.md carries a frontmatter `name:` matching its directory —
    # a hyphen-vs-underscore drift silently breaks invocation routing.
    # Runs in ~100ms.
    run python3 -c "
import json, os, re, sys
root = '${DOTFILES_DIR}/.agents/skills'
d = json.load(open(f'{root}/surface.yaml'))
declared = {s['name'] for s in d['skills']}
dirs = {e for e in os.listdir(root) if os.path.isdir(os.path.join(root, e))}
only_declared = declared - dirs
only_dirs = dirs - declared
name_drift = []
for subdir in sorted(dirs):
    skill_md = os.path.join(root, subdir, 'SKILL.md')
    if not os.path.isfile(skill_md):
        continue
    text = open(skill_md).read()
    m = re.search(r'^name:\s*(\S+)', text, re.M)
    if m and m.group(1) != subdir:
        name_drift.append(f'{subdir}: frontmatter name={m.group(1)}')
if only_declared or only_dirs or name_drift:
    print(f'declared_but_no_dir={sorted(only_declared)}')
    print(f'dir_but_not_declared={sorted(only_dirs)}')
    print(f'name_drift={name_drift}')
    sys.exit(1)
print(f'skills={len(declared)} drift=0')
"
    assert_success
    assert_output --partial "drift=0"
}

@test "install.sh --list-services units all resolve to a unit file" {
    # install.sh's desktop_service_units + desktop_passive_units arrays drive
    # systemctl --user enable/start. A unit declared in the array but missing
    # on disk silently fails to enable — systemctl reports "Unit … not found"
    # and the installer moves on. Gate every listed unit against:
    #   systemd/<name>                   — plain unit
    #   systemd/<base>@.<suffix>         — template for name@instance.service
    #   home/dot_config/systemd/user/symlink_<name>   — chezmoi placeholder
    #   home/dot_config/systemd/user/<name>           — literal placeholder
    #   manjaro/systemd/<name>           — Manjaro-only variant
    run bash -c '
        declared=$(bash "'"${DOTFILES_DIR}"'"/install.sh --list-services 2>&1 \
            | grep -E "^[[:space:]]+[a-z]" | awk "{print \$1}")
        missing=0
        while IFS= read -r unit; do
            [[ -z "$unit" ]] && continue
            base="${unit%%@*}"
            suffix="${unit##*.}"
            found=""
            for p in \
                "'"${DOTFILES_DIR}"'/systemd/${unit}" \
                "'"${DOTFILES_DIR}"'/systemd/${base}@.${suffix}" \
                "'"${DOTFILES_DIR}"'/manjaro/systemd/${unit}" \
                "'"${DOTFILES_DIR}"'/home/dot_config/systemd/user/${unit}" \
                "'"${DOTFILES_DIR}"'/home/dot_config/systemd/user/symlink_${unit}"; do
                if [[ -f "$p" || -L "$p" ]]; then
                    found="$p"
                    break
                fi
            done
            if [[ -z "$found" ]]; then
                echo "MISSING: $unit"
                missing=$((missing + 1))
            fi
        done <<< "$declared"
        echo "missing_units=$missing"
        [[ $missing -eq 0 ]]
    '
    assert_success
    assert_output --partial "missing_units=0"
}

@test "shader playlists all resolve to darkwindow .glsl files" {
    # Parallel to the kitty theme playlist gate (ok 12), but for the
    # shader rotation. kitty-shader-playlist.sh silently drops entries
    # that don't resolve, so a rename or a cleanup that deletes a .glsl
    # without updating the playlist quietly shrinks the rotation.
    run bash "${SCRIPTS_DIR}/validate-shader-playlists.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "MISSING"
}

@test "shader preset blocks all resolve to darkwindow .glsl files" {
    # Every [shaders.<name>.presets.*] in presets.toml must correspond
    # to <name>.glsl. The shader_preset_apply MCP tool opens the shader
    # source by this derived path; a rename that misses presets.toml
    # makes the preset silently no-op at invocation time.
    run bash "${SCRIPTS_DIR}/validate-shader-presets.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "MISSING"
}

@test "install.sh retroarch entries resolve to executable scripts" {
    # install.sh maps \$DOTFILES_DIR/scripts/retroarch-*.{py,sh} to
    # \$HOME/.local/bin/retroarch-*. A rename that misses one side silently
    # breaks the installer; this test catches that by verifying every
    # retroarch link-spec on the left side exists and is executable.
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success

    local missing=()
    local non_executable=()
    while IFS='|' read -r src dest; do
        [[ "${src}" == *retroarch-* ]] || continue
        local resolved="${src/\$DOTFILES_DIR/${DOTFILES_DIR}}"
        if [[ ! -f "${resolved}" ]]; then
            missing+=("${resolved}")
        elif [[ ! -x "${resolved}" ]]; then
            non_executable+=("${resolved}")
        fi
    done <<< "${output}"

    [[ ${#missing[@]} -eq 0 ]] || fail "missing retroarch sources: ${missing[*]}"
    [[ ${#non_executable[@]} -eq 0 ]] || fail "retroarch scripts missing +x bit: ${non_executable[*]}"
}

@test "compositor/bar configs only reference existing dotfiles paths" {
    # Bypass the $HOME/.local/bin indirection: hyprland binds and
    # ironbar/pypr commands sometimes point straight at
    # $HOME/hairglasses-studio/dotfiles/<path>. ok 15 only covers
    # ~/.local/bin/ wrappers — direct dotfile paths slip through, and
    # a rename leaves a bind that silently no-ops on hyprctl dispatch.
    run bash "${SCRIPTS_DIR}/validate-dotfiles-refs.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "MISSING"
}

@test "retroarch_archive_homebrew_verified.json entries match manifest sources" {
    # verified.json declares source_identifier + archive_path per entry.
    # The orchestrator splits archive_path on the first `/` and looks
    # the display name up in the SourceItem.system_dirs map from
    # retroarch-archive-homebrew-manifest.py. An entry with a prefix
    # that isn't in the map produces no playlist row — silent skip.
    # Likewise a source_identifier that doesn't match any SourceItem.
    run bash "${SCRIPTS_DIR}/validate-archive-homebrew-manifest.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "CLAUDE.md GEMINI.md copilot-instructions.md stay thin AGENTS.md mirrors" {
    # The per-repo convention is that AGENTS.md is the canonical
    # instruction file and CLAUDE.md / GEMINI.md / .github/copilot-
    # instructions.md are thin compatibility mirrors pointing at it.
    # Catches two drift modes: (a) a mirror loses its AGENTS.md link
    # so tools silently read only the mirror, (b) a mirror grows
    # into a parallel instruction source (the per-repo convention
    # says Claude-specific notes may only live in CLAUDE.md when
    # they cannot live in AGENTS.md — so 50 lines is plenty).
    run bash "${SCRIPTS_DIR}/validate-agents-mirror-contract.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "hyprland pypr toggle binds resolve to declared scratchpads" {
    # `pypr toggle <name>` in hyprland binds depends on a
    # [scratchpads.<name>] block in pypr/config.toml, or on a pypr
    # built-in command. A rename or removal of the scratchpad
    # leaves the keybind dispatching to pypr which politely does
    # nothing — no dialog, no log, no compositor-level feedback.
    run bash "${SCRIPTS_DIR}/validate-pypr-binds.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "orphan-script audit allowlist entries all exist on disk" {
    # scripts/audit-orphan-scripts.sh allowlists certain scripts as
    # "intentionally unreferenced" — user-global Claude hooks, manual
    # helpers, hardware-specific utilities. A rename or removal that
    # misses the allowlist leaves a ghost entry: nothing trips, and
    # the list slowly stops meaning what its comments say. This gate
    # walks the allowlist block and asserts each stem has a matching
    # scripts/<stem>.* file (ignoring .bak variants).
    run bash "${SCRIPTS_DIR}/validate-orphan-allowlist.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "tracked markdown internal links all resolve" {
    # README.md, AGENTS.md, CLAUDE.md, GEMINI.md, ROADMAP.md,
    # docs/**/*.md, and .agents/skills/**/*.md — every relative
    # [text](path.md) target must be a real file. External URLs
    # (http/https/mailto) and pure #anchor-only refs are skipped.
    # Catches rename/move drift that leaves docs pointing at the
    # old filename — GitHub renders the link and follows 404.
    run bash "${SCRIPTS_DIR}/validate-md-links.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "skill aliases in .claude/skills/ point at real canonical dirs" {
    # Every .claude/skills/<dir-with-hyphen>/ is a compatibility alias
    # for a snake_case canonical at both .claude/skills/<snake>/ and
    # .agents/skills/<snake>/. Catches the class where a rename on
    # the canonical leaves the hyphen alias stranded — invocation
    # routing silently drops on the floor because the runtime can't
    # resolve the alias target.
    run bash "${SCRIPTS_DIR}/validate-skill-aliases.sh"
    assert_success
    assert_output --partial "errors=0"
    refute_output --partial "DRIFT"
}

@test "every install.sh scripts/* row resolves to an executable source" {
    # Generalizes ok 28 (retroarch-*) to the whole catalog. A
    # rename or delete that misses install.sh leaves a broken
    # symlink that either errors on `hyprctl dispatch exec, <name>`
    # or silently points at a dangling path. Same class of drift
    # — ok 28 caught it for retroarch, this catches it everywhere.
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success

    local missing=()
    local non_executable=()
    while IFS='|' read -r src dest; do
        # Only sources under scripts/ with .sh or .py extensions.
        [[ "${src}" == *scripts/*.sh || "${src}" == *scripts/*.py ]] || continue
        local resolved="${src/\$DOTFILES_DIR/${DOTFILES_DIR}}"
        if [[ ! -f "${resolved}" ]]; then
            missing+=("${resolved}")
        elif [[ ! -x "${resolved}" ]]; then
            non_executable+=("${resolved}")
        fi
    done <<< "${output}"

    [[ ${#missing[@]} -eq 0 ]] || fail "missing script sources: ${missing[*]}"
    [[ ${#non_executable[@]} -eq 0 ]] || fail "script sources missing +x bit: ${non_executable[*]}"
}
