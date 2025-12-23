#!/usr/bin/env node

/**
 * MCP Server for AFTRS CLI - Asset-Driven Network Management
 * Provides AI access to network monitoring, asset management, and automation tools
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from '@modelcontextprotocol/sdk/types.js';
import { spawn } from 'child_process';

// AFTRS CLI base path
const AFTRS_CLI_PATH = process.env.AFTRS_CLI_PATH || '/home/hg/Docs/aftrs-void/aftrs_cli';

interface NetworkAsset {
  name: string;
  type: string;
  ip: string;
  hostname?: string;
  site?: string;
  services?: string[];
  status?: 'online' | 'offline' | 'unknown';
  lastChecked?: string;
}

interface TestResult {
  assetName: string;
  testType: string;
  success: boolean;
  duration: number;
  error?: string;
  details?: any;
}

interface MonitoringData {
  timestamp: string;
  assetName: string;
  metrics: {
    latency?: number;
    uptime?: boolean;
    serviceStatus?: Record<string, boolean>;
  };
}

class AFTRSCLIMCPServer {
  private server: Server;

  constructor() {
    this.server = new Server(
      {
        name: 'aftrs-cli',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private setupHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          // Asset Management Tools
          {
            name: 'list_assets',
            description: 'List all network assets with filtering and status information',
            inputSchema: {
              type: 'object',
              properties: {
                type: {
                  type: 'string',
                  enum: ['all', 'router', 'server', 'device', 'service'],
                  description: 'Filter assets by type',
                  default: 'all',
                },
                site: {
                  type: 'string',
                  description: 'Filter assets by site location',
                },
                status: {
                  type: 'string',
                  enum: ['all', 'online', 'offline', 'unknown'],
                  description: 'Filter assets by status',
                  default: 'all',
                },
                format: {
                  type: 'string',
                  enum: ['json', 'table', 'yaml'],
                  description: 'Output format',
                  default: 'table',
                },
              },
            },
          },
          {
            name: 'asset_details',
            description: 'Get detailed information about a specific network asset',
            inputSchema: {
              type: 'object',
              properties: {
                asset_name: {
                  type: 'string',
                  description: 'Name or ID of the asset to examine',
                },
                include_history: {
                  type: 'boolean',
                  description: 'Include monitoring history',
                  default: false,
                },
              },
              required: ['asset_name'],
            },
          },
          {
            name: 'import_assets',
            description: 'Import assets from Notion CSV or other sources',
            inputSchema: {
              type: 'object',
              properties: {
                source_file: {
                  type: 'string',
                  description: 'Path to CSV or YAML file to import',
                },
                merge_existing: {
                  type: 'boolean',
                  description: 'Merge with existing assets instead of replacing',
                  default: true,
                },
                validate_only: {
                  type: 'boolean',
                  description: 'Only validate import without applying changes',
                  default: false,
                },
              },
              required: ['source_file'],
            },
          },
          // Network Testing & Diagnostics
          {
            name: 'run_diagnostics',
            description: 'Run comprehensive network diagnostics and health checks',
            inputSchema: {
              type: 'object',
              properties: {
                target: {
                  type: 'string',
                  description: 'Target asset name or "all" for system-wide diagnostics',
                  default: 'all',
                },
                test_types: {
                  type: 'array',
                  items: {
                    type: 'string',
                    enum: ['connectivity', 'services', 'performance', 'security', 'all'],
                  },
                  description: 'Types of tests to run',
                  default: ['all'],
                },
                parallel: {
                  type: 'boolean',
                  description: 'Run tests in parallel for faster execution',
                  default: true,
                },
              },
            },
          },
          {
            name: 'generate_tests',
            description: 'Generate asset-driven test configurations',
            inputSchema: {
              type: 'object',
              properties: {
                asset_filter: {
                  type: 'string',
                  description: 'Filter for which assets to generate tests',
                },
                test_categories: {
                  type: 'array',
                  items: {
                    type: 'string',
                    enum: ['connectivity', 'services', 'performance'],
                  },
                  description: 'Categories of tests to generate',
                  default: ['connectivity', 'services'],
                },
                save_to_file: {
                  type: 'boolean',
                  description: 'Save generated tests to configuration files',
                  default: true,
                },
              },
            },
          },
          // Network Monitoring
          {
            name: 'monitor_status',
            description: 'Get real-time network monitoring status and metrics',
            inputSchema: {
              type: 'object',
              properties: {
                asset_name: {
                  type: 'string',
                  description: 'Specific asset to monitor (empty for all)',
                },
                include_services: {
                  type: 'boolean',
                  description: 'Include service-level monitoring',
                  default: true,
                },
                time_range: {
                  type: 'string',
                  enum: ['1h', '4h', '24h', '7d', '30d'],
                  description: 'Time range for historical data',
                  default: '4h',
                },
              },
            },
          },
          {
            name: 'network_topology',
            description: 'Generate and display network topology visualization',
            inputSchema: {
              type: 'object',
              properties: {
                output_format: {
                  type: 'string',
                  enum: ['ascii', 'svg', 'png', 'json'],
                  description: 'Output format for topology diagram',
                  default: 'ascii',
                },
                include_services: {
                  type: 'boolean',
                  description: 'Show services in topology',
                  default: false,
                },
                show_connections: {
                  type: 'boolean',
                  description: 'Display connection details',
                  default: true,
                },
              },
            },
          },
          // Tailscale Management
          {
            name: 'tailscale_status',
            description: 'Get Tailscale network status and device information',
            inputSchema: {
              type: 'object',
              properties: {
                detailed: {
                  type: 'boolean',
                  description: 'Include detailed device information',
                  default: false,
                },
                include_routes: {
                  type: 'boolean',
                  description: 'Show routing table information',
                  default: false,
                },
              },
            },
          },
          {
            name: 'tailscale_manage',
            description: 'Manage Tailscale network configuration and devices',
            inputSchema: {
              type: 'object',
              properties: {
                action: {
                  type: 'string',
                  enum: ['up', 'down', 'logout', 'set-dns', 'ping', 'netcheck'],
                  description: 'Tailscale management action',
                },
                target_device: {
                  type: 'string',
                  description: 'Target device name or IP (for ping/netcheck)',
                },
                dns_servers: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'DNS servers to set (for set-dns action)',
                },
              },
              required: ['action'],
            },
          },
          // Documentation & Reporting
          {
            name: 'generate_documentation',
            description: 'Generate network documentation and asset reports',
            inputSchema: {
              type: 'object',
              properties: {
                doc_type: {
                  type: 'string',
                  enum: ['asset-table', 'network-diagram', 'status-report', 'all'],
                  description: 'Type of documentation to generate',
                  default: 'all',
                },
                output_format: {
                  type: 'string',
                  enum: ['markdown', 'html', 'pdf'],
                  description: 'Output format for documentation',
                  default: 'markdown',
                },
                include_history: {
                  type: 'boolean',
                  description: 'Include monitoring history in reports',
                  default: false,
                },
              },
            },
          },
          // DMZ & Firewall Analysis
          {
            name: 'dmz_analysis',
            description: 'Analyze DMZ configuration and port forwarding',
            inputSchema: {
              type: 'object',
              properties: {
                analysis_type: {
                  type: 'string',
                  enum: ['reachability', 'bridge', 'full'],
                  description: 'Type of DMZ analysis to perform',
                  default: 'full',
                },
                target_hosts: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'Specific hosts to analyze (empty for all)',
                },
              },
            },
          },
          {
            name: 'nat_analysis',
            description: 'Analyze NAT chains and packet flow',
            inputSchema: {
              type: 'object',
              properties: {
                analysis_type: {
                  type: 'string',
                  enum: ['topology', 'live', 'report'],
                  description: 'Type of NAT analysis',
                  default: 'report',
                },
                protocol: {
                  type: 'string',
                  enum: ['tcp', 'udp', 'both'],
                  description: 'Protocol to analyze',
                  default: 'both',
                },
              },
            },
          },
          // Git & Repository Management
          {
            name: 'git_operations',
            description: 'Perform bulk Git operations across AFTRS repositories',
            inputSchema: {
              type: 'object',
              properties: {
                operation: {
                  type: 'string',
                  enum: ['status', 'pull', 'push', 'sync', 'commit', 'clean', 'report'],
                  description: 'Git operation to perform',
                },
                repositories: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'Specific repositories (empty for all)',
                },
                commit_message: {
                  type: 'string',
                  description: 'Commit message (for commit operation)',
                },
                dry_run: {
                  type: 'boolean',
                  description: 'Preview changes without applying',
                  default: false,
                },
              },
              required: ['operation'],
            },
          },
          {
            name: 'git_clone_org',
            description: 'Clone entire GitHub organizations with proper identity setup',
            inputSchema: {
              type: 'object',
              properties: {
                organization: {
                  type: 'string',
                  description: 'GitHub organization name to clone',
                },
                target_directory: {
                  type: 'string',
                  description: 'Target directory for cloning',
                },
                identity_profile: {
                  type: 'string',
                  description: 'Git identity profile to use',
                },
              },
              required: ['organization'],
            },
          },
          // Performance & Optimization
          {
            name: 'performance_analysis',
            description: 'Analyze network performance and identify optimization opportunities',
            inputSchema: {
              type: 'object',
              properties: {
                analysis_scope: {
                  type: 'string',
                  enum: ['network', 'assets', 'services', 'all'],
                  description: 'Scope of performance analysis',
                  default: 'all',
                },
                optimization_suggestions: {
                  type: 'boolean',
                  description: 'Include optimization recommendations',
                  default: true,
                },
              },
            },
          },
          // Web Dashboard Management
          {
            name: 'dashboard_control',
            description: 'Control AFTRS CLI web dashboard',
            inputSchema: {
              type: 'object',
              properties: {
                action: {
                  type: 'string',
                  enum: ['start', 'stop', 'restart', 'status'],
                  description: 'Dashboard control action',
                },
                port: {
                  type: 'number',
                  description: 'Port number for dashboard (default: 8080)',
                  default: 8080,
                },
              },
              required: ['action'],
            },
          },
          // System Management
          {
            name: 'system_status',
            description: 'Get comprehensive AFTRS CLI system status',
            inputSchema: {
              type: 'object',
              properties: {
                include_dependencies: {
                  type: 'boolean',
                  description: 'Check system dependencies',
                  default: true,
                },
                include_performance: {
                  type: 'boolean',
                  description: 'Include performance metrics',
                  default: false,
                },
              },
            },
          },
        ],
      };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      try {
        const { name, arguments: args } = request.params;

        switch (name) {
          case 'list_assets':
            return await this.listAssets(args.type, args.site, args.status, args.format);

          case 'asset_details':
            return await this.getAssetDetails(args.asset_name, args.include_history);

          case 'import_assets':
            return await this.importAssets(args.source_file, args.merge_existing, args.validate_only);

          case 'run_diagnostics':
            return await this.runDiagnostics(args.target, args.test_types, args.parallel);

          case 'generate_tests':
            return await this.generateTests(args.asset_filter, args.test_categories, args.save_to_file);

          case 'monitor_status':
            return await this.getMonitoringStatus(args.asset_name, args.include_services, args.time_range);

          case 'network_topology':
            return await this.generateNetworkTopology(args.output_format, args.include_services, args.show_connections);

          case 'tailscale_status':
            return await this.getTailscaleStatus(args.detailed, args.include_routes);

          case 'tailscale_manage':
            return await this.manageTailscale(args.action, args.target_device, args.dns_servers);

          case 'generate_documentation':
            return await this.generateDocumentation(args.doc_type, args.output_format, args.include_history);

          case 'dmz_analysis':
            return await this.analyzeDMZ(args.analysis_type, args.target_hosts);

          case 'nat_analysis':
            return await this.analyzeNAT(args.analysis_type, args.protocol);

          case 'git_operations':
            return await this.performGitOperations(args.operation, args.repositories, args.commit_message, args.dry_run);

          case 'git_clone_org':
            return await this.cloneGitHubOrg(args.organization, args.target_directory, args.identity_profile);

          case 'performance_analysis':
            return await this.analyzePerformance(args.analysis_scope, args.optimization_suggestions);

          case 'dashboard_control':
            return await this.controlDashboard(args.action, args.port);

          case 'system_status':
            return await this.getSystemStatus(args.include_dependencies, args.include_performance);

          default:
            throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${name}`);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        throw new McpError(ErrorCode.InternalError, `AFTRS CLI error: ${errorMessage}`);
      }
    });
  }

  private async execAFTRSCommand(command: string[]): Promise<{ stdout: string; stderr: string; exitCode: number }> {
    return new Promise((resolve) => {
      const aftrsProcess = spawn('./aftrs.sh', command, {
        cwd: AFTRS_CLI_PATH,
        stdio: ['pipe', 'pipe', 'pipe'],
      });

      let stdout = '';
      let stderr = '';

      aftrsProcess.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      aftrsProcess.stderr?.on('data', (data) => {
        stderr += data.toString();
      });

      aftrsProcess.on('close', (exitCode) => {
        resolve({ stdout, stderr, exitCode: exitCode || 0 });
      });
    });
  }

  private async listAssets(type = 'all', site?: string, status = 'all', format = 'table'): Promise<any> {
    const command = ['assets', 'list'];

    if (type !== 'all') {
      command.push('--type', type);
    }

    if (site) {
      command.push('--site', site);
    }

    if (status !== 'all') {
      command.push('--status', status);
    }

    command.push('--format', format);

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Network Assets:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async getAssetDetails(assetName: string, includeHistory = false): Promise<any> {
    const command = ['assets', 'details', assetName];

    if (includeHistory) {
      command.push('--history');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Asset Details for ${assetName}:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async importAssets(sourceFile: string, mergeExisting = true, validateOnly = false): Promise<any> {
    const command = ['import'];

    if (sourceFile.endsWith('.csv')) {
      command.push('notion', sourceFile);
    } else {
      command.push('asset', sourceFile);
    }

    if (mergeExisting) {
      command.push('--merge');
    }

    if (validateOnly) {
      command.push('--validate-only');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Asset Import Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async runDiagnostics(target = 'all', testTypes = ['all'], parallel = true): Promise<any> {
    let command: string[];

    if (target === 'all') {
      command = ['diag'];
    } else {
      command = ['tests', 'run', '--target', target];
    }

    if (testTypes.length > 0 && !testTypes.includes('all')) {
      command.push('--category', testTypes.join(','));
    }

    if (parallel) {
      command.push('--parallel', '4');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Diagnostics Results:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async generateTests(assetFilter?: string, testCategories = ['connectivity', 'services'], saveToFile = true): Promise<any> {
    const command = ['tests', 'generate'];

    if (assetFilter) {
      command.push('--filter', assetFilter);
    }

    if (testCategories.length > 0) {
      command.push('--category', testCategories.join(','));
    }

    if (saveToFile) {
      command.push('--save');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Test Generation Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async getMonitoringStatus(assetName?: string, includeServices = true, timeRange = '4h'): Promise<any> {
    const command = ['monitor', 'status'];

    if (assetName) {
      command.push('--asset', assetName);
    }

    if (includeServices) {
      command.push('--services');
    }

    command.push('--time-range', timeRange);

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Monitoring Status:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async generateNetworkTopology(outputFormat = 'ascii', includeServices = false, showConnections = true): Promise<any> {
    const command = ['monitor', 'topology'];

    command.push('--format', outputFormat);

    if (includeServices) {
      command.push('--services');
    }

    if (showConnections) {
      command.push('--connections');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Network Topology:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async getTailscaleStatus(detailed = false, includeRoutes = false): Promise<any> {
    const command = ['tailscale', 'status'];

    if (detailed) {
      command.push('--detailed');
    }

    if (includeRoutes) {
      command.push('--routes');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Tailscale Status:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async manageTailscale(action: string, targetDevice?: string, dnsServers?: string[]): Promise<any> {
    const command = ['tailscale', action];

    if (targetDevice && (action === 'ping' || action === 'netcheck')) {
      command.push(targetDevice);
    }

    if (dnsServers && action === 'set-dns') {
      command.push(...dnsServers);
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Tailscale ${action} Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async generateDocumentation(docType = 'all', outputFormat = 'markdown', includeHistory = false): Promise<any> {
    let command: string[];

    if (docType === 'all') {
      command = ['docs', 'generate'];
    } else {
      command = ['docs', docType];
    }

    command.push('--format', outputFormat);

    if (includeHistory) {
      command.push('--history');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Documentation Generation Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async analyzeDMZ(analysisType = 'full', targetHosts?: string[]): Promise<any> {
    const command = ['dmz', analysisType];

    if (targetHosts && targetHosts.length > 0) {
      command.push('--hosts', targetHosts.join(','));
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `DMZ Analysis Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async analyzeNAT(analysisType = 'report', protocol = 'both'): Promise<any> {
    const command = ['nat', analysisType];

    if (protocol !== 'both') {
      command.push('--protocol', protocol);
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `NAT Analysis Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async performGitOperations(operation: string, repositories?: string[], commitMessage?: string, dryRun = false): Promise<any> {
    const command = ['git', 'bulk', operation];

    if (repositories && repositories.length > 0) {
      command.push('--repos', repositories.join(','));
    }

    if (commitMessage && operation === 'commit') {
      command.push('--message', commitMessage);
    }

    if (dryRun) {
      command.push('--dry-run');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Git ${operation} Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async cloneGitHubOrg(organization: string, targetDirectory?: string, identityProfile?: string): Promise<any> {
    const command = ['git', 'clone', 'org', organization];

    if (targetDirectory) {
      command.push(targetDirectory);
    }

    if (identityProfile) {
      command.push('--identity', identityProfile);
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `GitHub Organization Clone Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async analyzePerformance(analysisScope = 'all', optimizationSuggestions = true): Promise<any> {
    const command = ['optimize', 'analyze'];

    if (analysisScope !== 'all') {
      command.push('--scope', analysisScope);
    }

    if (optimizationSuggestions) {
      command.push('--suggestions');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Performance Analysis Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async controlDashboard(action: string, port = 8080): Promise<any> {
    const command = ['dashboard', action];

    if (action === 'start' && port !== 8080) {
      command.push('--port', port.toString());
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `Dashboard ${action} Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async getSystemStatus(includeDependencies = true, includePerformance = false): Promise<any> {
    const command = ['status'];

    if (includeDependencies) {
      command.push('--deps');
    }

    if (includePerformance) {
      command.push('--performance');
    }

    const result = await this.execAFTRSCommand(command);

    return {
      content: [
        {
          type: 'text',
          text: `System Status:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('AFTRS CLI MCP server running on stdio');
  }
}

const server = new AFTRSCLIMCPServer();
server.run().catch(console.error);
