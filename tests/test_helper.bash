#!/usr/bin/env bash
# test_helper.bash -- shared setup for all bats tests

# Load bats helpers (try system install first, then submodule)
if [[ -d /usr/lib/bats/bats-support ]]; then
    load '/usr/lib/bats/bats-support/load'
    load '/usr/lib/bats/bats-assert/load'
elif [[ -d /usr/lib/bats-support ]]; then
    load '/usr/lib/bats-support/load'
    load '/usr/lib/bats-assert/load'
elif [[ -d "${BATS_TEST_DIRNAME}/lib/bats-support" ]]; then
    load "${BATS_TEST_DIRNAME}/lib/bats-support/load"
    load "${BATS_TEST_DIRNAME}/lib/bats-assert/load"
fi

# Set up paths
export DOTFILES_DIR="${BATS_TEST_DIRNAME}/.."
export LIB_DIR="${DOTFILES_DIR}/scripts/lib"
export SCRIPTS_DIR="${DOTFILES_DIR}/scripts"

# Create temp directory for each test
setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# Mock notify-send for tests that trigger notifications
mock_notify_send() {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/notify-send" << 'MOCK'
#!/usr/bin/env bash
echo "NOTIFY: $*" >> "${BATS_TEST_TMPDIR}/notifications.log"
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/notify-send"
}

# Mock compositor commands (hyprctl, aerospace, swaymsg)
mock_compositor() {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    for cmd in hyprctl aerospace swaymsg; do
        cat > "${BATS_TEST_TMPDIR}/${cmd}" << MOCK
#!/usr/bin/env bash
echo "MOCK_${cmd}: \$*" >> "${BATS_TEST_TMPDIR}/compositor.log"
echo '{"ok": true}'
MOCK
        chmod +x "${BATS_TEST_TMPDIR}/${cmd}"
    done
}
