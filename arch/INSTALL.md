# Quick Installation Guide

## 🚀 One-Line Install

```bash
git clone https://github.com/hairglasses/dotfiles-arch.git && cd dotfiles-arch && chmod +x setup.sh && ./setup.sh
```

## 📋 Step-by-Step

### 1. Clone the Repository
```bash
git clone https://github.com/hairglasses/dotfiles-arch.git
cd dotfiles-arch
```

### 2. Run Setup Script
```bash
chmod +x setup.sh
./setup.sh
```

### 3. Choose Installation Type
- **Option 1**: Quick Setup - Just yay and configs
- **Option 2**: Essential Apps - Your requested applications
- **Option 3**: Extended Installer - Interactive with 40+ apps
- **Option 4**: Configure Services - System services setup
- **Option 5**: Optimize Electron - Performance tweaks
- **Option 6**: Full Setup - Everything automated

## 🎯 Recommended First Run

For new Arch installations:
```bash
./setup.sh  # Choose option 6 (Full Setup)
```

For existing Arch installations:
```bash
./setup.sh  # Choose option 3 (Extended Installer)
```

## ⚡ Quick Commands

### Install specific category
```bash
# Development tools only
cat lists/development.txt | xargs yay -S --needed

# Productivity apps only  
cat lists/productivity.txt | xargs yay -S --needed

# Media apps only
cat lists/media.txt | xargs yay -S --needed
```

### Just the essentials
```bash
./scripts/arch-installer.sh
```

### Configure services (requires sudo)
```bash
sudo ./scripts/helpers/configure-services.sh
```

### Optimize Electron apps
```bash
./scripts/helpers/optimize-electron.sh
```

## 📝 Post-Installation

1. **Reboot** for all changes to take effect
2. **Authenticate services**:
   - `claude auth` - Claude Code
   - `sudo tailscale up` - Tailscale VPN
   - `gh auth login` - GitHub CLI
3. **Configure backups**: `sudo timeshift-gtk`
4. **Set up cloud sync** for your preferred services

## 🔧 Customization

Edit these files to customize your setup:
- `lists/essential.txt` - Core packages
- `configs/electron-flags.conf` - Electron optimizations
- `scripts/arch-installer.sh` - Add your own apps

## 📚 Documentation

- [README.md](README.md) - Full documentation
- [Migration Guide](docs/MIGRATION_GUIDE.md) - From Ubuntu to Arch
- [App Recommendations](docs/APP_RECOMMENDATIONS.md) - Detailed app info
- [Quick Reference](docs/QUICK_REFERENCE.md) - Commands cheatsheet

## 💡 Tips

- Run `yay -Syu` weekly to keep system updated
- Use `timeshift` before major changes
- Check `/var/log/pacman.log` if issues arise
- Join r/archlinux for community support

---

*Happy Arching! 🐧*
