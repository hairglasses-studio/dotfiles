#!/bin/bash

# Unraid Backup System Setup Script
# Prepares the backup and analysis tools for use

set -euo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up Unraid Configuration Backup System${NC}"
echo "=================================================="

# Make scripts executable
echo -e "${GREEN}📝 Making scripts executable...${NC}"
chmod +x backup_unraid.sh
chmod +x backup_unraid_enhanced.sh
chmod +x analyze_unraid.py
chmod +x analyze_unraid_enhanced.py
chmod +x monitor_unraid.py

# Create backup directory
echo -e "${GREEN}📁 Creating backup directory...${NC}"
mkdir -p backups

# Check Python dependencies
echo -e "${GREEN}🐍 Checking Python dependencies...${NC}"
python3 -c "import json, re, argparse, pathlib, yaml, statistics" 2>/dev/null || {
    echo -e "${YELLOW}⚠️  Warning: Some Python modules may not be available${NC}"
    echo "   The enhanced analysis script requires: json, re, argparse, pathlib, yaml, statistics"
    echo "   Installing PyYAML if needed..."
    pip3 install PyYAML 2>/dev/null || echo "   PyYAML installation failed - some features may be limited"
}

# Check system requirements
echo -e "${GREEN}🔍 Checking system requirements...${NC}"

# Check if running on Unraid
if [ -f "/proc/version" ] && grep -q "unraid" /proc/version 2>/dev/null; then
    echo -e "${GREEN}✅ Running on Unraid system${NC}"
else
    echo -e "${YELLOW}⚠️  Not detected as Unraid system - some features may be limited${NC}"
fi

# Check for Docker
if command -v docker >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker available${NC}"
else
    echo -e "${YELLOW}⚠️  Docker not found - Docker backup features will be skipped${NC}"
fi

# Check for libvirt (VM support)
if command -v virsh >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Libvirt available for VM backup${NC}"
else
    echo -e "${YELLOW}⚠️  Libvirt not found - VM backup features will be skipped${NC}"
fi

# Check for performance tools
if command -v iostat >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Performance monitoring tools available${NC}"
else
    echo -e "${YELLOW}⚠️  iostat not found - some performance metrics will be limited${NC}"
fi

# Create .gitignore for the project
echo -e "${GREEN}📋 Creating .gitignore...${NC}"
cat > .gitignore << 'EOF'
# Backup directories (large files)
backups/

# Analysis output files
*.json

# Temporary files
*.tmp
*.log

# Python cache
__pycache__/
*.pyc
*.pyo

# Sensitive data
*.ini
*.conf
EOF

# Create a quick test script
echo -e "${GREEN}🧪 Creating test script...${NC}"
cat > test_backup.sh << 'EOF'
#!/bin/bash

# Quick test of backup functionality
echo "🧪 Testing enhanced backup system..."

# Test backup script syntax
bash -n backup_unraid.sh && echo "✅ Basic backup script syntax OK" || echo "❌ Basic backup script syntax error"
bash -n backup_unraid_enhanced.sh && echo "✅ Enhanced backup script syntax OK" || echo "❌ Enhanced backup script syntax error"

# Test analysis script syntax
python3 -m py_compile analyze_unraid.py && echo "✅ Basic analysis script syntax OK" || echo "❌ Basic analysis script syntax error"
python3 -m py_compile analyze_unraid_enhanced.py && echo "✅ Enhanced analysis script syntax OK" || echo "❌ Enhanced analysis script syntax error"
python3 -m py_compile monitor_unraid.py && echo "✅ Monitoring script syntax OK" || echo "❌ Monitoring script syntax error"

# Test if we can create backup directory
mkdir -p test_backup && echo "✅ Can create backup directories" || echo "❌ Cannot create backup directories"
rmdir test_backup

# Test configuration file
if [ -f "config.yaml" ]; then
    echo "✅ Configuration file exists"
else
    echo "⚠️  Configuration file not found - will use defaults"
fi

echo "🧪 Enhanced test completed!"
EOF

chmod +x test_backup.sh

echo ""
echo -e "${BLUE}✅ Setup completed successfully!${NC}"
echo ""
echo -e "${GREEN}📋 Next steps:${NC}"
echo "1. Run: sudo ./backup_unraid_enhanced.sh (recommended)"
echo "2. Analyze: python3 analyze_unraid_enhanced.py backups/unraid_backup_*/ --html report.html"
echo "3. Monitor: python3 monitor_unraid.py --once (test monitoring)"
echo "4. Test: ./test_backup.sh"
echo ""
echo -e "${YELLOW}💡 Tips:${NC}"
echo "- Run as root (sudo) for complete system access"
echo "- The enhanced backup creates both full and git-friendly versions"
echo "- Use config.yaml to customize backup and monitoring behavior"
echo "- Check the README.md for detailed usage instructions"
echo "- Monitor your system continuously with the monitoring tool"
echo ""
echo -e "${BLUE}🚀 Ready to optimize your Unraid server!${NC}" 