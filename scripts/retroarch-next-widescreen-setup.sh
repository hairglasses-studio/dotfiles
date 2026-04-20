#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

"${script_dir}/retroarch-install-widescreen-cores.sh"
"${script_dir}/retroarch-dolphin-sync-sys.sh"
"${script_dir}/retroarch-flycast-apply-widescreen-defaults.py"
"${script_dir}/retroarch-dolphin-apply-widescreen-defaults.py"
"${script_dir}/retroarch-flycast-widescreen-audit.py"
"${script_dir}/retroarch-dolphin-widescreen-audit.py"
