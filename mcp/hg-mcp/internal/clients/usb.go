// Package clients provides USB drive management across macOS and Linux/WSL2.
package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// USBDrive represents an external USB drive
type USBDrive struct {
	Identifier string `json:"identifier"` // e.g., disk4 (macOS) or sda (Linux)
	Name       string `json:"name"`
	Size       string `json:"size"`
	Type       string `json:"type"`     // external, internal
	Protocol   string `json:"protocol"` // USB, SATA, NVMe
	Mountpoint string `json:"mountpoint"`
	IsVentoy   bool   `json:"is_ventoy"`
}

// ISOFile represents an ISO file on a Ventoy drive
type ISOFile struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
}

// KnownISO represents a downloadable ISO with known URL
type KnownISO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"` // rescue, os, utility
	SizeHuman   string `json:"size_human"`
}

// GetKnownISOs returns a curated list of useful bootable ISOs
func GetKnownISOs() []KnownISO {
	return []KnownISO{
		// Rescue & Recovery
		{Name: "GParted Live", Description: "Partition editor for disk management", URL: "https://downloads.sourceforge.net/gparted/gparted-live-1.6.0-3-amd64.iso", Category: "rescue", SizeHuman: "~500MB"},
		{Name: "SystemRescue", Description: "Full Linux rescue environment with disk tools", URL: "https://downloads.sourceforge.net/systemrescuecd/systemrescue-11.03-amd64.iso", Category: "rescue", SizeHuman: "~950MB"},
		{Name: "Clonezilla", Description: "Disk imaging and cloning utility", URL: "https://downloads.sourceforge.net/clonezilla/clonezilla-live-3.2.0-5-amd64.iso", Category: "rescue", SizeHuman: "~450MB"},
		{Name: "Hiren's BootCD PE", Description: "Windows PE with recovery tools", URL: "https://www.hirensbootcd.org/files/HBCD_PE_x64.iso", Category: "rescue", SizeHuman: "~2.9GB"},

		// Operating Systems
		{Name: "Ubuntu Desktop 24.04", Description: "Popular Linux desktop distribution", URL: "https://releases.ubuntu.com/24.04.1/ubuntu-24.04.1-desktop-amd64.iso", Category: "os", SizeHuman: "~5.9GB"},
		{Name: "Fedora Workstation 41", Description: "Cutting-edge Linux desktop", URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/41/Workstation/x86_64/iso/Fedora-Workstation-Live-x86_64-41-1.4.iso", Category: "os", SizeHuman: "~2.3GB"},
		{Name: "Linux Mint 22", Description: "User-friendly Linux desktop", URL: "https://mirrors.kernel.org/linuxmint/stable/22/linuxmint-22-cinnamon-64bit.iso", Category: "os", SizeHuman: "~2.8GB"},

		// Utility
		{Name: "Memtest86+", Description: "Memory testing utility", URL: "https://www.memtest.org/download/v7.00/mt86plus_7.00_64.iso.zip", Category: "utility", SizeHuman: "~25MB"},
		{Name: "ShredOS", Description: "Secure disk wiping (DBAN successor)", URL: "https://github.com/PartialVolume/shredos.x86_64/releases/download/v2024.02.2/shredos-2024.02.2_a_x86-64_0.38.img", Category: "utility", SizeHuman: "~70MB"},
		{Name: "UEFI Shell", Description: "UEFI environment for diagnostics", URL: "https://github.com/pbatard/UEFI-Shell/releases/download/22H2/UEFI-Shell-2.2-22H2-RELEASE.iso", Category: "utility", SizeHuman: "~3MB"},
	}
}

// defaultVentoyPath returns the platform-appropriate default Ventoy mount path.
func defaultVentoyPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/Volumes/Ventoy"
	default:
		// Linux: check common mount locations
		for _, p := range []string{"/media/Ventoy", "/mnt/Ventoy", "/run/media/" + os.Getenv("USER") + "/Ventoy"} {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "/media/Ventoy"
	}
}

// ListUSBDrives returns all external USB drives, dispatching by platform.
func ListUSBDrives(ctx context.Context) ([]USBDrive, error) {
	switch runtime.GOOS {
	case "darwin":
		return listUSBDrivesDarwin(ctx)
	case "linux":
		return listUSBDrivesLinux(ctx)
	default:
		return nil, fmt.Errorf("USB drive listing not supported on %s", runtime.GOOS)
	}
}

// ── Linux implementation ────────────────────────────────────────────────────

// lsblkOutput mirrors the JSON output of `lsblk -J`.
type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name       string        `json:"name"`
	Size       string        `json:"size"`
	Type       string        `json:"type"` // disk, part
	Tran       string        `json:"tran"` // usb, sata, nvme, ""
	RM         bool          `json:"rm"`   // removable
	Model      string        `json:"model"`
	Mountpoint string        `json:"mountpoint"`
	Label      string        `json:"label"`
	Children   []lsblkDevice `json:"children"`
}

func listUSBDrivesLinux(ctx context.Context) ([]USBDrive, error) {
	cmd := exec.CommandContext(ctx, "lsblk", "-J", "-o", "NAME,SIZE,TYPE,TRAN,RM,MODEL,MOUNTPOINT,LABEL")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed: %w", err)
	}

	var parsed lsblkOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	var drives []USBDrive
	for _, dev := range parsed.Blockdevices {
		if dev.Type != "disk" {
			continue
		}
		// Filter to removable or USB-transport devices
		if !dev.RM && dev.Tran != "usb" {
			continue
		}

		drive := USBDrive{
			Identifier: dev.Name,
			Name:       strings.TrimSpace(dev.Model),
			Size:       dev.Size,
			Protocol:   strings.ToUpper(dev.Tran),
			Type:       "external",
		}
		if drive.Name == "" {
			drive.Name = dev.Name
		}

		// Check partitions for mount points and Ventoy
		for _, child := range dev.Children {
			if child.Mountpoint != "" && drive.Mountpoint == "" {
				drive.Mountpoint = child.Mountpoint
			}
			if strings.EqualFold(child.Label, "Ventoy") || strings.EqualFold(child.Label, "ventoy") {
				drive.IsVentoy = true
				if child.Mountpoint != "" {
					drive.Mountpoint = child.Mountpoint
				}
			}
		}
		// Also check the device-level mountpoint
		if dev.Mountpoint != "" && drive.Mountpoint == "" {
			drive.Mountpoint = dev.Mountpoint
		}

		drives = append(drives, drive)
	}

	return drives, nil
}

// ── macOS implementation ────────────────────────────────────────────────────

func listUSBDrivesDarwin(ctx context.Context) ([]USBDrive, error) {
	cmd := exec.CommandContext(ctx, "diskutil", "list", "external")
	listOut, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("diskutil list failed: %w", err)
	}

	var drives []USBDrive
	lines := strings.Split(string(listOut), "\n")
	diskRe := regexp.MustCompile(`^/dev/(disk\d+)`)

	for _, line := range lines {
		matches := diskRe.FindStringSubmatch(line)
		if len(matches) > 1 {
			diskID := matches[1]
			drive, err := getUSBDriveInfo(ctx, diskID)
			if err != nil {
				continue
			}
			drives = append(drives, drive)
		}
	}

	// Also add mounted Ventoy volumes
	if _, err := os.Stat("/Volumes/Ventoy"); err == nil {
		for i := range drives {
			if drives[i].Mountpoint == "/Volumes/Ventoy" || isVentoyDrive(drives[i].Identifier) {
				drives[i].IsVentoy = true
			}
		}
	}

	return drives, nil
}

func getUSBDriveInfo(ctx context.Context, diskID string) (USBDrive, error) {
	cmd := exec.CommandContext(ctx, "diskutil", "info", diskID)
	out, err := cmd.Output()
	if err != nil {
		return USBDrive{}, err
	}

	drive := USBDrive{Identifier: diskID}
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Device / Media Name":
			drive.Name = val
		case "Disk Size":
			if idx := strings.Index(val, "("); idx > 0 {
				drive.Size = strings.TrimSpace(val[:idx])
			} else {
				drive.Size = val
			}
		case "Protocol":
			drive.Protocol = val
		case "Device Location":
			if strings.Contains(strings.ToLower(val), "external") {
				drive.Type = "external"
			} else {
				drive.Type = "internal"
			}
		case "Mount Point":
			drive.Mountpoint = val
		}
	}

	return drive, nil
}

func isVentoyDrive(diskID string) bool {
	efiPath := fmt.Sprintf("/Volumes/EFI/ventoy")
	if _, err := os.Stat(efiPath); err == nil {
		return true
	}
	return false
}

// ListVentoyISOs lists all ISO files on a Ventoy drive
func ListVentoyISOs(ctx context.Context, mountpoint string) ([]ISOFile, error) {
	if mountpoint == "" {
		mountpoint = defaultVentoyPath()
	}

	if _, err := os.Stat(mountpoint); os.IsNotExist(err) {
		return nil, fmt.Errorf("Ventoy drive not mounted at %s", mountpoint)
	}

	var isos []ISOFile
	err := filepath.Walk(mountpoint, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".iso") {
			isos = append(isos, ISOFile{
				Name:      info.Name(),
				Path:      path,
				Size:      info.Size(),
				SizeHuman: humanSize(info.Size()),
			})
		}
		return nil
	})

	return isos, err
}

// DownloadISO downloads an ISO to the Ventoy drive
func DownloadISO(ctx context.Context, url, destDir string) (string, error) {
	if destDir == "" {
		destDir = defaultVentoyPath()
	}

	filename := filepath.Base(url)
	destPath := filepath.Join(destDir, filename)

	if _, err := os.Stat(destPath); err == nil {
		return destPath, fmt.Errorf("file already exists: %s", destPath)
	}

	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", destPath, url)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("download failed: %w - %s", err, stderr.String())
	}

	return destPath, nil
}

// InstallVentoy installs Ventoy to a USB drive
func InstallVentoy(ctx context.Context, diskID, ventoyPath string) error {
	return fmt.Errorf("automated Ventoy installation not yet implemented on %s — use the Ventoy GUI or CLI manually", runtime.GOOS)
}

// UnmountDisk unmounts all volumes on a disk
func UnmountDisk(ctx context.Context, diskID string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "diskutil", "unmountDisk", "/dev/"+diskID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount %s: %w", diskID, err)
		}
	case "linux":
		cmd := exec.CommandContext(ctx, "udisksctl", "unmount", "-b", "/dev/"+diskID)
		if err := cmd.Run(); err != nil {
			// Fallback to umount
			cmd2 := exec.CommandContext(ctx, "umount", "/dev/"+diskID)
			if err2 := cmd2.Run(); err2 != nil {
				return fmt.Errorf("failed to unmount %s: %w", diskID, err)
			}
		}
	default:
		return fmt.Errorf("unmount not supported on %s", runtime.GOOS)
	}
	return nil
}

// MountDisk mounts a disk
func MountDisk(ctx context.Context, diskID string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.CommandContext(ctx, "diskutil", "mountDisk", "/dev/"+diskID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to mount %s: %w", diskID, err)
		}
	case "linux":
		cmd := exec.CommandContext(ctx, "udisksctl", "mount", "-b", "/dev/"+diskID)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to mount %s: %w", diskID, err)
		}
	default:
		return fmt.Errorf("mount not supported on %s", runtime.GOOS)
	}
	return nil
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetDiskInfo returns detailed info about a disk
func GetDiskInfo(ctx context.Context, diskID string) (map[string]interface{}, error) {
	switch runtime.GOOS {
	case "darwin":
		return getDiskInfoDarwin(ctx, diskID)
	case "linux":
		return getDiskInfoLinux(ctx, diskID)
	default:
		return nil, fmt.Errorf("disk info not supported on %s", runtime.GOOS)
	}
}

func getDiskInfoDarwin(ctx context.Context, diskID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "diskutil", "info", "-plist", diskID)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out, &result); err != nil {
		// Fallback to text parsing
		drive, err := getUSBDriveInfo(ctx, diskID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"identifier": drive.Identifier,
			"name":       drive.Name,
			"size":       drive.Size,
			"protocol":   drive.Protocol,
			"type":       drive.Type,
			"mountpoint": drive.Mountpoint,
			"is_ventoy":  drive.IsVentoy,
		}, nil
	}
	return result, nil
}

func getDiskInfoLinux(ctx context.Context, diskID string) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "lsblk", "-J", "-o", "NAME,SIZE,TYPE,TRAN,RM,MODEL,MOUNTPOINT,LABEL,FSTYPE,UUID", "/dev/"+diskID)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lsblk failed for %s: %w", diskID, err)
	}

	var parsed lsblkOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
	}

	if len(parsed.Blockdevices) == 0 {
		return nil, fmt.Errorf("device %s not found", diskID)
	}

	dev := parsed.Blockdevices[0]
	result := map[string]interface{}{
		"identifier": dev.Name,
		"name":       strings.TrimSpace(dev.Model),
		"size":       dev.Size,
		"protocol":   strings.ToUpper(dev.Tran),
		"type":       "disk",
		"removable":  dev.RM,
		"mountpoint": dev.Mountpoint,
	}

	// Include partition info
	if len(dev.Children) > 0 {
		parts := make([]map[string]interface{}, 0, len(dev.Children))
		for _, child := range dev.Children {
			parts = append(parts, map[string]interface{}{
				"name":       child.Name,
				"size":       child.Size,
				"label":      child.Label,
				"mountpoint": child.Mountpoint,
			})
		}
		result["partitions"] = parts
	}

	return result, nil
}
