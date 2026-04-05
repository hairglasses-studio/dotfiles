# UNRAID Server Thermal Troubleshooting Notes

**Date:** 2026-01-01
**System:** MSI MEG X670E ACE + AMD Ryzen (UNRAID)
**Cooler:** Corsair H150i Elite LCD XT

---

## Problem Summary

UNRAID server repeatedly shuts down during boot due to thermal issues. System passes POST and begins UNRAID boot, then powers off.

---

## Hardware Configuration

| Component | Details |
|-----------|---------|
| Motherboard | MSI MEG X670E ACE |
| CPU Cooler | Corsair H150i Elite LCD XT (360mm AIO) |
| Cooler Controller | Corsair Commander Core |
| Cache Pool | btrfs RAID5, 5x NVMe (2x Samsung 990 PRO + 3x Sabrent Rocket 5) |

---

## Troubleshooting Timeline

### BIOS Issues
- **Original BIOS:** 7D69v1K (Jan 2, 2025) - likely beta
- **Symptom:** Hardware monitor showed all fans at ~16,000 RPM (impossible/bug)
- **Fix:** Flashed to stable 7D69v1J (Oct 21, 2024) via Flash BIOS Button
- **Result:** BIOS bug resolved, but thermal shutdowns continued

### Cooler Connections Tested

| Connection | Header | Result |
|------------|--------|--------|
| Pump tach cable | CPU_FAN1 | Shutdown - fan curve throttles pump |
| Pump tach cable | W_PUMP+ | Shutdown - still occurring |
| Commander Core power | 6-pin PCIe | Connected ✓ |
| Pump head USB | USB 2.0 internal | Unknown - LCD stays off |

### POST Codes Observed
- **15** - Memory training (normal after CMOS reset, takes 5-10 min)
- **A6** - SCSI/storage device detection
- **D4** - PCI bus initialization

---

## Corsair H150i Elite LCD XT Connections

The cooler requires THREE connections:

1. **Pump Power/Tach** → W_PUMP+ or AIO_PUMP header (needs 100% duty cycle)
2. **Commander Core Power** → 6-pin PCIe from PSU (powers fans + controller)
3. **Pump Head USB** → Internal USB 2.0 header (required for LCD display)

### Important Notes
- LCD will NOT display without USB connection (even if pump runs)
- Commander Core controls radiator fans - if not powered, fans don't spin
- On Linux/UNRAID, use `liquidctl` for AIO control (no iCUE)

---

## Remaining Suspects

1. **Radiator fans not spinning** - Commander Core may not be powering fans despite 6-pin connected
2. **USB connection missing** - Pump head USB may need connection for full functionality
3. **Commander Core failure** - Hardware issue with fan controller
4. **Thermal paste** - May need reapplication if pump is running but no heat transfer

---

## Next Steps to Try

1. **Verify radiator fans spin** - Physically check all 3 fans on radiator spin at boot
2. **Check Commander Core LED** - Should have indicator light if powered
3. **Bypass Commander Core** - Connect radiator fans directly to motherboard headers (CHA_FAN1/2/3)
4. **Check USB connection** - Ensure pump head USB cable connected to motherboard
5. **Monitor CPU temp in BIOS** - If accessible, check temps in Hardware Monitor before shutdown
6. **Reseat cooler** - Remove and reinstall with fresh thermal paste

---

## Files Prepared

| File | Location | Purpose |
|------|----------|---------|
| MSI.ROM (v1.J) | USB drive "MSIBIOS" | Flash BIOS Button recovery |
| 7D69v1J.zip | Backup | October 2024 stable BIOS |
| 7D69v1K.zip | Backup | January 2025 beta BIOS (buggy) |

---

## Original Goal (Blocked)

Shrink btrfs RAID5 cache pool from 5 to 4 NVMe drives. Cannot proceed until server boots stably.

See: `docs/unraid-cache-shrink-plan.md`

---

## Commands for When Server Boots

```bash
# SSH into UNRAID
ssh root@192.168.x.x

# Check cache pool status
btrfs filesystem show /mnt/cache

# Check device usage
btrfs device usage /mnt/cache

# Remove a device (after identifying which one)
btrfs device remove /dev/nvmeXn1 /mnt/cache
```
