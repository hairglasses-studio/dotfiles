#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

NAME=""
SEED_REPO="codexkit"
VISIBILITY="private"
DESCRIPTION=""
DRY_RUN=false
SKIP_REMOTE=false
SKIP_DOCS=false

usage() {
  cat <<'EOF'
Usage: hg-providerkit-bootstrap.sh <name> [options]

Bootstrap a new provider toolkit repo from an existing baseline repo, then
register it in the studio workspace and docs control plane.

Options:
  --seed-from <repo>     Baseline repo to copy (default: codexkit)
  --visibility <value>   GitHub visibility for remote creation (default: private)
  --description <text>   Optional repo description for gh repo create
  --skip-remote          Do not create or push a GitHub remote
  --skip-docs            Do not refresh docs architecture after bootstrapping
  --dry-run              Print planned actions without writing
  -h, --help             Show this help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --seed-from)
      [[ $# -ge 2 ]] || hg_die "--seed-from requires a repo name"
      SEED_REPO="$2"
      shift 2
      ;;
    --visibility)
      [[ $# -ge 2 ]] || hg_die "--visibility requires a value"
      VISIBILITY="$2"
      shift 2
      ;;
    --description)
      [[ $# -ge 2 ]] || hg_die "--description requires text"
      DESCRIPTION="$2"
      shift 2
      ;;
    --skip-remote)
      SKIP_REMOTE=true
      shift
      ;;
    --skip-docs)
      SKIP_DOCS=true
      shift
      ;;
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      hg_die "Unknown argument: $1"
      ;;
    *)
      [[ -z "$NAME" ]] || hg_die "Only one repo name may be provided"
      NAME="$1"
      shift
      ;;
  esac
done

[[ -n "$NAME" ]] || {
  usage >&2
  exit 1
}

hg_require git jq rg perl rsync
$SKIP_REMOTE || hg_require gh

REPO_DIR="$HG_STUDIO_ROOT/$NAME"
SEED_DIR="$HG_STUDIO_ROOT/$SEED_REPO"
MANIFEST_PATH="$(hg_workspace_manifest)"
GO_WORK_PATH="$HG_STUDIO_ROOT/go.work"
DOCS_DIR="$HG_STUDIO_ROOT/docs"
SKILL_OLD="${SEED_REPO//-/_}"
SKILL_NEW="${NAME//-/_}"

[[ -d "$SEED_DIR/.git" ]] || hg_die "Seed repo not found: $SEED_DIR"
[[ ! -e "$REPO_DIR" ]] || hg_die "Target already exists: $REPO_DIR"
[[ "$VISIBILITY" == "private" || "$VISIBILITY" == "public" || "$VISIBILITY" == "internal" ]] ||   hg_die "--visibility must be private, public, or internal"

run_cmd() {
  if $DRY_RUN; then
    hg_warn "Would run: $*"
  else
    "$@"
  fi
}

copy_seed_repo() {
  run_cmd mkdir -p "$REPO_DIR"
  run_cmd rsync -a     --exclude .git     --exclude .DS_Store     --exclude .gemini/.hg-gemini-settings-sync.json     --exclude .claude/settings.local.json     "$SEED_DIR/" "$REPO_DIR/"
}

rename_seed_paths() {
  local old
  local new
  for old in "$REPO_DIR/cmd/$SEED_REPO" "$REPO_DIR/cmd/${SEED_REPO}-mcp"; do
    [[ -d "$old" ]] || continue
    new="${old/$SEED_REPO/$NAME}"
    run_cmd mv "$old" "$new"
  done

  if [[ -d "$REPO_DIR/.agents/skills/$SKILL_OLD" ]]; then
    run_cmd mv "$REPO_DIR/.agents/skills/$SKILL_OLD" "$REPO_DIR/.agents/skills/$SKILL_NEW"
  fi
}

rewrite_seed_strings() {
  local file
  while IFS= read -r file; do
    $DRY_RUN && continue
    perl -0pi -e "s#github.com/hairglasses-studio/$SEED_REPO#github.com/hairglasses-studio/$NAME#g; s#$SEED_REPO#$NAME#g; s#$SKILL_OLD#$SKILL_NEW#g" "$file"
  done < <(
    rg -l --hidden --follow       --glob '!.git/**'       --glob '!bin/**'       --glob '!node_modules/**'       --glob '!dist/**'       --glob '!build/**'       --glob '!vendor/**'       --glob '!*.png'       --glob '!*.jpg'       --glob '!*.jpeg'       --glob '!*.gif'       --glob '!*.ico'       --glob '!*.pdf'       --glob '!*.woff'       --glob '!*.woff2'       --glob '!*.ttf'       --glob '!*.otf'       --glob '!*.sumdb'       "$SEED_REPO|$SKILL_OLD|github.com/hairglasses-studio/$SEED_REPO" "$REPO_DIR"
  )
}

ensure_git_repo() {
  if $DRY_RUN; then
    hg_warn "Would initialize git repo at $REPO_DIR"
    return 0
  fi
  git -C "$REPO_DIR" init -q -b main
}

update_manifest() {
  [[ -f "$MANIFEST_PATH" ]] || return 0
  if $DRY_RUN; then
    hg_warn "Would register $NAME in workspace/manifest.json"
    return 0
  fi

  local tmp
  tmp="$(mktemp)"
  jq     --arg name "$NAME"     --arg seed "$SEED_REPO"     --arg visibility "$VISIBILITY"     '
      if any(.repos[]; .name == $name) then
        .
      else
        (.repos | map(select(.name == $seed)) | first) as $seed_repo
        | .repos += [
            (
              if $seed_repo != null then
                $seed_repo
              else
                {
                  category: "application",
                  scope: "active_first_party",
                  language: "Go",
                  baseline_target: true,
                  go_work_member: true,
                  visibility: $visibility,
                  lifecycle: "canonical",
                  automation_profile: "application",
                  ci_profile: "go_standard",
                  review_profile: (if $visibility == "private" then "private_ai_review" else "public_ai_review" end),
                  mcp_surface_class: "none",
                  env_profile: "root_inherit",
                  mirror_of: null
                }
              end
            )
            | .name = $name
            | .visibility = $visibility
          ]
      end
    ' "$MANIFEST_PATH" > "$tmp"
  mv "$tmp" "$MANIFEST_PATH"
}

update_go_work() {
  [[ -f "$REPO_DIR/go.mod" ]] || return 0
  [[ -f "$GO_WORK_PATH" ]] || return 0
  [[ -f "$MANIFEST_PATH" ]] || return 0

  local member
  member="$(jq -r --arg repo "$NAME" 'first(.repos[] | select(.name == $repo) | .go_work_member) // false' "$MANIFEST_PATH")"
  [[ "$member" == "true" ]] || return 0

  if $DRY_RUN; then
    hg_warn "Would add ./$NAME to go.work"
    return 0
  fi
  (cd "$HG_STUDIO_ROOT" && go work use "./$NAME")
}

run_repo_syncs() {
  run_cmd "$SCRIPT_DIR/hg-onboard-repo.sh" "$REPO_DIR" --language=go
  run_cmd "$SCRIPT_DIR/hg-provider-role-sync.sh" "$REPO_DIR"
  run_cmd "$SCRIPT_DIR/hg-gemini-settings-sync.sh" "$REPO_DIR" --allow-dirty
}

refresh_docs() {
  $SKIP_DOCS && return 0
  [[ -d "$DOCS_DIR/.git" ]] || return 0
  if $DRY_RUN; then
    hg_warn "Would refresh docs repo architecture"
    return 0
  fi
  (cd "$DOCS_DIR" && ./scripts/refresh-repo-architecture.sh)
}

commit_bootstrap() {
  if $DRY_RUN; then
    hg_warn "Would create initial bootstrap commit in $REPO_DIR"
    return 0
  fi
  git -C "$REPO_DIR" add .
  if git -C "$REPO_DIR" diff --cached --quiet; then
    hg_warn "$NAME: no bootstrap changes staged"
    return 0
  fi
  git -C "$REPO_DIR" commit -m "chore: bootstrap $NAME from $SEED_REPO"
}

ensure_remote() {
  $SKIP_REMOTE && return 0
  local full_name="hairglasses-studio/$NAME"
  if $DRY_RUN; then
    hg_warn "Would ensure GitHub remote for $full_name"
    return 0
  fi

  if gh repo view "$full_name" >/dev/null 2>&1; then
    hg_info "$full_name already exists on GitHub"
    if ! git -C "$REPO_DIR" remote get-url origin >/dev/null 2>&1; then
      git -C "$REPO_DIR" remote add origin "git@github.com:$full_name.git" ||         git -C "$REPO_DIR" remote add origin "https://github.com/$full_name.git"
    fi
    git -C "$REPO_DIR" push -u origin main
    return 0
  fi

  local args=(repo create "$full_name" "--$VISIBILITY" --source "$REPO_DIR" --remote origin --push)
  if [[ -n "$DESCRIPTION" ]]; then
    args+=(--description "$DESCRIPTION")
  fi
  gh "${args[@]}"
}

hg_info "Bootstrapping provider toolkit $NAME from $SEED_REPO"
copy_seed_repo
rename_seed_paths
rewrite_seed_strings
ensure_git_repo
update_manifest
update_go_work
run_repo_syncs
refresh_docs
commit_bootstrap
ensure_remote
hg_ok "$NAME bootstrap complete"
