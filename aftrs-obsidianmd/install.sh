#!/bin/bash

set -e

# AFTRS Obsidian-Supermemory Integration Setup Script
echo "🔮 AFTRS Obsidian-Supermemory Integration Setup"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}$1${NC}"
}

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    print_error "Node.js is not installed. Please install Node.js 18+ first."
    print_status "Visit: https://nodejs.org"
    exit 1
fi

NODE_VERSION=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    print_error "Node.js version 18+ is required. Current version: $(node --version)"
    exit 1
fi

print_status "Node.js version: $(node --version) ✅"

# Check if we're in the right directory
if [ ! -f "README.md" ]; then
    print_error "Please run this script from the aftrs-obsidianmd directory"
    exit 1
fi

# Install Obsidian Importer dependencies
print_header "📦 Installing Obsidian Importer dependencies..."
cd obsidian-importer
if [ -f "package.json" ]; then
    npm install
    print_status "Obsidian Importer dependencies installed"
else
    print_warning "Obsidian Importer package.json not found - this is expected for the plugin"
fi
cd ..

# Setup CLI tool
print_header "🛠️ Setting up CLI tools..."
cd cli
npm install
chmod +x obsidian-sync.ts

# Create global symlink
if command -v npm &> /dev/null; then
    npm link
    print_status "CLI tool 'obsidian-sync' is now available globally"
else
    print_warning "Could not create global symlink. Run CLI tool directly with: node obsidian-sync.js"
fi

cd ..

# Setup environment variables
print_header "🔑 Environment Configuration"
if [ ! -f "obsidian-importer/.env.local" ]; then
    print_status "Environment file obsidian-importer/.env.local already exists"
else
    print_status "Environment file created with default configuration"
fi

# Check API key
if grep -q "sm_6ZGopX4A3gvVKq8pLjwBZH_QlEUlRUFXWpxNhWJHRLBqQAzsPsyKFwsyRAObqSCysoRbnowHKgLZiqkPDIjYAfr" obsidian-importer/.env.local; then
    print_status "Default Supermemory API key is configured"
else
    print_warning "Please update your API key in obsidian-importer/.env.local"
fi

# Ask for Obsidian vault path
print_header "🏠 Obsidian Vault Configuration"
read -p "Enter your Obsidian vault path (or press Enter to set later): " VAULT_PATH

if [ ! -z "$VAULT_PATH" ]; then
    if [ -d "$VAULT_PATH" ]; then
        export OBSIDIAN_VAULT_PATH="$VAULT_PATH"
        echo "export OBSIDIAN_VAULT_PATH=\"$VAULT_PATH\"" >> ~/.bashrc
        print_status "Vault path set to: $VAULT_PATH"
    else
        print_warning "Directory not found: $VAULT_PATH"
        print_status "You can set it later with: export OBSIDIAN_VAULT_PATH=/path/to/vault"
    fi
else
    print_status "Vault path can be set later with: export OBSIDIAN_VAULT_PATH=/path/to/vault"
fi

# Check if Obsidian is installed
print_header "🔍 Checking Obsidian Installation"
if command -v obsidian &> /dev/null; then
    print_status "Obsidian found: $(which obsidian)"
elif [ -d "/Applications/Obsidian.app" ]; then
    print_status "Obsidian found: /Applications/Obsidian.app"
elif [ -d "$HOME/.local/share/obsidian" ]; then
    print_status "Obsidian found: $HOME/.local/share/obsidian"
else
    print_warning "Obsidian not found. Please install Obsidian from https://obsidian.md"
fi

# Setup complete
print_header "✅ Setup Complete!"
echo
print_status "Available commands:"
echo "  obsidian-sync --help                           # Show CLI help"
echo "  obsidian-sync --vault-path ~/MyVault sync      # One-time sync"
echo "  obsidian-sync --auto-sync 30                   # Auto-sync every 30 min"
echo "  obsidian-sync setup-plugin                     # Create plugin config"
echo
print_status "Obsidian Importer:"
echo "  • Open Obsidian → Settings → Community Plugins → Browse"
echo "  • Search for 'Importer' and install"
echo "  • Enable the plugin and select 'Supermemory (API)'"
echo
print_status "Next steps:"
echo "  1. Set OBSIDIAN_VAULT_PATH environment variable"
echo "  2. Set SUPERMEMORY_API_KEY environment variable"
echo "  3. Run: obsidian-sync --help"
echo "  4. Try: obsidian-sync sync"
echo
print_status "Documentation: README.md"
print_status "Issues: https://github.com/aftrs-void/aftrs-obsidianmd/issues"
echo
print_header "🎉 Happy note-taking!"
