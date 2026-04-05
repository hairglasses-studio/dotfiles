#!/bin/bash

#############################################
# Git Repository Initialization Script
# Sets up the dotfiles-arch repository
#############################################

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Initializing Arch Linux Dotfiles Repository${NC}"
echo "==========================================="
echo ""

# Initialize git repository
if [ ! -d .git ]; then
    git init
    echo -e "${GREEN}✓${NC} Initialized git repository"
else
    echo -e "${GREEN}✓${NC} Git repository already initialized"
fi

# Make all scripts executable
chmod +x install.sh
chmod +x scripts/*.sh
chmod +x backup/*.sh
echo -e "${GREEN}✓${NC} Made scripts executable"

# Set up git remote
read -p "Add GitHub remote? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git remote add origin https://github.com/hairglasses/dotfiles-arch.git
    echo -e "${GREEN}✓${NC} Added GitHub remote"
fi

# Create initial commit
read -p "Create initial commit? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git add .
    git commit -m "Initial commit: Arch Linux dotfiles and installation scripts

- Complete installation automation for Arch Linux
- Essential and extended application installers
- Development environment setup
- System configuration and optimization
- Dotfiles management system
- Backup and restore functionality
- Comprehensive documentation"
    echo -e "${GREEN}✓${NC} Created initial commit"
fi

echo ""
echo -e "${GREEN}Repository setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review and customize scripts for your needs"
echo "  2. Run: ./install.sh to start installation"
echo "  3. Push to GitHub: git push -u origin main"
echo ""
echo "For help, see README.md and docs/"
