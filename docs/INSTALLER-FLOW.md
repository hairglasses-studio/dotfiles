# Installer Flow Architecture

The dotfiles installer (`install.sh`) is an idempotent shell script designed for Arch-based Linux (specifically Manjaro). It orchestrates system configuration, package installation, and runtime environment linking.

## Entrypoint: `install.sh`

The core script is designed to be safe to run repeatedly. It performs the following sequential stages:
1. **Pre-flight Checks**: Verifies the host OS, ensures essential binaries are present, and elevates privileges if required.
2. **Profile Loading**: Parses `dotfiles.toml` to determine which feature flags and profiles are active (e.g., minimal vs desktop).
3. **System Setup**: 
   - Uses helpers like `scripts/etc-deploy.sh` to safely copy tracked `/etc/` configurations into the system directory.
   - Deploys critical system services (e.g., greetd) and configures the bootloader (rEFInd) if flagged.
4. **Package Management**: Resolves dependencies using `metapac`, `yay`, and `Pacfile` lists.
5. **Symlink Generation**: Maps local directories in the repository to user configuration paths (`~/.config/`, `~/.local/bin/`). It backs up any existing non-symlink configurations before linking.
6. **Service Enablement**: Links and enables user-scoped systemd services (found in `systemd/`) to ensure the desktop environment functions on the next login.

## Feature Flags (`dotfiles.toml`)

The installer behaves declaratively based on `dotfiles.toml`. Key toggles include:
- Desktop Environment: Hyprland or Sway.
- Tools: Specific CLI utilities or visual tweaks.

## Idempotency and State

To ensure the installer is safe to re-run:
- Existing symlinks matching the expected target are skipped.
- Backups are only created when actual file conflicts exist.
- Systemd enablement checks current status before issuing `systemctl enable`.
