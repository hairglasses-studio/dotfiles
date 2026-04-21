#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export XDG_CONFIG_HOME="${HOME}/.config"
    export XDG_STATE_HOME="${HOME}/.local/state"
    mkdir -p "${HOME}" "${XDG_CONFIG_HOME}" "${XDG_STATE_HOME}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "retroarch-archive-homebrew-playlists honors the shared RetroArch profile mapping" {
    local manifest="${BATS_TEST_TMPDIR}/manifest.json"
    local rom_root="${HOME}/Games/RetroArch/roms"
    local playlist_root="${HOME}/.config/retroarch/playlists"
    local profile="${BATS_TEST_TMPDIR}/retroarch.yaml"

    mkdir -p "${rom_root}/gb/archive-homebrew" "${playlist_root}"
    printf 'payload' > "${BATS_TEST_TMPDIR}/Demo Game.gb"
    python3 - <<PY
import zipfile
with zipfile.ZipFile("${rom_root}/gb/archive-homebrew/Demo Game.zip", "w") as archive:
    archive.write("${BATS_TEST_TMPDIR}/Demo Game.gb", arcname="Demo Game.gb")
PY

    cat > "${manifest}" <<'JSON'
{
  "entries": [
    {
      "system": "gb",
      "tier": "public_domain",
      "file_name": "Demo Game.zip",
      "default_selected": true
    }
  ]
}
JSON

    cat > "${profile}" <<'YAML'
id: retroarch
name: "RetroArch"
description: "test"
path_template: "/roms/{platform}/{game}{ext}"
max_filename_length: 255
strip_region: false
strip_revision: false
supported_platforms:
  - gb
preferred_extensions:
  - .gb
default_regions: ["USA"]
default_languages: ["En"]
retroarch_playlist_map:
  - "gb|Custom - Game Boy.lpl|sameboy_libretro.so|Custom Core|.gb|libretro-sameboy"
YAML

    run env HOME="${HOME}" XDG_CONFIG_HOME="${XDG_CONFIG_HOME}" \
        python3 "${DOTFILES_DIR}/scripts/retroarch-archive-homebrew-playlists.py" \
        --manifest "${manifest}" \
        --rom-root "${rom_root}" \
        --playlist-root "${playlist_root}" \
        --retroarch-profile "${profile}"
    assert_success

    run python3 - <<PY
import json
from pathlib import Path
path = Path("${playlist_root}/Custom - Game Boy.lpl")
payload = json.loads(path.read_text())
item = payload["items"][0]
assert item["db_name"] == "Custom - Game Boy.lpl"
assert item["label"] == "Demo Game"
assert item["path"].endswith("Demo Game.zip#Demo Game.gb")
print("ok")
PY
    assert_success
    assert_output "ok"
}

@test "retroarch-workstation-audit reports zip-backed ROM counts and conditional BIOS correctly" {
    local config_dir="${XDG_CONFIG_HOME}/retroarch"
    local rom_root="${HOME}/Games/RetroArch/roms"
    local system_dir="${config_dir}/system"
    local audit_path="${BATS_TEST_TMPDIR}/audit.json"
    local profile="${BATS_TEST_TMPDIR}/retroarch.yaml"

    mkdir -p "${config_dir}" "${rom_root}/psx/archive-homebrew" "${rom_root}/nds" \
        "${system_dir}/pcsx2/bios" "${system_dir}/pcsx2/resources"
    : > "${system_dir}/scph5501.bin"
    : > "${system_dir}/pcsx2/resources/GameIndex.yaml"
    printf 'nds' > "${rom_root}/nds/Test.nds"
    printf 'zipme' > "${BATS_TEST_TMPDIR}/psx.txt"
    python3 - <<PY
import zipfile
with zipfile.ZipFile("${rom_root}/psx/archive-homebrew/Test Game.zip", "w") as archive:
    archive.write("${BATS_TEST_TMPDIR}/psx.txt", arcname="psx.txt")
PY

    cat > "${config_dir}/retroarch.cfg" <<EOF
system_directory = "${system_dir}"
playlist_directory = "${config_dir}/playlists"
video_fullscreen_x = "5120"
video_fullscreen_y = "1440"
video_monitor_index = "2"
network_cmd_enable = "false"
EOF

    cat > "${profile}" <<'YAML'
id: retroarch
name: "RetroArch"
description: "test"
path_template: "/roms/{platform}/{game}{ext}"
max_filename_length: 255
strip_region: false
strip_revision: false
supported_platforms:
  - psx
  - nds
preferred_extensions:
  - .zip
default_regions: ["USA"]
default_languages: ["En"]
retroarch_playlist_map:
  - "psx|Sony - PlayStation.lpl|mednafen_psx_hw_libretro.so|Sony - PlayStation (Beetle PSX HW)|.cue,.chd|libretro-beetle-psx-hw"
  - "nds|Nintendo - Nintendo DS.lpl|desmume_libretro.so|Nintendo - DS (DeSmuME)|.nds|libretro-desmume"
retroarch_requirements:
  - "psx|file|required|scph5501.bin|d41d8cd98f00b204e9800998ecf8427e|US BIOS"
  - "ps2|dir|required|pcsx2/bios||LRPS2 BIOS directory"
  - "ps2|file|required|pcsx2/resources/GameIndex.yaml||LRPS2 compatibility database"
  - "nds|file|conditional|bios7.bin||ARM7 BIOS"
YAML

    run env HOME="${HOME}" XDG_CONFIG_HOME="${XDG_CONFIG_HOME}" \
        python3 "${DOTFILES_DIR}/scripts/retroarch-workstation-audit.py" \
        --config-dir "${config_dir}" \
        --roms-dir "${rom_root}" \
        --retroarch-profile "${profile}" \
        --output "${audit_path}"
    assert_success

    run python3 - <<PY
import json
from pathlib import Path
payload = json.loads(Path("${audit_path}").read_text())
assert payload["display"]["aspect_class"] == "32:9"
psx = next(row for row in payload["cores"] if row["system"] == "psx")
assert psx["rom_count"] == 1
nds_req = next(row for row in payload["requirements"] if row["system"] == "nds")
assert nds_req["requirement"] == "optional"
assert nds_req["status"] == "optional_missing"
print("ok")
PY
    assert_success
    assert_output "ok"
}

@test "retroarch-archive-homebrew-sync orchestrates the end-to-end flow with overridable tools" {
    local tools_dir="${BATS_TEST_TMPDIR}/tools"
    local log_path="${BATS_TEST_TMPDIR}/steps.log"
    local summary_path="${BATS_TEST_TMPDIR}/summary.json"
    local manifest_path="${BATS_TEST_TMPDIR}/manifest.json"
    mkdir -p "${tools_dir}"

    cat > "${tools_dir}/manifest.py" <<EOF
#!/usr/bin/env python3
import json
import pathlib
import sys
from pathlib import Path
log = Path("${log_path}")
args = sys.argv[1:]
prefix = log.read_text() if log.exists() else ""
log.write_text(prefix + "manifest " + " ".join(args) + "\\n")
output = pathlib.Path(args[args.index("--output") + 1])
output.parent.mkdir(parents=True, exist_ok=True)
output.write_text(json.dumps({"entries": []}) + "\\n")
EOF
    cat > "${tools_dir}/fetch.py" <<EOF
#!/usr/bin/env python3
import sys
from pathlib import Path
log = Path("${log_path}")
args = sys.argv[1:]
prefix = log.read_text() if log.exists() else ""
log.write_text(prefix + "fetch " + " ".join(args) + "\\n")
EOF
    cat > "${tools_dir}/import.py" <<EOF
#!/usr/bin/env python3
import sys
from pathlib import Path
log = Path("${log_path}")
args = sys.argv[1:]
prefix = log.read_text() if log.exists() else ""
log.write_text(prefix + "import " + " ".join(args) + "\\n")
EOF
    cat > "${tools_dir}/playlists.py" <<EOF
#!/usr/bin/env python3
import sys
from pathlib import Path
log = Path("${log_path}")
args = sys.argv[1:]
prefix = log.read_text() if log.exists() else ""
log.write_text(prefix + "playlists " + " ".join(args) + "\\n")
EOF
    cat > "${tools_dir}/audit.py" <<EOF
#!/usr/bin/env python3
import json
import pathlib
import sys
from pathlib import Path
log = Path("${log_path}")
args = sys.argv[1:]
prefix = log.read_text() if log.exists() else ""
log.write_text(prefix + "audit " + " ".join(args) + "\\n")
output = pathlib.Path(args[args.index("--output") + 1])
output.parent.mkdir(parents=True, exist_ok=True)
output.write_text(json.dumps({
    "runtime": {
        "process_running": False,
        "network_cmd_enable": False,
        "network_cmd_port": 55355
    }
}) + "\\n")
print(output)
EOF
    chmod +x "${tools_dir}/manifest.py" "${tools_dir}/fetch.py" "${tools_dir}/import.py" \
        "${tools_dir}/playlists.py" "${tools_dir}/audit.py"

    run env HOME="${HOME}" XDG_CONFIG_HOME="${XDG_CONFIG_HOME}" \
        RETROARCH_ARCHIVE_MANIFEST_TOOL="${tools_dir}/manifest.py" \
        RETROARCH_ARCHIVE_FETCH_TOOL="${tools_dir}/fetch.py" \
        RETROARCH_ARCHIVE_IMPORT_TOOL="${tools_dir}/import.py" \
        RETROARCH_ARCHIVE_PLAYLIST_TOOL="${tools_dir}/playlists.py" \
        RETROARCH_WORKSTATION_AUDIT_TOOL="${tools_dir}/audit.py" \
        python3 "${DOTFILES_DIR}/scripts/retroarch-archive-homebrew-sync.py" \
        --manifest "${manifest_path}" \
        --output "${summary_path}" \
        --system psx \
        --tier public_domain \
        --dry-run
    assert_success

    run python3 - <<PY
import json
from pathlib import Path
summary = json.loads(Path("${summary_path}").read_text())
assert summary["ok"] is True
assert [step["name"] for step in summary["steps"]] == [
    "manifest",
    "fetch",
    "import",
    "playlists",
    "audit",
]
fetch_args = next(step["argv"] for step in summary["steps"] if step["name"] == "fetch")
assert "--dry-run" in fetch_args
assert fetch_args.count("--system") == 1
assert fetch_args[fetch_args.index("--system") + 1] == "psx"
print("ok")
PY
    assert_success
    assert_output "ok"
}
