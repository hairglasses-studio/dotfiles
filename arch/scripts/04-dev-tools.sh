#!/bin/bash

#############################################
# 04 - Development Tools
# Complete development environment setup
#############################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}==>${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }

# Version control
install_vcs() {
    print_status "Installing version control systems..."
    
    sudo pacman -S --needed --noconfirm \
        git \
        git-lfs \
        subversion \
        mercurial \
        github-cli \
        lazygit
    
    # Configure Git LFS
    git lfs install
    
    print_success "Version control systems installed"
}

# Programming languages
install_languages() {
    print_status "Installing programming languages..."
    
    # Node.js and npm
    sudo pacman -S --needed --noconfirm nodejs npm yarn pnpm
    
    # Python
    sudo pacman -S --needed --noconfirm \
        python \
        python-pip \
        python-pipenv \
        python-poetry \
        python-virtualenv \
        ipython \
        jupyter-notebook
    
    # Rust
    if ! command -v rustc &> /dev/null; then
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source "$HOME/.cargo/env"
        print_success "Rust installed"
    else
        print_success "Rust already installed"
    fi
    
    # Go
    sudo pacman -S --needed --noconfirm go gopls
    
    # Java
    sudo pacman -S --needed --noconfirm \
        jdk-openjdk \
        maven \
        gradle
    
    # Ruby
    sudo pacman -S --needed --noconfirm ruby rubygems
    
    # PHP
    sudo pacman -S --needed --noconfirm \
        php \
        php-apache \
        composer
    
    # C/C++
    sudo pacman -S --needed --noconfirm \
        gcc \
        clang \
        cmake \
        make \
        gdb \
        valgrind
    
    print_success "Programming languages installed"
}

# Docker and containers
install_containers() {
    print_status "Installing Docker and container tools..."
    
    sudo pacman -S --needed --noconfirm \
        docker \
        docker-compose \
        docker-buildx \
        podman \
        buildah \
        skopeo
    
    # Add user to docker group
    sudo usermod -aG docker "$USER"
    
    # Enable Docker service
    sudo systemctl enable --now docker.service
    
    # Install Docker Desktop (optional)
    echo -n "Install Docker Desktop? (y/N): "
    read -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        yay -S --needed --noconfirm docker-desktop
    fi
    
    print_success "Docker and container tools installed"
    print_warning "Log out and back in for docker group membership to take effect"
}

# Kubernetes tools
install_kubernetes() {
    print_status "Installing Kubernetes tools..."
    
    sudo pacman -S --needed --noconfirm \
        kubectl \
        kubectx \
        helm \
        k9s
    
    # Minikube for local development
    yay -S --needed --noconfirm minikube-bin
    
    # Kind (Kubernetes in Docker)
    yay -S --needed --noconfirm kind-bin
    
    print_success "Kubernetes tools installed"
}

# Database clients and tools
install_databases() {
    print_status "Installing database tools..."
    
    sudo pacman -S --needed --noconfirm \
        postgresql-libs \
        mariadb-clients \
        redis \
        sqlite \
        mongodb-tools
    
    # GUI tools
    sudo pacman -S --needed --noconfirm dbeaver
    yay -S --needed --noconfirm \
        beekeeper-studio-bin \
        robo3t-bin
    
    print_success "Database tools installed"
}

# Cloud CLI tools
install_cloud_cli() {
    print_status "Installing cloud CLI tools..."
    
    # AWS CLI
    sudo pacman -S --needed --noconfirm aws-cli-v2
    
    # Google Cloud SDK
    yay -S --needed --noconfirm google-cloud-cli
    
    # Azure CLI
    yay -S --needed --noconfirm azure-cli
    
    # Terraform
    sudo pacman -S --needed --noconfirm terraform
    
    # Ansible
    sudo pacman -S --needed --noconfirm ansible ansible-lint
    
    print_success "Cloud CLI tools installed"
}

# Development utilities
install_dev_utils() {
    print_status "Installing development utilities..."
    
    sudo pacman -S --needed --noconfirm \
        httpie \
        jq \
        yq \
        ripgrep \
        fd \
        bat \
        exa \
        fzf \
        the_silver_searcher \
        ctags \
        neovim \
        tmux \
        meld \
        diffuse
    
    # Install from AUR
    yay -S --needed --noconfirm \
        ngrok-bin \
        localtunnel \
        postman-bin \
        insomnia-bin
    
    print_success "Development utilities installed"
}

# IDEs and editors
install_ides() {
    print_status "Installing additional IDEs..."
    
    echo -n "Install JetBrains Toolbox? (y/N): "
    read -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        yay -S --needed --noconfirm jetbrains-toolbox
    fi
    
    echo -n "Install Android Studio? (y/N): "
    read -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        yay -S --needed --noconfirm android-studio
    fi
    
    echo -n "Install Arduino IDE? (y/N): "
    read -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        sudo pacman -S --needed --noconfirm arduino arduino-cli
    fi
    
    print_success "IDEs installation completed"
}

# Shell enhancements
install_shell_tools() {
    print_status "Installing shell enhancements..."
    
    # Zsh and Oh My Zsh
    sudo pacman -S --needed --noconfirm zsh zsh-completions
    
    if [ ! -d "$HOME/.oh-my-zsh" ]; then
        sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
    fi
    
    # Fish shell
    sudo pacman -S --needed --noconfirm fish fisher
    
    # Starship prompt
    sudo pacman -S --needed --noconfirm starship
    
    # Shell utilities
    yay -S --needed --noconfirm \
        zoxide \
        atuin-bin \
        mcfly
    
    print_success "Shell enhancements installed"
}

# Configure development environment
configure_dev_env() {
    print_status "Configuring development environment..."
    
    # Create common directories
    mkdir -p ~/Projects
    mkdir -p ~/Scripts
    mkdir -p ~/.local/bin
    
    # NPM global directory
    mkdir -p ~/.npm-global
    npm config set prefix '~/.npm-global'
    
    # Add to PATH (for bash)
    if ! grep -q "npm-global" ~/.bashrc; then
        echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
    fi
    
    # Git global configuration
    if [ ! -f ~/.gitconfig ]; then
        print_status "Setting up Git configuration..."
        read -p "Enter your Git name: " git_name
        read -p "Enter your Git email: " git_email
        
        git config --global user.name "$git_name"
        git config --global user.email "$git_email"
        git config --global init.defaultBranch main
        git config --global pull.rebase false
        git config --global core.editor "code --wait"
    fi
    
    print_success "Development environment configured"
}

# Main execution
main() {
    print_status "Setting up development environment..."
    
    install_vcs
    install_languages
    install_containers
    install_kubernetes
    install_databases
    install_cloud_cli
    install_dev_utils
    install_ides
    install_shell_tools
    configure_dev_env
    
    print_success "Development environment setup completed!"
    print_warning "Remember to:"
    print_warning "  - Log out and back in for docker group membership"
    print_warning "  - Source ~/.bashrc for PATH changes"
    print_warning "  - Configure your cloud CLI tools with credentials"
}

main
