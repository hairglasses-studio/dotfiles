#!/bin/bash

#############################################
# 06 - Dotfiles Setup
# Symlink and manage configuration files
#############################################

set -e

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
CONFIGS_DIR="$REPO_DIR/configs"
BACKUP_DIR="$HOME/.dotfiles-backup"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

print_status() { echo -e "${BLUE}==>${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Create backup directory
create_backup() {
    if [ ! -d "$BACKUP_DIR" ]; then
        mkdir -p "$BACKUP_DIR"
        print_success "Created backup directory: $BACKUP_DIR"
    fi
}

# Backup existing file
backup_file() {
    local file=$1
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    if [ -e "$file" ] && [ ! -L "$file" ]; then
        local backup_name="$(basename "$file").$timestamp"
        cp -r "$file" "$BACKUP_DIR/$backup_name"
        print_info "Backed up $file to $BACKUP_DIR/$backup_name"
    fi
}

# Create symlink
create_symlink() {
    local source=$1
    local target=$2
    
    # Create parent directory if it doesn't exist
    mkdir -p "$(dirname "$target")"
    
    # Backup existing file
    if [ -e "$target" ] && [ ! -L "$target" ]; then
        backup_file "$target"
        rm -rf "$target"
    elif [ -L "$target" ]; then
        rm "$target"
    fi
    
    # Create symlink
    ln -s "$source" "$target"
    print_success "Linked $source -> $target"
}

# Setup shell configurations
setup_shell() {
    print_status "Setting up shell configurations..."
    
    # Create shell config directory
    mkdir -p "$CONFIGS_DIR/shell"
    
    # Bash configuration
    if [ ! -f "$CONFIGS_DIR/shell/bashrc" ]; then
        cat > "$CONFIGS_DIR/shell/bashrc" << 'EOF'
# Arch Linux Bash Configuration

# History
export HISTSIZE=10000
export HISTFILESIZE=20000
export HISTCONTROL=ignoreboth:erasedups
shopt -s histappend

# Aliases
alias ls='ls --color=auto'
alias ll='ls -lah'
alias la='ls -A'
alias l='ls -CF'
alias grep='grep --color=auto'
alias df='df -h'
alias du='du -h'
alias free='free -h'
alias vim='nvim'
alias vi='nvim'

# Modern replacements (if installed)
command -v exa &> /dev/null && alias ls='exa --icons' && alias ll='exa -l --icons' && alias la='exa -la --icons'
command -v bat &> /dev/null && alias cat='bat'
command -v fd &> /dev/null && alias find='fd'
command -v rg &> /dev/null && alias grep='rg'

# Arch specific
alias update='yay -Syu'
alias install='yay -S'
alias search='yay -Ss'
alias remove='yay -R'
alias orphans='yay -Qtdq | yay -Rns -'
alias mirrors='sudo reflector --latest 20 --protocol https --sort rate --save /etc/pacman.d/mirrorlist'

# Development
alias g='git'
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias gd='git diff'
alias gco='git checkout'
alias gb='git branch'

# Docker
alias d='docker'
alias dc='docker-compose'
alias dps='docker ps'
alias dimg='docker images'

# Navigation
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias ~='cd ~'
alias projects='cd ~/Projects'

# Safety
alias cp='cp -i'
alias mv='mv -i'
alias rm='rm -i'

# Colored man pages
export LESS_TERMCAP_mb=$'\e[1;32m'
export LESS_TERMCAP_md=$'\e[1;32m'
export LESS_TERMCAP_me=$'\e[0m'
export LESS_TERMCAP_se=$'\e[0m'
export LESS_TERMCAP_so=$'\e[01;33m'
export LESS_TERMCAP_ue=$'\e[0m'
export LESS_TERMCAP_us=$'\e[1;4;31m'

# PATH
export PATH="$HOME/.local/bin:$HOME/.npm-global/bin:$HOME/.cargo/bin:$PATH"

# Editor
export EDITOR=nvim
export VISUAL=nvim

# Prompt (if not using starship)
if ! command -v starship &> /dev/null; then
    PS1='[\u@\h \W]\$ '
else
    eval "$(starship init bash)"
fi

# Node
export NPM_CONFIG_PREFIX="$HOME/.npm-global"

# FZF
if [ -f /usr/share/fzf/key-bindings.bash ]; then
    source /usr/share/fzf/key-bindings.bash
fi
if [ -f /usr/share/fzf/completion.bash ]; then
    source /usr/share/fzf/completion.bash
fi

# Zoxide
if command -v zoxide &> /dev/null; then
    eval "$(zoxide init bash)"
fi

# Atuin (shell history)
if command -v atuin &> /dev/null; then
    eval "$(atuin init bash)"
fi

# Custom functions
mkcd() {
    mkdir -p "$1" && cd "$1"
}

extract() {
    if [ -f "$1" ]; then
        case "$1" in
            *.tar.bz2) tar xjf "$1" ;;
            *.tar.gz) tar xzf "$1" ;;
            *.tar.xz) tar xJf "$1" ;;
            *.bz2) bunzip2 "$1" ;;
            *.rar) unrar e "$1" ;;
            *.gz) gunzip "$1" ;;
            *.tar) tar xf "$1" ;;
            *.tbz2) tar xjf "$1" ;;
            *.tgz) tar xzf "$1" ;;
            *.zip) unzip "$1" ;;
            *.Z) uncompress "$1" ;;
            *.7z) 7z x "$1" ;;
            *) echo "'$1' cannot be extracted" ;;
        esac
    else
        echo "'$1' is not a valid file"
    fi
}

# Source additional configs
[ -f "$HOME/.config/shell/aliases" ] && source "$HOME/.config/shell/aliases"
EOF
    fi
    
    create_symlink "$CONFIGS_DIR/shell/bashrc" "$HOME/.bashrc"
    
    # Zsh configuration
    if command -v zsh &> /dev/null && [ ! -f "$CONFIGS_DIR/shell/zshrc" ]; then
        cat > "$CONFIGS_DIR/shell/zshrc" << 'EOF'
# Arch Linux Zsh Configuration

# Oh My Zsh
export ZSH="$HOME/.oh-my-zsh"
ZSH_THEME="robbyrussell"  # Will be overridden by Starship or P10k
plugins=(
    git
    docker
    docker-compose
    npm
    node
    python
    rust
    golang
    kubectl
    helm
    terraform
    ansible
    aws
    zsh-autosuggestions
    zsh-syntax-highlighting
    fast-syntax-highlighting
)

if [ -f "$ZSH/oh-my-zsh.sh" ]; then
    source "$ZSH/oh-my-zsh.sh"
fi

# Source bash aliases
if [ -f "$HOME/.bashrc" ]; then
    source "$HOME/.bashrc"
fi

# Powerlevel10k theme
if [ -f "$ZSH_CUSTOM/themes/powerlevel10k/powerlevel10k.zsh-theme" ]; then
    source "$ZSH_CUSTOM/themes/powerlevel10k/powerlevel10k.zsh-theme"
fi

# Starship (overrides theme if installed)
if command -v starship &> /dev/null; then
    eval "$(starship init zsh)"
fi

# Additional integrations
if command -v zoxide &> /dev/null; then
    eval "$(zoxide init zsh)"
fi

if command -v atuin &> /dev/null; then
    eval "$(atuin init zsh)"
fi
EOF
    fi
    
    [ -f "$CONFIGS_DIR/shell/zshrc" ] && create_symlink "$CONFIGS_DIR/shell/zshrc" "$HOME/.zshrc"
    
    # Fish configuration
    if command -v fish &> /dev/null; then
        mkdir -p "$CONFIGS_DIR/shell/fish"
        if [ ! -f "$CONFIGS_DIR/shell/fish/config.fish" ]; then
            cat > "$CONFIGS_DIR/shell/fish/config.fish" << 'EOF'
# Arch Linux Fish Configuration

# Starship prompt
if command -v starship &> /dev/null
    starship init fish | source
end

# Aliases
alias ls='exa --icons'
alias ll='exa -l --icons'
alias la='exa -la --icons'
alias update='yay -Syu'
alias install='yay -S'

# Environment
set -gx EDITOR nvim
set -gx PATH $HOME/.local/bin $HOME/.npm-global/bin $PATH

# Integrations
if command -v zoxide &> /dev/null
    zoxide init fish | source
end

if command -v atuin &> /dev/null
    atuin init fish | source
end
EOF
        fi
        mkdir -p "$HOME/.config/fish"
        create_symlink "$CONFIGS_DIR/shell/fish/config.fish" "$HOME/.config/fish/config.fish"
    fi
    
    print_success "Shell configurations set up"
}

# Setup Git configuration
setup_git() {
    print_status "Setting up Git configuration..."
    
    mkdir -p "$CONFIGS_DIR/git"
    
    if [ ! -f "$CONFIGS_DIR/git/gitconfig" ]; then
        cat > "$CONFIGS_DIR/git/gitconfig" << 'EOF'
[core]
    editor = nvim
    pager = less
    excludesfile = ~/.gitignore_global

[init]
    defaultBranch = main

[pull]
    rebase = false

[push]
    default = current

[alias]
    st = status
    co = checkout
    ci = commit
    br = branch
    unstage = reset HEAD --
    last = log -1 HEAD
    visual = !gitk
    lg = log --oneline --graph --decorate --all

[color]
    ui = auto

[diff]
    tool = meld

[merge]
    tool = meld

[fetch]
    prune = true
EOF
    fi
    
    if [ ! -f "$CONFIGS_DIR/git/gitignore_global" ]; then
        cat > "$CONFIGS_DIR/git/gitignore_global" << 'EOF'
# OS files
.DS_Store
Thumbs.db
*.swp
*~

# IDE
.vscode/
.idea/
*.iml
.project
.classpath
.settings/

# Dependencies
node_modules/
vendor/
__pycache__/
*.pyc
.env
.env.local

# Build
dist/
build/
target/
out/
*.log

# Archives
*.tar
*.tar.gz
*.zip
*.rar
EOF
    fi
    
    create_symlink "$CONFIGS_DIR/git/gitconfig" "$HOME/.gitconfig"
    create_symlink "$CONFIGS_DIR/git/gitignore_global" "$HOME/.gitignore_global"
    
    print_success "Git configuration set up"
}

# Setup terminal configurations
setup_terminal() {
    print_status "Setting up terminal configurations..."
    
    mkdir -p "$CONFIGS_DIR/terminal"
    
    # Kitty configuration
    if command -v kitty &> /dev/null; then
        mkdir -p "$HOME/.config/kitty"
        
        if [ ! -f "$CONFIGS_DIR/terminal/kitty.conf" ]; then
            cat > "$CONFIGS_DIR/terminal/kitty.conf" << 'EOF'
# Kitty Terminal Configuration

# Font
font_family      JetBrains Mono
font_size        11.0
bold_font        auto
italic_font      auto
bold_italic_font auto

# Colors (One Dark theme)
background #282c34
foreground #abb2bf
cursor #528bff
selection_background #3e4451
selection_foreground #abb2bf

# Black
color0  #282c34
color8  #5c6370

# Red
color1  #e06c75
color9  #e06c75

# Green
color2  #98c379
color10 #98c379

# Yellow
color3  #e5c07b
color11 #e5c07b

# Blue
color4  #61afef
color12 #61afef

# Magenta
color5  #c678dd
color13 #c678dd

# Cyan
color6  #56b6c2
color14 #56b6c2

# White
color7  #abb2bf
color15 #ffffff

# Window
remember_window_size  yes
initial_window_width  100
initial_window_height 30
window_padding_width 5

# Tabs
tab_bar_edge top
tab_bar_style powerline

# Performance
repaint_delay 10
input_delay 3
sync_to_monitor yes

# Bell
enable_audio_bell no
visual_bell_duration 0.1

# URLs
url_style curly
open_url_with default
detect_urls yes

# Clipboard
copy_on_select yes
strip_trailing_spaces smart

# Keyboard shortcuts
map ctrl+shift+c copy_to_clipboard
map ctrl+shift+v paste_from_clipboard
map ctrl+shift+t new_tab
map ctrl+shift+w close_tab
map ctrl+tab next_tab
map ctrl+shift+tab previous_tab
EOF
        fi
        
        create_symlink "$CONFIGS_DIR/terminal/kitty.conf" "$HOME/.config/kitty/kitty.conf"
    fi
    
    # Alacritty configuration
    if command -v alacritty &> /dev/null; then
        mkdir -p "$HOME/.config/alacritty"
        
        if [ ! -f "$CONFIGS_DIR/terminal/alacritty.yml" ]; then
            cat > "$CONFIGS_DIR/terminal/alacritty.yml" << 'EOF'
# Alacritty Terminal Configuration

window:
  padding:
    x: 5
    y: 5
  dynamic_padding: true
  decorations: full

font:
  normal:
    family: JetBrains Mono
    style: Regular
  bold:
    family: JetBrains Mono
    style: Bold
  italic:
    family: JetBrains Mono
    style: Italic
  size: 11.0

colors:
  primary:
    background: '#282c34'
    foreground: '#abb2bf'
  cursor:
    cursor: '#528bff'
  selection:
    text: '#abb2bf'
    background: '#3e4451'
  normal:
    black:   '#282c34'
    red:     '#e06c75'
    green:   '#98c379'
    yellow:  '#e5c07b'
    blue:    '#61afef'
    magenta: '#c678dd'
    cyan:    '#56b6c2'
    white:   '#abb2bf'
  bright:
    black:   '#5c6370'
    red:     '#e06c75'
    green:   '#98c379'
    yellow:  '#e5c07b'
    blue:    '#61afef'
    magenta: '#c678dd'
    cyan:    '#56b6c2'
    white:   '#ffffff'

cursor:
  style: Block
  unfocused_hollow: true

live_config_reload: true
EOF
        fi
        
        create_symlink "$CONFIGS_DIR/terminal/alacritty.yml" "$HOME/.config/alacritty/alacritty.yml"
    fi
    
    print_success "Terminal configurations set up"
}

# Setup VS Code settings
setup_vscode() {
    print_status "Setting up VS Code configuration..."
    
    if command -v code &> /dev/null; then
        mkdir -p "$CONFIGS_DIR/vscode"
        mkdir -p "$HOME/.config/Code/User"
        
        if [ ! -f "$CONFIGS_DIR/vscode/settings.json" ]; then
            cat > "$CONFIGS_DIR/vscode/settings.json" << 'EOF'
{
    "editor.fontSize": 14,
    "editor.fontFamily": "'JetBrains Mono', 'monospace'",
    "editor.fontLigatures": true,
    "editor.tabSize": 4,
    "editor.insertSpaces": true,
    "editor.wordWrap": "on",
    "editor.minimap.enabled": false,
    "editor.renderWhitespace": "trailing",
    "editor.rulers": [80, 120],
    "editor.formatOnSave": true,
    "editor.formatOnPaste": true,
    "editor.suggestSelection": "first",
    "editor.snippetSuggestions": "top",
    
    "terminal.integrated.fontSize": 13,
    "terminal.integrated.fontFamily": "'JetBrains Mono'",
    "terminal.integrated.shell.linux": "/bin/bash",
    
    "workbench.colorTheme": "One Dark Pro",
    "workbench.iconTheme": "material-icon-theme",
    "workbench.startupEditor": "none",
    
    "files.autoSave": "afterDelay",
    "files.autoSaveDelay": 1000,
    "files.trimTrailingWhitespace": true,
    "files.insertFinalNewline": true,
    
    "git.enableSmartCommit": true,
    "git.autofetch": true,
    "git.confirmSync": false,
    
    "extensions.autoUpdate": true,
    "telemetry.telemetryLevel": "off",
    
    "[python]": {
        "editor.defaultFormatter": "ms-python.black-formatter"
    },
    "[javascript]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[typescript]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[json]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[html]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[css]": {
        "editor.defaultFormatter": "esbenp.prettier-vscode"
    },
    "[markdown]": {
        "editor.defaultFormatter": "yzhang.markdown-all-in-one"
    }
}
EOF
        fi
        
        create_symlink "$CONFIGS_DIR/vscode/settings.json" "$HOME/.config/Code/User/settings.json"
        
        # Install recommended extensions
        print_status "Installing recommended VS Code extensions..."
        code --install-extension ms-python.python
        code --install-extension ms-vscode.cpptools
        code --install-extension golang.go
        code --install-extension rust-lang.rust-analyzer
        code --install-extension esbenp.prettier-vscode
        code --install-extension dbaeumer.vscode-eslint
        code --install-extension eamodio.gitlens
        code --install-extension ms-azuretools.vscode-docker
        code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
        code --install-extension hashicorp.terraform
        code --install-extension redhat.vscode-yaml
        code --install-extension zhuangtongfa.material-theme
        code --install-extension pkief.material-icon-theme
    fi
    
    print_success "VS Code configuration set up"
}

# Main execution
main() {
    print_status "Setting up dotfiles..."
    
    create_backup
    setup_shell
    setup_git
    setup_terminal
    setup_vscode
    
    print_success "Dotfiles setup completed!"
    print_info "Backup of original files saved to: $BACKUP_DIR"
    print_warning "Restart your terminal to apply shell changes"
}

main
