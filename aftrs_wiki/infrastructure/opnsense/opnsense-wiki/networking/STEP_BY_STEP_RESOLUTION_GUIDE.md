# 🚨 CRITICAL: Step-by-Step Network Crisis Resolution Guide

**EMERGENCY NETWORK REPAIR PROCEDURE**
**Status:** ACTIVE CRISIS - Multiple devices unreachable
**Priority:** URGENT - Complete immediately

## 🎯 **IMMEDIATE CRITICAL STEPS (DO FIRST)**

### **Step 1: Fix UNRAID Server (MOST CRITICAL)**
Your UNRAID server at `11.1.2` is completely unreachable on the new subnet.

#### **Option A: UNRAID GUI Method (if accessible)**
1. Try accessing UNRAID at: `http://11.1.2` (old IP)
2. If accessible: Go to **Settings > Network Settings**
3. Change settings:
   - **IP Address:** `10.11.1.10` 
   - **Netmask:** `255.255.255.0`
   - **Gateway:** `10.1.1.1`
   - **DNS Server:** `10.1.1.1`
4. Click **Apply** and **reboot** UNRAID

#### **Option B: Physical Console Method (if GUI inaccessible)**
1. Connect monitor/keyboard to UNRAID server
2. At console menu, select **Network Settings**
3. Configure:
   - **IP:** `10.11.1.10`
   - **Netmask:** `255.255.255.0` 
   - **Gateway:** `10.1.1.1`
   - **DNS:** `10.1.1.1`
4. Apply and reboot

### **Step 2: Update OPNsense DHCP Reservations**
1. Access OPNsense at: `http://10.1.1.1`
2. Navigate: **Services > DHCPv4 > LAN**
3. Find and update static reservations:
   - **UNRAID Server:** Change to `10.11.1.10`
   - **Any device with `11.1.x`:** Update to `10.11.1.x`
4. **Save** and **Apply Changes**

### **Step 3: Update OPNsense DNS Overrides**
1. Navigate: **Services > Unbound DNS > Overrides**
2. Update these domain overrides:
   - `aftrs-main.aftrs.void`: `11.1.1.3` → `10.11.1.3`
   - `baeblade.aftrs.void`: `11.1.1.4` → `10.11.1.4`
   - `spinblade.aftrs.void`: `11.1.1.5` → `10.11.1.5`
3. **Save** and **Apply Changes**

### **Step 4: Restart OPNsense Services**
1. Navigate: **Services > Unbound DNS** → Click **Restart**
2. Navigate: **Services > DHCPv4** → Click **Restart**
3. Or full reboot: **System > Reboot**

## 🔧 **ADDITIONAL CONFIGURATION FILES TO UPDATE**

The automated script found **20+ additional files** with hardcoded 11.1.x IPs:

### **Critical Files (Update ASAP)**

#### **1. AFTRS CLI Installation Script**
**File:** `aftrs_cli/install.sh`
**Issues Found:**
- `OPNSENSE_IP=11.1.11.1` → **Change to:** `10.1.1.1`
- Asset definitions with `11.1.11.x` IPs

#### **2. Network Asset Configuration** 
**File:** `aftrs_cli/network_assets/assets.yaml`
**Issues Found:**
- Multiple devices with `11.1.11.x` IPs
- **Update all to:** `10.11.1.x` range

#### **3. DHCP Device Mappings**
**File:** `aftrs_cli/dhcp/manifest.json`
**Issues Found:**
- `"aftrs-main": "11.1.11.3"` → **Change to:** `10.11.1.3`
- `"baeblade": "11.1.11.4"` → **Change to:** `10.11.1.4`

#### **4. Network Testing Scripts**
**Files:** Multiple scripts in `aftrs_cli/scripts/`
**Issues:** Hardcoded ping targets and connectivity tests

### **Automated Fix for Remaining Files**
```bash
cd /Users/mitch/Docs/aftrs-void

# Fix AFTRS CLI install script
sed -i '' 's/11\.1\.11\.1/10.1.1.1/g' aftrs_cli/install.sh
sed -i '' 's/11\.1\.11\.10/10.11.1.10/g' aftrs_cli/install.sh

# Fix network assets
sed -i '' 's/11\.1\.11\./10.11.1./g' aftrs_cli/network_assets/assets.yaml

# Fix DHCP manifest
sed -i '' 's/11\.1\.11\./10.11.1./g' aftrs_cli/dhcp/manifest.json

# Fix test scripts
find aftrs_cli/scripts/ -name "*.sh" -exec sed -i '' 's/11\.1\.11\./10.11.1./g' {} \;
```

## 🧪 **VERIFICATION TESTS**

After completing all updates:

### **Test 1: UNRAID Connectivity**
```bash
# Test new UNRAID IP
ping 10.11.1.10

# Test UNRAID web interface
curl -I http://10.11.1.10
```

### **Test 2: Grafana Dashboard**
```bash
# Test Grafana on new IP
curl -I http://10.11.1.2:3000
```

### **Test 3: DNS Resolution**
```bash
# Test DNS overrides
nslookup aftrs-main.aftrs.void 10.1.1.1
nslookup baeblade.aftrs.void 10.1.1.1
nslookup spinblade.aftrs.void 10.1.1.1
```

### **Test 4: Network Connectivity**
```bash
# Test subnet connectivity
ping 10.11.1.1  # Router
ping 10.11.1.10 # UNRAID
ping 10.11.1.2  # Grafana/monitoring
```

## ⚡ **PRIORITY ORDER**

**Do these in exact order:**

1. **🚨 URGENT:** Fix UNRAID server IP (Step 1)
2. **🚨 URGENT:** Update OPNsense DHCP reservations (Step 2)  
3. **⚠️ CRITICAL:** Update DNS overrides (Step 3)
4. **⚠️ CRITICAL:** Restart OPNsense services (Step 4)
5. **📝 Important:** Update remaining configuration files
6. **🧪 Verify:** Run all connectivity tests

## 🆘 **EMERGENCY CONTACT POINTS**

**If UNRAID is completely inaccessible:**
- **Physical access required** - connect monitor/keyboard
- **Network console access** - use serial/IPMI if available
- **Recovery mode** - boot UNRAID in safe mode

**If OPNsense is inaccessible:**
- Access via: `http://10.1.1.1` (new router IP)
- **Fallback:** Physical console access to router
- **Recovery:** Factory reset and restore config (LAST RESORT)

---
**⚡ THIS IS A NETWORK EMERGENCY - RESOLVE IMMEDIATELY TO RESTORE CONNECTIVITY**
