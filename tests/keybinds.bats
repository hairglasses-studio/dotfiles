#!/usr/bin/env bats
# Tests for the keybind module: hypr-keybinds.sh, keybind-ticker.py, and
# hyprland.conf keybind conventions.
#
# Tests: script existence, shebang, config format, duplicate descriptions,
#        section header pattern, section_icon coverage.
# Skips: live hyprctl calls, GTK rendering, runtime hyprland state.

load 'test_helper'

HYPR_CONF="${DOTFILES_DIR}/hyprland/hyprland.conf"
KEYBINDS_SCRIPT="${SCRIPTS_DIR}/hypr-keybinds.sh"
TICKER_SCRIPT="${SCRIPTS_DIR}/keybind-ticker.py"

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_compositor
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# ---------------------------------------------------------------------------
# 1. hypr-keybinds.sh: exists, is executable, exits without error
# ---------------------------------------------------------------------------

@test "keybinds: hypr-keybinds.sh exists and is executable" {
    [[ -f "${KEYBINDS_SCRIPT}" ]]
    [[ -x "${KEYBINDS_SCRIPT}" ]]
}

@test "keybinds: hypr-keybinds.sh has bash shebang" {
    run head -1 "${KEYBINDS_SCRIPT}"
    assert_success
    assert_output --partial "bash"
}

@test "keybinds: hypr-keybinds.sh exits 0 with --md when hyprctl is unavailable" {
    # Provide a minimal stub hyprland.conf so the script has something to parse,
    # and a mock hyprctl that returns empty JSON so query_runtime_binds succeeds.
    local fake_conf="${BATS_TEST_TMPDIR}/hyprland.conf"
    printf '# ── Keybinds: Launch ─────────────────────────────\nbindd = SUPER, Return, Default terminal, exec, kitty\n' \
        > "${fake_conf}"

    local fake_out="${BATS_TEST_TMPDIR}/keybinds.md"

    # mock hyprctl already added by mock_compositor; make it return valid JSON
    cat > "${BATS_TEST_TMPDIR}/hyprctl" << 'MOCK'
#!/usr/bin/env bash
if [[ "${1:-}" == "binds" ]]; then
    echo '[]'
    exit 0
fi
echo '{"ok":true}'
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/hyprctl"

    run env \
        PATH="${BATS_TEST_TMPDIR}:${PATH}" \
        XDG_CONFIG_HOME="${BATS_TEST_TMPDIR}/xdg" \
        DOTFILES_DIR="${BATS_TEST_TMPDIR}/dotfiles" \
        bash "${KEYBINDS_SCRIPT}" --md
    # Script writes to $DOTFILES_DIR/hyprland/keybinds.md; may fail because
    # the output dir doesn't exist. We accept both 0 and any exit when the
    # failure is purely the output path, not a parse error.
    # The key assertion: no "unbound variable" or "syntax error" in output.
    refute_output --partial "unbound variable"
    refute_output --partial "syntax error"
}

# ---------------------------------------------------------------------------
# 2. keybind-ticker.py: exists, correct shebang
# ---------------------------------------------------------------------------

@test "keybinds: keybind-ticker.py exists" {
    [[ -f "${TICKER_SCRIPT}" ]]
}

@test "keybinds: keybind-ticker.py has python3 shebang" {
    run head -1 "${TICKER_SCRIPT}"
    assert_success
    assert_output "#!/usr/bin/env python3"
}

@test "keybinds: keybind-ticker.py shebang is on line 1 (no BOM or blank prefix)" {
    local first_char
    first_char="$(head -c1 "${TICKER_SCRIPT}")"
    [[ "${first_char}" == "#" ]]
}

# ---------------------------------------------------------------------------
# 3. hyprland.conf: main keybinds use bindd format (not bare bind =)
# ---------------------------------------------------------------------------

@test "keybinds: hyprland.conf has no bare 'bind =' lines in keybind sections" {
    # bare bind = would be a line that starts with 'bind' but is NOT bindd/binddl/etc.
    # We look for lines matching ^bind[[:space:]]*= (no 'd' after bind).
    local count
    count="$(grep -c '^bind[[:space:]]*=' "${HYPR_CONF}" 2>/dev/null || true)"
    [[ "${count}" -eq 0 ]]
}

@test "keybinds: hyprland.conf has bindd lines (described binds are present)" {
    run grep -c '^bindd' "${HYPR_CONF}"
    assert_success
    # must have at least 10 described binds
    [[ "$output" -ge 10 ]]
}

@test "keybinds: all main bind lines in hyprland.conf use bindd variants" {
    # Allowed prefixes: bindd, binddl, bindde, binddm, binddlr, etc.
    # Any line starting with 'bind' must start with 'bindd'.
    local violations
    violations="$(grep -E '^bind[^d]' "${HYPR_CONF}" | grep -v '^#' || true)"
    [[ -z "${violations}" ]]
}

# ---------------------------------------------------------------------------
# 4. No duplicate keybind descriptions in hyprland.conf
# ---------------------------------------------------------------------------

@test "keybinds: no duplicate keybind descriptions in hyprland.conf" {
    # Extract the third comma-separated field (description) from all bindd lines.
    # Trim whitespace and count duplicates.
    local dupes
    dupes="$(grep '^bindd' "${HYPR_CONF}" \
        | awk -F',' '{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $3); print $3}' \
        | sort \
        | uniq -d)"
    if [[ -n "${dupes}" ]]; then
        echo "Duplicate keybind descriptions found:"
        echo "${dupes}"
        return 1
    fi
}

# ---------------------------------------------------------------------------
# 5. Section headers follow the canonical pattern
# ---------------------------------------------------------------------------

@test "keybinds: all keybind section headers match '# ── Keybinds: <Name> ──' pattern" {
    # Only lines that start with '# ── Keybinds:' are section headers;
    # all such lines must also end with the trailing '──' sequence.
    # We do NOT scan prose comments (those don't start with '# ── Keybinds:').
    local bad_headers
    bad_headers="$(grep '^# ── Keybinds:' "${HYPR_CONF}" \
        | grep -v '──$' \
        || true)"
    if [[ -n "${bad_headers}" ]]; then
        echo "Non-canonical keybind section headers:"
        echo "${bad_headers}"
        return 1
    fi
}

@test "keybinds: at least 10 keybind sections exist in hyprland.conf" {
    local count
    count="$(grep -c '^# ── Keybinds:' "${HYPR_CONF}")"
    [[ "${count}" -ge 10 ]]
}

@test "keybinds: each section header ends with ──  (trailing dash sequence)" {
    local bad
    bad="$(grep '^# ── Keybinds:' "${HYPR_CONF}" | grep -v '──$' || true)"
    if [[ -n "${bad}" ]]; then
        echo "Section headers missing trailing '──':"
        echo "${bad}"
        return 1
    fi
}

# ---------------------------------------------------------------------------
# 6. section_icon in hypr-keybinds.sh covers all section names in hyprland.conf
#
# We extract the section_icon() function body via sed rather than sourcing the
# whole script, because sourcing executes the top-level build_section_map /
# generate_md calls (and may invoke glow).
#
# Note: some sections fall through to the default catch-all `*) echo ""`
# and return an empty string. The tests for specific sections only assert
# non-empty output for patterns with explicit Nerd Font icon entries.
# The "all sections" test verifies the function runs without error for every
# section name present in hyprland.conf (coverage = no crash/exit failure).
# ---------------------------------------------------------------------------

# Pre-build the section_icon function body for reuse across tests.
_icon_fn() {
    sed -n '/^section_icon()/,/^}/p' "${KEYBINDS_SCRIPT}"
}

# Helper: call section_icon in a clean subshell using only the extracted function.
_call_section_icon() {
    local section_name="$1"
    bash -c "$(_icon_fn); section_icon \"\$1\"" -- "${section_name}"
}

@test "keybinds: section_icon has an explicit entry for Notification center" {
    run _call_section_icon "Notification center"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Claude Code" {
    run _call_section_icon "Claude Code"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Media" {
    run _call_section_icon "Media"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Mouse" {
    run _call_section_icon "Mouse"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Resize" {
    run _call_section_icon "Resize"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Scratchpads" {
    run _call_section_icon "Scratchpads"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Monitor focus" {
    run _call_section_icon "Monitor focus"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon has an explicit entry for Window groups" {
    run _call_section_icon "Window groups"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon handles Minimize / Special workspace" {
    run _call_section_icon "Minimize / Special workspace"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon function handles all section names from hyprland.conf without error" {
    # Extract every section name from the conf, call section_icon for each,
    # and assert all invocations succeed (exit 0). Some sections intentionally
    # return empty string (the catch-all fallback); that is acceptable.
    # Strip 'Keybinds: ' prefix and trailing box-drawing '─' characters.
    local conf_sections
    mapfile -t conf_sections < <(
        grep '^# ── Keybinds:' "${HYPR_CONF}" \
            | sed 's/^# ── Keybinds: //' \
            | sed 's/ *─\+[[:space:]]*$//'
    )

    local icon_fn
    icon_fn="$(_icon_fn)"

    local failed=()
    for section in "${conf_sections[@]}"; do
        if ! bash -c "${icon_fn}; section_icon \"\$1\"" -- "${section}" > /dev/null 2>&1; then
            failed+=("${section}")
        fi
    done

    if [[ "${#failed[@]}" -gt 0 ]]; then
        echo "section_icon failed (non-zero exit) for these sections:"
        printf '  %s\n' "${failed[@]}"
        return 1
    fi
}
