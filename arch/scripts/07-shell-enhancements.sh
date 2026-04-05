#!/bin/bash

#############################################
# 07 - Shell Enhancements
# Advanced shell configuration with Starship
#############################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

print_status() { echo -e "${BLUE}==>${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
print_info() { echo -e "${CYAN}ℹ${NC} $1"; }

# Script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
CONFIGS_DIR="$REPO_DIR/configs"

# Check if package is installed
is_installed() {
    pacman -Qi "$1" &> /dev/null
}

# Install modern shell (Zsh)
install_zsh() {
    print_status "Installing Zsh and Oh My Zsh..."
    
    # Install Zsh
    if ! is_installed zsh; then
        sudo pacman -S --needed --noconfirm zsh zsh-completions
        print_success "Zsh installed"
    else
        print_success "Zsh already installed"
    fi
    
    # Install Oh My Zsh
    if [ ! -d "$HOME/.oh-my-zsh" ]; then
        print_status "Installing Oh My Zsh..."
        sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended
        print_success "Oh My Zsh installed"
    else
        print_success "Oh My Zsh already installed"
    fi
    
    # Install Zsh plugins
    print_status "Installing Zsh plugins..."
    
    # zsh-autosuggestions
    if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-autosuggestions" ]; then
        git clone https://github.com/zsh-users/zsh-autosuggestions \
            ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
    fi
    
    # zsh-syntax-highlighting
    if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting" ]; then
        git clone https://github.com/zsh-users/zsh-syntax-highlighting \
            ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting
    fi
    
    # zsh-history-substring-search
    if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-history-substring-search" ]; then
        git clone https://github.com/zsh-users/zsh-history-substring-search \
            ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/zsh-history-substring-search
    fi
    
    # fzf-tab
    if [ ! -d "${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/fzf-tab" ]; then
        git clone https://github.com/Aloxaf/fzf-tab \
            ${ZSH_CUSTOM:-$HOME/.oh-my-zsh/custom}/plugins/fzf-tab
    fi
    
    print_success "Zsh plugins installed"
}

# Install Fish shell
install_fish() {
    echo -n "Install Fish shell? (y/N): "
    read -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if ! is_installed fish; then
            sudo pacman -S --needed --noconfirm fish fisher
            print_success "Fish shell installed"
            
            # Install Fisher plugins
            fish -c "fisher install jorgebucaran/fisher"
            fish -c "fisher install jethrokuan/z"
            fish -c "fisher install PatrickF1/fzf.fish"
        else
            print_success "Fish shell already installed"
        fi
    fi
}

# Install Starship prompt
install_starship() {
    print_status "Installing Starship prompt..."
    
    if ! command -v starship &> /dev/null; then
        print_status "Downloading and installing Starship..."
        curl -sS https://starship.rs/install.sh | sh -s -- --yes
        print_success "Starship installed"
    else
        print_success "Starship already installed"
    fi
    
    # Install Starship configuration
    print_status "Installing Starship configuration..."
    mkdir -p "$HOME/.config"
    
    # Create comprehensive Starship config
    if [ ! -f "$HOME/.config/starship.toml" ]; then
        cat > "$HOME/.config/starship.toml" << 'EOF'
# Starship Configuration - Arch Linux

# Format
format = """
$os\
$username\
$hostname\
$directory\
$git_branch\
$git_status\
$git_state\
$git_metrics\
$docker_context\
$package\
$golang\
$helm\
$java\
$kotlin\
$nodejs\
$php\
$python\
$ruby\
$rust\
$terraform\
$aws\
$gcloud\
$kubernetes\
$cmd_duration\
$line_break\
$jobs\
$battery\
$time\
$status\
$character"""

# Timeout
command_timeout = 1000
scan_timeout = 30
add_newline = true

[os]
disabled = false
format = "[$symbol](bold white) "

[os.symbols]
Arch = " "
Ubuntu = " "
Macos = " "
Windows = " "
Linux = " "

[username]
style_user = "bold cyan"
style_root = "bold red"
format = "[$user]($style) "
disabled = false
show_always = false

[hostname]
ssh_only = true
format = "[@$hostname](bold yellow) "
disabled = false

[directory]
truncation_length = 3
truncate_to_repo = true
format = "[$path]($style)[$read_only]($read_only_style) "
style = "bold blue"
read_only = " "
read_only_style = "red"
home_symbol = " ~"

[directory.substitutions]
"Documents" = " "
"Downloads" = " "
"Music" = " "
"Pictures" = " "
"Projects" = " "

[git_branch]
symbol = " "
format = "[$symbol$branch]($style) "
style = "bold purple"

[git_status]
format = '([\[$all_status$ahead_behind\]]($style) )'
style = "bold red"
conflicted = "⚔️ "
ahead = "⬆️ "
behind = "⬇️ "
diverged = "🔀 "
up_to_date = ""
untracked = "🤷 "
stashed = "📦 "
modified = "📝 "
staged = '[ ++\($count\)](green)'
renamed = "✏️ "
deleted = "🗑️ "

[git_state]
format = '\([$state( $progress_current/$progress_total)]($style)\) '
style = "bright-black"

[git_metrics]
disabled = false
added_style = "bold green"
deleted_style = "bold red"
only_nonzero_diffs = true
format = '([+$added]($added_style) )([-$deleted]($deleted_style) )'

[docker_context]
symbol = " "
style = "bold blue"
format = "[$symbol$context]($style) "
only_with_files = true
disabled = false
detect_extensions = []
detect_files = ["docker-compose.yml", "docker-compose.yaml", "Dockerfile"]
detect_folders = []

[kubernetes]
format = '[$symbol$context( \($namespace\))]($style) '
symbol = "☸ "
style = "bold cyan"
disabled = false

[terraform]
format = "[$symbol$workspace]($style) "
symbol = "💠 "
style = "bold purple"

[aws]
format = '[$symbol($profile )(\($region\) )]($style)'
style = "bold yellow"
symbol = "☁️ "
disabled = false

[gcloud]
format = '[$symbol$account(@$domain)(\($region\))]($style) '
symbol = "☁️ "
style = "bold blue"
disabled = false

[package]
format = "[$symbol$version]($style) "
symbol = "📦 "
style = "208"
display_private = false
disabled = false

# Programming Languages
[golang]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold cyan"

[java]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold red"

[kotlin]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold purple"

[nodejs]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold green"
not_capable_style = "bold red"

[php]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold blue"

[python]
format = '[${symbol}${pyenv_prefix}(${version} )(\($virtualenv\) )]($style)'
symbol = " "
style = "bold yellow"
pyenv_version_name = false
python_binary = ["python3", "python"]

[ruby]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold red"

[rust]
format = "[$symbol($version )]($style)"
symbol = " "
style = "bold red"

[helm]
format = "[$symbol($version )]($style)"
symbol = "⎈ "
style = "bold white"

# Utilities
[cmd_duration]
min_time = 2_000
format = "[ took $duration]($style) "
style = "bold yellow"
show_milliseconds = false

[jobs]
symbol = "✦ "
style = "bold blue"
number_threshold = 1
format = "[$symbol$number]($style) "

[battery]
full_symbol = "🔋 "
charging_symbol = "⚡ "
discharging_symbol = "💀 "
unknown_symbol = "❓ "
empty_symbol = "🪫 "
format = "[$symbol$percentage]($style) "

[[battery.display]]
threshold = 10
style = "bold red"

[[battery.display]]
threshold = 30
style = "bold yellow"

[[battery.display]]
threshold = 100
style = "bold green"

[time]
disabled = false
format = '[$time]($style) '
style = "bold yellow"
time_format = "%T"
utc_time_offset = "local"

[status]
style = "bold red"
symbol = "✖"
success_symbol = ""
format = '[$symbol $common_meaning$signal_name$maybe_int]($style) '
map_symbol = true
disabled = false

[character]
success_symbol = "[❯](bold green)"
error_symbol = "[❯](bold red)"
vimcmd_symbol = "[❮](bold green)"
vimcmd_replace_one_symbol = "[❮](bold purple)"
vimcmd_replace_symbol = "[❮](bold purple)"
vimcmd_visual_symbol = "[❮](bold yellow)"
EOF
        print_success "Starship configuration created"
    else
        print_warning "Starship configuration already exists"
    fi
    
    # Configure shells for Starship
    print_status "Configuring shells for Starship..."
    
    # Bash
    if [ -f "$HOME/.bashrc" ] && ! grep -q "starship init bash" "$HOME/.bashrc"; then
        echo '' >> "$HOME/.bashrc"
        echo '# Starship prompt' >> "$HOME/.bashrc"
        echo 'eval "$(starship init bash)"' >> "$HOME/.bashrc"
        print_success "Starship configured for Bash"
    fi
    
    # Zsh
    if [ -f "$HOME/.zshrc" ] && ! grep -q "starship init zsh" "$HOME/.zshrc"; then
        echo '' >> "$HOME/.zshrc"
        echo '# Starship prompt' >> "$HOME/.zshrc"
        echo 'eval "$(starship init zsh)"' >> "$HOME/.zshrc"
        print_success "Starship configured for Zsh"
    fi
    
    # Fish
    if command -v fish &> /dev/null; then
        mkdir -p "$HOME/.config/fish"
        if [ -f "$HOME/.config/fish/config.fish" ] && ! grep -q "starship init fish" "$HOME/.config/fish/config.fish"; then
            echo '' >> "$HOME/.config/fish/config.fish"
            echo '# Starship prompt' >> "$HOME/.config/fish/config.fish"
            echo 'starship init fish | source' >> "$HOME/.config/fish/config.fish"
            print_success "Starship configured for Fish"
        fi
    fi
}

# Install Nerd Fonts
install_nerd_fonts() {
    print_status "Installing Nerd Fonts..."
    
    # Create fonts directory
    mkdir -p "$HOME/.local/share/fonts"
    
    echo -n "Install FiraCode Nerd Font? (Y/n): "
    read -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        print_status "Downloading FiraCode Nerd Font..."
        cd /tmp
        wget -q https://github.com/ryanoasis/nerd-fonts/releases/latest/download/FiraCode.zip
        unzip -q FiraCode.zip -d FiraCode
        cp FiraCode/*.ttf "$HOME/.local/share/fonts/"
        rm -rf FiraCode FiraCode.zip
        print_success "FiraCode Nerd Font installed"
    fi
    
    echo -n "Install JetBrains Mono Nerd Font? (Y/n): "
    read -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        print_status "Downloading JetBrains Mono Nerd Font..."
        cd /tmp
        wget -q https://github.com/ryanoasis/nerd-fonts/releases/latest/download/JetBrainsMono.zip
        unzip -q JetBrainsMono.zip -d JetBrainsMono
        cp JetBrainsMono/*.ttf "$HOME/.local/share/fonts/"
        rm -rf JetBrainsMono JetBrainsMono.zip
        print_success "JetBrains Mono Nerd Font installed"
    fi
    
    # Update font cache
    if command -v fc-cache &> /dev/null; then
        print_status "Updating font cache..."
        fc-cache -f "$HOME/.local/share/fonts"
        print_success "Font cache updated"
    fi
    
    print_warning "Configure your terminal to use a Nerd Font for best results"
}

# Install modern CLI tools
install_cli_tools() {
    print_status "Installing modern CLI tools..."
    
    # Essential modern CLI replacements
    local tools=(
        "bat"           # Better cat
        "exa"           # Better ls
        "fd"            # Better find
        "ripgrep"       # Better grep
        "fzf"           # Fuzzy finder
        "zoxide"        # Smart cd
        "delta"         # Better git diff
        "duf"           # Better df
        "dust"          # Better du
        "procs"         # Better ps
        "bottom"        # Better top
        "httpie"        # Better curl
        "jq"            # JSON processor
        "yq"            # YAML processor
        "tldr"          # Simplified man pages
        "thefuck"       # Command correction
        "mcfly"         # Better shell history
        "atuin"         # Enhanced shell history
    )
    
    for tool in "${tools[@]}"; do
        local package_name=$tool
        
        # Handle special package names
        case $tool in
            "exa") package_name="exa" ;;
            "delta") package_name="git-delta" ;;
            "mcfly") package_name="mcfly" ;;
            "atuin") 
                if ! command -v atuin &> /dev/null; then
                    print_status "Installing atuin..."
                    yay -S --needed --noconfirm atuin-bin
                    print_success "atuin installed"
                fi
                continue
                ;;
        esac
        
        if ! is_installed "$package_name" && ! command -v "$tool" &> /dev/null; then
            print_status "Installing $tool..."
            if sudo pacman -Si "$package_name" &> /dev/null; then
                sudo pacman -S --needed --noconfirm "$package_name"
            else
                yay -S --needed --noconfirm "$package_name"
            fi
            print_success "$tool installed"
        else
            print_success "$tool already installed"
        fi
    done
    
    # Configure tools
    configure_cli_tools
}

# Configure CLI tools
configure_cli_tools() {
    print_status "Configuring CLI tools..."
    
    # Create config directory
    mkdir -p "$HOME/.config"
    
    # Configure bat
    if command -v bat &> /dev/null; then
        mkdir -p "$HOME/.config/bat"
        if [ ! -f "$HOME/.config/bat/config" ]; then
            cat > "$HOME/.config/bat/config" << 'EOF'
# Bat configuration
--theme="OneHalfDark"
--style="numbers,changes,header"
--italic-text=always
--map-syntax "*.jenkinsfile:Groovy"
--map-syntax "*.js:JavaScript"
--map-syntax "*.ts:TypeScript"
EOF
            print_success "bat configured"
        fi
    fi
    
    # Configure git delta
    if command -v delta &> /dev/null; then
        git config --global core.pager "delta"
        git config --global interactive.diffFilter "delta --color-only"
        git config --global delta.navigate true
        git config --global delta.light false
        git config --global delta.line-numbers true
        git config --global delta.side-by-side true
        print_success "git delta configured"
    fi
    
    # Configure zoxide
    if command -v zoxide &> /dev/null; then
        # Bash
        if [ -f "$HOME/.bashrc" ] && ! grep -q "zoxide init bash" "$HOME/.bashrc"; then
            echo 'eval "$(zoxide init bash)"' >> "$HOME/.bashrc"
        fi
        
        # Zsh
        if [ -f "$HOME/.zshrc" ] && ! grep -q "zoxide init zsh" "$HOME/.zshrc"; then
            echo 'eval "$(zoxide init zsh)"' >> "$HOME/.zshrc"
        fi
        
        # Fish
        if [ -f "$HOME/.config/fish/config.fish" ] && ! grep -q "zoxide init fish" "$HOME/.config/fish/config.fish"; then
            echo 'zoxide init fish | source' >> "$HOME/.config/fish/config.fish"
        fi
        
        print_success "zoxide configured"
    fi
    
    # Configure fzf
    if command -v fzf &> /dev/null; then
        # Set FZF defaults
        export FZF_DEFAULT_OPTS='
            --height 40%
            --layout=reverse
            --border
            --color=dark
            --color=fg:#c0c5ce,bg:#2b303b,hl:#65737e
            --color=fg+:#c0caf5,bg+:#292e42,hl+:#bb9af7
            --color=info:#7dcfff,prompt:#bb9af7,pointer:#bb9af7
            --color=marker:#9ece6a,spinner:#9ece6a,header:#9ece6a'
        
        # Add to shell configs
        for rc_file in "$HOME/.bashrc" "$HOME/.zshrc"; do
            if [ -f "$rc_file" ] && ! grep -q "FZF_DEFAULT_OPTS" "$rc_file"; then
                echo "export FZF_DEFAULT_OPTS='$FZF_DEFAULT_OPTS'" >> "$rc_file"
            fi
        done
        
        print_success "fzf configured"
    fi
    
    # Configure atuin
    if command -v atuin &> /dev/null; then
        # Bash
        if [ -f "$HOME/.bashrc" ] && ! grep -q "atuin init bash" "$HOME/.bashrc"; then
            echo '[[ -f ~/.bash-preexec.sh ]] && source ~/.bash-preexec.sh' >> "$HOME/.bashrc"
            echo 'eval "$(atuin init bash)"' >> "$HOME/.bashrc"
        fi
        
        # Zsh
        if [ -f "$HOME/.zshrc" ] && ! grep -q "atuin init zsh" "$HOME/.zshrc"; then
            echo 'eval "$(atuin init zsh)"' >> "$HOME/.zshrc"
        fi
        
        print_success "atuin configured"
    fi
}

# Create shell aliases
create_aliases() {
    print_status "Creating modern shell aliases..."
    
    local aliases_file="$HOME/.shell_aliases"
    
    cat > "$aliases_file" << 'EOF'
# Modern CLI Aliases

# Better defaults
alias ls='exa --icons --group-directories-first'
alias ll='exa -l --icons --group-directories-first'
alias la='exa -la --icons --group-directories-first'
alias lt='exa --tree --icons'
alias tree='exa --tree --icons'

alias cat='bat --paging=never'
alias less='bat'
alias grep='rg'
alias find='fd'
alias ps='procs'
alias top='btm'
alias htop='btm'
alias df='duf'
alias du='dust'
alias cd='z'

# Git aliases with delta
alias gd='git diff'
alias gds='git diff --staged'
alias gl='git log --oneline --graph --decorate'

# Safety aliases
alias rm='rm -i'
alias cp='cp -i'
alias mv='mv -i'

# Shortcuts
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias ~='cd ~'

# System
alias update='yay -Syu'
alias install='yay -S'
alias search='yay -Ss'
alias remove='yay -R'
alias orphans='yay -Qtdq | yay -Rns -'

# Docker
alias d='docker'
alias dc='docker-compose'
alias dps='docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"'
alias dex='docker exec -it'

# Kubernetes
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kaf='kubectl apply -f'
alias kdel='kubectl delete'
alias klog='kubectl logs'

# Quick edits
alias zshrc='${EDITOR:-vim} ~/.zshrc'
alias bashrc='${EDITOR:-vim} ~/.bashrc'
alias vimrc='${EDITOR:-vim} ~/.vimrc'
alias aliases='${EDITOR:-vim} ~/.shell_aliases'

# Useful functions
mkcd() { mkdir -p "$1" && cd "$1"; }
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

# FZF shortcuts
if command -v fzf &> /dev/null; then
    # Interactive file search and open
    fo() {
        local file
        file=$(fzf --preview 'bat --style=numbers --color=always {}' --preview-window=right:60%)
        [ -n "$file" ] && ${EDITOR:-vim} "$file"
    }
    
    # Interactive directory change
    fcd() {
        local dir
        dir=$(fd --type d | fzf --preview 'exa --tree --level=2 {}' --preview-window=right:40%)
        [ -n "$dir" ] && cd "$dir"
    }
    
    # Git branch switch
    fgb() {
        local branch
        branch=$(git branch -a | fzf | sed 's/^[* ]*//' | sed 's/^remotes\///')
        [ -n "$branch" ] && git checkout "$branch"
    }
    
    # Process kill
    fkill() {
        local pid
        pid=$(ps aux | fzf | awk '{print $2}')
        [ -n "$pid" ] && kill -9 "$pid"
    }
fi

# Clipboard
if command -v xclip &> /dev/null; then
    alias pbcopy='xclip -selection clipboard'
    alias pbpaste='xclip -selection clipboard -o'
fi
EOF
    
    # Source aliases in shell configs
    for rc_file in "$HOME/.bashrc" "$HOME/.zshrc"; do
        if [ -f "$rc_file" ] && ! grep -q "shell_aliases" "$rc_file"; then
            echo '' >> "$rc_file"
            echo '# Shell aliases' >> "$rc_file"
            echo '[ -f ~/.shell_aliases ] && source ~/.shell_aliases' >> "$rc_file"
        fi
    done
    
    print_success "Shell aliases created"
}

# Install and configure tmux
install_tmux() {
    print_status "Installing tmux and configuration..."
    
    if ! is_installed tmux; then
        sudo pacman -S --needed --noconfirm tmux
        print_success "tmux installed"
    else
        print_success "tmux already installed"
    fi
    
    # Install tmux plugin manager
    if [ ! -d "$HOME/.tmux/plugins/tpm" ]; then
        print_status "Installing Tmux Plugin Manager..."
        git clone https://github.com/tmux-plugins/tpm "$HOME/.tmux/plugins/tpm"
        print_success "TPM installed"
    fi
    
    # Create tmux configuration
    if [ ! -f "$HOME/.tmux.conf" ]; then
        cat > "$HOME/.tmux.conf" << 'EOF'
# Tmux Configuration

# Prefix key
unbind C-b
set -g prefix C-a
bind C-a send-prefix

# Options
set -g mouse on
set -g base-index 1
setw -g pane-base-index 1
set -g renumber-windows on
set -g history-limit 50000
set -g display-time 4000
set -g status-interval 5
set -g default-terminal "screen-256color"
set -g focus-events on
setw -g aggressive-resize on
set -sg escape-time 0

# Key bindings
bind r source-file ~/.tmux.conf \; display "Config reloaded!"
bind | split-window -h -c "#{pane_current_path}"
bind - split-window -v -c "#{pane_current_path}"
bind h select-pane -L
bind j select-pane -D
bind k select-pane -U
bind l select-pane -R
bind -r H resize-pane -L 5
bind -r J resize-pane -D 5
bind -r K resize-pane -U 5
bind -r L resize-pane -R 5

# Copy mode
setw -g mode-keys vi
bind-key -T copy-mode-vi v send-keys -X begin-selection
bind-key -T copy-mode-vi y send-keys -X copy-pipe-and-cancel 'xclip -in -selection clipboard'

# Status bar
set -g status-position top
set -g status-style 'bg=#1e1e2e fg=#cdd6f4'
set -g status-left '#[fg=#89b4fa,bold]  #S  '
set -g status-right '#[fg=#f38ba8]  %Y-%m-%d  %H:%M '
set -g status-left-length 50
set -g status-right-length 50

# Window status
setw -g window-status-current-style 'fg=#1e1e2e bg=#89b4fa bold'
setw -g window-status-current-format ' #I:#W#F '
setw -g window-status-style 'fg=#cdd6f4 bg=#313244'
setw -g window-status-format ' #I:#W#F '

# Pane borders
set -g pane-border-style 'fg=#313244'
set -g pane-active-border-style 'fg=#89b4fa'

# Messages
set -g message-style 'fg=#1e1e2e bg=#f9e2af bold'

# Plugins
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-sensible'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'
set -g @plugin 'tmux-plugins/tmux-yank'
set -g @plugin 'tmux-plugins/tmux-prefix-highlight'

# Plugin settings
set -g @resurrect-capture-pane-contents 'on'
set -g @continuum-restore 'on'
set -g @continuum-boot 'on'
set -g @prefix_highlight_show_copy_mode 'on'
set -g @prefix_highlight_show_sync_mode 'on'

# Initialize TPM (keep at bottom)
run '~/.tmux/plugins/tpm/tpm'
EOF
        print_success "tmux configuration created"
    else
        print_warning "tmux configuration already exists"
    fi
}

# Main execution
main() {
    print_status "Setting up shell enhancements..."
    
    # Check if running in auto mode
    if [ "${AUTO_INSTALL:-false}" = "true" ]; then
        print_info "Running in automatic mode - installing all enhancements"
        
        install_zsh
        install_starship
        install_nerd_fonts
        install_cli_tools
        create_aliases
        install_tmux
        
        # Skip Fish in auto mode
        print_info "Skipping Fish shell in automatic mode"
    else
        # Interactive mode
        echo ""
        echo "Select shell enhancements to install:"
        echo ""
        
        echo -n "Install Zsh with Oh My Zsh and plugins? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && install_zsh
        
        echo -n "Install Fish shell? (y/N): "
        read -n 1 -r
        echo
        [[ $REPLY =~ ^[Yy]$ ]] && install_fish
        
        echo -n "Install Starship prompt? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && install_starship
        
        echo -n "Install Nerd Fonts? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && install_nerd_fonts
        
        echo -n "Install modern CLI tools (bat, exa, fzf, etc.)? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && install_cli_tools
        
        echo -n "Create modern shell aliases? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && create_aliases
        
        echo -n "Install and configure tmux? (Y/n): "
        read -n 1 -r
        echo
        [[ ! $REPLY =~ ^[Nn]$ ]] && install_tmux
    fi
    
    print_success "Shell enhancements setup completed!"
    print_warning "Restart your terminal or source your shell config to apply changes"
    print_info "You may want to change your default shell with: chsh -s /usr/bin/zsh"
}

main
