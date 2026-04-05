# usb

> USB drive management including Ventoy bootable drives and ISO management

**7 tools**

## Tools

- [`aftrs_iso_catalog`](#aftrs-iso-catalog)
- [`aftrs_iso_download`](#aftrs-iso-download)
- [`aftrs_usb_info`](#aftrs-usb-info)
- [`aftrs_usb_list`](#aftrs-usb-list)
- [`aftrs_usb_mount`](#aftrs-usb-mount)
- [`aftrs_usb_unmount`](#aftrs-usb-unmount)
- [`aftrs_ventoy_isos`](#aftrs-ventoy-isos)

---

## aftrs_iso_catalog

Get catalog of recommended bootable ISOs with download URLs.

**Complexity:** simple

**Tags:** `iso`, `catalog`, `download`, `bootable`

**Use Cases:**
- Find ISOs to download
- Browse utility ISOs

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `category` | string |  | Filter by category: rescue, os, utility (default: all) |

### Example

```json
{
  "category": "example"
}
```

---

## aftrs_iso_download

Download an ISO file to a Ventoy USB drive.

**Complexity:** moderate

**Tags:** `iso`, `download`, `ventoy`, `bootable`

**Use Cases:**
- Add ISO to Ventoy
- Download bootable ISO

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `dest_dir` | string |  | Destination directory (default: /Volumes/Ventoy) |
| `url` | string | Yes | URL of the ISO to download |

### Example

```json
{
  "dest_dir": "example",
  "url": "example"
}
```

---

## aftrs_usb_info

Get detailed information about a specific USB drive.

**Complexity:** simple

**Tags:** `usb`, `info`, `disk`, `details`

**Use Cases:**
- Get disk details
- Check disk properties

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `disk_id` | string | Yes | Disk identifier (e.g., disk4) |

### Example

```json
{
  "disk_id": "example"
}
```

---

## aftrs_usb_list

List external USB drives with their details including Ventoy status.

**Complexity:** simple

**Tags:** `usb`, `drives`, `storage`, `external`

**Use Cases:**
- Find USB drives
- Check drive status

---

## aftrs_usb_mount

Mount a USB disk.

**Complexity:** simple

**Tags:** `usb`, `mount`

**Use Cases:**
- Mount USB drive
- Access USB contents

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `disk_id` | string | Yes | Disk identifier to mount (e.g., disk4) |

### Example

```json
{
  "disk_id": "example"
}
```

---

## aftrs_usb_unmount

Unmount all volumes on a USB disk.

**Complexity:** simple

**Tags:** `usb`, `unmount`, `eject`

**Use Cases:**
- Safely remove USB
- Prepare for imaging

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `disk_id` | string | Yes | Disk identifier to unmount (e.g., disk4) |

### Example

```json
{
  "disk_id": "example"
}
```

---

## aftrs_ventoy_isos

List ISO files on a Ventoy USB drive.

**Complexity:** simple

**Tags:** `ventoy`, `iso`, `bootable`, `list`

**Use Cases:**
- List bootable ISOs
- Check Ventoy contents

### Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `mountpoint` | string |  | Mount point of Ventoy drive (default: /Volumes/Ventoy) |

### Example

```json
{
  "mountpoint": "example"
}
```

---

