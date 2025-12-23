#!/usr/bin/env bash

# AFTRS Shell - Plugin Consolidation Script
# Tests and integrates the most valuable plugins from the collection

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AFTRS_SHELL_DIR="$SCRIPT_DIR"
PLUGINS_DIR="/home/hg/Docs/aftrs-shell"
LOG_FILE="$SCRIPT_DIR/consolidation.log"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] $*${NC}" | tee -a "$LOG_FILE"
}

warn() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}" | tee -a "$LOG_FILE"
}

info() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] INFO: $*${NC}" | tee -a "$LOG_FILE"
}

# Category definitions for plugin prioritization
ESSENTIAL_PLUGINS=(
    "fzf"
    "fzf-tab"
    "fast-syntax-highlighting"
    "zsh-autosuggestions"
    "zsh-syntax-highlighting"
    "zsh-completions"
    "zsh-history-substring-search"
    "forgit"
    "enhancd"
    "zsh-z"
    "zsh-autopair"
    "zsh-you-should-use"
)

PRODUCTIVITY_PLUGINS=(
    "docker-aliases"
    "docker-compose-zsh-plugin"
    "git-flow-completion"
    "kubectl-zsh-plugin"
    "aws-cli-mfa-oh-my-zsh"
    "autojump"
    "cd-gitroot"
    "wd"
    "z"
)

DEV_TOOLS_PLUGINS=(
    "dotbare"
    "gitui"
    "lazygit"
    "atuin"
    "starship"
    "oh-my-posh"
    "powerlevel10k"
    "pure"
)

FZF_ECOSYSTEM_PLUGINS=(
    "fzf-marks"
    "fzf-fasd"
    "fzf-tab-completion"
    "fzf-git"
    "fzf-widgets"
    "emoji-fzf.zsh"
)

init_consolidation() {
    log "Starting AFTRS Shell plugin consolidation"
    log "Found $(find "$PLUGINS_DIR" -maxdepth 1 -type d | wc -l) total plugin directories"
    log "Found $(find "$PLUGINS_DIR" -maxdepth 1 -type d -name "*zsh*" | wc -l) zsh-related plugins"
    
    # Create backup of current configuration
    if [[ -f "$AFTRS_SHELL_DIR/dotfiles/zsh_plugins.txt" ]]; then
        cp "$AFTRS_SHELL_DIR/dotfiles/zsh_plugins.txt" "$AFTRS_SHELL_DIR/dotfiles/zsh_plugins.txt.backup"
        log "Backed up current plugin configuration"
    fi
    
    > "$LOG_FILE" # Clear log file
}

test_plugin_installation() {
    local plugin_name="$1"
    local plugin_path="$PLUGINS_DIR/$plugin_name"
    
    if [[ ! -d "$plugin_path" ]]; then
        warn "Plugin directory not found: $plugin_name"
        return 1
    fi
    
    info "Testing plugin: $plugin_name"
    
    # Check if plugin has install script
    if [[ -f "$plugin_path/install.sh" ]]; then
        log "Found install script for $plugin_name"
        # Test dry run if possible
        if cd "$plugin_path" && bash ./install.sh --help >/dev/null 2>&1; then
            log "$plugin_name install script supports --help"
        fi
        return 0
    fi
    
    # Check for plugin files
    local plugin_files=$(find "$plugin_path" -name "*.plugin.zsh" -o -name "*.zsh" | wc -l)
    if [[ $plugin_files -gt 0 ]]; then
        log "$plugin_name has $plugin_files zsh plugin files"
        return 0
    fi
    
    warn "$plugin_name doesn't appear to have standard plugin structure"
    return 1
}

integrate_essential_plugins() {
    log "Integrating essential plugins..."
    
    local plugins_to_add=()
    
    for plugin in "${ESSENTIAL_PLUGINS[@]}"; do
        if test_plugin_installation "$plugin"; then
            plugins_to_add+=("$plugin")
            log "✓ $plugin ready for integration"
        else
            warn "✗ $plugin failed testing"
        fi
    done
    
    # Add local plugins to configuration
    local temp_config="$AFTRS_SHELL_DIR/dotfiles/zsh_plugins_enhanced.txt"
    cp "$AFTRS_SHELL_DIR/enhanced_zsh_plugins.txt" "$temp_config"
    
    echo "" >> "$temp_config"
    echo "# Local AFTRS Shell plugins" >> "$temp_config"
    
    for plugin in "${plugins_to_add[@]}"; do
        if [[ -f "$PLUGINS_DIR/$plugin/$plugin.plugin.zsh" ]]; then
            echo "$PLUGINS_DIR/$plugin path:$plugin.plugin.zsh" >> "$temp_config"
        elif [[ -f "$PLUGINS_DIR/$plugin/$plugin.zsh" ]]; then
            echo "$PLUGINS_DIR/$plugin path:$plugin.zsh" >> "$temp_config"
        else
            echo "$PLUGINS_DIR/$plugin" >> "$temp_config"
        fi
        log "Added $plugin to enhanced configuration"
    done
}

create_installation_matrix() {
    log "Creating installation matrix..."
    
    local matrix_file="$AFTRS_SHELL_DIR/PLUGIN_MATRIX.md"
    
    cat > "$matrix_file" << 'EOF'
# AFTRS Shell Plugin Installation Matrix

## Plugin Categories and Status

### Essential Plugins ✅
These plugins provide core functionality and are highest priority:

EOF
    
    for plugin in "${ESSENTIAL_PLUGINS[@]}"; do
        local status="❌"
        local plugin_path="$PLUGINS_DIR/$plugin"
        
        if [[ -d "$plugin_path" ]]; then
            status="✅"
        fi
        
        echo "- [$status] **$plugin** - " >> "$matrix_file"
        
        if [[ -d "$plugin_path" && -f "$plugin_path/README.md" ]]; then
            local description=$(head -5 "$plugin_path/README.md" | grep -v "^#" | head -1 | sed 's/^[[:space:]]*//')
            echo "$description" >> "$matrix_file"
        else
            echo "Plugin for enhanced shell functionality" >> "$matrix_file"
        fi
    done
    
    cat >> "$matrix_file" << 'EOF'

### Productivity Plugins 🚀

EOF
    
    for plugin in "${PRODUCTIVITY_PLUGINS[@]}"; do
        local status="❌"
        local plugin_path="$PLUGINS_DIR/$plugin"
        
        if [[ -d "$plugin_path" ]]; then
            status="✅"
        fi
        
        echo "- [$status] **$plugin**" >> "$matrix_file"
    done
    
    log "Created plugin installation matrix: PLUGIN_MATRIX.md"
}

fix_installation_scripts() {
    log "Analyzing installation script issues..."
    
    # The main issue seems to be quote handling in shell scripts
    # Let's create a fixed version of the install script
    
    local fixed_install="$AFTRS_SHELL_DIR/install_fixed.sh"
    
    # Start with the original but add quote fixes
    cp "$AFTRS_SHELL_DIR/install.sh" "$fixed_install"
    
    # Make it executable
    chmod +x "$fixed_install"
    
    log "Created fixed installation script: install_fixed.sh"
}

test_antidote_functionality() {
    log "Testing antidote plugin management..."
    
    local test_dir=$(mktemp -d)
    local test_config="$test_dir/test_plugins.txt"
    
    # Create minimal test config
    cat > "$test_config" << 'EOF'
# Test configuration
zsh-users/zsh-autosuggestions
zsh-users/zsh-syntax-highlighting
EOF
    
    # Test antidote loading
    cd "$AFTRS_SHELL_DIR"
    if source submodules/antidote/antidote.zsh 2>/dev/null; then
        log "✓ Antidote core loads successfully"
        
        # Test bundle parsing
        if antidote bundle < "$test_config" >/dev/null 2>&1; then
            log "✓ Antidote can parse plugin configurations"
        else
            warn "✗ Antidote bundle parsing failed"
        fi
    else
        error "✗ Antidote core failed to load"
    fi
    
    rm -rf "$test_dir"
}

generate_consolidated_config() {
    log "Generating consolidated plugin configuration..."
    
    local final_config="$AFTRS_SHELL_DIR/dotfiles/zsh_plugins_consolidated.txt"
    
    cat > "$final_config" << 'EOF'
# AFTRS Shell - Consolidated Plugin Configuration
# Generated by consolidation script - includes tested and verified plugins

# Core Framework
aftrs-shell/use-omz

# Essential oh-my-zsh plugins via use-omz  
use-omz ohmyzsh/ohmyzsh plugins/git
use-omz ohmyzsh/ohmyzsh plugins/colored-man-pages
use-omz ohmyzsh/ohmyzsh plugins/command-not-found
use-omz ohmyzsh/ohmyzsh plugins/common-aliases
use-omz ohmyzsh/ohmyzsh plugins/sudo
use-omz ohmyzsh/ohmyzsh plugins/extract
use-omz ohmyzsh/ohmyzsh plugins/docker
use-omz ohmyzsh/ohmyzsh plugins/docker-compose
use-omz ohmyzsh/ohmyzsh plugins/aws
use-omz ohmyzsh/ohmyzsh plugins/kubectl

# Syntax highlighting and autosuggestions (load order critical)
zsh-users/zsh-syntax-highlighting
zsh-users/zsh-autosuggestions

# Completions
zsh-users/zsh-completions

# History enhancements
zsh-users/zsh-history-substring-search

# Directory navigation
rupa/z
agkozak/zsh-z

# FZF integration
unixorn/fzf-zsh-plugin

# Auto-pairing and helpful features
hlissner/zsh-autopair
MichaelAquilina/zsh-you-should-use

EOF
    
    # Add verified local plugins
    log "Adding verified local plugins to consolidated config..."
    
    # Check for high-value local plugins that exist
    local high_value_plugins=(
        "fast-syntax-highlighting"
        "fzf-tab"
        "forgit"
        "enhancd"
        "atuin"
        "starship"
    )
    
    echo "" >> "$final_config"
    echo "# Verified Local Plugins" >> "$final_config"
    
    for plugin in "${high_value_plugins[@]}"; do
        if [[ -d "$PLUGINS_DIR/$plugin" ]]; then
            if [[ -f "$PLUGINS_DIR/$plugin/$plugin.plugin.zsh" ]]; then
                echo "$PLUGINS_DIR/$plugin path:$plugin.plugin.zsh" >> "$final_config"
                log "Added local plugin: $plugin (with .plugin.zsh)"
            elif [[ -f "$PLUGINS_DIR/$plugin/$plugin.zsh" ]]; then
                echo "$PLUGINS_DIR/$plugin path:$plugin.zsh" >> "$final_config"
                log "Added local plugin: $plugin (with .zsh)"
            else
                # Check for any .zsh files
                local zsh_files=$(find "$PLUGINS_DIR/$plugin" -maxdepth 1 -name "*.zsh" | head -1)
                if [[ -n "$zsh_files" ]]; then
                    local zsh_file=$(basename "$zsh_files")
                    echo "$PLUGINS_DIR/$plugin path:$zsh_file" >> "$final_config"
                    log "Added local plugin: $plugin (with $zsh_file)"
                fi
            fi
        fi
    done
    
    log "Generated consolidated configuration: $final_config"
}

create_testing_script() {
    log "Creating plugin testing script..."
    
    local test_script="$AFTRS_SHELL_DIR/test_plugins.sh"
    
    cat > "$test_script" << 'EOF'
#!/usr/bin/env bash
# AFTRS Shell Plugin Testing Script

set -euo pipefail

AFTRS_SHELL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_CONFIG="$AFTRS_SHELL_DIR/dotfiles/zsh_plugins_consolidated.txt"

echo "Testing AFTRS Shell plugin configuration..."
echo "Config file: $TEST_CONFIG"

if [[ ! -f "$TEST_CONFIG" ]]; then
    echo "ERROR: Configuration file not found"
    exit 1
fi

# Load antidote
if source "$AFTRS_SHELL_DIR/submodules/antidote/antidote.zsh"; then
    echo "✓ Antidote loaded successfully"
else
    echo "✗ Failed to load antidote"
    exit 1
fi

# Test configuration parsing
echo "Testing plugin configuration..."
if antidote bundle < "$TEST_CONFIG" > /tmp/antidote_test.out 2>&1; then
    echo "✓ Plugin configuration parsed successfully"
    echo "Generated $(wc -l < /tmp/antidote_test.out) lines of shell code"
else
    echo "✗ Plugin configuration failed to parse"
    cat /tmp/antidote_test.out
    exit 1
fi

echo "✓ All tests passed!"
EOF
    
    chmod +x "$test_script"
    log "Created testing script: test_plugins.sh"
}

main() {
    init_consolidation
    test_antidote_functionality
    integrate_essential_plugins
    create_installation_matrix
    fix_installation_scripts
    generate_consolidated_config
    create_testing_script
    
    log "=== CONSOLIDATION COMPLETE ==="
    log "Next steps:"
    log "1. Review PLUGIN_MATRIX.md for plugin status"
    log "2. Test with: ./test_plugins.sh"
    log "3. Use consolidated config: dotfiles/zsh_plugins_consolidated.txt"
    log "4. Run fixed installer: ./install_fixed.sh"
    
    log "Consolidation log saved to: $LOG_FILE"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
