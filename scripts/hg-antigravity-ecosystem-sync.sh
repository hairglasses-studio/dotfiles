#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="write"
WORKSPACE_ROOT="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}"
ANTIGRAVITY_DIR="${HG_ANTIGRAVITY_DIR:-$HOME/.gemini/antigravity}"
SOURCES_MANIFEST="${HG_ANTIGRAVITY_SOURCES_MANIFEST:-$HG_DOTFILES/config/antigravity/sources.json}"
GLOBAL_SKILLS_DIR="${HG_ANTIGRAVITY_SKILLS_DIR:-$ANTIGRAVITY_DIR/skills}"
SKILLS_LIBRARY_DIR="${HG_ANTIGRAVITY_SKILLS_LIBRARY_DIR:-$ANTIGRAVITY_DIR/skills_library}"
WORKFLOW_LIBRARY_DIR="${HG_ANTIGRAVITY_WORKFLOW_LIBRARY_DIR:-$ANTIGRAVITY_DIR/workflow_library}"
GLOBAL_WORKFLOWS_DIR="${HG_ANTIGRAVITY_GLOBAL_WORKFLOWS_DIR:-$ANTIGRAVITY_DIR/global_workflows}"
TOOL_CACHE_DIR="${HG_ANTIGRAVITY_TOOL_CACHE_DIR:-$ANTIGRAVITY_DIR/tool_cache}"
MCP_TEMPLATE_DIR="${HG_ANTIGRAVITY_MCP_TEMPLATE_DIR:-$ANTIGRAVITY_DIR/mcp_templates}"
METADATA_PATH="${HG_ANTIGRAVITY_ECOSYSTEM_METADATA_PATH:-$ANTIGRAVITY_DIR/.hg-antigravity-ecosystem.json}"
MANAGER_DIR="${HG_ANTIGRAVITY_MANAGER_DIR:-$HOME/.local/opt/AntigravityManager}"
MANAGER_LAUNCHER_PATH="${HG_ANTIGRAVITY_MANAGER_LAUNCHER_PATH:-$HOME/.local/bin/antigravity-manager-launch}"
MANAGER_DESKTOP_PATH="${HG_ANTIGRAVITY_MANAGER_DESKTOP_PATH:-$HOME/.local/share/applications/antigravity-manager.desktop}"
SSH_PROXY_INSTALLER_PATH="${HG_ANTIGRAVITY_SSH_PROXY_INSTALLER_PATH:-$HOME/.local/bin/antigravity-install-ssh-proxy}"
AGENTMEMORY_INSTALLER_PATH="${HG_ANTIGRAVITY_AGENTMEMORY_INSTALLER_PATH:-$HOME/.local/bin/antigravity-install-agentmemory}"
AGENTMEMORY_TEMPLATE_PATH="${HG_ANTIGRAVITY_AGENTMEMORY_TEMPLATE_PATH:-$MCP_TEMPLATE_DIR/agentmemory.disabled.json}"

pending=0
live_global_skill_count=0
global_workflow_count=0
archived_skill_source_count=0
archived_workflow_source_count=0
sidecar_count=0
installed_extension_count=0

declare -A managed_global_skill_set=()
declare -A managed_global_workflow_set=()
declare -A managed_tool_set=()

usage() {
  cat <<'EOF'
Usage: hg-antigravity-ecosystem-sync.sh [--dry-run|--check]

Provision archived Antigravity libraries, activate the selected live global
skill/workflow subset, and cache recommended ecosystem tools.
EOF
}

for arg in "$@"; do
  case "$arg" in
    --dry-run)
      MODE="dry-run"
      ;;
    --check)
      MODE="check"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $arg"
      ;;
  esac
done

hg_require jq git mktemp diff cmp curl awk sed find node npm getent runuser /usr/bin/antigravity

WORKSPACE_ROOT="$(cd "$WORKSPACE_ROOT" && pwd)"
[[ -d "$WORKSPACE_ROOT" ]] || hg_die "Workspace root not found: $WORKSPACE_ROOT"
[[ -f "$SOURCES_MANIFEST" ]] || hg_die "Antigravity sources manifest not found: $SOURCES_MANIFEST"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

sync_rendered_file() {
  local target="$1"
  local rendered="$2"
  local label="$3"

  if [[ -f "$target" ]] && cmp -s "$target" "$rendered"; then
    rm -f "$rendered"
    return 0
  fi

  pending=1
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      mv "$rendered" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      rm -f "$rendered"
      hg_warn "Would update $label: $target"
      ;;
    check)
      rm -f "$rendered"
      hg_warn "Out of date $label: $target"
      ;;
  esac
}

remove_path_if_present() {
  local target="$1"
  local label="$2"
  [[ -e "$target" ]] || return 0
  pending=1
  case "$MODE" in
    write)
      rm -rf "$target"
      hg_ok "Removed stale $label: $target"
      ;;
    dry-run)
      hg_warn "Would remove stale $label: $target"
      ;;
    check)
      hg_warn "Unexpected stale $label: $target"
      ;;
  esac
}

sync_symlink() {
  local target="$1"
  local source="$2"
  local label="$3"

  if [[ -e "$target" && ! -L "$target" ]]; then
    hg_die "$label target exists and is not a symlink: $target"
  fi

  if [[ -L "$target" && "$(readlink "$target")" == "$source" ]]; then
    return 0
  fi

  pending=1
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      ln -sfn "$source" "$target"
      hg_ok "$label: $target"
      ;;
    dry-run)
      hg_warn "Would update $label: $target -> $source"
      ;;
    check)
      hg_warn "Out of date $label: $target"
      ;;
  esac
}

strip_frontmatter() {
  local file="$1"
  awk '
    BEGIN { in_frontmatter = 0; frontmatter_done = 0 }
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { in_frontmatter = 0; frontmatter_done = 1; next }
    !in_frontmatter { print }
  ' "$file"
}

extract_frontmatter_scalar() {
  local file="$1"
  local key="$2"
  awk -v key="$key" '
    BEGIN { in_frontmatter = 0 }
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { exit }
    in_frontmatter && $0 ~ ("^" key ":[[:space:]]*") {
      sub("^" key ":[[:space:]]*", "", $0)
      gsub(/^["'\''"]|["'\''"]$/, "", $0)
      print
      exit
    }
  ' "$file"
}

validate_markdown_budget() {
  local path="$1"
  local label="$2"
  local length
  length="$(wc -m <"$path" | tr -d ' ')"
  if [[ "$length" -gt 12000 ]]; then
    hg_die "$label exceeds Antigravity 12,000 character limit: $path ($length chars)"
  fi
}

antigravity_target_owner() {
  if [[ -e "$ANTIGRAVITY_DIR" ]]; then
    stat -c '%U' "$ANTIGRAVITY_DIR"
    return 0
  fi
  stat -c '%U' "$(dirname "$ANTIGRAVITY_DIR")"
}

antigravity_target_home() {
  local owner="$1"
  getent passwd "$owner" | cut -d: -f6
}

run_antigravity_cli() {
  local owner owner_home
  owner="$(antigravity_target_owner)"
  [[ "$owner" != "root" ]] || return 125
  owner_home="$(antigravity_target_home "$owner")"
  [[ -n "$owner_home" ]] || hg_die "Failed to resolve home directory for Antigravity owner: $owner"

  if [[ "$(id -un)" == "$owner" ]]; then
    HOME="$owner_home" /opt/Antigravity/bin/antigravity "$@"
    return 0
  fi

  runuser -u "$owner" -- env HOME="$owner_home" /opt/Antigravity/bin/antigravity "$@"
}

ensure_repo_checkout() {
  local repo_url="$1"
  local ref="$2"
  local target="$3"
  local label="$4"

  if [[ ! -d "$target/.git" ]]; then
    pending=1
    case "$MODE" in
      write)
        mkdir -p "$(dirname "$target")"
        git clone "$repo_url" "$target" >/dev/null
        hg_ok "Cloned $label: $target"
        ;;
      dry-run)
        hg_warn "Would clone $label: $repo_url -> $target"
        return 0
        ;;
      check)
        hg_warn "Missing $label clone: $target"
        return 1
        ;;
    esac
  fi

  if [[ "$MODE" == "write" ]]; then
    git -C "$target" fetch --tags origin >/dev/null 2>&1 || true
    git -C "$target" checkout --detach "$ref" >/dev/null 2>&1
  else
    local head=""
    local expected=""
    head="$(git -C "$target" rev-parse HEAD 2>/dev/null || true)"
    expected="$(git -C "$target" rev-parse "${ref}^{commit}" 2>/dev/null || true)"
    if [[ -n "$head" && -n "$expected" && "$head" != "$expected" ]]; then
      pending=1
      if [[ "$MODE" == "dry-run" ]]; then
        hg_warn "Would checkout $label to $ref: $target"
      else
        hg_warn "Unexpected $label ref in check mode: $target (have $head want $expected from $ref)"
      fi
    fi
  fi
}

download_asset() {
  local url="$1"
  local target="$2"
  local label="$3"

  if [[ -f "$target" ]]; then
    return 0
  fi

  pending=1
  case "$MODE" in
    write)
      mkdir -p "$(dirname "$target")"
      curl -fsSL "$url" -o "$target"
      hg_ok "Cached $label: $target"
      ;;
    dry-run)
      hg_warn "Would download $label: $url"
      ;;
    check)
      hg_warn "Missing $label cache: $target"
      ;;
  esac
}

sync_global_workflow_file() {
  local name="$1"
  local staged="$2"
  local target="$GLOBAL_WORKFLOWS_DIR/$name"
  managed_global_workflow_set["$name"]=1
  validate_markdown_budget "$staged" "Antigravity global workflow"
  sync_rendered_file "$target" "$staged" "Synced Antigravity global workflow"
}

build_addyosmani_workflow() {
  local repo_dir="$1"
  local command_name="$2"
  local source_file="$repo_dir/.claude/commands/$command_name.md"
  local staged="$tmpdir/$command_name.md"
  local description
  description="$(extract_frontmatter_scalar "$source_file" "description")"

  {
    printf '# /%s\n\n' "$command_name"
    printf '%s\n\n' "$description"
    printf '## Source\n\n'
    printf -- '- Library: `addyosmani/agent-skills@%s`\n' "$(jq -r '.libraries.workflows[] | select(.id == "addyosmani-agent-skills") | .ref' "$SOURCES_MANIFEST")"
    printf -- '- Command: `%s`\n\n' "$source_file"
    printf '## Workflow\n\n'
    strip_frontmatter "$source_file"
    printf '\n## Notes\n\n'
    printf -- '- Keep the workflow scoped to the current repo and confirm before destructive changes.\n'
  } >"$staged"

  sync_global_workflow_file "$command_name.md" "$staged"
}

build_maintenance_workflows() {
  local staged

  staged="$tmpdir/antigravity-sync.md"
  {
    printf '# /antigravity-sync\n\n'
    printf 'Refresh the managed Antigravity overlays for the local `hairglasses-studio` workstation.\n\n'
    printf '## Steps\n\n'
    printf '1. Verify the workspace root is `/home/hg/hairglasses-studio`.\n'
    printf '2. Run `/home/hg/hairglasses-studio/dotfiles/scripts/hg-antigravity-sync.sh` from the terminal.\n'
    printf '3. Review the generated overlay status in `~/.gemini/antigravity/.hg-antigravity-sync.json`.\n'
    printf '4. If extensions or sidecars changed, review `~/.gemini/antigravity/.hg-antigravity-ecosystem.json`.\n'
  } >"$staged"
  sync_global_workflow_file "antigravity-sync.md" "$staged"

  staged="$tmpdir/skills-activate.md"
  {
    printf '# /skills-activate\n\n'
    printf 'Refresh the curated live Antigravity global skills and workflow subset from the pinned ecosystem manifest.\n\n'
    printf '## Steps\n\n'
    printf '1. Update `dotfiles/config/antigravity/sources.json` if the live subset should change.\n'
    printf '2. Run `/home/hg/hairglasses-studio/dotfiles/scripts/hg-antigravity-ecosystem-sync.sh`.\n'
    printf '3. Confirm the live skills directory `~/.gemini/antigravity/skills/` only contains the intended subset.\n'
    printf '4. Confirm the workflow library and archive directories still point at the pinned refs.\n'
  } >"$staged"
  sync_global_workflow_file "skills-activate.md" "$staged"
}

activate_live_skills() {
  local skill_repo_dir="$1"
  local live_skill
  while IFS= read -r live_skill; do
    [[ -n "$live_skill" ]] || continue
    local source_dir="$skill_repo_dir/skills/$live_skill"
    [[ -d "$source_dir" ]] || hg_die "Missing live skill in source library: $source_dir"
    managed_global_skill_set["$live_skill"]=1
    sync_symlink "$GLOBAL_SKILLS_DIR/$live_skill" "$source_dir" "Synced Antigravity global skill"
  done < <(jq -r '.libraries.skills[] | select(.id == "sickn33-antigravity-awesome-skills") | .live_skill_ids[]' "$SOURCES_MANIFEST")
}

sync_global_skills_and_workflows() {
  local skill_repo_ref workflow_repo_ref
  local skill_repo_dir="$SKILLS_LIBRARY_DIR/sickn33-antigravity-awesome-skills"
  local agentmemory_repo_dir="$SKILLS_LIBRARY_DIR/webzler-agentmemory"
  local addyosmani_repo_dir="$WORKFLOW_LIBRARY_DIR/addyosmani-agent-skills"
  local harik_repo_dir="$WORKFLOW_LIBRARY_DIR/harikrishna8121999-antigravity-workflows"

  skill_repo_ref="$(jq -r '.libraries.skills[] | select(.id == "sickn33-antigravity-awesome-skills") | .ref' "$SOURCES_MANIFEST")"
  workflow_repo_ref="$(jq -r '.libraries.workflows[] | select(.id == "addyosmani-agent-skills") | .ref' "$SOURCES_MANIFEST")"

  ensure_repo_checkout \
    "$(jq -r '.libraries.skills[] | select(.id == "sickn33-antigravity-awesome-skills") | .repo' "$SOURCES_MANIFEST")" \
    "$skill_repo_ref" \
    "$skill_repo_dir" \
    "sickn33 skill library"
  archived_skill_source_count=$((archived_skill_source_count + 1))

  ensure_repo_checkout \
    "$(jq -r '.libraries.skills[] | select(.id == "webzler-agentmemory") | .repo' "$SOURCES_MANIFEST")" \
    "$(jq -r '.libraries.skills[] | select(.id == "webzler-agentmemory") | .ref' "$SOURCES_MANIFEST")" \
    "$agentmemory_repo_dir" \
    "agentMemory library"
  archived_skill_source_count=$((archived_skill_source_count + 1))

  ensure_repo_checkout \
    "$(jq -r '.libraries.workflows[] | select(.id == "addyosmani-agent-skills") | .repo' "$SOURCES_MANIFEST")" \
    "$workflow_repo_ref" \
    "$addyosmani_repo_dir" \
    "addyosmani workflow library"
  archived_workflow_source_count=$((archived_workflow_source_count + 1))

  ensure_repo_checkout \
    "$(jq -r '.libraries.workflows[] | select(.id == "harikrishna8121999-antigravity-workflows") | .repo' "$SOURCES_MANIFEST")" \
    "$(jq -r '.libraries.workflows[] | select(.id == "harikrishna8121999-antigravity-workflows") | .ref' "$SOURCES_MANIFEST")" \
    "$harik_repo_dir" \
    "community workflow library"
  archived_workflow_source_count=$((archived_workflow_source_count + 1))

  activate_live_skills "$skill_repo_dir"

  while IFS= read -r command_name; do
    [[ -n "$command_name" ]] || continue
    build_addyosmani_workflow "$addyosmani_repo_dir" "$command_name"
  done < <(jq -r '.libraries.workflows[] | select(.id == "addyosmani-agent-skills") | .global_workflows[]' "$SOURCES_MANIFEST")

  build_maintenance_workflows
}

sync_cached_extensions_and_sidecars() {
  local panel_url panel_vsix_path panel_ext panel_version antigravity_owner
  panel_ext="$(jq -r '.extensions[] | select(.id == "n2ns.antigravity-panel") | .id' "$SOURCES_MANIFEST")"
  panel_version="$(jq -r '.extensions[] | select(.id == "n2ns.antigravity-panel") | .version' "$SOURCES_MANIFEST")"
  panel_url="$(jq -r '.extensions[] | select(.id == "n2ns.antigravity-panel") | .vsix_url' "$SOURCES_MANIFEST")"
  panel_vsix_path="$TOOL_CACHE_DIR/${panel_ext}-${panel_version}.vsix"
  antigravity_owner="$(antigravity_target_owner)"
  download_asset "$panel_url" "$panel_vsix_path" "Antigravity Panel VSIX"
  managed_tool_set["$panel_vsix_path"]=1

  if [[ "$antigravity_owner" != "root" ]]; then
    local installed_panel=""
    installed_panel="$(
      { run_antigravity_cli --list-extensions --show-versions 2>/dev/null || true; } \
        | awk '$0 ~ /^n2ns\.antigravity-panel@/ {print $0; exit}'
    )"
    if [[ "$installed_panel" != "n2ns.antigravity-panel@$panel_version" ]]; then
      pending=1
      case "$MODE" in
        write)
          run_antigravity_cli --install-extension "$panel_vsix_path" >/dev/null
          hg_ok "Installed Antigravity extension: n2ns.antigravity-panel@$panel_version"
          ;;
        dry-run)
          hg_warn "Would install Antigravity extension: n2ns.antigravity-panel@$panel_version"
          ;;
        check)
          hg_warn "Missing Antigravity extension: n2ns.antigravity-panel@$panel_version"
          ;;
      esac
    fi
  fi

  local manager_asset_name manager_asset_url manager_appimage
  manager_asset_name="$(jq -r '.sidecars[] | select(.id == "draculabo-antigravity-manager") | .asset_name' "$SOURCES_MANIFEST")"
  manager_asset_url="$(jq -r '.sidecars[] | select(.id == "draculabo-antigravity-manager") | .asset_url' "$SOURCES_MANIFEST")"
  manager_appimage="$MANAGER_DIR/$manager_asset_name"
  download_asset "$manager_asset_url" "$manager_appimage" "Antigravity Manager AppImage"
  managed_tool_set["$manager_appimage"]=1
  if [[ "$MODE" == "write" && -f "$manager_appimage" ]]; then
    chmod 755 "$manager_appimage"
  fi

  local manager_launcher="$tmpdir/antigravity-manager-launch"
  {
    printf '#!/usr/bin/env bash\n'
    printf 'set -euo pipefail\n'
    printf 'exec %q "$@"\n' "$manager_appimage"
  } >"$manager_launcher"
  chmod 755 "$manager_launcher"
  sync_rendered_file "$MANAGER_LAUNCHER_PATH" "$manager_launcher" "Synced Antigravity Manager launcher"

  local manager_desktop="$tmpdir/antigravity-manager.desktop"
  {
    printf '[Desktop Entry]\n'
    printf 'Name=Antigravity Manager\n'
    printf 'Comment=Multi-account manager for Antigravity\n'
    printf 'Exec=%s\n' "$MANAGER_LAUNCHER_PATH"
    printf 'Icon=antigravity\n'
    printf 'Type=Application\n'
    printf 'Categories=Development;IDE;\n'
  } >"$manager_desktop"
  sync_rendered_file "$MANAGER_DESKTOP_PATH" "$manager_desktop" "Synced Antigravity Manager desktop entry"
  sidecar_count=$((sidecar_count + 1))

  local ssh_proxy_name ssh_proxy_url ssh_proxy_vsix
  ssh_proxy_name="$(jq -r '.sidecars[] | select(.id == "dinobot22-antigravity-ssh-proxy") | .asset_name' "$SOURCES_MANIFEST")"
  ssh_proxy_url="$(jq -r '.sidecars[] | select(.id == "dinobot22-antigravity-ssh-proxy") | .asset_url' "$SOURCES_MANIFEST")"
  ssh_proxy_vsix="$TOOL_CACHE_DIR/$ssh_proxy_name"
  download_asset "$ssh_proxy_url" "$ssh_proxy_vsix" "Antigravity SSH Proxy VSIX"
  managed_tool_set["$ssh_proxy_vsix"]=1
  local ssh_proxy_installer="$tmpdir/antigravity-install-ssh-proxy"
  {
    printf '#!/usr/bin/env bash\n'
    printf 'set -euo pipefail\n'
    printf 'exec /opt/Antigravity/bin/antigravity --install-extension %q\n' "$ssh_proxy_vsix"
  } >"$ssh_proxy_installer"
  chmod 755 "$ssh_proxy_installer"
  sync_rendered_file "$SSH_PROXY_INSTALLER_PATH" "$ssh_proxy_installer" "Synced Antigravity SSH proxy installer"
  sidecar_count=$((sidecar_count + 1))
}

build_agentmemory_vsix() {
  local repo_dir="$SKILLS_LIBRARY_DIR/webzler-agentmemory"
  local version vsix_path tmp_vsix installer
  version="$(jq -r '.version' "$repo_dir/package.json")"
  vsix_path="$TOOL_CACHE_DIR/webzler.agentmemory-$version.vsix"
  managed_tool_set["$vsix_path"]=1

  if [[ ! -f "$vsix_path" ]]; then
    pending=1
    case "$MODE" in
      write)
        (
          cd "$repo_dir"
          npm install --no-fund --no-audit >/dev/null
          npm run compile >/dev/null
          npx @vscode/vsce package -o "$vsix_path" >/dev/null
        )
        hg_ok "Built agentMemory VSIX: $vsix_path"
        ;;
      dry-run)
        hg_warn "Would build agentMemory VSIX: $vsix_path"
        ;;
      check)
        hg_warn "Missing agentMemory VSIX: $vsix_path"
        ;;
    esac
  fi

  installer="$tmpdir/antigravity-install-agentmemory"
  {
    printf '#!/usr/bin/env bash\n'
    printf 'set -euo pipefail\n'
    printf 'exec /opt/Antigravity/bin/antigravity --install-extension %q\n' "$vsix_path"
  } >"$installer"
  chmod 755 "$installer"
  sync_rendered_file "$AGENTMEMORY_INSTALLER_PATH" "$installer" "Synced agentMemory installer"

  tmp_vsix="$tmpdir/agentmemory.disabled.json"
  {
    printf '{\n'
    printf '  "mcpServers": {\n'
    printf '    "agentmemory": {\n'
    printf '      "command": "node",\n'
    printf '      "args": [\n'
    printf '        "%s/out/mcp-server/server.js",\n' "$repo_dir"
    printf '        "<project_id>",\n'
    printf '        "<absolute_workspace_path>"\n'
    printf '      ],\n'
    printf '      "disabled": true\n'
    printf '    }\n'
    printf '  }\n'
    printf '}\n'
  } >"$tmp_vsix"
  sync_rendered_file "$AGENTMEMORY_TEMPLATE_PATH" "$tmp_vsix" "Synced agentMemory MCP template"
  sidecar_count=$((sidecar_count + 1))
}

remove_stale_managed_paths() {
  if [[ -f "$METADATA_PATH" ]]; then
    while IFS= read -r old_name; do
      [[ -n "$old_name" ]] || continue
      [[ -n "${managed_global_skill_set[$old_name]:-}" ]] && continue
      remove_path_if_present "$GLOBAL_SKILLS_DIR/$old_name" "Antigravity global skill"
    done < <(jq -r '.managed_global_skills[]? // empty' "$METADATA_PATH" 2>/dev/null || true)

    while IFS= read -r old_name; do
      [[ -n "$old_name" ]] || continue
      [[ -n "${managed_global_workflow_set[$old_name]:-}" ]] && continue
      remove_path_if_present "$GLOBAL_WORKFLOWS_DIR/$old_name" "Antigravity global workflow"
    done < <(jq -r '.managed_global_workflows[]? // empty' "$METADATA_PATH" 2>/dev/null || true)
  fi
}

write_metadata() {
  local managed_skills_file="$tmpdir/managed-global-skills.txt"
  local managed_workflows_file="$tmpdir/managed-global-workflows.txt"
  local installed_extensions
  local sickn33_ref webzler_ref addyosmani_ref harik_ref
  local manager_asset_path ssh_proxy_asset_path agentmemory_version agentmemory_vsix_path
  printf '%s\n' "${!managed_global_skill_set[@]}" | sort -u >"$managed_skills_file"
  printf '%s\n' "${!managed_global_workflow_set[@]}" | sort -u >"$managed_workflows_file"

  live_global_skill_count="$(wc -l <"$managed_skills_file" | tr -d ' ')"
  global_workflow_count="$(wc -l <"$managed_workflows_file" | tr -d ' ')"
  if [[ "$(antigravity_target_owner)" == "root" ]]; then
    installed_extensions=""
  else
    installed_extensions="$(
      { run_antigravity_cli --list-extensions --show-versions 2>/dev/null || true; } \
        | awk '$0 ~ /^[A-Za-z0-9._-]+@/ {print}'
    )"
  fi
  installed_extension_count="$(printf '%s\n' "$installed_extensions" | sed '/^$/d' | wc -l | tr -d ' ')"
  sickn33_ref="$(jq -r '.libraries.skills[] | select(.id == "sickn33-antigravity-awesome-skills") | .ref' "$SOURCES_MANIFEST")"
  webzler_ref="$(jq -r '.libraries.skills[] | select(.id == "webzler-agentmemory") | .ref' "$SOURCES_MANIFEST")"
  addyosmani_ref="$(jq -r '.libraries.workflows[] | select(.id == "addyosmani-agent-skills") | .ref' "$SOURCES_MANIFEST")"
  harik_ref="$(jq -r '.libraries.workflows[] | select(.id == "harikrishna8121999-antigravity-workflows") | .ref' "$SOURCES_MANIFEST")"
  manager_asset_path="$MANAGER_DIR/$(jq -r '.sidecars[] | select(.id == "draculabo-antigravity-manager") | .asset_name' "$SOURCES_MANIFEST")"
  ssh_proxy_asset_path="$TOOL_CACHE_DIR/$(jq -r '.sidecars[] | select(.id == "dinobot22-antigravity-ssh-proxy") | .asset_name' "$SOURCES_MANIFEST")"
  agentmemory_version="$(jq -r '.version // "0.1.0"' "$SKILLS_LIBRARY_DIR/webzler-agentmemory/package.json" 2>/dev/null || printf '0.1.0')"
  agentmemory_vsix_path="$TOOL_CACHE_DIR/webzler.agentmemory-$agentmemory_version.vsix"

  {
    printf '{\n'
    printf '  "generator": "dotfiles/scripts/hg-antigravity-ecosystem-sync.sh",\n'
    printf '  "workspace_root": %s,\n' "$(jq -Rn --arg v "$WORKSPACE_ROOT" '$v')"
    printf '  "sources_manifest": %s,\n' "$(jq -Rn --arg v "$SOURCES_MANIFEST" '$v')"
    printf '  "generated_on": %s,\n' "$(jq -Rn --arg v "$(date +%Y-%m-%d)" '$v')"
    printf '  "live_global_skill_count": %s,\n' "$live_global_skill_count"
    printf '  "global_workflow_count": %s,\n' "$global_workflow_count"
    printf '  "archived_skill_source_count": %s,\n' "$archived_skill_source_count"
    printf '  "archived_workflow_source_count": %s,\n' "$archived_workflow_source_count"
    printf '  "sidecar_count": %s,\n' "$sidecar_count"
    printf '  "installed_extension_count": %s,\n' "$installed_extension_count"
    printf '  "env_bridge_mode": "runtime-resolved",\n'
    printf '  "managed_global_skills": %s,\n' "$(jq -Rsc 'split("\n") | map(select(length > 0))' <"$managed_skills_file")"
    printf '  "managed_global_workflows": %s,\n' "$(jq -Rsc 'split("\n") | map(select(length > 0))' <"$managed_workflows_file")"
    printf '  "installed_extensions": %s,\n' "$(printf '%s\n' "$installed_extensions" | jq -Rsc 'split("\n") | map(select(length > 0))')"
    printf '  "skill_library_refs": {\n'
    printf '    "sickn33-antigravity-awesome-skills": %s,\n' "$(jq -Rn --arg v "$sickn33_ref" '$v')"
    printf '    "webzler-agentmemory": %s\n' "$(jq -Rn --arg v "$webzler_ref" '$v')"
    printf '  },\n'
    printf '  "workflow_library_refs": {\n'
    printf '    "addyosmani-agent-skills": %s,\n' "$(jq -Rn --arg v "$addyosmani_ref" '$v')"
    printf '    "harikrishna8121999-antigravity-workflows": %s\n' "$(jq -Rn --arg v "$harik_ref" '$v')"
    printf '  },\n'
    printf '  "manual_sidecars": {\n'
    printf '    "antigravity_manager_appimage": %s,\n' "$(jq -Rn --arg v "$manager_asset_path" '$v')"
    printf '    "ssh_proxy_vsix": %s,\n' "$(jq -Rn --arg v "$ssh_proxy_asset_path" '$v')"
    printf '    "agentmemory_vsix": %s\n' "$(jq -Rn --arg v "$agentmemory_vsix_path" '$v')"
    printf '  }\n'
    printf '}\n'
  } >"$tmpdir/antigravity-ecosystem.json"

  sync_rendered_file "$METADATA_PATH" "$tmpdir/antigravity-ecosystem.json" "Synced Antigravity ecosystem metadata"
}

sync_global_skills_and_workflows
sync_cached_extensions_and_sidecars
build_agentmemory_vsix
remove_stale_managed_paths
write_metadata

case "$MODE" in
  write)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "Antigravity ecosystem sync already up to date"
    else
      hg_info "Antigravity ecosystem sync refreshed"
    fi
    ;;
  dry-run)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "No Antigravity ecosystem changes needed"
    fi
    ;;
  check)
    if [[ "$pending" -eq 0 ]]; then
      hg_ok "Antigravity ecosystem sync up to date"
    else
      exit 1
    fi
    ;;
esac
