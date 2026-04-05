# Starship Configuration Presets

This directory contains multiple Starship prompt configurations for different use cases.

## Available Presets

### 1. Default (starship.toml)
Full-featured configuration with all modules enabled. Best for development workstations with modern hardware.

### 2. Minimal (starship-minimal.toml)
Lightweight configuration focusing on essential information. Ideal for:
- Older hardware
- Remote servers
- Quick terminal startup

### 3. Performance (starship-performance.toml)
Optimized for speed with selective module loading. Good balance between features and performance.

### 4. Server (starship-server.toml)
Designed for server environments with emphasis on:
- System status
- User/hostname visibility
- Resource usage

### 5. Git-focused (starship-git.toml)
Enhanced git information display for heavy git users.

## Usage

To use a preset, copy it to your Starship config location:

```bash
# Use default preset
cp ~/.config/starship-presets/starship.toml ~/.config/starship.toml

# Use minimal preset
cp ~/.config/starship-presets/starship-minimal.toml ~/.config/starship.toml

# Use performance preset
cp ~/.config/starship-presets/starship-performance.toml ~/.config/starship.toml
```

## Switching Presets

You can create a simple function to switch between presets:

```bash
# Add to your .bashrc or .zshrc
starship_preset() {
    local preset="${1:-default}"
    local preset_dir="$HOME/.config/starship-presets"
    local preset_file="$preset_dir/starship-${preset}.toml"
    
    if [ "$preset" = "default" ]; then
        preset_file="$preset_dir/starship.toml"
    fi
    
    if [ -f "$preset_file" ]; then
        cp "$preset_file" "$HOME/.config/starship.toml"
        echo "Switched to $preset preset"
    else
        echo "Preset not found: $preset"
        echo "Available presets: default, minimal, performance, server, git"
    fi
}
```

Then use:
```bash
starship_preset minimal
starship_preset performance
starship_preset default
```

## Customization

Feel free to modify any preset to match your preferences. Common customizations:

1. **Change symbols**: Edit the symbol fields in each module
2. **Adjust colors**: Modify style attributes
3. **Toggle modules**: Set `disabled = true` for unwanted modules
4. **Reorder modules**: Change the order in the `format` string

## Performance Tips

- Disable unused modules with `disabled = true`
- Increase `command_timeout` for slow systems
- Reduce `scan_timeout` for faster directory scanning
- Use minimal preset on older hardware

## Resources

- [Starship Documentation](https://starship.rs/config/)
- [Nerd Fonts](https://www.nerdfonts.com/)
- [Starship Presets Gallery](https://starship.rs/presets/)
