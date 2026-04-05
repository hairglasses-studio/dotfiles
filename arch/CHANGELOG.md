# Changelog - Arch Linux Dotfiles

## [2.0.0] - 2024-11-06

### Added
- **Unattended Installation Mode** (`--full-auto`)
  - Complete hands-free installation of all applications
  - AUTO_INSTALL environment variable support
  - No user prompts required - perfect for automation

- **Shell Enhancements Module** (`07-shell-enhancements.sh`)
  - Starship prompt with 3 presets (default, minimal, performance)
  - Modern terminal tools (bat, exa, ripgrep, fd, fzf, zoxide)
  - Nerd Fonts installation (JetBrainsMono, FiraCode, Hack, etc.)
  - Oh My Zsh with plugins (autosuggestions, syntax highlighting)
  - Powerlevel10k theme option
  - Fish shell with plugins
  - Terminal tool suite (lazygit, lazydocker, gitui, bottom, atuin)
  - Shell history sync and smart search
  - Custom aliases for modern replacements

- **Starship Configuration System**
  - Three configuration presets
  - Preset switcher script (`starship-preset`)
  - Comprehensive prompt elements
  - Git integration with detailed status
  - Language detection and version display
  - Performance optimizations

- **Enhanced Installation Options**
  - `--shell` flag for shell-only installation
  - Updated interactive menu with 11 options
  - Custom selection now includes shell enhancements

- **Future Feature Recommendations**
  - Ansible integration roadmap
  - Multi-machine profile support
  - Cloud sync capabilities
  - Theme management system
  - Hardware auto-detection
  - Container support for testing
  - CI/CD pipeline plans

### Modified
- **Main Installer (`install.sh`)**
  - Added `--full-auto` option for unattended installation
  - Added `--shell` option for shell enhancements only
  - Updated menu from 9 to 11 options
  - Improved help documentation

- **Extended Apps Script (`03-extended-apps.sh`)**
  - Added AUTO_INSTALL mode support
  - Automatic yes to all prompts in unattended mode
  - Improved package detection

- **Dotfiles Script (`06-dotfiles.sh`)**
  - Enhanced shell configurations
  - Added support for Fish shell
  - Integrated modern tool aliases
  - Added Powerlevel10k support
  - Improved Zsh plugin management

- **README.md**
  - Added unattended installation documentation
  - Added shell enhancements section
  - Added future features roadmap
  - Added contributing section
  - Updated quick start guide

### Integrated
- **Starship Ultimate Setup**
  - Incorporated configuration from starship-ultimate-setup
  - Added multiple prompt presets
  - Integrated Nerd Font management
  - Added preset switching capability

### Improved
- **Shell Configuration**
  - Unified bash, zsh, and fish configurations
  - Smart command detection for aliases
  - Conditional loading of integrations
  - Better PATH management
  - Enhanced prompt fallbacks

- **Documentation**
  - More comprehensive feature list
  - Better installation examples
  - Clearer option descriptions
  - Added shell customization guide

## [1.0.0] - 2024-11-06 (Initial Release)

### Created
- Base system setup script
- Essential applications installer
- Extended applications installer
- Development tools installer
- System configuration script
- Dotfiles management
- Backup and restore functionality
- Migration guide from Ubuntu
- Application recommendations
- Troubleshooting guide
- Quick reference guide

---

## Upgrade Instructions

For existing users upgrading from v1.0.0 to v2.0.0:

1. **Backup your current configuration:**
   ```bash
   ./backup/backup-configs.sh
   ```

2. **Pull the latest changes:**
   ```bash
   git pull origin main
   ```

3. **Install new shell enhancements:**
   ```bash
   ./install.sh --shell
   ```

4. **Or do a full reinstall (will skip existing packages):**
   ```bash
   ./install.sh --full-auto
   ```

5. **Switch to a Starship preset:**
   ```bash
   starship-preset minimal  # or 'default' or 'performance'
   ```

6. **Restart your terminal or reload shell:**
   ```bash
   source ~/.${SHELL##*/}rc
   ```

## Breaking Changes

None - all changes are backward compatible. Existing configurations will continue to work.

## Notes

- The unattended mode is perfect for CI/CD environments or automated deployments
- Shell enhancements significantly improve terminal productivity
- Starship prompt is much faster than traditional prompts while being more feature-rich
- Modern terminal tools (bat, exa, etc.) are aliased to replace traditional commands but originals remain available
