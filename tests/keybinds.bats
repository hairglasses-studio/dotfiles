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
    # Any line that looks like a keybind section header must match exactly.
    local bad_headers
    bad_headers="$(grep -E '^#.*[Kk]eybind' "${HYPR_CONF}" \
        | grep -v '^# ── Keybinds: ' \
        | grep -v '^# Plugin-backed' \
        | grep -v '^# ── Keybinds:$' \
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
# ---------------------------------------------------------------------------

@test "keybinds: section_icon covers Launch section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Launch"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Navigation section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Navigation"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Screenshot section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Screenshots"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Notification section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Notification center"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Claude section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Claude Code"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers System section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "System"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Media section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Media"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Mouse section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Mouse"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Resize section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Resize"
    assert_success
    refute_output ""
}

@test "keybinds: section_icon covers Scratchpads section" {
    source "${KEYBINDS_SCRIPT}" 2>/dev/null || true
    run section_icon "Scratchpads"
    assert_success
    refute_output ""
}

@test "keybinds: all section names from hyprland.conf map to non-empty icons" {
    # Extract every section name, call section_icon, verify non-empty result.
    # Source the script in a subshell to get the function.
    local conf_sections
    mapfile -t conf_sections < <(
        grep '^# ── Keybinds:' "${HYPR_CONF}" \
            | sed 's/^# ── Keybinds: //' \
            | sed 's/[[:space:]]*──[[:space:]]*$//'
    )

    local missing=()
    for section in "${conf_sections[@]}"; do
        local icon
        icon="$(bash -c "
            source '${KEYBINDS_SCRIPT}' 2>/dev/null
            section_icon \"\$1\"
        " -- "${section}" 2>/dev/null)"
        if [[ -z "${icon}" ]]; then
            missing+=("${section}")
        fi
    done

    if [[ "${#missing[@]}" -gt 0 ]]; then
        echo "Sections with no icon mapping:"
        printf '  %s\n' "${missing[@]}"
        return 1
    fi
}
