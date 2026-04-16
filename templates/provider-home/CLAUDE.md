# Global Instructions

## Launch Contract
- Use the canonical dotfiles launchers for coding sessions: `hg-codex-launch.sh`, `hg-claude-launch.sh`, and `hg-gemini-launch.sh`.
- Interactive shell wrappers route `codex`, `claude`, and `gemini` through those launchers automatically.
- Git repos launch into fresh managed worktrees under `$HOME/.codex/worktrees`; non-git directories run in place.
- If a session already starts inside `$HOME/.codex/worktrees`, keep working there instead of nesting another worktree.

## Environment
- OS: Manjaro Linux (Arch-based, rolling release), packages via pacman/yay
- WM: Hyprland (Wayland) with Sway fallback
- Terminal: Kitty
- Shell: zsh with starship prompt
- Workspace root: `/home/hg/hairglasses-studio`
- Shared research repo: `/home/hg/hairglasses-studio/docs`

## Operating Rules
- Prefer repo-local `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md` before global instructions.
- Use `wl-copy` and `wl-paste` for clipboard work.
- Keep US QWERTY keyboard layout in window-manager configs.
- Treat `.agents/skills/` as the canonical skill surface. Compatibility mirrors are generated.
