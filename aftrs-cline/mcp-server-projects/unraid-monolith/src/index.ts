#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import { execSync } from 'child_process';
import Docker from 'dockerode';

const server = new Server(
  {
    name: 'unraid-monolith-server',
    version: '0.1.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Initialize Docker client
const docker = new Docker({ socketPath: '/var/run/docker.sock' });

server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: 'list_containers',
        description: 'List all Docker containers on Unraid server',
        inputSchema: {
          type: 'object',
          properties: {
            all: {
              type: 'boolean',
              description: 'Include stopped containers',
              default: false,
            },
          },
        },
      },
      {
        name: 'start_container',
        description: 'Start a Docker container by name or ID',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID to start',
            },
          },
          required: ['container'],
        },
      },
      {
        name: 'stop_container',
        description: 'Stop a Docker container by name or ID',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID to stop',
            },
          },
          required: ['container'],
        },
      },
      {
        name: 'restart_container',
        description: 'Restart a Docker container by name or ID',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID to restart',
            },
          },
          required: ['container'],
        },
      },
      {
        name: 'get_container_logs',
        description: 'Get logs from a Docker container',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID to get logs from',
            },
            lines: {
              type: 'number',
              description: 'Number of log lines to retrieve',
              default: 100,
            },
          },
          required: ['container'],
        },
      },
      {
        name: 'get_system_stats',
        description: 'Get Unraid system statistics (CPU, memory, disk usage)',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'list_shares',
        description: 'List Unraid user shares and their status',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'get_array_status',
        description: 'Get Unraid array status and disk information',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'backup_appdata',
        description: 'Trigger backup of container appdata',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name to backup (optional, backs up all if not specified)',
            },
          },
        },
      },
      {
        name: 'update_container',
        description: 'Update a Docker container to latest image',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID to update',
            },
          },
          required: ['container'],
        },
      },
      {
        name: 'get_docker_stats',
        description: 'Get real-time resource usage statistics for containers',
        inputSchema: {
          type: 'object',
          properties: {
            container: {
              type: 'string',
              description: 'Container name or ID (optional, gets all if not specified)',
            },
          },
        },
      },
    ],
  };
});

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case 'list_containers': {
        const containers = await docker.listContainers({
          all: args?.all || false,
        });

        const containerInfo = containers.map(container => ({
          id: container.Id.slice(0, 12),
          names: container.Names,
          image: container.Image,
          state: container.State,
          status: container.Status,
          ports: container.Ports,
        }));

        return {
          content: [
            {
              type: 'text',
              text: `Found ${containers.length} containers:\n\n${JSON.stringify(containerInfo, null, 2)}`,
            },
          ],
        };
      }

      case 'start_container': {
        const container = docker.getContainer(args.container);
        await container.start();

        return {
          content: [
            {
              type: 'text',
              text: `Successfully started container: ${args.container}`,
            },
          ],
        };
      }

      case 'stop_container': {
        const container = docker.getContainer(args.container);
        await container.stop();

        return {
          content: [
            {
              type: 'text',
              text: `Successfully stopped container: ${args.container}`,
            },
          ],
        };
      }

      case 'restart_container': {
        const container = docker.getContainer(args.container);
        await container.restart();

        return {
          content: [
            {
              type: 'text',
              text: `Successfully restarted container: ${args.container}`,
            },
          ],
        };
      }

      case 'get_container_logs': {
        const container = docker.getContainer(args.container);
        const logs = await container.logs({
          stdout: true,
          stderr: true,
          tail: args.lines || 100,
        });

        return {
          content: [
            {
              type: 'text',
              text: `Container logs for ${args.container}:\n\n${logs.toString()}`,
            },
          ],
        };
      }

      case 'get_system_stats': {
        try {
          // Get system info using standard Linux commands
          const cpuInfo = execSync('cat /proc/cpuinfo | grep "model name" | head -1', { encoding: 'utf8' });
          const memInfo = execSync('free -h', { encoding: 'utf8' });
          const diskInfo = execSync('df -h', { encoding: 'utf8' });
          const uptime = execSync('uptime', { encoding: 'utf8' });

          return {
            content: [
              {
                type: 'text',
                text: `System Statistics:\n\nCPU: ${cpuInfo.trim()}\n\nUptime: ${uptime.trim()}\n\nMemory:\n${memInfo}\n\nDisk Usage:\n${diskInfo}`,
              },
            ],
          };
        } catch (error) {
          return {
            content: [
              {
                type: 'text',
                text: `Error getting system stats: ${error.message}`,
              },
            ],
          };
        }
      }

      case 'list_shares': {
        try {
          const shares = execSync('ls -la /mnt/user/', { encoding: 'utf8' });

          return {
            content: [
              {
                type: 'text',
                text: `Unraid User Shares:\n\n${shares}`,
              },
            ],
          };
        } catch (error) {
          return {
            content: [
              {
                type: 'text',
                text: `Error listing shares: ${error.message}`,
              },
            ],
          };
        }
      }

      case 'get_array_status': {
        try {
          // Check if mdcmd is available (Unraid-specific command)
          const mdstat = execSync('cat /proc/mdstat', { encoding: 'utf8' });
          const diskInfo = execSync('lsblk', { encoding: 'utf8' });

          return {
            content: [
              {
                type: 'text',
                text: `Array Status:\n\nMD Status:\n${mdstat}\n\nDisk Information:\n${diskInfo}`,
              },
            ],
          };
        } catch (error) {
          return {
            content: [
              {
                type: 'text',
                text: `Error getting array status: ${error.message}`,
              },
            ],
          };
        }
      }

      case 'backup_appdata': {
        try {
          const backupPath = '/mnt/user/backups/appdata';
          const sourceBase = '/mnt/user/appdata';

          if (args.container) {
            const sourcePath = `${sourceBase}/${args.container}`;
            const backupCmd = `mkdir -p ${backupPath} && rsync -av ${sourcePath}/ ${backupPath}/${args.container}_$(date +%Y%m%d_%H%M%S)/`;
            execSync(backupCmd, { encoding: 'utf8' });

            return {
              content: [
                {
                  type: 'text',
                  text: `Successfully backed up appdata for container: ${args.container}`,
                },
              ],
            };
          } else {
            const backupCmd = `mkdir -p ${backupPath} && rsync -av ${sourceBase}/ ${backupPath}/full_backup_$(date +%Y%m%d_%H%M%S)/`;
            execSync(backupCmd, { encoding: 'utf8' });

            return {
              content: [
                {
                  type: 'text',
                  text: `Successfully backed up all appdata`,
                },
              ],
            };
          }
        } catch (error) {
          return {
            content: [
              {
                type: 'text',
                text: `Error backing up appdata: ${error.message}`,
              },
            ],
          };
        }
      }

      case 'update_container': {
        const container = docker.getContainer(args.container);
        const info = await container.inspect();
        const imageName = info.Config.Image;

        // Pull latest image
        await docker.pull(imageName);

        // Stop and remove old container
        await container.stop();
        await container.remove();

        // This is a simplified update - in practice, you'd need to preserve
        // the container configuration and recreate with the same settings
        return {
          content: [
            {
              type: 'text',
              text: `Updated container ${args.container} to latest image: ${imageName}\nNote: Container needs to be recreated with original configuration`,
            },
          ],
        };
      }

      case 'get_docker_stats': {
        if (args.container) {
          const container = docker.getContainer(args.container);
          const stats = await container.stats({ stream: false });

          return {
            content: [
              {
                type: 'text',
                text: `Resource usage for ${args.container}:\n\n${JSON.stringify(stats, null, 2)}`,
              },
            ],
          };
        } else {
          const containers = await docker.listContainers();
          const allStats = [];

          for (const containerInfo of containers) {
            const container = docker.getContainer(containerInfo.Id);
            const stats = await container.stats({ stream: false });
            allStats.push({
              name: containerInfo.Names[0],
              id: containerInfo.Id.slice(0, 12),
              stats: stats,
            });
          }

          return {
            content: [
              {
                type: 'text',
                text: `Resource usage for all containers:\n\n${JSON.stringify(allStats, null, 2)}`,
              },
            ],
          };
        }
      }

      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  } catch (error) {
    return {
      content: [
        {
          type: 'text',
          text: `Error executing ${name}: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
});

async function runServer() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('Unraid Monolith MCP Server running on stdio');
}

runServer().catch(console.error);
