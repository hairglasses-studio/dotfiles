// Package usb provides USB drive management tools for hg-mcp.
package usb

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for USB drive management
type Module struct{}

func (m *Module) Name() string {
	return "usb"
}

func (m *Module) Description() string {
	return "USB drive management including Ventoy bootable drives and ISO management"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_usb_list",
				mcp.WithDescription("List external USB drives with their details including Ventoy status."),
			),
			Handler:             handleUSBList,
			Category:            "usb",
			Subcategory:         "drives",
			Tags:                []string{"usb", "drives", "storage", "external"},
			UseCases:            []string{"Find USB drives", "Check drive status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
		},
		{
			Tool: mcp.NewTool("aftrs_usb_info",
				mcp.WithDescription("Get detailed information about a specific USB drive."),
				mcp.WithString("disk_id",
					mcp.Description("Disk identifier (e.g., disk4)"),
					mcp.Required(),
				),
			),
			Handler:             handleUSBInfo,
			Category:            "usb",
			Subcategory:         "drives",
			Tags:                []string{"usb", "info", "disk", "details"},
			UseCases:            []string{"Get disk details", "Check disk properties"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
		},
		{
			Tool: mcp.NewTool("aftrs_ventoy_isos",
				mcp.WithDescription("List ISO files on a Ventoy USB drive."),
				mcp.WithString("mountpoint",
					mcp.Description("Mount point of Ventoy drive (auto-detected per platform)"),
				),
			),
			Handler:             handleVentoyISOs,
			Category:            "usb",
			Subcategory:         "ventoy",
			Tags:                []string{"ventoy", "iso", "bootable", "list"},
			UseCases:            []string{"List bootable ISOs", "Check Ventoy contents"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
		},
		{
			Tool: mcp.NewTool("aftrs_iso_catalog",
				mcp.WithDescription("Get catalog of recommended bootable ISOs with download URLs."),
				mcp.WithString("category",
					mcp.Description("Filter by category: rescue, os, utility (default: all)"),
				),
			),
			Handler:             handleISOCatalog,
			Category:            "usb",
			Subcategory:         "ventoy",
			Tags:                []string{"iso", "catalog", "download", "bootable"},
			UseCases:            []string{"Find ISOs to download", "Browse utility ISOs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
		},
		{
			Tool: mcp.NewTool("aftrs_iso_download",
				mcp.WithDescription("Download an ISO file to a Ventoy USB drive."),
				mcp.WithString("url",
					mcp.Description("URL of the ISO to download"),
					mcp.Required(),
				),
				mcp.WithString("dest_dir",
					mcp.Description("Destination directory (auto-detected per platform)"),
				),
			),
			Handler:             handleISODownload,
			Category:            "usb",
			Subcategory:         "ventoy",
			Tags:                []string{"iso", "download", "ventoy", "bootable"},
			UseCases:            []string{"Add ISO to Ventoy", "Download bootable ISO"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "usb",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_usb_unmount",
				mcp.WithDescription("Unmount all volumes on a USB disk."),
				mcp.WithString("disk_id",
					mcp.Description("Disk identifier to unmount (e.g., disk4)"),
					mcp.Required(),
				),
			),
			Handler:             handleUSBUnmount,
			Category:            "usb",
			Subcategory:         "drives",
			Tags:                []string{"usb", "unmount", "eject"},
			UseCases:            []string{"Safely remove USB", "Prepare for imaging"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_usb_mount",
				mcp.WithDescription("Mount a USB disk."),
				mcp.WithString("disk_id",
					mcp.Description("Disk identifier to mount (e.g., disk4)"),
					mcp.Required(),
				),
			),
			Handler:             handleUSBMount,
			Category:            "usb",
			Subcategory:         "drives",
			Tags:                []string{"usb", "mount"},
			UseCases:            []string{"Mount USB drive", "Access USB contents"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "usb",
			IsWrite:             true,
		},
	}
}

func handleUSBList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drives, err := clients.ListUSBDrives(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# External USB Drives\n\n")

	if len(drives) == 0 {
		sb.WriteString("No external USB drives found.\n\n")
		sb.WriteString("**Tips:**\n")
		sb.WriteString("- Ensure USB drive is connected\n")
		if runtime.GOOS == "darwin" {
			sb.WriteString("- Check System Preferences > Security for drive access\n")
		} else {
			sb.WriteString("- On WSL2, USB passthrough requires usbipd-win: `usbipd attach --wsl`\n")
		}
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Device | Name | Size | Protocol | Ventoy |\n")
	sb.WriteString("|--------|------|------|----------|--------|\n")

	for _, drive := range drives {
		ventoy := ""
		if drive.IsVentoy {
			ventoy = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| /dev/%s | %s | %s | %s | %s |\n",
			drive.Identifier, drive.Name, drive.Size, drive.Protocol, ventoy))
	}

	sb.WriteString("\n**Note:** Use `aftrs_usb_info` for detailed information about a specific drive.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUSBInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskID, errResult := tools.RequireStringParam(req, "disk_id")
	if errResult != nil {
		return errResult, nil
	}

	// Remove /dev/ prefix if present
	diskID = strings.TrimPrefix(diskID, "/dev/")

	info, err := clients.GetDiskInfo(ctx, diskID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Disk Info: /dev/%s\n\n", diskID))

	for key, val := range info {
		sb.WriteString(fmt.Sprintf("- **%s:** %v\n", key, val))
	}

	return tools.TextResult(sb.String()), nil
}

func handleVentoyISOs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mountpoint := tools.GetStringParam(req, "mountpoint")
	// Empty mountpoint is handled by clients.ListVentoyISOs with platform-aware default

	isos, err := clients.ListVentoyISOs(ctx, mountpoint)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# ISOs on Ventoy Drive (%s)\n\n", mountpoint))

	if len(isos) == 0 {
		sb.WriteString("No ISO files found.\n\n")
		sb.WriteString("**To add ISOs:**\n")
		sb.WriteString("1. Use `aftrs_iso_catalog` to browse available ISOs\n")
		sb.WriteString("2. Use `aftrs_iso_download` to download ISOs to the drive\n")
		sb.WriteString("3. Or manually copy .iso files to the Ventoy drive\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| ISO Name | Size |\n")
	sb.WriteString("|----------|------|\n")

	var totalSize int64
	for _, iso := range isos {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", iso.Name, iso.SizeHuman))
		totalSize += iso.Size
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d ISOs, %s\n", len(isos), humanSize(totalSize)))
	sb.WriteString("\n**Usage:** Boot from USB and select an ISO from the Ventoy menu.\n")

	return tools.TextResult(sb.String()), nil
}

func handleISOCatalog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(req, "category")

	isos := clients.GetKnownISOs()

	var sb strings.Builder
	sb.WriteString("# Bootable ISO Catalog\n\n")

	// Group by category
	categories := map[string][]clients.KnownISO{
		"rescue":  {},
		"os":      {},
		"utility": {},
	}

	for _, iso := range isos {
		if category == "" || iso.Category == category {
			categories[iso.Category] = append(categories[iso.Category], iso)
		}
	}

	// Rescue & Recovery
	if len(categories["rescue"]) > 0 {
		sb.WriteString("## Rescue & Recovery\n\n")
		sb.WriteString("| Name | Description | Size |\n")
		sb.WriteString("|------|-------------|------|\n")
		for _, iso := range categories["rescue"] {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", iso.Name, iso.Description, iso.SizeHuman))
		}
		sb.WriteString("\n")
	}

	// Operating Systems
	if len(categories["os"]) > 0 {
		sb.WriteString("## Operating Systems\n\n")
		sb.WriteString("| Name | Description | Size |\n")
		sb.WriteString("|------|-------------|------|\n")
		for _, iso := range categories["os"] {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", iso.Name, iso.Description, iso.SizeHuman))
		}
		sb.WriteString("\n")
	}

	// Utilities
	if len(categories["utility"]) > 0 {
		sb.WriteString("## Utilities\n\n")
		sb.WriteString("| Name | Description | Size |\n")
		sb.WriteString("|------|-------------|------|\n")
		for _, iso := range categories["utility"] {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", iso.Name, iso.Description, iso.SizeHuman))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("**Note:** Windows 11 ISO must be downloaded from Microsoft: https://microsoft.com/software-download/windows11\n\n")
	sb.WriteString("**Download:** Use `aftrs_iso_download url=<iso_url>` to download an ISO to your Ventoy drive.\n")

	return tools.TextResult(sb.String()), nil
}

func handleISODownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	destDir := tools.GetStringParam(req, "dest_dir")
	// Empty destDir is handled by clients.DownloadISO with platform-aware default

	var sb strings.Builder
	sb.WriteString("# ISO Download\n\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n", url))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n\n", destDir))
	sb.WriteString("Starting download... (this may take several minutes)\n")

	path, err := clients.DownloadISO(ctx, url, destDir)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sb.WriteString(fmt.Sprintf("\n**Download complete:** %s\n", path))
	sb.WriteString("\nThe ISO is now available in your Ventoy boot menu.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUSBUnmount(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskID, errResult := tools.RequireStringParam(req, "disk_id")
	if errResult != nil {
		return errResult, nil
	}

	diskID = strings.TrimPrefix(diskID, "/dev/")

	if err := clients.UnmountDisk(ctx, diskID); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Disk Unmounted: /dev/%s\n\n", diskID))
	sb.WriteString("All volumes on this disk have been unmounted.\n")
	sb.WriteString("The disk can now be safely removed or used for imaging.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUSBMount(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskID, errResult := tools.RequireStringParam(req, "disk_id")
	if errResult != nil {
		return errResult, nil
	}

	diskID = strings.TrimPrefix(diskID, "/dev/")

	if err := clients.MountDisk(ctx, diskID); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Disk Mounted: /dev/%s\n\n", diskID))
	sb.WriteString("The disk has been mounted.\n")
	sb.WriteString("Use `aftrs_usb_list` to see mount points.\n")

	return tools.TextResult(sb.String()), nil
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

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
