# Unraid Monolith MCP Server

A Model Context Protocol server for managing Unraid server operations, including Docker container management, system monitoring, and backup operations.

## Features

### Container Management
- **list_containers**: List all Docker containers with filtering options
- **start_container**: Start containers by name or ID
- **stop_container**: Stop running containers
- **restart_container**: Restart containers
- **get_container_logs**: Retrieve container logs with configurable line limits
- **update_container**: Update containers to latest image versions
- **get_docker_stats**: Get real-time resource usage statistics

### System Monitoring
- **get_system_stats**: CPU, memory, disk usage, and uptime information
- **get_array_status**: Unraid array status and disk information
- **list_shares**: List user shares and their status

### Backup Operations
- **backup_appdata**: Create backups of container appdata directories

## Installation

```bash
cd aftrs-cline/mcp-server-projects/unraid-monolith
npm install
npm run build
```

## Configuration

Add to your Cline MCP configuration:

```json
{
  "mcpServers": {
    "unraid-monolith": {
      "command": "node",
      "args": ["path/to/unraid-monolith/dist/index.js"],
      "env": {
        "NODE_ENV": "production"
      }
    }
  }
}
```

## Requirements

- Node.js 18+
- Docker socket access (`/var/run/docker.sock`)
- Unraid server environment
- Appropriate permissions for system commands

## Usage Examples

### List all containers
```typescript
await use_mcp_tool({
  server_name: "unraid-monolith",
  tool_name: "list_containers",
  arguments: { all: true }
});
```

### Start a container
```typescript
await use_mcp_tool({
  server_name: "unraid-monolith",
  tool_name: "start_container",
  arguments: { container: "plex" }
});
```

### Get system statistics
```typescript
await use_mcp_tool({
  server_name: "unraid-monolith",
  tool_name: "get_system_stats",
  arguments: {}
});
```

### Backup container appdata
```typescript
await use_mcp_tool({
  server_name: "unraid-monolith",
  tool_name: "backup_appdata",
  arguments: { container: "plex" }
});
```

## Security Considerations

- This server requires elevated privileges for system operations
- Docker socket access provides full container control
- Backup operations can consume significant disk space
- Monitor resource usage when using stats tools

## Troubleshooting

### Docker Socket Access
Ensure the Docker socket is accessible:
```bash
sudo chmod 666 /var/run/docker.sock
```

### Unraid Specific Commands
Some tools use Unraid-specific paths (`/mnt/user/`, `/proc/mdstat`). Ensure the server is running on an Unraid system for full functionality.

### Resource Usage
Container statistics can be resource-intensive. Use sparingly in production environments.

## Architecture Integration

This server integrates with the AFTRS ecosystem's:
- **Container orchestration**: Manages Docker workloads across the homelab
- **Backup strategy**: Automated appdata protection
- **Monitoring pipeline**: System health and resource tracking
- **Network management**: Works alongside OPNsense for complete infrastructure control

## Development

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check
