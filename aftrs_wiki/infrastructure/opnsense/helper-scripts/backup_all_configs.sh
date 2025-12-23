#!/bin/bash
# OPNsense Backup Coordination Script
# Coordinates backups across all OPNsense configurations

set -e

echo "🔄 Starting OPNsense backup coordination..."

# Backup AFTRS OPNsense config
if [ -d "../aftrs_opnsense_config_backup" ]; then
    echo "📁 Backing up AFTRS OPNsense configuration..."
    cd ../aftrs_opnsense_config_backup
    # Add backup commands here
    cd - > /dev/null
fi

# Backup Secretstudios OPNsense config  
if [ -d "../secretstudios_opnsense_router_backup" ]; then
    echo "📁 Backing up Secretstudios OPNsense configuration..."
    cd ../secretstudios_opnsense_router_backup
    # Add backup commands here
    cd - > /dev/null
fi

echo "✅ OPNsense backup coordination complete!"
