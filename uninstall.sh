#!/usr/bin/env bash
set -euo pipefail

# ── Dotfiles Uninstaller ───────────────────────
# Removes symlinks created by install.sh.
# Restores from backup if available.

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

log_info()    { printf "\033[0;34m[INFO]\033[0m  %s\n" "$1"; }
log_success() { printf "\033[0;32m[OK]\033[0m    %s\n" "$1"; }
log_warn()    { printf "\033[0;33m[WARN]\033[0m  %s\n" "$1"; }

unlink_file() {
    local dst="$1"

    if [[ -L "$dst" ]]; then
        local target
        target="$(readlink "$dst")"
        if [[ "$target" == "$DOTFILES_DIR"* ]]; then
            rm "$dst"
            log_success "Removed symlink: $dst"
        else
            log_warn "Skipped (not ours): $dst -> $target"
        fi
    else
        log_warn "Not a symlink: $dst"
    fi
}

echo ""
echo "  dotfiles uninstaller"
echo "  ────────────────────"
echo ""

# Remove individual file symlinks
unlink_file "$HOME/.zshrc"
unlink_file "$HOME/.zshenv"
unlink_file "$HOME/.gitconfig"
unlink_file "$HOME/.ssh/config"
unlink_file "$HOME/.config/starship.toml"

# Remove directory symlinks
unlink_file "$HOME/.config/nvim"
unlink_file "$HOME/.config/bat"
unlink_file "$HOME/.config/fastfetch"
unlink_file "$HOME/.config/delta"
unlink_file "$HOME/.config/git/ignore"
unlink_file "$HOME/.config/gh"
unlink_file "$HOME/.config/k9s"

# Check for backups
echo ""
LATEST_BACKUP=$(ls -dt "$HOME"/.dotfiles-backup-* 2>/dev/null | head -1)
if [[ -n "$LATEST_BACKUP" ]]; then
    log_info "Most recent backup: $LATEST_BACKUP"
    echo "  To restore: cp -r $LATEST_BACKUP/. $HOME/"
else
    log_info "No backups found."
fi

echo ""
log_success "Uninstall complete. Symlinks removed."
echo ""
