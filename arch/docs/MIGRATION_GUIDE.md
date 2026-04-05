# Ubuntu to Arch Linux Migration Guide

## 🔄 Complete Migration Checklist

### Pre-Migration (On Ubuntu)

#### 1. Backup Essential Data
```bash
# Create backup directory
mkdir -p ~/ubuntu-backup

# VS Code settings & extensions
cp -r ~/.config/Code/User ~/ubuntu-backup/vscode-settings
code --list-extensions > ~/ubuntu-backup/vscode-extensions.txt

# Git configuration
cp ~/.gitconfig ~/ubuntu-backup/
cp -r ~/.ssh ~/ubuntu-backup/ssh-keys

# Browser bookmarks (Chrome)
cp -r ~/.config/google-chrome/Default/Bookmarks ~/ubuntu-backup/

# Package list
dpkg --get-selections | grep -v deinstall > ~/ubuntu-backup/ubuntu-packages.txt

# Application configs
cp -r ~/.config/discord ~/ubuntu-backup/
cp -r ~/.config/slack ~/ubuntu-backup/
cp -r ~/.config/obsidian ~/ubuntu-backup/  # if using

# Custom scripts
cp -r ~/bin ~/ubuntu-backup/scripts  # if you have any

# Dotfiles
cp ~/.bashrc ~/.zshrc ~/.vimrc ~/ubuntu-backup/  # as applicable
```

#### 2. Document System Information
```bash
# Hardware info
inxi -Fxz > ~/ubuntu-backup/system-info.txt

# Disk partitions
lsblk -f > ~/ubuntu-backup/disk-layout.txt

# Network configuration
ip addr > ~/ubuntu-backup/network-config.txt
```

### Installation Phase

#### 1. Create Arch Installation Media
```bash
# Download Arch ISO
wget https://archlinux.org/iso/latest/archlinux-x86_64.iso

# Verify signature (recommended)
wget https://archlinux.org/iso/latest/archlinux-x86_64.iso.sig
gpg --keyserver-options auto-key-retrieve --verify archlinux-x86_64.iso.sig

# Create bootable USB (replace /dev/sdX with your USB device)
sudo dd bs=4M if=archlinux-x86_64.iso of=/dev/sdX status=progress oflag=sync
```

#### 2. Partition Recommendations
```
# UEFI System
/dev/nvme0n1p1  512M  EFI System  /boot/efi
/dev/nvme0n1p2  40G   Linux       /
/dev/nvme0n1p3  Rest  Linux       /home

# Optional: Separate /var for AUR builds
/dev/nvme0n1p4  20G   Linux       /var
```

### Post-Installation Setup

#### 1. Essential System Configuration
```bash
# Set timezone
ln -sf /usr/share/zoneinfo/Region/City /etc/localtime
hwclock --systohc

# Locale
echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
locale-gen
echo "LANG=en_US.UTF-8" > /etc/locale.conf

# Hostname
echo "arch-workstation" > /etc/hostname

# Enable NetworkManager
systemctl enable NetworkManager
```

#### 2. Install Base Desktop Environment

##### Option A: GNOME (Most Ubuntu-like)
```bash
sudo pacman -S gnome gnome-extra gdm
sudo systemctl enable gdm
```

##### Option B: KDE Plasma (Most customizable)
```bash
sudo pacman -S plasma kde-applications sddm
sudo systemctl enable sddm
```

##### Option C: XFCE (Lightweight)
```bash
sudo pacman -S xfce4 xfce4-goodies lightdm lightdm-gtk-greeter
sudo systemctl enable lightdm
```

### Application Migration

#### 1. Restore VS Code
```bash
# Install VS Code
yay -S visual-studio-code-bin

# Restore settings
cp -r ~/ubuntu-backup/vscode-settings ~/.config/Code/User

# Restore extensions
cat ~/ubuntu-backup/vscode-extensions.txt | xargs -L 1 code --install-extension
```

#### 2. Restore Git & SSH
```bash
# Install git
sudo pacman -S git

# Restore configs
cp ~/ubuntu-backup/.gitconfig ~/
cp -r ~/ubuntu-backup/ssh-keys ~/.ssh
chmod 700 ~/.ssh
chmod 600 ~/.ssh/*
```

#### 3. Restore Browser Data
```bash
# Install Chrome
yay -S google-chrome

# Restore bookmarks (after first launch)
cp ~/ubuntu-backup/Bookmarks ~/.config/google-chrome/Default/
```

## 🔀 Package Equivalents

### Core System Tools
| Ubuntu Package | Arch Package | Notes |
|----------------|--------------|-------|
| build-essential | base-devel | Development tools |
| linux-generic | linux | Kernel package |
| apt-file | pkgfile | Search package files |
| software-properties-common | N/A | Use /etc/pacman.conf |
| update-manager | N/A | Rolling release |

### Common Applications
| Ubuntu | Arch | Install Command |
|--------|------|-----------------|
| firefox (snap) | firefox | `sudo pacman -S firefox` |
| thunderbird (snap) | thunderbird | `sudo pacman -S thunderbird` |
| libreoffice (snap) | libreoffice-fresh | `sudo pacman -S libreoffice-fresh` |
| gimp (snap) | gimp | `sudo pacman -S gimp` |
| vlc (snap) | vlc | `sudo pacman -S vlc` |
| code (deb) | visual-studio-code-bin | `yay -S visual-studio-code-bin` |
| chrome (deb) | google-chrome | `yay -S google-chrome` |
| docker.io | docker | `sudo pacman -S docker` |

### Development Tools
| Ubuntu | Arch | Notes |
|--------|------|-------|
| nodejs | nodejs | Same package name |
| npm | npm | Same package name |
| python3-pip | python-pip | Python included by default |
| openjdk-11-jdk | jdk11-openjdk | Multiple versions available |
| mysql-server | mariadb | MariaDB is default |
| postgresql | postgresql | Same package name |

## 🎨 Desktop Environment Tweaks

### Make GNOME Feel Like Ubuntu
```bash
# Install Ubuntu-like themes
yay -S yaru-gtk-theme yaru-icon-theme

# Install GNOME extensions
yay -S gnome-shell-extension-dash-to-dock
yay -S gnome-shell-extension-appindicator
yay -S gnome-shell-extension-system-monitor

# Ubuntu fonts
yay -S ttf-ubuntu-font-family

# Apply theme
gsettings set org.gnome.desktop.interface gtk-theme "Yaru"
gsettings set org.gnome.desktop.interface icon-theme "Yaru"
```

### Essential GNOME Extensions
1. **Dash to Dock** - Ubuntu-style dock
2. **AppIndicator** - System tray icons
3. **System Monitor** - Resource usage in top bar
4. **Clipboard Indicator** - Clipboard history
5. **Caffeine** - Prevent screen sleep

## 🔧 System Maintenance Differences

### Ubuntu vs Arch Maintenance

#### Ubuntu Style
```bash
# Update system (Ubuntu)
sudo apt update && sudo apt upgrade

# Clean packages (Ubuntu)
sudo apt autoremove && sudo apt autoclean
```

#### Arch Style
```bash
# Update system (Arch)
yay -Syu

# Clean packages (Arch)
yay -Sc

# Remove orphans
yay -Rns $(yay -Qtdq)
```

### Scheduled Maintenance
```bash
# Create update alias in ~/.bashrc
alias update='yay -Syu'
alias cleanup='yay -Sc && yay -Rns $(yay -Qtdq)'

# Optional: Set up automatic updates (not recommended for Arch)
# Better: Set reminder to update weekly
```

## 📦 Snap to AUR/Flatpak Migration

### Snap Applications → Arch Alternatives
| Snap Package | Arch Alternative | Source |
|--------------|------------------|--------|
| spotify | spotify | AUR |
| discord | discord | Official |
| slack | slack-desktop | AUR |
| code | visual-studio-code-bin | AUR |
| postman | postman-bin | AUR |
| notion-snap | notion-app-electron | AUR |
| obsidian | obsidian | Official |

### If You Need Snap Compatibility
```bash
# Install snapd (not recommended)
yay -S snapd
sudo systemctl enable --now snapd.socket
sudo ln -s /var/lib/snapd/snap /snap
```

### Better Alternative: Flatpak
```bash
# Install Flatpak
sudo pacman -S flatpak

# Add Flathub
flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo

# Install apps
flatpak install flathub com.spotify.Client
flatpak install flathub com.slack.Slack
```

## 🚀 Performance Optimizations

### Make Arch Faster Than Ubuntu

#### 1. Boot Time
```bash
# Analyze boot time
systemd-analyze

# Disable unnecessary services
systemd-analyze blame  # See what's slow
sudo systemctl disable bluetooth  # If not using
sudo systemctl disable cups  # If not printing
```

#### 2. Pacman Optimization
```bash
# Edit /etc/pacman.conf
# Uncomment:
Color
ParallelDownloads = 5
ILoveCandy  # Easter egg progress bar

# Add to [options]:
NoExtract = usr/share/help/* !usr/share/help/C/*
NoExtract = usr/share/gtk-doc/html/*
NoExtract = usr/share/locale/* !usr/share/locale/en_US/*
```

#### 3. Makepkg Optimization
```bash
# Edit /etc/makepkg.conf
# Set for your CPU (example for 8 cores):
MAKEFLAGS="-j8"

# Enable compilation cache
yay -S ccache
# Add to /etc/makepkg.conf:
BUILDENV=(ccache fakeroot !distcc color !check !sign)
```

## ⚠️ Common Pitfalls to Avoid

### 1. **Partial Upgrades**
```bash
# NEVER do this:
sudo pacman -Sy package-name  # Wrong!

# ALWAYS do this:
sudo pacman -Syu package-name  # Correct!
```

### 2. **AUR Security**
```bash
# Always review PKGBUILD before installing
yay -G package-name
cd package-name
vim PKGBUILD  # Review the build script
```

### 3. **Config File Management**
```bash
# Use pacdiff to manage .pacnew files
sudo pacman -S pacman-contrib
pacdiff  # After updates
```

### 4. **Kernel Updates**
```bash
# Keep LTS kernel as fallback
sudo pacman -S linux-lts linux-lts-headers

# Configure GRUB to show menu
# Edit /etc/default/grub:
GRUB_TIMEOUT=5
GRUB_TIMEOUT_STYLE=menu
```

## 🆘 Recovery Preparation

### Create Recovery Tools
```bash
# Install system backup
sudo pacman -S timeshift

# Configure automatic snapshots
sudo timeshift-gtk

# Create rescue USB with your config
# Use archiso or EndeavourOS ISO
```

### Essential Recovery Commands
```bash
# Boot from live USB and mount system
mount /dev/nvme0n1p2 /mnt
mount /dev/nvme0n1p1 /mnt/boot/efi
arch-chroot /mnt

# Fix broken packages
pacman -Syu
pacman -S package-name --overwrite '*'

# Restore from Timeshift
timeshift --restore
```

## 📚 Learning Resources

### Must-Read Documentation
1. **Arch Wiki** - https://wiki.archlinux.org/
   - Best Linux documentation available
   - Solution to 99% of problems

2. **Arch Forums** - https://bbs.archlinux.org/
   - Active community support
   - Search before posting

3. **Reddit** - r/archlinux
   - Friendly community
   - Good for beginners

### Useful Commands Reference
```bash
# Package management
pacman -Ss keyword     # Search packages
pacman -Si package     # Package info
pacman -Ql package     # List package files
pacman -Qo /path/file  # Which package owns file
pacman -Qe             # List explicitly installed
pacman -Qm             # List AUR packages

# System information
journalctl -xe         # System logs
systemctl status       # Service status
lspci -k              # Hardware info
```

## 🎯 First Week Goals

### Day 1-2: Core System
- [ ] Install Arch with DE
- [ ] Configure network
- [ ] Install yay
- [ ] Run initial system update

### Day 3-4: Essential Apps
- [ ] Install browsers
- [ ] Install development tools
- [ ] Install communication apps
- [ ] Configure backup solution

### Day 5-6: Customization
- [ ] Configure themes
- [ ] Set up dotfiles
- [ ] Install productivity apps
- [ ] Configure shortcuts

### Day 7: Documentation
- [ ] Document your setup
- [ ] Create backup of configs
- [ ] Note any issues/solutions
- [ ] Plan next improvements

## 💡 Pro Tips

1. **Read the Wiki** - Seriously, the Arch Wiki has everything
2. **Update Regularly** - Weekly is good, daily is fine
3. **Learn Pacman** - It's more powerful than apt
4. **Use AUR Wisely** - Review PKGBUILDs
5. **Keep Notes** - Document your customizations
6. **Join Community** - Arch users are helpful
7. **Be Patient** - Learning curve is worth it
8. **Backup Often** - Especially before major changes
9. **Experiment** - Arch encourages learning
10. **Have Fun** - Enjoy the control and freedom!

---

*Welcome to Arch Linux - You're going to love the control and performance!*
