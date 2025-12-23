#!/usr/bin/env bash

# AFTRS-Shell installer: clones org repos, scans local tool tree, wires completions, seeds configs, and sets up oh-my-posh.
set -euo pipefail

# Defaults (override via env)
ORG="${AFTRS_SHELL_ORG:-aftrs-shell}"
SKIP_REPO="${ORG}/aftrs-shell"
CLONE_DIR="${AFTRS_SHELL_HOME:-$HOME/Docs}"
LOCAL_ROOT="${AFTRS_LOCAL_ROOT:-/home/hg/Docs/aftrs-shell}"
COMPL_BASE="${AFTRS_COMPLETIONS_DIR:-$HOME/.config/aftrs/completions}"
ZSH_SITE_FUNCS="${ZSH_SITE_FUNCS:-$HOME/.local/share/zsh/site-functions}"
BASH_USER_COMPLETIONS="${BASH_USER_COMPLETIONS:-$HOME/.local/share/bash-completion/completions}"
FISH_COMPLETIONS="${FISH_COMPLETIONS:-$HOME/.config/fish/completions}"
OMP_CONFIG_DIR="${OMP_CONFIG_DIR:-$HOME/.config/oh-my-posh}"
OMP_THEME="${OMP_CONFIG_DIR}/aftrs-powerline.omp.json"

# Optional pretty output with gum (disabled on wizardstower)
HOST="$(hostname -s || echo unknown)"
USE_GUM=false
if command -v gum >/dev/null 2>&1 && [ "$HOST" != "wizardstower" ]; then USE_GUM=true; fi
log()   { if $USE_GUM; then gum log --level info "$@"; else echo "[INFO] $*"; fi; }
warn()  { if $USE_GUM; then gum log --level warn "$@"; else echo "[WARN] $*" >&2; fi; }
error() { if $USE_GUM; then gum log --level error "$@"; else echo "[ERROR] $*" >&2; fi; }

dry_run=false
run_install=true
skip_clone=false
refresh_completions=true

usage() {
  cat <<'USAGE'
Usage: install.sh [OPTIONS]

Install/update aftrs-shell tools (from GitHub org and local tree), set up shell completions, seed configs, and enable an oh-my-posh theme.

Options:
  -n, --dry-run        List actions without changing anything
      --no-install     Do not run per-repo install.sh scripts
      --skip-clone     Skip cloning/updating org repos (only operate on LOCAL_ROOT)
      --no-completions Skip completions aggregation/installation
  -h, --help           Show this help and exit
Env:
  AFTRS_SHELL_HOME        Destination for cloned repos (default: ~/Docs)
  AFTRS_LOCAL_ROOT        Local tools root to scan (default: /home/hg/Docs/aftrs-shell)
  AFTRS_COMPLETIONS_DIR   Base dir for collected completions (default: ~/.config/aftrs/completions)
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -n|--dry-run) dry_run=true ;;
    --no-install) run_install=false ;;
    --skip-clone) skip_clone=true ;;
    --no-completions) refresh_completions=false ;;
    -h|--help) usage; exit 0 ;;
    --) shift; break ;;
    *) error "Unknown option: $1"; usage >&2; exit 1 ;;
  esac
  shift
done

clone_org_repos() {
  mkdir -p "$CLONE_DIR"
  log "Fetching repositories for org: $ORG"
  if ! command -v gh >/dev/null 2>&1; then
    warn "gh CLI not found; skipping org clone. Install https://cli.github.com/ to enable."
    return 0
  fi
  local repos
  repos=$(SKIP_REPO="$SKIP_REPO" gh repo list "$ORG" --limit 1000 --json nameWithOwner --jq \
      '.[] | select(.nameWithOwner != env.SKIP_REPO) | .nameWithOwner')
  [ -z "$repos" ] && { warn "No repos found via gh for org $ORG"; return 0; }

  (cd "$CLONE_DIR"
    for repo in $repos; do
      local name="${repo#*/}"
      if [ -d "$name/.git" ]; then
        log "Updating $repo ..."
        $dry_run || (cd "$name" && git pull --ff-only)
      else
        log "Cloning $repo into $name ..."
        $dry_run || gh repo clone "$repo" "$name"
      fi
      if $run_install && [ -x "$name/install.sh" ]; then
        log "Running install.sh in $name"
        $dry_run || (cd "$name" && bash ./install.sh)
      fi
    done
  )
}

scan_local_repos_install() {
  if [ ! -d "$LOCAL_ROOT" ]; then
    warn "LOCAL_ROOT not found: $LOCAL_ROOT"
    return 0
  fi
  log "Scanning local tools in $LOCAL_ROOT"
  find "$LOCAL_ROOT" -maxdepth 2 -type f -name 'install.sh' -print0 2>/dev/null | while IFS= read -r -d '' f; do
    local dir; dir="$(dirname "$f")"
    if [[ "$dir" == *"/aftrs-shell/aftrs-shell" ]]; then continue; fi
    log "Running local install: $dir/install.sh"
    if $run_install; then
      if $dry_run; then echo "DRYRUN: (cd $dir && bash ./install.sh)"; else (cd "$dir" && bash ./install.sh); fi
    fi
  done
}

collect_and_install_completions() {
  log "Refreshing completions into $COMPL_BASE"
  $dry_run || mkdir -p "$COMPL_BASE"/{zsh,bash,fish} "$ZSH_SITE_FUNCS" "$BASH_USER_COMPLETIONS" "$FISH_COMPLETIONS"

  # zsh
  while IFS= read -r -d '' f; do
    base="$(basename "$f")"
    log "Found zsh completion: $f"
    if ! $dry_run; then
      cp -f "$f" "$COMPL_BASE/zsh/$base"
      ln -sf "$COMPL_BASE/zsh/$base" "$ZSH_SITE_FUNCS/$base"
    fi
  done < <(find "$LOCAL_ROOT" -type f \( -name '_*' -o -path '*/zsh/_*' -o -name '*.zsh' -a -path '*/etc/*' \) -print0 2>/dev/null || true)

  # bash
  while IFS= read -r -d '' f; do
    base="$(basename "$f")"
    log "Found bash completion: $f"
    if ! $dry_run; then
      cp -f "$f" "$COMPL_BASE/bash/$base"
      ln -sf "$COMPL_BASE/bash/$base" "$BASH_USER_COMPLETIONS/$base"
    fi
  done < <(find "$LOCAL_ROOT" -type f \( -name '*bash-completion*' -o -name '*completion.bash' -o -path '*/bash/*' -o -path '*/bash_completion.d/*' \) -print0 2>/dev/null || true)

  # fish
  while IFS= read -r -d '' f; do
    base="$(basename "$f")"
    log "Found fish completion: $f"
    if ! $dry_run; then
      cp -f "$f" "$COMPL_BASE/fish/$base"
      ln -sf "$COMPL_BASE/fish/$base" "$FISH_COMPLETIONS/$base"
    fi
  done < <(find "$LOCAL_ROOT" -type f \( -name '*.fish' -o -path '*/completions/*.fish' \) -print0 2>/dev/null || true)

  # Known locations
  if [ -f "$LOCAL_ROOT/nnn/misc/auto-completion/zsh/_nnn" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/nnn/misc/auto-completion/zsh/_nnn" "$ZSH_SITE_FUNCS/_nnn"
  fi
  if [ -f "$LOCAL_ROOT/nnn/misc/auto-completion/bash/nnn-completion.bash" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/nnn/misc/auto-completion/bash/nnn-completion.bash" "$BASH_USER_COMPLETIONS/nnn"
  fi
  if [ -f "$LOCAL_ROOT/nnn/misc/auto-completion/fish/nnn.fish" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/nnn/misc/auto-completion/fish/nnn.fish" "$FISH_COMPLETIONS/nnn.fish"
  fi
  if [ -f "$LOCAL_ROOT/mpv/etc/_mpv.zsh" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/mpv/etc/_mpv.zsh" "$ZSH_SITE_FUNCS/_mpv"
  fi
  if [ -f "$LOCAL_ROOT/mpv/etc/mpv.fish" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/mpv/etc/mpv.fish" "$FISH_COMPLETIONS/mpv.fish"
  fi
  if [ -f "$LOCAL_ROOT/zellij/zellij-utils/assets/completions/comp.fish" ]; then
    $dry_run || ln -sf "$LOCAL_ROOT/zellij/zellij-utils/assets/completions/comp.fish" "$FISH_COMPLETIONS/zellij.fish"
  fi

  # Generate from binaries if available
  if command -v zellij >/dev/null 2>&1 && ! $dry_run; then
    zellij setup --generate-completion zsh > "$ZSH_SITE_FUNCS/_zellij" 2>/dev/null || true
    zellij setup --generate-completion bash > "$BASH_USER_COMPLETIONS/zellij" 2>/dev/null || true
    zellij setup --generate-completion fish > "$FISH_COMPLETIONS/zellij.fish" 2>/dev/null || true
  fi

  # Wire shells
  if [ -n "${ZDOTDIR:-}" ] && [ -w "${ZDOTDIR}/.zshrc" ]; then ZSHRC="${ZDOTDIR}/.zshrc"; else ZSHRC="$HOME/.zshrc"; fi
  if ! grep -qs "$ZSH_SITE_FUNCS" "$ZSHRC" 2>/dev/null; then
    log "Wiring zsh fpath -> $ZSH_SITE_FUNCS"
    $dry_run || printf '\n# aftrs-shell completions\nfpath=(%s $fpath)\nautoload -Uz compinit && compinit\n' "$ZSH_SITE_FUNCS" >> "$ZSHRC"
  fi

  BASHRC="$HOME/.bashrc"
  if ! grep -qs "$BASH_USER_COMPLETIONS" "$BASHRC" 2>/dev/null; then
    log "Wiring bash completions -> $BASH_USER_COMPLETIONS"
    $dry_run || cat >> "$BASHRC" <<EOF

# aftrs-shell completions
if [ -f /usr/share/bash-completion/bash_completion ]; then
  . /usr/share/bash-completion/bash_completion
elif [ -f /etc/bash_completion ]; then
  . /etc/bash_completion
fi
export BASH_COMPLETION_USER_DIR="$BASH_USER_COMPLETIONS"
EOF
  fi
}

seed_configs() {
  local unpack="$LOCAL_ROOT/aftrs-shell/aftrs_initial_unpack.sh"
  if [ -x "$unpack" ]; then
    log "Seeding configs via $unpack"
    if $dry_run; then "$unpack" --dry-run; else "$unpack"; fi
  else
    warn "Unpack script not found at $unpack; skipping config seeding"
  fi
}

install_oh_my_posh() {
  if command -v oh-my-posh >/dev/null 2>&1; then
    log "oh-my-posh already installed"
  else
    log "Installing oh-my-posh"
    if $dry_run; then
      echo "DRYRUN: install oh-my-posh"
    else
      sudo wget -q https://github.com/JanDeDobbeleer/oh-my-posh/releases/latest/download/posh-linux-amd64 -O /usr/local/bin/oh-my-posh
      sudo chmod +x /usr/local/bin/oh-my-posh
    fi
  fi

  $dry_run || mkdir -p "$OMP_CONFIG_DIR"
  # Fallback default theme if seed wasn't applied
  if [ ! -f "$OMP_THEME" ] && ! $dry_run; then
    cat > "$OMP_THEME" <<'JSON'
{
  "$schema": "https://raw.githubusercontent.com/JanDeDobbeleer/oh-my-posh/main/themes/schema.json",
  "final_space": true, "version": 2,
  "blocks": [{ "type": "prompt", "alignment": "left", "segments": [
    { "type": "session", "style": "powerline", "foreground": "#1e222a", "background": "#61afef", "properties": { "display_host": true } },
    { "type": "path", "style": "powerline", "powerline_symbol": "\uE0B0", "foreground": "#1e222a", "background": "#98c379", "properties": { "style": "folder", "max_depth": 3, "enable_hyperlink": true } },
    { "type": "git", "style": "powerline", "foreground": "#1e222a", "background": "#e5c07b", "properties": { "branch_icon": "\ue725 ", "display_stash_count": true, "display_status": true } },
    { "type": "executiontime", "style": "powerline", "foreground": "#1e222a", "background": "#c678dd", "properties": { "threshold": 100 } }
  ]}]
}
JSON
  fi

  # Enable in zsh and bash
  if [ -n "${ZDOTDIR:-}" ] && [ -w "${ZDOTDIR}/.zshrc" ]; then ZSHRC="${ZDOTDIR}/.zshrc"; else ZSHRC="$HOME/.zshrc"; fi
  if ! grep -qs 'oh-my-posh init zsh' "$ZSHRC" 2>/dev/null; then
    log "Enabling oh-my-posh in $ZSHRC"
    $dry_run || printf '\n# aftrs-shell: oh-my-posh\neval "$(oh-my-posh init zsh --config %s)"\n' "$OMP_THEME" >> "$ZSHRC"
  fi
  BASHRC="$HOME/.bashrc"
  if ! grep -qs 'oh-my-posh init bash' "$BASHRC" 2>/dev/null; then
    log "Enabling oh-my-posh in $BASHRC"
    $dry_run || printf '\n# aftrs-shell: oh-my-posh\neval "$(oh-my-posh init bash --config %s)"\n' "$OMP_THEME" >> "$BASHRC"
  fi
}

main() {
  $skip_clone || clone_org_repos
  scan_local_repos_install
  seed_configs
  $refresh_completions && collect_and_install_completions
  install_oh_my_posh
  if $dry_run; then
    echo "Dry run complete. No changes were made."
  else
    log "Installation complete. New shells will load completions and theme automatically."
  fi
}

main "$@"
