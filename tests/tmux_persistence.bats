#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export PATH="${TEST_BIN}:${PATH}"
    export TMUX_LOG="${BATS_TEST_TMPDIR}/tmux.log"
    export GIT_LOG="${BATS_TEST_TMPDIR}/git.log"
    mkdir -p "${TEST_BIN}" "${HOME}"

    cat > "${HOME}/.tmux.conf" <<'EOF'
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'
EOF

    cat > "${TEST_BIN}/tmux" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${TMUX_LOG}"
if [[ "${1:-}" == "has-session" ]]; then
    exit "${TMUX_MAIN_SESSION_EXIT:-1}"
fi
exit 0
EOF
    chmod +x "${TEST_BIN}/tmux"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "tmux-persistence-health fails when TPM and restore plugins are missing" {
    run bash "${SCRIPTS_DIR}/tmux-persistence-health.sh" --json
    assert_failure
    assert_output --partial '"check":"tpm_runtime","status":"fail"'
    assert_output --partial '"check":"resurrect_runtime","status":"fail"'
    assert_output --partial '"check":"continuum_runtime","status":"fail"'
}

@test "tmux-persistence-health passes when TPM and restore plugins are installed" {
    export TMUX_MAIN_SESSION_EXIT=0
    mkdir -p "${HOME}/.tmux/plugins/tpm/bin" "${HOME}/.tmux/plugins/tmux-resurrect" "${HOME}/.tmux/plugins/tmux-continuum"
    printf '#!/usr/bin/env bash\n' > "${HOME}/.tmux/plugins/tpm/tpm"
    printf '#!/usr/bin/env bash\n' > "${HOME}/.tmux/plugins/tpm/bin/install_plugins"
    chmod +x "${HOME}/.tmux/plugins/tpm/tpm" "${HOME}/.tmux/plugins/tpm/bin/install_plugins"

    run bash "${SCRIPTS_DIR}/tmux-persistence-health.sh" --json
    assert_success
    assert_output --partial '"errors":0'
    assert_output --partial '"check":"main_session","status":"pass"'
}

@test "tmux-persistence-bootstrap clones TPM and installs restore plugins" {
    export TMUX_PLUGIN_ROOT="${HOME}/.tmux/plugins"
    export TMUX_TPM_DIR="${TMUX_PLUGIN_ROOT}/tpm"

    cat > "${TEST_BIN}/git" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${GIT_LOG}"
mkdir -p "${TMUX_TPM_DIR}/bin"
printf '#!/usr/bin/env bash\n' > "${TMUX_TPM_DIR}/tpm"
cat > "${TMUX_TPM_DIR}/bin/install_plugins" <<'INSTALL'
#!/usr/bin/env bash
mkdir -p "${TMUX_PLUGIN_ROOT}/tmux-resurrect" "${TMUX_PLUGIN_ROOT}/tmux-continuum"
INSTALL
chmod +x "${TMUX_TPM_DIR}/tpm" "${TMUX_TPM_DIR}/bin/install_plugins"
EOF
    chmod +x "${TEST_BIN}/git"

    run bash "${SCRIPTS_DIR}/tmux-persistence-bootstrap.sh"
    assert_success

    run test -d "${HOME}/.tmux/plugins/tmux-resurrect"
    assert_success
    run test -d "${HOME}/.tmux/plugins/tmux-continuum"
    assert_success

    run cat "${GIT_LOG}"
    assert_success
    assert_output --partial "clone https://github.com/tmux-plugins/tpm ${HOME}/.tmux/plugins/tpm"

    run cat "${TMUX_LOG}"
    assert_success
    assert_output --partial "start-server"
    assert_output --partial "source-file ${HOME}/.tmux.conf"
}
