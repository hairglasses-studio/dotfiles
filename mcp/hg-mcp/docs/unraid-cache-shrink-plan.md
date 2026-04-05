# UNRAID Cache Pool Shrink Operation

## Objective
Remove 5th NVMe drive from btrfs RAID5 cache pool to run on 4 native M.2 slots.

## Prerequisites
- [ ] UNRAID booted with 4 NVMe (native) + 1 NVMe (USB-C enclosure)
- [ ] All 5 cache pool drives detected
- [ ] SSH access to UNRAID server
- [ ] No critical operations running on cache

---

## Phase 1: Discovery & Connection

### 1.1 Find UNRAID on Network
```bash
# Scan for UNRAID (typically responds on port 80/443/22)
nmap -sn 192.168.x.0/24 | grep -B2 -i "unraid\|your-server"

# Or check DHCP leases if you have router access
# Or use mDNS:
dns-sd -B _ssh._tcp local

# Direct ping test (if hostname known)
ping your-server.local
```

### 1.2 SSH Connection
```bash
ssh root@your-server.local
# or
ssh root@<IP_ADDRESS>
```

---

## Phase 2: Pre-Flight Checks

### 2.1 Verify All 5 Cache Drives Present
```bash
# Show btrfs filesystem with all devices
btrfs filesystem show /mnt/cache

# Expected output should show 5 devices:
# - Samsung_SSD_990_PRO_4TB_S7KGNU0Y105137N
# - Samsung_SSD_990_PRO_4TB_S7KGNU0Y105156D
# - Sabrent_SB-RKT5-4TB_48836385700287
# - Sabrent_SB-RKT5-4TB_48836385700293
# - Sabrent_SB-RKT5-4TB_48836385700285
```

### 2.2 Check Device Health
```bash
# Check for any existing errors
btrfs device stats /mnt/cache

# All counters should be 0
```

### 2.3 Identify USB-Connected Drive
```bash
# List all block devices with transport type
lsblk -o NAME,SIZE,MODEL,SERIAL,TRAN,MOUNTPOINT

# USB-connected drive will show TRAN=usb
# Native NVMe drives will show TRAN=nvme
```

### 2.4 Check Cache Usage
```bash
# See how much data is on cache
btrfs filesystem df /mnt/cache
df -h /mnt/cache

# Check what's using cache
du -sh /mnt/cache/*
```

### 2.5 Document Current State
```bash
# Save current state for reference
echo "=== PRE-MIGRATION STATE ===" > /boot/logs/cache-shrink-$(date +%Y%m%d).log
btrfs filesystem show /mnt/cache >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
btrfs filesystem df /mnt/cache >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
lsblk -o NAME,SIZE,MODEL,SERIAL,TRAN >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
```

---

## Phase 3: Prepare for Migration

### 3.1 Stop Services Using Cache
```bash
# Check what's using the cache mount
lsof +D /mnt/cache 2>/dev/null | head -20

# Stop Docker (via UNRAID UI preferred, or):
/etc/rc.d/rc.docker stop

# Stop VMs (via UNRAID UI preferred, or):
/etc/rc.d/rc.libvirt stop

# Verify nothing using cache
lsof +D /mnt/cache 2>/dev/null
```

### 3.2 Identify Target Drive for Removal
```bash
# Find the USB-connected drive's device path
USB_DRIVE=$(lsblk -o NAME,TRAN -n | grep usb | awk '{print "/dev/"$1}')
echo "USB drive to remove: $USB_DRIVE"

# Verify this is one of the cache pool members
btrfs filesystem show /mnt/cache | grep -i $(basename $USB_DRIVE)
```

---

## Phase 4: Execute Migration

### 4.1 Remove Drive from Pool
```bash
# CRITICAL: This is the main operation
# Replace /dev/sdX with the actual USB device path identified above

# Option A: Remove by device path
btrfs device remove /dev/sdX /mnt/cache

# Option B: Remove by device ID (safer)
# First get device ID:
btrfs filesystem show /mnt/cache
# Then remove by ID:
btrfs device remove <devid> /mnt/cache
```

### 4.2 Monitor Progress
```bash
# Watch rebalance progress (run in separate SSH session)
watch -n 10 'btrfs balance status /mnt/cache; echo "---"; btrfs filesystem df /mnt/cache'

# Alternative: check device removal status
btrfs device stats /mnt/cache
```

### 4.3 Estimated Duration
- USB 3.2 Gen 2 speed: ~1 GB/s theoretical, ~700 MB/s real
- If cache has 2TB data: ~45-60 minutes minimum
- If cache has 8TB data: ~3-4 hours minimum
- btrfs rebalance adds overhead: expect 2-3x longer

---

## Phase 5: Post-Migration Verification

### 5.1 Verify Removal Complete
```bash
# Should now show only 4 devices
btrfs filesystem show /mnt/cache

# Check filesystem health
btrfs filesystem df /mnt/cache

# Verify no errors introduced
btrfs device stats /mnt/cache

# Run scrub to verify data integrity
btrfs scrub start /mnt/cache
btrfs scrub status /mnt/cache
```

### 5.2 Restart Services
```bash
# Start Docker
/etc/rc.d/rc.docker start

# Start VMs
/etc/rc.d/rc.libvirt start

# Verify services running
docker ps
virsh list --all
```

### 5.3 Document Final State
```bash
echo "=== POST-MIGRATION STATE ===" >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
btrfs filesystem show /mnt/cache >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
btrfs filesystem df /mnt/cache >> /boot/logs/cache-shrink-$(date +%Y%m%d).log
```

---

## Phase 6: Physical Cleanup

### 6.1 Safe to Disconnect USB
```bash
# Verify drive is no longer part of pool
btrfs filesystem show /mnt/cache | grep -c devid
# Should show: 4

# Now safe to physically disconnect USB enclosure
```

### 6.2 Reboot Test
```bash
# Reboot without USB drive to confirm clean boot
reboot
```

---

## Optional: Convert RAID Profile

After shrinking to 4 drives, consider converting from RAID5 to RAID1:

```bash
# Check current profile
btrfs filesystem df /mnt/cache

# Convert to RAID1 (more reliable, uses 50% capacity)
btrfs balance start -dconvert=raid1 -mconvert=raid1 /mnt/cache

# Monitor conversion
watch -n 10 'btrfs balance status /mnt/cache'
```

### RAID Profile Comparison (4x 4TB drives = 16TB raw)

| Profile | Usable Space | Fault Tolerance | Recommended |
|---------|--------------|-----------------|-------------|
| RAID5   | 12TB         | 1 drive         | ⚠️ Risky with 4 drives |
| RAID1   | 8TB          | 1 drive         | ✅ More reliable |
| RAID10  | 8TB          | 1 per mirror    | ✅ Best performance |

---

## Rollback Plan

If something goes wrong during device removal:

```bash
# Cancel ongoing balance/removal
btrfs balance cancel /mnt/cache

# Check filesystem status
btrfs filesystem show /mnt/cache

# If filesystem is degraded, mount with recovery
mount -o degraded,ro /dev/nvme0n1p1 /mnt/cache

# Run recovery
btrfs rescue zero-log /dev/nvme0n1p1
```

---

## Quick Reference Commands

```bash
# Status check one-liner
btrfs filesystem show /mnt/cache && btrfs device stats /mnt/cache && df -h /mnt/cache

# Find USB drive
lsblk -o NAME,SIZE,SERIAL,TRAN | grep usb

# Monitor rebalance
watch -n 5 'btrfs balance status /mnt/cache 2>&1'

# Check for errors
dmesg | grep -i btrfs | tail -20
```

---

## Timeline

| Step | Duration | Notes |
|------|----------|-------|
| Discovery & SSH | 5 min | |
| Pre-flight checks | 10 min | |
| Stop services | 5 min | |
| Device removal | 1-6 hours | Depends on data size |
| Verification | 15 min | |
| Scrub (optional) | 1-2 hours | Can run in background |
| **Total** | **2-8 hours** | |

---

## Notes

- **Do not disconnect USB during removal** - This will corrupt the filesystem
- **Keep SSH session alive** - Use `screen` or `tmux` for long operations
- **Have backups** - Critical data should be backed up before this operation
- **UPS recommended** - Power loss during rebalance can cause issues
