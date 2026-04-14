# Contributing

## Setup

```bash
git clone https://github.com/hairglasses-studio/dotfiles
cd dotfiles
./install.sh
```

## Structure

- `kitty/` — Kitty config, CRTty shader assets, and theme playlists
- `hyprland/` — Hyprland window manager config  
- `ironbar/` — Ironbar menubar config
- `zsh/` — Shell config + starship prompt
- `scripts/` — Helper scripts + shared libraries

## Shader Contributions

Kitty CRTty shaders live in `kitty/shaders/crtty/`. To add a new shader:
1. Place your `.glsl` file in the directory
2. Add metadata to `kitty/shaders/crtty/shaders.toml`
3. Test with `shader-test` (requires glslangValidator)
4. Submit a PR

## Code Style

- Shell scripts: use `shellcheck`
- GLSL: self-contained (no external includes at runtime)
- Configs: follow existing formatting
