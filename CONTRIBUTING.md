# Contributing

## Setup

```bash
git clone https://github.com/hairglasses-studio/dotfiles
cd dotfiles
./install.sh
```

## Structure

- `ghostty/` — Terminal config + 132+ GLSL shaders
- `hyprland/` — Hyprland window manager config  
- `eww/` — eww bar widgets
- `zsh/` — Shell config + starship prompt
- `scripts/` — Helper scripts + shared libraries

## Shader Contributions

Shaders live in `ghostty/shaders/`. To add a new shader:
1. Place your `.glsl` file in the directory
2. Add metadata to `ghostty/shaders/shaders.toml`
3. Test with `shader-test` (requires glslangValidator)
4. Submit a PR

## Code Style

- Shell scripts: use `shellcheck`
- GLSL: self-contained (no external includes at runtime)
- Configs: follow existing formatting
