#!/usr/bin/env bash
# link-specs.sh — Output managed dotfiles symlink specs in pipe-delimited format.
# Compatible with the install.sh --print-link-specs interface consumed by
# dotfiles-mcp link_inventory.go.
#
# Output format: <source_absolute_path>|<destination_absolute_path>
# One line per managed symlink.
set -euo pipefail

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHEZMOI_SOURCE="$DOTFILES_DIR/home"

# Parse symlink_ files directly from the chezmoi source tree.
# Each symlink_ file's content is the absolute path to the link target.
find "$CHEZMOI_SOURCE" -name 'symlink_*' -type f | sort | while IFS= read -r symlink_file; do
    src="$(cat "$symlink_file")"
    [[ -z "$src" ]] && continue

    # Convert chezmoi source path to target home path
    rel="${symlink_file#"$CHEZMOI_SOURCE/"}"
    dir="$(dirname "$rel")"
    base="$(basename "$rel")"
    base="${base#symlink_}"

    # Reconstruct target path: dot_config -> .config, dot_ssh -> .ssh, etc.
    if [[ "$dir" == "." ]]; then
        # Root-level dotfile (e.g., symlink_dot_zshrc -> ~/.zshrc)
        target="$HOME/${base//dot_/.}"
    else
        # Nested path (e.g., dot_config/symlink_kitty -> ~/.config/kitty)
        target_dir="${dir//dot_/.}"
        target="$HOME/$target_dir/$base"
    fi

    printf '%s|%s\n' "$src" "$target"
done
