# 🚨 CRITICAL: IP Subnet Migration Analysis

**URGENT NETWORK ISSUE DETECTED**
**Date:** 2025-09-23 00:38:00
**Status:** ACTIVE CRISIS - Devices stuck on old subnet

## 🔍 **Crisis Summary**

**Problem:** OPNsense router moved from `11.1.1.1` to `10.1.1.1` with new subnet `10.11.1.0/24`, but multiple devices still configured for old `11.1.x` ranges.

## 📊 **Discovered Network Configurations**

### **Current Router Status**
- **NEW Router IP:** `10.1.1.1` 
- **NEW Subnet:** `10.11.1.0/24`
- **OLD Router IP:** `11.1.1.1` (deprecated)
- **OLD Subnet:** `11.1.1.0/24` (deprecated)

### **Devices Stuck on Old Subnet (CONFIRMED)**

#### **1. UNRAID Server**
- **Current IP:** `11.1.2` (STUCK - user confirmed)
- **Target IP:** `10.11.1.2`
- **Status:** 🚨 CRITICAL - Server not reachable on new subnet
- **Found in:** User report, likely DHCP reservation or static config

#### **2. Grafana Dashboard** 
- **Current IP:** `11.1.11.2:3000` 
- **Target IP:** `10.11.1.2:3000`
- **Status:** 🚨 CRITICAL - Monitoring dashboard unreachable
- **Found in:** `unraid-observability-stack/DASHBOARD_GUIDE.md`

#### **3. DNS Domain Overrides (HARDCODED)**
**Found in:** `aftrs_cli/scripts/bootstrap_opnsense_agent.sh`
- **aftrs-main.aftrs.void:** `11.1.1.3` → **SHOULD BE:** `10.11.1.3`
- **baeblade.aftrs.void:** `11.1.1.4` → **SHOULD BE:** `10.11.1.4`  
- **spinblade.aftrs.void:** `11.1.1.5` → **SHOULD BE:** `10.11.1.5`

#### **4. Network Documentation (OUTDATED)**
**Found in:** `agentctl/infrastructure/network.md`
- **Primary Network:** `11.1.11.0/24` → **SHOULD BE:** `10.11.1.0/24`
- **UNRAID Server:** `11.1.11.10` → **SHOULD BE:** `10.11.1.10`
- **Dev PC:** `11.1.11.20` → **SHOULD BE:** `10.11.1.20`

## 🎯 **IMMEDIATE ACTION REQUIRED**

### **Step 1: Update OPNsense DHCP Reservations**
```bash
# Access OPNsense Web Interface at 10.1.1.1
# Navigate to: Services > DHCPv4 > LAN
# Update static DHCP reservations:

UNRAID Server: 
  - OLD: 11.1.2 or 11.1.11.10
  - NEW: 10.11.1.10
  - MAC: [Find in OPNsense DHCP leases]

Grafana/Monitoring:
  - OLD: 11.1.11.2  
  - NEW: 10.11.1.2
```

### **Step 2: Update DNS Overrides**
```bash
# Navigate to: Services > Unbound DNS > Overrides
# Update domain overrides:

aftrs-main.aftrs.void: 11.1.1.3 → 10.11.1.3
baeblade.aftrs.void: 11.1.1.4 → 10.11.1.4
spinblade.aftrs.void: 11.1.1.5 → 10.11.1.5
```

### **Step 3: Update UNRAID Network Configuration**
```bash
# SSH to UNRAID server (if accessible)
# Method 1: If UNRAID GUI accessible
# - Navigate to Settings > Network Settings
# - Change IP from 11.1.2 to 10.11.1.10
# - Gateway: 10.1.1.1
# - DNS: 10.1.1.1

# Method 2: Via console/monitor
# - Access UNRAID directly via monitor/keyboard
# - Use network configuration menu
```

### **Step 4: Restart Network Services**
```bash
# On OPNsense router:
# Navigate to: System > Reboot
# OR restart specific services:
# - Services > Unbound DNS > Restart
# - Services > DHCPv4 > Restart
```

## 🔧 **Automated Resolution Scripts**

### **IP Migration Mapping**
```
OLD SUBNET: 11.1.1.0/24  → NEW SUBNET: 10.11.1.0/24
OLD SUBNET: 11.1.11.0/24 → NEW SUBNET: 10.11.1.0/24

Device Mappings:
11.1.1.1   → 10.1.1.1   (Router - already migrated)
11.1.1.3   → 10.11.1.3  (aftrs-main)
11.1.1.4   → 10.11.1.4  (baeblade)
11.1.1.5   → 10.11.1.5  (spinblade)
11.1.2     → 10.11.1.10 (UNRAID Server)
11.1.11.2  → 10.11.1.2  (Grafana)
11.1.11.10 → 10.11.1.10 (UNRAID Server alt)
11.1.11.20 → 10.11.1.20 (Dev PC)
```

## ⚠️ **CRITICAL NEXT STEPS**

1. **IMMEDIATELY** update OPNsense DHCP reservations for UNRAID server
2. **UPDATE** all DNS overrides in OPNsense Unbound DNS
3. **RECONFIGURE** UNRAID server network settings
4. **UPDATE** all hardcoded IPs in configuration files
5. **TEST** network connectivity after changes
6. **UPDATE** documentation to reflect new subnet

## 🔍 **Files Requiring Updates**

1. `aftrs_cli/scripts/bootstrap_opnsense_agent.sh` - Update DNS overrides
2. `agentctl/infrastructure/network.md` - Update network documentation  
3. `unraid-observability-stack/DASHBOARD_GUIDE.md` - Update Grafana URL
4. All configuration files with hardcoded `11.1.x` IPs

---
**⚡ URGENT: This is blocking network connectivity and must be resolved immediately!**
