# Arch Linux Installation Scripts (DEPRECATED)

> **Note:** These scripts are from the initial Arch/Manjaro setup and reference outdated packages (JetBrains Mono instead of Maple Mono NF CN, grim instead of wayshot). For current setup, use `install.sh` at the repo root and `metapac sync` for package management.

Personal Arch Linux configuration and automated installation scripts for quickly setting up new systems.

## 🚀 Quick Start

```bash
# Clone this repository
git clone https://github.com/hairglasses/dotfiles-arch.git
cd dotfiles-arch

# Make scripts executable
chmod +x scripts/*.sh

# Run the main installer
./install.sh
```

## 📁 Repository Structure

```
dotfiles-arch/
├── install.sh                 # Main installation orchestrator
├── README.md                  # This file
├── scripts/                   # Installation scripts
│   ├── 01-base-system.sh     # System updates, AUR helper, repositories
│   ├── 02-essential-apps.sh  # Core applications installer
│   ├── 03-extended-apps.sh   # Extended applications installer
│   ├── 04-dev-tools.sh       # Development environment setup
│   ├── 05-system-config.sh   # System configuration and optimizations
│   └── 06-dotfiles.sh        # Dotfiles symlink manager
├── configs/                   # Configuration files
│   ├── electron/             # Electron app configurations
│   ├── shell/                # Shell configurations (bash, zsh, fish)
│   ├── terminal/             # Terminal emulator configs
│   └── system/               # System-wide configurations
├── docs/                      # Documentation
│   ├── APP_RECOMMENDATIONS.md
│   ├── MIGRATION_GUIDE.md
│   ├── QUICK_REFERENCE.md
│   └── TROUBLESHOOTING.md
└── backup/                    # Backup scripts and configurations
    └── backup-configs.sh

```

## 🎯 Features

- **Automated Installation**: Single command to set up a complete Arch Linux environment
- **Modular Scripts**: Run individual components as needed
- **Smart Detection**: Skips already installed packages
- **Interactive Mode**: Choose what to install with menu-driven interface
- **Backup & Restore**: Save and restore your configurations across machines
- **Migration Support**: Smooth transition from Ubuntu/other distros

## 💻 What Gets Installed

### Core Applications
- **Browsers**: Google Chrome, Firefox
- **Development**: VS Code, Cursor IDE, Claude Code, Sublime Text
- **Communication**: Discord, Slack, Telegram, Signal
- **Utilities**: Tailscale, Chrome Remote Desktop
- **Gaming**: Steam
- **Productivity**: Obsidian, Notion, Bitwarden

### Development Environment
- **Languages**: Node.js, Python, Rust, Go
- **Tools**: Docker, Git, GitHub Desktop
- **Databases**: DBeaver, PostgreSQL client
- **API Testing**: Postman, Insomnia

### System Enhancements
- **AUR Helper**: yay
- **Backup**: Timeshift
- **Screenshots**: Flameshot
- **Launcher**: Ulauncher
- **Terminal**: Kitty, Alacritty

## 🔧 Installation Options

### Full Installation (Recommended for new systems)
```bash
./install.sh --full
```

### Essential Apps Only
```bash
./install.sh --essential
```

### Development Tools Only
```bash
./scripts/04-dev-tools.sh
```

### Interactive Selection
```bash
./install.sh --interactive
```

## 📝 Configuration Management

### Backup Current Configurations
```bash
./backup/backup-configs.sh
```

### Restore Configurations
```bash
./install.sh --restore
```

### Update Dotfiles
```bash
git pull
./scripts/06-dotfiles.sh
```

## 🎨 Customization

### Adding Your Own Applications

Edit `scripts/02-essential-apps.sh` or `scripts/03-extended-apps.sh`:

```bash
# Add to OFFICIAL_PACKAGES array for official repo packages
OFFICIAL_PACKAGES+=(
    "your-package-name"
)

# Add to AUR_PACKAGES array for AUR packages
AUR_PACKAGES+=(
    "your-aur-package"
)
```

### Adding Configuration Files

1. Place your config files in the appropriate `configs/` subdirectory
2. Update `scripts/06-dotfiles.sh` to create symlinks
3. Commit and push your changes

## 🔄 Keeping Your System Updated

```bash
# Update everything (system + AUR packages)
yay -Syu

# Update this repository
git pull

# Reapply configurations
./scripts/06-dotfiles.sh
```

## 🐛 Troubleshooting

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common issues and solutions.

## 🤝 Migration from Other Distributions

If you're coming from Ubuntu or another distribution, check out:
- [docs/MIGRATION_GUIDE.md](docs/MIGRATION_GUIDE.md) - Complete migration guide
- [docs/QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md) - Command comparison and tips

## ⚡ Performance Optimizations

The installation automatically applies several optimizations:
- Enables hardware video acceleration for Electron apps
- Configures Wayland support where applicable
- Sets up zram for better memory management
- Optimizes pacman mirror list for fastest downloads

## 🔐 Security Considerations

- Scripts require sudo for system packages but run as user
- AUR packages are built in user space
- Sensitive configurations (SSH keys, tokens) should be manually copied
- Consider encrypting sensitive dotfiles with git-crypt

## 📊 System Requirements

- **OS**: Arch Linux (or Arch-based distributions)
- **RAM**: Minimum 4GB, 8GB+ recommended
- **Storage**: 20GB+ free space
- **Network**: Active internet connection for package downloads

## 🚦 Post-Installation Checklist

After running the installer:

- [ ] Reboot system
- [ ] Configure Timeshift backups
- [ ] Set up Tailscale: `sudo tailscale up`
- [ ] Authenticate Claude Code: `claude auth`
- [ ] Sign into browsers and sync settings
- [ ] Configure Git: `git config --global user.name "Your Name"`
- [ ] Set up SSH keys for GitHub
- [ ] Install VS Code extensions
- [ ] Configure your shell (bash/zsh/fish)

## 📚 Additional Resources

- [Arch Wiki](https://wiki.archlinux.org/) - The best Linux documentation
- [AUR](https://aur.archlinux.org/) - Arch User Repository
- [r/archlinux](https://reddit.com/r/archlinux) - Community support

## 🚀 Future Feature Recommendations

### Planned Enhancements
- **Ansible Integration**: Convert scripts to Ansible playbooks for better idempotency
- **Multi-Machine Profiles**: Support for desktop/laptop/server configurations
- **Cloud Sync**: Automatic backup to cloud storage (Google Drive, Dropbox)
- **Theme Manager**: Unified theme switching across all applications
- **Hardware Detection**: Auto-configure based on detected hardware
- **Container Support**: Run in Docker for testing configurations
- **CI/CD Pipeline**: GitHub Actions for testing script changes

### Advanced Features Under Consideration
- **Declarative Configuration**: NixOS-style system configuration
- **Secrets Management**: Integration with pass or Bitwarden CLI
- **Remote Deployment**: Deploy configurations over SSH
- **Version Pinning**: Lock package versions for stability
- **Rollback System**: Automated rollback on failed updates
- **Performance Monitoring**: System performance tracking and optimization
- **Security Hardening**: Automated security configuration
- **Network Profiles**: Different configurations for different networks

### Community Features
- **Template System**: Easy customization for different use cases
- **Plugin Architecture**: Community-contributed modules
- **Configuration Marketplace**: Share configurations with others
- **Web Dashboard**: Browser-based configuration management
- **Mobile App**: Monitor and manage systems from phone

### Integration Ideas
- **Home Assistant**: Smart home integration
- **Kubernetes**: Local k8s cluster setup
- **Gaming Optimizations**: Steam, Lutris, Proton configuration
- **Content Creation**: OBS, DaVinci Resolve, Blender setup
- **Privacy Tools**: VPN, Tor, encrypted communication setup
- **Backup Strategies**: Automated 3-2-1 backup implementation

## 📄 License

This repository is for personal use. Feel free to fork and modify for your own needs.

## 🙏 Acknowledgments

- Arch Linux community for excellent documentation
- AUR maintainers for keeping packages updated
- All the open-source projects that make this possible
- Starship prompt developers for the amazing shell experience

## 🤝 Contributing

Contributions are welcome! Please feel free to:
- Submit bug reports and feature requests
- Improve documentation
- Submit pull requests with enhancements
- Share your configurations and customizations

## 🚀 Future Features & Roadmap

### Planned Enhancements

#### 🔧 System Features
- **Automated Testing**: GitHub Actions CI/CD for script validation
- **Modular Profiles**: Desktop, Server, Laptop, Gaming profiles
- **Hardware Detection**: Auto-configure based on detected hardware
- **Dual Boot Support**: Windows/Linux dual boot configuration
- **Encrypted Install**: Full disk encryption setup automation
- **Btrfs Snapshots**: Automatic snapshot management with snapper

#### 🎨 Customization
- **Theme Manager**: Easy switching between color schemes
- **Desktop Environments**: Automated setup for GNOME/KDE/XFCE/i3/Hyprland
- **Wallpaper Collection**: Curated wallpaper pack with dynamic switching
- **Icon Themes**: Automatic icon theme installation and configuration
- **Plymouth Themes**: Beautiful boot splash screens

#### 📦 Package Management
- **Package Profiles**: Curated lists for different use cases (Gaming, Development, Creative)
- **Flatpak Integration**: Sandboxed application support
- **AppImage Manager**: Centralized AppImage management
- **Version Pinning**: Lock critical packages to specific versions
- **Rollback System**: Easy package rollback functionality

#### 🔒 Security & Privacy
- **Security Hardening**: CIS benchmark compliance automation
- **VPN Integration**: Auto-configure WireGuard/OpenVPN
- **Firewall Profiles**: Pre-configured firewall rules for common scenarios
- **Privacy Tools**: Tor, DNSCrypt, privacy-focused browser configs
- **Audit System**: Security audit and compliance checking

#### 🛠️ Development
- **Language Managers**: asdf/rtx for multiple language versions
- **Container Orchestration**: Kubernetes local development setup
- **Database Containers**: Pre-configured database containers
- **IDE Configurations**: Automated IDE setup (IntelliJ, VSCode, Neovim)
- **Git Workflows**: Advanced git hooks and workflows

#### 🌐 Cloud & Remote
- **Cloud Sync**: Automated config sync across machines
- **Remote Desktop**: X2Go/XRDP/VNC setup automation
- **SSH Hardening**: Automated SSH security configuration
- **Backup Automation**: Automated backup to cloud storage
- **Distributed Setup**: Multi-machine orchestration

#### 📊 Monitoring & Maintenance
- **System Dashboard**: Grafana/Prometheus monitoring setup
- **Log Management**: Centralized logging with Loki/Elasticsearch
- **Performance Tuning**: Automated kernel parameter optimization
- **Resource Monitoring**: Real-time system resource tracking
- **Update Scheduling**: Intelligent update scheduling

#### 🎮 Gaming Optimization
- **Steam Tweaks**: Proton/Wine optimization
- **GPU Configuration**: NVIDIA/AMD driver optimization
- **Game Mode**: Automated gaming performance mode
- **Emulator Setup**: RetroArch and emulator configuration
- **Streaming Setup**: OBS/Sunshine configuration

#### 🤖 AI & Automation
- **AI Integration**: Local LLM setup (Ollama, LocalAI)
- **Voice Assistant**: Mycroft/Rhasspy integration
- **Automation Rules**: IFTTT-style automation rules
- **Smart Home**: Home Assistant integration
- **Workflow Automation**: n8n/Node-RED setup

### How to Request Features

1. Open an issue with the `enhancement` label
2. Describe your use case
3. Provide examples if possible
4. Vote on existing feature requests

---

**Note**: This is a living repository. Configurations and scripts will evolve based on needs and new discoveries.

**Last Updated**: November 2024
**Arch Linux Version**: Rolling Release
**Primary Use Case**: Development workstation with productivity focus
**Repository**: https://github.com/hairglasses/dotfiles-arch
**Shell Enhancement**: Starship prompt with modern CLI tools
