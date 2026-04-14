#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export CRON_FILE="${BATS_TEST_TMPDIR}/crontab.txt"
    export CRONTAB_LOG="${BATS_TEST_TMPDIR}/crontab.log"
    export MOCK_BIN="${BATS_TEST_TMPDIR}/bin"
    mkdir -p "${MOCK_BIN}"

    cat > "${MOCK_BIN}/crontab" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail

log_file="${CRONTAB_LOG:?}"
cron_file="${CRON_FILE:?}"

printf '%s\n' "$*" >> "$log_file"

if [[ "${1:-}" == "-u" ]]; then
  shift 2
fi

case "${1:-}" in
  -l)
    if [[ -f "$cron_file" ]]; then
      cat "$cron_file"
      exit 0
    fi
    exit 1
    ;;
  *)
    cp "$1" "$cron_file"
    ;;
esac
MOCK
    chmod +x "${MOCK_BIN}/crontab"

    export PATH="${MOCK_BIN}:${PATH}"
    export CRONTAB_BIN="crontab"
    export MIGRATE_SCRIPT="${SCRIPTS_DIR}/migrate-repo-crontab-paths.sh"
    export HG_STUDIO_ROOT="$(cd "${DOTFILES_DIR}/.." && pwd)"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "migrate-repo-crontab-paths: check mode reports legacy chromecast4k paths" {
    cat > "${CRON_FILE}" <<EOF
0 9 * * * ${HG_STUDIO_ROOT}/chromecast4k/scripts/monitor-cve.sh >> /home/hg/.kirkwood/cve-monitor.log 2>&1
0 */6 * * * ${HG_STUDIO_ROOT}/chromecast4k/scripts/kirkwood-health-cron.sh
EOF

    run bash "${MIGRATE_SCRIPT}" --check
    assert_failure 1
    assert_output --partial "Legacy repo paths detected"
    assert_output --partial "${HG_STUDIO_ROOT}/hg-android/scripts/monitor-cve.sh"
    assert_output --partial "${HG_STUDIO_ROOT}/hg-android/scripts/kirkwood-health-cron.sh"
}

@test "migrate-repo-crontab-paths: apply mode rewrites legacy paths in place" {
    cat > "${CRON_FILE}" <<EOF
0 9 * * * ${HG_STUDIO_ROOT}/chromecast4k/scripts/monitor-cve.sh >> /home/hg/.kirkwood/cve-monitor.log 2>&1
0 */6 * * * ${HG_STUDIO_ROOT}/chromecast4k/scripts/kirkwood-health-cron.sh
EOF

    run bash "${MIGRATE_SCRIPT}" --apply
    assert_success
    assert_output --partial "Updated"

    run cat "${CRON_FILE}"
    assert_success
    assert_output --partial "${HG_STUDIO_ROOT}/hg-android/scripts/monitor-cve.sh"
    assert_output --partial "${HG_STUDIO_ROOT}/hg-android/scripts/kirkwood-health-cron.sh"
}

@test "migrate-repo-crontab-paths: clean crontab exits successfully without rewrite" {
    cat > "${CRON_FILE}" <<EOF
0 9 * * * ${HG_STUDIO_ROOT}/hg-android/scripts/monitor-cve.sh >> /home/hg/.kirkwood/cve-monitor.log 2>&1
0 */6 * * * ${HG_STUDIO_ROOT}/hg-android/scripts/kirkwood-health-cron.sh
EOF

    run bash "${MIGRATE_SCRIPT}" --check
    assert_success
    assert_output --partial "No legacy repo paths found in "
    assert_output --partial " crontab."
}
