// Package backup provides backup automation tools for hg-mcp.
package backup

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for backup automation
type Module struct{}

func (m *Module) Name() string {
	return "backup"
}

func (m *Module) Description() string {
	return "Project backup automation and restore capabilities"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_backup_projects",
				mcp.WithDescription("Create a backup of a project directory."),
				mcp.WithString("project", mcp.Required(), mcp.Description("Project name")),
				mcp.WithString("source", mcp.Required(), mcp.Description("Source directory path")),
				mcp.WithString("exclude", mcp.Description("Comma-separated patterns to exclude")),
			),
			Handler:             handleBackupProjects,
			Category:            "backup",
			Subcategory:         "create",
			Tags:                []string{"backup", "archive", "project", "tar"},
			UseCases:            []string{"Backup project files", "Create project archive"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "backup",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_backup_status",
				mcp.WithDescription("Get backup status for all projects."),
				mcp.WithString("project", mcp.Description("Filter by project name")),
			),
			Handler:             handleBackupStatus,
			Category:            "backup",
			Subcategory:         "status",
			Tags:                []string{"backup", "status", "list", "overview"},
			UseCases:            []string{"Check backup status", "View backup history"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "backup",
		},
		{
			Tool: mcp.NewTool("aftrs_backup_restore",
				mcp.WithDescription("Restore a backup to a target directory."),
				mcp.WithString("project", mcp.Required(), mcp.Description("Project name")),
				mcp.WithString("backup_id", mcp.Required(), mcp.Description("Backup ID to restore")),
				mcp.WithString("target", mcp.Required(), mcp.Description("Target directory for restore")),
			),
			Handler:             handleBackupRestore,
			Category:            "backup",
			Subcategory:         "restore",
			Tags:                []string{"backup", "restore", "recover", "extract"},
			UseCases:            []string{"Restore from backup", "Recover project files"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "backup",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_backup_list",
				mcp.WithDescription("List all backups for a project."),
				mcp.WithString("project", mcp.Required(), mcp.Description("Project name")),
			),
			Handler:             handleBackupList,
			Category:            "backup",
			Subcategory:         "list",
			Tags:                []string{"backup", "list", "history", "versions"},
			UseCases:            []string{"View backup versions", "List available backups"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "backup",
		},
	}
}

var getClient = tools.LazyClient(clients.NewBackupClient)

// handleBackupProjects handles the aftrs_backup_projects tool
func handleBackupProjects(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectName, errResult := tools.RequireStringParam(req, "project")
	if errResult != nil {
		return errResult, nil
	}

	sourcePath, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	var exclude []string
	if excludeStr := tools.GetStringParam(req, "exclude"); excludeStr != "" {
		exclude = strings.Split(excludeStr, ",")
		for i := range exclude {
			exclude[i] = strings.TrimSpace(exclude[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.BackupProject(ctx, projectName, sourcePath, exclude)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Backup Result\n\n")

	if info.Status == "completed" {
		sb.WriteString("**Status:** Completed\n\n")
		sb.WriteString(fmt.Sprintf("**Project:** %s\n", info.ProjectName))
		sb.WriteString(fmt.Sprintf("**Backup ID:** %s\n", info.ID))
		sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", info.SourcePath))
		sb.WriteString(fmt.Sprintf("**Backup:** `%s`\n", info.BackupPath))
		sb.WriteString(fmt.Sprintf("**Size:** %s\n", formatBytes(info.SizeBytes)))
		sb.WriteString(fmt.Sprintf("**Files:** %d\n", info.FileCount))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", info.Duration))
	} else {
		sb.WriteString("**Status:** Failed\n\n")
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", info.Error))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBackupStatus handles the aftrs_backup_status tool
func handleBackupStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectFilter := tools.GetStringParam(req, "project")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetBackupStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Backup Status\n\n")

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Total Backups:** %d\n", status.TotalBackups))
	sb.WriteString(fmt.Sprintf("**Total Size:** %s\n", formatBytes(status.TotalSizeBytes)))

	if status.LastBackup != nil {
		sb.WriteString(fmt.Sprintf("**Last Backup:** %s (%s)\n",
			status.LastBackup.ProjectName,
			status.LastBackup.Timestamp.Format("2006-01-02 15:04")))
	}

	sb.WriteString("\n## Projects\n\n")

	if len(status.Projects) == 0 {
		sb.WriteString("No backup projects found.\n")
		sb.WriteString("\nUse `aftrs_backup_projects` to create a backup.\n")
	} else {
		sb.WriteString("| Project | Backups | Total Size | Last Backup |\n")
		sb.WriteString("|---------|---------|------------|-------------|\n")

		for _, proj := range status.Projects {
			if projectFilter != "" && proj.Name != projectFilter {
				continue
			}

			lastBackup := "Never"
			if proj.LastBackup != nil {
				lastBackup = proj.LastBackup.Timestamp.Format("2006-01-02 15:04")
			}

			sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
				proj.Name, proj.BackupCount, formatBytes(proj.TotalSize), lastBackup))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleBackupRestore handles the aftrs_backup_restore tool
func handleBackupRestore(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectName, errResult := tools.RequireStringParam(req, "project")
	if errResult != nil {
		return errResult, nil
	}

	backupID, errResult := tools.RequireStringParam(req, "backup_id")
	if errResult != nil {
		return errResult, nil
	}

	targetPath, errResult := tools.RequireStringParam(req, "target")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.RestoreBackup(ctx, projectName, backupID, targetPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Restore Result\n\n")

	if result.Success {
		sb.WriteString("**Status:** Completed\n\n")
		sb.WriteString(fmt.Sprintf("**Backup ID:** %s\n", result.BackupID))
		sb.WriteString(fmt.Sprintf("**Restored To:** `%s`\n", result.RestorePath))
		sb.WriteString(fmt.Sprintf("**Files:** %d\n", result.FileCount))
		sb.WriteString(fmt.Sprintf("**Size:** %s\n", formatBytes(result.SizeBytes)))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", result.Duration))
	} else {
		sb.WriteString("**Status:** Failed\n\n")
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", result.Error))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBackupList handles the aftrs_backup_list tool
func handleBackupList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectName, errResult := tools.RequireStringParam(req, "project")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	backups, err := client.ListBackups(ctx, projectName)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Backups for %s\n\n", projectName))

	if len(backups) == 0 {
		sb.WriteString("No backups found for this project.\n")
		sb.WriteString("\nUse `aftrs_backup_projects` to create a backup.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found **%d** backups:\n\n", len(backups)))
		sb.WriteString("| ID | Date | Size | Status |\n")
		sb.WriteString("|----|------|------|--------|\n")

		for _, backup := range backups {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				backup.ID,
				backup.Timestamp.Format("2006-01-02 15:04"),
				formatBytes(backup.SizeBytes),
				backup.Status))
		}

		sb.WriteString("\n## Restore\n\n")
		sb.WriteString("To restore a backup, use:\n")
		sb.WriteString("```\n")
		sb.WriteString(fmt.Sprintf("aftrs_backup_restore project=%s backup_id=<ID> target=/path/to/restore\n", projectName))
		sb.WriteString("```\n")
	}

	return tools.TextResult(sb.String()), nil
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
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
