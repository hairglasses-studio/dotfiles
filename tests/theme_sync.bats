#!/usr/bin/env bats

load 'test_helper'

setup() {
    BATS_TEST_TMPDIR="$(mktemp -d)"
    export BATS_TEST_TMPDIR
    export HOME="${BATS_TEST_TMPDIR}/home"
    export DOTFILES_DIR="${BATS_TEST_TMPDIR}/dotfiles"
    export THEME_SYNC_ARGS_LOG="${BATS_TEST_TMPDIR}/palette-propagate.args"
    export THEME_SYNC_CALLS_LOG="${BATS_TEST_TMPDIR}/runtime.calls"

    mkdir -p "${HOME}" "${DOTFILES_DIR}/matugen/templates" "${DOTFILES_DIR}/scripts/lib" "${DOTFILES_DIR}/theme" "${BATS_TEST_TMPDIR}/bin"
    touch "${DOTFILES_DIR}/AGENTS.md"
    cp "${SCRIPTS_DIR}/palette-propagate.sh" "${DOTFILES_DIR}/scripts/"
    cp "${SCRIPTS_DIR}/theme-sync.sh" "${DOTFILES_DIR}/scripts/"
    cp "${SCRIPTS_DIR}/lib/hg-core.sh" "${DOTFILES_DIR}/scripts/lib/"

    cat > "${DOTFILES_DIR}/theme/palette.env" <<'PALETTE'
THEME_NAME="Test Palette"
THEME_PRIMARY="12ab34"
THEME_SECONDARY="56cd78"
THEME_TERTIARY="90ef12"
THEME_ICON_THEME="Papirus-Test"
THEME_CURSOR_THEME="Cursor-Test"
THEME_CURSOR_SIZE="32"
PALETTE

    cat > "${DOTFILES_DIR}/scripts/palette-propagate.stub.sh" <<'STUB'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${THEME_SYNC_ARGS_LOG:?}"
STUB
    chmod +x "${DOTFILES_DIR}/scripts/palette-propagate.stub.sh"

    cat > "${DOTFILES_DIR}/scripts/lib/shell-stack.sh" <<'STUB'
#!/usr/bin/env bash
shell_stack_load() { :; }
shell_stack_quickshell_wanted() { return 1; }
STUB

    cat > "${DOTFILES_DIR}/matugen/templates/qt-colorscheme.conf" <<'TEMPLATE'
cursor_theme=${THEME_CURSOR_THEME}
cursor_size=${THEME_CURSOR_SIZE}
icon_theme=${THEME_ICON_THEME}
TEMPLATE

    for cmd in gsettings xfconf-query systemctl dbus-update-activation-environment; do
        cat > "${BATS_TEST_TMPDIR}/bin/${cmd}" <<'STUB'
#!/usr/bin/env bash
printf '%s %s\n' "$(basename "$0")" "$*" >> "${THEME_SYNC_CALLS_LOG:?}"
STUB
        chmod +x "${BATS_TEST_TMPDIR}/bin/${cmd}"
    done

    export PATH="${BATS_TEST_TMPDIR}/bin:${PATH}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "theme-sync --quiet does not suppress palette reload hooks" {
    mv "${DOTFILES_DIR}/scripts/palette-propagate.stub.sh" "${DOTFILES_DIR}/scripts/palette-propagate.sh"
    run bash "${DOTFILES_DIR}/scripts/theme-sync.sh" --quiet
    assert_success
    assert_output ""

    run cat "${THEME_SYNC_ARGS_LOG}"
    assert_success
    assert_output ""
}

@test "theme-sync --no-reload passes no-reload through to palette propagation" {
    mv "${DOTFILES_DIR}/scripts/palette-propagate.stub.sh" "${DOTFILES_DIR}/scripts/palette-propagate.sh"
    run bash "${DOTFILES_DIR}/scripts/theme-sync.sh" --quiet --no-reload
    assert_success
    assert_output ""

    run cat "${THEME_SYNC_ARGS_LOG}"
    assert_success
    assert_output "--no-reload"
}

@test "theme-sync applies icon and cursor palette tokens to runtime settings" {
    mv "${DOTFILES_DIR}/scripts/palette-propagate.stub.sh" "${DOTFILES_DIR}/scripts/palette-propagate.sh"
    run bash "${DOTFILES_DIR}/scripts/theme-sync.sh" --quiet --no-reload
    assert_success

    run cat "${THEME_SYNC_CALLS_LOG}"
    assert_success
    assert_output --partial "gsettings set org.gnome.desktop.interface icon-theme Papirus-Test"
    assert_output --partial "gsettings set org.gnome.desktop.interface cursor-theme Cursor-Test"
    assert_output --partial "gsettings set org.gnome.desktop.interface cursor-size 32"
    assert_output --partial "xfconf-query -c xsettings -p /Net/IconThemeName -n -t string -s Papirus-Test"
    assert_output --partial "xfconf-query -c xsettings -p /Gtk/CursorThemeName -n -t string -s Cursor-Test"
    assert_output --partial "xfconf-query -c xsettings -p /Gtk/CursorThemeSize -n -t int -s 32"
}

@test "palette-propagate substitutes cursor and icon tokens in templates" {
    run bash "${DOTFILES_DIR}/scripts/palette-propagate.sh" --no-reload
    assert_success

    run cat "${HOME}/.config/qt5ct/colors/hairglasses.conf"
    assert_success
    assert_output --partial "cursor_theme=Cursor-Test"
    assert_output --partial "cursor_size=32"
    assert_output --partial "icon_theme=Papirus-Test"
}
