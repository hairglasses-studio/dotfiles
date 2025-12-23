#!/bin/bash
# 🚨 CRITICAL IP Migration Fix Script
# Automatically updates hardcoded 11.1.x IPs to new 10.11.1.x subnet

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${RED}🚨 CRITICAL IP MIGRATION FIX${NC}"
echo -e "${YELLOW}⚠️  This script will update hardcoded 11.1.x IPs to 10.11.1.x subnet${NC}"
echo ""

BASE_DIR="/Users/mitch/Docs/aftrs-void"

# Backup function
backup_file() {
    local file="$1"
    local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"
    cp "$file" "$backup"
    echo -e "${BLUE}💾 Backed up: $file → $backup${NC}"
}

# Fix DNS overrides in bootstrap script
fix_dns_overrides() {
    echo -e "${YELLOW}🔧 Fixing DNS overrides in bootstrap script...${NC}"
    
    local file="$BASE_DIR/aftrs_cli/scripts/bootstrap_opnsense_agent.sh"
    if [ -f "$file" ]; then
        backup_file "$file"
        
        # Update DNS overrides
        sed -i '' 's/11\.1\.1\.3/10.11.1.3/g' "$file"
        sed -i '' 's/11\.1\.1\.4/10.11.1.4/g' "$file"
        sed -i '' 's/11\.1\.1\.5/10.11.1.5/g' "$file"
        
        echo -e "${GREEN}✅ Updated DNS overrides in bootstrap script${NC}"
        echo "   - aftrs-main.aftrs.void: 11.1.1.3 → 10.11.1.3"
        echo "   - baeblade.aftrs.void: 11.1.1.4 → 10.11.1.4"  
        echo "   - spinblade.aftrs.void: 11.1.1.5 → 10.11.1.5"
    else
        echo -e "${RED}❌ Bootstrap script not found: $file${NC}"
    fi
}

# Fix network documentation
fix_network_docs() {
    echo -e "${YELLOW}🔧 Fixing network documentation...${NC}"
    
    local file="$BASE_DIR/agentctl/infrastructure/network.md"
    if [ -f "$file" ]; then
        backup_file "$file"
        
        # Update network ranges
        sed -i '' 's/11\.1\.11\.0\/24/10.11.1.0\/24/g' "$file"
        sed -i '' 's/11\.1\.12\.0\/24/10.11.2.0\/24/g' "$file"
        sed -i '' 's/11\.1\.13\.0\/24/10.11.3.0\/24/g' "$file"
        sed -i '' 's/11\.1\.14\.0\/24/10.11.4.0\/24/g' "$file"
        
        # Update device IPs
        sed -i '' 's/11\.1\.11\.1/10.11.1.1/g' "$file"
        sed -i '' 's/11\.1\.11\.10/10.11.1.10/g' "$file"
        sed -i '' 's/11\.1\.11\.20/10.11.1.20/g' "$file"
        sed -i '' 's/11\.1\.11\.30/10.11.1.30/g' "$file"
        
        echo -e "${GREEN}✅ Updated network documentation${NC}"
        echo "   - Primary Network: 11.1.11.0/24 → 10.11.1.0/24"
        echo "   - UNRAID Server: 11.1.11.10 → 10.11.1.10"
        echo "   - Dev PC: 11.1.11.20 → 10.11.1.20"
    else
        echo -e "${RED}❌ Network documentation not found: $file${NC}"
    fi
}

# Fix Grafana dashboard URL
fix_grafana_url() {
    echo -e "${YELLOW}🔧 Fixing Grafana dashboard URL...${NC}"
    
    local file="$BASE_DIR/unraid-observability-stack/DASHBOARD_GUIDE.md"
    if [ -f "$file" ]; then
        backup_file "$file"
        
        # Update Grafana URL
        sed -i '' 's/11\.1\.11\.2:3000/10.11.1.2:3000/g' "$file"
        
        echo -e "${GREEN}✅ Updated Grafana dashboard URL${NC}"
        echo "   - Grafana: 11.1.11.2:3000 → 10.11.1.2:3000"
    else
        echo -e "${RED}❌ Grafana dashboard guide not found: $file${NC}"
    fi
}

# Search for additional hardcoded IPs
find_remaining_ips() {
    echo -e "${YELLOW}🔍 Searching for remaining hardcoded 11.1.x IPs...${NC}"
    
    cd "$BASE_DIR"
    
    # Search for remaining 11.1.x IPs
    echo -e "${BLUE}🔍 Scanning for 11.1.x patterns...${NC}"
    grep -r "11\.1\." . --include="*.sh" --include="*.py" --include="*.md" --include="*.yml" --include="*.yaml" --include="*.json" --include="*.conf" --include="*.cfg" --exclude-dir=".git" --exclude-dir=".venv" --exclude-dir="node_modules" | grep -v "citeseerx" | grep -v "\.backup\." | head -20
    
    echo ""
    echo -e "${YELLOW}⚠️  Review the above results and manually update any remaining hardcoded IPs${NC}"
}

# Generate IP migration report
generate_migration_report() {
    echo -e "${YELLOW}📊 Generating IP migration report...${NC}"
    
    local report_file="$BASE_DIR/opnsense-monolith/IP_MIGRATION_REPORT.md"
    
    cat > "$report_file" << EOF
# IP Migration Report

**Generated:** $(date '+%Y-%m-%d %H:%M:%S')
**Status:** Automated fix applied

## Changes Made

### DNS Overrides Updated
- aftrs-main.aftrs.void: 11.1.1.3 → 10.11.1.3
- baeblade.aftrs.void: 11.1.1.4 → 10.11.1.4
- spinblade.aftrs.void: 11.1.1.5 → 10.11.1.5

### Network Documentation Updated
- Primary Network: 11.1.11.0/24 → 10.11.1.0/24
- UNRAID Server: 11.1.11.10 → 10.11.1.10
- Dev PC: 11.1.11.20 → 10.11.1.20

### Monitoring Updated
- Grafana Dashboard: 11.1.11.2:3000 → 10.11.1.2:3000

## Files Modified
$(find "$BASE_DIR" -name "*.backup.*" -newer "$report_file" 2>/dev/null | head -10)

## Manual Steps Still Required

1. **CRITICAL:** Update OPNsense DHCP reservations
   - Access: http://10.1.1.1
   - Navigate: Services > DHCPv4 > LAN
   - Update UNRAID server: 11.1.2 → 10.11.1.10

2. **CRITICAL:** Update UNRAID network settings
   - Access UNRAID GUI or console
   - Change IP: 11.1.2 → 10.11.1.10
   - Gateway: 10.1.1.1
   - DNS: 10.1.1.1

3. **Update:** OPNsense DNS overrides via GUI
   - Navigate: Services > Unbound DNS > Overrides
   - Apply the DNS changes listed above

4. **Restart:** Network services
   - Restart OPNsense DHCP and DNS services
   - Reboot UNRAID server after IP change

## Verification Commands

\`\`\`bash
# Test new UNRAID connectivity
ping 10.11.1.10

# Test Grafana dashboard
curl -I http://10.11.1.2:3000

# Test DNS resolution
nslookup aftrs-main.aftrs.void 10.1.1.1
\`\`\`

---
*Report generated by IP Migration Fix Script*
EOF

    echo -e "${GREEN}✅ Migration report created: $report_file${NC}"
}

# Main execution
main() {
    echo -e "${RED}🚨 STARTING CRITICAL IP MIGRATION FIX${NC}"
    echo ""
    
    # Check if we're in the right directory
    if [ ! -d "$BASE_DIR" ]; then
        echo -e "${RED}❌ Base directory not found: $BASE_DIR${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}📁 Working directory: $BASE_DIR${NC}"
    echo ""
    
    # Execute fixes
    fix_dns_overrides
    echo ""
    
    fix_network_docs  
    echo ""
    
    fix_grafana_url
    echo ""
    
    find_remaining_ips
    echo ""
    
    generate_migration_report
    echo ""
    
    echo -e "${GREEN}🎉 Automated IP migration fixes completed!${NC}"
    echo ""
    echo -e "${RED}⚠️  MANUAL STEPS STILL REQUIRED:${NC}"
    echo -e "   1. Update OPNsense DHCP reservations (CRITICAL)"
    echo -e "   2. Reconfigure UNRAID server IP (CRITICAL)"  
    echo -e "   3. Update OPNsense DNS overrides"
    echo -e "   4. Restart network services"
    echo ""
    echo -e "${BLUE}📋 See IP_MIGRATION_REPORT.md for complete details${NC}"
}

# Execute main function
main "$@"
