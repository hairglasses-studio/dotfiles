#!/usr/bin/env bash
# hg-workspace-check.sh — Validate manifest-backed workspace inventory and shared script drift.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

hg_workspace_require_manifest

MANIFEST="$(hg_workspace_manifest)"

hg_require jq git

jq -e '
  .version == 2
  and (.repos | type == "array")
  and (([.repos[].name] | length) == ([.repos[].name] | unique | length))
' "$MANIFEST" >/dev/null || hg_die "workspace manifest must be version=2 with unique repo names"

jq -e '
  [.repos[] as $repo
   | select(
       (["public", "private", "never_publish"] | index($repo.visibility // "")) == null
       or (["canonical", "mirror", "compatibility", "deprecated"] | index($repo.lifecycle // "")) == null
       or (["hub", "application", "standalone_mcp", "tooling", "private_ops"] | index($repo.automation_profile // "")) == null
       or (["go_standard", "go_custom", "python_standard", "node_standard", "none"] | index($repo.ci_profile // "")) == null
       or (["public_ai_review", "private_ai_review", "none"] | index($repo.review_profile // "")) == null
       or (["none", "small", "large", "gateway"] | index($repo.mcp_surface_class // "")) == null
       or (["root_inherit", "root_inherit_oauth_exception", "local_only"] | index($repo.env_profile // "")) == null
     )
  ] | length == 0
' "$MANIFEST" >/dev/null || hg_die "workspace manifest contains invalid enum values"

while IFS= read -r repo; do
  [[ -d "$HG_STUDIO_ROOT/$repo" ]] || hg_die "manifest repo missing on disk: $repo"
done < <(hg_workspace_repo_names)

while IFS= read -r repo; do
  [[ -n "$repo" ]] || continue
  mirror_of="$(hg_workspace_repo_field "$repo" "mirror_of" "")"
  [[ -n "$mirror_of" ]] || hg_die "mirror repo missing mirror_of: $repo"
  [[ -d "$HG_STUDIO_ROOT/$mirror_of" ]] || hg_die "mirror source missing for $repo: $mirror_of"
done < <(hg_workspace_repo_names '.repos[] | select((.lifecycle // "") == "mirror")')

go_work_entries="$(awk '
  BEGIN { in_use = 0 }
  /^use[[:space:]]*\(/ { in_use = 1; next }
  in_use && /^\)/ { exit }
  in_use {
    gsub(/^[[:space:]]+|[[:space:]]+$/, "", $0)
    sub(/^\.\//, "", $0)
    if ($0 != "") print $0
  }
' "$HG_STUDIO_ROOT/go.work" | sort)"

manifest_go_entries="$(jq -r '.repos[] | select(.go_work_member == true) | .name' "$MANIFEST" | sort)"

missing_from_manifest="$(comm -23 <(printf '%s\n' "$go_work_entries") <(printf '%s\n' "$manifest_go_entries"))"
missing_from_go_work="$(comm -13 <(printf '%s\n' "$go_work_entries") <(printf '%s\n' "$manifest_go_entries"))"

[[ -z "$missing_from_manifest" ]] || hg_die "go.work contains repos missing from manifest: $(paste -sd ', ' <<<"$missing_from_manifest")"
[[ -z "$missing_from_go_work" ]] || hg_die "manifest go_work_member=true repos missing from go.work: $(paste -sd ', ' <<<"$missing_from_go_work")"

for script in \
  "$SCRIPT_DIR/hg-go-sync.sh" \
  "$SCRIPT_DIR/hg-new-repo.sh" \
  "$SCRIPT_DIR/hg-onboard-repo.sh" \
  "$SCRIPT_DIR/hg-repo-profile-sync.sh" \
  "$SCRIPT_DIR/hg-workflow-sync.sh" \
  "$SCRIPT_DIR/sync-standalone-mcp-repos.sh"; do
  grep -q 'source "\$SCRIPT_DIR/lib/hg-workspace.sh"' "$script" || hg_die "shared script must use manifest helper: $script"
done

hg_ok "Workspace manifest and shared automation checks passed"
